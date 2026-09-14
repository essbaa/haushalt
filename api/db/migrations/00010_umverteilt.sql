-- Umverteilen von Hand ist ein Ereignis, das es noch nicht gab.
--
-- Das Protokoll hat ein geschlossenes Vokabular: `kind` ist durch einen CHECK
-- auf sechs Wörter begrenzt. Das ist Absicht — ein Tippfehler soll keine
-- siebte Ereignisart erfinden, und jede neue Art ist eine Produktentscheidung,
-- die eine Zeile in einer Migration wert ist.
--
-- `verschoben` wäre naheliegend gewesen und ist es nicht: Das Wort gehört der
-- Verschiebung auf einen anderen *Tag*, die noch kommt. Eine Aufgabe, die die
-- Person wechselt, und eine, die den Tag wechselt, sind zwei verschiedene
-- Rückmeldungen — und das Protokoll wird gelesen, um daraus zu lernen. Zwei
-- Dinge unter einem Namen sind dort nicht zu trennen.

-- +goose Up

ALTER TABLE event DROP CONSTRAINT event_kind_check;
ALTER TABLE event ADD CONSTRAINT event_kind_check CHECK (kind IN
    ('erstellt', 'erledigt', 'verschoben', 'abgegeben', 'geloescht',
     'wieder_geoeffnet', 'umverteilt'));

-- +goose Down
--
-- Schlägt fehl, sobald es Umverteilungen gibt, und das ist richtig so:
-- Ereignisse sind unveränderlich (siehe den Trigger in 00001). Eine
-- Rückmigration, die sie stillschweigend wegräumt, wäre schlimmer als eine,
-- die scheitert und den Menschen entscheiden lässt.

ALTER TABLE event DROP CONSTRAINT event_kind_check;
ALTER TABLE event ADD CONSTRAINT event_kind_check CHECK (kind IN
    ('erstellt', 'erledigt', 'verschoben', 'abgegeben', 'geloescht',
     'wieder_geoeffnet'));
