-- Eine Anmeldung darf in mehreren Haushalten Mitglied sein.
--
-- In 00001 stand `auth_user_id text UNIQUE` — global eindeutig. Die Annahme
-- dahinter („ein Mensch, ein Haushalt") hat nie jemand ausgesprochen, sie stand
-- einfach im Schema. Aufgefallen ist sie erst, als jemand mit eigenem Haushalt
-- einer Einladung folgte: Fremdschlüsselverletzung, 500, und die Frage, welche
-- der beiden Wahrheiten gilt.
--
-- Es gilt diese: Wer ein Konto hat, kann in mehreren Haushalten leben — die
-- Familie und die WG, die Patchwork-Konstellation, der Haushalt der Eltern, um
-- den man sich mitkümmert. Der Umschalter auf der Startseite setzte das ohnehin
-- voraus, und ListHouseholdsForAuthUser gibt seit dem ersten Tag eine Liste
-- zurück.
--
-- Eindeutig bleibt es trotzdem: je Haushalt genau einmal.

-- +goose Up

ALTER TABLE member DROP CONSTRAINT member_auth_user_id_key;
ALTER TABLE member ADD CONSTRAINT member_auth_user_je_haushalt
    UNIQUE (household_id, auth_user_id);

-- +goose Down

ALTER TABLE member DROP CONSTRAINT member_auth_user_je_haushalt;
ALTER TABLE member ADD CONSTRAINT member_auth_user_id_key UNIQUE (auth_user_id);
