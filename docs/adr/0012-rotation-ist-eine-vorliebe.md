# 0012 — Rotation ist eine Vorliebe, keine Bedingung

**Status:** angenommen · 14. September 2026 · hebt einen der vier
unantastbaren Punkte aus [0003](0003-ausgleich-als-zweiter-durchgang.md) auf
und korrigiert die Diagnose in [0011](0011-massstab-statt-filter.md)

## Kontext

In ADR-0011 steht unter „Offen geblieben", der Ausgleich könne Aufgaben nur
ganz verschieben und überschieße deshalb bei kleiner Kapazität systematisch —
weshalb in Familie B ausgerechnet die dreizehnjährige Mia am höchsten
ausgelastet sei (79 % ihrer Zeit, ihr Vater 67 %).

**Das war geraten, und es war falsch.** Vier Messungen am echten Plan:

| Eingriff | Mia | Tarek |
|---|---|---|
| nichts (Ausgangslage) | 79 % | 67 % |
| Dreierzyklen statt Paartausche | 79 % | 67 % |
| Minutenabweichung zusätzlich bestraft | 79 % | 67 % |
| Kapazitätsprüfung abgeschaltet | 80 % | 69 % |
| **Wochen-Rotation im Ausgleich abgeschaltet** | **74 %** | **70 %** |

Der Ausgleich macht bei Familie B fünf Tausche und drückt die Kennzahl um
85 %. Er steckt danach nicht fest, weil Aufgaben zu grob sind — er steckt
fest, weil die Rotation ihn festhält. Der blockierte Tausch im Klartext:

```
Böden wischen (Mia) <-> Wäsche zusammenlegen (Tarek): Kennzahl 3,92e9 -> 2,89e9
```

**Mia wischt die Böden, weil ihr Vater sie letzte Woche gewischt hat.** Die
Rotation schützt den Erwachsenen, und bezahlt wird mit der Zeit des Kindes.

Der Befund hat eine zweite Hälfte, und die ist die eigentliche: **Dieselbe
Regel hatte zwei Stärken.** Im `assign` ist die Rotation eine Vorliebe —
`rankMembers` sortiert die Person von letzter Woche nach hinten und nimmt sie
trotzdem, wenn sonst niemand kann. Im `rebalance` war sie ein Veto. Die
härtere Fassung stand ausgerechnet in dem Durchgang, der ausgleichen soll.
Absolut war die Rotation ohnehin nie: Im Plan von Familie B stehen schon vor
jeder Änderung 5 Wiederholungen bei 52 Aufgaben.

Das ist auch kein neues Wissen. In ADR-0003 steht es unter „Dagegen":

> **Die Rotationsregel begrenzt, was reparabel ist.** Die größten Schieflagen
> entstehen genau dadurch, dass Rotation vor Ausgleich geht.

Was fehlte, war die Zahl daneben — und wer die Rechnung bezahlt.

## Entscheidung

Der Ausgleich läuft in zwei Durchgängen.

**Erster Durchgang:** wie bisher, die Wochen-Rotation ist unantastbar.

**Zweiter Durchgang:** derselbe Suchlauf, aber die Wochen-Rotation darf
gebrochen werden — und nur, solange jemand um mehr als `MaxOvershootPermille`
seiner **eigenen** Kapazität über seinem gerechten Anteil liegt. Sobald das
nicht mehr zutrifft, hört er auf.

`MaxOvershootPermille` steht in den `Limits`, ist also eine Produktaussage und
keine Konstante im Code. Der Startwert ist 25 ‰ — etwa eine kleine Aufgabe zu
viel. Null schaltet den zweiten Durchgang ab; dann gilt wieder ADR-0003
unverändert.

Der Maßstab ist bewusst die eigene Kapazität und nicht der Haushalt. 20
gewichtete Minuten zu viel sind für eine Dreizehnjährige etwas anderes als für
ihren Vater, und genau dieser Unterschied ist der Grund für diesen Durchgang.

