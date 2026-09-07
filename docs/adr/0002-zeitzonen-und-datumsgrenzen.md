# 0002 — UTC im Kern, `Europe/Berlin` an der Grenze

**Status:** angenommen · 7. September 2026

## Kontext

Eine Haushaltsanwendung ist eine Datumsanwendung. „Diese Woche", „heute
fällig", „alle drei Tage", „am dritten Sonntag im Monat" — jede dieser Aussagen
hängt daran, wo die Tagesgrenze liegt. Legt man sie nicht bewusst fest, legt
sie sich von selbst fest, und zwar an der Zeitzone des Servers.

Die Fehler, die daraus entstehen, sind unauffällig und teuer: Eine Aufgabe, die
sonntags um 23:30 Uhr abgehakt wird, landet in der Statistik der Folgewoche.
Ein Plan, erzeugt um 01:00 Uhr Sommerzeit, beginnt am falschen Montag. Ein
Nutzer im Urlaub in Bangkok sieht den Dienstag, während zu Hause noch Montag
ist. Zwei Tage im Jahr — bei der Zeitumstellung — hat ein Tag 23 oder 25
Stunden, und jede Rechnung mit „plus 24 Stunden" ist an diesen Tagen falsch.

Der Planer rechnet bereits heute mit Kalendertagen, nicht mit Zeitpunkten. Das
war eine stille Entscheidung im Code und wird hier zur ausdrücklichen.

## Entscheidung

**Der Planer kennt keine Uhrzeiten.** Er rechnet in Kalendertagen und ISO-Wochen
(`2026-W37`, Montag als erster Tag, nach ISO 8601). Wo im Kern ein `time.Time`
steht, ist es Mitternacht UTC und steht für einen Tag, nicht für einen
Zeitpunkt.

**Gespeichert wird in UTC.** Zeitpunkte in der Datenbank sind `timestamptz`;
reine Daten sind `date`. Nie eine Ortszeit ohne Zone speichern.

**Umgerechnet wird an der Grenze, nicht innen.** Die Zeitzone des Haushalts —
Vorgabe `Europe/Berlin`, später ein Feld an der Haushaltstabelle — bestimmt drei
Dinge: welcher Tag „heute" ist, wo die Woche beginnt, und wie ein Zeitstempel
angezeigt wird. Diese Umrechnung passiert in der HTTP-Schicht und im Browser,
nicht im Planer.

**„Heute" wird nie aus der Serverzeit abgeleitet.** Immer aus
`time.Now().In(haushaltszone)`. Ein Server in UTC hält zwischen 22:00 und
Mitternacht deutscher Sommerzeit bereits den Folgetag.

**In JSON stehen Daten als Zeichenketten:** `"2026-09-07"` für einen Tag,
`"2026-W37"` für eine Woche, RFC 3339 mit Zonenangabe für echte Zeitpunkte.
Kein Unix-Zeitstempel für etwas, das ein Kalendertag ist — genau dort entsteht
die Verwechslung.

## Konsequenzen

**Dafür:**

- Der Planer bleibt rein und deterministisch. Derselbe Haushalt, dieselbe Woche,
  dasselbe Ergebnis — unabhängig davon, wo und wann er läuft. Genau das macht
  ihn testbar.
- Die Zeitumstellung berührt den Kern nicht, weil er nie mit Stunden rechnet.
- Ein späterer Umzug oder ein Haushalt mit Mitgliedern in zwei Ländern ist ein
  Feld, keine Umbaumaßnahme.

**Dagegen:**

- Jede Grenze braucht eine bewusste Umrechnung. Vergisst man sie an einer
  Stelle, ist der Fehler still — er zeigt sich nur nachts und nur an manchen
  Tagen.
- Zeitfenster („morgens", „abends") sind damit im Kern reine Etiketten ohne
  Uhrzeit. Sobald sie zu echten Kapazitätsgrenzen werden sollen, braucht es
  einen eigenen Eintrag hier.
- Der Browser darf `new Date("2026-09-07")` nicht naiv anzeigen — das ist
  Mitternacht UTC und wird westlich von Greenwich zum Vortag.

## Verworfene Alternativen

**Alles in Ortszeit rechnen und speichern.** Einfacher zu lesen, solange alle im
selben Land sitzen. Verworfen: Bei der Zeitumstellung gibt es eine Stunde
doppelt und eine gar nicht, und Vergleiche zwischen gespeicherten Zeiten werden
mehrdeutig.

**Zeitzone pro Person statt pro Haushalt.** Näher an der Wirklichkeit einer
Fernbeziehung, aber der Wochenplan ist ein gemeinsames Objekt: Zwei Menschen
mit unterschiedlichen Wochengrenzen sähen unterschiedliche Pläne. Der Haushalt
ist die richtige Einheit; eine abweichende Anzeige pro Person kann später
obendrauf.

## Wann wir das revidieren

Wenn Aufgaben echte Uhrzeiten bekommen — „Kita-Abholung 16:30" statt „Dienstag
nachmittags". Dann wird aus dem Kalendertag ein Zeitpunkt, der Planer braucht
eine Zone, und dieser Eintrag wird durch einen neuen abgelöst.
