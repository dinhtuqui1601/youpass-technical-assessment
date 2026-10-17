package users

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Role is the set of supported user roles.
const (
	RoleStudent = "student"
	RoleAdmin   = "admin"
)

// User is the account used for authentication.
type User struct {
	ID           primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Email        string             `bson:"email" json:"email"`
	PasswordHash string             `bson:"passwordHash" json:"-"`
	Name         string             `bson:"name" json:"name"`
	Role         string             `bson:"role" json:"role"`
	CreatedAt    time.Time          `bson:"createdAt" json:"createdAt"`
	UpdatedAt    time.Time          `bson:"updatedAt" json:"updatedAt"`
}
