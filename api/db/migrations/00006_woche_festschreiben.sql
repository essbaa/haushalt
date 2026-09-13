-- Die festgeschriebene Woche.
--
-- Bis hierher war ein Wochenplan eine Rechnung: Bei jedem Aufruf neu aus
-- Vorlagen und Verlauf ermittelt, nirgends gespeichert. Das war richtig,
-- solange man ihn nur ansehen konnte. Sobald jemand eine Aufgabe abhaken oder
-- abgeben soll, braucht sie eine Zeile, an der das hängen kann — und der Plan
-- darf sich beim Neuladen nicht mehr verschieben.
--
-- task_instance und assignment gibt es seit dem ersten Tag; gefüllt hat sie
-- bisher nur der Importer. week_plan ist das Fehlende: die Marke „diese Woche
-- ist geschrieben". Ohne sie wäre eine Woche ohne Aufgaben nicht von einer
-- ungeschriebenen zu unterscheiden, und jeder Aufruf würde sie erneut zu
-- schreiben versuchen.

-- +goose Up

CREATE TABLE week_plan (
    household_id uuid NOT NULL REFERENCES household (id) ON DELETE CASCADE,
    iso_week     text NOT NULL CHECK (iso_week ~ '^\d{4}-W\d{2}$'),

    -- Was nicht im Plan steht und warum. Als JSONB, weil es Begründungen für
    -- Menschen sind und keine Beziehungen — niemand fragt „alle Haushalte, in
    -- denen Rasenmähen wegen Startdichte fehlte".
    skipped      jsonb NOT NULL DEFAULT '[]',

    created_at   timestamptz NOT NULL DEFAULT now(),

    -- Der zusammengesetzte Schlüssel ist zugleich die Absicherung gegen zwei
    -- gleichzeitige erste Aufrufe: Der zweite läuft in den Konflikt, schreibt
    -- nichts und liest, was der erste geschrieben hat.
    PRIMARY KEY (household_id, iso_week)
);

-- +goose Down

DROP TABLE IF EXISTS week_plan;
