-- Was ein Haushalt hat und was nicht.
--
-- Bis hierher kannte das Modell fünf Dinge: Wohnform, Auto, Garten und
-- Haustiere. Alles andere galt als vorhanden — und deshalb stand „Pflanzen
-- gießen" im Plan eines Haushalts ohne Pflanzen und „Geschenk für den
-- Kindergeburtstag" ohne Geburtstag.
--
-- facts hält die offene Liste: pflanzen, spuelmaschine, keller, fahrrad. Als
-- JSONB, weil die Liste wächst, ohne dass das Schema mitwachsen soll.
--
-- Entscheidend ist, was NICHT drinsteht: Ein fehlender Schlüssel heißt
-- „unbekannt", nicht „nein". Unbekannt wird nicht geplant, sondern gefragt.
-- Ohne diesen dritten Zustand wäre die Tabelle nur eine bequemere Art, wieder
-- alles anzunehmen.

-- +goose Up

ALTER TABLE household
    ADD COLUMN facts jsonb NOT NULL DEFAULT '{}';

-- +goose Down

ALTER TABLE household DROP COLUMN IF EXISTS facts;
