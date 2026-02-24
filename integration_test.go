package gmqb_test

import (
	"context"
	"log"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	mongooptions "go.mongodb.org/mongo-driver/v2/mongo/options"

	"github.com/tryvium-travels/memongo"

	"github.com/squall-chua/gmqb"
)

// --- Test fixtures ---

type User struct {
	ID      bson.ObjectID `bson:"_id,omitempty"`
	Name    string        `bson:"name"`
	Age     int           `bson:"age"`
	Email   string        `bson:"email"`
	Country string        `bson:"country"`
	Active  bool          `bson:"active"`
	Tags    []string      `bson:"tags,omitempty"`
}

// Shared across all integration tests.
var (
	mongoServer *memongo.Server
	testClient  *mongo.Client
	testDB      *mongo.Database
)

func TestMain(m *testing.M) {
	var err error
	mongoServer, err = memongo.StartWithOptions(&memongo.Options{
		MongoVersion: "8.2.5",
	})
	if err != nil {
		log.Fatalf("memongo start: %v", err)
	}

	dbName := memongo.RandomDatabase()
	clientOpts := mongooptions.Client().ApplyURI(mongoServer.URI())
	testClient, err = mongo.Connect(clientOpts)
	if err != nil {
		log.Fatalf("mongo connect: %v", err)
	}

	testDB = testClient.Database(dbName)

	code := m.Run()

	_ = testClient.Disconnect(context.Background())
	mongoServer.Stop()
	os.Exit(code)
}

// freshCollection drops and returns a clean collection per test.
func freshCollection(t *testing.T) *gmqb.Collection[User] {
	t.Helper()
	coll := testDB.Collection(t.Name())
	_ = coll.Drop(context.Background())
	return gmqb.Wrap[User](coll)
}

// seedUsers inserts a standard set of test users.
func seedUsers(t *testing.T, coll *gmqb.Collection[User]) {
	t.Helper()
	ctx := context.Background()
	users := []User{
		{Name: "Alice", Age: 30, Email: "alice@example.com", Country: "US", Active: true, Tags: []string{"admin", "dev"}},
		{Name: "Bob", Age: 25, Email: "bob@example.com", Country: "UK", Active: true, Tags: []string{"dev"}},
		{Name: "Charlie", Age: 35, Email: "charlie@example.com", Country: "US", Active: false, Tags: []string{"ops"}},
		{Name: "Diana", Age: 28, Email: "diana@example.com", Country: "DE", Active: true, Tags: []string{"dev", "ops"}},
		{Name: "Eve", Age: 22, Email: "eve@example.com", Country: "UK", Active: false},
	}
	_, err := coll.InsertMany(ctx, users)
	require.NoError(t, err, "seed users")
}

// --- CRUD Tests ---

func TestIntegration_InsertOne_FindOne(t *testing.T) {
	coll := freshCollection(t)
	ctx := context.Background()

	user := &User{Name: "Zara", Age: 40, Email: "zara@test.com", Country: "FR", Active: true}
	_, err := coll.InsertOne(ctx, user)
	require.NoError(t, err)

	found, err := coll.FindOne(ctx, gmqb.Eq("name", "Zara"))
	require.NoError(t, err)
	assert.Equal(t, "Zara", found.Name)
	assert.Equal(t, 40, found.Age)
	assert.Equal(t, "FR", found.Country)
}

func TestIntegration_InsertMany_Find(t *testing.T) {
	coll := freshCollection(t)
	ctx := context.Background()
	seedUsers(t, coll)

	users, err := coll.Find(ctx, gmqb.Eq("active", true))
	require.NoError(t, err)
	assert.Len(t, users, 3)
}

func TestIntegration_Find_WithSortAndLimit(t *testing.T) {
	coll := freshCollection(t)
	ctx := context.Background()
	seedUsers(t, coll)

	users, err := coll.Find(ctx, gmqb.Eq("active", true),
		gmqb.WithSort(gmqb.Asc("age")),
		gmqb.WithLimit(2),
	)
	require.NoError(t, err)
	require.Len(t, users, 2)
	assert.Equal(t, "Bob", users[0].Name)
	assert.Equal(t, "Diana", users[1].Name)
}

