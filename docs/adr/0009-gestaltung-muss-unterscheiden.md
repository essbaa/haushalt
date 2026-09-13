# 0009 — Ein Token-System, und Gestaltung, die unterscheidet

**Status:** angenommen · 13. September 2026

## Kontext

Die Oberfläche entstand nebenbei: sieben Farbwerte in `globals.css`, alles
Übrige als Tailwind-Ketten in den Komponenten. Das reicht für eine Seite. Bei
sechs Seiten erfindet jede ihre eigenen Abstufungen, und das Ergebnis sieht aus
wie von drei Leuten an drei Tagen — was es auch war.

Zwei Fehler traten dabei zweimal auf, und beide sind derselbe:

1. **Jede Zeile sah gleich aus.** Gleicher Rahmen, gleiche Marken, dieselbe
   orange Kopflast an fast jeder Aufgabe. Wenn alles hervorgehoben ist, ist
   nichts hervorgehoben.
2. **Das Lauteste war erledigt.** Ein grün gefüllter Knopf auf der einen Zeile,
   die niemanden mehr interessiert.

Beide Male lenkte die Gestaltung Aufmerksamkeit dorthin, wo keine hingehört.

## Entscheidung

**Farben, Abstände und Radien stehen als Token in `globals.css`**, mit drei
Flächenstufen, zwei Linienstärken und einem hellen Ton für Ausgewähltes. Genau
diese Zwischenstufen fehlten; deshalb wirkte alles flach.

**Die Bausteine stehen einmal in `app/components/ui.tsx`** — Knopf, Karte,
Auswahl, Marke, Feld, Hinweis. Wer eine Form ändert, ändert sie überall.

**Gestaltung muss unterscheiden.** Fläche und Rahmen bekommen nur die eigenen,
offenen Aufgaben. Fremde Zeilen sind ruhig, Erledigtes verblasst. Die Kopflast
wird erst ab 2 farbig.

**Die Struktur ist die Woche.** Links Tag und Datum an einer Linie, heute
gefüllt, rechts die Aufgaben; vergangene Tage klappen zusammen. Ein Kasten je
Tag hätte nur optisch getrennt und nichts gesagt.

**Die Bilanz ist eine Gegenüberstellung**, kein Fortschrittsbalken je Person:
Das Versprechen lautet „wer halb so viel Zeit hat, trägt halb so viel" — das
sieht man nur im Vergleich auf einer Skala.

**Plus Jakarta Sans** über `next/font`, von der eigenen Adresse ausgeliefert.
Kein Aufruf zu Google im Browser des Nutzers.

**Jede anfassbare Fläche ist mindestens 44 Pixel hoch**, der Fokusring ist
überall sichtbar, und `prefers-reduced-motion` wird respektiert.

Verworfen wurden dabei bewusst vier Muster, die als „von einer KI gestaltet"
gelesen werden: Versalien als Beschriftung, Mittelpunkte als Trenner,
Dachzeilen über Überschriften, und dasselbe Kartenraster für alles.

## Konsequenzen

**Positiv.** Die Frage „was ist heute meins" beantwortet die Seite jetzt in der
ersten Bildschirmhöhe. Neue Ansichten greifen auf vorhandene Bausteine zu statt
neue Abstufungen zu erfinden.

**Negativ.** Die Token sind umbenannt (`foreground` → `fg`, `accent` →
`primary`), also musste jede vorhandene Klasse angefasst werden — ein
verpasster Name fällt nicht auf, weil Tailwind die Klasse einfach nicht
erzeugt und nichts bricht. Und ein eigenes Token-System ist Pflege: Es gibt
keine Bibliothek, die es für uns aktualisiert.

## Wann wir das revidieren

Wenn die App mehr als eine Handvoll Ansichten bekommt, lohnt sich eine
Komponentenbibliothek mit Zuständen und Tests statt einer Datei. Und sobald
jemand außer uns damit arbeitet, gehören die Token in eine Datei, die auch ein
Entwurfswerkzeug lesen kann.
