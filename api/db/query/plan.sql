-- name: DeleteWeek :exec
-- Ein Wochenplan wird ersetzt, nicht fortgeschrieben: Der Planer ist
-- deterministisch, ein zweiter Lauf mit derselben Eingabe ergibt dasselbe.
-- Die Zuteilungen hängen per ON DELETE CASCADE mit dran.
DELETE FROM task_instance WHERE household_id = $1 AND iso_week = $2;

-- name: InsertTaskInstance :one
INSERT INTO task_instance
    (household_id, template_id, iso_week, day, slot, duration_min, head_load, deadline)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
RETURNING id;

-- name: InsertAssignment :exec
INSERT INTO assignment (task_instance_id, member_id, reason_code, reason_previous, manual)
VALUES ($1, $2, $3, $4, $5);

-- name: GetWeekPlan :many
-- Der gespeicherte Plan einer Woche, mit der Vorlage daneben.
--
-- LEFT JOIN auf assignment mit Absicht: Eine Aufgabe ohne Zuständige ist ein
-- gültiger Zustand — jemand hat sie abgegeben und noch niemand übernommen.
SELECT
    t.id,
    t.template_id,
    t.day,
    t.slot,
    t.duration_min,
    t.head_load,
    t.deadline,
    tt.definition,
    a.member_id,
    a.reason_code,
    a.reason_previous,
    a.manual
FROM task_instance t
JOIN task_template tt ON tt.id = t.template_id
LEFT JOIN assignment a ON a.task_instance_id = t.id
WHERE t.household_id = $1 AND t.iso_week = $2
ORDER BY t.day, t.id;

-- name: ReassignTask :exec
-- Von Hand umverteilt. manual = true ist die eigentliche Information: Jede
-- Korrektur sagt, wo der Planer danebenlag.
UPDATE assignment
SET member_id = $2, manual = true
WHERE task_instance_id = $1;
