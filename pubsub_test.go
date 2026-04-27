package gmqb_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tryvium-travels/memongo"
	"go.mongodb.org/mongo-driver/v2/mongo"
	mongooptions "go.mongodb.org/mongo-driver/v2/mongo/options"

	"github.com/squall-chua/gmqb"
)

type MyEvent struct {
	ID   int    `bson:"id"`
	Data string `bson:"data"`
}

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

func TestTailablePubSub_PublishSubscribe(t *testing.T) {
	db, _ := startStandalone(t)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	topic := "test_topic"
	bus, err := gmqb.NewTailablePubSub[MyEvent](db, topic, gmqb.DefaultCappedOpts)
	require.NoError(t, err)

	ch, subCancel := bus.Subscribe(ctx)
	defer subCancel()

	// Give it a tiny bit of time to start the goroutine (though not strictly necessary as tailable cursors wait)
	time.Sleep(100 * time.Millisecond)

	event := MyEvent{ID: 1, Data: "hello"}
	err = bus.Publish(ctx, event)
	require.NoError(t, err)

	select {
	case received := <-ch:
		assert.Equal(t, event, received)
	case <-time.After(2 * time.Second):
		t.Fatal("timed out waiting for event")
	}
}

func TestTailablePubSub_MultipleSubscribers(t *testing.T) {
	db, _ := startStandalone(t)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	topic := "multi_topic"
	bus, err := gmqb.NewTailablePubSub[MyEvent](db, topic, gmqb.DefaultCappedOpts)
	require.NoError(t, err)

	ch1, cancel1 := bus.Subscribe(ctx)
	defer cancel1()
	ch2, cancel2 := bus.Subscribe(ctx)
	defer cancel2()

	time.Sleep(100 * time.Millisecond)

	event := MyEvent{ID: 42, Data: "broadcast"}
	err = bus.Publish(ctx, event)
	require.NoError(t, err)

	// Both should receive it
	select {
	case r1 := <-ch1:
		assert.Equal(t, event, r1)
	case <-time.After(1 * time.Second):
		t.Fatal("sub1 timed out")
	}

	select {
	case r2 := <-ch2:
		assert.Equal(t, event, r2)
	case <-time.After(1 * time.Second):
		t.Fatal("sub2 timed out")
	}
}

func TestTailablePubSub_CancelStopsSubscriber(t *testing.T) {
	db, _ := startStandalone(t)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	bus, _ := gmqb.NewTailablePubSub[MyEvent](db, "cancel_topic", gmqb.DefaultCappedOpts)
	ch, subCancel := bus.Subscribe(ctx)

	subCancel()

	// Channel should be closed
	select {
	case _, ok := <-ch:
		assert.False(t, ok, "channel should be closed after cancel")
	case <-time.After(1 * time.Second):
		t.Fatal("channel didn't close in time")
	}
}

func TestTailablePubSub_LargeBurst(t *testing.T) {
	db, _ := startStandalone(t)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	bus, _ := gmqb.NewTailablePubSub[MyEvent](db, "burst_topic", gmqb.CappedOpts{SizeBytes: 1024 * 1024})
	ch, subCancel := bus.Subscribe(ctx)
	defer subCancel()

	count := 100
	for i := 0; i < count; i++ {
		err := bus.Publish(ctx, MyEvent{ID: i})
		require.NoError(t, err)
	}

	receivedCount := 0
	for i := 0; i < count; i++ {
		select {
		case ev := <-ch:
			assert.Equal(t, i, ev.ID)
			receivedCount++
		case <-time.After(2 * time.Second):
			t.Fatalf("timed out at index %d", i)
		}
	}
	assert.Equal(t, count, receivedCount)
}

func TestTailablePubSub_IdempotentCreation(t *testing.T) {
	db, _ := startStandalone(t)
	topic := "idempotent_topic"

	// First creation
	_, err := gmqb.NewTailablePubSub[MyEvent](db, topic, gmqb.DefaultCappedOpts)
	require.NoError(t, err)

	// Second creation should not fail even if it already exists
	_, err = gmqb.NewTailablePubSub[MyEvent](db, topic, gmqb.DefaultCappedOpts)
	require.NoError(t, err, "second creation should be idempotent")
}

func TestTailablePubSub_ReconnectAfterCursorDeath(t *testing.T) {
	db, _ := startStandalone(t)
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	topic := "reconnect_topic"
	bus, err := gmqb.NewTailablePubSub[MyEvent](db, topic, gmqb.DefaultCappedOpts)
	require.NoError(t, err)

	ch, subCancel := bus.Subscribe(ctx)
	defer subCancel()

	// 1. Publish first event
	err = bus.Publish(ctx, MyEvent{ID: 1})
	require.NoError(t, err)

	select {
	case ev := <-ch:
		assert.Equal(t, 1, ev.ID)
	case <-time.After(2 * time.Second):
		t.Fatal("initial event timeout")
	}

	// 2. Force cursor death by dropping the collection
	err = db.Collection(topic).Drop(ctx)
	require.NoError(t, err)

	// 3. Re-create the collection (simulated by NewTailablePubSub)
	_, err = gmqb.NewTailablePubSub[MyEvent](db, topic, gmqb.DefaultCappedOpts)
	require.NoError(t, err)

	// 4. Publish second event — it should be picked up after reconnect
	// The backoff starts at 1s, so we wait a bit
	time.Sleep(2 * time.Second)
	err = bus.Publish(ctx, MyEvent{ID: 2})
	require.NoError(t, err)

	select {
	case ev := <-ch:
		assert.Equal(t, 2, ev.ID)
	case <-time.After(10 * time.Second):
		t.Fatal("reconnected event timeout")
	}
}

func ExampleTailablePubSub() {
	// 1. Initialize PubSub for a specific event type.
	// This idempotently creates a capped collection.
	type UserEvent struct {
		UserID string `bson:"user_id"`
		Type   string `bson:"type"`
	}

	// Assuming 'db' is a *mongo.Database
	var db *mongo.Database
	bus, _ := gmqb.NewTailablePubSub[UserEvent](db, "user_events", gmqb.DefaultCappedOpts)

	// 2. Subscribe to events.
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	ch, stop := bus.Subscribe(ctx)
	defer stop()

	go func() {
		for event := range ch {
			fmt.Printf("Received event: %s for user %s\n", event.Type, event.UserID)
		}
	}()

	// 3. Publish an event.
	_ = bus.Publish(ctx, UserEvent{UserID: "123", Type: "signup"})
}
