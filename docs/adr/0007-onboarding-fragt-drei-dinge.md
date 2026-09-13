# 0007 — Das Onboarding fragt drei Dinge und rät den Rest

**Status:** angenommen · 13. September 2026

## Kontext

Bis heute entstand der Haushalt als Nebenwirkung: Der erste angemeldete GET auf
`/api/haushalte` legte ihn an, mit Vorgabewerten für alles. Das war bequem und
aus drei Gründen falsch. Ein Lesezugriff, der schreibt, ist für jeden
Zwischenspeicher eine Lüge. Der Name „Haushalt von Zakaria" ist niemandes
Haushalt. Und vor allem: Der Planer bekam keine einzige Angabe, an der er
arbeiten könnte — ohne Wohnform keine Unterscheidung zwischen Rasen und
Balkon, ohne Alter keine Kinderaufgaben, ohne Kapazität eine erfundene
Fairness.

Dagegen steht die härteste Regel aus dem Produktkonzept: **Eine App gegen
Mental Load darf beim Einrichten keine erzeugen.** Genau daran scheitern die
Mitbewerber — sie fragen Mitglieder, Arbeitszeiten, Betreuungszeiten und
Vorlieben ab, bevor irgendein Nutzen entstanden ist.

Die Frage lautet also nicht „viel oder wenig fragen", sondern: *Welche Angabe
verändert den ersten Plan so stark, dass ihr Fehlen auffällt?*

## Entscheidung

**Das Onboarding fragt drei Dinge und rät alles Übrige.**

1. **Wohnen** — Name des Haushalts, Wohnung oder Haus, Garten, Auto,
   Haustiere. Jedes Ja bringt Vorlagen mit, jedes Nein spart sie.
2. **Wer gehört dazu** — Name, erwachsen / Kind / wird betreut, bei den
   letzten beiden ein Geburtsjahr, bei Erwachsenen die Frage „plant mit?".
3. **Zeit** — je ausführender Person eine von drei Stufen: wenig, mittel, viel.

Nicht gefragt werden Arbeitszeiten, Betreuungszeiten und Vorlieben. Vorlieben
lernt die App aus dem Umverteilen; die Zeiten stecken grob schon in der
Zeitstufe.

**Kapazität in drei Stufen statt in Minuten.** Hinter den Stufen liegen
Minutenprofile für Werktag und Wochenende (`planner.TimeBudget`). Niemand weiß,
wie viele Minuten Haushalt er dienstags hat — eine erfundene Zahl sieht nur
präziser aus als eine ehrliche Stufe, tragen muss sie dieselbe Verteilung.

**Die Betreuungsform wird aus dem Alter geraten**, nicht abgefragt: unter drei
keine, ab drei Kita, ab sechs Schule.

**Alles in einem Aufruf und einer Transaktion.** `POST /api/haushalte` nimmt
Haushalt und Mitglieder zusammen entgegen. Ein Haushalt ohne Mitglieder wäre
nicht bloß unvollständig, sondern unerreichbar — der Zugriff läuft immer über
`member`, es gäbe niemanden mehr, der ihn je wieder sehen könnte.

**`Setup.Members[0]` ist die einrichtende Person** und muss planen. An ihr
hängt die Anmeldung; ein Haushalt ohne planende Person wäre von der ersten
Sekunde an verwaist.

## Konsequenzen

**Positiv.** Der erste Plan passt zum Haushalt statt zu einer Vorgabe. Der
schreibende GET verschwindet. Die Prüfung sitzt in `planner.Setup.Validate()`,
also an einer Stelle statt in Formular und Handler. Und die Fehlermeldungen
sind für Menschen geschrieben — sie landen unverändert im Formular.

**Negativ.** Drei Schritte sind drei Gelegenheiten abzubrechen; wir messen
nicht, an welcher. Die Zeitstufen sind geraten und tragen die gesamte
Verteilung — eine Person, die „mittel" wählt und tatsächlich zwanzig Minuten
hat, bekommt zu viel. Die geratene Betreuungsform liegt bei jedem Kind daneben,
das später eingeschult wird oder noch nicht in die Kita geht. Und es gibt
bisher keine Einstellungsseite: Was hier falsch gerät, lässt sich in der App
noch nicht korrigieren — das ist die nächste Schuld.

## Verworfene Alternativen

**Nur nach den Personen fragen** (unter 30 Sekunden). Der erste Plan hätte
Aufgaben enthalten, die es im Haushalt nicht gibt, und andere weggelassen.
Genau der erste Plan ist aber der Moment, in dem die App beweisen muss, dass
sie etwas weiß, das man selbst vergessen hätte.

**Fünf Schritte mit Arbeits- und Betreuungszeiten.** Ergäbe den besten ersten
Plan und verletzte die eine Regel, an der die Mitbewerber scheitern.

**Gar nicht nach Zeit fragen.** Spart einen halben Schritt. Dann wäre die
Fairness im ersten Plan reine Erfindung — und Fairness ist das, was das
Produkt verspricht.

**Eine Zahl statt drei Stufen.** Präziser in der Darstellung, falscher im
Inhalt: Die Zahl wäre geschätzt, sähe aber aus wie gemessen.

## Wann wir das revidieren

Wenn die Einstellungsseite steht, wird die Zeitstufe dort zu echten Minuten je
Tag — die Stufen bleiben dann nur noch das, was das Onboarding daraus macht.

Wenn wir sehen, dass Haushalte reihenweise in Schritt 2 abbrechen, ist die
Personenliste zu schwer und gehört auf „nur du, den Rest lädst du ein"
reduziert.

Und wenn die Betreuungsform sich als der häufigste Korrekturgrund erweist, ist
sie doch eine Frage wert — dann aber in Schritt 2, nicht als vierter Schritt.
