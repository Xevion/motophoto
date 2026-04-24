-- name: CreatePhoto :one
INSERT INTO photos (id, gallery_id, storage_key, preview_key, filename, content_type, size_bytes, status)
VALUES ($1, $2, $3, $4, $5, $6, $7, 'pending')
RETURNING *;

-- name: GetPhoto :one
SELECT * FROM photos
WHERE id = $1 AND gallery_id = $2;

-- name: GetPhotoTimeRange :one
SELECT 
    MIN(taken_at) as min_taken_at,
    MAX(taken_at) as max_taken_at
FROM photos
WHERE gallery_id = $1 AND deleted_at IS NULL AND status = 'ready' AND taken_at IS NOT NULL;

-- name: ConfirmPhoto :one
UPDATE photos SET
    width      = $3,
    height     = $4,
    size_bytes = $5,
    taken_at   = $6,
    status     = 'ready',
    updated_at = now()
WHERE id = $1 AND gallery_id = $2 AND status = 'pending'
RETURNING *;

-- name: ListPhotosByGallery :many
SELECT *
FROM photos
WHERE gallery_id = $1 
    AND deleted_at IS NULL 
    AND status = 'ready'
    AND (sqlc.narg('taken_after')::timestamptz IS NULL OR taken_at >= sqlc.narg('taken_after')::timestamptz)
    AND (sqlc.narg('taken_before')::timestamptz IS NULL OR taken_at <= sqlc.narg('taken_before')::timestamptz)
ORDER BY sort_order, id;
