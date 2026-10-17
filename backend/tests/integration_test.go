package tests

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"os"
	"strconv"
	"testing"
	"time"

	"github.com/redis/go-redis/v9"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"

	"github.com/youpass/backend/internal/auth"
	"github.com/youpass/backend/internal/share"
	"github.com/youpass/backend/internal/submissions"
	"github.com/youpass/backend/internal/tracking"
	"github.com/youpass/backend/internal/users"
	"github.com/youpass/backend/pkg/cache"
	"github.com/youpass/backend/pkg/database"
	redispkg "github.com/youpass/backend/pkg/redis"
	"github.com/youpass/backend/pkg/response"
)

type testEnv struct {
	authSvc      *auth.Service
	shareSvc     *share.Service
	trackingSvc  *tracking.Service
	trackingRepo *tracking.Repository
	tokens       *auth.TokenManager

	mongo *mongo.Client
	rdb   *redis.Client
	db    *mongo.Database
}

func env(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

// setup connects to real MongoDB/Redis and skips the test when they are not
// reachable. It uses a dedicated database and Redis DB to isolate test data.
func setup(t *testing.T) *testEnv {
	t.Helper()

	mongoURI := env("TEST_MONGO_URI", env("MONGO_URI", "mongodb://admin:password_2026@localhost:27017/?authSource=admin"))
	redisAddr := env("TEST_REDIS_ADDR", env("REDIS_ADDR", "localhost:6379"))
	redisPassword := env("TEST_REDIS_PASSWORD", env("REDIS_PASSWORD", ""))

	ctx := context.Background()

	mc, err := database.Connect(ctx, mongoURI)
	if err != nil {
		t.Skipf("MongoDB not available, skipping integration test: %v", err)
	}
	dbName := "youpass_it_" + strconv.FormatInt(time.Now().UnixNano(), 10)
	db := mc.Database(dbName)
	if err := database.EnsureIndexes(ctx, db); err != nil {
		t.Fatalf("EnsureIndexes: %v", err)
	}

	rdb := redispkg.NewClient(redisAddr, redisPassword, 1) // Redis DB 1 isolates tests
	if err := redispkg.Ping(ctx, rdb); err != nil {
		t.Skipf("Redis not available, skipping integration test: %v", err)
	}
	rdb.FlushDB(ctx)

	tm := auth.NewTokenManager("it-test-secret", 15*time.Minute, "youpass", "youpass-web")
	usersRepo := users.NewRepository(db)
	sessionsRepo := auth.NewRepository(db)
	authSvc := auth.NewService(usersRepo, sessionsRepo, tm, 7*24*time.Hour, 30*time.Second)

	submissionsRepo := submissions.NewRepository(db)
	shareRepo := share.NewRepository(db)
	shareSvc := share.NewService(shareRepo, submissionsRepo, cache.New(rdb), 10, 5)

	trackingRepo := tracking.NewRepository(db)
	trackingSvc := tracking.NewService(shareSvc, trackingRepo, rdb, discardLogger())

	env := &testEnv{
		authSvc:      authSvc,
		shareSvc:     shareSvc,
		trackingSvc:  trackingSvc,
		trackingRepo: trackingRepo,
		tokens:       tm,
		mongo:        mc,
		rdb:          rdb,
		db:           db,
	}

	t.Cleanup(func() {
		_ = db.Drop(context.Background())
		_ = mc.Disconnect(context.Background())
		_ = rdb.Close()
	})
	return env
}

func discardLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(io.Discard, nil))
}

type account struct {
	user         *users.User
	accessToken  string
	refreshToken string
}

func registerAndLogin(t *testing.T, e *testEnv, email string) account {
	t.Helper()
	ctx := context.Background()

	u, err := e.authSvc.Register(ctx, auth.RegisterRequest{Email: email, Password: "password123", Name: "Integration"})
	if err != nil {
		t.Fatalf("Register: %v", err)
	}
	resp, refreshToken, err := e.authSvc.Login(ctx, auth.LoginRequest{Email: email, Password: "password123"}, nil)
	if err != nil {
		t.Fatalf("Login: %v", err)
	}
	return account{user: u, accessToken: resp.AccessToken, refreshToken: refreshToken}
}

func startWorker(t *testing.T, e *testEnv) func() {
	t.Helper()
	worker := tracking.NewWorker(e.rdb, e.trackingRepo, 4, 5, discardLogger())
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan struct{})
	go func() {
		_ = worker.Run(ctx)
		close(done)
	}()
	return func() {
		cancel()
		<-done
	}
}

