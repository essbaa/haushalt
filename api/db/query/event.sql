-- name: AppendEvent :one
-- Die einzige schreibende Operation auf dem Protokoll. Ändern und Löschen
-- verhindert der Trigger in der Migration, nicht die Disziplin des Aufrufers.
INSERT INTO event (household_id, member_id, task_instance_id, kind, payload)
VALUES ($1, $2, $3, $4, $5)
RETURNING *;

-- name: ListEvents :many
SELECT * FROM event
WHERE household_id = $1
ORDER BY occurred_at DESC, id DESC
LIMIT $2;

-- name: LastDonePerTemplate :many
-- Verdichtet das Protokoll zu dem, was der Planer als Historie braucht.
--
-- AT TIME ZONE h.timezone vor dem Schnitt auf das Datum: Der Zeitstempel liegt
-- in UTC, die Frage „an welchem Tag war das?" ist aber eine Frage nach der
-- Ortszeit des Haushalts (ADR-0002). Ohne die Umrechnung landet eine Aufgabe,
-- die sonntags um 23:30 abgehakt wurde, im Montag.
SELECT
    t.template_id,
    max((e.occurred_at AT TIME ZONE h.timezone))::date AS last_done
FROM event e
JOIN task_instance t ON t.id = e.task_instance_id
JOIN household h ON h.id = e.household_id
WHERE e.household_id = $1 AND e.kind = 'erledigt'
GROUP BY t.template_id;

-- name: LastAssigneePerTemplate :many
-- Wer eine Vorlage zuletzt hatte — die Grundlage der Rotation.
--
-- DISTINCT ON ist Postgres-eigen und hier genau richtig: erste Zeile je
-- Vorlage, sortiert nach Tag absteigend. In Standard-SQL bräuchte es dafür
-- eine Fensterfunktion und eine Unterabfrage.
SELECT DISTINCT ON (t.template_id)
    t.template_id,
    a.member_id
FROM task_instance t
JOIN assignment a ON a.task_instance_id = t.id
WHERE t.household_id = $1
ORDER BY t.template_id, t.day DESC;

-- name: CountEvents :one
-- Der Import legt Historie nur an, wenn noch keine da ist.
--
-- Das Protokoll ist reines Anhängen — es lässt sich nicht bereinigen, auch
-- nicht vom Import. Also darf er nur einmal schreiben, sonst wächst bei jedem
-- Lauf ein weiterer Satz „erledigt" hinzu.
SELECT count(*) FROM event WHERE household_id = $1;

-- name: AppendHistoricEvent :one
-- Nur für den Import: ein Ereignis mit einem Zeitpunkt aus der Vergangenheit.
--
-- Im Betrieb setzt die Datenbank now(), und das ist keine Bequemlichkeit: Ein
-- Aufrufer, der den Zeitpunkt bestimmen darf, kann die Vergangenheit
-- umschreiben — und dagegen steht das ganze Protokoll. Der Import ist die eine
-- Ausnahme, und sie hat deshalb ihre eigene Abfrage statt eines Parameters an
-- der normalen.
INSERT INTO event (household_id, member_id, task_instance_id, kind, payload, occurred_at)
VALUES ($1, $2, $3, $4, $5, $6)
RETURNING *;
