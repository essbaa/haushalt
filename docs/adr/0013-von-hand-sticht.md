# 0013 — Von Hand sticht: Umverteilen prüft Bedingungen, nicht Vorlieben

**Status:** angenommen · 14. September 2026 · setzt
[0012](0012-rotation-ist-eine-vorliebe.md) an der Grenze zum Menschen fort

## Kontext

Der Plan ist ein Vorschlag. Das steht so im Produktkonzept — „Ausführende
brauchen Handlungsmacht, nicht nur Pflichten. Eine reine Empfangsliste wird
gelöscht" — und bisher gab es davon genau eine Form: abgeben. Beim Abgeben
sucht die App den Nachfolger nach ihren eigenen Regeln. Wer jemand Bestimmtes
im Sinn hat, konnte das nicht sagen.

Die Spalte `assignment.manual` steht seit Migration 00001 im Schema und wurde
nie geschrieben. Sie ist die wertvollste Rückmeldung, die dieses Produkt
bekommen kann: Jede Korrektur von Hand sagt, wo der Planer danebenlag. Ein
Planer, der nie erfährt, dass er falsch lag, wird nie besser.

Damit stellt sich die Frage, die diesen Eintrag nötig macht: **Wie viel darf
der Planer gegen einen Menschen einwenden, der gerade eine Aufgabe
verschiebt?**

## Entscheidung

Der Planer prüft **Bedingungen** und keine **Vorlieben** — dieselbe
Unterscheidung wie in ADR-0012, hier an der Grenze zur Oberfläche.

**Geprüft wird, was die Person nicht übernehmen *kann*:**

- Alter (`MinAge` der Vorlage),
- Rolle und Verteilungsregel (`nur Erwachsene`, `nur Kinder`, betreute Personen
  führen nichts aus),
- dass eine eigene Aufgabe der Person selbst gehört (ADR-0005). „Dein Bett
  beziehen" bei jemand anderem ist eine andere Aufgabe, nicht dieselbe in
  anderen Händen.

**Nicht geprüft wird, was der Planer nur *bevorzugt* hätte:**

- Rotation über die Wochen und innerhalb der Woche,
- Auslastung und gerechter Anteil,
- Kopflast je Tag, Aufgaben je Tag, freie Minuten am Tag,
- die Längengrenze für Einzelaufgaben von Minderjährigen.

Ein Test hält das fest, und er ist der eigentliche Inhalt dieses Eintrags:
`TestUmverteilenStichtDieVorliebenDesPlaners`. Denn eine Korrektur, die der
Planer erst genehmigen muss, sagt nichts — sie bestätigt nur seine Annahmen.

**Erlaubt ist es der zuständigen Person und jeder planenden** — dieselbe Regel
wie beim Abhaken und Abgeben, keine neue.

**Die Zuteilung bekommt `von_hand` als Begründung und `manual = true`.** Die
alte Begründung stehen zu lassen wäre dieselbe Lüge wie beim Ausgleich
(ADR-0003): Nicht die Rotation hat entschieden, sondern ein Mensch.

**Geschrieben wird ein Ereignis `umverteilt`**, mit *von* und *an*. Das hat
eine Folge, die aus dem Schema kommt und nicht aus einer Absicht:
`DeleteUntouchedWeekTasks` verschont jede Aufgabe, an der ein Ereignis hängt —
**eine von Hand verteilte Aufgabe überlebt damit das Neurechnen der Woche.**
Das ist genau richtig und war nicht geplant. Es folgt aus der Regel aus
ADR-0008: Was Spuren hinterlassen hat, bleibt; was nur ein Vorschlag war, wird
neu gerechnet. Eine Umverteilung ist eine Spur.

## Konsequenzen

**Dafür:**

- Der Plan ist sichtbar ein Vorschlag. Das ist der Unterschied zwischen einer
  App, die man behält, und einem Dienstplan.
- `manual = true` sammelt ab jetzt echte Daten darüber, wo der Planer
  danebenliegt — die einzige Quelle dafür, die nicht aus Vermutungen besteht.
- Eine offene Aufgabe lässt sich mit demselben Handgriff übernehmen. Derselbe
  Knopf heißt dann „Übernehmen" statt „Wer macht das?".

**Dagegen:**

- **Die Woche kann von Hand unfair werden, und die Bilanz zeigt es.** Das ist
  die richtige Reihenfolge: Sie zeigt, was ist, nicht, was geplant war.
- **Eine planende Person kann eine Aufgabe auf jemanden schieben, der zu viel
  hat.** Die App hält nicht dagegen. Sie ist ein Werkzeug für einen Haushalt
  und keine Aufsichtsbehörde — und eine App, die einem Erwachsenen erklärt,
  er dürfe seinem eigenen Kind das Bad nicht geben, wird gelöscht.
- **`manual` wird gesammelt und noch nicht ausgewertet.** Der Planer lernt
  daraus vorerst nichts. Das ist ehrlicher als eine Auswertung auf zwölf
  Datenpunkten.

## Verworfene Alternativen

**Tauschen statt Verschieben** — zwei Aufgaben auswählen, die Personen
vertauschen. Erhält die Verteilung und ist doppelt so umständlich. Der
häufige Fall ist eine Aufgabe und eine Person („ich bin Samstag nicht da"),
nicht zwei und zwei.

**Nur für planende Personen.** Verworfen: Die eigene Aufgabe jemand Bestimmtem
zu geben ist genau die Handlungsmacht, die das Konzept verlangt — und abgeben
darf die zuständige Person längst. Die engere Regel wäre eine neue gewesen,
und eine Regel mehr braucht einen Grund.

**Nach der Umverteilung neu rechnen.** Klingt sauber und macht die Korrektur
zunichte: Der nächste Ausgleich schiebt die Aufgabe zurück, weil seine
Kennzahl das so will. Eine Korrektur, die der Planer wieder wegrechnet, ist
keine.

**Die Vorlieben als Warnung anzeigen** („Mia hat diese Woche schon viel").
Nicht verworfen, nur vertagt: Ein Hinweis neben dem Namen ist sinnvoll, ein
Hinweis, der den Klick kostet, nicht. Kommt, wenn die Auswahl mehr als vier
Personen zeigt.

## Wann wir das revidieren

Wenn `manual = true` genug Zeilen hat, um daraus zu lernen. Eine Vorlage, die
ein Haushalt dreimal hintereinander derselben Person gibt, ist eine feste
Zuständigkeit, die noch niemand eingetragen hat — und das kann die App fragen,
statt es zu raten (ADR-0007: raten ist in Ordnung, solange man widersprechen
kann).
