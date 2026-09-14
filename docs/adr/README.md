# Entscheidungen

Ein Eintrag je Entscheidung, die sich nicht aus dem Code ablesen lässt.
Durchnummeriert, unveränderlich: Wer eine Entscheidung revidiert, schreibt
einen neuen Eintrag und setzt den alten auf „abgelöst durch ####" — er wird
nicht überschrieben. Der Wert steckt in der Kette, nicht im letzten Stand.

| Nr. | Titel | Status |
|---|---|---|
| [0001](0001-go-neben-nextjs.md) | Go-Dienst neben Next.js statt Next.js allein | angenommen |
| [0002](0002-zeitzonen-und-datumsgrenzen.md) | UTC im Kern, `Europe/Berlin` an der Grenze | angenommen |
| [0003](0003-ausgleich-als-zweiter-durchgang.md) | Ausgleich als zweiter Durchgang, Abstand vor Kopflast | Kriterium abgelöst durch 0004, Rotationsregel durch 0012 |
| [0004](0004-fairness-nach-kapazitaet.md) | Fairness nach Kapazität, Kopflast als eigener Posten | angenommen |
| [0005](0005-aufgaben-die-einer-person-gehoeren.md) | Aufgaben, die einer Person selbst gehören | angenommen |
| [0006](0006-anmeldung-better-auth.md) | Anmeldung mit Better Auth in der Web-App, Go verifiziert | angenommen |
| [0007](0007-onboarding-fragt-drei-dinge.md) | Das Onboarding fragt drei Dinge und rät den Rest | angenommen |
| [0008](0008-woche-festschreiben.md) | Die Woche wird beim ersten Ansehen festgeschrieben | angenommen |
| [0009](0009-gestaltung-muss-unterscheiden.md) | Ein Token-System, und Gestaltung, die unterscheidet | angenommen |
| [0010](0010-unbekannt-ist-nicht-nein.md) | Unbekannt ist nicht nein | angenommen |
| [0011](0011-massstab-statt-filter.md) | Maßstab statt Filter | angenommen |
| [0012](0012-rotation-ist-eine-vorliebe.md) | Rotation ist eine Vorliebe, keine Bedingung | angenommen |
| [0013](0013-von-hand-sticht.md) | Von Hand sticht: Umverteilen prüft Bedingungen, nicht Vorlieben | angenommen |
| [0014](0014-anmeldung-bleibt-bei-mail.md) | Anmeldung bleibt bei E-Mail und Passwort | angenommen |

## Aufbau eines Eintrags

**Status** · **Kontext** (die Lage, die die Entscheidung erzwingt — Fakten,
noch keine Wertung) · **Entscheidung** (ein Satz im Aktiv) · **Konsequenzen**
(positiv *und* negativ) · **Verworfene Alternativen** · **Wann wir das
revidieren**.

Die letzten beiden Blöcke sind der Unterschied zwischen einem Eintrag, der
später hilft, und einer nachträglichen Rechtfertigung. Ein Eintrag ohne
Nachteile ist keine Entscheidung, sondern Werbung.
