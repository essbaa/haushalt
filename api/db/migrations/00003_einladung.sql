-- Einladungen: wie ein zweiter Mensch in einen bestehenden Haushalt kommt.

-- +goose Up

CREATE TABLE invitation (
    -- Der Code ist der Primärschlüssel, nicht eine Nummer daneben. Er wird
    -- getippt oder aus einem Link gelesen; eine zweite Kennung daneben wäre
    -- eine zweite Wahrheit.
    code         text PRIMARY KEY,
    household_id uuid NOT NULL REFERENCES household (id) ON DELETE CASCADE,

    -- Die Rolle, die der Eingeladene bekommt. 'betreut' fehlt hier
    -- absichtlich: Wer betreut wird, meldet sich nicht an — ein Zweijähriger
    -- erzeugt Arbeit, ohne ein Konto zu haben.
    role         text NOT NULL CHECK (role IN ('planend', 'ausfuehrend')),

    created_by   uuid REFERENCES member (id) ON DELETE SET NULL,
    created_at   timestamptz NOT NULL DEFAULT now(),

    -- Einmal gültig und begrenzt haltbar. Beides ist Absicht: Ein Code, der
    -- ewig gilt, liegt irgendwann in einem alten Chatverlauf und öffnet einen
    -- Haushalt mit Kinderdaten.
    expires_at   timestamptz NOT NULL,
    used_at      timestamptz,
    used_by      uuid REFERENCES member (id) ON DELETE SET NULL
);

CREATE INDEX invitation_household_idx ON invitation (household_id);

-- +goose Down

DROP TABLE IF EXISTS invitation;
