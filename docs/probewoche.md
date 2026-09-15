# Die Probewoche

Anleitung und Prüfplan für die erste Woche mit einem echten Haushalt.

Zwei Teile in einer Datei, weil sie zusammengehören: **Wie die App benutzt
wird** — das können alle lesen — und **was wir dabei herausfinden wollen** —
das ist für dich. Wer nur mitmacht, liest Abschnitt 1 und 4 und hört auf.

---

## 0. Wozu die Woche da ist

Sie beantwortet drei Fragen, und nur diese drei:

1. **Trägt das Modell?** Zwei Achsen, Fairness nach Kapazität, Rotation vor
   Ausgleich — das ist bisher eine Behauptung, geprüft an erfundenen
   Haushalten. Ein echter Haushalt sagt in sieben Tagen, ob die Verteilung
   sich richtig *anfühlt*. Das ist die Frage, die kein Test beantworten kann.

2. **Wird die App benutzt, wenn niemand sie benutzen muss?** Nicht von dir —
   du hast sie gebaut. Von den anderen.

3. **Was fehlt so sehr, dass es weh tut?** Nicht „wäre schön", sondern: Was
   führt dazu, dass jemand die App am Mittwoch zuklappt und nicht wieder
   aufmacht.

**Was sie nicht beantwortet:** ob die App fertig ist. Sie ist es nicht, und
das ist in Ordnung. Eine Probewoche mit einer fertigen App wäre keine Probe,
sondern eine Vorführung.

**Die wichtigste Regel der Woche: In diesen sieben Tagen wird nichts
gebaut.** Kein Fehler wird währenddessen behoben, keine Vorlage nachgeschoben,
keine Zahl nachjustiert. Wer an einem Ding schraubt, während er es misst,
misst das Schrauben. Alles wird notiert und am Freitagabend ausgewertet.
Einzige Ausnahme: Die App ist unbenutzbar (siehe Abschnitt 8).

---

## 1. Wie die App denkt — zehn Sätze

Das ist der Teil, den alle lesen sollten. Wer versteht, *warum* etwas so
dasteht, ärgert sich weniger und korrigiert öfter.

1. **Aufgaben haben zwei Größen: Dauer und Kopflast.** Dauer ist die Zeit,
   Kopflast der Aufwand, überhaupt daran zu denken. „Zahnarzttermin für alle
   machen" dauert acht Minuten und liegt trotzdem schwerer auf einem Menschen
   als vierzig Minuten Staubsaugen. Die App rechnet beides.

2. **Verteilt wird nach Kapazität, nicht nach Kopfzahl.** Wer zehn Stunden in
   der Woche hat, bekommt nicht dasselbe wie jemand mit dreißig. Die Bilanz
   zeigt deshalb **Prozent der eigenen Zeit** und nicht Minuten.

3. **Rotation geht vor Ausgleich.** Wer eine Aufgabe zuletzt hatte, bekommt
   sie nicht sofort wieder — außer die Woche bliebe sonst deutlich schief.
   Dann sticht die Fairness.

4. **Was die App nicht weiß, plant sie nicht — sie fragt.** Oben im Plan
   stehen höchstens zwei Fragen. Jede beantwortete Frage schaltet echte
   Aufgaben frei. Unbeantwortet heißt *nicht* „nein".

5. **Alles Geratene ist korrigierbar.** Beim Einrichten hat die App Kapazität,
   Betreuungsform und Wohnform geschätzt. Alles steht in den Einstellungen und
   lässt sich ändern.

6. **Was ihr absprecht, verteilt sie nicht.** Kita bringen und abholen hängt
   am Arbeitsplan, nicht an freier Zeit. Dafür gibt es ein Wochenraster, das
   ihr selbst setzt. Es gilt, bis jemand es ändert.

7. **Die laufende Woche steht fest.** Einmal angesehen, ändert sie sich nicht
   mehr von selbst — sonst wäre es kein Plan. Änderungen wirken ab nächster
   Woche, oder auf Knopfdruck sofort.

8. **„Diesmal nicht" und „Brauchen wir nicht" sind zwei verschiedene Dinge.**
   Das erste streicht einen Termin, das zweite schaltet die Aufgabe für immer
   ab. Beide sind umkehrbar, aber nicht dasselbe.

