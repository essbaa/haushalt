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

-- name: UpsertDemoHousehold :one
-- Haushalte aus dem Repo. Der Slug ist der Schlüssel, damit ein zweiter
-- Import dieselbe Zeile trifft statt eine neue anzulegen.
INSERT INTO household (slug, name, home, has_car, has_yard, pets, timezone)
VALUES ($1, $2, $3, $4, $5, $6, $7)
ON CONFLICT (slug) DO UPDATE
    SET name     = EXCLUDED.name,
        home     = EXCLUDED.home,
        has_car  = EXCLUDED.has_car,
        has_yard = EXCLUDED.has_yard,
        pets     = EXCLUDED.pets,
        timezone = EXCLUDED.timezone
RETURNING *;

-- name: GetHouseholdBySlug :one
SELECT * FROM household WHERE slug = $1;

-- name: ListHouseholds :many
-- Alle Haushalte. Ab der Anmeldung tritt ListHouseholdsForAuthUser an diese
-- Stelle — bis dahin sind es die Beispielhaushalte aus dem Repo.
SELECT * FROM household ORDER BY created_at;
