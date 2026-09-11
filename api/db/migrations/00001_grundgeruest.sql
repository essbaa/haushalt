-- Das Grundgerüst: sieben Tabellen.
--
-- Mandant ist der Haushalt. Jede Fachabfrage filtert auf household_id — das
-- ist die einzige Trennlinie, die es zwischen zwei Familien gibt, und sie muss
-- deshalb überall stehen.
--
-- Die Anmeldetabellen gehören NICHT hierher. Better Auth verwaltet user,
-- session, account und verification aus TypeScript heraus und mit eigenen
-- Migrationen. goose fasst sie nie an; die Verbindung zwischen beiden Welten
-- ist die einzige Spalte member.auth_user_id.

-- +goose Up

-- ------------------------------------------------------------------ Haushalt

CREATE TABLE household (
    id          uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    name        text        NOT NULL,
    home        text        NOT NULL CHECK (home IN ('wohnung', 'haus')),
    has_car     boolean     NOT NULL DEFAULT false,
    has_yard    boolean     NOT NULL DEFAULT false,
    pets        text[]      NOT NULL DEFAULT '{}',
    -- Zeitzone am Haushalt, nicht an der Person (ADR-0002): Der Wochenplan ist
    -- ein gemeinsames Objekt, zwei Wochengrenzen ergäben zwei Pläne.
    timezone    text        NOT NULL DEFAULT 'Europe/Berlin',
    created_at  timestamptz NOT NULL DEFAULT now()
);

-- ----------------------------------------------------------------- Mitglied

