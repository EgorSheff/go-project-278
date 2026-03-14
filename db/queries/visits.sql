-- name: ListVisits :many
SELECT * FROM visits ORDER BY id LIMIT $1 OFFSET $2;

-- name: CountVisits :one
SELECT count(*) from visits;

-- name: CreateVisit :one
INSERT INTO visits (link_id, ip, user_agent, status) VALUES ($1, $2, $3, $4) RETURNING *;
