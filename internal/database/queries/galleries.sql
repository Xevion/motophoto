-- name: ListGalleriesByEvent :many
SELECT g.*,
       (SELECT count(*) FROM photos p WHERE p.gallery_id = g.id AND p.deleted_at IS NULL)::bigint AS photo_count
FROM galleries g
WHERE g.event_id = $1
ORDER BY g.sort_order, g.id;

-- name: GetGallery :one
SELECT g.*,
       (SELECT count(*) FROM photos p WHERE p.gallery_id = g.id AND p.deleted_at IS NULL)::bigint AS photo_count
FROM galleries g
WHERE g.id = $1 AND g.event_id = $2;

-- name: CreateGallery :one
INSERT INTO galleries (id, event_id, slug, name, description, sort_order)
VALUES ($1, $2, $3, $4, $5, $6)
RETURNING *;

-- name: UpdateGallery :one
UPDATE galleries SET
    name        = COALESCE(sqlc.narg('name'), name),
    slug        = COALESCE(sqlc.narg('slug'), slug),
    description = COALESCE(sqlc.narg('description'), description),
    sort_order  = COALESCE(sqlc.narg('sort_order'), sort_order),
    updated_at  = now()
WHERE id = sqlc.arg('id') AND event_id = sqlc.arg('event_id')
RETURNING *;

-- name: DeleteGallery :exec
DELETE FROM galleries WHERE id = $1 AND event_id = $2;
