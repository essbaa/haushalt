# Stolperstellen

Jeder Fehler, der uns beim Bau von `haushalt` aufgehalten hat — mit Symptom,
Ursache, Lösung und dem, was davon bleibt. Die Reihenfolge ist thematisch, nicht
zeitlich. Wer einen ähnlichen Fehler sieht, soll ihn hier in einer Minute
wiederfinden.

Zwei Sätze fassen die meisten Einträge zusammen:

> **Grüne Tests sagen, dass der Code tut, was er soll. Ob er das Richtige soll,
> steht in der Ausgabe.**
> Kein einziger Planer-Fehler unten wurde von der Testsuite gefunden — drei
> beim Lesen von `make plan`, einer beim Lesen des Codes selbst.

> **Das Schema trifft Produktentscheidungen, ob man will oder nicht.**
> Ein `UNIQUE` an der falschen Spalte hat eine Produktregel behauptet, die nie
> jemand beschlossen hatte.

Und eine Regel für Werkzeuge:

> **Eine Regel, an die man sich erinnern muss, ist keine Regel.**
> Was zweimal vergessen wurde, gehört nicht in den Kopf, sondern in die
> Werkzeugkette.

---

## 1. Umgebung und Werkzeugkette

### 1.1 Go-Proxy des alten Arbeitgebers

**Symptom** — `go mod download` scheiterte mit Verbindungsfehlern gegen
`athens.unimed.systems`. Dreimal, an drei verschiedenen Tagen.

**Ursache** — Eine `GOPROXY`-Umgebungsvariable aus einer Shell-Konfiguration des
alten Arbeitgebers. Eine echte Umgebungsvariable schlägt `go env -w` — deshalb
half das Zurücksetzen der Go-Konfiguration nicht.

**Lösung** — `GOPROXY` im `api/Makefile` unbedingt setzen und exportieren:

```make
GOPROXY := https://proxy.golang.org,direct
export GOPROXY
```

Damit ist das Repo immun, egal was in der Shell steht. Um die Quelle zu finden:
`zsh -xlic exit 2>&1 | grep -in "goproxy\|athens"`.

**Lehre** — Reste alter Infrastruktur (Go-Proxy, npm-Registry) sind nicht nur
lästig, sondern auch eine Frage der IP-Hygiene: Baue nie ein eigenes Produkt
über die Infrastruktur eines Arbeitgebers.

### 1.2 Neon lokal nicht erreichbar

**Symptom** — Verbindungen nach `…neon.tech:5432` liefen in einen Timeout.

**Ursache** — Netzwerk beziehungsweise VPN, **nicht** die Zugangsdaten. Erkannt
mit `nc -vz <host> 5432`: später ging derselbe Aufruf ohne Änderung durch.

**Lösung** — Erst die Erreichbarkeit prüfen, dann die Zugangsdaten verdächtigen.
Zusätzlich startet der Dienst im Entwicklungsmodus auch ohne Datenbank
(`unavailableDB`), damit ein Netzwerkproblem nicht die ganze Arbeit blockiert.

**Lehre** — Ein Timeout ist kein Authentifizierungsfehler. Die beiden sehen im
Log ähnlich aus und haben nichts miteinander zu tun.

### 1.3 Zu altes Node in der Shell

**Symptom** — `structuredClone is not defined` beim Linten. Zweimal.

**Ursache** — Die Shell hatte eine alte Node-Version aktiv, ESLint braucht eine
neuere.

**Lösung** — `web/.nvmrc` mit `22`, `engines.node >= 20.9` in der
`package.json`, und im Zweifel `nvm use` vor der Arbeit im `web`-Verzeichnis.

### 1.4 Turbopack fand die falsche Projektwurzel

**Symptom** — Jeder `@/…`-Import scheiterte, obwohl die Pfade stimmten.

**Ursache** — Eine verirrte `package-lock.json` im Heimatverzeichnis. Turbopack
sucht nach oben nach einer Sperrdatei und erklärte `/Users/zessbaa` zur Wurzel.

**Lösung** — Die fremde Datei löschen **und** die Wurzel festnageln, damit es
nicht wiederkommt:

```ts
// web/next.config.ts
turbopack: { root: __dirname }
```

**Lehre** — Werkzeuge, die „automatisch die Wurzel finden“, finden irgendwann
die falsche. Wenn es eine Möglichkeit gibt, sie explizit zu setzen: setzen.

