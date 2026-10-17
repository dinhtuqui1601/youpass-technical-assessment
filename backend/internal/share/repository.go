package share

import (
	"context"
	"errors"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// ErrNotFound is returned when a share link does not exist.
var ErrNotFound = errors.New("share link not found")

// ErrDuplicateCode signals a short-code collision.
var ErrDuplicateCode = errors.New("duplicate short code")

// Repository provides MongoDB persistence for share links.
type Repository struct {
	coll *mongo.Collection
}

// NewRepository builds a Repository bound to the share_links collection.
func NewRepository(db *mongo.Database) *Repository {
	return &Repository{coll: db.Collection("share_links")}
}

// Create inserts a new share link.
func (r *Repository) Create(ctx context.Context, s *ShareLink) error {
	if s.ID.IsZero() {
		s.ID = primitive.NewObjectID()
	}
	_, err := r.coll.InsertOne(ctx, s)
	if err != nil {
		if mongo.IsDuplicateKeyError(err) {
			return ErrDuplicateCode
		}
		return err
	}
	return nil
}

// FindByCode returns the share link with the given code (any status).
func (r *Repository) FindByCode(ctx context.Context, code string) (*ShareLink, error) {
	var s ShareLink
	err := r.coll.FindOne(ctx, bson.M{"code": code}).Decode(&s)
	if errors.Is(err, mongo.ErrNoDocuments) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return &s, nil
}

// FindActiveBySubmission returns an ACTIVE link for the given submission.
func (r *Repository) FindActiveBySubmission(ctx context.Context, submissionID, ownerID primitive.ObjectID) (*ShareLink, error) {
	var s ShareLink
	err := r.coll.FindOne(ctx, bson.M{
		"submissionId": submissionID,
		"ownerId":      ownerID,
		"status":       StatusActive,
	}).Decode(&s)
	if errors.Is(err, mongo.ErrNoDocuments) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return &s, nil
}

// ListByOwner returns the user's share links, newest first.
func (r *Repository) ListByOwner(ctx context.Context, ownerID primitive.ObjectID, limit, offset int64) ([]*ShareLink, error) {
	opts := options.Find().
		SetSort(bson.D{{Key: "createdAt", Value: -1}}).
		SetLimit(limit).
		SetSkip(offset)

	cur, err := r.coll.Find(ctx, bson.M{"ownerId": ownerID}, opts)
	if err != nil {
		return nil, err
	}
	defer cur.Close(ctx)

	var out []*ShareLink
	if err := cur.All(ctx, &out); err != nil {
		return nil, err
	}
	return out, nil
}

// UpdateStatus atomically transitions a link from one status to another,
// scoped to the owner.
func (r *Repository) UpdateStatus(ctx context.Context, id, ownerID primitive.ObjectID, from, to Status, now time.Time) (bool, error) {
	res, err := r.coll.UpdateOne(ctx,
		bson.M{"_id": id, "ownerId": ownerID, "status": from},
		bson.M{"$set": bson.M{"status": to, "updatedAt": now}},
	)
	if err != nil {
		return false, err
	}
	return res.MatchedCount == 1, nil
}

// SoftDelete marks a link as DELETED, scoped to the owner.
func (r *Repository) SoftDelete(ctx context.Context, id, ownerID primitive.ObjectID, now time.Time) (bool, error) {
	res, err := r.coll.UpdateOne(ctx,
		bson.M{"_id": id, "ownerId": ownerID, "status": bson.M{"$ne": StatusDeleted}},
		bson.M{"$set": bson.M{"status": StatusDeleted, "updatedAt": now}},
	)
	if err != nil {
		return false, err
	}
	return res.MatchedCount == 1, nil
}
