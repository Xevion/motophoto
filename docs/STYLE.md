# Style Guide

Cross-cutting conventions shared between Go and SvelteKit. Language-specific details live in [GO.md](GO.md) and [SVELTE.md](SVELTE.md).

## Vocabulary

Use consistent terms across the entire codebase — backend, frontend, database, comments, commit messages, and UI text.

### Core Entities

| Term | Definition | Notes |
|------|-----------|-------|
| **Event** | A sporting occasion where photos are taken — motocross race, BMX competition, rodeo, swim meet, etc. | Top-level organizing entity. Has a date, location, and sport type. |
| **Gallery** | A collection of photos from a single event, shot by one photographer. | An event can have multiple galleries (one per photographer). |
| **Photo** | A single image captured at an event, available for purchase. | Has a storage key, optional watermark, dimensions, price, and metadata tags. |
| **Photographer** | A user who captures and uploads photos to galleries. | A user with `role = 'photographer'`. |
| **Customer** | A user who browses and purchases photos. | A user with `role = 'customer'`. Default role for new signups. |
| **Admin** | A user with platform management privileges. | A user with `role = 'admin'`. |
| **Tag** | A key-value pair attached to a photo for searchability. | Examples: `rider_number: 42`, `color: red`, `position: 1st`. Stored in `photo_tags`. |

### Supporting Concepts

| Term | Definition | Notes |
|------|-----------|-------|
| **Sport** | The type of athletic activity at an event. | Values: `motocross`, `bmx`, `rodeo`, `swimming`, etc. Stored as text, not an enum. |
| **Watermark** | A visual overlay on a photo to prevent unpaid use. | Stored as a separate `watermarked_key` alongside the original `storage_key`. |
| **Storage key** | The identifier for a photo file in object storage. | Opaque string — the storage backend (S3, R2, local) determines the actual URL. |
| **Price** | The cost to purchase a photo, in **cents** (USD). | Always stored as integer cents (`price_cents`) to avoid floating-point issues. |

### Entity Relationships

```
User (photographer) ──creates──► Event
                     ──creates──► Gallery ──belongs to──► Event
                                  Gallery ──contains──► Photo
                                                         Photo ──has many──► Tag

User (customer) ──browses──► Event ──► Gallery ──► Photo
                ──purchases──► Photo
```

### Anti-Patterns

| Don't say | Say instead | Why |
|-----------|-------------|-----|
| image, picture | photo | We sell photos, not generic images |
| album, folder, collection | gallery | Galleries are tied to events and photographers |
| user (when role matters) | photographer, customer, admin | Be specific about which role |
| tournament, game, match, contest, competition | event | Single canonical term for any occasion |
| price (ambiguous) | price in cents, `price_cents` | Always clarify the unit |
| label | tag | Tags are key-value pairs on photos |

## Naming

| Context | Convention | Example |
|---------|-----------|---------|
| Go packages | lowercase, single word | `server`, `database`, `middleware` |
| Go exported names | PascalCase | `ListEvents`, `NewServer` |
| Go unexported names | camelCase | `setupRoutes`, `demoEvents` |
| TypeScript variables/functions | camelCase | `apiFetch`, `eventId` |
| TypeScript types/interfaces | PascalCase | `Event`, `ApiResponse` |
| Svelte components | PascalCase files | `EventCard.svelte` |
| SvelteKit routes | kebab-case directories | `events/[id]/` |
| SQL tables | snake_case, plural | `photo_tags`, `galleries` |
| SQL columns | snake_case | `event_date`, `price_cents` |
| JSON keys | snake_case | `{"event_date": "..."}` |
| URL paths | kebab-case | `/api/v1/events` |
| Environment variables | SCREAMING_SNAKE | `DATABASE_URL`, `PORT` |

## Comments

- Explain **why**, not **what** — the code shows what it does
- Don't reference old implementations, migrations, or "this was refactored from..."
- Don't add banner comments (`// ===`, `// ---`, section separators)
- TODO comments must include context: `// TODO: handle pagination once we have >100 events`

```go
// BAD
// This function gets events from the database
func ListEvents() {}

// GOOD
// ListEvents returns all events, ordered by date descending.
// Only upcoming events are included — past events are filtered server-side
// to avoid exposing stale gallery data.
func ListEvents() {}
```

## Error Handling

### Principles

1. **Wrap with context** — every error should describe what failed, not just that something failed
2. **Handle at the boundary** — functions return errors, HTTP handlers translate them to status codes
3. **Don't swallow errors** — if you catch an error, log it or return it. Never `_ = err`

### Go

```go
if err != nil {
    return fmt.Errorf("fetching event %s: %w", id, err)
}
```

### TypeScript

```typescript
if (!response.ok) {
    throw error(response.status, `Failed to load event: ${response.statusText}`);
}
```

## Logging

Structured logging via Go's `log/slog`. JSON format in production, text in development.

| Level | Use for |
|-------|---------|
| `slog.Debug` | Verbose data useful only when debugging a specific issue |
| `slog.Info` | Normal operations — server start, request handling, config loaded |
| `slog.Warn` | Recoverable problems — missing optional config, degraded service |
| `slog.Error` | Failures that need attention — DB connection lost, handler panics |

Always include relevant context as structured fields:

```go
slog.Info("event fetched", "event_id", id, "duration_ms", elapsed)
slog.Error("database query failed", "error", err, "query", "ListEvents")
```

## API Design

### URL Structure

```
/api/health                 # Health check (no versioning)
/api/v1/{resource}          # Collection
/api/v1/{resource}/{id}     # Individual item
```

### Response Format

All API responses return JSON. Collections return arrays, individual items return objects.

```json
// Collection
[{"id": "...", "name": "..."}]

// Individual
{"id": "...", "name": "..."}

// Error
{"error": "description of what went wrong"}
```

### HTTP Status Codes

| Status | When |
|--------|------|
| 200 | Successful GET |
| 201 | Successful POST (resource created) |
| 400 | Invalid request (bad params, missing fields) |
| 404 | Resource not found |
| 500 | Server error (log it, don't expose internals) |

## Git

- Commit messages use conventional commit style: `feat:`, `fix:`, `chore:`, `docs:`, etc.
- Scale message length to change impact — a rename doesn't need a paragraph
- Default branch is `master`

## Testing

- Go tests live alongside the code they test (`foo_test.go` next to `foo.go`)
- Frontend tests (when added) go in `web/src/**/*.test.ts`
- Run `just test` before pushing — CI will catch failures, but it's faster to catch locally
