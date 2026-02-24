package main

import (
	"context"
	"fmt"
	"log"

	"github.com/squall-chua/gmqb"
	"github.com/tryvium-travels/memongo"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

type Session struct {
	UserID         string      `bson:"userId"`
	Status         string      `bson:"status"`
	LoginCount     int         `bson:"loginCount"`
	TemporaryToken string      `bson:"temporaryToken,omitempty"`
	LastModified   interface{} `bson:"lastModified,omitempty"`
}

func main() {
	// 1. Setup in-memory MongoDB
	mongoServer, err := memongo.StartWithOptions(&memongo.Options{
		MongoVersion: "8.2.5",
	})
	if err != nil {
		log.Fatalf("Failed to start memongo: %v", err)
	}
	defer mongoServer.Stop()

	client, err := mongo.Connect(options.Client().ApplyURI(mongoServer.URI()))
	if err != nil {
		log.Fatalf("Failed to connect: %v", err)
	}
	defer client.Disconnect(context.Background())

	// 2. Wrap collection
	db := client.Database("testdb")
	coll := gmqb.Wrap[Session](db.Collection("sessions"))
	ctx := context.Background()

	// 3. Seed some data
	_, _ = coll.InsertOne(ctx, &Session{
		UserID:         "user-1",
		Status:         "inactive",
		LoginCount:     5,
		TemporaryToken: "xyz123",
	})

	// 4. Define updates
	update := gmqb.NewUpdate().
		Set("status", "active").
		Inc("loginCount", 1).
		Unset("temporaryToken").
		CurrentDateAsTimestamp("lastModified")

	fmt.Println("Field Update JSON:")
	fmt.Println(update.JSON())

	// 5. Execute update
	_, err = coll.UpdateOne(ctx, gmqb.Eq("userId", "user-1"), update)
	if err != nil {
		log.Fatal(err)
	}

	session, _ := coll.FindOne(ctx, gmqb.Eq("userId", "user-1"))
	fmt.Printf("\nUpdated Session:\n%+v\n", session)
}
