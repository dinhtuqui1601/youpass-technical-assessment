package auth

import (
	"context"
	"errors"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

// ErrSessionNotFound is returned when a session does not exist.
var ErrSessionNotFound = errors.New("session not found")

// Repository provides MongoDB persistence for authentication sessions.
type Repository struct {
	coll *mongo.Collection
}

// NewRepository builds a Repository bound to the auth_sessions collection.
func NewRepository(db *mongo.Database) *Repository {
	return &Repository{coll: db.Collection("auth_sessions")}
}

// Create inserts a new session.
func (r *Repository) Create(ctx context.Context, s *Session) error {
	if s.ID.IsZero() {
		s.ID = primitive.NewObjectID()
	}
	_, err := r.coll.InsertOne(ctx, s)
	return err
}

// FindByHash returns the session whose current or previous refresh-token hash
// matches the given hash.
func (r *Repository) FindByHash(ctx context.Context, hash string) (*Session, error) {
	var s Session
	filter := bson.M{
		"$or": []bson.M{
			{"refreshTokenHash": hash},
			{"prevRefreshTokenHash": hash},
		},
	}
	err := r.coll.FindOne(ctx, filter).Decode(&s)
	if errors.Is(err, mongo.ErrNoDocuments) {
		return nil, ErrSessionNotFound
	}
	if err != nil {
		return nil, err
	}
	return &s, nil
}

// Rotate atomically replaces oldHash with newHash. It succeeds only when the
// stored current hash still equals oldHash and the session is still active, so
// concurrent rotations of the same token cannot both succeed.
func (r *Repository) Rotate(ctx context.Context, id primitive.ObjectID, oldHash, newHash string, graceUntil, now time.Time) (bool, error) {
	res, err := r.coll.UpdateOne(ctx,
		bson.M{
			"_id":              id,
			"refreshTokenHash": oldHash,
			"revoked":          false,
			"expiresAt":        bson.M{"$gt": now},
		},
		bson.M{"$set": bson.M{
			"refreshTokenHash":     newHash,
			"prevRefreshTokenHash": oldHash,
			"prevHashGraceUntil":   graceUntil,
			"updatedAt":            now,
		}},
	)
	if err != nil {
		return false, err
	}
	return res.MatchedCount == 1, nil
}

// Revoke marks a session (token family) as revoked.
func (r *Repository) Revoke(ctx context.Context, id primitive.ObjectID, now time.Time) error {
	_, err := r.coll.UpdateOne(ctx,
		bson.M{"_id": id},
		bson.M{"$set": bson.M{"revoked": true, "revokedAt": now, "updatedAt": now}},
	)
	return err
}
