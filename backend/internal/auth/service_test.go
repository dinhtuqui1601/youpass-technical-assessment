package auth

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"

	"github.com/youpass/backend/internal/users"
	"github.com/youpass/backend/pkg/response"
)

// ---- in-memory fakes ----

type fakeUserRepo struct {
	mu     sync.Mutex
	byMail map[string]*users.User
	byID   map[string]*users.User
}

func newFakeUserRepo() *fakeUserRepo {
	return &fakeUserRepo{byMail: map[string]*users.User{}, byID: map[string]*users.User{}}
}

func (f *fakeUserRepo) Create(_ context.Context, u *users.User) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	if u.ID.IsZero() {
		u.ID = primitive.NewObjectID()
	}
	if _, ok := f.byMail[u.Email]; ok {
		return users.ErrEmailExists
	}
	cp := *u
	f.byMail[u.Email] = &cp
	f.byID[u.ID.Hex()] = &cp
	return nil
}

func (f *fakeUserRepo) FindByEmail(_ context.Context, email string) (*users.User, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	u, ok := f.byMail[email]
	if !ok {
		return nil, users.ErrNotFound
	}
	cp := *u
	return &cp, nil
}

func (f *fakeUserRepo) FindByID(_ context.Context, id primitive.ObjectID) (*users.User, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	u, ok := f.byID[id.Hex()]
	if !ok {
		return nil, users.ErrNotFound
	}
	cp := *u
	return &cp, nil
}

type fakeSessionRepo struct {
	mu       sync.Mutex
	sessions map[string]*Session // key: id.Hex()
}

func newFakeSessionRepo() *fakeSessionRepo {
	return &fakeSessionRepo{sessions: map[string]*Session{}}
}

func (f *fakeSessionRepo) Create(_ context.Context, s *Session) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	if s.ID.IsZero() {
		s.ID = primitive.NewObjectID()
	}
	cp := *s
	f.sessions[s.ID.Hex()] = &cp
	return nil
}

func (f *fakeSessionRepo) FindByHash(_ context.Context, hash string) (*Session, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	for _, s := range f.sessions {
		if s.RefreshTokenHash == hash || s.PrevRefreshTokenHash == hash {
			cp := *s
			return &cp, nil
		}
	}
	return nil, ErrSessionNotFound
}

func (f *fakeSessionRepo) Rotate(_ context.Context, id primitive.ObjectID, oldHash, newHash string, graceUntil, now time.Time) (bool, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	s, ok := f.sessions[id.Hex()]
	if !ok {
		return false, ErrSessionNotFound
	}
	if s.RefreshTokenHash != oldHash || s.Revoked || !s.ExpiresAt.After(now) {
		return false, nil
	}
	s.RefreshTokenHash = newHash
	s.PrevRefreshTokenHash = oldHash
	s.PrevHashGraceUntil = graceUntil
	s.UpdatedAt = now
	return true, nil
}

func (f *fakeSessionRepo) Revoke(_ context.Context, id primitive.ObjectID, now time.Time) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	if s, ok := f.sessions[id.Hex()]; ok {
		s.Revoked = true
		rt := now
		s.RevokedAt = &rt
		s.UpdatedAt = now
	}
	return nil
}

// expireAll forces every session to be expired (test helper).
func (f *fakeSessionRepo) expireAll() {
	f.mu.Lock()
	defer f.mu.Unlock()
	for _, s := range f.sessions {
		s.ExpiresAt = time.Now().UTC().Add(-time.Second)
	}
}

// ---- helpers ----

func newTestService(t *testing.T, gracePeriod time.Duration) (*Service, *fakeUserRepo, *fakeSessionRepo) {
	t.Helper()
	ur := newFakeUserRepo()
	sr := newFakeSessionRepo()
	tm := NewTokenManager("test-secret", 15*time.Minute, "youpass", "youpass-web")
	svc := NewService(ur, sr, tm, 7*24*time.Hour, gracePeriod)
	return svc, ur, sr
}

func registerAndLogin(t *testing.T, svc *Service, email, password string) (*LoginResponse, string) {
	t.Helper()
	ctx := context.Background()
	if _, err := svc.Register(ctx, RegisterRequest{Email: email, Password: password, Name: "Tester"}); err != nil {
		t.Fatalf("Register error: %v", err)
	}
	resp, refreshToken, err := svc.Login(ctx, LoginRequest{Email: email, Password: password}, nil)
	if err != nil {
		t.Fatalf("Login error: %v", err)
	}
	return resp, refreshToken
}

func assertAPIErrorCode(t *testing.T, err error, wantCode string) {
	t.Helper()
	var apiErr *response.APIError
	if !errors.As(err, &apiErr) {
		t.Fatalf("error = %v, want APIError with code %s", err, wantCode)
	}
	if apiErr.Code != wantCode {
		t.Fatalf("error code = %s, want %s", apiErr.Code, wantCode)
	}
}

// ---- tests ----

func TestRegisterAndLogin(t *testing.T) {
	svc, _, _ := newTestService(t, 30*time.Second)
	resp, refreshToken := registerAndLogin(t, svc, "a@example.com", "password123")

	if resp.AccessToken == "" || refreshToken == "" {
		t.Fatal("expected non-empty tokens")
	}
	if resp.User.Email != "a@example.com" {
		t.Errorf("user email = %q", resp.User.Email)
	}
}

func TestRegisterDuplicateEmail(t *testing.T) {
	svc, _, _ := newTestService(t, 30*time.Second)
	ctx := context.Background()
	req := RegisterRequest{Email: "dup@example.com", Password: "password123", Name: "Tester"}
	if _, err := svc.Register(ctx, req); err != nil {
		t.Fatalf("Register error: %v", err)
	}
	if _, err := svc.Register(ctx, req); err == nil {
		t.Fatal("expected duplicate email error")
	} else {
		assertAPIErrorCode(t, err, response.CodeEmailExists)
	}
}