func TestIntegration_Find_ComparisonOperators(t *testing.T) {
	coll := freshCollection(t)
	ctx := context.Background()
	seedUsers(t, coll)

	// $gte
	users, err := coll.Find(ctx, gmqb.Gte("age", 30))
	require.NoError(t, err)
	assert.Len(t, users, 2, "$gte 30")

	// $lt
	users, err = coll.Find(ctx, gmqb.Lt("age", 25))
	require.NoError(t, err)
	assert.Len(t, users, 1, "$lt 25")

	// $ne
	users, err = coll.Find(ctx, gmqb.Ne("country", "US"))
	require.NoError(t, err)
	assert.Len(t, users, 3, "$ne US")

	// $in
	users, err = coll.Find(ctx, gmqb.In("country", "US", "UK"))
	require.NoError(t, err)
	assert.Len(t, users, 4, "$in [US,UK]")
}

func TestIntegration_Find_LogicalOperators(t *testing.T) {
	coll := freshCollection(t)
	ctx := context.Background()
	seedUsers(t, coll)

	// $and: active AND country=US
	users, err := coll.Find(ctx, gmqb.And(
		gmqb.Eq("active", true),
		gmqb.Eq("country", "US"),
	))
	require.NoError(t, err)
	require.Len(t, users, 1)
	assert.Equal(t, "Alice", users[0].Name)

	// $or: country=UK OR country=DE
	users, err = coll.Find(ctx, gmqb.Or(
		gmqb.Eq("country", "UK"),
		gmqb.Eq("country", "DE"),
	))
	require.NoError(t, err)
	assert.Len(t, users, 3)
}

func TestIntegration_Find_ElementOperators(t *testing.T) {
	coll := freshCollection(t)
	ctx := context.Background()
	seedUsers(t, coll)

	// Eve has no tags (omitempty → field absent)
	users, err := coll.Find(ctx, gmqb.Exists("tags", true))
	require.NoError(t, err)
	assert.Len(t, users, 4)
}

func TestIntegration_Find_ArrayOperators(t *testing.T) {
	coll := freshCollection(t)
	ctx := context.Background()
	seedUsers(t, coll)

	// $all: users with both "dev" and "ops" tags
	users, err := coll.Find(ctx, gmqb.All("tags", "dev", "ops"))
	require.NoError(t, err)
	require.Len(t, users, 1)
	assert.Equal(t, "Diana", users[0].Name)

	// $size: users with exactly 2 tags
	users, err = coll.Find(ctx, gmqb.Size("tags", 2))
	require.NoError(t, err)
	assert.Len(t, users, 2, "$size 2 (Alice, Diana)")
}

func TestIntegration_Find_RegexOperator(t *testing.T) {
	coll := freshCollection(t)
	ctx := context.Background()
	seedUsers(t, coll)

	users, err := coll.Find(ctx, gmqb.Regex("email", "^(alice|bob)", "i"))
	require.NoError(t, err)
	assert.Len(t, users, 2)
}

func TestIntegration_UpdateOne(t *testing.T) {
	coll := freshCollection(t)
	ctx := context.Background()
	seedUsers(t, coll)

	result, err := coll.UpdateOne(ctx,
		gmqb.Eq("name", "Alice"),
		gmqb.NewUpdate().Set("age", 31).Set("email", "alice-updated@example.com"),
	)
	require.NoError(t, err)
	assert.Equal(t, int64(1), result.ModifiedCount)

	found, err := coll.FindOne(ctx, gmqb.Eq("name", "Alice"))
	require.NoError(t, err)
	assert.Equal(t, 31, found.Age)
	assert.Equal(t, "alice-updated@example.com", found.Email)
}