### 1.5 Mehrzeiliges Einfügen im Terminal

**Symptom** — Mehrere Befehle in einem Codeblock kamen im Terminal
zusammengeklebt an und liefen falsch.

**Lösung** — Ein Befehl pro Codeblock. Keine technische Erkenntnis, aber sie hat
uns mehrmals Zeit gekostet.

---

## 2. Go, Build und Erzeugung

### 2.1 Go-Version im Docker-Build

**Symptom** —

```
go: go.mod requires go >= 1.25.0 (running go 1.24.x; GOTOOLCHAIN=local)
```

und eine Woche später dasselbe mit `>= 1.26.0`.

**Ursache** — `api/Dockerfile` hatte die Version fest verdrahtet
(`FROM golang:1.25-alpine`), `api/go.mod` zog weiter. Beim ersten Mal haben wir
nur das Basisimage hochgezogen — also den Fehler verschoben, nicht behoben.

**Lösung** — Eine Wahrheit, die `go.mod` heißt. Das Dockerfile nimmt die Version
als Argument entgegen:

```dockerfile
ARG GO_VERSION=1.26
FROM golang:${GO_VERSION}-alpine AS build
```

Das Makefile liest sie aus der `go.mod` und reicht sie herein:

```make
GO_VERSION := $(shell awk '/^go /{split($$2,v,"."); print v[1]"."v[2]; exit}' go.mod)
```

Die CI tut dasselbe (`.github/workflows/api.yml`), und `actions/setup-go` liest
sie ohnehin schon über `go-version-file: api/go.mod`.

**Lehre** — Eine Regel, an die man sich erinnern muss, ist keine Regel. Beim
zweiten Auftreten desselben Fehlers ist nicht der Fehler das Problem, sondern
die Stelle, die ihn ermöglicht.

### 2.2 `run:` ohne Block-Skalar in GitHub Actions

**Symptom** — Der Deploy-Schritt sah richtig aus, aber `--build-arg
GO_VERSION=` kam leer an. Die Datei war gültiges YAML, deshalb schlug keine
Prüfung an.

**Ursache** — Ein einfacher Skalar über zwei Zeilen wird von YAML mit einem
Leerzeichen verbunden. Aus

```yaml
run: GO_VERSION=$(awk …)
  flyctl deploy --build-arg GO_VERSION="$GO_VERSION"
```

wurde eine einzige Zeile `GO_VERSION=… flyctl deploy …` — also eine
Zuweisung als Präfix eines Befehls. Die Shell ersetzt `"$GO_VERSION"` aber
**bevor** diese Zuweisung wirkt, das Argument ist leer.

**Lösung** — Mehrzeilige Befehle immer als Block-Skalar:

```yaml
run: |
  GO_VERSION=$(awk '/^go /{split($2,v,"."); print v[1]"."v[2]; exit}' go.mod)
  flyctl deploy --remote-only --build-arg GO_VERSION="$GO_VERSION"
```

**Lehre** — „Das YAML parst“ heißt nicht „der Befehl stimmt“. Bei
Workflow-Änderungen den erzeugten Shell-Text ansehen, nicht nur die Einrückung.

### 2.3 `gofmt` bricht `make check`

**Symptom** — `make check` rot, ohne Compilerfehler. Zweimal.

**Ursache** — Beide Male der Importblock: einmal eine übrig gebliebene
Leerzeile, einmal `time` an der falschen alphabetischen Stelle.

**Lösung** — `gofmt -w` vor dem Commit; `make check` läuft ohnehin lokal, bevor
etwas gepusht wird.

### 2.4 sqlc erzeugt Zeigertypen

**Symptom** — Übersetzungsfehler beim Befüllen erzeugter Strukturen:
`pgtype.Text` passte nicht auf `*string`.

**Ursache** — `emit_pointers_for_null_types: true` in der `sqlc.yaml`. Damit
werden nullable `text`/`int` zu `*string`/`*int32`, während `uuid`, `date` und
`timestamptz` ihre `pgtype`-Typen mit `Valid`-Flag behalten.

**Lösung** — Die Regel kennen statt sie zu raten: nullable Text und Zahlen sind
Zeiger, die übrigen bleiben `pgtype`.

### 2.5 Erzeugter Code lief auseinander

