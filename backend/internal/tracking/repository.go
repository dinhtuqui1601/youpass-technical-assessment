package tracking

import (
	"context"
	"errors"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// ErrDuplicateView signals that the view event was already recorded (idempotency).
var ErrDuplicateView = errors.New("duplicate view event")

// ViewRecord is a single (idempotent) view event. The unique index on
// (shareLinkId, viewId) makes the events collection the source of truth for
// view counts.
type ViewRecord struct {
	ID          primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	ShareLinkID primitive.ObjectID `bson:"shareLinkId" json:"shareLinkId"`
	ViewID      string             `bson:"viewId" json:"viewId"`
	ViewedAt    time.Time          `bson:"viewedAt" json:"viewedAt"`
	IP          string             `bson:"ip,omitempty" json:"-"`
	UserAgent   string             `bson:"userAgent,omitempty" json:"-"`
	CreatedAt   time.Time          `bson:"createdAt" json:"createdAt"`
}

// Stats is the aggregated view statistics for a share link.
type Stats struct {
	TotalViews   int64      `json:"totalViews"`
	LastViewedAt *time.Time `json:"lastViewedAt"`
}

// Repository provides MongoDB persistence for view events.
type Repository struct {
	coll *mongo.Collection
}

// NewRepository builds a Repository bound to the share_views collection.
func NewRepository(db *mongo.Database) *Repository {
	return &Repository{coll: db.Collection("share_views")}
}

// Record inserts a view event. A duplicate (shareLinkId, viewId) returns
// ErrDuplicateView so callers can treat the event as already processed.
func (r *Repository) Record(ctx context.Context, v *ViewRecord) error {
	if v.ID.IsZero() {
		v.ID = primitive.NewObjectID()
	}
	_, err := r.coll.InsertOne(ctx, v)
	if err != nil {
		if mongo.IsDuplicateKeyError(err) {
			return ErrDuplicateView
		}
		return err
	}
	return nil
}

// Count returns the number of recorded view events for a share link.
func (r *Repository) Count(ctx context.Context, shareLinkID primitive.ObjectID) (int64, error) {
	return r.coll.CountDocuments(ctx, bson.M{"shareLinkId": shareLinkID})
}

// LastViewedAt returns the most recent view time for a share link (nil if none).
func (r *Repository) LastViewedAt(ctx context.Context, shareLinkID primitive.ObjectID) (*time.Time, error) {
	opts := options.FindOne().SetSort(bson.D{{Key: "viewedAt", Value: -1}}).SetProjection(bson.M{"viewedAt": 1})
	var rec struct {
		ViewedAt time.Time `bson:"viewedAt"`
	}
	err := r.coll.FindOne(ctx, bson.M{"shareLinkId": shareLinkID}, opts).Decode(&rec)
	if errors.Is(err, mongo.ErrNoDocuments) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	t := rec.ViewedAt
	return &t, nil
}
