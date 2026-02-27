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

1. **RequestID** -- unique ID per request
2. **RealIP** -- trust `X-Forwarded-For` headers
3. **Logger** -- request/response logging
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

1. Define the handler function in `internal/server/server.go`:

```go
func handleCreateEvent(w http.ResponseWriter, r *http.Request) {
    // Decode request body
    var req CreateEventRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        http.Error(w, `{"error": "invalid request body"}`, http.StatusBadRequest)
        return
    }

    // Validate
    if req.Name == "" {
        http.Error(w, `{"error": "name is required"}`, http.StatusBadRequest)
        return
    }

    // Call database (once wired up)
    // event, err := queries.CreateEvent(r.Context(), db.CreateEventParams{...})

    // Return JSON
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(http.StatusCreated)
    json.NewEncoder(w).Encode(event)
}
```

2. Register it in `setupRoutes()`:

```go
r.Route("/api/v1", func(r chi.Router) {
    r.Get("/events", handleListEvents)
    r.Post("/events", handleCreateEvent)  // new
    r.Get("/events/{id}", handleGetEvent)
})
```

### Handler Pattern

Every handler follows the same structure:

1. Extract parameters (URL params, query string, request body)
2. Validate inputs
3. Call business logic / database
4. Return JSON response with appropriate status code

### URL Parameters

```go
id := chi.URLParam(r, "id")     // from route pattern {id}
sport := r.URL.Query().Get("sport")  // from query string ?sport=motocross
```

### JSON Responses

```go
w.Header().Set("Content-Type", "application/json")
json.NewEncoder(w).Encode(data)
```

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

Test files live next to the code they test: `server_test.go` alongside `server.go`.

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