9. **Abgeben ist erlaubt und kein Versagen.** Die App sucht dann jemand
   anderen. Findet sie niemanden, steht die Aufgabe offen da — sichtbar, nicht
   verschwunden.

10. **Die App erinnert an nichts.** Noch nicht. Wer nicht hineinsieht, weiß
    nicht, was ansteht. **Genau das ist eine der Messungen dieser Woche.**

---

## 2. Vor dem Start

### 2a. Technisch (nur du, am Wochenende davor)

| # | Was | Warum jetzt |
|---|---|---|
| 1 | Den Fehler mit den wiederkehrenden Fragen zu Ende diagnostizieren | Eine Frage, die nach dem Neuladen wieder dasteht, macht Messung 4 wertlos |
| 2 | Alles committen, `make check` grün, deployen | Der erste echte Lauf des `release_command` mit den Migrationen 00012 und 00013 |
| 3 | Domain kaufen, Resend verifizieren, dann `requireEmailVerification` umlegen | Ohne das kommt keine Bestätigungsmail an, und der Einladungsweg endet im Nichts |
| 4 | **Einladungen per Mail verschicken** — sobald Resend steht | Bis dahin ist eine Einladung ein Code, den du abschreibst und per Nachricht schickst. Das geht für zwei Personen und für niemanden sonst. Mit Resend wird daraus eine Mail mit Link: Adresse eintragen, abschicken, fertig — und der Beitritt beginnt dort, wo er ankommt, statt bei einer Zeichenkette in der Zwischenablage |
| 5 | Den Einladungsweg einmal vollständig gehen — mit einem echten zweiten Menschen und einem echten zweiten Telefon | Der einzige Weg, der die App für alle anderen öffnet. Bricht er, ist die Woche vorbei, bevor sie anfängt |

**Punkt 5 zuerst ausprobieren, nicht zuletzt.** Wenn das Einladen nicht
funktioniert, ist alles andere egal — und es ist der Weg mit den meisten
beweglichen Teilen: Mailversand, Code, Anmeldung, Zuordnung zum Haushalt.

### 2b. Den Haushalt einrichten (zusammen, 20 Minuten)

Setzt euch dafür einmal zusammen hin. Nicht, weil es lange dauert, sondern
weil die Zahlen sonst eine Person allein für alle anderen schätzt — und genau
das ist der Streit, den die App verhindern soll.

**Der Reihe nach:**

1. **Mitglieder anlegen.** Für jede Person: Name, Rolle, Alter.
   - *planend* — plant und führt aus. Sieht die Bilanz. In der Regel die
     Erwachsenen.
   - *ausführend* — führt aus, plant nicht. Für Jugendliche, die eigene
     Aufgaben übernehmen (ab etwa 12 sinnvoll; ein Siebzehnjähriger gehört
     hierher, nicht zu den Betreuten).
   - *betreut* — erzeugt Arbeit, übernimmt keine. Kleine Kinder.

2. **Zeit je Person prüfen.** Die App hat aus einer Stufe (keine / wenig /
   mittel / viel) Minuten je Wochentag geraten. **Diese Zahlen tragen die
   ganze Verteilung.** Wer hier großzügig schätzt, bekommt zu viel; wer knapp
   schätzt, zu wenig. Ehrlich schätzen heißt: Zeit, die realistisch für
   Haushalt übrig ist, nachdem Arbeit, Wege und Schlaf abgezogen sind. Nicht
   die Zeit, die man gerne hätte.

3. **Die Fragen oben im Plan beantworten.** Es kommen immer zwei. Beantwortet
   sie, ladet neu, beantwortet die nächsten zwei. So lange, bis keine mehr
   kommt — oder bis ihr genug habt; der Rest kommt nächste Woche wieder.

