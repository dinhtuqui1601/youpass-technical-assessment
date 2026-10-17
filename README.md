# YouPass Technical Assignment

Monorepo for the **YouPass Technical Assignment**, covering two modules:

1. **URL Shortener** — Golang backend to manage share links for learner submissions,
   access control and view statistics.
2. **Seamless Authentication** — JWT access tokens + opaque rotating refresh tokens
   so the ReactJS frontend can refresh silently and keep the learner's session alive.

## Repository structure

```
youpass/
├── front-end/          # ReactJS + TypeScript + Vite + Ant Design 6
├── backend/            # Golang + Gin + MongoDB + Redis
├── deploy/             # Docker Compose, Nginx, environment config
├── .gitignore
└── README.md
```

```
backend/
├── cmd/
│   ├── api/            # HTTP API entrypoint
│   └── worker/         # View-tracking worker entrypoint
├── internal/
│   ├── auth/           # JWT + refresh-token rotation, sessions
│   ├── users/          # User model & persistence
│   ├── submissions/    # Minimal submission module (demo)
│   ├── share/          # URL shortener: short codes, CRUD, redirect, access control
│   ├── tracking/       # View events (Redis Streams) + stats + worker
│   └── middleware/     # CORS, request id, auth, rate limit, origin check
├── pkg/
│   ├── database/       # MongoDB connect + indexes
│   ├── cache/          # Redis cache helpers
│   ├── redis/          # Redis client helpers
│   ├── response/       # Unified response envelope & error codes
│   └── logger/         # slog JSON logger
├── configs/            # Environment configuration
├── docs/               # Generated OpenAPI / Swagger
├── tests/              # Integration tests (real Mongo + Redis)
├── go.mod
├── Dockerfile
└── .env.example
```

## Architecture

Modular monolith split by domain:

- **Handler** — HTTP only: bind/validate input, call service, write the unified response.
- **Service** — business logic (authentication, share links, tracking).
- **Repository** — MongoDB access.
- **Middleware** — cross-cutting concerns (CORS, request ID, JWT auth, rate limiting).

Dependencies are injected via constructors; all I/O uses `context.Context`.

### Unified response envelope

```json
{ "success": true, "data": {}, "error": null, "requestId": "uuid" }
```

Errors use machine-readable codes (see below). `GET /s/:code` returns a plain HTTP
302 redirect and is **not** wrapped in the envelope.

## API

| Method | Endpoint | Auth | Description |
|---|---|---|---|
| GET | `/health` | – | Health check |
| GET | `/swagger/*` | – | OpenAPI UI |
| POST | `/api/auth/register` | – | Register (demo) |
| POST | `/api/auth/login` | – | Login (access token + HttpOnly refresh cookie) |
| POST | `/api/auth/refresh` | cookie | Rotate refresh token, return new access token |
| POST | `/api/auth/logout` | cookie | Revoke session, clear cookie |
| GET | `/api/auth/me` | Bearer | Current user profile |
| GET | `/api/submissions` | Bearer | List my submissions (newest first) |
| POST | `/api/submissions` | Bearer | Create a submission (demo) |
| POST | `/api/submissions/:id/share` | Bearer | Create/reuse a share link |
| GET | `/api/share-links` | Bearer | List my share links |
| GET | `/api/share-links/:code` | Bearer | Get my link detail |
| PATCH | `/api/share-links/:code` | Bearer | Enable/disable a link |
| DELETE | `/api/share-links/:code` | Bearer | Soft-delete a link |
| GET | `/api/share-links/:code/stats` | Bearer | View statistics (owner only) |
| GET | `/s/:code` | – | Resolve short code → 302 redirect |
| GET | `/api/public/shares/:code` | – | Shared submission content |
| POST | `/api/public/shares/:code/views` | – | Record a view event |

### Error codes

