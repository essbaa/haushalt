# Offene Fragen

Entscheidungen, die noch nicht getroffen sind.

Die ADRs halten fest, was entschieden wurde, `stolperstellen.md` was schiefging
— beides schaut zurück. Diese Datei schaut nach vorn: Fragen, die anstehen,
mit dem, was für eine Entscheidung fehlt. Sie stehen hier, damit sie nicht nur
in einem Chatverlauf existieren, und damit der Unterschied sichtbar bleibt
zwischen „noch nicht gebaut" und „noch nicht entschieden".

Wandert eine Frage in einen ADR, verschwindet sie hier.

---

## 0. Vor der Probewoche — die Reihenfolge

1. **Den Fragen-Fehler zu Ende bringen.** Beantwortete Fragen kommen nach dem
   Neuladen wieder. Nicht weil die Fragen so wichtig sind, sondern weil ein
   ungeklärter Fehler jeden anderen Befund der Woche verunreinigt: Bei jeder
   Merkwürdigkeit wüsste niemand, ob sie zum Produkt gehört oder zu diesem Bug.
2. **Committen, prüfen, deployen.** Streichen, „Diesmal nicht", Häufigkeit,
   fehlende Bedingung im Klartext, Tagestrennung und `OpenFacts` liegen
   unversioniert. Dieser Deploy ist zugleich der erste echte Lauf des
   `release_command` — Migration 00012 muss vor dem Umschalten auf Neon.
3. **Domain und Resend.** Ohne sie kann niemand außer dem Besitzer des Kontos
   sein Passwort zurücksetzen, und ein verlorenes Passwort ist ein verlorener
   Haushalt.
4. **Einladungen per Mail verschicken**, sobald Resend steht. Heute ist eine
   Einladung ein Code, den der Einladende abschreibt und über einen anderen
   Kanal weitergibt. Das geht für zwei Personen und für niemanden sonst — und
   es ist der Bruch mitten im einzigen Weg, der die App für andere öffnet:
   Adresse eintragen, Mail mit Link, Beitritt beginnt dort, wo er ankommt.
   Bewusst **nach** Punkt 3, weil derselbe Versandweg dranhängt; der Code
   dafür steht in `lib/mail.ts` und wartet nur auf den Schlüssel.
5. **Den Beitritt einmal echt gehen.** Einladungslink, Registrierung, Beitritt,
   erster Blick auf den Plan — von einem Menschen, der nicht weiß, wie die App
   gebaut ist. Der einzige Weg, den nie jemand unter echten Bedingungen
   gegangen ist.

Dazu: **mitschreiben** — inzwischen nicht mehr nur auf Papier. Seit dem
15. September steht unten auf jeder Seite „Stimmt etwas nicht?", und jedes
Mitglied darf melden, nicht nur die planenden (Migration 00014, Tabelle
`feedback`). Der Zettel behält genau eine Aufgabe: festzuhalten, was **nicht**
passiert ist — „heute hat außer mir niemand die App geöffnet" meldet niemand.
Alles Weitere in `docs/probewoche.md`.

---


## 1. Wann kommt das Captcha?

**Stand.** Nicht gebaut, bewusst. Better Auth hat ein Plugin (Cloudflare
Turnstile, hCaptcha, Google reCAPTCHA), das Einbauen ist eine halbe Stunde.

**Was fehlt, ist kein Code, sondern ein Auslöser.** Vor dem ersten Missbrauch
ist ein Captcha eine Hürde für echte Menschen ohne Gegenleistung — und für die
Person, die gerade über einen Einladungslink dazukommt, eine Aufgabe, die sie
nicht versteht.

**Vorschlag für den Auslöser:** sobald die Registrierung öffentlich beworben
wird, oder sobald im Protokoll Anmeldeversuche auftauchen, die niemand aus dem
Haushalt war. Das erste ist eine Entscheidung, das zweite ein Messwert — beide
sind besser als ein Bauchgefühl.

**Zu klären:** Turnstile setzt ein Cloudflare-Konto voraus. Wenn die Domain
ohnehin dort liegt, ist es der kürzeste Weg; sonst ist hCaptcha der neutralere.

---

## 2. Was passiert, wenn die einzige planende Person aussperrt ist?

**Stand.** Passwort zurücksetzen gibt es jetzt (ADR-0014). Es setzt aber
voraus, dass die Adresse stimmt und erreichbar ist.

**Die offene Lücke:** Ein Haushalt mit genau einer planenden Person, deren
Adresse nicht mehr geht — altes Postfach, Tippfehler bei der Registrierung,
Konto beim Anbieter weg. Dann gibt es keinen Weg zurück, der nicht über uns
läuft. Bei einer Familien-App mit zwei Erwachsenen ist das selten; bei einer
App mit fremden Nutzern ist es der häufigste Supportfall überhaupt.