4. **Die Aufgabenliste einmal durchgehen.** Bereich für Bereich aufklappen.
   Was bei euch nicht vorkommt: **Abschalten**. Was fehlt: als **eigene
   Aufgabe** anlegen (das „+" im jeweiligen Bereich).

   Beim Anlegen einer eigenen Aufgabe kommen drei Fragen, die zählen: *Muss
   jemand daran denken?* (Kopflast) · *Muss man erst etwas klären?* (Kopflast)
   · *Müssen wir das absprechen?* (dann verteilt die App sie nicht).

5. **Absprachen setzen.** In der Aufgabenliste haben „Zur Kita bringen" und
   „Von der Kita abholen" ein Wochenraster. Tragt ein, wer wann. Leere Tage
   sind erlaubt (Wochenende, Urlaub). Wenn ihr es für die laufende Woche
   braucht: „Ab heute eintragen".

6. **Anlässe eintragen.** Geburtstage, Elternabend, Zahnarzt. Daraus entstehen
   Aufgaben mit Vorlauf — ein Geschenk kauft sich nicht am Vortag.

7. **Auf jedes Telefon legen.** Safari öffnen, Teilen-Symbol, „Zum
   Home-Bildschirm". Die App startet dann ohne Browserleiste und sieht aus wie
   eine App. Wer sie über ein Lesezeichen aufruft, benutzt sie nicht.

**Am Sonntagabend fertig sein.** Der Plan schreibt sich beim ersten Ansehen am
Montag fest; wer am Dienstag noch einrichtet, misst eine halbe Woche.

---

## 3. Die Woche — was jeder tut

**Für alle: drei Gesten, mehr nicht.**

| Was | Wie |
|---|---|
| **Erledigt** | Kreis rechts antippen — oder die Zeile nach rechts wischen |
| **Abgeben** | Pfeil-Zeichen neben dem Kreis, Grund eintippen (oder leer lassen), „Zurückgeben" |
| **Diesmal nicht** | Zeile nach links wischen — oder antippen und „Diesmal nicht" |

Dazu, seltener: **„Wer macht das?"** verteilt eine Aufgabe gezielt an eine
Person (sticht alles, was die App sich gedacht hat). **„Brauchen wir nicht"**
schaltet eine Aufgabe für immer ab.

**Einmal am Tag hineinsehen.** Am besten morgens beim Kaffee. Die App erinnert
an nichts — das ist bekannt und wird gemessen, nicht ausgeglichen. Wer
hineinsieht, weil er es sich vorgenommen hat, ist ein gültiger Messwert. Wer
eine Erinnerung von außen bekommt, ist keiner.

**Für dich zusätzlich: nichts reparieren.** Wenn etwas falsch aussieht, ist
das ein Ergebnis. Aufschreiben (Abschnitt 5), weitermachen.

---

## 4. Was wir messen

Sieben Behauptungen. Zu jeder steht, **woran** man sie merkt und **was es
heißt**, wenn sie fällt. Eine Behauptung, zu der man das nicht sagen kann,
gehört nicht in die Liste.

### M1 — Die Verteilung fühlt sich fair an

*Woran:* Am Sonntag die Bilanz ansehen. Liegen alle Erwachsenen innerhalb von
etwa zehn Prozentpunkten? **Und, wichtiger:** Fragt euch gegenseitig, *ohne
vorher auf die Zahl zu schauen*, wer diese Woche mehr getragen hat. Stimmen
Gefühl und Zahl überein?

*Wenn nicht:* Die Kapazitätszahlen stimmen nicht (häufigster Fall), oder die
Kopflast einzelner Aufgaben ist zu niedrig angesetzt, oder `MaxOvershootPermille`
(25 ‰) ist zu großzügig.

### M2 — Kopflast ist mehr als eine Zahl im Modell

*Woran:* Zählt am Ende, wie viele Aufgaben mit Kopflast ≥ 2 bei wem lagen. Und
fragt: Hat die Person, die sonst „an alles denkt", diese Woche weniger daran
gedacht?

*Wenn nicht:* Das ist der wichtigste Einzelwert der Woche. Fällt M2, ist das
Alleinstellungsmerkmal der App eine Rechnung ohne Wirkung — dann ist die Frage
nicht „welche Zahl stimmt nicht", sondern „reicht Verteilen überhaupt, oder
braucht es Erinnern".

### M3 — Dem Plan wird geglaubt

*Woran:* Wie oft wurde „Wer macht das?" benutzt? Jede Umverteilung von Hand
steht als `manual` in der Datenbank. Am Sonntag zählen.

```sql
SELECT reason_code, count(*), sum(manual::int) AS von_hand
FROM assignment GROUP BY reason_code ORDER BY 2 DESC;
```

*Deutung:* Nie umverteilt heißt nicht „perfekt" — es kann auch heißen, dass
niemand hingesehen hat. Viel umverteilt, immer in dieselbe Richtung (immer weg
von derselben Person), heißt: Der Planer liegt systematisch daneben, und die
Korrektur ist die Arbeit, die er abnehmen sollte.

### M4 — Der Einstieg trägt

*Woran:* Wie viele Fragen wurden beantwortet, wie viele Aufgaben abgeschaltet,
wie viele eigene angelegt? Und: Kam eine beantwortete Frage je wieder?

*Wenn schlecht:* Wenn nach einer Woche noch zwanzig Aufgaben ungeklärt sind,
ist der Einstieg zu langsam. Wenn zwanzig abgeschaltet wurden, ist die
Bibliothek für diesen Haushalt zu voll.

### M5 — Die Absprache ist die richtige Antwort

*Woran:* Wurden die Kita-Termine abgehakt? Wie oft musste das Raster geändert
werden? Wie oft wurde ein einzelner Tag umverteilt?

*Wenn schlecht:* Ändert sich das Raster mehrmals pro Woche, ist ein festes
Wochenraster die falsche Form — dann braucht es etwas Wöchentliches
(„wer kann diese Woche?"). Wurden die Termine nie abgehakt, sind sie im Plan
überflüssig, und die Absprache ist nur eine Notiz.

### M6 — Die Erinnerungslücke

*Woran:* Zählt mit, an wie vielen Tagen jeder die App überhaupt geöffnet hat.
Ein Strich pro Tag auf dem Zettel reicht.

*Deutung:* **Das ist die Messung, die über das nächste große Stück
entscheidet.** Öffnet ihr sie täglich von selbst, sind Benachrichtigungen
Komfort. Öffnet sie außer dir niemand ohne Aufforderung, sind sie das nächste
Feature — und alles andere kann warten.

### M7 — Was fehlt

*Woran:* Die Liste in den Einstellungen, Art *Idee* und *Stört mich*. Dazu,
was jemand laut gesagt und nicht gemeldet hat — dafür ist die letzte Zeile auf
dem Zettel da. Wortwörtlich, nicht übersetzt.

*Deutung:* Das ist die einzige Messung ohne Zielwert. Drei Sätze, die
zweimal fallen, sind die Roadmap der nächsten zwei Wochen.

---

## 5. Melden in der App — und der Zettel

**In der App** steht unten auf jeder Seite „Stimmt etwas nicht?". Drei Arten:
*Stimmt nicht* · *Stört mich* · *Idee*. Ein Satz reicht; wo man gerade war,
schickt die App von selbst mit. **Das darf jeder im Haushalt benutzen, nicht
nur du** — die ausführenden Personen sehen die Stellen, an denen es klemmt,
zuerst. Gemeldetes steht in den Einstellungen unter „Was gemeldet wurde".

Das ist der bessere Weg für alles, was **in dem Moment** auffällt, in dem es
auffällt — und der Wortlaut ist bei einer Rückmeldung das Ganze. Abends aus
dem Gedächtnis aufgeschrieben ist er ein anderer.

**Der Zettel bleibt trotzdem**, weil er etwas kann, was das Formular nicht
kann: Er hält fest, was **nicht** passiert ist. „Heute hat außer mir niemand
die App geöffnet" meldet niemand — und genau das ist Messung M6.

Zwei Minuten am Abend, und er ist kürzer geworden:

```
Tag ____________

App geöffnet:      ich [ ]   Partnerin [ ]   Sohn [ ]
Abgehakt:          ____ von ____
Etwas umverteilt?  ja / nein
Etwas gestrichen?  ja / nein

Was jemand über die App gesagt und NICHT gemeldet hat:
_________________________________________________
```

Die letzte Zeile ist die wichtigste am Zettel. Was jemand laut sagt und nicht
tippt, ist der Satz, der sonst verlorengeht — und oft der ehrlichste.

Am Sonntag den Zettel und die Liste aus den Einstellungen zusammen in
`docs/probewoche-protokoll.md` übertragen.

---

## 6. Freitagabend: die Auswertung

Zwanzig Minuten, gemeinsam. Sechs Fragen, in dieser Reihenfolge — die erste
ist die wichtigste und muss gestellt werden, bevor jemand die Zahlen gesehen
hat:

1. **Wer hat diese Woche mehr getragen?** (Erst danach die Bilanz zeigen.)
2. Was hat die App dir abgenommen?
3. Was hat sie dir *zusätzlich* aufgehalst?
4. Wann hast du sie zugemacht, ohne etwas zu tun — und warum?
5. Was hat dich geärgert?
6. Würdest du nächste Woche weitermachen, wenn ich nicht fragen würde?

**Frage 6 ist das Ergebnis der Woche.** Alles andere sind Details dazu.

Und eine Frage an dich selbst, ehrlich beantwortet: *Habe ich diese Woche
Dinge im Kopf erledigt, die eigentlich die App hätte tun sollen?* Wer sein
eigenes Produkt heimlich umgeht, hat ein Ergebnis — nur kein gutes.

---

## 7. Was diese Woche nicht geprüft wird

Damit niemand darauf wartet und niemand es als Fehler meldet:

- **Erinnerungen und Benachrichtigungen** — gibt es nicht, mit Absicht (M6).
- **Einkaufsliste, Essensplanung** — nicht gebaut.
- **Punkte, Belohnungen für Kinder** — nicht gebaut.
- **Mehrere Haushalte gleichzeitig** — geht, ist aber nicht das Ziel.
- **Offline** — die App braucht Netz.
- **Fremde Menschen** — der Einladungsweg wird mit euch geprüft, nicht mit
  Unbekannten. Der Fall „einzige planende Person ausgesperrt" bleibt offen.

---

## 8. Wenn etwas kaputtgeht

**Erste Regel: aufschreiben, nicht reparieren.** Zweite Regel: Der Haushalt
läuft weiter, auch wenn die App nicht läuft. Niemand soll eine Aufgabe
liegenlassen, weil die App klemmt.

**Was du ohne Codeänderung tun kannst:**

| Problem | Weg |
|---|---|
| Eine Aufgabe steht bei der falschen Person | „Wer macht das?" |
| Eine Aufgabe ist diese Woche unsinnig | „Diesmal nicht" |
| Eine Aufgabe ist immer unsinnig | „Brauchen wir nicht" |
| Aus Versehen abgeschaltet | Aufgabenliste → Filter „Abgeschaltet" → Anschalten |
| Der halbe Plan ist falsch, weil eine Einstellung nicht stimmte | Einstellungen ändern, dann „Schon diese Woche" |
| Eine Absprache soll sofort gelten | „Ab heute eintragen" |

**Wenn der Dienst nicht antwortet:**

```
fly logs -a haushalt-api
```

**Unbenutzbar** heißt: Niemand kommt hinein, oder der Plan ist leer, oder
Abhaken geht nicht. Nur dann wird während der Woche eingegriffen — und dann
steht im Protokoll, an welchem Tag und was, damit die Woche vorher und nachher
getrennt ausgewertet werden kann.

---

## 9. Wann wir abbrechen

Ein Abbruch ist kein Scheitern, sondern gesparte Zeit. Abgebrochen wird, wenn
eines davon eintritt:

- **Niemand außer dir öffnet die App bis Mittwoch.** Dann ist die Frage nicht
  mehr, ob die Verteilung fair ist.
- **Der Plan steht drei Tage lang falsch da**, und die vorhandenen Wege
  (umverteilen, streichen) reichen nicht, ihn geradezuziehen.
- **Die App erzeugt Streit statt ihn zu ersparen.** Das ist der einzige
  Abbruchgrund, der sofort gilt, und der einzige, bei dem nicht das Produkt
  Vorrang hat.

---

## 10. Danach

Am Montag nach der Woche: `docs/probewoche-protokoll.md` auswerten,
`docs/offene-fragen.md` um das ergänzen, was die Woche beantwortet und was sie
neu aufgeworfen hat, und in `claude/stand.md` einen Absatz schreiben, der
anfängt mit: *Was eine Woche echter Nutzung gezeigt hat.*

Die drei wahrscheinlichsten Ergebnisse, damit sie schon aufgeschrieben sind
und nicht erst hinterher plausibel wirken:

1. **Erinnerungen sind das nächste große Stück** (wenn M6 schlecht ausfällt).
2. **Die Kapazitätszahlen sind der Hebel** (wenn M1 schief ist — dann nicht am
   Algorithmus drehen, sondern daran, wie die Zahlen zustande kommen).
3. **Die Bibliothek ist zu voll oder zu leer** (M4) — das ist Fleißarbeit und
   kein Entwurf, und deshalb die gute Nachricht unter den dreien.
