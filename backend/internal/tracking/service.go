package tracking

import (
	"context"
	"log/slog"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"go.mongodb.org/mongo-driver/bson/primitive"

	"github.com/youpass/backend/internal/share"
	"github.com/youpass/backend/pkg/response"
)

// Redis stream names used by the view-tracking pipeline.
const (
	StreamName = "view_events"
	GroupName  = "view_workers"
	DLQStream  = "view_events_dlq"
)

// ShareAccess is the subset of the share module used by tracking to verify
// link accessibility and ownership.
type ShareAccess interface {
	ResolvePublic(ctx context.Context, code string) (*share.ShareLink, error)
	GetOwned(ctx context.Context, ownerID primitive.ObjectID, code string) (*share.ShareLink, error)
}

// Service enqueues view events and serves view statistics.
type Service struct {
	shareAccess ShareAccess
	repo        *Repository
	rdb         *redis.Client
	logger      *slog.Logger
}

// NewService builds a tracking Service.
func NewService(shareAccess ShareAccess, repo *Repository, rdb *redis.Client, logger *slog.Logger) *Service {
	return &Service{shareAccess: shareAccess, repo: repo, rdb: rdb, logger: logger}
}

// RecordView validates access and enqueues the view event into Redis Streams.
// It never blocks the content path; queue writes are the only synchronous work.
func (s *Service) RecordView(ctx context.Context, code, viewID, ip, userAgent string) error {
	if _, err := uuid.Parse(viewID); err != nil {
		return response.NewError(response.CodeValidationError, "viewId must be a valid UUID")
	}

	link, err := s.shareAccess.ResolvePublic(ctx, code)
	if err != nil {
		return err
	}

	values := map[string]any{
		"shareLinkId": link.ID.Hex(),
		"viewId":      viewID,
		"viewedAt":    time.Now().UTC().Format(time.RFC3339Nano),
		"ip":          ip,
		"userAgent":   userAgent,
		"retries":     "0",
	}

	if err := s.rdb.XAdd(ctx, &redis.XAddArgs{Stream: StreamName, Values: values}).Err(); err != nil {
		// Surface queue write failures instead of silently dropping the event.
		if s.logger != nil {
			s.logger.Error("failed to enqueue view event", "error", err)
		}
		return response.NewError(response.CodeQueueUnavailable, "View event could not be enqueued, please retry")
	}
	return nil
}

// Stats returns view statistics for a link owned by the given user.
func (s *Service) Stats(ctx context.Context, ownerID primitive.ObjectID, code string) (*Stats, error) {
	link, err := s.shareAccess.GetOwned(ctx, ownerID, code)
	if err != nil {
		return nil, err
	}

	total, err := s.repo.Count(ctx, link.ID)
	if err != nil {
		return nil, err
	}
	last, err := s.repo.LastViewedAt(ctx, link.ID)
	if err != nil {
		return nil, err
	}
	return &Stats{TotalViews: total, LastViewedAt: last}, nil
}

