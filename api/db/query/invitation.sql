-- name: CreateInvitation :one
-- member_id darf leer sein: Dann kommt jemand dazu, den es im Haushalt noch
-- nicht gibt. Ist er gesetzt, übernimmt der Beitritt genau diese Person.
INSERT INTO invitation (code, household_id, role, created_by, expires_at, member_id)
VALUES ($1, $2, $3, $4, $5, $6)
RETURNING *;

-- name: GetInvitation :one
-- Der Code allein genügt; ob er noch gilt, entscheidet der Aufrufer anhand
-- von expires_at und used_at. Absichtlich nicht in der Abfrage gefiltert:
-- „abgelaufen" und „schon benutzt" sind verschiedene Auskünfte, und wer sie
-- hier wegfiltert, kann sie später nicht mehr geben.
SELECT * FROM invitation WHERE code = $1;

-- name: ConsumeInvitation :one
-- Verbraucht die Einladung — und zwar nur, wenn sie noch zu haben ist.
--
-- Die Bedingungen stehen in der UPDATE-Anweisung und nicht davor in Go. Der
-- Unterschied ist nicht Stil, sondern Korrektheit: „erst prüfen, dann
-- schreiben" lässt zwei gleichzeitige Beitritte beide durch, weil beide
-- prüfen, bevor einer schreibt. So gewinnt genau einer, und der andere
-- bekommt keine Zeile zurück.
UPDATE invitation
SET used_at = now(), used_by = $2
WHERE code = $1 AND used_at IS NULL AND expires_at > now()
RETURNING *;

-- name: ListOpenInvitations :many
SELECT * FROM invitation
WHERE household_id = $1 AND used_at IS NULL AND expires_at > now()
ORDER BY created_at DESC;

-- name: RoleInHousehold :one
-- Die Rolle des Aufrufers in diesem Haushalt. Keine Zeile heißt: kein
-- Mitglied.
SELECT role FROM member WHERE household_id = $1 AND auth_user_id = $2;

-- name: CreateMemberWithRole :one
-- Wie CreateMember, aber für den Beitritt über eine Einladung: Der Haushalt
-- steht fest, die Rolle kommt aus der Einladung.
INSERT INTO member (household_id, name, role, capacity_minutes, auth_user_id)
VALUES ($1, $2, $3, $4, $5)
RETURNING *;

-- name: GetInvitableMember :one
-- Die Person, auf die eine Einladung zeigen soll — nur wenn sie zu diesem
-- Haushalt gehört und noch kein Konto hat. Die Prüfung steht in der Abfrage
-- und nicht in Go: Ein Haushalt darf niemanden aus einem fremden Haushalt
-- einladen, und das soll nicht an einem vergessenen if hängen.
SELECT * FROM member
WHERE id = $1 AND household_id = $2 AND auth_user_id IS NULL AND role <> 'betreut';

-- name: ClaimMember :one
-- Verbindet eine vorhandene Person mit einer Anmeldung.
--
-- Wieder mit den Bedingungen in der Anweisung statt davor: „auth_user_id IS
-- NULL" sorgt dafür, dass zwei gleichzeitige Beitritte nicht beide dieselbe
-- Person übernehmen. Wer keine Zeile zurückbekommt, war der Zweite.
UPDATE member
SET auth_user_id = $1
WHERE id = $2 AND household_id = $3 AND auth_user_id IS NULL
RETURNING *;
