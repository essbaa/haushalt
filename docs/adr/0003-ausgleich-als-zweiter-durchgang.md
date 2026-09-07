# 0003 — Ausgleich als zweiter Durchgang, Abstand vor Kopflast

**Status:** angenommen · 7. September 2026 · das Vergleichskriterium ist
abgelöst durch [0004](0004-fairness-nach-kapazitaet.md) — der Durchgang
selbst, seine Regeln und seine Grenzen gelten unverändert weiter

## Kontext

Die Zuteilung ist gierig: `assign` geht die fälligen Aufgaben einmal in
Dringlichkeitsreihenfolge durch und gibt jede der Person, die in *diesem
Moment* am wenigsten trägt. Was einmal zugeteilt ist, wird nie wieder
angefasst.

Das erzeugt Ergebnisse, die messbar schief sind. Drei Aufgaben, zwei
Erwachsene, keine Historie: zwei Organisationsaufgaben zu je 10 Minuten mit
Kopflast 3 (gewichtet je 55) und eine Putzaufgabe über 110 Minuten. Der gierige
Durchgang verteilt die beiden kleinen Aufgaben auf beide Personen und legt die
große danach auf die Person, die zufällig zuerst dran war: **165 zu 55**.
Möglich wäre **110 zu 110** — beide Organisationsaufgaben bei einer Person, die
Putzaufgabe bei der anderen.

Dazu kommt eine zweite Blindheit: Der gierige Durchgang vergleicht nur die
gewichtete Last. Zwei Personen können dieselbe gewichtete Last haben, während
eine davon die gesamte Kopfarbeit trägt und die andere nur Minuten. Das ist
exakt die Schieflage, gegen die diese App antritt.

## Entscheidung

Nach der Zuteilung läuft ein zweiter Durchgang, der Paare von Aufgaben
durchgeht und die zuständigen Personen tauscht, solange sich das Ergebnis
verbessert. **Nur echte Verbesserungen werden angenommen**, die Reihenfolge ist
fest — damit bleibt das Ergebnis deterministisch.

Verglichen wird zweistufig:

1. die Summe der quadrierten Abweichungen der **gewichteten Last** vom
   Mittelwert,
2. bei Gleichstand dasselbe für die **reine Kopflast**.

Das Quadrat statt der Spannweite, weil die Spannweite bei drei oder mehr
Personen blind für alles zwischen dem Höchst- und dem Tiefstwert ist — und
größere Haushalte sind ausdrücklich Teil des Produkts.

Vier Dinge sind dabei unantastbar:

- **Eignung.** Alters- und Rollenregeln gelten nach dem Tausch wie davor.
- **Kapazität.** Minuten je Tag und die Höchstzahl Aufgaben je Person und Tag.
- **Rotation über die Wochen.** Ein Tausch darf eine Aufgabe nie an die Person
  zurückgeben, die sie zuletzt hatte. Rotation geht vor Ausgleich — auch vor
  diesem Durchgang.
- **Rotation innerhalb der Woche.** Niemand sammelt durch einen Tausch einen
  zweiten Termin derselben Vorlage ein. Nachgetragen am 7. September, nachdem
  die erste Fassung genau das tat: Der Ausgleich holte das Kochen vom Mittwoch
  zu der Person, die schon Montag und Freitag kochte, weil die Zahlen dadurch
  aufgingen. Dreimal Abendessen bei einer Person ist die Beschwerde, gegen die
  diese App antritt — die Kennzahl war besser, das Ergebnis schlechter.
  Gefunden wurde es beim Lesen des Wochenplans, nicht von den Tests.
- **Feste Zuordnungen** bleiben, wo sie sind.

**Tage werden nicht verschoben.** Getauscht werden nur die Personen, die Tage
bleiben, wie der erste Durchgang sie gesetzt hat.

**Getauschte Aufgaben bekommen `Reason{Code: ausgleich}`.** Das ist keine
Kosmetik: Nach einem Tausch ist die ursprünglich gespeicherte Begründung nicht
mehr wahr, und eine Begründung, die nicht stimmt, ist schlimmer als keine.
Der Grund für die neue Zuordnung *ist* der Ausgleich.

## Konsequenzen

**Dafür:**

- Das Ergebnis ist messbar gleichmäßiger, und der Test kann es an konkreten
  Zahlen festmachen statt an einem Gefühl.
- Die Kopflast wird zum ersten Mal eigenständig ausgeglichen, nicht nur als
  Summand im gewichteten Wert.
- Der Durchgang ist vom ersten Schritt getrennt und einzeln testbar.

**Dagegen:**

- **Es ist eine lokale Suche, kein Optimum.** Ein Tausch, der erst über zwei
  Zwischenschritte besser wird, wird nicht gefunden. Für die Größenordnung
  eines Wochenplans ist das unerheblich, aber es ist so.
- **Die Rotationsregel begrenzt, was reparabel ist.** Die größten Schieflagen
  entstehen genau dadurch, dass Rotation vor Ausgleich geht — und die darf
  dieser Durchgang nicht anfassen. Er räumt auf, was danach noch übrig ist.
- **Quadratisch in der Zahl der Aufgaben je Durchlauf.** Bei zwanzig bis
  dreißig Aufgaben pro Woche irrelevant; bei einem Planer über ein Quartal
  wäre es eine andere Rechnung.
- Die Begründungen werden unschärfer: Wo vorher „Rotation, zuletzt bei Anna"
  stand, steht nach einem Tausch nur noch „Ausgleich".

## Verworfene Alternativen

**Nur die Spitzenlast senken** (klassisches Makespan-Problem). Einfacher und
bei zwei Erwachsenen gleichwertig. Verworfen, weil es bei drei oder vier
Personen ignoriert, wie sich alle unterhalb des Maximums verteilen — und ein
Haushalt mit zwei Jugendlichen ist erklärter Anwendungsfall.

**Kopflast strikt zuerst, Minuten danach.** Die radikalste Lesart der
Zwei-Achsen-Idee. Verworfen, weil ein Plan, in dem die Kopflast perfekt
aufgeht und eine Person dafür zwei Stunden mehr putzt, nicht als fair
empfunden wird — und Fairness ist hier kein mathematischer, sondern ein
wahrgenommener Zustand.

**Das echte Optimum ausrechnen** (Rückverfolgung oder ganzzahlige
Optimierung). Verworfen: Der Aufwand steht in keinem Verhältnis, die Laufzeit
wäre nicht mehr offensichtlich, und vor allem sind die Eingaben geschätzt —
Dauer und Kopflast jeder Vorlage sind gerundete Annahmen. Ein exaktes Optimum
über ungenauen Zahlen ist Scheingenauigkeit.

## Wann wir das revidieren

Wenn Haushalte zurückmelden, dass die Pläne zwar gleichmäßig, aber an den
falschen Tagen liegen. Dann muss der Durchgang auch Tage verschieben dürfen,
und das ist ein anderer Algorithmus — mit eigenem Eintrag.