func TestIntegration_UpdateMany_Inc(t *testing.T) {
	coll := freshCollection(t)
	ctx := context.Background()
	seedUsers(t, coll)

	result, err := coll.UpdateMany(ctx,
		gmqb.Eq("country", "UK"),
		gmqb.NewUpdate().Inc("age", 1),
	)
	require.NoError(t, err)
	assert.Equal(t, int64(2), result.ModifiedCount)

	bob, err := coll.FindOne(ctx, gmqb.Eq("name", "Bob"))
	require.NoError(t, err)
	assert.Equal(t, 26, bob.Age)

	eve, err := coll.FindOne(ctx, gmqb.Eq("name", "Eve"))
	require.NoError(t, err)
	assert.Equal(t, 23, eve.Age)
}

func TestIntegration_UpdateOne_Upsert(t *testing.T) {
	coll := freshCollection(t)
	ctx := context.Background()

	result, err := coll.UpdateOne(ctx,
		gmqb.Eq("name", "NewUser"),
		gmqb.NewUpdate().Set("name", "NewUser").Set("age", 18).Set("country", "JP"),
		gmqb.WithUpsert(true),
	)
	require.NoError(t, err)
	assert.Equal(t, int64(1), result.UpsertedCount)

	found, err := coll.FindOne(ctx, gmqb.Eq("name", "NewUser"))
	require.NoError(t, err)
	assert.Equal(t, 18, found.Age)
	assert.Equal(t, "JP", found.Country)
}

func TestIntegration_UpdateOne_Push(t *testing.T) {
	coll := freshCollection(t)
	ctx := context.Background()
	seedUsers(t, coll)

	_, err := coll.UpdateOne(ctx,
		gmqb.Eq("name", "Bob"),
		gmqb.NewUpdate().Push("tags", "lead"),
	)
	require.NoError(t, err)

	bob, err := coll.FindOne(ctx, gmqb.Eq("name", "Bob"))
	require.NoError(t, err)
	assert.Equal(t, []string{"dev", "lead"}, bob.Tags)
}

func TestIntegration_DeleteOne(t *testing.T) {
	coll := freshCollection(t)
	ctx := context.Background()
	seedUsers(t, coll)

	result, err := coll.DeleteOne(ctx, gmqb.Eq("name", "Charlie"))
	require.NoError(t, err)
	assert.Equal(t, int64(1), result.DeletedCount)

	count, err := coll.CountDocuments(ctx, gmqb.Raw(bson.D{}))
	require.NoError(t, err)
	assert.Equal(t, int64(4), count)
}

func TestIntegration_DeleteMany(t *testing.T) {
	coll := freshCollection(t)
	ctx := context.Background()
	seedUsers(t, coll)

	result, err := coll.DeleteMany(ctx, gmqb.Eq("active", false))
	require.NoError(t, err)
	assert.Equal(t, int64(2), result.DeletedCount)

	count, err := coll.CountDocuments(ctx, gmqb.Raw(bson.D{}))
	require.NoError(t, err)
	assert.Equal(t, int64(3), count)
}

func TestIntegration_CountDocuments(t *testing.T) {
	coll := freshCollection(t)
	ctx := context.Background()
	seedUsers(t, coll)

	count, err := coll.CountDocuments(ctx, gmqb.Eq("country", "US"))
	require.NoError(t, err)
	assert.Equal(t, int64(2), count)

	count, err = coll.CountDocuments(ctx, gmqb.Raw(bson.D{}))
	require.NoError(t, err)
	assert.Equal(t, int64(5), count)
}

func TestIntegration_Distinct(t *testing.T) {
	coll := freshCollection(t)
	ctx := context.Background()
	seedUsers(t, coll)

	result := coll.Distinct(ctx, "country", gmqb.Raw(bson.D{}))
	var countries []string
	require.NoError(t, result.Decode(&countries))
	assert.Len(t, countries, 3)
}