| HTTP | Code | Meaning |
|---|---|---|
| 401 | `ACCESS_TOKEN_EXPIRED` | Access token expired (frontend triggers refresh) |
| 401 | `INVALID_ACCESS_TOKEN` | Access token failed verification |
| 401 | `MISSING_ACCESS_TOKEN` | No Bearer token |
| 401 | `REFRESH_TOKEN_EXPIRED` | Refresh token expired |
| 401 | `INVALID_REFRESH_TOKEN` | Refresh token invalid / reused |
| 401 | `SESSION_REVOKED` | Session family revoked |
| 403 | `FORBIDDEN` | Not allowed / not owner |
| 403 | `SHARE_LINK_DISABLED` / `SHARE_LINK_DELETED` / `SHARE_LINK_EXPIRED` | Link not accessible |
| 404 | `NOT_FOUND` / `SHARE_LINK_NOT_FOUND` / `SUBMISSION_NOT_FOUND` | Missing resource |
| 409 | `EMAIL_ALREADY_EXISTS` / `CONFLICT` | Conflict |
| 429 | `RATE_LIMITED` | Too many requests |
| 503 | `QUEUE_UNAVAILABLE` | View event could not be enqueued (retryable) |

## MongoDB collections & indexes

| Collection | Indexes |
|---|---|
| `users` | `email` (unique) |
| `auth_sessions` | `sessionId` (unique), `userId`, `refreshTokenHash` |
| `submissions` | `ownerId` |
| `share_links` | `code` (unique), `ownerId`, `submissionId` |
| `share_views` | `(shareLinkId, viewId)` (unique), `(shareLinkId, viewedAt desc)` |

Only the SHA-256 hash of the refresh token is stored, never the plaintext.

## Authentication design

- **Access token**: HS256 JWT, 15 min TTL (configurable). Claims: `sub`, `role`, `iat`,
  `exp`, `jti`, `iss`, `aud`. Returned in the JSON body; the frontend keeps it in memory.
- **Refresh token**: opaque 32-byte `crypto/rand` value, SHA-256 hash stored in
  `auth_sessions`, delivered via an **HttpOnly** cookie (path `/api/auth`). In production
  enable `COOKIE_SECURE=true` and HTTPS.
- **Rotation**: every refresh issues a brand-new refresh token. Rotation is atomic
  (compare-and-swap on the current hash), so the same token can never rotate twice.
- **Grace window** (`REFRESH_GRACE_PERIOD`, default 30s): tolerates two concurrent
  refreshes of the same token (e.g. two tabs) without treating them as reuse.
- **Reuse detection**: presenting a rotated-out token after the grace window revokes the
  whole token family.
- **Limitations (documented)**: a burst of ≥3 simultaneous refreshes of the *same* token
  can reject the stragglers with `INVALID_REFRESH_TOKEN` (never spurious revocation).
  The frontend single-flights refreshes per tab and falls back to re-login.

## URL shortener design

- Short codes are cryptographically-random **Base62** strings (default length 10),
  generated with `crypto/rand` + rejection sampling. A unique index on `code` plus a
  bounded retry loop handles collisions. ObjectIDs are never used as codes.
- `status`: `ACTIVE`, `DISABLED`, `DELETED` (soft delete keeps history; codes are never
  reissued). Re-creating a link for a submission returns the existing `ACTIVE` link.
- **Redirect** (`/s/:code`): validates the code, checks a short-lived Redis mapping
  (30s TTL), falls back to MongoDB on miss, and returns `302` → `/shared/:code` with
  `Cache-Control: no-store`.
- **Access control is authoritative in MongoDB**: the content endpoint
  `/api/public/shares/:code` always re-checks status, expiry and submission
  shareability (fail closed). Redis caches only the code mapping, never the final
  authorization decision. Status changes invalidate the cache key.

## View tracking design

- `POST /api/public/shares/:code/views` validates access, then enqueues a
  `{shareLinkId, viewId, viewedAt, ...}` event into the Redis Stream `view_events`.
  If the queue is unavailable the API returns `503 QUEUE_UNAVAILABLE` (never silently
  drops the event).
- The worker consumes `view_events` via a **consumer group** with a **bounded worker
  pool**, **ACK after success**, **bounded retries**, a **dead-letter stream**
  (`view_events_dlq`), periodic pending-entry reclamation and graceful shutdown.
