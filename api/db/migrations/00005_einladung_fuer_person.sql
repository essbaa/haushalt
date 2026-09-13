-- Eine Einladung kann auf eine Person zeigen, die es schon gibt.
--
-- Seit dem Onboarding (T6) stehen Partner und Kinder von Anfang an im Plan,
-- ohne Anmeldung: Die faire Verteilung soll am ersten Tag sichtbar sein, auch
-- wenn der andere die App nie geöffnet hat. Eine Einladung legte bisher aber
-- immer eine neue Person an — aus „Asmae" wurde beim Beitritt „Asmae (2)",
-- mit halber Kapazität und ohne den Verlauf der ersten.
--
-- member_id sagt: Dieser Code macht dich zu *dieser* Person. Bleibt er leer,
-- kommt jemand dazu, den es noch nicht gab.

-- +goose Up

ALTER TABLE invitation
    ADD COLUMN member_id uuid REFERENCES member (id) ON DELETE CASCADE;

-- Ein Mensch, der schon ein Konto hat, wird nicht noch einmal übernommen —
-- dafür ist der Anspruch beim Beitritt bedingt formuliert (auth_user_id IS
-- NULL). Zwei offene Einladungen für dieselbe Person sind dagegen erlaubt:
-- Der erste Code gewinnt, der zweite läuft ins Leere. Das ist der Fall
-- „Link nochmal geschickt, weil der erste in einem alten Chat liegt".
CREATE INDEX invitation_member_idx ON invitation (member_id);

-- +goose Down

DROP INDEX IF EXISTS invitation_member_idx;
ALTER TABLE invitation DROP COLUMN IF EXISTS member_id;
