# 0011 — Maßstab statt Filter

**Status:** angenommen · 14. September 2026

## Kontext

Alle Bedingungen des Planers waren bis hierher **Filter**: Sie entschieden, ob
eine Aufgabe überhaupt existiert. Auto ja oder nein, Garten ja oder nein, Kind
im Kita-Alter ja oder nein.

Beim Nachsehen, was die Frage „Wohnung oder Haus?" eigentlich bewirkt, stellte
sich heraus: nichts. Keine einzige Vorlage hing daran. Sie stand als
Pflichtschritt im Onboarding, als Schalter in den Einstellungen und als Spalte
in der Datenbank — und veränderte am Plan nichts.

Die bessere Frage kam von Zakaria: *Wie groß ist die Wohnung?* Und sie ist
nicht nur besser, sie ist von anderer Art. Bis dahin bekam jeder Haushalt
dieselben Zahlen — Staubsaugen 25 Minuten in der Zweizimmerwohnung wie im
Fünfzimmerhaus. Das ist nicht bloß ungenau: Die Dauer trägt die
Fairnessrechnung. Wenn Anna staubsaugt und Ben einkauft, entscheidet genau
diese Zahl, wer mehr trägt.

## Entscheidung

**Es gibt zwei Arten von Eigenschaften, und sie gehören auseinandergehalten.**

**Filter** entscheiden, *ob* eine Aufgabe existiert: Auto, Garten, Haustierart,
Betreuungsform, die Fakten aus ADR-0010. Sie sind Ja/Nein.

**Maßstäbe** entscheiden, *wie groß* eine Aufgabe ist. Drei davon:

- `zimmer` — Fläche. Staubsaugen, Böden wischen, Staub wischen, Fenster putzen.
- `personen` — Kopfzahl. Betten beziehen, Wäsche, Bettwäsche. Vier Personen
  haben vier Betten, ob in drei oder fünf Zimmern; Zimmer sind ein guter
  Maßstab für Fläche und ein schlechter für Kopfzahl.
- `baeder` — **Anzahl, nicht Ausmaß.** Zwei Bäder sind nicht eine doppelt so
  lange Aufgabe, sondern zwei Aufgaben. Man putzt Bad A und Bad B, womöglich an
  verschiedenen Tagen und von verschiedenen Personen.

**Der Bezug ist drei Zimmer, ein Bad, drei Personen** — der Haushalt, für den
die Bibliothek kuratiert ist. Alle Dauern in `vorlagen.json` beziehen sich
darauf, und ein Haushalt dieser Größe bekommt exakt die Zahlen aus der Datei.

**Gedeckelt zwischen 0,6 und 2,0.** Ohne Deckel bekäme eine Einzimmerwohnung
acht Minuten Staubsaugen und ein Zehnzimmerhaus dreiundachtzig — beides Zahlen,
die niemand ernst nimmt, und der Plan verliert seine Glaubwürdigkeit an einer
Rechnung.

**Gerechnet wird in Promille, nicht mit Fließkomma.** Der Planer ist
deterministisch, und zwei Läufe mit derselben Eingabe sollen bis auf die Minute
dasselbe ergeben — auch auf einer anderen Maschine.

**Umgerechnet wird einmal, ganz vorne** (`ScaleTemplates` in `Plan()`). Sonst
rechnet die Zuteilung mit der einen Zahl, die Kapazitätsprüfung mit der anderen
und die Anzeige mit einer dritten. Die Seite „Eure Woche" benutzt dieselbe
Funktion.

**Und die Wohnform wird nicht mehr abgefragt.** Sie bleibt im Modell — sie ist
eine echte Eigenschaft eines Haushalts —, aber eine Frage, die nichts bewirkt,
gehört nicht in die ersten neunzig Sekunden. Sobald eine Vorlage daran hängt
(Streupflicht, Dachrinne, Schornsteinfeger), kommt sie zurück.

## Konsequenzen

**Positiv.** Die Dauern stimmen ungefähr statt gar nicht, und sie stimmen dort,
wo es zählt: in der Bilanz. Ein Haushalt der Bezugsgröße merkt keinen
Unterschied — die Änderung ist rückwärtskompatibel, weil die Vorgaben genau der
bisherige Zustand sind.

**Negativ.** Die Zahl in der Bibliothek ist nicht mehr die Zahl im Plan. Wer
`vorlagen.json` liest, sieht 25 Minuten und im Plan stehen 42; das ist gewollt
und trotzdem eine Stelle, an der man sich wundert. Und die Schätzungen bleiben
Schätzungen: Ein Faktor auf eine geratene Zahl ergibt eine genauer aussehende
geratene Zahl.

**Offen geblieben ist ein strukturelles Problem**, das dabei sichtbar wurde:
Der Ausgleich kann Aufgaben nur ganz verschieben, nicht teilen. Bei kleiner
Kapazität überschießt er deshalb systematisch — eine Aufgabe von 42 Minuten ist
für eine Dreizehnjährige 14 % ihrer Woche und für ihren Vater 5 %. In Familie B
ist damit ausgerechnet das jüngste Kind am höchsten ausgelastet. Das trifft
immer die Schwächsten im Haushalt: Kinder, Teilzeit, wer krank ist. Eine
Obergrenze je Einzelaufgabe für Minderjährige (45 Minuten) mildert das
Auffälligste, löst es aber nicht.

## Verworfene Alternativen

**Quadratmeter statt Zimmer.** Genauer und langsamer: „Drei-Zimmer-Wohnung"
weiß jeder auswendig, die Quadratmeter müssen die meisten nachsehen.

**Die Wohnform verdienen statt streichen** — also die fehlenden Hausaufgaben
ergänzen (Streupflicht, Dachrinne, Schornsteinfeger). Inhaltlich richtig und
trotzdem die falsche Reihenfolge: Erst Inhalt erfinden, um eine Frage zu
rechtfertigen, ist der Weg, auf dem Fragebögen wachsen. Die Vorlagen kommen,
wenn sie gebraucht werden, und dann kommt die Frage mit.

**Zwei Bäder als doppelte Dauer.** So war es zuerst gebaut, und es sah aus wie
Mathematik. Es war eine Behauptung darüber, wie Menschen putzen — und sie
erzeugte eine Einzelaufgabe von 70 Minuten, die bei einer Dreizehnjährigen
landete.

## Wann wir das revidieren

Wenn die Teilbarkeit von Aufgaben kommt („Bad putzen, 20 von 35 Minuten"), löst
sich das Granularitätsproblem von selbst, und die Obergrenze für Kinder wird
überflüssig.

Und sobald ein Haushalt seine Dauern selbst korrigieren kann, ist der Maßstab
nur noch der Startwert — so wie die Zeitstufen aus ADR-0007.
