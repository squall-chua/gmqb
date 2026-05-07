package rest

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/squall-chua/gmqb"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tryvium-travels/memongo"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

type User struct {
	ID    bson.ObjectID `json:"id" bson:"_id,omitempty"`
	Name  string        `json:"name" bson:"name"`
	Age   int           `json:"age" bson:"age"`
	Email string        `json:"email" bson:"email"`
}

var (
	testDB   *mongo.Database
	mongoSvr *memongo.Server
)

func TestMain(m *testing.M) {
	opts := &memongo.Options{
		MongoVersion: "6.0.5",
	}
	server, err := memongo.StartWithOptions(opts)
	if err != nil {
		fmt.Printf("failed to start memongo: %v", err)
		os.Exit(1)
	}
	mongoSvr = server
	defer mongoSvr.Stop()

	client, err := mongo.Connect(options.Client().ApplyURI(mongoSvr.URI()))
	if err != nil {
		fmt.Printf("failed to connect to mongo: %v", err)
		os.Exit(1)
	}
	testDB = client.Database("test")

	os.Exit(m.Run())
}

func TestResource(t *testing.T) {
	ctx := context.Background()
	rawColl := testDB.Collection("users_" + fmt.Sprint(time.Now().UnixNano()))
	coll := gmqb.Wrap[User](rawColl)

	cfg := Config[User, bson.ObjectID]{
		IDField: "_id",
		IDParser: func(s string) (bson.ObjectID, error) {
			return bson.ObjectIDFromHex(s)
		},
		FilterableFields: []FilterField{
			{Name: "name", BsonKey: "name", Op: OpEq},
			{Name: "age", BsonKey: "age", Op: OpGte | OpLte},
		},
		SortableFields: []string{"name", "age"},
		DefaultLimit:   10,
	}
	res := NewResource(coll, cfg)

	// --- Helper: Seed Data ---
	seedUsers := []User{
		{Name: "Alice", Age: 30, Email: "alice@example.com"},
		{Name: "Bob", Age: 25, Email: "bob@example.com"},
		{Name: "Charlie", Age: 35, Email: "charlie@example.com"},
	}
	for i := range seedUsers {
		_, _ = coll.InsertOne(ctx, &seedUsers[i])
		// Get generated ID if using strings usually we'd have it, 
		// but here we just let mongo generate and we'll fetch them.
		// For simplicity in tests, let's just use the seeded ones if we set IDs.
	}
	// Fetch all to get IDs
	allUsers, err := coll.Find(ctx, gmqb.NewFilter())
	require.NoError(t, err)
	require.NotEmpty(t, allUsers)
	aliceID := allUsers[0].ID

	t.Run("Create", func(t *testing.T) {
		newUser := User{Name: "David", Age: 28, Email: "david@example.com"}
		body, _ := json.Marshal(newUser)
		req := httptest.NewRequest(http.MethodPost, "/", bytes.NewBuffer(body))
		rec := httptest.NewRecorder()

		res.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusCreated, rec.Code)
		var envelope Envelope
		json.Unmarshal(rec.Body.Bytes(), &envelope)
		data := envelope.Data.(map[string]interface{})
		assert.Equal(t, "David", data["name"])
		assert.NotEmpty(t, rec.Header().Get("Location"))
	})

	t.Run("List", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/?limit=2", nil)
		rec := httptest.NewRecorder()

		res.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
		var envelope Envelope
		json.Unmarshal(rec.Body.Bytes(), &envelope)
		items := envelope.Data.([]interface{})
		assert.Len(t, items, 2)
		assert.Equal(t, int64(2), envelope.Meta.Limit)
	})

	t.Run("Read", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/"+aliceID.Hex(), nil)
		// Set PathValue manually for testing without a mux
		req.SetPathValue("id", aliceID.Hex())
		rec := httptest.NewRecorder()

		res.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
		var envelope Envelope
		json.Unmarshal(rec.Body.Bytes(), &envelope)
		data := envelope.Data.(map[string]interface{})
		assert.Equal(t, "Alice", data["name"])
	})

	t.Run("Read NotFound", func(t *testing.T) {
		nonExistentID := bson.NewObjectID().Hex()
		req := httptest.NewRequest(http.MethodGet, "/"+nonExistentID, nil)
		req.SetPathValue("id", nonExistentID)
		rec := httptest.NewRecorder()

		res.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusNotFound, rec.Code)
	})

	t.Run("Patch", func(t *testing.T) {
		patch := map[string]interface{}{"age": 31}
		body, _ := json.Marshal(patch)
		req := httptest.NewRequest(http.MethodPatch, "/"+aliceID.Hex(), bytes.NewBuffer(body))
		req.SetPathValue("id", aliceID.Hex())
		rec := httptest.NewRecorder()

		res.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
		var envelope Envelope
		json.Unmarshal(rec.Body.Bytes(), &envelope)
		data := envelope.Data.(map[string]interface{})
		assert.Equal(t, float64(31), data["age"])
	})

	t.Run("Delete", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodDelete, "/"+aliceID.Hex(), nil)
		req.SetPathValue("id", aliceID.Hex())
		rec := httptest.NewRecorder()

		res.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusNoContent, rec.Code)

		// Verify deleted
		_, err := coll.FindOne(ctx, gmqb.Eq("_id", aliceID))
		assert.ErrorIs(t, err, mongo.ErrNoDocuments)
	})
}
