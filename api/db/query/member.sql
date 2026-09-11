-- name: CreateMember :one
INSERT INTO member (household_id, name, role, birth_year, care, capacity_minutes, auth_user_id)
VALUES ($1, $2, $3, $4, $5, $6, $7)
RETURNING *;

-- name: ListMembers :many
SELECT * FROM member WHERE household_id = $1 ORDER BY created_at;

-- name: GetMemberByAuthUserID :one
-- Der Übergang von der Anmeldung in die Fachwelt: Better Auth kennt nur eine
-- Nutzerkennung, alles Weitere hängt an dieser einen Zeile.
SELECT * FROM member WHERE auth_user_id = $1;

-- name: SetMemberAuthUser :one
-- Verbindet eine bestehende Person mit einem Login — der Fall, dass jemand
-- in einen Haushalt eingeladen wird, der ihn schon als Mitglied führt.
UPDATE member SET auth_user_id = $2 WHERE id = $1 RETURNING *;