func waitForViews(t *testing.T, e *testEnv, linkID primitive.ObjectID, want int64) {
	t.Helper()
	deadline := time.Now().Add(10 * time.Second)
	for time.Now().Before(deadline) {
		got, err := e.trackingRepo.Count(context.Background(), linkID)
		if err == nil && got == want {
			return
		}
		time.Sleep(100 * time.Millisecond)
	}
	got, _ := e.trackingRepo.Count(context.Background(), linkID)
	t.Fatalf("view count = %d, want %d within timeout", got, want)
}

func mustCreateSubmission(t *testing.T, e *testEnv, ownerID primitive.ObjectID) *submissions.Submission {
	t.Helper()
	now := time.Now().UTC()
	sub := &submissions.Submission{
		OwnerID:   ownerID,
		Title:     "Demo submission",
		Content:   "Demo content",
		Status:    submissions.StatusActive,
		CreatedAt: now,
		UpdatedAt: now,
	}
	if err := submissions.NewRepository(e.db).Create(context.Background(), sub); err != nil {
		t.Fatalf("Create submission: %v", err)
	}
	return sub
}

func assertCode(t *testing.T, err error, wantCode string) {
	t.Helper()
	var apiErr *response.APIError
	if !errors.As(err, &apiErr) {
		t.Fatalf("error = %v, want code %s", err, wantCode)
	}
	if apiErr.Code != wantCode {
		t.Fatalf("error code = %s, want %s", apiErr.Code, wantCode)
	}
}

// ---- Flows ----

func TestAuthFlow(t *testing.T) {
	e := setup(t)
	ctx := context.Background()
	acc := registerAndLogin(t, e, "auth@example.com")

	// Access token is valid.
	if _, err := e.tokens.Parse(acc.accessToken); err != nil {
		t.Fatalf("access token invalid: %v", err)
	}

	// Refresh rotates.
	resp, newToken, err := e.authSvc.Refresh(ctx, acc.refreshToken)
	if err != nil {
		t.Fatalf("Refresh: %v", err)
	}
	if newToken == acc.refreshToken {
		t.Fatal("refresh token not rotated")
	}
	if _, err := e.tokens.Parse(resp.AccessToken); err != nil {
		t.Fatalf("new access token invalid: %v", err)
	}

	// Retry protected API after refresh (simulated by parsing the new token).
	if _, err := e.tokens.Parse(resp.AccessToken); err != nil {
		t.Fatalf("retry access token invalid: %v", err)
	}

	// Logout revokes the session.
	if err := e.authSvc.Logout(ctx, newToken); err != nil {
		t.Fatalf("Logout: %v", err)
	}
	if _, _, err := e.authSvc.Refresh(ctx, newToken); err == nil {
		t.Fatal("expected refresh after logout to fail")
	} else {
		assertCode(t, err, response.CodeSessionRevoked)
	}
}

func TestShareFlow(t *testing.T) {
	e := setup(t)
	stop := startWorker(t, e)
	defer stop()

	ctx := context.Background()
	acc := registerAndLogin(t, e, "share@example.com")

	ownerID, _ := primitive.ObjectIDFromHex(acc.user.ID.Hex())
	sub := mustCreateSubmission(t, e, ownerID)

	// Create share link.
	linkDTO, err := e.shareSvc.Create(ctx, ownerID, sub.ID)
	if err != nil {
		t.Fatalf("Create share link: %v", err)
	}
	code := linkDTO.Code

	// Public content is accessible.
	public, err := e.shareSvc.PublicShare(ctx, code)
	if err != nil {
		t.Fatalf("PublicShare: %v", err)
	}
	if public.Title != sub.Title {
		t.Fatalf("public title = %q, want %q", public.Title, sub.Title)
	}

	// Record a view and verify it is counted exactly once.
	viewID := "11111111-1111-4111-8111-111111111111"
	if err := e.trackingSvc.RecordView(ctx, code, viewID, "127.0.0.1", "test"); err != nil {
		t.Fatalf("RecordView: %v", err)
	}
	// Idempotent duplicate of the same viewId.
	if err := e.trackingSvc.RecordView(ctx, code, viewID, "127.0.0.1", "test"); err != nil {
		t.Fatalf("RecordView duplicate: %v", err)
	}
	link, err := e.shareSvc.ResolvePublic(ctx, code)
	if err != nil {
		t.Fatalf("ResolvePublic: %v", err)
	}
	waitForViews(t, e, link.ID, 1)

	// Disable -> access denied.
	if _, err := e.shareSvc.UpdateStatus(ctx, ownerID, code, share.StatusDisabled); err != nil {
		t.Fatalf("Disable: %v", err)
	}
	if _, err := e.shareSvc.PublicShare(ctx, code); err == nil {
		t.Fatal("expected disabled link to be inaccessible")
	} else {
		assertCode(t, err, response.CodeShareLinkDisabled)
	}

	// Re-enable -> accessible again.
	if _, err := e.shareSvc.UpdateStatus(ctx, ownerID, code, share.StatusActive); err != nil {
		t.Fatalf("Re-enable: %v", err)
	}
	if _, err := e.shareSvc.PublicShare(ctx, code); err != nil {
		t.Fatalf("PublicShare after re-enable: %v", err)
	}

	// Delete -> inaccessible and cannot be reactivated.
	if err := e.shareSvc.Delete(ctx, ownerID, code); err != nil {
		t.Fatalf("Delete: %v", err)
	}
	if _, err := e.shareSvc.PublicShare(ctx, code); err == nil {
		t.Fatal("expected deleted link to be inaccessible")
	} else {
		assertCode(t, err, response.CodeShareLinkDeleted)
	}
	if _, err := e.shareSvc.UpdateStatus(ctx, ownerID, code, share.StatusActive); err == nil {
		t.Fatal("expected deleted link to reject reactivation")
	}
}

