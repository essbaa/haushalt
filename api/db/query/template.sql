-- name: UpsertCuratedTemplate :exec
-- Import der Bibliothek aus dem Repo. Beim Start oder in einem eigenen
-- Kommando aufgerufen; die Datei bleibt die Quelle, die Tabelle ist die Kopie.
INSERT INTO task_template (id, household_id, source, version, definition)
VALUES ($1, NULL, 'kuratiert', $2, $3)
ON CONFLICT (id) DO UPDATE
    SET version    = EXCLUDED.version,
        definition = EXCLUDED.definition;

-- name: ListTemplatesForHousehold :many
-- Kuratierte Vorlagen gelten für alle (household_id IS NULL), dazu die
-- eigenen. Die Reihenfolge ist stabil, damit der Planer deterministisch bleibt.
SELECT * FROM task_template
WHERE household_id IS NULL OR household_id = $1
ORDER BY id;
