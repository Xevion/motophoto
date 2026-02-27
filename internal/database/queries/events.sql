-- name: ListEvents :many
SELECT e.*,
       (SELECT count(*) FROM photos p
        JOIN galleries g ON g.id = p.gallery_id
        WHERE g.event_id = e.id AND p.deleted_at IS NULL)::bigint AS photo_count
FROM events e
WHERE e.status = 'published'
  AND (sqlc.narg('cursor_sort_order')::int IS NULL
       OR (e.sort_order, e.id) > (sqlc.narg('cursor_sort_order')::int, sqlc.narg('cursor_id')::text))
ORDER BY e.sort_order, e.id
LIMIT sqlc.arg('limit_val')::int;

-- name: GetEvent :one
SELECT e.*,
       (SELECT count(*) FROM photos p
        JOIN galleries g ON g.id = p.gallery_id
        WHERE g.event_id = e.id AND p.deleted_at IS NULL)::bigint AS photo_count
FROM events e
WHERE e.id = sqlc.arg('id_or_slug') OR e.slug = sqlc.arg('id_or_slug');

-- name: CreateEvent :one
INSERT INTO events (id, photographer_id, slug, name, sport, location, description, tags, status, date, sort_order)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
RETURNING *;

-- name: UpdateEvent :one
UPDATE events SET
    name        = COALESCE(sqlc.narg('name'), name),
    slug        = COALESCE(sqlc.narg('slug'), slug),
    sport       = COALESCE(sqlc.narg('sport'), sport),
    location    = COALESCE(sqlc.narg('location'), location),
    description = COALESCE(sqlc.narg('description'), description),
    tags        = COALESCE(sqlc.narg('tags'), tags),
    status      = COALESCE(sqlc.narg('status'), status),
    date        = COALESCE(sqlc.narg('date'), date),
    sort_order  = COALESCE(sqlc.narg('sort_order'), sort_order),
    updated_at  = now()
WHERE id = sqlc.arg('id')
RETURNING *;

-- name: DeleteEvent :exec
DELETE FROM events WHERE id = $1;

-- name: GetEventOwner :one
SELECT photographer_id FROM events WHERE id = $1;
