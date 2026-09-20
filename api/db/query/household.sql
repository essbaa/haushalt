-- name: CreateHousehold :one
INSERT INTO household (name, home, has_car, has_yard, pets, timezone, facts, rooms, baths)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
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
INSERT INTO household (slug, name, home, has_car, has_yard, pets, timezone, facts, rooms, baths)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
ON CONFLICT (slug) DO UPDATE
    SET name     = EXCLUDED.name,
        home     = EXCLUDED.home,
        has_car  = EXCLUDED.has_car,
        has_yard = EXCLUDED.has_yard,
        pets     = EXCLUDED.pets,
        timezone = EXCLUDED.timezone,
        facts    = EXCLUDED.facts,
        rooms    = EXCLUDED.rooms,
        baths    = EXCLUDED.baths
RETURNING *;

-- name: GetHouseholdBySlug :one
SELECT * FROM household WHERE slug = $1;

-- name: ListHouseholds :many
-- Alle Haushalte. Ab der Anmeldung tritt ListHouseholdsForAuthUser an diese
-- Stelle — bis dahin sind es die Beispielhaushalte aus dem Repo.
SELECT * FROM household ORDER BY created_at;

-- name: ListDemoHouseholds :many
-- Haushalte mit Slug sind die aus dem Repo: öffentlich sichtbar, ohne
-- Anmeldung. Alles ohne Slug gehört jemandem.
SELECT * FROM household WHERE slug IS NOT NULL ORDER BY created_at;

-- name: UpdateHousehold :one
-- Einstellungen ändern. Jedes Feld darf fehlen; was fehlt, bleibt stehen.
--
-- COALESCE statt einer gebauten Anweisung: Der Unterschied zwischen „nicht
-- mitgeschickt" und „auf leer gesetzt" wird hier entschieden und nicht in Go
-- zusammengestückelt.
UPDATE household SET
    name     = COALESCE(sqlc.narg('name'), name),
    home     = COALESCE(sqlc.narg('home'), home),
    has_car  = COALESCE(sqlc.narg('has_car'), has_car),
    has_yard = COALESCE(sqlc.narg('has_yard'), has_yard),
    pets     = COALESCE(sqlc.narg('pets'), pets),
    timezone = COALESCE(sqlc.narg('timezone'), timezone),
    rooms    = COALESCE(sqlc.narg('rooms'), rooms),
    baths    = COALESCE(sqlc.narg('baths'), baths)
WHERE id = sqlc.arg('id')
RETURNING *;

-- name: SetFacts :one
-- Fakten zusammenführen, nicht ersetzen: || vereinigt zwei JSONB-Objekte,
-- rechts gewinnt. Ein Formular, das nur eine Antwort schickt, soll die
-- übrigen nicht löschen.
UPDATE household
SET facts = facts || $2::jsonb
WHERE id = $1
RETURNING *;

-- name: ListAgreements :many
-- Das Wochenraster des Haushalts: je Vorlage und Wochentag eine Person.
SELECT template_id, weekday, member_id
FROM agreement
WHERE household_id = $1
ORDER BY template_id, weekday;

-- name: SetAgreement :exec
-- Einen Platz im Raster besetzen. Zweimal dasselbe zu setzen ist kein Fehler.
INSERT INTO agreement (household_id, template_id, weekday, member_id)
VALUES ($1, $2, $3, $4)
ON CONFLICT (household_id, template_id, weekday)
DO UPDATE SET member_id = EXCLUDED.member_id;

-- name: ClearAgreementRow :exec
-- Das ganze Raster einer Vorlage leeren — der erste Schritt beim Setzen einer
-- kompletten Zeile. Sieben Plätze einzeln abzugleichen wäre sieben Abfragen
-- und dieselbe Wirkung.
DELETE FROM agreement WHERE household_id = $1 AND template_id = $2;

-- name: ListTemplateWeekdays :many
-- Die Wochentage, die dieser Haushalt für Vorlagen festgelegt hat.
SELECT template_id, weekday
FROM template_weekday
WHERE household_id = $1
ORDER BY template_id, weekday;

-- name: SetTemplateWeekday :exec
-- Einen Wochentag festlegen. Zweimal denselben zu setzen ist kein Fehler.
INSERT INTO template_weekday (household_id, template_id, weekday)
VALUES ($1, $2, $3)
ON CONFLICT (household_id, template_id, weekday) DO NOTHING;

-- name: ClearTemplateWeekdays :exec
-- Alle Wochentage einer Vorlage löschen — der erste Schritt beim Setzen einer
-- neuen Auswahl und zugleich das Zurücksetzen auf den Rhythmus der Bibliothek.
DELETE FROM template_weekday WHERE household_id = $1 AND template_id = $2;
