package share

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"

	"github.com/youpass/backend/internal/submissions"
	"github.com/youpass/backend/pkg/response"
)

const (
	linkCacheTTL   = 30 * time.Second
	cacheKeyPrefix = "share:code:"
)

// SubmissionProvider is the subset of the submissions module used by share.
type SubmissionProvider interface {
	FindByID(ctx context.Context, id primitive.ObjectID) (*submissions.Submission, error)
}

// Cache is the subset of the cache layer used by share.
type Cache interface {
	Get(ctx context.Context, key string) (string, error)
	Set(ctx context.Context, key, value string, ttl time.Duration) error
	Del(ctx context.Context, keys ...string) error
}

// Service contains the share-link business logic.
type Service struct {
	repo        *Repository
	submissions SubmissionProvider
	cache       Cache
	codeLength  int
	maxRetries  int
}

// NewService builds a share Service.
func NewService(repo *Repository, submissions SubmissionProvider, cache Cache, codeLength, maxRetries int) *Service {
	return &Service{
		repo:        repo,
		submissions: submissions,
		cache:       cache,
		codeLength:  codeLength,
		maxRetries:  maxRetries,
	}
}

// Create creates (or returns the existing ACTIVE) share link for a submission.
// Only the owner of the submission may create links.
func (s *Service) Create(ctx context.Context, ownerID, submissionID primitive.ObjectID) (*ShareLinkDTO, error) {
	sub, err := s.submissions.FindByID(ctx, submissionID)
	if errors.Is(err, submissions.ErrNotFound) {
		return nil, response.NewError(response.CodeSubmissionNotFound, "Submission not found")
	}
	if err != nil {
		return nil, err
	}
	if sub.OwnerID != ownerID {
		return nil, response.NewError(response.CodeForbidden, "You do not own this submission")
	}
	if sub.Status != submissions.StatusActive {
		return nil, response.NewError(response.CodeForbidden, "Submission is not shareable")
	}

	// Reuse an existing ACTIVE link instead of creating duplicates.
	if existing, err := s.repo.FindActiveBySubmission(ctx, submissionID, ownerID); err == nil {
		return toDTO(existing), nil
	}

	now := time.Now().UTC()
	for attempt := 0; attempt < s.maxRetries; attempt++ {
		code, err := GenerateCode(s.codeLength)
		if err != nil {
			return nil, err
		}

		link := &ShareLink{
			Code:         code,
			SubmissionID: submissionID,
			OwnerID:      ownerID,
			Status:       StatusActive,
			CreatedAt:    now,
			UpdatedAt:    now,
		}
		if err := s.repo.Create(ctx, link); err != nil {
			if errors.Is(err, ErrDuplicateCode) {
				continue // collision: retry with a fresh code
			}
			return nil, err
		}
		return toDTO(link), nil
	}

	return nil, response.NewError(response.CodeInternalError, "Could not generate a unique short code")
}

// List returns the authenticated user's share links.
func (s *Service) List(ctx context.Context, ownerID primitive.ObjectID, limit, offset int64) ([]*ShareLinkDTO, error) {
	links, err := s.repo.ListByOwner(ctx, ownerID, limit, offset)
	if err != nil {
		return nil, err
	}
	out := make([]*ShareLinkDTO, 0, len(links))
	for _, l := range links {
		out = append(out, toDTO(l))
	}
	return out, nil
}

// Get returns the owner-facing detail of a link.
func (s *Service) Get(ctx context.Context, ownerID primitive.ObjectID, code string) (*ShareLinkDTO, error) {
	link, err := s.repo.FindByCode(ctx, code)
	if err != nil {
		return nil, err
	}
	if link.OwnerID != ownerID {
		return nil, response.NewError(response.CodeForbidden, "You do not own this share link")
	}
	return toDTO(link), nil
}

// UpdateStatus toggles a link between ACTIVE and DISABLED, scoped to the owner.
// DELETED links can never be reactivated.
func (s *Service) UpdateStatus(ctx context.Context, ownerID primitive.ObjectID, code string, target Status) (*ShareLinkDTO, error) {
	link, err := s.repo.FindByCode(ctx, code)
	if err != nil {
		return nil, err
	}
	if link.OwnerID != ownerID {
		return nil, response.NewError(response.CodeForbidden, "You do not own this share link")
	}
	if link.Status == StatusDeleted {
		return nil, response.NewError(response.CodeShareLinkDeleted, "Deleted links cannot be modified")
	}

	now := time.Now().UTC()
	ok := false
	switch {
	case link.Status == StatusActive && target == StatusDisabled:
		ok, err = s.repo.UpdateStatus(ctx, link.ID, ownerID, StatusActive, StatusDisabled, now)
	case link.Status == StatusDisabled && target == StatusActive:
		ok, err = s.repo.UpdateStatus(ctx, link.ID, ownerID, StatusDisabled, StatusActive, now)
	case link.Status == target:
		ok, err = true, nil // already in desired state
	default:
		return nil, response.NewError(response.CodeValidationError, "Invalid status transition")
	}
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, response.NewError(response.CodeConflict, "Status changed concurrently, please retry")
	}

	// Invalidate any cached mapping so redirects stop promptly.
	_ = s.cache.Del(ctx, cacheKeyPrefix+code)

	link.Status = target
	link.UpdatedAt = now
	return toDTO(link), nil
}

