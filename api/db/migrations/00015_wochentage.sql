-- An welchen Wochentagen eine Vorlage in DIESEM Haushalt liegt.
--
-- Der Fall: Müll rausbringen steht in der Bibliothek auf Dienstag, weil
-- irgendein Tag dastehen musste. Bei euch kommt die Tonne donnerstags. Bisher
-- gab es dagegen genau ein Mittel — die ganze Vorlage abbestellen und von Hand
-- neu anlegen, womit ihr die Bibliothek verliert und die Aufgabe als „eigene"
-- mit geschätzten Zahlen dasteht.
--
-- Dasselbe gilt andersherum: Eine selbst angelegte Aufgabe kennt bisher nur
-- „alle N Tage". Die gelbe Tonne, die donnerstags rausmuss, landete damit
-- irgendwo in der Woche.
--
-- Eine Zeile je Vorlage und Wochentag, Index 0 = Montag — dieselbe Form wie
-- bei `agreement`. Kein Wochenbezug: Der Tag gilt, bis jemand ihn ändert.
-- Keine Zeile heißt: Es gilt der Rhythmus aus der Bibliothek. Damit ist
-- „zurücksetzen" ein DELETE und braucht kein eigenes Kennzeichen — und es gibt
-- keinen dritten Zustand, in dem jemand einen Haken gesetzt hat und trotzdem
-- die Vorgabe gilt.
--
-- Kein FOREIGN KEY auf task_template: Kuratierte Vorlagen gehören keinem
-- Haushalt, und ihre Kennung ist Text aus dem Repo. Dieselbe Entscheidung wie
-- bei `agreement`, aus demselben Grund.

-- +goose Up

CREATE TABLE template_weekday (
    household_id uuid NOT NULL REFERENCES household (id) ON DELETE CASCADE,
    template_id  text NOT NULL,
    weekday      int  NOT NULL CHECK (weekday BETWEEN 0 AND 6),
    PRIMARY KEY (household_id, template_id, weekday)
);

-- +goose Down

DROP TABLE template_weekday;
