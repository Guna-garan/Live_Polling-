package vote

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

// ErrDuplicateVote is returned when the unique (pollId, voterId) index
// rejects an insert. This is the backend's authoritative duplicate-vote
// guard — it holds even if two requests from the same voter race each
// other, which purely application-level checks cannot guarantee.
var ErrDuplicateVote = errDuplicateVote{}

type errDuplicateVote struct{}

func (errDuplicateVote) Error() string { return "you have already voted in this poll" }

type Repository struct {
	col *mongo.Collection
}

func NewRepository(db *mongo.Database) *Repository {
	return &Repository{col: db.Collection("votes")}
}

// Insert records a vote. Relies on the unique (pollId, voterId) index
// created in database.ensureIndexes to atomically reject duplicates at
// the database level — the check-then-insert race is closed by the
// index, not by application logic.
func (r *Repository) Insert(ctx context.Context, v *Vote) error {
	v.CreatedAt = time.Now().UTC()
	res, err := r.col.InsertOne(ctx, v)
	if err != nil {
		if mongo.IsDuplicateKeyError(err) {
			return ErrDuplicateVote
		}
		return err
	}
	v.ID = res.InsertedID.(primitive.ObjectID)
	return nil
}

// CountByOption aggregates persisted votes for a poll, grouped by
// option. Used to rebuild Redis counters from the MongoDB source of
// truth (startup recovery and lazy rebuild-on-read).
func (r *Repository) CountByOption(ctx context.Context, pollID primitive.ObjectID) (map[string]int64, error) {
	pipeline := mongo.Pipeline{
		{{Key: "$match", Value: bson.M{"pollId": pollID}}},
		{{Key: "$group", Value: bson.M{"_id": "$optionId", "count": bson.M{"$sum": 1}}}},
	}
	cur, err := r.col.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, err
	}
	defer cur.Close(ctx)

	out := make(map[string]int64)
	var row struct {
		ID    string `bson:"_id"`
		Count int64  `bson:"count"`
	}
	for cur.Next(ctx) {
		if err := cur.Decode(&row); err != nil {
			return nil, err
		}
		out[row.ID] = row.Count
	}
	return out, nil
}

// HasVoted checks whether a voter has already voted in a poll. Used to
// give the frontend an accurate "you've already voted" state on load,
// separate from the authoritative insert-time duplicate check.
func (r *Repository) HasVoted(ctx context.Context, pollID primitive.ObjectID, voterID string) (bool, error) {
	count, err := r.col.CountDocuments(ctx, bson.M{"pollId": pollID, "voterId": voterID})
	return count > 0, err
}