func TestLoginWrongPassword(t *testing.T) {
	svc, _, _ := newTestService(t, 30*time.Second)
	registerAndLogin(t, svc, "x@example.com", "password123")

	_, _, err := svc.Login(context.Background(), LoginRequest{Email: "x@example.com", Password: "wrong-pass"}, nil)
	if err == nil {
		t.Fatal("expected invalid credentials error")
	}
	assertAPIErrorCode(t, err, response.CodeInvalidCredentials)
}

func TestRefreshRotation(t *testing.T) {
	svc, _, _ := newTestService(t, 30*time.Second)
	_, refreshToken := registerAndLogin(t, svc, "r@example.com", "password123")

	resp, newToken, err := svc.Refresh(context.Background(), refreshToken)
	if err != nil {
		t.Fatalf("Refresh error: %v", err)
	}
	if newToken == refreshToken {
		t.Fatal("refresh token was not rotated")
	}
	if resp.AccessToken == "" {
		t.Fatal("expected new access token")
	}

	// The new token must still be valid for a subsequent refresh.
	if _, _, err := svc.Refresh(context.Background(), newToken); err != nil {
		t.Fatalf("Refresh with rotated token error: %v", err)
	}
}

func TestRefreshExpired(t *testing.T) {
	svc, _, sr := newTestService(t, 30*time.Second)
	_, refreshToken := registerAndLogin(t, svc, "e@example.com", "password123")

	sr.expireAll()

	if _, _, err := svc.Refresh(context.Background(), refreshToken); err == nil {
		t.Fatal("expected refresh expired error")
	} else {
		assertAPIErrorCode(t, err, response.CodeRefreshTokenExpired)
	}
}

func TestLogoutRevokesSession(t *testing.T) {
	svc, _, _ := newTestService(t, 30*time.Second)
	_, refreshToken := registerAndLogin(t, svc, "l@example.com", "password123")

	if err := svc.Logout(context.Background(), refreshToken); err != nil {
		t.Fatalf("Logout error: %v", err)
	}

	if _, _, err := svc.Refresh(context.Background(), refreshToken); err == nil {
		t.Fatal("expected revoked session error")
	} else {
		assertAPIErrorCode(t, err, response.CodeSessionRevoked)
	}
}

func TestConcurrentRefreshTwoTabs(t *testing.T) {
	svc, _, _ := newTestService(t, 30*time.Second)
	_, refreshToken := registerAndLogin(t, svc, "c2@example.com", "password123")

	const n = 2
	type result struct {
		token string
		err   error
	}
	results := make(chan result, n)
	for i := 0; i < n; i++ {
		go func() {
			_, tk, err := svc.Refresh(context.Background(), refreshToken)
			results <- result{token: tk, err: err}
		}()
	}

	var newTokens []string
	for i := 0; i < n; i++ {
		r := <-results
		if r.err != nil {
			t.Fatalf("concurrent refresh %d failed: %v", i, r.err)
		}
		newTokens = append(newTokens, r.token)
	}

	// At least one of the freshly issued tokens must remain usable.
	if _, _, err := svc.Refresh(context.Background(), newTokens[0]); err != nil {
		t.Fatalf("refresh with a rotated-out token failed: %v", err)
	}
}

func TestConcurrentRefreshNeverRevokes(t *testing.T) {
	svc, _, sr := newTestService(t, 30*time.Second)
	_, refreshToken := registerAndLogin(t, svc, "c5@example.com", "password123")

	const n = 5
	type result struct {
		token string
		err   error
	}
	results := make(chan result, n)
	for i := 0; i < n; i++ {
		go func() {
			_, tk, err := svc.Refresh(context.Background(), refreshToken)
			results <- result{token: tk, err: err}
		}()
	}

	successes := 0
	for i := 0; i < n; i++ {
		r := <-results
		if r.err == nil {
			successes++
			continue
		}
		// A legitimate concurrent burst must never look like reuse: no family revocation.
		var apiErr *response.APIError
		if errors.As(r.err, &apiErr) && apiErr.Code == response.CodeSessionRevoked {
			t.Fatalf("concurrent refresh caused spurious family revocation: %v", r.err)
		}
	}

	if successes == 0 {
		t.Fatal("expected at least one concurrent refresh to succeed")
	}

	// No session may have been spuriously revoked.
	sr.mu.Lock()
	defer sr.mu.Unlock()
	for _, s := range sr.sessions {
		if s.Revoked {
			t.Fatal("concurrent refresh caused spurious family revocation")
		}
	}
}

func TestRefreshReuseRevokesFamily(t *testing.T) {
	svc, _, _ := newTestService(t, 50*time.Millisecond)
	_, refreshToken := registerAndLogin(t, svc, "reuse@example.com", "password123")

	// Rotate once: refreshToken becomes the previous token.
	_, newToken, err := svc.Refresh(context.Background(), refreshToken)
	if err != nil {
		t.Fatalf("Refresh error: %v", err)
	}

	// Wait for the grace window to lapse, then reuse the old token.
	time.Sleep(80 * time.Millisecond)
	if _, _, err := svc.Refresh(context.Background(), refreshToken); err == nil {
		t.Fatal("expected reuse to fail")
	} else {
		assertAPIErrorCode(t, err, response.CodeInvalidRefreshToken)
	}

	// The whole family must now be revoked.
	if _, _, err := svc.Refresh(context.Background(), newToken); err == nil {
		t.Fatal("expected family revocation after reuse")
	} else {
		assertAPIErrorCode(t, err, response.CodeSessionRevoked)
	}
}
