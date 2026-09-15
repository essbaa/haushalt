-- Wer wann — das Wochenraster für Aufgaben, die abgesprochen werden.
--
-- Der Fall: ein Kind in die Kita bringen und wieder abholen, bei zwei
-- berufstätigen Eltern. Der Planer verteilt nach Kapazität — wie viel Zeit
-- jemand hat. Gebraucht wird Verfügbarkeit — ob jemand um 7:45 an einem
-- bestimmten Ort sein kann. Das sind verschiedene Größen, und die App kennt
-- nur die erste; ohne Absprache teilt sie mit voller Überzeugung den
-- Elternteil ein, der um acht eine Besprechung hat.
--
-- Eine Zeile je Vorlage und Wochentag. Kein Wochenbezug: Die Absprache gilt,
-- bis jemand sie ändert — damit ist „diese Woche wie letzte" der Normalfall
-- und braucht keinen Knopf. Eine einzelne Ausnahme ist Umverteilen und
-- betrifft nur den einen Termin.
--
-- ON DELETE CASCADE auf member: Zieht jemand aus, verschwindet seine Zeile,
-- und die Aufgabe steht wieder als „braucht Absprache" da. Still auf jemand
-- anderen umzubuchen wäre eine Entscheidung, die dem Haushalt gehört.

-- +goose Up

CREATE TABLE agreement (
    household_id uuid NOT NULL REFERENCES household (id) ON DELETE CASCADE,
    template_id  text NOT NULL,
    weekday      int  NOT NULL CHECK (weekday BETWEEN 0 AND 6),
    member_id    uuid NOT NULL REFERENCES member (id) ON DELETE CASCADE,
    PRIMARY KEY (household_id, template_id, weekday)
);

ALTER TABLE assignment DROP CONSTRAINT assignment_reason_code_check;
ALTER TABLE assignment ADD CONSTRAINT assignment_reason_code_check
    CHECK (reason_code IN ('rotation', 'ausgleich', 'feste_person',
                           'einzige_moeglichkeit', 'frist', 'eigene_aufgabe',
                           'importiert', 'von_hand', 'absprache'));

-- +goose Down

ALTER TABLE assignment DROP CONSTRAINT assignment_reason_code_check;
ALTER TABLE assignment ADD CONSTRAINT assignment_reason_code_check
    CHECK (reason_code IN ('rotation', 'ausgleich', 'feste_person',
                           'einzige_moeglichkeit', 'frist', 'eigene_aufgabe',
                           'importiert', 'von_hand'));

DROP TABLE agreement;