Die drei anderen unantastbaren Punkte aus ADR-0003 bleiben unantastbar:
Eignung, Kapazität, Rotation **innerhalb** der Woche. Niemand kocht dreimal in
einer Woche, auch wenn die Zahlen dadurch aufgingen — dieselbe Person zwei
Wochen hintereinander am Bad ist etwas anderes als dreimal Abendessen in
sieben Tagen.

## Konsequenzen

**Dafür:**

- Familie B: 74 / 70 / 70 / 74 statt 74 / 67 / 70 / 79. Die Spanne fällt von
  12 auf 4 Prozentpunkte, und das Kind ist nicht mehr die am höchsten
  ausgelastete Person im Haushalt.
- Die Regel hat wieder eine einzige Stärke. „Vorliebe mit Rückfallebene" ist
  das Muster, das im Planer ohnehin überall gilt — `earliestFreeDay` macht es
  mit der Kopflast je Tag, `nichtZuLangFuerKinder` mit der Aufgabenlänge.
- Familie A bleibt unverändert (60 / 60 %). Wo der erste Durchgang reicht,
  passiert nichts.

**Dagegen:**

- **Wiederholungen steigen: 5 → 7 von 52 Aufgaben.** Zwei Menschen machen
  dieselbe Sache zwei Wochen hintereinander, damit ein Kind fünf Prozentpunkte
  weniger trägt. Das ist der Tausch, und er ist bewusst so herum.
- Eine Zahl mehr in den `Limits`, und sie ist geraten wie die anderen. 25 ‰ ist
  ein Startwert, kein Messergebnis.
- Der Ausgleich läuft im schlechtesten Fall doppelt so lange. Bei fünfzig
  Aufgaben irrelevant.
- „Rotation geht vor Ausgleich" stimmt jetzt nur noch mit Nachsatz. Der Satz
  war griffig; er war nur nicht die ganze Regel.

## Verworfene Alternativen

**Große Aufgaben zuerst verteilen** (das klassische Mittel gegen gierige
Zuteilung). Gemessen, nicht vermutet: Mia geht damit auf **82 %**, Jonas
verliert ein Drittel seiner Last. Die Reihenfolge nach Dringlichkeit ist nicht
zufällig — sie entscheidet auch über den Tag.

**Eine eigene Obergrenze für Minderjährige** („niemand unter 18 über dem
Haushaltsschnitt"). Behebt den sichtbaren Fall und lässt die Erwachsenen
untereinander schief. Und es hätte eine Regel für Kinder eingeführt, wo das
Problem keine Frage des Alters ist, sondern der Kapazität: Es trifft genauso
den Elternteil in Teilzeit und den, der krank ist.

**Nur die Anzeige ändern.** Verteilt wird nach gewichteter Last, und da ist der
Plan auf ±4 % fair; die 79 % sind reine Zeit und genau die Größe, nach der
*nicht* verteilt wird. Man hätte also die Zahl erklären können statt den Planer
zu ändern. Verworfen, weil die Zahl richtig ist: Mias Zeit war wirklich zu 79 %
verplant. Eine Erklärung, warum das in Ordnung sein soll, ist eine Ausrede.

**Die Rotation ganz aus dem Ausgleich nehmen.** Ergibt dieselben Zahlen wie der
zweite Durchgang, aber ohne Schranke — dann würde eine Wiederholung auch für
eine Verbesserung im Promillebereich in Kauf genommen, und aus „Rotation vor
Ausgleich" wäre unbemerkt das Gegenteil geworden.

## Wann wir das revidieren

Wenn Haushalte zurückmelden, dass sie Wiederholungen schlimmer finden als
Schieflagen — dann steigt `MaxOvershootPermille`, nicht der Code.

Wenn Aufgaben teilbar werden („Bad putzen, 20 von 35 Minuten"), wird die
Schranke seltener überschritten, und der zweite Durchgang läuft von selbst
kaum noch.
