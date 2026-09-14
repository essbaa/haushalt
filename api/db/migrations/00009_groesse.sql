-- Wie groß der Haushalt ist.
--
-- Anders als die Wohnform, die abgefragt wurde und nichts bewirkte: Diese
-- beiden Zahlen filtern nichts, sie skalieren. Staubsaugen dauert in fünf
-- Zimmern länger als in zwei, und wer zwei Bäder hat, putzt zwei — und die
-- Dauer trägt die Fairnessrechnung.
--
-- Die Vorgaben sind der Haushalt, für den die Bibliothek kuratiert ist: drei
-- Zimmer, ein Bad. Bestehende Zeilen behalten damit genau die Dauern, die sie
-- heute haben.

-- +goose Up

ALTER TABLE household
    ADD COLUMN rooms int NOT NULL DEFAULT 3 CHECK (rooms BETWEEN 1 AND 15),
    ADD COLUMN baths int NOT NULL DEFAULT 1 CHECK (baths BETWEEN 0 AND 5);

-- +goose Down

ALTER TABLE household DROP COLUMN IF EXISTS rooms;
ALTER TABLE household DROP COLUMN IF EXISTS baths;
