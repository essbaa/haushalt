-- +goose Up

-- Rückmeldungen aus der App heraus.
--
-- Der Zettel am Kühlschrank ist eine gute Idee mit einer bekannten
-- Schwäche: Er wird abends ausgefüllt, von der Person, die ihn aufgehängt
-- hat, und aus dem Gedächtnis. Was jemand am Mittwochmorgen dachte, als der
-- Plan etwas Falsches behauptete, steht am Abend nicht mehr in demselben
-- Wortlaut da — und der Wortlaut ist genau das, was hier zählt.
--
-- Deshalb der Weg aus der App selbst: an der Stelle, an der es auffällt, und
-- von der Person, der es auffällt. Auch von denen, die nicht planen — gerade
-- von denen.
--
-- Eigene Tabelle und kein Ereignis: `event` hängt an einer Aufgabe und an
-- einem geschlossenen Vokabular. Eine Rückmeldung hängt an nichts davon; sie
-- in das Protokoll zu drücken hieße, zwei verschiedene Dinge unter einen
-- Begriff zu zwingen, nur weil beide zeitlich geordnet sind.
CREATE TABLE feedback (
    id           uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    household_id uuid        NOT NULL REFERENCES household (id) ON DELETE CASCADE,
    -- Wer gemeldet hat. NULL ist erlaubt: Eine Person kann den Haushalt
    -- verlassen, und ihre Rückmeldung bleibt trotzdem wahr.
    member_id    uuid        REFERENCES member (id) ON DELETE SET NULL,
    art          text        NOT NULL CHECK (art IN ('fehler', 'idee', 'stoert')),
    text         text        NOT NULL CHECK (length(text) BETWEEN 1 AND 2000),
    -- Woher die Meldung kam: Seite, Woche, App-Fassung. Ohne das beginnt jede
    -- Auswertung mit der Rückfrage „wo warst du gerade?", und die Antwort
    -- darauf ist eine Woche später niemandem mehr geläufig.
    kontext      text        NOT NULL DEFAULT '',
    created_at   timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX feedback_haushalt_zeit ON feedback (household_id, created_at DESC);

-- +goose Down

DROP TABLE feedback;