- **Idempotency / correctness**: events are persisted to `share_views` with a unique
  index on `(shareLinkId, viewId)`, so retries and duplicate `viewId`s never double
  count. `totalViews` is computed from this collection (source of truth); the duplicate
  path is treated as success and ACKed.

## Prerequisites

- Node.js 20.19+ / 22.12+ (tested with 22.x)
- Go 1.27+
- Docker & Docker Compose v2

## Running

### Frontend

```bash
cd front-end
npm install
npm run dev            # http://localhost:5173
```

### Backend API

```bash
cd backend
go run ./cmd/api       # http://localhost:8080
curl http://localhost:8080/health   # {"status":"ok"}
```

OpenAPI UI: <http://localhost:8080/swagger/index.html>

### Worker (view tracking)

```bash
cd backend
go run ./cmd/worker
```

### Docker Compose (full stack)

```bash
cd deploy
cp .env.example .env
docker compose up --build
```

- Nginx reverse proxy: <http://localhost:8080>
- `/api/*` → backend, `/s/*` → backend, `/swagger/*` → backend, everything else → SPA
- Services: `mongo`, `redis`, `backend`, `worker`, `frontend`, `nginx`

## Testing

```bash
cd backend

# Unit tests
go test ./internal/...

# Unit tests with race detector
go test -race ./internal/auth/... ./internal/share/...

# Integration tests (require running MongoDB + Redis; see tests/integration_test.go)
go test ./tests/...
```

Integration tests cover the full auth flow, share flow (create → open → track →
disable → verify denied), ownership enforcement, and view-idempotency under
concurrency. They use a dedicated database/Redis DB and clean up after themselves.

## Example curl flow

```bash
BASE=http://localhost:8080

# 1. Register + login (cookie jar stores the refresh token)
curl -s -c cookies.txt -X POST $BASE/api/auth/register \
  -H 'Content-Type: application/json' \
  -d '{"email":"demo@youpass.vn","password":"password123","name":"Demo"}'

curl -s -c cookies.txt -X POST $BASE/api/auth/login \
  -H 'Content-Type: application/json' \
  -d '{"email":"demo@youpass.vn","password":"password123"}'

# 2. Access a protected route (use the returned accessToken)
curl -s $BASE/api/auth/me -H "Authorization: Bearer $ACCESS_TOKEN"

# 3. Refresh (uses the cookie)
curl -s -b cookies.txt -c cookies.txt -X POST $BASE/api/auth/refresh

# 4. Create a submission + share link
curl -s -X POST $BASE/api/submissions \
  -H "Authorization: Bearer $ACCESS_TOKEN" -H 'Content-Type: application/json' \
  -d '{"title":"My work","content":"Hello YouPass"}'

curl -s -X POST $BASE/api/submissions/$SUBMISSION_ID/share \
  -H "Authorization: Bearer $ACCESS_TOKEN"

# 5. Open the short link (302) then read shared content
curl -s -i $BASE/s/$CODE
curl -s $BASE/api/public/shares/$CODE

# 6. Record a view, then check stats
curl -s -X POST $BASE/api/public/shares/$CODE/views \
  -H 'Content-Type: application/json' -d '{"viewId":"11111111-1111-4111-8111-111111111111"}'

curl -s $BASE/api/share-links/$CODE/stats -H "Authorization: Bearer $ACCESS_TOKEN"

# 7. Disable the link (owner only)
curl -s -X PATCH $BASE/api/share-links/$CODE \
  -H "Authorization: Bearer $ACCESS_TOKEN" -H 'Content-Type: application/json' \
  -d '{"status":"DISABLED"}'
```

## Environment variables

- `backend/.env.example` — port, MongoDB, Redis, CORS, JWT, refresh TTL, cookie, rate limits, worker concurrency.
- `deploy/.env.example` — Compose-level ports, credentials, JWT secret.
- `front-end/.env.example` — API base URL.

No real credentials are committed; copy the `*.env.example` files to `.env` and adjust.
