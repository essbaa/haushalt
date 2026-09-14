-- Anlässe: Geburtstage, Elternabend, Arzttermine.
--
-- Der Grund steht in ADR-0010. „Geschenk für Kindergeburtstag besorgen" trug
-- den Rhythmus „Auslöser, alle 45 Tage" und erfand damit alle anderthalb
-- Monate einen Geburtstag. Eine Aufgabe ohne Anlass ist keine Erinnerung,
-- sondern eine Behauptung. Hier ist der Anlass.
--
-- Ein Kalendertag, kein Zeitpunkt (ADR-0002): Ein Geburtstag hat keine
-- Uhrzeit, und eine erfundene würde über Zeitzonen hinweg den Tag verschieben.

-- +goose Up

CREATE TABLE occasion (
    id           uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    household_id uuid NOT NULL REFERENCES household (id) ON DELETE CASCADE,

    title        text NOT NULL CHECK (length(btrim(title)) > 0),
    day          date NOT NULL,

    -- Verbindet den Anlass mit den Vorlagen, die daran hängen. Kein
    -- Fremdschlüssel: Die Arten stehen in der Bibliothek im Repo, nicht in
    -- der Datenbank — sie sind Inhalt, kein Zustand.
    kind         text NOT NULL CHECK (length(btrim(kind)) > 0),

    -- Geburtstage wiederholen sich. Ohne diese Spalte müsste man jeden
    -- einzeln jährlich neu eintragen, und genau das täte niemand.
    yearly       boolean NOT NULL DEFAULT false,

    created_at   timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX occasion_household_idx ON occasion (household_id, day);

-- +goose Down

DROP TABLE IF EXISTS occasion;
