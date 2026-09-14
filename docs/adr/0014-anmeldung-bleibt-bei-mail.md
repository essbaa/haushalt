# 0014 — Anmeldung bleibt bei E-Mail und Passwort

**Status:** angenommen · 14. September 2026 · ergänzt
[0006](0006-anmeldung-better-auth.md)

## Kontext

ADR-0006 entschied, *womit* angemeldet wird (Better Auth in der Web-App, Go
prüft über JWKS). Offen blieb, *wie* sich jemand ausweist — und die
Registrierung per Mail war nur zur Hälfte gebaut: Adresse und Passwort,
sonst nichts.

Die Frage ist nicht technisch. Sie lautet: *Wer meldet sich an?* Heute zwei
Erwachsene in einem Haushalt, und einer davon hat die App nicht gebaut. Für
den zählt, wie viele Schritte zwischen Einladungslink und erstem Wochenplan
liegen — und was passiert, wenn das Passwort weg ist.

## Entscheidung

**Keine Drittanbieter.** Weder Google noch Apple.

**Die Registrierung per Mail wird vollständig**, in dieser Reihenfolge, nach
Schmerz sortiert und nicht nach Aufwand:

1. **Passwort zurücksetzen.** Ohne das heißt ein vergessenes Passwort:
   verlorener Haushalt. Kein Komfortmangel — Datenverlust.
2. **Adresse bestätigen.** Ohne bestätigte Adresse ginge die Rücksetzmail an
   einen Tippfehler, und der Mensch wartet auf etwas, das nie kommt.
3. **Ein Passwortfeld mit Auge** statt zweier Felder. Menschen machen
   denselben Tippfehler zweimal; ein zweites Feld bestätigt ihn dann nur.
   Sehen zu können, was man getippt hat, fängt ihn wirklich — und auf dem
   Telefon ist es ein Feld weniger über der Tastatur.
   *Revidiert am selben Abend — siehe Nachtrag.*
4. **Acht Zeichen Mindestlänge, keine erzwungenen Sonderzeichen.** Regeln wie
   „mindestens eine Ziffer und ein Satzzeichen" erzeugen `Passwort1!`.
5. **Begrenzung der Versuche** (Better Auth bringt sie mit, sie war nur aus).

**Verschickt wird über Resend mit eigener Domain.** Der Grund ist eine
Einschränkung, keine Vorliebe: An beliebige Adressen zustellen darf nur, wer
eine Domain verifiziert hat. Ohne das käme die Bestätigungsmail beim Absender
an und sonst nirgends — die Person, die sie braucht, bekäme sie nie.

**Solange kein Schlüssel gesetzt ist, landet die Mail in der Serverkonsole**,
statt dass etwas scheitert. Damit ist die Registrierung vollständig prüfbar,
bevor eine Domain existiert, und der Übergang kostet keine Zeile Code.

**`requireEmailVerification` bleibt vorerst aus.** Eine Anmeldung, die eine
Bestätigung verlangt, die niemand bekommen kann, sperrt alle aus. Der Auslöser
zum Umlegen ist benannt: sobald Resend mit eigener Domain läuft und eine
Testmail ankommt.

## Konsequenzen

**Dafür:**

- Kein Dritter erfährt, wer diese App benutzt. Bei einer App mit Namen und
  Alter von Kindern ist das die stärkste Antwort, die es gibt — und sie
  passt zu ADR-0006, wo aus demselben Grund kein Anmeldedienst gewählt wurde.
- Keine 99 $ im Jahr für das Apple Developer Program, das auch eine reine
  Web-App ohne App Store braucht.
- Keine Datenschutzerklärung, die eine Weitergabe an Google erklären müsste.
- Der Mailversand, den das hier braucht, ist derselbe, den die fehlenden
  Erinnerungen brauchen. Das sind nicht zwei Aufgaben, das ist eine.

**Dagegen:**

- **Ein Passwort mehr im Kopf.** Genau der Schritt, den „Mit Apple anmelden"
  auf dem iPhone spart — und der Mensch, der ihn gehen muss, ist der, der die
  App nicht gebaut hat.
- Eine Domain kostet Geld und Einrichtung (DNS-Einträge).
- Unbestätigte Adressen sind möglich, solange die Pflicht aus ist. Wer in
  diesem Fenster sein Passwort vergisst, hat keinen Weg zurück, der nicht
  über uns läuft.

