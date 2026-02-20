# Go Backend

The Go backend is a JSON API server built with Chi, serving the SvelteKit frontend and handling all data operations. See [ARCHITECTURE.md](ARCHITECTURE.md) for overall system design and project structure.

## Entry Point

`main.go` does three things:

1. Loads `.env` via godotenv
2. Creates a structured JSON logger (`slog`)
3. Creates the server and starts it with graceful shutdown on SIGINT/SIGTERM

## Router & Middleware

The server uses Chi (`go-chi/chi/v5`) with this middleware stack (order matters):

1. **RequestID** — unique ID per request
2. **RealIP** — trust `X-Forwarded-For` headers
3. **Logger** — request/response logging
4. **Recoverer** — panic recovery → 500 instead of crash
5. **Compress(5)** — gzip responses
6. **CORS** — allows `localhost:5173` and `localhost:3000` origins

### Current Routes

```
GET  /api/health          → {"status": "ok"}
GET  /api/v1/events       → list of demo events
GET  /api/v1/events/{id}  → single demo event
```

These currently return hardcoded data. The database layer will replace this.

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

## Database & sqlc

The database layer uses **pgx/v5** (connection pool) with **sqlc** for type-safe SQL.

> **Current status**: The schema and queries exist but aren't wired into the API handlers yet. Endpoints return hardcoded demo data.

### sqlc Workflow

When you need to add or change database queries:

1. Write or edit SQL in `internal/database/queries/*.sql`
2. Run `just generate` to regenerate Go code in `internal/database/db/`
3. Use the generated functions in your handlers
4. **Never hand-edit files in `internal/database/db/`** — they're overwritten on every generate

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

SQL migrations live in `internal/database/migrations/` with numeric prefixes (`001_`, `002_`, etc.). These are applied manually for now — no migration runner is configured yet.

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

## Testing

```bash
just test             # Run all Go tests
go test ./...         # Equivalent
go test ./internal/server/...  # Test specific package
```

Test files live next to the code they test: `server_test.go` alongside `server.go`.
