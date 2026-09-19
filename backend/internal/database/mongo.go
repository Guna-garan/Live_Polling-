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
		log.Printf("warning: users index creation: %v", err)
	}

	pollsIdx := mongo.IndexModel{
		Keys: bson.D{{Key: "creatorId", Value: 1}},
	}
	if _, err := db.Collection("polls").Indexes().CreateOne(ctx, pollsIdx); err != nil {
		log.Printf("warning: polls index creation: %v", err)
	}

	votesCol := db.Collection("votes")

	votesPollIdx := mongo.IndexModel{
		Keys: bson.D{{Key: "pollId", Value: 1}},
	}
	votesUniqueIdx := mongo.IndexModel{
		Keys:    bson.D{{Key: "pollId", Value: 1}, {Key: "voterId", Value: 1}},
		Options: options.Index().SetUnique(true),
	}
	if err := reconcileVotesIndexes(ctx, votesCol, votesPollIdx, votesUniqueIdx); err != nil {
		return err
	}
	return nil
}

// reconcileVotesIndexes makes the votes collection's indexes match exactly
// what the application needs, dropping anything else first.
//
// This collection previously used different field names (an older schema
// had "poll_id"/"voter_key" instead of today's "pollId"/"voterId"). A
// unique index left over from that schema is actively dangerous: today's
// documents don't have that old field at all, so MongoDB treats it as
// null on every insert — and a unique index only allows ONE document with
// a null value for that field in the whole collection. The result is that
// the very first vote ever inserted succeeds, and every vote after that,
// from any voter on any poll, fails as a duplicate key error, which the
// application then (wrongly) reports as "you've already voted".
//
// Rather than guess the exact name of a stale index (which silently does
// nothing if the guess is wrong), this lists every index that actually
// exists and drops anything that isn't one of the two we want, by
// comparing key specs — not names. This is safe to run on every startup.
func reconcileVotesIndexes(ctx context.Context, col *mongo.Collection, wanted ...mongo.IndexModel) error {
	wantedKeys := make([]bson.D, len(wanted))
	for i, w := range wanted {
		wantedKeys[i] = w.Keys.(bson.D)
	}

	cursor, err := col.Indexes().List(ctx)
	if err != nil {
		return err
	}
	var existing []bson.M
	if err := cursor.All(ctx, &existing); err != nil {
		return err
	}

	for _, idx := range existing {
		name, _ := idx["name"].(string)
		if name == "_id_" || name == "" {
			continue // never touch the default _id index
		}
		keyDoc, _ := idx["key"].(bson.M)
		if indexKeysMatchAny(keyDoc, wantedKeys) {
			continue // already exactly what we want; leave it alone
		}
		log.Printf("mongo: dropping stale votes index %q (does not match current schema)", name)
		if _, err := col.Indexes().DropOne(ctx, name); err != nil {
			log.Printf("mongo: WARNING failed to drop stale index %q: %v — duplicate-vote detection may misbehave until this is removed manually", name, err)
		}
	}

	_, err = col.Indexes().CreateMany(ctx, wanted)
	return err
}

func indexKeysMatchAny(existing bson.M, wanted []bson.D) bool {
	for _, w := range wanted {
		if len(existing) != len(w) {
			continue
		}
		match := true
		for _, field := range w {
			v, ok := existing[field.Key]
			if !ok {
				match = false
				break
			}
			// Compare as float64 since BSON numeric types decode inconsistently.
			existingVal, _ := toFloat64(v)
			wantedVal, _ := toFloat64(field.Value)
			if existingVal != wantedVal {
				match = false
				break
			}
		}
		if match {
			return true
		}
	}
	return false
}

func toFloat64(v interface{}) (float64, bool) {
	switch n := v.(type) {
	case int32:
		return float64(n), true
	case int64:
		return float64(n), true
	case float64:
		return n, true
	case int:
		return float64(n), true
	default:
		return 0, false
	}
}