func TestIntegration_Aggregate_GroupBy(t *testing.T) {
	coll := freshCollection(t)
	ctx := context.Background()
	seedUsers(t, coll)

	type CountryStats struct {
		Country string `bson:"_id"`
		Count   int    `bson:"count"`
	}

	pipeline := gmqb.NewPipeline().
		Group(gmqb.GroupSpec("$country",
			gmqb.GroupAcc("count", gmqb.AccSum(1)),
		)).
		Sort(gmqb.Desc("count"))

	stats, err := gmqb.Aggregate[CountryStats](coll, ctx, pipeline)
	require.NoError(t, err)
	require.Len(t, stats, 3)
	assert.Equal(t, 2, stats[0].Count)
	assert.Equal(t, 1, stats[2].Count)
	assert.Equal(t, "DE", stats[2].Country)
}

func TestIntegration_Aggregate_MatchAndProject(t *testing.T) {
	coll := freshCollection(t)
	ctx := context.Background()
	seedUsers(t, coll)

	type NameOnly struct {
		Name string `bson:"name"`
	}

	pipeline := gmqb.NewPipeline().
		Match(gmqb.Eq("active", true)).
		Project(bson.D{
			{Key: "name", Value: 1},
			{Key: "_id", Value: 0},
		}).
		Sort(gmqb.Asc("name"))

	names, err := gmqb.Aggregate[NameOnly](coll, ctx, pipeline)
	require.NoError(t, err)
	require.Len(t, names, 3)
	assert.Equal(t, "Alice", names[0].Name)
	assert.Equal(t, "Bob", names[1].Name)
	assert.Equal(t, "Diana", names[2].Name)
}

func TestIntegration_Aggregate_Unwind(t *testing.T) {
	coll := freshCollection(t)
	ctx := context.Background()
	seedUsers(t, coll)

	type TagCount struct {
		Tag   string `bson:"_id"`
		Count int    `bson:"count"`
	}

	pipeline := gmqb.NewPipeline().
		Unwind("$tags").
		Group(gmqb.GroupSpec("$tags",
			gmqb.GroupAcc("count", gmqb.AccSum(1)),
		)).
		Sort(gmqb.Desc("count"))

	tags, err := gmqb.Aggregate[TagCount](coll, ctx, pipeline)
	require.NoError(t, err)
	require.Len(t, tags, 3)
	assert.Equal(t, "dev", tags[0].Tag)
	assert.Equal(t, 3, tags[0].Count)
}

func TestIntegration_Aggregate_AddFieldsAndLimit(t *testing.T) {
	coll := freshCollection(t)
	ctx := context.Background()
	seedUsers(t, coll)

	type UserWithLabel struct {
		Name     string `bson:"name"`
		AgeGroup string `bson:"ageGroup"`
	}

	pipeline := gmqb.NewPipeline().
		AddFields(gmqb.AddFieldsSpec(
			gmqb.AddField("ageGroup", gmqb.ExprCond(
				gmqb.ExprGte("$age", 30),
				"senior",
				"junior",
			)),
		)).
		Project(bson.D{
			{Key: "name", Value: 1},
			{Key: "ageGroup", Value: 1},
			{Key: "_id", Value: 0},
		}).
		Sort(gmqb.Asc("name")).
		Limit(3)

	results, err := gmqb.Aggregate[UserWithLabel](coll, ctx, pipeline)
	require.NoError(t, err)
	require.Len(t, results, 3)
	assert.Equal(t, "senior", results[0].AgeGroup, "Alice(30)=senior")
	assert.Equal(t, "junior", results[1].AgeGroup, "Bob(25)=junior")
}

func TestIntegration_FindOne_NotFound(t *testing.T) {
	coll := freshCollection(t)
	ctx := context.Background()

	_, err := coll.FindOne(ctx, gmqb.Eq("name", "NonExistent"))
	assert.ErrorIs(t, err, mongo.ErrNoDocuments)
}

