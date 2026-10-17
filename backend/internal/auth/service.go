package auth

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"golang.org/x/crypto/bcrypt"

	"github.com/youpass/backend/internal/users"
	"github.com/youpass/backend/pkg/response"
)

const bcryptCost = bcrypt.DefaultCost

// UserRepository is the consumer-side interface for user persistence.
type UserRepository interface {
	Create(ctx context.Context, u *users.User) error
	FindByEmail(ctx context.Context, email string) (*users.User, error)
	FindByID(ctx context.Context, id primitive.ObjectID) (*users.User, error)
}

// SessionRepository is the consumer-side interface for session persistence.
type SessionRepository interface {
	Create(ctx context.Context, s *Session) error
	FindByHash(ctx context.Context, hash string) (*Session, error)
	Rotate(ctx context.Context, id primitive.ObjectID, oldHash, newHash string, graceUntil, now time.Time) (bool, error)
	Revoke(ctx context.Context, id primitive.ObjectID, now time.Time) error
}

// Service contains the authentication business logic.
type Service struct {
	users       UserRepository
	sessions    SessionRepository
	tokens      *TokenManager
	refreshTTL  time.Duration
	gracePeriod time.Duration
}

// NewService builds an auth Service.
func NewService(usersRepo UserRepository, sessionsRepo SessionRepository, tokens *TokenManager, refreshTTL, gracePeriod time.Duration) *Service {
	return &Service{
		users:       usersRepo,
		sessions:    sessionsRepo,
		tokens:      tokens,
		refreshTTL:  refreshTTL,
		gracePeriod: gracePeriod,
	}
}

// Pre-defined API errors returned by the service.
var (
	ErrInvalidCredentials  = response.NewError(response.CodeInvalidCredentials, "Invalid email or password")
	ErrRefreshTokenExpired = response.NewError(response.CodeRefreshTokenExpired, "Refresh token has expired")
	ErrInvalidRefreshToken = response.NewError(response.CodeInvalidRefreshToken, "Invalid refresh token")
	ErrEmailExists         = response.NewError(response.CodeEmailExists, "Email already registered")
	ErrUserNotFound        = response.NewError(response.CodeNotFound, "User not found")
)

// Register creates a new learner account (demo endpoint).
func (s *Service) Register(ctx context.Context, req RegisterRequest) (*users.User, error) {
	if _, err := s.users.FindByEmail(ctx, req.Email); err == nil {
		return nil, ErrEmailExists
	} else if !errors.Is(err, users.ErrNotFound) {
		return nil, err
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcryptCost)
	if err != nil {
		return nil, err
	}

	now := time.Now().UTC()
	u := &users.User{
		Email:        req.Email,
		PasswordHash: string(hash),
		Name:         req.Name,
		Role:         users.RoleStudent,
		CreatedAt:    now,
		UpdatedAt:    now,
	}

	if err := s.users.Create(ctx, u); err != nil {
		if errors.Is(err, users.ErrEmailExists) {
			return nil, ErrEmailExists
		}
		return nil, err
	}
	return u, nil
}

// Login verifies credentials and issues a new session. It returns the API
// response and the plaintext refresh token (to be set as an HttpOnly cookie).
func (s *Service) Login(ctx context.Context, req LoginRequest, metadata map[string]string) (*LoginResponse, string, error) {
	u, err := s.users.FindByEmail(ctx, req.Email)
	if errors.Is(err, users.ErrNotFound) {
		return nil, "", ErrInvalidCredentials
	}
	if err != nil {
		return nil, "", err
	}

	if bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(req.Password)) != nil {
		return nil, "", ErrInvalidCredentials
	}

	accessToken, expiresIn, err := s.tokens.Generate(u.ID.Hex(), u.Role)
	if err != nil {
		return nil, "", err
	}

	refreshToken, refreshHash, err := newRefreshToken()
	if err != nil {
		return nil, "", err
	}

	now := time.Now().UTC()
	session := &Session{
		SessionID:        uuid.NewString(),
		UserID:           u.ID.Hex(),
		RefreshTokenHash: refreshHash,
		ExpiresAt:        now.Add(s.refreshTTL),
		CreatedAt:        now,
		UpdatedAt:        now,
		Metadata:         metadata,
	}
	if err := s.sessions.Create(ctx, session); err != nil {
		return nil, "", err
	}

	return &LoginResponse{
		AccessToken: accessToken,
		ExpiresIn:   expiresIn,
		User:        toUserDTO(u),
	}, refreshToken, nil
}

