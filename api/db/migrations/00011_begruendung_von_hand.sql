-- Die Begründung „von Hand" fehlte im zweiten Vokabular.
--
-- Dieselbe Sperre wie in 00010, eine Tabelle weiter: `assignment.reason_code`
-- ist ebenfalls durch einen CHECK begrenzt. Zwei Migrationen für eine Funktion
-- sind unschön und trotzdem richtig — 00010 war beim Fund schon gelaufen, und
-- eine gelaufene Migration wird nicht umgeschrieben. Goose merkt sich die
-- Nummer, nicht den Inhalt.
--
-- Die Lehre steht in docs/stolperstellen.md: Wer ein neues Wort in die
-- Datenbank schreibt, sucht vorher ALLE Stellen, an denen Wörter begrenzt
-- sind, und nicht die erste.

-- +goose Up

ALTER TABLE assignment DROP CONSTRAINT assignment_reason_code_check;
ALTER TABLE assignment ADD CONSTRAINT assignment_reason_code_check
    CHECK (reason_code IN ('rotation', 'ausgleich', 'feste_person',
                           'einzige_moeglichkeit', 'frist', 'eigene_aufgabe',
                           'importiert', 'von_hand'));

-- +goose Down
--
-- Schlägt fehl, solange Zuteilungen mit „von_hand" stehen. Anders als beim
-- Protokoll wäre ein Aufräumen hier möglich — und es bliebe falsch: Die
-- Begründung ist die einzige Spur, dass ein Mensch widersprochen hat.

ALTER TABLE assignment DROP CONSTRAINT assignment_reason_code_check;
ALTER TABLE assignment ADD CONSTRAINT assignment_reason_code_check
    CHECK (reason_code IN ('rotation', 'ausgleich', 'feste_person',
                           'einzige_moeglichkeit', 'frist', 'eigene_aufgabe',
                           'importiert'));
