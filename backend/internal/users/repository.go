package users

import (
	"context"
	"errors"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

// ErrNotFound is returned when a user does not exist.
var ErrNotFound = errors.New("user not found")

// Repository provides MongoDB persistence for users.
type Repository struct {
	coll *mongo.Collection
}

// NewRepository builds a Repository bound to the users collection.
func NewRepository(db *mongo.Database) *Repository {
	return &Repository{coll: db.Collection("users")}
}

// Create inserts a new user.
func (r *Repository) Create(ctx context.Context, u *User) error {
	if u.ID.IsZero() {
		u.ID = primitive.NewObjectID()
	}
	_, err := r.coll.InsertOne(ctx, u)
	if err != nil {
		if mongo.IsDuplicateKeyError(err) {
			return ErrEmailExists
		}
		return err
	}
	return nil
}

// FindByEmail returns the user with the given email.
func (r *Repository) FindByEmail(ctx context.Context, email string) (*User, error) {
	var u User
	err := r.coll.FindOne(ctx, bson.M{"email": email}).Decode(&u)
	if errors.Is(err, mongo.ErrNoDocuments) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return &u, nil
}

// FindByID returns the user with the given id.
func (r *Repository) FindByID(ctx context.Context, id primitive.ObjectID) (*User, error) {
	var u User
	err := r.coll.FindOne(ctx, bson.M{"_id": id}).Decode(&u)
	if errors.Is(err, mongo.ErrNoDocuments) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return &u, nil
}

// ErrEmailExists signals a duplicate email on insert.
var ErrEmailExists = errors.New("email already exists")