**Denkbare Antworten**, keine davon entschieden: eine zweite planende Person
zur Pflicht machen · ein Wiederherstellungscode beim Einrichten · gar nichts
und ehrlich dazu stehen, solange es die eigene Familie ist.

**Zu klären, bevor Fremde die App benutzen.**

---

## 3. Eigene Domain — welche, und wofür alles?

**Stand.** Entschieden ist, dass es eine gibt (ADR-0014, für den Mailversand).
Gekauft ist keine.

Sie trägt dann drei Dinge auf einmal: den Absender der Mails, die Adresse der
App statt `haushalt-omega.vercel.app`, und — falls das Projekt je eines wird —
den Namen des Produkts.

**Zu klären:** Der Name. Das ist keine technische Frage und die einzige hier,
die man nicht nachrechnen kann.

---

## 4. Kita bringen und abholen — die Aufgabe, die der Planer nicht verteilen kann

**Der Fall**, aus dem echten Leben: Ein Kind in der Kita, beide Eltern
berufstätig. Einer bringt, einer holt ab — und wer was macht, klären sie je
nach Arbeitsplan, oft sonntags für die kommende Woche.

**Warum der Planer hier scheitert**, und zwar nicht aus Versehen:

Er verteilt nach **Kapazität** — wie viel Zeit jemand hat. Was er braucht, ist
**Verfügbarkeit** — ob jemand um 7:45 an einem bestimmten Ort sein kann. Das
sind verschiedene Größen, und die App kennt nur die erste. „30 Minuten am
Dienstag" trifft auf den Elternteil zu, der dienstags um acht eine Besprechung
hat, genauso wie auf den anderen.

Das ist kein neuer Befund. In `stand.md` steht seit Tagen unter „bewusst
offen": **Zeitfenster sind Anzeige.** Was dieser Fall zeigt: Für
ortsgebundene, uhrzeitgebundene Aufgaben ist das kein Schönheitsfehler,
sondern das Ende des Verteilungsmodells. Der Planer darf hier nicht
entscheiden.

**Drei weitere Eigenschaften**, die der Fall mitbringt:

- **Es ist ein Paar.** Bringen und Abholen gehören zusammen und können
  verschiedene Personen sein. Ist nur die eine Hälfte besetzt, ist das kein
  halbes Problem, sondern ein ganzes.
- **Das Scheitern ist hart.** Ein vergessenes Abholen ist keine weiche
  Aufgabe, die morgen auch geht. Abgeben ins Leere darf es nicht geben.
- **Die eigentliche Arbeit ist die Absprache.** „Wer holt Safiya am
  Donnerstag?" — dieses Gespräch *ist* die Kopflast. Damit ist der Fall nicht
  ein Randfall des Produkts, sondern sein schärfster.

**Entschieden am 15. September 2026 — [ADR-0016](adr/0016-absprache-statt-raten.md), gebaut.**
Die Antworten auf die Fragen unten, der Reihe nach: **keine neue Vorlagenart**,
sondern ein Kreuzchen `braucht_absprache` an jeder Vorlage — auch an eigenen;
ob abgesprochen werden muss, ist eine Eigenschaft des Haushalts und nicht der
Aufgabe. **Ja, sie zählt voll in die Bilanz** — und damit steuert der Ausgleich
bei allem anderen gegen: Wer jeden Morgen fährt, bekommt abends weniger. Und
der Knopf „diese Woche wie letzte" fiel weg, weil das Raster gilt, bis jemand
es ändert; Wiederholen ist damit der Normalfall und kostet keinen Handgriff.
Offen bleibt nur die dritte Frage — **wie oft es wirklich beißt**, und das
beantwortet die Probewoche.

Der ursprüngliche Stand, zum Nachlesen:

**Die Richtung**, noch nicht entschieden: Eine **Wochenabsprache** — ein
kleines Raster Wochentag × (bringen / abholen), das die Menschen setzen und
nicht der Planer. Der Plan zeigt sie wie jede andere Aufgabe und rechnet sie in
die Bilanz (es ist Arbeit), aber er verteilt sie nicht. Dazu ein Knopf „diese
Woche wie letzte", weil sich das Raster meistens wiederholt — und genau das
Wiederholen ist die Entlastung.

**Zu klären:**

- Ist das eine neue Vorlagenart („gebunden": wird abgesprochen, nicht
  verteilt) oder eine Erweiterung von `feste_person` auf einzelne Wochentage?
- Zählt eine abgesprochene Aufgabe voll in die Fairnessrechnung? Wenn ja, kann
  ein Elternteil durch die Absprache dauerhaft mehr tragen, ohne dass der
  Ausgleich gegensteuern darf. Das ist vielleicht richtig — aber es muss
  sichtbar sein.
- Wie oft beißt es wirklich? **Das beantwortet die Probewoche**, und nicht
  dieser Absatz.
