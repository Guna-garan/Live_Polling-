package database

import (
	"context"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// ConnectMongo dials MongoDB and returns a ready-to-use *mongo.Database.
// It pings the server before returning so that startup fails fast if the
// database is unreachable, rather than surfacing confusing errors later
// on the first request.
func ConnectMongo(uri, dbName string) (*mongo.Database, func(context.Context) error, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client, err := mongo.Connect(ctx, options.Client().ApplyURI(uri))
	if err != nil {
		return nil, nil, err
	}

	if err := client.Ping(ctx, nil); err != nil {
		return nil, nil, err
	}

	db := client.Database(dbName)
	if err := ensureIndexes(ctx, db); err != nil {
		return nil, nil, err
	}

	log.Println("mongodb: connected and indexes ensured")
	return db, client.Disconnect, nil
}

// ensureIndexes creates every index the application relies on for
// correctness (not just performance) — most importantly the unique
// (pollId, voterId) index on votes, which is the backend's last line of
// defense against duplicate voting even if application logic has a bug
// or two requests race each other.
func ensureIndexes(ctx context.Context, db *mongo.Database) error {
	usersIdx := mongo.IndexModel{
		Keys:    bson.D{{Key: "email", Value: 1}},
		Options: options.Index().SetUnique(true),
	}
	if _, err := db.Collection("users").Indexes().CreateOne(ctx, usersIdx); err != nil {
		return err
	}

	pollsIdx := mongo.IndexModel{
		Keys: bson.D{{Key: "creatorId", Value: 1}},
	}
	if _, err := db.Collection("polls").Indexes().CreateOne(ctx, pollsIdx); err != nil {
		return err
	}

	votesPollIdx := mongo.IndexModel{
		Keys: bson.D{{Key: "pollId", Value: 1}},
	}
	votesUniqueIdx := mongo.IndexModel{
		Keys:    bson.D{{Key: "pollId", Value: 1}, {Key: "voterId", Value: 1}},
		Options: options.Index().SetUnique(true),
	}
	_, err := db.Collection("votes").Indexes().CreateMany(ctx, []mongo.IndexModel{votesPollIdx, votesUniqueIdx})
	return err
}
