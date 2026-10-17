package submissions

import (
	"context"
	"errors"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// ErrNotFound is returned when a submission does not exist.
var ErrNotFound = errors.New("submission not found")

// Repository provides MongoDB persistence for submissions.
type Repository struct {
	coll *mongo.Collection
}

// NewRepository builds a Repository bound to the submissions collection.
func NewRepository(db *mongo.Database) *Repository {
	return &Repository{coll: db.Collection("submissions")}
}

// Create inserts a new submission.
func (r *Repository) Create(ctx context.Context, s *Submission) error {
	if s.ID.IsZero() {
		s.ID = primitive.NewObjectID()
	}
	_, err := r.coll.InsertOne(ctx, s)
	return err
}

// FindByID returns the submission with the given id.
func (r *Repository) FindByID(ctx context.Context, id primitive.ObjectID) (*Submission, error) {
	var s Submission
	err := r.coll.FindOne(ctx, bson.M{"_id": id}).Decode(&s)
	if errors.Is(err, mongo.ErrNoDocuments) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return &s, nil
}

// FindByOwner lists the submissions owned by the given user, newest first.
func (r *Repository) FindByOwner(ctx context.Context, ownerID primitive.ObjectID) ([]*Submission, error) {
	cur, err := r.coll.Find(
		ctx,
		bson.M{"ownerId": ownerID, "status": StatusActive},
		options.Find().SetSort(bson.D{{Key: "createdAt", Value: -1}}),
	)
	if err != nil {
		return nil, err
	}
	defer cur.Close(ctx)

	var out []*Submission
	if err := cur.All(ctx, &out); err != nil {
		return nil, err
	}
	return out, nil
}