## Verworfene Alternativen

**Google-Anmeldung.** Kostenlos und ein Tipp statt eines Passworts. Verworfen,
weil sie eine dritte Partei in eine App über Haushaltsaufgaben und Kinder
bringt — dieselbe Abwägung wie in ADR-0006, dort zugunsten der eigenen
Datenbank entschieden. Es wäre seltsam, den Anmeldedienst wegzulassen und den
Anmeldeanbieter hereinzuholen.

**Apple-Anmeldung.** Auf dem iPhone der kürzeste Weg. Verworfen am Preis:
99 $ im Jahr sind eine jährliche Wette auf ein Projekt ohne Einnahmen.

**Gmail-SMTP statt Resend.** Sofort und kostenlos. Verworfen, weil der
Absender dann eine private Adresse ist und die Zustellbarkeit an Fremde
schlecht — für zwei Personen reicht es, und genau darauf sollte man eine
Anmeldung nicht bauen.

**Captcha jetzt.** Better Auth hat ein Plugin (Turnstile, hCaptcha,
reCAPTCHA). Nicht verworfen, sondern vertagt — mit benanntem Auslöser: sobald
die Registrierung offensteht und jemand sie missbraucht. Vorher ist es eine
Hürde für echte Menschen ohne Gegenleistung.

## Wann wir das revidieren

Wenn jemand außerhalb der Familie die App benutzt und am Anlegen eines
Passworts scheitert. Dann ist Google die naheliegende Ergänzung — und die
Datenschutzerklärung wird sowieso fällig.

Und sobald es eine App im Store gibt, ist die Apple-Anmeldung keine Wahl mehr,
sondern eine Auflage.

## Nachtrag, 14. September 2026 — zweites Feld doch, und eine Leiste dazu

Punkt 3 ist zurückgenommen: Bei der Registrierung steht jetzt **beides**, ein
Feld mit Auge *und* eine Wiederholung. Das Argument gegen die Wiederholung
bleibt richtig (wer zweimal gleich falsch tippt, bestätigt seinen Fehler nur),
aber es ist kein Argument dafür, sie wegzulassen: Sie fängt den Fall ab, in
dem jemand beim zweiten Mal *anders* danebengreift, und sie ist das, was
Menschen an dieser Stelle erwarten. Ein Formular, das eine berechtigte
Erwartung nicht erfüllt, muss dafür einen sichtbaren Vorteil bieten — hier
wäre es ein Feld weniger, und das wiegt den Verlust nicht auf.

Geprüft wird im Browser, bevor gesendet wird. Ein Rundgang zum Server, um zu
erfahren, dass zwei Felder nicht gleich sind, ist Wartezeit ohne Erkenntnis.

**Dazu eine Stärkeanzeige** (`lib/passwort.ts`), und die hat eine Eigenheit,
die Absicht ist: **Die Länge gibt den Ausschlag, die Zeichenklassen geben
einen Bonus.** Die naheliegende Rechnung zählt Klassen und ist falsch herum —
sie gibt `Passwort1!` die Bestnote und `pferdeklammerbatterie` die
schlechteste, obwohl das zweite um Größenordnungen schwerer zu raten ist. Wer
rät, rät mit Wörterbüchern und mit den Mustern, die Menschen benutzen: großer
Anfangsbuchstabe, Ziffer am Ende, Ausrufezeichen dahinter.

Die Klassen stehen trotzdem als Merkmale daneben, weil sie einen brauchbaren
Rat geben — aber als **Rat und nicht als Pflicht**. Ein langes Passwort ohne
Sonderzeichen bekommt „sehr stark", und das bleibt so. Erzwungene
Sonderzeichen erzeugen genau das `Passwort1!`, gegen das die Rechnung antritt.

Abgezählt wird außerdem, was offensichtlich nichts taugt: die zwanzig
häufigsten Passwörter, der eigene Name oder Adressteil darin, ein wiederholtes
Zeichen, eine Tastaturreihe.

**Kein zxcvbn.** Die Bibliothek rechnet besser und wiegt mehrere hundert
Kilobyte im Bundle einer App, die auf dem Telefon geöffnet wird. Für die Frage
„ist das offensichtlich zu wenig" reicht die eigene Rechnung.