**Symptom** — `next build` scheiterte, weil `meine_rolle` im Typ fehlte,
obwohl das Feld in `openapi.yaml` stand.

**Ursache** — `web/lib/api-types.ts` war veraltet: Spezifikation geändert,
Erzeugung vergessen.

**Lösung** — `npm run generate` beziehungsweise `make generate`, und beides in
der CI mit `git diff --exit-code` abgesichert. Wer die Spezifikation ändert und
die Erzeugung vergisst, bekommt sofort einen Diff statt drei Tage später einen
Client, der Felder erwartet, die der Server nicht schickt.

### 2.6 Signatur geändert, Aufrufer vergessen

**Symptom** — `catalog.Households(ctx)` — zu wenige Argumente.

**Ursache** — Das `Plans`-Interface bekam mit der Anmeldung überall ein
`subject`; die Bibliotheks-Implementierung zog nicht mit.

**Lösung** — Beide Implementierungen eines Interface gleichzeitig anfassen. Der
Compiler findet es zuverlässig, deshalb ist es billig — aber nur, wenn man
`make check` laufen lässt, bevor man weitergeht.

---

## 3. Der Planer

Keiner dieser Fehler kam aus der Testsuite. Drei wurden beim **Lesen von
`make plan`** gefunden, einer beim Lesen des Codes. Die Tests waren die ganze
Zeit grün.

### 3.1 Fenster-Vorlagen wurden nur einmal fällig

**Symptom** — Eine Vorlage mit Rhythmus „alle drei Tage“ tauchte im Wochenplan
genau einmal auf.

**Ursache** — `dueEvery` erzeugte einen einzigen Kandidaten mit einem Fenster
über die ganze Woche, statt einen Kandidaten je Fälligkeit.

**Lösung** — Eine Schleife über die Fälligkeiten, jede mit eigenem Fenster
(`api/internal/planner/due.go`):

```go
for start := next; !start.After(sunday); start = start.AddDays(every) {
    end := start.AddDays(every - 1)
    if end.After(sunday) { end = sunday }
    c := candidate{tmpl: t, days: daysBetween(start, end)}
    if t.Failure == FailureHard { c.deadline = end }
    out = append(out, c)
}
```

### 3.2 Der Ausgleich brach die Rotation innerhalb der Woche

**Symptom** — Anna kochte an drei Abenden hintereinander. Fair nach Zahlen,
falsch im Leben.

**Ursache** — Der Ausgleichsdurchgang tauschte Aufgaben paarweise, ohne zu
prüfen, ob dieselbe Person dieselbe Vorlage in derselben Woche schon hält. Die
Rotation über Wochen hinweg war geprüft, die innerhalb der Woche nicht.

**Lösung** — Fünfte Regel in `swapAllowed` (`rebalance.go`, `holdsOther`), ein
Regressionstest und ein Nachtrag in ADR-0003.

**Lehre** — Wer eine Invariante an einer Stelle prüft, muss sie an jeder Stelle
prüfen, die Zuordnungen verändert.

### 3.3 113 % Auslastung

**Symptom** — Die Bilanz zeigte eine Auslastung über 100 %.

**Ursache** — Zähler und Nenner waren verschiedene Größen: geteilt wurde das
**gewichtete** Maß (Minuten + Kopflast) durch eine Kapazität, die nur in Minuten
gemessen ist.

**Lösung** — `MemberLoad.Utilization()` rechnet mit `Minutes`, nicht mit
`Weighted`. Die Kopflast bleibt in der Fairness-Bewertung des Ausgleichs, wo sie
hingehört.

**Lehre** — Zwei Achsen (Dauer und Kopflast) sind ein Gewinn fürs Produkt und
eine ständige Einladung, Äpfel durch Birnen zu teilen.

### 3.4 Ein Erwachsener ohne Geburtsjahr galt als Kind

**Symptom** — Keiner. Gefunden beim Lesen von `IsAdult()`, während das
Onboarding gebaut wurde. Der Fehler stand zu dem Zeitpunkt bereits in
Produktion.

**Ursache** — `IsAdult()` lautete `m.Role == RolePlanner || m.Age >= 18`. Wer
über eine Einladung als **ausführende** Person dazukommt, bekommt aber kein
Geburtsjahr — wir fragen es nicht ab und erfinden es nicht. Alter 0, also Kind.

