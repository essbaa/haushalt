-- name: UpsertSignal :exec
-- Lernwerte sind ableitbar, nicht Quelle: Diese Tabelle darf jederzeit aus dem
-- Protokoll neu berechnet werden.
INSERT INTO signal (household_id, template_id, kind, value, updated_at)
VALUES ($1, $2, $3, $4, now())
ON CONFLICT (household_id, template_id, kind)
DO UPDATE SET value = EXCLUDED.value, updated_at = now();

-- name: ListSignals :many
SELECT * FROM signal WHERE household_id = $1 ORDER BY template_id, kind;

-- name: DeleteSignal :exec
-- Ein zurückgenommenes Abwählen löscht die Zeile, statt sie auf 0 zu setzen.
-- „Nie etwas gesagt" und „ausdrücklich wieder erlaubt" sollen im Protokoll
-- gleich aussehen — der Unterschied steht in den Ereignissen, nicht hier.
DELETE FROM signal
WHERE household_id = $1 AND template_id = $2 AND kind = $3;
