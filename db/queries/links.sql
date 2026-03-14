-- name: CreateLink :one
INSERT INTO links (original_url, short_name) VALUES ($1, $2) RETURNING *;

-- name: GetLink :one
SELECT * FROM links WHERE id = $1;

-- name: ListLinks :many
SELECT * FROM links ORDER BY id LIMIT $1 OFFSET $2;

-- name: UpdateLink :one
UPDATE links SET original_url = $1, short_name = $2 WHERE id = $3 RETURNING *;

-- name: DeleteLink :exec
DELETE FROM links WHERE id = $1;

-- name: CountLinks :one
SELECT count(*) from links;

-- name: GetLinkByShortName :one
SELECT * from links WHERE short_name = $1;
