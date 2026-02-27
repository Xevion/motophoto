# Go Backend

The Go backend is a JSON API server built with Chi, serving the SvelteKit frontend and handling all data operations. See [ARCHITECTURE.md](ARCHITECTURE.md) for overall system design and project structure.

## Entry Point

`main.go` does five things:

1. Loads `.env` via godotenv (ignored in production where env vars are set directly)
2. Initialises logging -- tint pretty-printer in development, JSON via `slog.NewJSONHandler` when `LOG_JSON=true`; log level controlled by `LOG_LEVEL`
3. Opens the pgx connection pool and runs goose migrations
4. Creates the session manager (scs backed by PostgreSQL)
5. Creates the server and starts it with graceful shutdown on SIGINT/SIGTERM

## Router & Middleware

The server uses Chi (`go-chi/chi/v5`) with this middleware stack (order matters):

1. **RequestID** -- reads `X-Railway-Request-Id` first (Railway's edge proxy header), falls back to `X-Request-Id`, then generates a UUID; stored in context via chi's `RequestIDKey`
2. **RealIP** -- chi's `chimw.RealIP`; checks `True-Client-IP`, then `X-Real-IP`, then `X-Forwarded-For` (in that order). In production the proxy chain is Client → Cloudflare → Fastly (Railway) → SvelteKit → Go, so the backend's network peer is SvelteKit. `hooks.server.ts` forwards `True-Client-IP` (set by Cloudflare to the real client) so chi resolves the correct address
3. **RequestLogger** -- logs each response with method, path, status, and duration; routes to Debug/Warn/Error by status code; stores a `request_id`-tagged logger in context for use by handlers
4. **Recoverer** -- panic recovery -> 500 instead of crash
5. **RateLimiter** -- 100 requests/minute per real IP (`httprate`)
6. **CORS** -- allows `localhost:5173` and `localhost:3000` origins; credentials enabled
7. **Compress(5)** -- gzip responses
8. **SessionManager** -- loads and saves the scs session for each request

### Current Routes

```
GET    /api/health                              -> {"status": "ok"}
GET    /api/v1/events                           -> list published events (cursor-paginated)
POST   /api/v1/events                           -> create event
GET    /api/v1/events/{id}                      -> get event by nanoid or slug (includes galleries)
PATCH  /api/v1/events/{id}                      -> partial update event
DELETE /api/v1/events/{id}                      -> delete event
GET    /api/v1/events/{eventId}/galleries       -> list galleries for event
POST   /api/v1/events/{eventId}/galleries       -> create gallery
PATCH  /api/v1/events/{eventId}/galleries/{id}  -> partial update gallery
DELETE /api/v1/events/{eventId}/galleries/{id}  -> delete gallery
```

All data endpoints use real database queries via sqlc. List endpoints return a `{"data": [...], "next_cursor": "..."}` envelope. Single-item endpoints return `{"data": {...}}`.

## Adding a New Endpoint

Handlers live in resource-specific files under `internal/server/` (e.g., `events.go`, `galleries.go`). They are methods on `*Server` so they can access the service layer and session manager.

1. Add the handler method to the appropriate file:

```go
func (s *Server) handleCreatePhoto(w http.ResponseWriter, r *http.Request) {
    // Decode and size-limit the request body
    r.Body = http.MaxBytesReader(w, r.Body, maxRequestBodySize)
    var req CreatePhotoRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        writeError(w, http.StatusBadRequest, "invalid request body")
        return
    }

    // Validate struct tags
    if err := validate.Struct(req); err != nil {
        writeError(w, http.StatusBadRequest, err.Error())
        return
    }

    // Call the service layer
    photo, err := s.photos.Create(r.Context(), req.GalleryID, req.PriceCents)
    if err != nil {
        writeServiceError(w, r, err, "create photo")
        return
    }

    writeJSON(w, http.StatusCreated, ItemResponse[PhotoResponse]{Data: photoResponseFromService(photo)})
}
```

2. Register it in `setupRoutes()` in `server.go`:

```go
r.Route("/api/v1", func(r chi.Router) {
    r.Get("/events", s.handleListEvents)
    r.Post("/events/{eventId}/galleries/{id}/photos", s.handleCreatePhoto)  // new
})
```

### Handler Pattern

Every handler follows the same structure:

1. Extract parameters (URL params, query string, request body)
2. Validate inputs
3. Call the service layer
4. Return JSON response with appropriate status code

### URL Parameters

```go
id := chi.URLParam(r, "id")     // from route pattern {id}
sport := r.URL.Query().Get("sport")  // from query string ?sport=motocross
```

### JSON Responses

Use the `writeJSON`, `writeError`, and `writeServiceError` helpers -- never write to `w` directly:

```go
writeJSON(w, http.StatusOK, ItemResponse[EventResponse]{Data: eventResponseFromService(e)})
writeError(w, http.StatusBadRequest, "name is required")
writeServiceError(w, r, err, "create event")  // translates service errors to HTTP codes
```

`writeServiceError` maps `service.ErrNotFound` -> 404, `service.ErrConflict` -> 409, and logs + returns 500 for anything else.

## Service Layer

Business logic lives in `internal/service/`, sitting between the HTTP handlers and the database. Handlers call services; services call sqlc-generated queries.

```
Handler -> Service -> sqlc queries -> Postgres
```

Each service is instantiated once in `server.New()` and stored on the `Server` struct:

```go
s.events    *service.EventService
s.galleries *service.GalleryService
```

### Service Types

Services define their own plain-Go types (no pgx/pgtype leaking into handlers):

```go
type Event struct {
    ID          string
    Slug        string
    Name        string
    Sport       string
    Status      string
    Location    *string
    Description *string
    Date        *string
    Tags        []string
    PhotoCount  int64
}
```

Conversion from sqlc-generated types to service types happens inside the service. Conversion from service types to HTTP response types happens in the handler (via `eventResponseFromService`, etc.).

### Service Errors

`internal/service/errors.go` defines sentinel errors that handlers translate to HTTP status codes via `writeServiceError`:

| Error              | HTTP status |
| ------------------ | ----------- |
| `service.ErrNotFound` | 404 Not Found |
| `service.ErrConflict` | 409 Conflict |
| anything else      | 500 Internal Server Error |

## Sessions

Session management uses **scs** (`alexedwards/scs/v2`) backed by a PostgreSQL store (`scs/pgxstore`). The session manager is created in `internal/session/session.go` and injected into the server.

Key settings:

| Setting | Value |
|---------|-------|
| Lifetime | 24 hours |
| Idle timeout | 30 minutes |
| Cookie name | `session_id` |
| SameSite | Lax |
| HttpOnly | true |

The store uses the `sessions` table (created by migration `001_create_sessions.sql`). An index on `expiry` ensures expired session cleanup stays fast.

## Database & sqlc

The database layer uses **pgx/v5** (connection pool) with **sqlc** for type-safe SQL.

### sqlc Workflow

When you need to add or change database queries:

1. Write or edit SQL in `internal/database/queries/*.sql`
2. Run `just generate` to regenerate Go code in `internal/database/db/`
3. Use the generated functions in your handlers
4. **Never hand-edit files in `internal/database/db/`** -- they're overwritten on every generate

### Query File Format

sqlc queries follow a specific annotation format:

```sql
-- name: GetEvent :one
SELECT * FROM events WHERE id = $1;

-- name: ListEvents :many
SELECT * FROM events ORDER BY event_date DESC;

-- name: CreateEvent :one
INSERT INTO events (name, description, sport, location, event_date, created_by)
VALUES ($1, $2, $3, $4, $5, $6)
RETURNING *;
```

The annotation (`-- name: ... :one/:many/:exec`) tells sqlc what Go function to generate and whether it returns one row, many rows, or no rows.

### Migrations

SQL migrations live in `internal/database/migrations/` in goose format. They run automatically at server startup via `database.Migrate()` in `main.go` -- no manual steps needed. Files use goose's `-- +goose Up` / `-- +goose Down` annotations and numeric prefixes (`001_`, `002_`, etc.).

## Error Handling

Always wrap errors with context describing what failed:

```go
if err != nil {
    return fmt.Errorf("fetching event %s: %w", id, err)
}
```

In handlers, translate errors to HTTP status codes at the boundary:

```go
event, err := queries.GetEvent(r.Context(), id)
if err != nil {
    if errors.Is(err, pgx.ErrNoRows) {
        http.Error(w, `{"error": "event not found"}`, http.StatusNotFound)
        return
    }
    slog.Error("failed to get event", "error", err, "event_id", id)
    http.Error(w, `{"error": "internal server error"}`, http.StatusInternalServerError)
    return
}
```

## Key Libraries

Beyond Chi, pgx, and scs, the following packages are available for use:

| Package | Purpose |
|---------|---------|
| `go-playground/validator/v10` | Struct field validation via struct tags -- use for validating request bodies |
| `aws-sdk-go-v2` + `service/s3` | S3-compatible object storage (AWS S3, Cloudflare R2) for photo uploads |
| `disintegration/imaging` | Image resizing, cropping, and format conversion for thumbnail/watermark generation |
| `gabriel-vasile/mimetype` | MIME type detection from file content (not file extension) |
| `google/uuid` | UUID generation for photo and entity IDs |
| `stretchr/testify` | Assertion helpers for Go tests |

### Request Validation

Use `go-playground/validator` for validating decoded request bodies:

```go
import "github.com/go-playground/validator/v10"

var validate = validator.New()

type CreatePhotoRequest struct {
    GalleryID uuid.UUID `json:"gallery_id" validate:"required"`
    Price     int64     `json:"price_cents" validate:"required,min=0"`
}

func handleCreatePhoto(w http.ResponseWriter, r *http.Request) {
    var req CreatePhotoRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
        return
    }
    if err := validate.Struct(req); err != nil {
        writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
        return
    }
    // ...
}
```

## Testing

```bash
just test             # Run all Go tests
go test ./...         # Equivalent
go test ./internal/server/...  # Test specific package
```

### Integration Tests with pgtestdb

Handler and service tests are integration tests that run against a real Postgres instance. [pgtestdb](https://github.com/peterldowns/pgtestdb) creates an isolated, migrated database per test using template-database cloning -- migrations run once and are cached across the test suite.

Tests use `testutil.NewEnv(t)`, which wires up a pool, queries, services, and an HTTP handler against the isolated database:

```go
func TestHandleListEvents(t *testing.T) {
    t.Parallel()
    env := testutil.NewEnv(t)

    // Create fixtures via the dbfactory helpers
    event := dbfactory.Event(ctx, t, env.Pool, env.Events, &dbfactory.EventOpts{
        Status: new("published"),
    })

    rr := doRequest(t, env.Handler, http.MethodGet, "/api/v1/events", "")
    assert.Equal(t, http.StatusOK, rr.Code)
}
```

`testutil.NewEnv` uses the local docker-compose Postgres on port 57512. In CI, `CI=true` switches it to the standard port 5432.

Test files use the `_test` package suffix (`server_test`, `service_test`) and live alongside the code they test.

## Field Alignment

Go struct fields are padded to satisfy alignment requirements, which can waste memory if fields are ordered carelessly. The `fieldalignment` tool from `golang.org/x/tools` detects structs with suboptimal layout and can reorder fields automatically.

Install:

```bash
go install golang.org/x/tools/go/analysis/passes/fieldalignment/cmd/fieldalignment@latest
```

Check for alignment issues:

```bash
fieldalignment ./...
```

Auto-fix by reordering fields:

```bash
fieldalignment -fix ./...
```

The tool rewrites the struct field order in-place to minimize padding. It does not change field names, types, or tags -- only order. Re-run after adding new fields to a struct, since the optimal order changes as the struct evolves.

Do not apply `-fix` blindly to generated files (e.g. `internal/database/db/`) -- those are overwritten by `just generate`.
