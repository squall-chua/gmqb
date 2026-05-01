package gmqb

import (
	"context"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

// queueDoc is the internal wrapper for messages stored in MongoDB.
type queueDoc[T any] struct {
	ID          bson.ObjectID `bson:"_id,omitempty"`
	Payload     T             `bson:"p"`
	Status      string        `bson:"s"` // pending | processing | done | dead
	Attempts    int           `bson:"a"`
	MaxAttempts int           `bson:"ma"`
	ClaimedAt   *time.Time    `bson:"ca,omitempty"` // visibility timeout anchor
	ClaimedBy   string        `bson:"cb,omitempty"` // worker instance ID
	CreatedAt   time.Time     `bson:"t"`
	FailedGroup string        `bson:"fg,omitempty"` // populated only in DLQ for fan-out messages
}

const (
	statusPending    = "pending"
	statusProcessing = "processing"
	statusDone       = "done"
	statusDead       = "dead"
)

// collection name helpers
func primaryCollName(base string) string {
	return base
}

func dlqCollName(base string) string {
	return fmt.Sprintf("%s_dlq", base)
}

func dedupCollName(base string) string {
	return fmt.Sprintf("%s_dedup", base)
}

func offsetCollName(base string) string {
	return fmt.Sprintf("%s_offsets", base)
}

func claimsCollName(base string) string {
	return fmt.Sprintf("%s_claims", base)
}

type offsetDoc struct {
	ID            string        `bson:"_id"` // group name
	LastProcessed bson.ObjectID `bson:"lp"`
}

type dedupDoc struct {
	ID        string    `bson:"_id"` // "{msgID}:{groupName}"
	CreatedAt time.Time `bson:"t"`
}

type groupClaimDoc struct {
	ID        groupClaimID `bson:"_id"`
	Status    string       `bson:"s"`
	WorkerID  string       `bson:"w"`
	ClaimedAt time.Time    `bson:"ca"`
	Attempts  int          `bson:"a"`
	CreatedAt time.Time    `bson:"t"`
}

type groupClaimID struct {
	Group string        `bson:"g"`
	MsgID bson.ObjectID `bson:"m"`
}

func claimOne[T any](ctx context.Context, coll *mongo.Collection, workerID string, lastID bson.ObjectID, visibilityTimeout time.Duration, targetStatus string) (*queueDoc[T], error) {
	now := time.Now()

	// Filter:
	// 1. status="pending" AND _id > lastID
	// 2. OR status="processing" AND claimedAt < now - visibilityTimeout (stale/crashed workers)
	// For fan-out mode, lastID is used to ensure we only get new messages.
	// For load-balancing mode, lastID is usually NilObjectID.

	filter := bson.M{
		"$or": []bson.M{
			{
				"s":   statusPending,
				"_id": bson.M{"$gt": lastID},
			},
			{
				"s":  statusProcessing,
				"ca": bson.M{"$lt": now.Add(-visibilityTimeout)},
			},
		},
	}

	update := bson.M{
		"$set": bson.M{
			"s":  targetStatus,
			"ca": now,
			"cb": workerID,
		},
		"$inc": bson.M{"a": 1},
	}

	opts := options.FindOneAndUpdate().
		SetReturnDocument(options.After).
		SetSort(bson.D{{Key: "_id", Value: 1}})

	var doc queueDoc[T]
	err := coll.FindOneAndUpdate(ctx, filter, update, opts).Decode(&doc)
	if err != nil {
		return nil, err
	}
	return &doc, nil
}

func ackOne(ctx context.Context, coll *mongo.Collection, id bson.ObjectID) error {
	_, err := coll.UpdateOne(ctx, bson.M{"_id": id}, bson.M{"$set": bson.M{"s": statusDone}})
	return err
}

func nackOne(ctx context.Context, coll *mongo.Collection, id bson.ObjectID) error {
	// Put it back to pending so it can be picked up again immediately or after visibility timeout
	// depends on the claimOne filter. Here we just set it back to pending.
	_, err := coll.UpdateOne(ctx, bson.M{"_id": id}, bson.M{"$set": bson.M{"s": statusPending}})
	return err
}

func moveToDLQ[T any](ctx context.Context, primaryColl, dlqColl *mongo.Collection, doc *queueDoc[T]) error {
	// 1. Mark as dead in primary (don't delete, to avoid breaking other fan-out groups)
	_, err := primaryColl.UpdateOne(ctx, bson.M{"_id": doc.ID}, bson.M{"$set": bson.M{"s": statusDead}})
	if err != nil {
		return fmt.Errorf("mark as dead in primary: %w", err)
	}

	// 2. Insert into DLQ
	doc.Status = statusDead
	doc.ClaimedAt = nil
	doc.ClaimedBy = ""
	_, err = dlqColl.InsertOne(ctx, doc)
	if err != nil {
		return fmt.Errorf("insert into dlq: %w", err)
	}

	return nil
}

func ackFanOut(ctx context.Context, claims *mongo.Collection, group string, msgID bson.ObjectID) error {
	_, err := claims.UpdateOne(ctx, bson.M{"_id": groupClaimID{Group: group, MsgID: msgID}}, bson.M{"$set": bson.M{"s": statusDone}})
	return err
}

func nackFanOut(ctx context.Context, claims *mongo.Collection, group string, msgID bson.ObjectID) error {
	// Put it back to pending for the group
	_, err := claims.UpdateOne(ctx, bson.M{"_id": groupClaimID{Group: group, MsgID: msgID}}, bson.M{"$set": bson.M{"s": statusPending}})
	return err
}

func moveToDLQFanOut[T any](ctx context.Context, claims *mongo.Collection, dlqColl *mongo.Collection, group string, doc *queueDoc[T]) error {
	// 1. Mark as dead in group claims
	_, err := claims.UpdateOne(ctx, bson.M{"_id": groupClaimID{Group: group, MsgID: doc.ID}}, bson.M{"$set": bson.M{"s": statusDead}})
	if err != nil {
		return err
	}

	// 2. Insert into DLQ
	doc.Status = statusDead
	doc.ClaimedAt = nil
	doc.ClaimedBy = ""
	doc.FailedGroup = group
	_, err = dlqColl.InsertOne(ctx, doc)
	return err
}

func claimFanOut[T any](ctx context.Context, coll *mongo.Collection, offsets *mongo.Collection, claims *mongo.Collection, groupName string, workerID string, visibilityTimeout time.Duration) (*queueDoc[T], error) {
	now := time.Now()
	staleTime := now.Add(-visibilityTimeout)

	// 1. Try to find a stale claim to retry
	retryFilter := bson.M{
		"_id.g": groupName,
		"s":     statusProcessing,
		"ca":    bson.M{"$lt": staleTime},
	}
	retryUpdate := bson.M{
		"$set": bson.M{
			"w":  workerID,
			"ca": now,
		},
		"$inc": bson.M{"a": 1},
	}

	var claim groupClaimDoc
	err := claims.FindOneAndUpdate(ctx, retryFilter, retryUpdate).Decode(&claim)
	if err == nil {
		// Found a stale claim, fetch the actual message
		var doc queueDoc[T]
		err = coll.FindOne(ctx, bson.M{"_id": claim.ID.MsgID}).Decode(&doc)
		if err == nil {
			return &doc, nil
		}
	}

	// 2. No stale claims, try to claim a NEW message
	// Get current offset
	var offset offsetDoc
	err = offsets.FindOne(ctx, bson.M{"_id": groupName}).Decode(&offset)
	if err != nil && err != mongo.ErrNoDocuments {
		return nil, err
	}

	// Find next message > offset
	filter := bson.M{"_id": bson.M{"$gt": offset.LastProcessed}}
	findOpts := options.FindOne().SetSort(bson.D{{Key: "_id", Value: 1}})

	var doc queueDoc[T]
	err = coll.FindOne(ctx, filter, findOpts).Decode(&doc)
	if err != nil {
		return nil, err
	}

	// Create a NEW claim
	newClaim := groupClaimDoc{
		ID:        groupClaimID{Group: groupName, MsgID: doc.ID},
		Status:    statusProcessing,
		WorkerID:  workerID,
		ClaimedAt: now,
		Attempts:  1,
		CreatedAt: now,
	}

	_, err = claims.InsertOne(ctx, newClaim)
	if err != nil {
		if mongo.IsDuplicateKeyError(err) {
			return nil, mongo.ErrNoDocuments // Someone else got it
		}
		return nil, err
	}

	// Advance the offset greedily to keep the "pointer" moving.
	// Note: If this fails, it's just an optimization loss (we'll re-scan next time).
	updateFilter := bson.M{"_id": groupName}
	if offset.ID != "" {
		updateFilter["lp"] = offset.LastProcessed
	}
	_, _ = offsets.UpdateOne(ctx, updateFilter, bson.M{"$set": bson.M{"lp": doc.ID}}, options.UpdateOne().SetUpsert(true))

	return &doc, nil
}