func TestIntegration_Find_WithProjection(t *testing.T) {
	coll := freshCollection(t)
	ctx := context.Background()
	seedUsers(t, coll)

	users, err := coll.Find(ctx, gmqb.Eq("name", "Alice"),
		gmqb.WithProjection(gmqb.Include("name", "age")),
	)
	require.NoError(t, err)
	require.Len(t, users, 1)
	assert.Equal(t, "Alice", users[0].Name)
	assert.Empty(t, users[0].Email, "non-projected field should be zero")
	assert.Empty(t, users[0].Country, "non-projected field should be zero")
}

func TestIntegration_Find_WithSkip(t *testing.T) {
	coll := freshCollection(t)
	ctx := context.Background()
	seedUsers(t, coll)

	users, err := coll.Find(ctx, gmqb.Raw(bson.D{}),
		gmqb.WithSort(gmqb.Asc("name")),
		gmqb.WithSkip(3),
	)
	require.NoError(t, err)
	require.Len(t, users, 2)
	assert.Equal(t, "Diana", users[0].Name)
	assert.Equal(t, "Eve", users[1].Name)
}

func TestIntegration_FindOneAndDelete(t *testing.T) {
	coll := freshCollection(t)
	ctx := context.Background()
	seedUsers(t, coll)

	deleted, err := coll.FindOneAndDelete(ctx, gmqb.Eq("name", "Bob"))
	require.NoError(t, err)
	assert.Equal(t, "Bob", deleted.Name)

	// Verify it's actually deleted
	count, err := coll.CountDocuments(ctx, gmqb.Eq("name", "Bob"))
	require.NoError(t, err)
	assert.Equal(t, int64(0), count)

	// Delete non-existent
	_, err = coll.FindOneAndDelete(ctx, gmqb.Eq("name", "NonExistent"))
	assert.ErrorIs(t, err, mongo.ErrNoDocuments)
}

func TestIntegration_FindOneAndUpdate(t *testing.T) {
	coll := freshCollection(t)
	ctx := context.Background()
	seedUsers(t, coll)

	// By default, returns the document BEFORE the update
	before, err := coll.FindOneAndUpdate(ctx,
		gmqb.Eq("name", "Alice"),
		gmqb.NewUpdate().Set("age", 31),
	)
	require.NoError(t, err)
	assert.Equal(t, 30, before.Age) // original age

	// With options.After, returns the document AFTER the update
	after, err := coll.FindOneAndUpdate(ctx,
		gmqb.Eq("name", "Alice"),
		gmqb.NewUpdate().Set("age", 32),
		gmqb.WithReturnDocument(mongooptions.After),
	)
	require.NoError(t, err)
	assert.Equal(t, 32, after.Age) // updated age

	// Update non-existent
	_, err = coll.FindOneAndUpdate(ctx,
		gmqb.Eq("name", "NonExistent"),
		gmqb.NewUpdate().Set("age", 100),
	)
	assert.ErrorIs(t, err, mongo.ErrNoDocuments)
}

func TestIntegration_FindOneAndReplace(t *testing.T) {
	coll := freshCollection(t)
	ctx := context.Background()
	seedUsers(t, coll)

	replacement := &User{Name: "Diana Replaced", Age: 99, Email: "diana-r@example.com", Country: "DK", Active: true}

	// Returns the document BEFORE the replace by default
	before, err := coll.FindOneAndReplace(ctx,
		gmqb.Eq("name", "Diana"),
		replacement,
	)
	require.NoError(t, err)
	assert.Equal(t, "Diana", before.Name)
	assert.Equal(t, 28, before.Age)

	// Replace again with ReturnDocument = After
	replacement2 := &User{Name: "Diana Replaced 2", Age: 100, Email: "diana-r2@example.com", Country: "DK", Active: true}
	after, err := coll.FindOneAndReplace(ctx,
		gmqb.Eq("name", "Diana Replaced"),
		replacement2,
		gmqb.WithReturnDocumentReplace(mongooptions.After),
	)
	require.NoError(t, err)
	assert.Equal(t, "Diana Replaced 2", after.Name)
	assert.Equal(t, 100, after.Age)

	// Replace non-existent
	_, err = coll.FindOneAndReplace(ctx,
		gmqb.Eq("name", "NonExistent"),
		&User{Name: "Ghost", Age: 0},
	)
	assert.ErrorIs(t, err, mongo.ErrNoDocuments)
}