// Delete soft-deletes a link (keeps history, never reuses the code).
func (s *Service) Delete(ctx context.Context, ownerID primitive.ObjectID, code string) error {
	link, err := s.repo.FindByCode(ctx, code)
	if err != nil {
		return err
	}
	if link.OwnerID != ownerID {
		return response.NewError(response.CodeForbidden, "You do not own this share link")
	}
	if link.Status == StatusDeleted {
		return nil // idempotent
	}

	if _, err := s.repo.SoftDelete(ctx, link.ID, ownerID, time.Now().UTC()); err != nil {
		return err
	}
	_ = s.cache.Del(ctx, cacheKeyPrefix+code)
	return nil
}

// GetOwned returns the link if the given user owns it.
func (s *Service) GetOwned(ctx context.Context, ownerID primitive.ObjectID, code string) (*ShareLink, error) {
	link, err := s.repo.FindByCode(ctx, code)
	if err != nil {
		return nil, err
	}
	if link.OwnerID != ownerID {
		return nil, response.NewError(response.CodeForbidden, "You do not own this share link")
	}
	return link, nil
}

// ResolveRedirect resolves a short code for the /s/:code redirect. Redis caches
// the code mapping; MongoDB remains the authority for access decisions.
func (s *Service) ResolveRedirect(ctx context.Context, code string) (*ShareLink, error) {
	key := cacheKeyPrefix + code

	if v, err := s.cache.Get(ctx, key); err == nil && v != "" {
		var cached cachedLink
		if json.Unmarshal([]byte(v), &cached) == nil &&
			cached.Status == string(StatusActive) &&
			(cached.ExpiresAt == nil || cached.ExpiresAt.After(time.Now().UTC())) {
			return &ShareLink{Code: code, SubmissionID: cached.SubmissionID, Status: StatusActive}, nil
		}
	}

	link, err := s.repo.FindByCode(ctx, code)
	if errors.Is(err, ErrNotFound) {
		return nil, response.NewError(response.CodeShareLinkNotFound, "Share link not found")
	}
	if err != nil {
		return nil, err
	}

	// Cache the mapping (best effort); content is still protected downstream.
	_ = s.cache.Set(ctx, key, marshalCached(link), linkCacheTTL)

	return s.checkAccessible(ctx, link)
}

// ResolvePublic returns the link if it is publicly accessible. It is the
// authoritative access check used by content and view-tracking endpoints.
func (s *Service) ResolvePublic(ctx context.Context, code string) (*ShareLink, error) {
	link, err := s.repo.FindByCode(ctx, code)
	if errors.Is(err, ErrNotFound) {
		return nil, response.NewError(response.CodeShareLinkNotFound, "Share link not found")
	}
	if err != nil {
		return nil, err
	}
	return s.checkAccessible(ctx, link)
}

// PublicShare returns the shared submission content after an authoritative
// access check.
func (s *Service) PublicShare(ctx context.Context, code string) (*PublicShareDTO, error) {
	link, err := s.ResolvePublic(ctx, code)
	if err != nil {
		return nil, err
	}

	sub, err := s.submissions.FindByID(ctx, link.SubmissionID)
	if err != nil {
		return nil, err
	}

	return &PublicShareDTO{
		Code:    code,
		Title:   sub.Title,
		Content: sub.Content,
		Status:  string(link.Status),
	}, nil
}

// checkAccessible is the authoritative access gate (fail closed).
func (s *Service) checkAccessible(ctx context.Context, link *ShareLink) (*ShareLink, error) {
	switch link.Status {
	case StatusDeleted:
		return nil, response.NewError(response.CodeShareLinkDeleted, "Share link has been deleted")
	case StatusDisabled:
		return nil, response.NewError(response.CodeShareLinkDisabled, "Share link is disabled")
	}

	if link.ExpiresAt != nil && !link.ExpiresAt.After(time.Now().UTC()) {
		return nil, response.NewError(response.CodeShareLinkExpired, "Share link has expired")
	}

	sub, err := s.submissions.FindByID(ctx, link.SubmissionID)
	if errors.Is(err, submissions.ErrNotFound) {
		return nil, response.NewError(response.CodeSubmissionNotFound, "Submission no longer exists")
	}
	if err != nil {
		// Fail closed: never serve content on a DB error.
		return nil, err
	}
	if sub.Status != submissions.StatusActive {
		return nil, response.NewError(response.CodeForbidden, "Submission is not shareable")
	}

	return link, nil
}

type cachedLink struct {
	SubmissionID primitive.ObjectID `json:"submissionId"`
	Status       string             `json:"status"`
	ExpiresAt    *time.Time         `json:"expiresAt,omitempty"`
}

func marshalCached(link *ShareLink) string {
	b, err := json.Marshal(cachedLink{
		SubmissionID: link.SubmissionID,
		Status:       string(link.Status),
		ExpiresAt:    link.ExpiresAt,
	})
	if err != nil {
		return ""
	}
	return string(b)
}

func toDTO(link *ShareLink) *ShareLinkDTO {
	return &ShareLinkDTO{
		ID:           link.ID.Hex(),
		Code:         link.Code,
		SubmissionID: link.SubmissionID.Hex(),
		Status:       string(link.Status),
		ExpiresAt:    link.ExpiresAt,
		URL:          "/s/" + link.Code,
		CreatedAt:    link.CreatedAt,
		UpdatedAt:    link.UpdatedAt,
	}
}

