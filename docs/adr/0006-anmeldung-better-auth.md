# 0006 — Anmeldung mit Better Auth in der Web-App, Go verifiziert

**Status:** angenommen · 12. September 2026

## Kontext

Die App führt Namen, Rollen und Geburtsjahre von Kindern. Damit ist die Frage
nach dem Anmelde-Anbieter keine Bequemlichkeitsfrage, sondern die Frage, wo
diese Daten liegen — und sie fällt vor einer Gründung noch einmal.

Drei Wege standen zur Wahl. **Clerk** wäre in einer halben Stunde eingebaut,
mit fertigen Masken für Next.js; dafür sitzt ein US-Anbieter mitten im Produkt,
und die Datenresidenz steht nicht in den veröffentlichten Preisen.
**Zitadel** bietet Datenstandort EU oder Schweiz, ist quelloffen und
selbstbetreibbar — aber es bleibt ein Dritter, dem Nutzerdaten gehören.
**Better Auth** ist kein Dienst, sondern eine TypeScript-Bibliothek: Nutzer und
Sitzungen liegen in unserer eigenen Datenbank.

Der Bauplan sagte „Identität einkaufen, nicht selbst bauen". Better Auth ist
der dritte Weg dazwischen: nicht gekauft, aber auch nicht selbst gebaut.

## Entscheidung

**Better Auth läuft in der Next.js-App**, mit dem Postgres, in dem auch die
Haushalte stehen. Angemeldet wird mit E-Mail und Passwort.

**Der Go-Dienst prüft, er stellt nicht aus.** Das JWT-Plugin stellt zwei
Endpunkte bereit: `/api/auth/token` gibt ein signiertes Token,
`/api/auth/jwks` die öffentlichen Schlüssel. Go verifiziert die Signatur
**lokal** gegen den zwischengespeicherten Schlüsselsatz — ohne Rückfrage, ohne
Datenbankabfrage, ohne die Web-App zu erreichen. Signaturverfahren ist EdDSA
(Ed25519), und nur das wird akzeptiert.

**Wem welche Tabelle gehört, ist festgelegt:** `user`, `session`, `account`,
`verification` gehören Better Auth und werden über dessen CLI migriert; alles
andere gehört goose. Keines der beiden fasst die Tabellen des anderen an. Die
einzige Verbindung ist die Spalte `member.auth_user_id`.

**Die Sitzung ist ein Cookie, das Token ist für den Fremddienst.** Die
Dokumentation des Plugins sagt das ausdrücklich, und wir halten uns daran: Im
Browser gilt die Sitzung, der Go-Dienst bekommt ein kurzlebiges Token. Zwei
Mechanismen, klar getrennte Zuständigkeit — nicht zwei, die dasselbe tun.

**Verifiziert wird mit einer Bibliothek, nicht von Hand.** Ein Ed25519-JWT
selbst zu prüfen ist überschaubar, und die Versuchung war da. Die Fehler sitzen
aber nicht in der Signaturprüfung, sondern in den Details: Algorithmus
festnageln (sonst Alg-Confusion), Schlüsselrotation, Ablaufzeiten,
Aussteller. Das ist die eine Stelle im Projekt, an der „selbst gemacht" kein
Senior-Signal ist, sondern das Gegenteil.

## Konsequenzen

**Dafür:**

- Es gibt keinen Dritten. Nutzerdaten liegen in derselben Datenbank in
  Frankfurt wie alles andere; bei einer Gründung ist nichts zu erklären.
- Kostenlos und ohne Nutzerobergrenze — kein Preissprung bei Erfolg.
- Die Verifikation in Go ist echte Backend-Arbeit und im Gespräch besser zu
  erklären als ein eingebundenes Anbieter-SDK.

**Dagegen — und das ist der Preis, den ADR-0001 nicht vorgesehen hatte:**

- **Die Next.js-App ist nicht mehr der dünne Client**, als der sie dort
  beschrieben ist. Sie ist der Autorisierungsserver, schreibt selbst in die
  Datenbank und braucht dafür eigene Zugangsdaten.
- **Die Anmeldung hängt an der Verfügbarkeit der Web-App.** Vorher war der
  Go-Dienst unabhängig benutzbar.
- **Zwei Migrationswerkzeuge auf einer Datenbank.** Die Regel oben hält das
  auseinander, aber es bleibt eine Regel, die jemand kennen muss.
- **Passwörter sind unser Problem.** Zurücksetzen, Sperren nach Fehlversuchen,
  Zwei-Faktor — alles Funktionen, die Better Auth mitbringt, aber die wir
  konfigurieren und verstehen müssen, statt sie zu bestellen.

## Verworfene Alternativen

**Clerk.** Schnellste Integration, beste Masken. Verworfen wegen der
Datenresidenz und weil die Go-Seite damit zur Formsache geworden wäre.

**Zitadel mit Datenstandort EU.** Sauberer OIDC-Standard, quelloffen,
selbstbetreibbar. Verworfen, weil Better Auth dasselbe Ergebnis ohne jeden
Dritten erreicht — und weil ein Anbieter weniger ein Vertrag weniger ist.

**Anmeldung im Go-Dienst selbst bauen.** Verworfen aus demselben Grund, aus
dem der Bauplan es ausschloss: Passwort-Hashing, Sitzungen, Bestätigungsmails
und Zurücksetzen sind Wochen Arbeit an einem Problem, das gelöst ist.

## Wann wir das revidieren

Wenn Haushalte über Single-Sign-on eines Arbeitgebers oder über
Behörden-Identitäten hereinkommen sollen. Dann ist ein Anbieter mit
SAML-Erfahrung im Vorteil, und die Frage wird neu gestellt — der Go-Dienst
merkt davon nichts, weil er ohnehin nur JWKS und ein Token kennt.
