# 0015 — Streichen trifft einen Termin, Abschalten eine Vorlage

**Status:** angenommen · 14. September 2026

## Kontext

Es gab genau einen Ausweg aus einer Aufgabe: „Brauchen wir nicht", und das
schaltet die Vorlage **für immer** ab. Damit war jede Ablehnung eine
Entscheidung über alle künftigen Wochen — auch die, die nur „heute passt es
nicht" meinte.

Das ist vor einer Probewoche kein Schönheitsfehler. Der eine Fehlgriff nimmt
dem Haushalt eine Vorlage weg, die er eigentlich wollte, und es fällt erst
nächste Woche auf, wenn sie fehlt. Ein Aufbau, der beim Ausprobieren stillschweigend
kaputtgeht, ist nicht zu bewerten.

## Entscheidung

**Streichen gilt für einen Termin, Abschalten für eine Vorlage.** Zwei
Handlungen, zwei Reichweiten, zwei Orte.

Gestrichen wird über zwei neue Ereignisarten (`gestrichen`,
`wieder_eingeplant`, Migration 00012) — nicht über eine Spalte. Wie bei
„erledigt": Nur das Protokoll kann erzählen, dass jemand es sich anders
überlegt hat. Bewusst nicht das vorhandene `geloescht`: Das Wort gehört dem
Löschen, nicht dem Aussetzen, und zwei Absichten unter einem Namen sind
später nicht mehr zu trennen (dieselbe Überlegung wie bei `verschoben` in
00010).

**Gestrichenes verschwindet aus Plan und Bilanz** — wer einen Termin streicht,
soll ihn nicht als Last angerechnet bekommen. Aber nicht wortlos: Am Ende der
Woche steht „Gestrichen" mit dem Tag daneben und „Doch wieder einplanen". Eine
Zeile, die spurlos verschwindet, lässt jemanden, der danebengewischt hat, ohne
Rückweg.

**Die Geste hat gewechselt.** Wischen nach links heißt jetzt „Diesmal nicht";
„Brauchen wir nicht" steht nur noch in den aufgeklappten Aktionen und nur für
planende Personen.

> **Die unumkehrbare Handlung darf nicht die sein, die man aus Versehen
> macht.**

Das ist keine Geschmacksfrage. Eine Wischgeste ist schnell, ungenau und wird
im Vorbeigehen ausgeführt — sie gehört der Handlung, die man zurücknehmen
kann.

## Konsequenzen

**Dafür:**

- Ein Fehlgriff kostet einen Termin, nicht eine Vorlage.
- Der Aufbau überlebt eine Woche Ausprobieren, und damit ist die Woche
  überhaupt auswertbar.
- Das Protokoll erzählt beides: dass es den Termin gab und dass er gestrichen
  wurde. Später ist das eine Aussage über die Bibliothek — eine Vorlage, die
  ein Haushalt jede Woche streicht, ist falsch kuratiert.

**Dagegen:**

- Eine Ereignisart mehr, und damit eine Migration, die vor dem Deploy laufen
  muss.
- Zwei ähnliche Handlungen nebeneinander. Wer den Unterschied nicht liest,
  hält sie für dasselbe — die Beschriftung trägt hier mehr Gewicht als der
  Code.
- „Brauchen wir nicht" ist einen Schritt weiter weg. Das ist Absicht und
  trotzdem ein Verlust für den, der genau das wollte.

## Verworfene Alternativen

**Eine Spalte `gestrichen` auf der Aufgabe.** Einfacher zu lesen und eine
zweite Wahrheit neben dem Protokoll. Dieselbe Begründung wie bei „erledigt".

**`geloescht` wiederverwenden.** Hätte die Migration gespart und zwei
verschiedene Absichten unter einem Namen begraben.

**Nur ein Ausweg, dafür mit Rückfrage** („Wirklich für immer?"). Ein Dialog
unterbricht genau die Geste, die schnell sein soll — und beantwortet die
falsche Frage: Die meisten wollen nicht „wirklich für immer", sondern „heute
nicht".

## Wann wir das revidieren

Wenn sich zeigt, dass niemand „Brauchen wir nicht" findet, weil es nur noch in
den aufgeklappten Aktionen steht. Dann braucht die Aufgaben-Seite den Weg
dorthin deutlicher — nicht die Wischgeste zurück.
