-- Was der Import braucht: stabile Kennungen und eine ehrliche Begründung.

-- +goose Up

-- Eine sprechende Kennung für Haushalte, die aus dem Repo kommen.
--
-- Ohne sie hieße die Demo-Adresse /api/haushalte/8f3c…-…/plan/2026-W38 statt
-- /api/haushalte/familie-a/…, und der Import hätte keinen Schlüssel, an dem er
-- erkennt, ob er diesen Haushalt schon angelegt hat. NULL ist der Normalfall:
-- Haushalte, die sich jemand selbst anlegt, brauchen keinen Namen in der URL.
ALTER TABLE household ADD COLUMN slug text UNIQUE;

-- Zwei Personen mit demselben Namen in einem Haushalt gibt es nicht — und der
-- Import braucht einen Schlüssel, um beim zweiten Lauf dieselbe Person
-- wiederzufinden statt eine zweite anzulegen.
ALTER TABLE member ADD CONSTRAINT member_name_je_haushalt UNIQUE (household_id, name);

-- 'importiert' als Begründung.
--
-- Für eine eingelesene Historie wissen wir, WER eine Aufgabe zuletzt hatte,
-- aber nicht WARUM. Eine der echten Begründungen hinzuschreiben wäre bequem
-- und falsch; die Oberfläche würde später „zuletzt bei Anna, zum Ausgleich"
-- behaupten, wo niemand etwas ausgeglichen hat.
ALTER TABLE assignment DROP CONSTRAINT assignment_reason_code_check;
ALTER TABLE assignment ADD CONSTRAINT assignment_reason_code_check
    CHECK (reason_code IN ('rotation', 'ausgleich', 'feste_person',
                           'einzige_moeglichkeit', 'frist', 'eigene_aufgabe', 'importiert'));

-- +goose Down

ALTER TABLE assignment DROP CONSTRAINT assignment_reason_code_check;
ALTER TABLE assignment ADD CONSTRAINT assignment_reason_code_check
    CHECK (reason_code IN ('rotation', 'ausgleich', 'feste_person',
                           'einzige_moeglichkeit', 'frist', 'eigene_aufgabe'));
ALTER TABLE member DROP CONSTRAINT member_name_je_haushalt;
ALTER TABLE household DROP COLUMN slug;
