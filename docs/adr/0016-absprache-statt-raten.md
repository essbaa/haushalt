# 0016 — Wo die App den Kalender nicht kennt, spricht der Haushalt ab

**Status:** angenommen · 15. September 2026

## Kontext

Ein Kind geht in die Kita. Beide Eltern arbeiten. Einer bringt, einer holt ab
— oder sie klären es untereinander, je nach Arbeitsplan, und das Ergebnis ist
jede Woche ein anderes.

Der Planer verteilt nach **Kapazität**: wie viele Minuten jemand am Dienstag
übrig hat. Das ist die richtige Größe für „Wäsche aufhängen", die irgendwann
zwischen 6 und 22 Uhr passiert. Für „um 7:45 in der Kita sein" ist sie die
falsche. Dort entscheidet nicht, wer Zeit *hat*, sondern wer zu dieser Uhrzeit
an diesem Ort sein *kann* — und das steht in einem Dienstplan, den die App
nicht kennt und nach ADR-0007 auch nicht abfragen will (drei Fragen, nicht
dreißig).

Ohne eine eigene Behandlung würde die App hier trotzdem einteilen, mit
derselben Selbstsicherheit wie überall sonst. Ein Plan, der im Kern falsch
ist und richtig aussieht, ist schlimmer als eine Lücke: Er kostet jeden
Morgen eine Korrektur, und nach der dritten glaubt niemand mehr den Rest.

## Entscheidung

**Aufgaben können als „braucht Absprache" markiert sein. Diese verteilt der
Planer nicht — er trägt ein, was der Haushalt abgesprochen hat.**

Die Absprache ist ein Wochenraster: sieben Plätze, Index 0 ist Montag, je
Platz eine Person oder niemand (Tabelle `agreement`, Migration 00013). Sie
gilt, **bis jemand sie ändert** — „diese Woche wie letzte" ist damit der
Normalfall und braucht keinen Knopf. Wer noch nichts abgesprochen hat, sieht
die Aufgabe unter „Nicht im Plan" mit dem Grund `braucht_absprache` und einem
Weg dorthin.

Bewusst **keine neue Vorlagenart**. `braucht_absprache` ist ein Kreuzchen an
jeder Vorlage — auch an eigenen. Eine Art wäre eine Aussage darüber, *was*
die Aufgabe ist; das Kreuzchen sagt, was die App über sie *nicht weiß*, und
das ist die richtige Aussage. Zwei Haushalte können bei derselben Vorlage
verschieden antworten.

Die Zuteilung bekommt die Begründung `absprache` und zählt **voll in die
Bilanz**. Wer viermal die Woche bringt, hat vier Wege geleistet; sie aus der
Auslastung herauszurechnen, weil der Planer sie nicht verteilt hat, würde
genau die Arbeit unsichtbar machen, um die es in diesem Produkt geht.

Eine einzelne Ausnahme — Dienstag hat Amir einen Termin — ist das vorhandene
**Umverteilen** (ADR-0013) und betrifft nur den einen Tag. Das Raster bleibt.

## Nachtrag, 15. September 2026 — „ab sofort"

Beim ersten Ausprobieren fiel die Lücke auf: Ein Raster, am Dienstag gesetzt,
tauchte in der laufenden Woche nicht auf. Richtig so — die Woche steht fest
(ADR-0008) —, aber die Frage „und was ist mit Absprachen, die ab sofort gelten
sollen?" war damit unbeantwortet.

Der naheliegende Weg wäre das vorhandene Neurechnen gewesen, und er ist der
falsche: Es verwirft alles, woran nichts hängt, und verteilt die **ganze**
Woche neu. Wer am Mittwoch die Kita-Absprache einträgt, bekäme nebenbei eine
neue Antwort darauf, wer am Freitag das Bad putzt. Ein Plan, der sich an
Stellen ändert, die niemand angefasst hat, ist genau der Plan, dem niemand
mehr glaubt — und das wiegt schwerer als der gesparte Handgriff.

