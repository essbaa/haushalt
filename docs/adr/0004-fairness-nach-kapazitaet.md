# 0004 — Fairness nach Kapazität, Kopflast als eigener Posten

**Status:** angenommen · 7. September 2026 · löst das Vergleichskriterium aus
[0003](0003-ausgleich-als-zweiter-durchgang.md) ab

## Kontext

Mit 62 Vorlagen und zwei Beispielhaushalten wurde zum ersten Mal sichtbar, was
die Kennzahl aus 0003 tatsächlich anrichtet. Zwei Befunde, beide aus dem
Wochenplan auf der Konsole und nicht aus den Tests.

**Erstens: gleiche Last ist nicht gleiche Belastung.** Familie B, vier Personen,
Woche 2026-W38:

| | Kapazität | verplant | ausgelastet |
|---|---|---|---|
| Nadja (46) | 550 min | 276 | 50 % |
| Tarek (48) | 540 min | 296 | 55 % |
| Jonas (15) | 345 min | 301 | 87 % |
| Mia (13) | 300 min | 280 | 93 % |

Die gewichteten Werte lagen zwischen 370 und 476 — auf dem Papier ein
ausgeglichener Plan. Tatsächlich war die Dreizehnjährige zu 93 Prozent
ausgebucht, während ihre Mutter die halbe Woche frei hatte. Der Ausgleich
strebte gleiche *absolute* Last an, und weil Jugendliche weniger Kapazität
haben, drückt genau das sie ans Limit. Mia hatte mit 17 Aufgaben die meisten
im ganzen Haushalt.

**Zweitens: die Kopflast entschied nie.** Familie A endete bei 737 zu 740
gewichteten Minuten — und Kopflast 22 zu 15. Nach 0003 ist die Kopflast das
zweite Kriterium, herangezogen bei Gleichstand des ersten. 737 ist nicht 740,
also kam sie nie zum Zug. Die Entscheidung war umgesetzt und wirkungslos, und
das ausgerechnet bei der Achse, auf der die These des Produkts steht.

## Entscheidung

**Der gerechte Anteil richtet sich nach der Kapazität.** Wer halb so viel Zeit
zur Verfügung hat, soll halb so viel tragen. Verglichen wird die Abweichung vom
kapazitätsgerechten Anteil, nicht vom Mittelwert über die Köpfe.

**Die Kopflast ist ein eigener Posten in derselben Summe**, keine
Gleichstandsregel mehr. Damit beide Abweichungen vergleichbar sind, wird ein
Kopflastpunkt in Minuten umgerechnet — mit `HeadLoadMinutes`, demselben
Wechselkurs, mit dem auch `PlannedTask.Weight` rechnet.

Die Kennzahl ist damit eine einzige Zahl statt eines zweistufigen Vergleichs:

	Σ  (Abweichung gewichtete Last)² + (Abweichung Kopflast × HeadLoadMinutes)²

**Auch die Zuteilung rechnet in Auslastung.** `rankMembers` fragt nicht mehr
nach der geringsten absoluten Last, sondern nach der geringsten Auslastung.
Sonst erzeugt der erste Durchgang eine Schieflage, die der zweite erst wieder
aufräumen muss — und was die Rotationsregeln blockieren, bliebe stehen.

**Die Auslastung wird sichtbar — aber die richtige.** `MemberLoad` trägt die
Wochenkapazität und liefert `Utilization()`: den Anteil der verfügbaren Zeit,
der verplant ist. Gerechnet wird dafür mit den reinen Minuten.

Die *Vergleichsgröße*, nach der verteilt wird, ist eine andere — gewichtete
Last im Verhältnis zur Kapazität — und sie bleibt im Paket. Der erste Anlauf
zeigte sie als Prozentzahl an und kam auf 113 Prozent: Kopflast wird nie gegen
Minuten verrechnet, eine Zahl mit Kopflast durch eine ohne zu teilen ergibt
keine Prozent. Der Fehler stand als Nachteil in diesem Eintrag, bevor er in der
Bilanz stand.

Gerechnet wird ganzzahlig: Statt durch die Gesamtkapazität zu teilen, wird jede
Abweichung mit ihr multipliziert. Der gemeinsame Faktor ändert die Reihenfolge
zweier Pläne nicht und erspart Fließkommazahlen im Kern.

## Konsequenzen

**Dafür:**

- Jugendliche und Teilzeit-Arbeitende bekommen einen Anteil, der zu ihrer Zeit
  passt, statt denselben absoluten Brocken wie alle anderen.
- Die Kopflast wirkt bei jeder Entscheidung mit, nicht nur bei rechnerischem
  Gleichstand.
- „Ausgelastet 62 %" ist eine Zahl, die man einem Haushalt zeigen kann.
  „Gewichtet 737" ist es nicht.

**Dagegen:**

- **Die Kapazitätswerte werden zur tragenden Größe.** Bisher waren sie eine
  Obergrenze, jetzt bestimmen sie die Verteilung. Sie sind aber geschätzt — aus
  Arbeitszeiten grob abgeleitet. Wer seine Kapazität zu niedrig angibt,
  bekommt dauerhaft weniger zu tun. Das ist eine offene Flanke und gehört
  beobachtet, sobald echte Haushalte Zahlen eintragen.
- **Quadrieren gewichtet Ausreißer stärker.** Eine Person, die weit daneben
  liegt, zieht mehr Aufmerksamkeit auf sich als drei, die leicht daneben
  liegen. Meist gewollt, aber es ist eine Setzung.
- **Die Auslastung mischt Einheiten.** Gewichtete Minuten enthalten
  Kopflastpunkte, die keine echte Zeit sind; geteilt durch verfügbare Minuten
  ergibt das keine Prozent im strengen Sinn. Für den Vergleich zweier Personen
  taugt es, als physikalische Größe nicht.
- Wer wenig Zeit hat, bekommt wenig Aufgaben — auch wenn er gern mehr täte.
  Ein „ich kann diese Woche mehr" fehlt im Modell.

## Verworfene Alternativen

**Toleranzband statt Summe:** Unterschiede unter fünf Prozentpunkten als gleich
behandeln, dann die Kopflast entscheiden lassen. Anschaulich erklärbar,
verworfen wegen der fünf — eine gegriffene Zahl im Kern, die niemand begründen
kann und die bei jedem zweiten Plan anders wirkt.

**Kopflast strikt zuerst.** Schon in 0003 verworfen, aus demselben Grund: Ein
Plan, in dem die Kopflast perfekt aufgeht und eine Person dafür zwei Stunden
mehr putzt, wird nicht als fair empfunden.

**Auslastung nur im zweiten Durchgang.** Weniger Änderung, aber der gierige
Durchgang erzeugte die Schieflage dann weiterhin, und was die Rotationsregeln
blockieren, könnte der Ausgleich nicht mehr reparieren.

## Wann wir das revidieren

Wenn die Kapazität nicht mehr geschätzt, sondern aus echten Daten kommt — etwa
aus einem verknüpften Kalender. Dann ändert sich, wie belastbar die Zahl ist,
und damit die Frage, wie stark man sie tragen lassen darf.

Und wenn Haushalte melden, dass der Anteil der Jugendlichen zu klein wirkt: Ein
Fünfzehnjähriger, der 87 Prozent trug, war zu viel — 20 Prozent könnten zu
wenig sein. Das ist eine Kalibrierung, keine neue Entscheidung, aber sie gehört
gemessen statt geraten.