// Refresh rotates the refresh token and issues a new access token. It is safe
// under concurrent requests: rotation is atomic, a short grace window tolerates
// multi-tab races, and reuse of a rotated-out token revokes the whole family.
func (s *Service) Refresh(ctx context.Context, refreshToken string) (*LoginResponse, string, error) {
	hash := hashRefreshToken(refreshToken)
	now := time.Now().UTC()

	for attempt := 0; attempt < 3; attempt++ {
		sess, err := s.sessions.FindByHash(ctx, hash)
		if errors.Is(err, ErrSessionNotFound) {
			return nil, "", ErrInvalidRefreshToken
		}
		if err != nil {
			return nil, "", err
		}

		if sess.Revoked {
			return nil, "", response.NewError(response.CodeSessionRevoked, "Session has been revoked")
		}
		if !sess.ExpiresAt.After(now) {
			return nil, "", ErrRefreshTokenExpired
		}

		switch {
		case sess.RefreshTokenHash == hash:
			// Normal rotation.
			newToken, newHash, err := newRefreshToken()
			if err != nil {
				return nil, "", err
			}
			ok, err := s.sessions.Rotate(ctx, sess.ID, hash, newHash, now.Add(s.gracePeriod), now)
			if err != nil {
				return nil, "", err
			}
			if ok {
				return s.issueTokens(ctx, sess.UserID, newToken)
			}
			// Lost a race with a concurrent refresh; reload and re-evaluate.

		case sess.PrevRefreshTokenHash == hash && now.Before(sess.PrevHashGraceUntil):
			// Legitimate concurrent refresh (multi-tab): rotate the current token.
			newToken, newHash, err := newRefreshToken()
			if err != nil {
				return nil, "", err
			}
			ok, err := s.sessions.Rotate(ctx, sess.ID, sess.RefreshTokenHash, newHash, now.Add(s.gracePeriod), now)
			if err != nil {
				return nil, "", err
			}
			if ok {
				return s.issueTokens(ctx, sess.UserID, newToken)
			}

		default:
			// Suspicious reuse: revoke the whole token family.
			_ = s.sessions.Revoke(ctx, sess.ID, now)
			return nil, "", ErrInvalidRefreshToken
		}
	}

	return nil, "", response.NewError(response.CodeInternalError, "Too many concurrent refreshes")
}

// Logout revokes the session associated with the refresh token. It is
// idempotent.
func (s *Service) Logout(ctx context.Context, refreshToken string) error {
	hash := hashRefreshToken(refreshToken)
	sess, err := s.sessions.FindByHash(ctx, hash)
	if errors.Is(err, ErrSessionNotFound) {
		return nil
	}
	if err != nil {
		return err
	}
	return s.sessions.Revoke(ctx, sess.ID, time.Now().UTC())
}

// GetUserByID returns the user with the given id.
func (s *Service) GetUserByID(ctx context.Context, id primitive.ObjectID) (*users.User, error) {
	u, err := s.users.FindByID(ctx, id)
	if errors.Is(err, users.ErrNotFound) {
		return nil, ErrUserNotFound
	}
	return u, err
}

// issueTokens generates an access token for the given user and pairs it with
// the provided refresh token.
func (s *Service) issueTokens(ctx context.Context, userID, refreshToken string) (*LoginResponse, string, error) {
	id, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return nil, "", err
	}
	u, err := s.users.FindByID(ctx, id)
	if errors.Is(err, users.ErrNotFound) {
		return nil, "", ErrUserNotFound
	}
	if err != nil {
		return nil, "", err
	}

	accessToken, expiresIn, err := s.tokens.Generate(u.ID.Hex(), u.Role)
	if err != nil {
		return nil, "", err
	}

	return &LoginResponse{
		AccessToken: accessToken,
		ExpiresIn:   expiresIn,
		User:        toUserDTO(u),
	}, refreshToken, nil
}

func toUserDTO(u *users.User) UserDTO {
	return UserDTO{
		ID:        u.ID.Hex(),
		Email:     u.Email,
		Name:      u.Name,
		Role:      u.Role,
		CreatedAt: u.CreatedAt,
	}
}
