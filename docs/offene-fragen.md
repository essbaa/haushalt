# Offene Fragen

Entscheidungen, die noch nicht getroffen sind.

Die ADRs halten fest, was entschieden wurde, `stolperstellen.md` was schiefging
— beides schaut zurück. Diese Datei schaut nach vorn: Fragen, die anstehen,
mit dem, was für eine Entscheidung fehlt. Sie stehen hier, damit sie nicht nur
in einem Chatverlauf existieren, und damit der Unterschied sichtbar bleibt
zwischen „noch nicht gebaut" und „noch nicht entschieden".

Wandert eine Frage in einen ADR, verschwindet sie hier.

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
