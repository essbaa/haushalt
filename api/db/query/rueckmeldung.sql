-- name: InsertFeedback :exec
-- Eine Rückmeldung aus der App. Jedes Mitglied darf melden, nicht nur die
-- planenden — die ausführenden sehen die Stellen, an denen es klemmt, zuerst.
INSERT INTO feedback (household_id, member_id, art, text, kontext)
VALUES ($1, $2, $3, $4, $5);

-- name: ListFeedback :many
-- Was dieser Haushalt gemeldet hat, das Neueste zuerst.
--
-- Mit Namen daneben: Eine Liste von Sätzen ohne Absender ist bei der
-- Auswertung nur halb so viel wert — „das hat sie gesagt" und „das hat er
-- gesagt" sind zwei verschiedene Befunde.
SELECT f.id, f.art, f.text, f.kontext, f.created_at, m.name AS wer
FROM feedback f
LEFT JOIN member m ON m.id = f.member_id
WHERE f.household_id = $1
ORDER BY f.created_at DESC
LIMIT 50;
