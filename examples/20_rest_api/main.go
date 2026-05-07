package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/squall-chua/gmqb"
	"github.com/squall-chua/gmqb/rest"
	"github.com/tryvium-travels/memongo"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

// User represents our database model
type User struct {
	ID        bson.ObjectID `bson:"_id,omitempty" json:"id"`
	Username  string        `bson:"username" json:"username"`
	Email     string        `bson:"email" json:"email"`
	Age       int           `bson:"age" json:"age"`
	Status    string        `bson:"status" json:"status"`
	CreatedAt time.Time     `bson:"createdAt" json:"createdAt"`
}

func main() {
	// 1. Start In-Memory MongoDB
	ctx := context.Background()
	mongoServer, err := memongo.StartWithOptions(&memongo.Options{
		MongoVersion: "8.2.5",
	})
	if err != nil {
		log.Fatalf("memongo start: %v", err)
	}
	defer mongoServer.Stop()

	// 2. Connect to In-Memory MongoDB
	client, err := mongo.Connect(options.Client().ApplyURI(mongoServer.URI()))
	if err != nil {
		log.Fatal(err)
	}
	defer client.Disconnect(ctx)

	db := client.Database("example_db")
	rawColl := db.Collection("users")

	// 3. Seed some initial data for testing
	seedUsers(ctx, rawColl)

	// 4. Wrap with gmqb generic collection
	coll := gmqb.Wrap[User](rawColl)

	// 5. Configure the REST resource
	// We use bson.ObjectID as the ID type.
	usersResource := rest.NewResource(coll, rest.Config[User, bson.ObjectID]{
		IDField: "_id",
		IDParser: func(s string) (bson.ObjectID, error) {
			return bson.ObjectIDFromHex(s)
		},

		// Allowlist for filtering via query params: ?age[gte]=18&status=active
		FilterableFields: []rest.FilterField{
			{Name: "username", BsonKey: "username", Op: rest.OpEq | rest.OpIn},
			{Name: "status", BsonKey: "status", Op: rest.OpEq},
			{Name: "age", BsonKey: "age", Op: rest.OpEq | rest.OpGt | rest.OpGte | rest.OpLt | rest.OpLte, ValueParser: func(s string) (any, error) {
				return strconv.Atoi(s)
			}},
		},

		// Allowlist for sorting: ?sort=createdAt,-age
		SortableFields: []string{"username", "age", "status", "createdAt"},
		DefaultSort:    bson.D{{Key: "createdAt", Value: -1}},

		// Pagination defaults
		DefaultLimit: 10,
		MaxLimit:     100,

		// Lifecycle hooks for business logic and auditing
		Hooks: rest.Hooks[User, bson.ObjectID]{
			BeforeCreate: func(ctx context.Context, doc *User) error {
				doc.CreatedAt = time.Now()
				if doc.Status == "" {
					doc.Status = "active"
				}
				return nil
			},
			AfterCreate: func(ctx context.Context, doc *User) {
				log.Printf("REST: Created user %s (%s)", doc.Username, doc.ID.Hex())
			},
			BeforeUpdate: func(ctx context.Context, id bson.ObjectID, patch map[string]any) error {
				// Example: Prohibit updating the username via PATCH
				if _, ok := patch["username"]; ok {
					return errors.New("username is immutable")
				}
				log.Printf("REST: Patching user %s: %v", id.Hex(), patch)
				return nil
			},
			AfterUpdate: func(ctx context.Context, id bson.ObjectID, patch map[string]any) {
				log.Printf("REST: Successfully updated user %s", id.Hex())
			},
			BeforeDelete: func(ctx context.Context, id bson.ObjectID) error {
				log.Printf("REST: Deleting user %s", id.Hex())
				return nil
			},
		},
	})

	// 6. Mount handlers
	// Note: We use Go 1.22+ routing features (PathValue)
	mux := http.NewServeMux()

	// Handle collection: GET /users (List), POST /users (Create)
	mux.Handle("GET /users", usersResource)
	mux.Handle("POST /users", usersResource)

	// Handle individual resource: GET/PUT/PATCH/DELETE /users/{id}
	mux.Handle("GET /users/{id}", usersResource)
	mux.Handle("PUT /users/{id}", usersResource)
	mux.Handle("PATCH /users/{id}", usersResource)
	mux.Handle("DELETE /users/{id}", usersResource)

	// 7. Start Server
	fmt.Println("REST API server starting on :8080...")
	fmt.Println("Try these endpoints:")
	fmt.Println("  List users:    GET  http://localhost:8080/users")
	fmt.Println("  Filter users:  GET  http://localhost:8080/users?age[gte]=18&status=active")
	fmt.Println("  Sort users:    GET  http://localhost:8080/users?sort=-createdAt,username")
	fmt.Println("  Paginate:      GET  http://localhost:8080/users?limit=5&offset=10")
	fmt.Println("  Cursor Paging: GET  http://localhost:8080/users?limit=5&cursor=XYZ...")
	fmt.Println("  Create user:   POST http://localhost:8080/users (Body: JSON)")
	fmt.Println("  Read user:     GET  http://localhost:8080/users/{id}")

	if err := http.ListenAndServe(":8080", mux); err != nil {
		log.Fatal(err)
	}
}

// seedUsers populates the collection with some test data
func seedUsers(ctx context.Context, coll *mongo.Collection) {
	users := []any{
		User{ID: bson.NewObjectID(), Username: "alice", Email: "alice@example.com", Age: 30, Status: "active", CreatedAt: time.Now()},
		User{ID: bson.NewObjectID(), Username: "bob", Email: "bob@example.com", Age: 25, Status: "active", CreatedAt: time.Now().Add(-time.Hour)},
		User{ID: bson.NewObjectID(), Username: "charlie", Email: "charlie@example.com", Age: 35, Status: "pending", CreatedAt: time.Now().Add(-2 * time.Hour)},
	}
	_, err := coll.InsertMany(ctx, users)
	if err != nil {
		log.Printf("failed to seed users: %v", err)
	}
}
