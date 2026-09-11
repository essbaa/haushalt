-- name: CreateHousehold :one
INSERT INTO household (name, home, has_car, has_yard, pets, timezone)
VALUES ($1, $2, $3, $4, $5, $6)
RETURNING *;

-- name: GetHousehold :one
SELECT * FROM household WHERE id = $1;

-- name: ListHouseholdsForAuthUser :many
-- Alle Haushalte, in denen diese angemeldete Person Mitglied ist.
-- Der Weg führt immer über member: Es gibt keine Beziehung zwischen einem
-- Login und einem Haushalt, die nicht über eine Person läuft.
SELECT h.*
FROM household h
JOIN member m ON m.household_id = h.id
WHERE m.auth_user_id = $1
ORDER BY h.created_at;
