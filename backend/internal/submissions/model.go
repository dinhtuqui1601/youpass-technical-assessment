package submissions

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Submission statuses.
const (
	StatusActive  = "ACTIVE"
	StatusDeleted = "DELETED"
)

// Submission is a learner's piece of work that can be shared via a short link.
type Submission struct {
	ID        primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	OwnerID   primitive.ObjectID `bson:"ownerId" json:"ownerId"`
	Title     string             `bson:"title" json:"title"`
	Content   string             `bson:"content" json:"content"`
	Status    string             `bson:"status" json:"status"`
	CreatedAt time.Time          `bson:"createdAt" json:"createdAt"`
	UpdatedAt time.Time          `bson:"updatedAt" json:"updatedAt"`
}