// --- Filter Chaining Tests ---

func TestIntegration_FilterChain_Basic(t *testing.T) {
	coll := freshCollection(t)
	ctx := context.Background()
	seedUsers(t, coll)

	// active=true AND age>=25 AND age<35 → Bob(25), Diana(28), Alice(30)
	users, err := coll.Find(ctx,
		gmqb.NewFilter().
			Eq("active", true).
			Gte("age", 25).
			Lt("age", 35),
		gmqb.WithSort(gmqb.Asc("age")),
	)
	require.NoError(t, err)
	require.Len(t, users, 3)
	assert.Equal(t, "Bob", users[0].Name)
	assert.Equal(t, "Diana", users[1].Name)
	assert.Equal(t, "Alice", users[2].Name)
}

func TestIntegration_FilterChain_WithOptions(t *testing.T) {
	coll := freshCollection(t)
	ctx := context.Background()
	seedUsers(t, coll)

	// country=US AND active=true, sorted by name, limit 1
	users, err := coll.Find(ctx,
		gmqb.NewFilter().
			Eq("country", "US").
			Eq("active", true),
		gmqb.WithSort(gmqb.Asc("name")),
		gmqb.WithLimit(1),
	)
	require.NoError(t, err)
	require.Len(t, users, 1)
	assert.Equal(t, "Alice", users[0].Name)
}

func TestIntegration_FilterChain_WithCount(t *testing.T) {
	coll := freshCollection(t)
	ctx := context.Background()
	seedUsers(t, coll)

	// country=UK AND active=false → Eve
	count, err := coll.CountDocuments(ctx,
		gmqb.NewFilter().
			Eq("country", "UK").
			Eq("active", false),
	)
	require.NoError(t, err)
	assert.Equal(t, int64(1), count)
}

func TestIntegration_FilterChain_RegexAndExists(t *testing.T) {
	coll := freshCollection(t)
	ctx := context.Background()
	seedUsers(t, coll)

	// email matches ^alice AND has tags field
	users, err := coll.Find(ctx,
		gmqb.NewFilter().
			Regex("email", "^alice", "i").
			Exists("tags", true),
	)
	require.NoError(t, err)
	require.Len(t, users, 1)
	assert.Equal(t, "Alice", users[0].Name)
}

func TestIntegration_BulkWrite(t *testing.T) {
	coll := freshCollection(t)
	ctx := context.Background()
	seedUsers(t, coll)

	models := []gmqb.WriteModel[User]{
		gmqb.NewInsertOneModel[User]().SetDocument(&User{Name: "Zack", Age: 20, Country: "US", Active: true}),
		gmqb.NewUpdateOneModel[User]().
			SetFilter(gmqb.Eq("name", "Alice")).
			SetUpdate(gmqb.NewUpdate().Set("age", 31)),
		gmqb.NewDeleteOneModel[User]().
			SetFilter(gmqb.Eq("name", "Charlie")),
		gmqb.NewReplaceOneModel[User]().
			SetFilter(gmqb.Eq("name", "Eve")).
			SetReplacement(&User{Name: "Eve Replaced", Age: 99, Country: "UK", Active: true}),
	}

	res, err := coll.BulkWrite(ctx, models)
	require.NoError(t, err)

	assert.Equal(t, int64(1), res.InsertedCount)
	assert.Equal(t, int64(2), res.ModifiedCount)
	assert.Equal(t, int64(1), res.DeletedCount)

	count, err := coll.CountDocuments(ctx, gmqb.Raw(bson.D{}))
	require.NoError(t, err)
	assert.Equal(t, int64(5), count)
}
