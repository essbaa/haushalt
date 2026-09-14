-- name: ListOccasions :many
SELECT * FROM occasion WHERE household_id = $1 ORDER BY day, title;

-- name: CreateOccasion :one
INSERT INTO occasion (household_id, title, day, kind, yearly)
VALUES ($1, $2, $3, $4, $5)
RETURNING *;

-- name: DeleteOccasion :exec
-- Mit Haushalt gefragt, nicht nur mit der Kennung: Eine Kennung allein wäre
-- eine Berechtigungsprüfung, die an einem vergessenen if hängt.
DELETE FROM occasion WHERE id = $1 AND household_id = $2;
