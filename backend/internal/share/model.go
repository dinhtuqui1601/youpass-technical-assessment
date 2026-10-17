package share

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Status is the lifecycle state of a share link.
type Status string

const (
	StatusActive   Status = "ACTIVE"
	StatusDisabled Status = "DISABLED"
	StatusDeleted  Status = "DELETED"
)

// ShareLink maps a short code to a learner's submission.
type ShareLink struct {
	ID           primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Code         string             `bson:"code" json:"code"`
	SubmissionID primitive.ObjectID `bson:"submissionId" json:"submissionId"`
	OwnerID      primitive.ObjectID `bson:"ownerId" json:"ownerId"`
	Status       Status             `bson:"status" json:"status"`
	ExpiresAt    *time.Time         `bson:"expiresAt,omitempty" json:"expiresAt,omitempty"`
	CreatedAt    time.Time          `bson:"createdAt" json:"createdAt"`
	UpdatedAt    time.Time          `bson:"updatedAt" json:"updatedAt"`
}
