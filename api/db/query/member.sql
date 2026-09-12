-- name: CreateMember :one
INSERT INTO member (household_id, name, role, birth_year, care, capacity_minutes, auth_user_id)
VALUES ($1, $2, $3, $4, $5, $6, $7)
RETURNING *;

-- name: ListMembers :many
SELECT * FROM member WHERE household_id = $1 ORDER BY created_at, name;

-- name: GetMemberInHousehold :one
-- Der Übergang von der Anmeldung in die Fachwelt: Better Auth kennt nur eine
-- Nutzerkennung, alles Weitere hängt an dieser einen Zeile.
--
-- Immer MIT Haushalt gefragt. Dieselbe Anmeldung kann in mehreren Haushalten
-- eine Person sein, mit verschiedenen Rollen und verschiedenen Namen — eine
-- Abfrage ohne household_id würde irgendeine davon zurückgeben.
SELECT * FROM member WHERE household_id = $1 AND auth_user_id = $2;

-- name: SetMemberAuthUser :one
-- Verbindet eine bestehende Person mit einem Login — der Fall, dass jemand
-- in einen Haushalt eingeladen wird, der ihn schon als Mitglied führt.
UPDATE member SET auth_user_id = $2 WHERE id = $1 RETURNING *;

-- name: UpsertMember :one
-- Schlüssel ist der Name im Haushalt. auth_user_id bleibt unangetastet: Wer
-- sich einmal angemeldet hat, verliert die Verbindung nicht, weil jemand die
-- Beispieldaten neu einliest.
INSERT INTO member (household_id, name, role, birth_year, care, capacity_minutes)
VALUES ($1, $2, $3, $4, $5, $6)
ON CONFLICT (household_id, name) DO UPDATE
    SET role             = EXCLUDED.role,
        birth_year       = EXCLUDED.birth_year,
        care             = EXCLUDED.care,
        capacity_minutes = EXCLUDED.capacity_minutes
RETURNING *;

-- name: HasMembership :one
-- Ist diese angemeldete Person überhaupt irgendwo Mitglied? Die Frage
-- entscheidet, ob beim ersten Zugriff ein Haushalt entsteht.
SELECT EXISTS(SELECT 1 FROM member WHERE auth_user_id = $1);

-- name: IsMemberOf :one
-- Darf diese Person diesen Haushalt sehen?
--
-- Eine Abfrage, kein Abgleich in Go: Die Antwort soll aus derselben Quelle
-- kommen wie die Daten, sonst driften Berechtigung und Inhalt auseinander.
SELECT EXISTS(
    SELECT 1 FROM member WHERE household_id = $1 AND auth_user_id = $2
);
