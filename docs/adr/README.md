# Entscheidungen

Ein Eintrag je Entscheidung, die sich nicht aus dem Code ablesen lässt.
Durchnummeriert, unveränderlich: Wer eine Entscheidung revidiert, schreibt
einen neuen Eintrag und setzt den alten auf „abgelöst durch ####" — er wird
nicht überschrieben. Der Wert steckt in der Kette, nicht im letzten Stand.

| Nr. | Titel | Status |
|---|---|---|
| [0001](0001-go-neben-nextjs.md) | Go-Dienst neben Next.js statt Next.js allein | angenommen |
| [0002](0002-zeitzonen-und-datumsgrenzen.md) | UTC im Kern, `Europe/Berlin` an der Grenze | angenommen |
| [0003](0003-ausgleich-als-zweiter-durchgang.md) | Ausgleich als zweiter Durchgang, Abstand vor Kopflast | Kriterium abgelöst durch 0004 |
| [0004](0004-fairness-nach-kapazitaet.md) | Fairness nach Kapazität, Kopflast als eigener Posten | angenommen |
| [0005](0005-aufgaben-die-einer-person-gehoeren.md) | Aufgaben, die einer Person selbst gehören | angenommen |
| [0006](0006-anmeldung-better-auth.md) | Anmeldung mit Better Auth in der Web-App, Go verifiziert | angenommen |

## Aufbau eines Eintrags

**Status** · **Kontext** (die Lage, die die Entscheidung erzwingt — Fakten,
noch keine Wertung) · **Entscheidung** (ein Satz im Aktiv) · **Konsequenzen**
(positiv *und* negativ) · **Verworfene Alternativen** · **Wann wir das
revidieren**.

Die letzten beiden Blöcke sind der Unterschied zwischen einem Eintrag, der
später hilft, und einer nachträglichen Rechtfertigung. Ein Eintrag ohne
Nachteile ist keine Entscheidung, sondern Werbung.
