# 0008 — Die Woche wird beim ersten Ansehen festgeschrieben

**Status:** angenommen · 13. September 2026 · präzisiert am 14. September
(siehe Nachtrag am Ende)

## Kontext

Bis hierher war ein Wochenplan eine Rechnung. `Plan()` las Vorlagen und
Verlauf, ließ den Planer laufen und gab das Ergebnis zurück; gespeichert wurde
nichts. Für eine Ansicht war das genau richtig — der Planer ist
deterministisch, zwei Läufe mit derselben Eingabe ergeben dasselbe.

Zwei Dinge brechen diese Annahme.

**Die Eingabe ist nicht stabil.** Deterministisch heißt „gleiche Eingabe,
gleiches Ergebnis" — und die Eingabe wächst. Der Verlauf bekommt Ereignisse,
Gleichstände brechen nach Mitglieds-Kennung, und der heutige Tag ist morgen ein
anderer. Ein Plan, der sich beim Neuladen verschiebt, ist kein Plan. Wer am
Montag gelesen hat, dass er den Müll rausbringt, muss das am Mittwoch noch so
vorfinden.

**Eine Aufgabe ohne Zeile lässt sich nicht anfassen.** Abhaken, abgeben,
umverteilen — alles braucht etwas, woran es hängen kann. Eine gerechnete
Aufgabe hat keine Kennung.

## Entscheidung

**Sieht ein Mitglied eine Woche zum ersten Mal an, wird sie geschrieben.**
`task_instance` und `assignment` bekommen ihre Zeilen, `week_plan` die Marke
„diese Woche steht" samt der übersprungenen Vorlagen. Jeder weitere Aufruf
liest, statt zu rechnen.

**Nur für Mitglieder.** Die öffentlichen Beispielhaushalte bleiben eine
Rechnung: Ein Besucher soll keine Zeilen erzeugen, und die Demo soll mit der
Bibliothek mitwachsen statt im Januar einzufrieren.

**Die Marke ist eine eigene Tabelle, kein Zählen von Aufgaben.** Eine Woche
ohne Aufgaben wäre sonst nicht von einer ungeschriebenen zu unterscheiden, und
jeder Aufruf würde sie erneut zu schreiben versuchen. Der zusammengesetzte
Primärschlüssel `(household_id, iso_week)` ist zugleich die Absicherung gegen
zwei gleichzeitige erste Aufrufe: Der zweite läuft in den Konflikt, schreibt
nichts und liest, was der erste hinterlassen hat.

**Die Bilanz wird weiter gerechnet, nie gespeichert.** Sie ist eine Sicht auf
die Aufgaben; eine gespeicherte Summe neben den Posten läuft auseinander.
Dasselbe gilt für Titel, Kategorie und Art: Die stehen in der Vorlage und
werden beim Lesen nachgeschlagen, damit ein korrigierter Titel überall
ankommt.

## Konsequenzen

**Positiv.** Der Plan steht still. Aufgaben haben Kennungen, also kann man sie
abhaken und abgeben. Die bekannte Unschönheit „Demo-Pläne verschieben sich nach
`db-reset`" betrifft nur noch die Demo. Und das Ereignisprotokoll bekommt
endlich Aufgaben, auf die es zeigen kann.

**Negativ, und das wiegt.** Ein GET schreibt — genau das, was ADR-0007 beim
Onboarding abgeschafft hat. Der Unterschied ist schmal: Dort entstand eine
sichtbare Sache (ein Haushalt) als Nebenwirkung eines Lesezugriffs, hier wird
das Ergebnis einer deterministischen Rechnung materialisiert, und die Antwort
ist für Angemeldete ohnehin `no-store`. Schmal, aber vorhanden. Wir schreiben
es hierhin, statt es wegzureden.

Weiter: Ändert sich eine Vorlage oder zieht jemand ein, merkt die laufende
Woche davon nichts mehr — es gibt noch keinen Weg, eine Woche neu rechnen zu
lassen. `DeleteWeek` liegt bereit, ein Knopf dafür fehlt.

Und eine Woche, die ein Mitglied nie öffnet, wird nie geschrieben. Für
Auswertungen über die Zeit heißt das: Es gibt Lücken, und die sind keine
Aussage über den Haushalt, sondern über sein Klickverhalten.

## Verworfene Alternativen

**Weiter rechnen, „erledigt" an (Vorlage, Datum) hängen.** Am billigsten, ohne
Migration. Damit lässt sich aber keine Umverteilung von Hand darstellen — und
genau das ist T8. Außerdem bliebe der Plan wackelig.

**Beim Bestätigen durch eine planende Person festschreiben.** Ehrlicher in der
Sprache: Niemand hat zugestimmt, bis jemand zustimmt. Dafür eine Pflicht mehr
pro Woche, und im Solo-Betrieb eine Hürde ohne Gegenwert. Verworfen, aber nicht
für immer — siehe unten.

**Jede Woche im Voraus schreiben, etwa per Hintergrundlauf.** Klingt sauber und
erzeugt Zeilen für Wochen, die niemand ansieht, samt Zuteilungen, über die nie
jemand gesprochen hat.

## Wann wir das revidieren

Wenn Haushalte anfangen, Pläne gemeinsam zu besprechen, wird aus dem stillen
Festschreiben ein „Woche übernehmen" mit Datum und Person — dann ist die Marke
in `week_plan` schon da und bekommt nur zwei Spalten mehr.

Und sobald jemand mitten in der Woche einzieht oder eine Vorlage ändert,
brauchen wir „neu rechnen, aber Erledigtes behalten". Das ist kein neues
Konzept, sondern eine Abfrage: `DeleteWeek`, neu schreiben, die Ereignisse
stehen lassen.

## Nachtrag, 14. September 2026 — nur die laufende Woche

Der Titel sagt „beim ersten Ansehen". Solange es keinen Weg zu einer anderen
Woche gab, war das dasselbe wie „die laufende Woche" — und der Unterschied
fiel niemandem auf.

Mit Blätterpfeilen im Wochenplan fällt er auf, und zwar in beide Richtungen:

**Nach vorn.** Der erste Klick auf „nächste Woche" hätte sie festgeschrieben.
Ein danach eingetragener Anlass, eine abgeschaltete Vorlage, eine geänderte
Kapazität — nichts davon wäre dort je angekommen. Das Versprechen dieses
Eintrags lautet: Der Plan ändert sich nicht unter dir. Es gilt der Woche, in
der jemand lebt. Ein Blick nach vorn verspricht nichts, und ihn zu einer
Zusage zu machen ist kein Dienst, sondern ein Nebeneffekt des Hinsehens.

**Nach hinten.** Eine vergangene Woche, die nie geschrieben wurde, hat niemand
gesehen. Sie beim Zurückblättern zu schreiben hätte Vergangenheit erfunden:
Zuteilungen, die nie jemand hatte, und eine Rotation, die daraus weiterrechnet.

Festgeschrieben wird deshalb nur die **laufende** Woche, in der Zeitzone des
Haushalts. Bereits geschriebene Wochen bleiben, wie sie sind — auch wenn sie
vorbei sind. Die Oberfläche sagt es dazu: Eine künftige Woche trägt ein
Banner „Vorschau — diese Woche steht noch nicht fest", samt der Folge, dass
dort nichts abzuhaken ist.

Das ist keine Kursumkehr, sondern eine Präzisierung. Bemerkenswert ist, wie
sie zustande kam: **Eine neue Funktion hat eine alte Zweideutigkeit sichtbar
gemacht.** Die Regel war nie eindeutig; es gab nur keinen Weg, den Unterschied
zu bemerken.
