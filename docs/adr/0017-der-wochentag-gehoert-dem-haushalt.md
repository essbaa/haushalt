# 0017 — Der Wochentag gehört dem Haushalt, nicht der Bibliothek

**Status:** angenommen · 20. September 2026

## Kontext

Die Bibliothek kennt seit dem ersten Tag feste Wochentage. Neun Vorlagen
tragen sie:

```
t-muell         Müll rausbringen        di
t-abendessen    Abendessen kochen       mo, mi, fr
t-kita-tasche   Kita-Tasche packen      so
t-hausaufgaben  Hausaufgaben begleiten  mo, di, mi, do
```

Diese Tage sind **geraten**. Nicht schlampig geraten — irgendein Tag musste
dastehen, und Dienstag ist für Müll so gut wie jeder andere. Aber der Müll
kommt nicht in jeder Straße dienstags, und Abendessen kocht nicht jede Familie
montags, mittwochs, freitags.

Dagegen gab es bisher genau ein Mittel: die Vorlage abbestellen und von Hand
neu anlegen. Damit verliert der Haushalt die Zahlen der Bibliothek — Dauer,
Kopflast, Skalierung, Bedingungen — und bekommt eine Aufgabe, die als
„selbst angelegt, geschätzt" gekennzeichnet ist und die Fairnessrechnung mit
geschätzten Werten trägt. Eine Vorlage wegzuwerfen, weil ein einziges Feld
nicht stimmt, ist kein Rückweg, sondern eine Strafe.

Dasselbe von der anderen Seite: Eine selbst angelegte Aufgabe kannte nur
„alle N Tage". Die gelbe Tonne, die donnerstags rausmuss, landete damit
irgendwo in der Woche — und eine Aufgabe, die am falschen Tag steht, ist
nicht einmal falsch, sondern jede Woche.

**Der Planer konnte das die ganze Zeit.** `RhythmFixed` mit `Weekdays`
existiert, `dueFixed` setzt die Aufgaben exakt auf ihre Tage. Was fehlte, war
die Eingabe. Die Lücke lag nicht im Modell, sondern zwischen Modell und
Mensch — und das ist die Sorte Lücke, die man nicht bemerkt, solange man
seine eigene Bibliothek benutzt.

## Entscheidung

Ein Haushalt darf die Wochentage jeder Vorlage überschreiben — der
kuratierten wie der eigenen. Gespeichert wird in `template_weekday`, eine
Zeile je Vorlage und Tag, Index 0 = Montag, dieselbe Form wie beim
Wochenraster der Absprache.

Drei Festlegungen dazu:

**1. Feste Tage schlagen jeden anderen Rhythmus.** Wer „donnerstags" sagt,
meint nicht „alle sieben Tage, bevorzugt donnerstags". Also wird aus der
Vorlage ein `fest`-Rhythmus, egal was vorher dastand.

Das ändert die **Häufigkeit** mit: Aus „alle 14 Tage" wird durch die Wahl
eines Tages „jeden Samstag". Das ist keine Nebenwirkung, die man wegdiskutiert
— es ist die Aussage. Verschwiegen wäre es der schlimmste Fall: ein Plan, der
doppelt so voll ist wie der, den jemand gesetzt hat, und eine Bilanz, die das
mitträgt, ohne dass es jemand erklären kann. Deshalb steht der Satz im
Bildschirm, während man wählt, und nicht in einer Hilfe.

**2. Keine Tage heißt: Vorgabe.** Es gibt keinen dritten Zustand und kein
Kennzeichen „zurückgesetzt". Kein Haken = kein Eintrag in der Tabelle = der
Rhythmus der Bibliothek gilt. Dieselbe Regel wie beim Abbestellen: „nie etwas
gesagt" und „wieder erlaubt" wirken gleich. Ein eigener Zurücksetzen-Knopf
wäre ein zweiter Weg zu demselben Zustand, und zwei Wege zu einem Zustand sind
die Gelegenheit, sie auseinanderlaufen zu lassen.

**3. Nicht für jeden Rhythmus.** Erlaubt bei `fest` und `fenster`, beides
Rhythmen, für die „an welchem Tag?" überhaupt eine Antwort ist. Abgelehnt bei:

- `ausloeser` — die Wäsche kommt, wenn der Korb voll ist. Ein Wochentag ist
  dort keine Antwort, sondern eine andere Frage, und sie zu beantworten hieße,
  den Auslöser abzuschaffen.
- `saison` und `phase` — die laufen über Monate und Jahre.
- alles mit `RequiresEvent` — das hat seinen Tag schon, vom Anlass.

Die Grenze steht im Dienst und nicht nur im Bildschirm. Eine Regel, die nur
die Oberfläche kennt, gilt für den nächsten Aufrufer nicht. Der Bildschirm
bietet die Wahl dort trotzdem gar nicht erst an (`wochentage_moeglich`) — eine
Oberfläche, die eine `400` provoziert, ist eine Einladung in einen Fehler.

Geprüft wird gegen die **Bibliotheksfassung**, nicht gegen die, die der
Haushalt gerade sieht. Sonst wäre die Prüfung nach dem ersten Festlegen
wirkungslos: Eine überschriebene Vorlage liest sich als `fest`, und `fest`
wäre dann immer erlaubt. Dafür gibt es `templatesRoh` neben `templates` — die
einzige Stelle im Dienst, die die rohe Fassung braucht.

## Folgen

Die Bibliothek darf ungenau sein. Das ist mehr als eine Bequemlichkeit: Solange
jeder geratene Tag ein Fehler war, den nur ein Neuanlegen behebt, musste die
Bibliothek Tage vermeiden — und ohne Tage ist „Abendessen kochen" eine
Aufgabe, die irgendwann in der Woche fünfmal auftaucht. Jetzt kann sie einen
plausiblen Tag vorschlagen und der Haushalt korrigiert ihn.

Was **nicht** entschieden ist: das Verschieben einer einzelnen geplanten
Aufgabe („heute nicht, Mittwoch"). Das ist eine andere Frage, und sie ist
größer, als sie klingt — innerhalb des Fensters ist es eine Korrektur der
Zuteilung, darüber hinaus hebelt es den Rhythmus aus, und dann braucht es eine
Antwort darauf, was mit der nächsten Fälligkeit geschieht und ob die Aufgabe
in der Bilanz derselben Woche bleibt. Die Probewoche beantwortet zuerst, wie
oft es überhaupt vorkommt (M7). Siehe offene Frage 5.
