package mq_test

import (
	"context"
	"testing"
	"time"

	"github.com/squall-chua/gmqb/mq"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tryvium-travels/memongo"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	mongooptions "go.mongodb.org/mongo-driver/v2/mongo/options"
)

func startStandalone(t *testing.T) (*mongo.Database, *mongo.Client) {
	t.Helper()

	srv, err := memongo.StartWithOptions(&memongo.Options{
		MongoVersion: "8.2.5",
	})
	require.NoError(t, err)
	t.Cleanup(srv.Stop)

	client, err := mongo.Connect(mongooptions.Client().ApplyURI(srv.URI()))
	require.NoError(t, err)
	t.Cleanup(func() { _ = client.Disconnect(context.Background()) })

	dbName := memongo.RandomDatabase()
	return client.Database(dbName), client
}

func TestQueue_Enqueue(t *testing.T) {
	db, _ := startStandalone(t)
	ctx := context.Background()

	type Job struct {
		Name string `bson:"name"`
	}

	q, err := mq.NewQueue[Job](db, "test_queue", mq.DefaultQueueOpts)
	require.NoError(t, err)

	err = q.EnsureIndexes(ctx)
	require.NoError(t, err)

	id, err := q.Enqueue(ctx, Job{Name: "task1"})
	require.NoError(t, err)
	assert.False(t, id.IsZero())

	// Verify it's in the collection
	var doc struct {
		ID     interface{} `bson:"_id"`
		Status string      `bson:"s"`
	}
	err = db.Collection("test_queue").FindOne(ctx, map[string]interface{}{"_id": id}).Decode(&doc)
	require.NoError(t, err)
	assert.Equal(t, "pending", doc.Status)
}

func TestDLQ_Purge(t *testing.T) {
	db, _ := startStandalone(t)
	ctx := context.Background()

	type Job struct {
		ID int `bson:"id"`
	}

	q, err := mq.NewQueue[Job](db, "dlq_test", mq.DefaultQueueOpts)
	require.NoError(t, err)

	// Manually insert into DLQ
	_, err = db.Collection("dlq_test_dlq").InsertOne(ctx, map[string]interface{}{
		"p": Job{ID: 1},
		"t": time.Now(),
	})
	require.NoError(t, err)

	dlq := q.DLQ()
	list, err := dlq.List(ctx, 10, 0)
	require.NoError(t, err)
	assert.Len(t, list, 1)

	err = dlq.Purge(ctx)
	require.NoError(t, err)

	list, err = dlq.List(ctx, 10, 0)
	require.NoError(t, err)
	assert.Len(t, list, 0)
}

func TestQueue_Enqueue_Idempotency(t *testing.T) {
	db, _ := startStandalone(t)
	ctx := context.Background()

	type Job struct {
		X int `bson:"x"`
	}

	q, err := mq.NewQueue[Job](db, "idempotency_test", mq.DefaultQueueOpts)
	require.NoError(t, err)
	_ = q.EnsureIndexes(ctx)

	id := bson.NewObjectID()

	// First enqueue
	id1, err := q.Enqueue(ctx, Job{X: 1}, mq.WithID(id), mq.WithIdempotent())
	require.NoError(t, err)
	assert.Equal(t, id, id1)

	// Second enqueue with same ID (should succeed idempotently)
	id2, err := q.Enqueue(ctx, Job{X: 1}, mq.WithID(id), mq.WithIdempotent())
	require.NoError(t, err)
	assert.Equal(t, id, id2)

	// Third enqueue with same ID but WITHOUT WithIdempotent (should fail)
	_, err = q.Enqueue(ctx, Job{X: 1}, mq.WithID(id))
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "duplicate key")
}
