# 0005 — Aufgaben, die einer Person selbst gehören

**Status:** angenommen · 7. September 2026

## Kontext

Im Wochenplan von Familie A stand „Eigenes Zimmer aufräumen" bei Anna (36), in
Familie B bei Tarek (48). Beides war nach den Regeln des Planers korrekt: Die
Vorlage rotiert, `ab_alter` grenzt nur Kinder ein, Erwachsene sind immer
geeignet — und wer gerade die niedrigste Auslastung hat, bekommt sie.

Dahinter steht keine Panne, sondern eine Lücke im Modell. Es gibt eine ganze
Klasse von Aufgaben, die *jedem selbst* gehören: das eigene Zimmer, die eigene
Wäsche, die eigene Schultasche, die eigenen Hausaufgaben. Für sie sind beide
Grundannahmen des Planers falsch. Sie rotieren nicht — niemand räumt
abwechselnd das Zimmer eines anderen auf. Und es gibt sie nicht einmal pro
Haushalt, sondern einmal pro Person.

## Entscheidung

Zwei getrennte Angaben an der Vorlage, weil es zwei getrennte Fragen sind:

**`Distribution: nur_kinder`** — das Gegenstück zu `nur_erwachsene`. Wer die
Aufgabe übernehmen darf.

**`PerPerson bool`** — ob aus einem Termin eine Aufgabe je berechtigter Person
wird, fest an sie gebunden. Wie viele es gibt.

Getrennt, weil sie sich kreuzen: „eigene Wäsche" gilt auch für Erwachsene,
„Hausaufgaben begleiten" ist Kindersache und trotzdem eine gemeinsame Aufgabe
der Eltern.

Personengebundene Termine nehmen nicht an der Rotation teil, können vom
Ausgleich nicht getauscht werden und bekommen die Begründung
`eigene_aufgabe`.

Nebenbei behoben, weil es dieselbe Sorte Lücke war: **Bedingungen können nach
der Tierart fragen** (`haustier_art: "hund"`). Vorher war es ein Ja/Nein — der
Haushalt wusste, dass er einen Hund hat, die Vorlage konnte aber nicht danach
fragen. Das Katzenklo stand viermal die Woche im Plan eines Hundehaushalts.

## Konsequenzen

**Dafür:**

- Aufgaben landen dort, wo sie hingehören, statt bei der Person mit der gerade
  niedrigsten Auslastung.
- Zwei Kinder ergeben zwei Aufgaben. In der Bilanz erscheinen sie bei beiden —
  richtig so, denn beide arbeiten.

**Dagegen:**

- **Die Last vervielfacht sich mit der Haushaltsgröße.** Vier Personen mit
  eigener Wäsche ergeben vier Aufgaben statt einer. Das ist gewollt und wird
  trotzdem dazu führen, dass große Haushalte volle Pläne bekommen — die
  Startdichte begrenzt bisher nur die Kopfarbeit.
- **Der Ausgleich kann diese Aufgaben nicht mehr anfassen.** Wer viele
  personengebundene Aufgaben hat, trägt sie; die Verteilung kann nur noch am
  Rest drehen.
- Eine Vorlage kann nicht ausdrücken, dass Eltern jüngeren Kindern dabei
  helfen. Für ein Kind, das sein Zimmer noch nicht allein aufräumt, gibt es
  weiterhin nur `ab_alter`.

## Verworfene Alternativen

**Nur `nur_kinder`, ohne Vervielfachung.** Die Hälfte der Arbeit: Die Aufgabe
wäre nicht mehr bei den Eltern gelandet, aber zwischen zwei Geschwistern
rotiert — Jonas räumt diese Woche Mias Zimmer auf. Sichtbar falsch.

**Aufgaben an Personen statt an Haushalte hängen**, also ein eigenes Feld
`gehoert_zu` mit einer Mitglieds-ID. Verworfen, weil Vorlagen kuratiert sind
und den Haushalt nicht kennen. Die Bindung entsteht erst beim Planen — genau
dort passiert sie jetzt auch.

## Wann wir das revidieren

Wenn Haushalte melden, dass ihre Pläne zu voll sind. Dann ist die
Vervielfachung ein Kandidat für die Startdichte — und die müsste dafür lernen,
Ausführungsaufgaben ebenfalls zu begrenzen.
