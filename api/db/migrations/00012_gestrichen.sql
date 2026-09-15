-- Eine Aufgabe für diese eine Woche streichen.
--
-- Bis hierher gab es nur „Brauchen wir nicht", und das schaltet die Vorlage
-- für immer ab. Damit war der einzige Ausweg aus einer Aufgabe eine
-- Entscheidung über alle künftigen Wochen — und ein Fehlgriff nahm dem
-- Haushalt eine Vorlage weg, die er eigentlich wollte.
--
-- Zwei neue Ereignisarten statt einer Spalte: Der Zustand steht im Protokoll,
-- wie bei „erledigt" und „wieder geöffnet". Eine Spalte wäre eine zweite
-- Wahrheit daneben, und nur das Protokoll kann erzählen, dass jemand es sich
-- anders überlegt hat.
--
-- Warum nicht `geloescht`, das schon in der Liste steht: Das Wort gehört dem
-- Löschen, nicht dem Aussetzen. Zwei verschiedene Absichten unter einem Namen
-- sind später nicht mehr zu trennen — dieselbe Überlegung wie bei
-- `verschoben` in 00010.

-- +goose Up

ALTER TABLE event DROP CONSTRAINT event_kind_check;
ALTER TABLE event ADD CONSTRAINT event_kind_check CHECK (kind IN
    ('erstellt', 'erledigt', 'verschoben', 'abgegeben', 'geloescht',
     'wieder_geoeffnet', 'umverteilt', 'gestrichen', 'wieder_eingeplant'));

-- +goose Down

ALTER TABLE event DROP CONSTRAINT event_kind_check;
ALTER TABLE event ADD CONSTRAINT event_kind_check CHECK (kind IN
    ('erstellt', 'erledigt', 'verschoben', 'abgegeben', 'geloescht',
     'wieder_geoeffnet', 'umverteilt'));
