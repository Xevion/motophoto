# Style Guide

Cross-cutting conventions shared between Go and SvelteKit. Language-specific details live in [GO.md](GO.md) and [SVELTE.md](SVELTE.md).

## Naming

### Vocabulary

Use consistent terms across the entire codebase — backend, frontend, database, comments, commit messages. See [VOCABULARY.md](VOCABULARY.md) for the full glossary.

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

### Don't Say / Say Instead

| Don't say | Say instead | Why |
|-----------|-------------|-----|
| image | photo | Domain term — we sell photos, not generic images |
| user (when specific) | photographer, customer | Be specific about the role |
| album | gallery | Domain term — galleries belong to events |
| tournament, game | event | Canonical term for any sporting occasion |
| price, cost | price (in cents) | Always clarify the unit — `price_cents` in code |

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
// Wrap errors with context
if err != nil {
    return fmt.Errorf("fetching event %s: %w", id, err)
}
```

### TypeScript

```typescript
// Throw descriptive errors in load functions
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
/api/v1/{resource}          # Collection
/api/v1/{resource}/{id}     # Individual item
/health                     # Health check (no versioning)
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
- Run `task test` before pushing — CI will catch failures, but it's faster to catch locally
