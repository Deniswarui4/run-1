-- name: GetEventByID :one
SELECT id, title, description, category, venue, address, city, country, image_url,
       start_date, end_date, status, is_featured, organizer_id, moderator_id,
       moderation_comment, moderated_at, created_at, updated_at, deleted_at
FROM events
WHERE id = $1 AND deleted_at IS NULL;

-- name: ListEvents :many
SELECT id, title, description, category, venue, address, city, country, image_url,
       start_date, end_date, status, is_featured, organizer_id, moderator_id,
       moderation_comment, moderated_at, created_at, updated_at, deleted_at
FROM events
WHERE deleted_at IS NULL
ORDER BY created_at DESC
LIMIT $1 OFFSET $2;

-- name: CreateEvent :one
INSERT INTO events (
    title, description, category, venue, address, city, country,
    image_url, start_date, end_date, status, is_featured, organizer_id
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13
)
RETURNING id, title, description, category, venue, address, city, country, image_url,
          start_date, end_date, status, is_featured, organizer_id, moderator_id,
          moderation_comment, moderated_at, created_at, updated_at, deleted_at;

-- name: UpdateEventStatus :one
UPDATE events
SET
    status             = $2,
    moderator_id       = sqlc.narg(moderator_id),
    moderation_comment = sqlc.narg(moderation_comment),
    moderated_at       = sqlc.narg(moderated_at),
    updated_at         = NOW()
WHERE id = $1 AND deleted_at IS NULL
RETURNING id, title, description, category, venue, address, city, country, image_url,
          start_date, end_date, status, is_featured, organizer_id, moderator_id,
          moderation_comment, moderated_at, created_at, updated_at, deleted_at;
