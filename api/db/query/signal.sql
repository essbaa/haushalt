-- name: UpsertSignal :exec
-- Lernwerte sind ableitbar, nicht Quelle: Diese Tabelle darf jederzeit aus dem
-- Protokoll neu berechnet werden.
INSERT INTO signal (household_id, template_id, kind, value, updated_at)
VALUES ($1, $2, $3, $4, now())
ON CONFLICT (household_id, template_id, kind)
DO UPDATE SET value = EXCLUDED.value, updated_at = now();

-- name: ListSignals :many
SELECT * FROM signal WHERE household_id = $1 ORDER BY template_id, kind;