CREATE TABLE member (
    id           uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    household_id uuid NOT NULL REFERENCES household (id) ON DELETE CASCADE,
    name         text NOT NULL,
    role         text NOT NULL CHECK (role IN ('planend', 'ausfuehrend', 'betreut')),

    -- Geburtsjahr, nicht Geburtsdatum. Der Planer braucht das Alter für
    -- Altersgrenzen; der Tag und der Monat tragen dazu nichts bei. Bei
    -- Kinderdaten ist das nicht Sparsamkeit aus Ordnungsliebe, sondern die
    -- Regel: Was nicht erhoben wird, kann nicht verloren gehen.
    birth_year   int  CHECK (birth_year BETWEEN 1900 AND 2200),
    care         text CHECK (care IN ('keine', 'kita', 'schule')),

    -- Verfügbare Minuten je Wochentag, Index 1 = Montag (Postgres zählt ab 1).
    -- Seit ADR-0004 bestimmt diese Zahl die Verteilung und ist damit mehr als
    -- eine Obergrenze.
    capacity_minutes int[] NOT NULL DEFAULT '{0,0,0,0,0,0,0}'
        CHECK (array_length(capacity_minutes, 1) = 7),

    -- Die Kennung aus Better Auth. NULL heißt: Diese Person hat keinen Zugang
    -- — ein Kleinkind erzeugt Arbeit, ohne sich anzumelden.
    auth_user_id text UNIQUE,

    created_at   timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX member_household_idx ON member (household_id);

-- ------------------------------------------------------------------ Vorlage

-- Die vierzehn Felder einer Vorlage stehen als JSON in definition, nicht als
-- fünfundzwanzig Spalten.
--
-- Grund: Niemand filtert in SQL nach Rhythmus oder Kopflast. Der Planer liest
-- die ganze Vorlage und entscheidet in Go — dieselbe Struktur, die auch aus
-- der Bibliothek im Repo kommt, also derselbe Codepfad. Sobald eine Abfrage
-- ein Feld wirklich braucht, wird es eine eigene Spalte; bis dahin wäre sie
-- Ballast, der bei jeder Modelländerung eine Migration erzwingt.
CREATE TABLE task_template (
    id           text PRIMARY KEY,
    -- NULL = kuratierte Vorlage aus dem Repo, gilt für alle.
    household_id uuid REFERENCES household (id) ON DELETE CASCADE,
    source       text NOT NULL CHECK (source IN ('kuratiert', 'haushalt', 'gelernt')),
    -- Stand der Bibliothek, aus der sie importiert wurde.
    version      text NOT NULL DEFAULT '',
    definition   jsonb NOT NULL,
    created_at   timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX task_template_household_idx ON task_template (household_id);

-- -------------------------------------------------------------- Aufgabe

CREATE TABLE task_instance (
    id           uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    household_id uuid NOT NULL REFERENCES household (id) ON DELETE CASCADE,
    template_id  text NOT NULL REFERENCES task_template (id),

    -- ISO-Kalenderwoche als Zeichenkette, genau wie im Vertrag: '2026-W38'.
    iso_week     text NOT NULL CHECK (iso_week ~ '^\d{4}-W\d{2}$'),
    -- Ein Kalendertag, kein Zeitpunkt (ADR-0002).
    day          date NOT NULL,

    slot         text NOT NULL CHECK (slot IN ('egal', 'morgens', 'abends')),
    duration_min int  NOT NULL CHECK (duration_min > 0),
    head_load    int  NOT NULL CHECK (head_load BETWEEN 0 AND 3),
    deadline     date,

    created_at   timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX task_instance_woche_idx ON task_instance (household_id, iso_week);

-- ---------------------------------------------------------------- Zuteilung

CREATE TABLE assignment (
    id               uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    task_instance_id uuid NOT NULL REFERENCES task_instance (id) ON DELETE CASCADE,
    member_id        uuid NOT NULL REFERENCES member (id) ON DELETE CASCADE,

    reason_code      text NOT NULL CHECK (reason_code IN
        ('rotation', 'ausgleich', 'feste_person', 'einzige_moeglichkeit', 'frist', 'eigene_aufgabe')),
    reason_previous  uuid REFERENCES member (id) ON DELETE SET NULL,

    -- Von Hand geändert oder vom Planer gesetzt. Der Unterschied ist die
    -- wichtigste Rückmeldung, die das Produkt bekommt: Jede Korrektur sagt,
    -- wo der Planer danebenlag.
    manual           boolean NOT NULL DEFAULT false,

    created_at       timestamptz NOT NULL DEFAULT now()
);

CREATE UNIQUE INDEX assignment_je_aufgabe_idx ON assignment (task_instance_id);
CREATE INDEX assignment_member_idx ON assignment (member_id);

-- ------------------------------------------------------------- Protokoll

-- Reines Anhängen. Erledigt, verschoben, abgegeben, gelöscht — nur schreiben,
-- nie ändern.
--
-- Damit werden Lernschleife, Aktivitätsanzeige und spätere Auswertungen zu
-- Leseabfragen auf derselben Quelle statt zu drei Mechanismen, die
-- auseinanderlaufen. Der Trigger unten macht aus dieser Absicht eine Regel:
-- Ein UPDATE oder DELETE scheitert, auch wenn jemand es später bequem fände.
CREATE TABLE event (
    id               bigint GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    household_id     uuid NOT NULL REFERENCES household (id) ON DELETE CASCADE,
    member_id        uuid REFERENCES member (id) ON DELETE SET NULL,
    task_instance_id uuid REFERENCES task_instance (id) ON DELETE SET NULL,

    kind             text NOT NULL CHECK (kind IN
        ('erstellt', 'erledigt', 'verschoben', 'abgegeben', 'geloescht', 'wieder_geoeffnet')),
    payload          jsonb NOT NULL DEFAULT '{}',
    occurred_at      timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX event_household_idx ON event (household_id, occurred_at DESC);

-- +goose StatementBegin
CREATE FUNCTION event_ist_unveraenderlich() RETURNS trigger AS $$
BEGIN
    RAISE EXCEPTION 'event ist ein reines Anhänge-Protokoll: % ist nicht erlaubt', TG_OP;
END;
$$ LANGUAGE plpgsql;
-- +goose StatementEnd

CREATE TRIGGER event_kein_update BEFORE UPDATE OR DELETE ON event
    FOR EACH ROW EXECUTE FUNCTION event_ist_unveraenderlich();

-- -------------------------------------------------------------- Lernwerte

-- Aus dem Protokoll verdichtet: Wie oft wird eine Vorlage gelöscht, wie lange
-- dauert sie wirklich, wer übernimmt sie freiwillig. Ableitbar, nicht Quelle —
-- diese Tabelle darf jederzeit neu berechnet werden.
CREATE TABLE signal (
    household_id uuid NOT NULL REFERENCES household (id) ON DELETE CASCADE,
    template_id  text NOT NULL REFERENCES task_template (id) ON DELETE CASCADE,
    kind         text NOT NULL,
    value        double precision NOT NULL,
    updated_at   timestamptz NOT NULL DEFAULT now(),
    PRIMARY KEY (household_id, template_id, kind)
);

-- +goose Down

DROP TABLE IF EXISTS signal;
DROP TRIGGER IF EXISTS event_kein_update ON event;
DROP FUNCTION IF EXISTS event_ist_unveraenderlich();
DROP TABLE IF EXISTS event;
DROP TABLE IF EXISTS assignment;
DROP TABLE IF EXISTS task_instance;
DROP TABLE IF EXISTS task_template;
DROP TABLE IF EXISTS member;
DROP TABLE IF EXISTS household;