**Folgen, beide unsichtbar** — Diese Person bekam keine einzige Aufgabe, die
Erwachsene voraussetzt. Und sie zählte als Kind des Haushalts, was darüber
entscheidet, welche Vorlagen überhaupt fällig werden — ein Haushalt „mit Kind"
hat andere Aufgaben als einer ohne. Betroffen war ausgerechnet der Fall, um den
sich das halbe Produktkonzept dreht: der planungsunwillige Partner.

**Lösung** — Ein unbekanntes Alter heißt jetzt erwachsen:

```go
case m.Age <= 0:
    return true
```

Die Gegenrichtung kostet weniger, weil Kinder im Onboarding immer ein
Geburtsjahr bekommen und betreute Personen eines haben müssen. Ein Kind ohne
Jahr kann nur noch über einen Import ohne Altersangabe entstehen — und dort ist
die fehlende Angabe der Fehler.

**Lehre** — Ein Nullwert, der „nicht gefragt" bedeutet, wird irgendwo als
„gemessen null" gelesen. Bei Alter, Preis und Menge ist das keine Frage des Ob,
sondern des Wann.

### 3.5 Demo-Pläne verschieben sich nach `db-reset`

**Symptom** — Nach `make db-reset` sieht der Beispielplan anders aus als vorher,
obwohl sich am Code nichts geändert hat.

**Ursache** — Bei Gleichstand entscheidet die Mitglieds-UUID, und die wird beim
Import neu vergeben.

**Status** — Bekannt und bewusst offen. Behebbar mit rund 20 Zeilen
(deterministische Kennungen beim Import), lohnt sich, sobald Screenshots oder
Demos stabil sein müssen.

---

## 4. Datenbank und Schema

### 4.1 Das Schema hatte eine Produktentscheidung getroffen

**Symptom** — Beim Beitritt zu einem zweiten Haushalt:
`duplicate key value violates unique constraint "member_auth_user_id_key"`.

**Ursache** — `auth_user_id` war global eindeutig. Damit behauptete das Schema
„ein Mensch, ein Haushalt“ — eine Regel, die nie jemand beschlossen hatte. Sie
fällt genau in dem Moment auf, in dem jemand Patchwork lebt oder den Haushalt
der Eltern mitplant.

**Lösung** — Migration `00004_mitglied_je_haushalt.sql`: eindeutig **je
Haushalt** statt global. `GetMemberByAuthUserID` wurde zu
`GetMemberInHousehold`.

**Lehre** — Jeder `UNIQUE` und jeder `NOT NULL` ist eine Produktaussage. Beim
Schreiben einer Migration laut aussprechen, was die Bedingung behauptet.

### 4.2 Der Importer datierte alle Ereignisse auf heute

**Symptom** — Die importierte Historie war da, aber jedes Ereignis trug den
Zeitpunkt des Imports. Die Rotation konnte damit nichts anfangen.

**Ursache** — `occurred_at` hatte einen Vorgabewert `now()`, der Importer setzte
ihn nicht.

**Erschwerend** — Nachbessern ging nicht: Der Append-only-Trigger
(`event_ist_unveraenderlich`) blockiert `UPDATE` und `DELETE` auf `event`. Genau
dafür ist er da.

**Lösung** — `make db-reset` und eine eigene Abfrage `AppendHistoricEvent`, die
den Zeitpunkt mitgibt (`mittags(tag)` = 12:00 UTC, damit keine Zeitzone die
Datumsgrenze verschiebt — siehe ADR-0002).

**Lehre** — Ein append-only-Log ist eine Zusage, keine Einstellung. Wer sie
gibt, muss Daten beim ersten Schreiben richtig schreiben.

---

## 5. Web

### 5.1 React Compiler verbietet die Zuweisung an `window.location`

**Symptom** — Der Build scheiterte an `window.location.href = …`.

**Ursache** — Der React Compiler verlangt, dass Komponenten frei von solchen
Seiteneffekten sind.

**Lösung** — `const router = useRouter()`, dann `router.push(ziel)` und, wenn
Server-Komponenten die neue Sitzung sehen sollen, `router.refresh()`.

### 5.2 Eine Regel, die das Formular nur anzeigte

**Symptom** — Beim Anlegen des ersten Haushalts:
`eingabe ergibt keinen haushalt: wer den Haushalt einrichtet, muss darin
planen`. Im Formular war das Häkchen „plant mit" bei der eigenen Person
gesetzt — ausgegraut, damit niemand es abwählt.

