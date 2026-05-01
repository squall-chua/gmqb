package gmqb

import (
	"context"
	"fmt"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

// DLQ provides access to dead-lettered messages.
type DLQ[T any] struct {
	q    *Queue[T]
	coll *mongo.Collection
}

// List returns dead-lettered messages with pagination.
func (d *DLQ[T]) List(ctx context.Context, limit, offset int64) ([]queueDoc[T], error) {
	opts := options.Find().
		SetLimit(limit).
		SetSkip(offset).
		SetSort(bson.D{{Key: "t", Value: -1}})

	cursor, err := d.coll.Find(ctx, bson.D{}, opts)
	if err != nil {
		return nil, fmt.Errorf("gmqb dlq: list: %w", err)
	}
	var results []queueDoc[T]
	if err := cursor.All(ctx, &results); err != nil {
		return nil, fmt.Errorf("gmqb dlq: decode: %w", err)
	}
	return results, nil
}

// Requeue moves a dead-lettered message back to the queue for processing.
func (d *DLQ[T]) Requeue(ctx context.Context, id bson.ObjectID) error {
	// 1. Find and delete from DLQ
	var doc queueDoc[T]
	err := d.coll.FindOneAndDelete(ctx, bson.D{{Key: "_id", Value: id}}).Decode(&doc)
	if err != nil {
		return fmt.Errorf("gmqb dlq: find and delete: %w", err)
	}

	if doc.FailedGroup != "" {
		// 2a. Fan-Out: Reset the group claim status instead of touching primary
		update := bson.M{
			"$set": bson.M{
				"s":  statusPending,
				"a":  0,
				"ca": nil,
				"w":  "",
			},
		}
		_, err = d.q.claims.UpdateOne(ctx, bson.M{"_id": groupClaimID{Group: doc.FailedGroup, MsgID: doc.ID}}, update)
		if err != nil {
			return fmt.Errorf("gmqb dlq: requeue group claim: %w", err)
		}
	} else {
		// 2b. Load-Balanced: Reset primary message state
		doc.Status = statusPending
		doc.Attempts = 0
		doc.ClaimedAt = nil
		doc.ClaimedBy = ""

		// Use ReplaceOne with upsert because the message might still exist with statusDead
		_, err = d.q.coll.ReplaceOne(ctx, bson.M{"_id": doc.ID}, doc, options.Replace().SetUpsert(true))
		if err != nil {
			return fmt.Errorf("gmqb dlq: requeue primary replace: %w", err)
		}
	}

	// Trigger wake channel
	select {
	case d.q.wake <- struct{}{}:
	default:
	}

	return nil
}

// Purge removes all messages from the DLQ.
func (d *DLQ[T]) Purge(ctx context.Context) error {
	_, err := d.coll.DeleteMany(ctx, bson.D{})
	if err != nil {
		return fmt.Errorf("gmqb dlq: purge: %w", err)
	}
	return nil
}