func TestOwnershipEnforced(t *testing.T) {
	e := setup(t)
	ctx := context.Background()

	owner := registerAndLogin(t, e, "owner@example.com")
	intruder := registerAndLogin(t, e, "intruder@example.com")

	ownerID, _ := primitive.ObjectIDFromHex(owner.user.ID.Hex())
	intruderID, _ := primitive.ObjectIDFromHex(intruder.user.ID.Hex())
	sub := mustCreateSubmission(t, e, ownerID)

	// Intruder cannot create a share link for someone else's submission.
	if _, err := e.shareSvc.Create(ctx, intruderID, sub.ID); err == nil {
		t.Fatal("expected intruder to be forbidden from creating a link")
	} else {
		assertCode(t, err, response.CodeForbidden)
	}

	linkDTO, err := e.shareSvc.Create(ctx, ownerID, sub.ID)
	if err != nil {
		t.Fatalf("Create share link: %v", err)
	}

	// Intruder cannot disable or delete the owner's link.
	if _, err := e.shareSvc.UpdateStatus(ctx, intruderID, linkDTO.Code, share.StatusDisabled); err == nil {
		t.Fatal("expected intruder to be forbidden from updating")
	} else {
		assertCode(t, err, response.CodeForbidden)
	}
	if err := e.shareSvc.Delete(ctx, intruderID, linkDTO.Code); err == nil {
		t.Fatal("expected intruder to be forbidden from deleting")
	} else {
		assertCode(t, err, response.CodeForbidden)
	}
}

func TestListSubmissionsByOwner(t *testing.T) {
	e := setup(t)
	ctx := context.Background()

	owner := registerAndLogin(t, e, "list-owner@example.com")
	other := registerAndLogin(t, e, "list-other@example.com")

	ownerID, _ := primitive.ObjectIDFromHex(owner.user.ID.Hex())
	otherID, _ := primitive.ObjectIDFromHex(other.user.ID.Hex())

	repo := submissions.NewRepository(e.db)

	s1 := mustCreateSubmission(t, e, ownerID)
	time.Sleep(5 * time.Millisecond)
	s2 := mustCreateSubmission(t, e, ownerID)
	mustCreateSubmission(t, e, otherID) // must not appear in owner's list

	got, err := repo.FindByOwner(ctx, ownerID)
	if err != nil {
		t.Fatalf("FindByOwner: %v", err)
	}
	if len(got) != 2 {
		t.Fatalf("got %d submissions, want 2", len(got))
	}
	// Newest first.
	if got[0].ID != s2.ID || got[1].ID != s1.ID {
		t.Fatalf("submissions not sorted newest first")
	}
}

func TestViewIdempotencyUnderConcurrency(t *testing.T) {
	e := setup(t)
	stop := startWorker(t, e)
	defer stop()

	ctx := context.Background()
	acc := registerAndLogin(t, e, "views@example.com")
	ownerID, _ := primitive.ObjectIDFromHex(acc.user.ID.Hex())
	sub := mustCreateSubmission(t, e, ownerID)

	linkDTO, err := e.shareSvc.Create(ctx, ownerID, sub.ID)
	if err != nil {
		t.Fatalf("Create share link: %v", err)
	}
	link, _ := e.shareSvc.ResolvePublic(ctx, linkDTO.Code)

	// Many concurrent events with the SAME viewId must count once.
	const n = 10
	errs := make(chan error, n)
	for i := 0; i < n; i++ {
		go func() {
			errs <- e.trackingSvc.RecordView(ctx, linkDTO.Code, "22222222-2222-4222-8222-222222222222", "127.0.0.1", "test")
		}()
	}
	for i := 0; i < n; i++ {
		if err := <-errs; err != nil {
			t.Fatalf("RecordView: %v", err)
		}
	}

	waitForViews(t, e, link.ID, 1)
}
