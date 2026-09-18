package poll

import (
	"context"
	"errors"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var ErrPollNotFound = errors.New("poll not found")

type Repository struct {
	col *mongo.Collection
}

func NewRepository(db *mongo.Database) *Repository {
	return &Repository{col: db.Collection("polls")}
}

func (r *Repository) Create(ctx context.Context, p *Poll) error {
	p.CreatedAt = time.Now().UTC()
	p.UpdatedAt = p.CreatedAt
	res, err := r.col.InsertOne(ctx, p)
	if err != nil {
		return err
	}
	p.ID = res.InsertedID.(primitive.ObjectID)
	return nil
}

func (r *Repository) FindByID(ctx context.Context, id primitive.ObjectID) (*Poll, error) {
	var p Poll
	err := r.col.FindOne(ctx, bson.M{"_id": id}).Decode(&p)
	if errors.Is(err, mongo.ErrNoDocuments) {
		return nil, ErrPollNotFound
	}
	if err != nil {
		return nil, err
	}
	return &p, nil
}

func (r *Repository) FindByCreator(ctx context.Context, creatorID primitive.ObjectID) ([]Poll, error) {
	cur, err := r.col.Find(ctx, bson.M{"creatorId": creatorID}, options.Find().SetSort(bson.M{"createdAt": -1}))
	if err != nil {
		return nil, err
	}
	defer cur.Close(ctx)

	var polls []Poll
	if err := cur.All(ctx, &polls); err != nil {
		return nil, err
	}
	return polls, nil
}

// FindActive returns every poll currently marked active — used at
// startup to know which polls' Redis counters need to be rebuilt from
// MongoDB.
func (r *Repository) FindActive(ctx context.Context) ([]Poll, error) {
	cur, err := r.col.Find(ctx, bson.M{"status": StatusActive})
	if err != nil {
		return nil, err
	}
	defer cur.Close(ctx)

	var polls []Poll
	if err := cur.All(ctx, &polls); err != nil {
		return nil, err
	}
	return polls, nil
}

func (r *Repository) UpdateStatus(ctx context.Context, id, creatorID primitive.ObjectID, status string) (*Poll, error) {
	res := r.col.FindOneAndUpdate(
		ctx,
		bson.M{"_id": id, "creatorId": creatorID},
		bson.M{"$set": bson.M{"status": status, "updatedAt": time.Now().UTC()}},
		options.FindOneAndUpdate().SetReturnDocument(options.After),
	)

	var p Poll
	if err := res.Decode(&p); err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, ErrPollNotFound
		}
		return nil, err
	}
	return &p, nil
}

// MarkExpired closes a poll whose expiration has passed. Called lazily
// whenever a poll is read or voted on, so expiration is enforced by the
// backend regardless of whether any background job has run.
func (r *Repository) MarkExpired(ctx context.Context, id primitive.ObjectID) error {
	_, err := r.col.UpdateOne(ctx,
		bson.M{"_id": id},
		bson.M{"$set": bson.M{"status": StatusClosed, "updatedAt": time.Now().UTC()}},
	)
	return err
}

func (r *Repository) Delete(ctx context.Context, id, creatorID primitive.ObjectID) error {
	res, err := r.col.DeleteOne(ctx, bson.M{"_id": id, "creatorId": creatorID})
	if err != nil {
		return err
	}
	if res.DeletedCount == 0 {
		return ErrPollNotFound
	}
	return nil
}
