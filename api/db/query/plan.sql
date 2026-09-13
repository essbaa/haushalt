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

-- name: MarkWeekWritten :one
-- Schreibt die Marke „diese Woche steht" — und nur der erste Aufrufer
-- gewinnt. ON CONFLICT DO NOTHING liefert dann keine Zeile zurück, und genau
-- daran erkennt der zweite, dass er lesen statt schreiben soll.
INSERT INTO week_plan (household_id, iso_week, skipped)
VALUES ($1, $2, $3)
ON CONFLICT (household_id, iso_week) DO NOTHING
RETURNING *;

-- name: GetWrittenWeek :one
SELECT * FROM week_plan WHERE household_id = $1 AND iso_week = $2;

-- name: ListDoneTasks :many
-- Welche Aufgaben dieser Woche erledigt sind.
--
-- Die Wahrheit steht im Ereignisprotokoll, nicht in einer Spalte auf der
-- Aufgabe. Ein Häkchen wäre eine zweite Wahrheit neben dem Protokoll — und
-- das Protokoll ist die, die auch „wieder geöffnet" erzählen kann.
SELECT DISTINCT ON (e.task_instance_id)
    e.task_instance_id, e.member_id, e.occurred_at, e.kind
FROM event e
JOIN task_instance t ON t.id = e.task_instance_id
WHERE t.household_id = $1
  AND t.iso_week = $2
  AND e.kind IN ('erledigt', 'wieder_geoeffnet')
ORDER BY e.task_instance_id, e.occurred_at DESC, e.id DESC;

-- name: GetTaskForMember :one
-- Eine Aufgabe samt ihrem Haushalt — aber nur, wenn der Aufrufer in diesem
-- Haushalt Mitglied ist.
--
-- Die Berechtigungsprüfung steht im JOIN und nicht davor in Go: Eine Aufgabe
-- aus einem fremden Haushalt gibt es für diesen Aufrufer nicht, und das soll
-- nicht an einem vergessenen if hängen.
SELECT
    t.id, t.household_id, t.template_id, t.iso_week, t.day,
    m.id AS member_id, m.role AS member_role,
    a.member_id AS assignee_id
FROM task_instance t
JOIN member m ON m.household_id = t.household_id AND m.auth_user_id = $2
LEFT JOIN assignment a ON a.task_instance_id = t.id
WHERE t.id = $1;

-- name: ClearAssignment :exec
-- Abgegeben und noch niemand übernommen. Ein gültiger Zwischenzustand — siehe
-- den LEFT JOIN in GetWeekPlan.
DELETE FROM assignment WHERE task_instance_id = $1;

-- name: SetAssignment :exec
-- Nach einer Abgabe: neue Zuständige, neue Begründung, von Hand markiert.
INSERT INTO assignment (task_instance_id, member_id, reason_code, reason_previous, manual)
VALUES ($1, $2, $3, $4, true);

-- name: DeleteUntouchedWeekTasks :exec
-- Verwirft die Aufgaben einer Woche, an denen nichts hängt.
--
-- Der Filter ist keine Vorsicht, sondern eine Konsequenz: event.task_instance_id
-- ist ON DELETE SET NULL, und der Trigger event_kein_update verbietet jedes
-- UPDATE auf event. Eine Aufgabe mit Ereignis zu löschen wirft also eine
-- Ausnahme — die Datenbank lässt gar nicht zu, dass Neurechnen Geschehenes
-- wegräumt.
--
-- Daraus wird eine Produktregel: Was Spuren hinterlassen hat, bleibt. Was nur
-- ein Vorschlag war, wird neu gerechnet.
DELETE FROM task_instance t
WHERE t.household_id = $1
  AND t.iso_week = $2
  AND NOT EXISTS (SELECT 1 FROM event e WHERE e.task_instance_id = t.id);

-- name: ListWeekTaskKeys :many
-- Was von einer Woche übrig ist, nach dem Verwerfen: Vorlage und Tag. Damit
-- rechnet das Neuschreiben nichts doppelt hin.
SELECT template_id, day FROM task_instance
WHERE household_id = $1 AND iso_week = $2;

-- name: UpdateWeekSkipped :exec
UPDATE week_plan SET skipped = $3
WHERE household_id = $1 AND iso_week = $2;