**Ursache** — Ausgegraut heißt nicht unveränderlich. Der Zustand dahinter
konnte kippen, sobald man oben die Art umschaltete
(`plant: art === "erwachsen" && p.plant`). Das Häkchen zeigte die Regel an,
statt sie durchzusetzen; durchgesetzt wurde sie erst im Server, und dort kam
sie als Fehlermeldung zurück statt als unmögliche Eingabe.

**Lösung** — Bei der einrichtenden Person steht jetzt kein Schalter mehr,
sondern ein Satz. Die Rolle wird beim Absenden gesetzt, nicht aus dem Formular
abgeleitet.

**Lehre** — Eine Regel, die man nur anzeigt, ist keine. Ein deaktiviertes
Bedienelement ist eine Erklärung, keine Absicherung — und wenn es einen
Zustand spiegelt, der sich woanders ändern kann, ist es eine falsche
Erklärung.

### 5.3 Offene Weiterleitung in `?weiter=`

**Symptom** — Kein Fehler, sondern ein Fund beim Durchlesen: Der
Anmelde-Parameter `?weiter=` hätte auf eine fremde Domain zeigen können.

**Lösung** — Nur Pfade akzeptieren, und keine, die mit `//` beginnen.

---

## 6. Betrieb

### 6.1 Fly verlangt eine Kreditkarte

**Symptom** — `fly secrets set` scheiterte mit
„trial has ended, please add a credit card“.

**Ursache** — Fly verlangt inzwischen für jede Organisation eine hinterlegte
Karte, auch bei kleinem Verbrauch.

**Lösung** — Karte hinterlegt und die Maschine auf 256 MB verkleinert
(`api/fly.toml`). Größenordnung: 256 MB rund 1,94 $/Monat, 512 MB rund
3,19 $/Monat, 1 GB rund 5,70 $/Monat; gestoppte Maschinen kosten nur Speicher.
Mit `auto_stop` und `min_machines_running = 0` läuft der Dienst ohnehin nur,
wenn jemand ihn braucht.

### 6.2 Zwei Maschinen statt einer

**Symptom** — `fly status` zeigte zwei Maschinen, obwohl eine reicht.

**Lösung** — `fly scale count 1`. Fly legt beim ersten Deploy gern ein Paar an.

### 6.3 Ein falscher `jq`-Filter

**Symptom** — `select(.id | test("-"))` sollte einen bestimmten Haushalt
herausfiltern und traf auch `familie-a`.

**Ursache** — Der Filter war schlicht falsch gedacht — beide Kennungen
enthalten einen Bindestrich.

**Lösung** — Die Kennung direkt setzen statt sie zu erraten.

**Lehre** — Ein Filter, der beim ersten Versuch „fast passt“, hat meistens gar
keine Bedingung, sondern eine Gewohnheit.

---

## 7. Fehler, die keine waren

- **`FAIL: TestWeekMonday`** — absichtlich gebrochen (Commit `29947e3`), um zu
  sehen, ob die CI wirklich rot wird. Sie wurde. Zurückgenommen in `81d4a79`.
  Eine CI, die noch nie rot war, ist keine geprüfte CI.
- **„This job was skipped“** — der Pfadfilter der Workflows. Eine reine
  Frontend-Änderung lässt den API-Workflow nicht laufen. Ein Repo, zwei Dienste,
  zwei unabhängige Auslieferungen.

---

## 8. Sicherheitsregeln, die aus diesen Tagen stammen

- `.env` und `.env.local` stehen in `.gitignore` und werden nie committet. Für
  die Neon-Zugangsdaten wurde geprüft, dass sie nie in der Git-Historie standen.
- Produktionstokens gehören nicht in einen Chat. Zum Debuggen reichen Kopf und
  Nutzlast eines JWT — die Signatur nie mitschicken.
- `NEXT_PUBLIC_*` ist per Definition öffentlich und gehört auf Vercel unter
  „Config“. `DATABASE_URL` und `BETTER_AUTH_SECRET` sind Secrets und dürfen das
  Präfix niemals tragen.
- Ein fremder Haushalt antwortet mit 404, nicht mit 403. Ein 403 verrät, dass es
  ihn gibt.
- Ungültige und abgelaufene Einladungscodes bekommen dieselbe, ununterscheidbare
  Antwort.