Also ein eigener, enger Aufruf: `POST …/absprachen/{vorlageId}/ab-heute`.
**Nur diese Vorlage, nur ab heute, nur wo noch nichts geschehen ist.**
Vergangene Tage bleiben unberührt — ein Kita-Termin von gestern früh ist keine
Verabredung mehr, die man noch treffen könnte. Zurück kommt die Zahl der
entstandenen Termine, damit die Oberfläche auch „es gab nichts einzutragen"
sagen kann statt „fertig" zu melden und nichts zu zeigen.

Damit hat die Absprache zwei Reichweiten, und beide stehen als Satz da: **ab
nächster Woche von selbst** — das ist der Normalfall und kostet keinen
Handgriff — oder **ab heute**, auf Knopfdruck, für diese eine Aufgabe.

## Konsequenzen

**Positiv.** Die App sagt zum ersten Mal „das weiß ich nicht" statt zu raten,
und zwar an der Stelle, an der Raten am teuersten ist. Das Raster ist
nebenbei die ehrlichste Form der Absprache, die viele Haushalte überhaupt
haben — sie steht sonst nirgends. Und da die Zuteilung voll zählt, korrigiert
sie die Verteilung aller übrigen Aufgaben in die richtige Richtung: Wer jeden
Morgen fährt, bekommt abends weniger.

**Negativ.** Es gibt jetzt einen Zustand, in dem eine geltende Aufgabe in
keinem Plan steht und keine Meldung erzwingt. Wer das Raster nie ausfüllt,
sieht die Kita-Aufgaben nie — die Lücke ist sichtbar, aber leise. Zweitens
ist das Raster starr: Wechselschicht im 14-Tage-Rhythmus lässt sich damit
nicht abbilden, nur wöchentlich neu eintragen. Drittens kostet es eine
Entscheidung beim Anlegen eigener Aufgaben, die man falsch treffen kann —
wer alles ankreuzt, hat die Verteilung abgeschaltet.

## Verworfene Alternativen

**Trotzdem verteilen und korrigieren lassen.** Der Planer teilt ein, der
Haushalt schiebt zurecht. Das ist das Verhalten, das wir kennen, und es ist
der Grund für die Entscheidung: Eine falsche Einteilung, die jede Woche
wiederkommt, ist keine Vorlage zum Korrigieren, sondern eine Aufgabe.

**Arbeitszeiten abfragen.** Die ehrlichste Lösung und die mit dem längsten
Fragebogen — Schichten, Gleitzeit, Homeoffice-Tage, Ausnahmen. Sie würde vor
den ersten Plan eine Stunde Eingabe stellen (ADR-0007) und wäre am ersten
Tag, an dem jemand früher raus muss, wieder falsch.

**Eine eigene Vorlagenart `absprache`.** Sauber im Modell und falsch in der
Sache: Ob eine Aufgabe abgesprochen werden muss, ist keine Eigenschaft der
Aufgabe, sondern eine des Haushalts. „Einkaufen" braucht bei uns keine
Absprache und bei jemandem mit einem Auto für zwei Erwachsene schon.

**Das Raster nur für diese Woche.** Wäre gleichlaufend mit dem Streichen
(ADR-0015) und im Alltag falsch: Eine Absprache ist genau das, was *nicht*
jede Woche neu getroffen wird. Der Aufwand, sie wöchentlich zu bestätigen,
wäre höher als der Nutzen — und ein leeres Raster am Montagmorgen ist der
Zustand, den dieses ADR verhindern soll.

## Wann wir das revidieren

Wenn die Probewoche zeigt, dass das Raster stillschweigend veraltet — also
Aufgaben nach Plan bei Personen stehen, die sie nicht gemacht haben. Dann
braucht es eine Rückfrage („stimmt das noch?"), ausgelöst von den Abhaken-
Daten, nicht vom Kalender.

Und wenn sich herausstellt, dass Haushalte das Kreuzchen bei eigenen Aufgaben
großzügig setzen, um der Verteilung auszuweichen. Dann ist `braucht_absprache`
zu einem bequemen Ausweg geworden, und die Frage muss anders gestellt werden.
