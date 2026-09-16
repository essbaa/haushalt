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

### 2.0 Der laufende Dienst war dreimal der Fehler

**Symptom** — `405 Method Not Allowed` auf einen neuen Endpunkt, `404` auf eine
neue Route, eine 500 auf eine neue Tabelle.

**Ursache** — Jedes Mal lief der Go-Dienst noch mit dem Stand von vor
`make generate` beziehungsweise `make migrate-up`.

**Lehre** — Ein 405 heißt „den Pfad gibt es, die Methode nicht" und ist damit
fast immer ein alter Prozess. Beim dritten Mal war es Zeit für eine Regel:
**Nach jedem Erzeugen neu starten.**

### 2.1 `make generate` erzeugte nur die Hälfte

**Symptom** — `p.db.DeleteSignal undefined` — eine Abfrage, die in der
SQL-Datei stand und im Go-Code fehlte.

**Ursache** — `make generate` erzeugte nur die Vertragstypen; die
Datenbankabfragen waren ein zweites Ziel (`make sqlc`), an das man denken
musste.

**Lösung** — `generate: sqlc` als Abhängigkeit, und `verify-generate` prüft
jetzt auch `internal/storage/db` auf Drift. Vorher wäre ein vergessener
sqlc-Lauf unbemerkt durch die CI gegangen.

**Lehre** — Wieder dieselbe: Eine Regel, an die man sich erinnern muss, ist
keine Regel.

### 2.2 gofmt richtet Feldblöcke aus — dreimal gestolpert

**Symptom** — `make check` rot, ohne Compilerfehler, dreimal an einem Tag an
derselben Art Stelle.

**Ursache** — gofmt richtet zusammenhängende Feldblöcke aneinander aus, in
Struktur-Definitionen wie in Literalen. Ein neues Feld, dessen Name länger ist
als alle bisherigen, verschiebt den ganzen Block. Ein Kommentar mittendrin
beginnt einen neuen Block, eine Leerzeile ebenfalls.

**Lösung** — Neue Felder ans Ende, durch eine Leerzeile abgetrennt. Dann bilden
sie ihre eigene Gruppe und stören die bestehende nicht.



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

### 3.5 Der Montag trug die halbe Woche

**Symptom** — Vier Organisationsaufgaben, alle am selben Tag: Vorsorgetermin,
Inspektion, Elternbeitrag, Post. Kopflast 3+2+2+3 an einem Montag, der Rest der
Woche zusammen etwa genauso viel.

**Ursache** — Die Startdichte begrenzt Organisationsaufgaben **je Woche** und
sagt nichts darüber, wie sie liegen. Organisationsaufgaben haben die höchste
Dringlichkeit, bekommen also den frühesten Tag — und der früheste Tag ist für
alle derselbe.

**Lösung** — `MaxHeadLoadPerDay` (Vorgabe 4). Erst der Versuch mit Grenze, dann
derselbe Durchgang ohne: Die Grenze ist eine Vorliebe, keine Bedingung, sonst
fällt eine Aufgabe ganz weg, statt einen vollen Tag zu bekommen.

**Nebeneffekt, der nicht geplant war** — `keine_kapazitaet` schrumpfte von drei
Aufgaben auf eine. Wer die Kopfarbeit verteilt, macht auch Platz für die
kleinen Sachen.

### 3.6 Die harte Frist als Freifahrtschein

**Symptom** — Nach der Kopflastgrenze blieb der Montag trotzdem voll.

**Ursache** — Ich hatte Aufgaben mit harter Frist von der Grenze *ausgenommen*.
Und genau diese Gruppe ist in der Bibliothek als „hart" markiert: Vorsorge,
Elternbeitrag, Post. Die Regel griff also überall außer dort, wofür es sie gab.

**Lösung** — Die Ausnahme wurde zum **Rückfall**: Erst suchen alle einen ruhigen
Tag, und nur wer keinen findet, nimmt einen vollen. Der Sonderfall „hart"
verschwand ersatzlos.

**Lehre** — Eine Ausnahme für die Wichtigsten ist fast immer falsch herum. Was
wichtig ist, braucht die Regel am dringendsten.

### 3.7 Eine Tagesgrenze im Durchgang, der keine Tage kennt

**Symptom** — Ein Test wurde rot: Der Ausgleich fand die faire Verteilung nicht
mehr (165 zu 55 statt 110 zu 110).

**Ursache** — Ich hatte die Kopflastgrenze auch in `feasible` geprüft, also im
Ausgleich. Der Ausgleich tauscht aber nur Zuständige und lässt den Tag stehen.
Eine Tagesgrenze in einem Durchgang zu prüfen, der den Tag nicht ändern kann,
nimmt Möglichkeiten weg, ohne je eine bessere zu finden.

**Lösung** — Die Grenze bleibt in der Zuteilung, die den Tag wählen kann, und
fällt im Ausgleich weg. Zwischen „eine Person trägt das Dreifache" und „ein
voller Montag" ist die Antwort nicht offen: Fairness ist das Versprechen, der
ruhige Montag ist die Kür.

**Lehre** — Dieselbe Familie wie 3.2, nur andersherum: Der Ausgleich ist die
Stelle, an der Regeln auseinanderfallen, weil er weniger weiß als die
Zuteilung.

### 3.8 Zwei Bäder als doppelte Dauer

**Symptom** — `Mia  Bad putzen  70 min`. Die Dreizehnjährige bekam die größte
Einzelaufgabe der Woche.

**Ursache** — Ich hatte die Bäderzahl als Faktor auf die Dauer gelegt. Das sah
aus wie Mathematik und war eine Behauptung darüber, wie Menschen putzen.

**Lösung** — Bäder sind eine **Anzahl**, kein Ausmaß: Zwei Bäder ergeben zwei
Aufgaben zu 35 Minuten, verteilbar auf zwei Tage und zwei Personen. Zimmer und
Personen bleiben ein Ausmaß — fünf Zimmer staubsaugen ist *eine* längere
Aufgabe.

**Dazu** — `MaxMinutesForChild` (45 Minuten): Das Alter in den Vorlagen sagt, ob
ein Kind eine Aufgabe *kann*, nicht wie groß ein einzelner Block sein darf.

**Lehre** — Wenn eine Zahl in der Ausgabe unangenehm aussieht, ist meistens
nicht die Rechnung falsch, sondern das Modell dahinter.

### 3.9 Ein Kinderzimmer wurde größer, wenn die Wohnung mehr Zimmer hat

**Symptom** — `Eigenes Zimmer aufräumen 42 min` in einer Fünfzimmerwohnung.

**Ursache** — Ich hatte `t-zimmer` in die Zimmer-Skalierung genommen. Es ist
die eine Vorlage, die sich auf *ein* Zimmer bezieht.

**Lehre** — Beim Kuratieren nach dem Titel zu greifen ist der kürzeste Weg zum
Fehler. „Zimmer" im Namen heißt nicht „skaliert mit Zimmern".

### 3.10 Eine Frage, die nichts bewirkte

**Symptom** — Keins. Gefunden durch die Frage: *Was bringt der Schalter
„Wohnung oder Haus" dem Nutzer?*

**Ursache** — Nichts. Keine einzige Vorlage wertete die Wohnform aus — die
letzte Bedingung daran hatte ich eine Stunde zuvor entfernt, weil sie falsch
herum war. Die Frage stand trotzdem als Pflichtschritt im Onboarding, als
Schalter in den Einstellungen und als Spalte in der Datenbank.

**Erschwerend** — Die Regel dagegen war längst aufgeschrieben und sogar
mechanisch durchgesetzt: Der Fakten-Loader weist eine Frage zurück, die nicht
sagt, was die Antwort bringt. Sie galt nur für die Fakten, nicht für das
Onboarding.

**Lösung** — Frage gestrichen. An ihre Stelle traten Zimmer und Bäder, die
Zahlen verändern statt Aufgaben ein- und auszublenden. Und ein Test hält die
Bibliothek jetzt gegen sich selbst: jedes Faktum wird gebraucht, jede
Voraussetzung ist fragbar, jede Anlassart eintragbar.

**Lehre** — Eine Regel, die nur für einen Teil des Systems mechanisch gilt,
wird im Rest verletzt. Der Test dafür war der erste im Projekt, der eine
Produktregel prüft statt Code.

### 3.11 Krippe vergessen

**Symptom** — Eine Zweijährige, die in die Kita geht, galt der App als
unbetreut. Damit fielen Kita-Tasche, Wechselkleidung und Elternbeitrag weg.

**Ursache** — Meine Faustregel lautete „Kindergarten ab drei". In Deutschland
gilt der Rechtsanspruch auf Betreuung ab dem vollendeten ersten Lebensjahr,
und Krippe ab eins ist der Normalfall.

**Lösung** — Kita ab 1, Schule ab 6. Und, wichtiger: Die Betreuungsform steht
jetzt in den Einstellungen. Eine bessere Faustregel liegt beim nächsten Kind
wieder daneben.

**Lehre** — **Ein Rateschluss, den man nicht korrigieren kann, ist eine
Behauptung.** Raten ist in Ordnung, solange man widersprechen kann.

### 3.12 Demo-Pläne verschieben sich nach `db-reset`

**Symptom** — Nach `make db-reset` sieht der Beispielplan anders aus als vorher,
obwohl sich am Code nichts geändert hat.

**Ursache** — Bei Gleichstand entscheidet die Mitglieds-UUID, und die wird beim
Import neu vergeben.

**Status** — Bekannt und bewusst offen. Behebbar mit rund 20 Zeilen
(deterministische Kennungen beim Import), lohnt sich, sobald Screenshots oder
Demos stabil sein müssen.

---

### 3.13 Eine Diagnose, die ich nicht nachgerechnet hatte

**Symptom** — Keins im Betrieb. In ADR-0011 stand unter „Offen geblieben", die
Dreizehnjährige sei am höchsten ausgelastet, weil der Ausgleich Aufgaben nur
ganz verschieben kann. Das klang plausibel, war sauber formuliert und
stand damit als Ursache fest.

**Ursache** — Ich hatte es nie gemessen. Vier Versuche später: Dreierzyklen
statt Paartausche ändern nichts, eine zusätzliche Strafe auf die
Minutenabweichung ändert nichts, ohne Kapazitätsprüfung wird es schlechter.
Schalte ich dagegen die Wochen-Rotation im Ausgleich ab, fällt das Kind von
79 auf 74 % und der Vater steigt von 67 auf 70.

**Der eigentliche Fehler** — Dieselbe Regel hatte zwei Stärken: Im `assign`
ist die Rotation eine Vorliebe (`rankMembers` sortiert die Person von letzter
Woche nur nach hinten), im `rebalance` war sie ein Veto. Die härtere Fassung
stand in dem Durchgang, der ausgleichen soll — und blockierte den Tausch
„Böden wischen gegen Wäsche zusammenlegen", der die Kennzahl um ein Viertel
verbessert hätte. Mia wischte die Böden, weil ihr Vater sie letzte Woche
gewischt hatte.

**Lösung** — Zweiter Durchgang im Ausgleich, der die Wochen-Rotation brechen
darf, solange jemand um mehr als `MaxOvershootPermille` seiner eigenen
Kapazität über seinem Anteil liegt (ADR-0012). Preis: 7 statt 5
Wiederholungen bei 52 Aufgaben.

**Lehre** — **Eine Vermutung, die man nicht nachrechnet, ist eine Behauptung
— auch in einem Architekturentscheid.** Ein ADR sieht aus wie ein Befund,
sobald es geschrieben ist. Der Abschnitt „Dagegen" in ADR-0003 nannte die
richtige Ursache übrigens schon am 7. September; es fehlte nur die Zahl
daneben und die Frage, wer die Rechnung bezahlt.

### 3.14 Eine Schnittstelle, drei Erfüllungen — zwei gefunden

**Symptom** — `go vet`: *\*library.Catalog does not implement httpapi.Plans
(missing method Reassign)*.

**Ursache** — `httpapi.Plans` hat drei Erfüllungen: den Speicher (echte
Datenbank), den Katalog (die Beispielhaushalte aus dem Repo) und die Attrappe
im Test. Beim Nachziehen hatte ich an Speicher und Attrappe gedacht und den
Katalog vergessen — er steht in einem anderen Paket und fällt beim Suchen
durchs Raster.

**Lösung** — `Reassign` gibt dort `ErrUnknownTask` zurück, wie `MarkDone` und
`HandOver` daneben: Die Wochen aus dem Repo werden gerechnet und nicht
geschrieben, es gibt keine Aufgabe, die man umverteilen könnte.

**Lehre** — Bei einer neuen Methode an einer Schnittstelle zuerst
`grep -rn "InterfaceName"` über das ganze Repo, nicht nur über das Paket, in
dem man gerade arbeitet. Der Übersetzer findet es sowieso — er findet es nur
später als man selbst könnte.

### 3.15 Der Ausgleich tauschte eine Absprache weg

**Symptom** — Ein Test, den ich als Gegenprobe geschrieben hatte, schlug fehl:
Das Bad landete bei der Person, die laut Absprache jeden Morgen zur Kita
fährt. Der Fehler saß nicht dort, wo er auffiel.

**Ursache** — `swapAllowed` kannte `braucht_absprache` nicht. Der zweite
Durchgang tauschte also den Montag an die andere Person, weil die Zahlen
dadurch aufgingen — und schrieb als Begründung „zum Ausgleich" daneben. Im
Plan stand damit jemand, der um 7:45 gar nicht dort sein kann, mitsamt einer
Behauptung über eine Vereinbarung, die niemand getroffen hatte. Genau der
Zustand, gegen den ADR-0016 antritt, erzeugt vom eigenen zweiten Durchgang.

**Lösung** — `ta.NeedsAgreement || tb.NeedsAgreement` verbietet den Tausch,
wie `DistFixed` und `PerPerson` daneben. Dasselbe in `Handover`: Abgeben sucht
die Person mit den meisten freien Minuten — die Größe, die hier nicht
entscheidet. Abgesprochenes steht danach offen da, und ein Mensch schließt die
Lücke.

**Lehre** — Eine neue Regel im Planer ist erst fertig, wenn **jeder**
Durchgang sie kennt, der Zuteilungen anfasst: `assign`, `rebalance`,
`Handover`, `Reassign`. Vier Stellen, und drei davon hatte ich beim Bauen der
vierten nicht im Kopf. Der Ort, an dem eine Regel entsteht, ist selten der
einzige, an dem sie gelten muss.

Und: Die Gegenprobe war es wert. Ich hatte sie fast weggelassen — „prüft doch
nur das Offensichtliche". Sie hat den einzigen echten Fehler der ganzen
Funktion gefunden, und zwar an einer Stelle, an der ich nicht gesucht hätte.

### 3.16 Der Vertrag kannte zwei eigene Antworten nicht

**Symptom** — TypeScript beim Bauen der Weboberfläche: *This comparison
appears to be unintentional … `"gilt_nicht" | … | "niemand_geeignet"` and
`"braucht_absprache"` have no overlap.*

**Ursache** — In `openapi.yaml` steht dasselbe Vokabular zweimal:
`VorlagenStand.grund` und `Uebersprungen.grund`. Die zweite Liste war
unvollständig — ohne `unbekannt`, ohne `braucht_termin` —, und zwar seit dem
Tag, an dem es diese Gründe gibt. Der Dienst hat beide Werte die ganze Zeit
ausgeliefert; der Vertrag behauptete, es könne sie nicht geben.

**Warum es niemandem auffiel** — Die Oberfläche hat die Gründe nur
*nachgeschlagen* (`grundText[u.grund] ?? u.grund`), und ein Nachschlagen in
`Record<string, string>` prüft nichts. Erst der erste *Vergleich* auf einen
dieser Werte brachte den Übersetzer dazu, die Liste ernst zu nehmen. Ein
falscher Vertrag, an dem nichts bricht, bleibt beliebig lange falsch.

**Lösung** — Die Liste vervollständigt, und zwei Tests in `internal/httpapi`
dagegengestellt: Jeder Wert aus `planner.AllSkipCodes` und
`planner.AllReasonCodes` muss in den erzeugten Aufzählungen `Valid()` sein.
`oapi-codegen` erzeugt diese Methode ohnehin — der Test kostet zwölf Zeilen
und hält drei Listen zusammen, die sonst nur Disziplin zusammenhält.

**Lehre** — Dieselbe Sorte Absicherung wie `git diff --exit-code` über die
erzeugten Dateien in der CI: Der Vertrag ist nur dann die eine Wahrheit, wenn
etwas nachrechnet, dass er es ist. Und der Reihe nach das dritte Mal dieselbe
Ursache in diesem Projekt — ein Wissen, das an einer Stelle entsteht und an
einer zweiten von Hand gepflegt wird (vgl. 3.15, und die geschlossenen Listen
im Schema unter 4.x).

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

### 4.3 Zwei geschlossene Vokabulare, eines nach dem anderen

**Symptom** — Umverteilen antwortete mit 500. Nach der Korrektur: wieder 500.

**Ursache** — `event.kind` ist durch einen CHECK auf sechs Wörter begrenzt,
„umverteilt" war keines davon. Migration geschrieben, angewendet — und danach
scheiterte dieselbe Anfrage an `assignment.reason_code`, der genauso begrenzt
ist und `von_hand` nicht kannte. Zwei Migrationen (00010, 00011) für eine
Funktion, weil ich nach dem ersten Fund nicht weitergesucht habe.

**Erschwerend** — `verschoben` stand bereits in der Liste und hätte gepasst.
Verworfen: Das Wort gehört der Verschiebung auf einen anderen *Tag*, die noch
kommt. Zwei verschiedene Rückmeldungen unter einem Namen sind später nicht
mehr zu trennen — und das Protokoll wird gelesen, um daraus zu lernen.

**Nicht gemacht** — 00010 nachträglich erweitern. Goose merkt sich die Nummer,
nicht den Inhalt; eine geänderte 00010 wäre lokal nie wieder gelaufen, auf Neon
aber in der neuen Fassung. Zwei Schemata, die beide „Stand 00010" heißen, sieht
man erst in der Produktion.

**Lehre** — **Wer ein neues Wort in die Datenbank schreibt, sucht vorher ALLE
Stellen, an denen Wörter begrenzt sind.** `grep -n "CHECK (.* IN" db/migrations`
dauert zehn Sekunden und hätte die zweite Migration gespart. Und: Eine neue
Ereignisart ist eine Schemaänderung, keine Codeänderung — sie gehört vor dem
Deploy nach Neon, sonst trifft es dort einen echten Haushalt.

### 4.4 Eine neue Funktion machte eine alte Zweideutigkeit sichtbar

**Symptom** — Keins. Gefunden beim Lesen von `storage.Plan`, bevor die
Blätterpfeile gebaut waren.

**Ursache** — `Plan` schreibt jede angesehene Woche fest. Die Regel aus
ADR-0008 heißt „beim ersten Ansehen", und solange es keinen Weg zu einer
anderen Woche gab, war das dasselbe wie „die laufende Woche". Mit Pfeilen
hätte der erste Klick auf „nächste Woche" sie eingefroren — jeder danach
eingetragene Anlass wäre dort nie angekommen. Und beim Zurückblättern hätte
die App Vergangenheit erfunden: Zuteilungen, die nie jemand hatte, und eine
Rotation, die daraus weiterrechnet.

**Lösung** — Festgeschrieben wird nur die laufende Woche, in der Zeitzone des
Haushalts. Künftige Wochen tragen ein Banner, das sagt, dass sie eine Vorschau
sind und warum dort nichts abzuhaken ist.

**Lehre** — **Eine Regel, die nie auf die Probe gestellt wurde, ist nicht
eindeutig — sie ist nur unangefochten.** Beim Bauen einer neuen Funktion lohnt
die Frage, welche bestehende Regel dadurch zum ersten Mal in einen Fall
gerät, für den sie nie geschrieben wurde.

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

### 5.3 `res.json()` auf eine 204

**Symptom** — Beim Abhaken:
`Failed to execute 'json' on 'Response': Unexpected end of JSON input`. Das
Abhaken hatte funktioniert, nur die Antwort war leer.

**Ursache** — `postMitToken` las immer JSON. Der Endpunkt antwortet mit **204 —
kein Inhalt**, weil es nichts zu sagen gibt. Beides für sich richtig, und
genau dazwischen fällt es durch: Zwei Generatoren erzeugen Typen aus derselben
Spezifikation, aber niemand erzeugt die Annahme „hier kommt ein Rumpf".

**Lösung** — Der Status wird geprüft, nicht geraten:

```ts
if (res.status === 204) return undefined as T;
```

**Lehre** — Ein Vertrag beschreibt auch die Antworten, die leer sind. Ein
Client, der jede Antwort gleich behandelt, hat den Vertrag nur zur Hälfte
gelesen — und der Nutzer sieht einen Fehler, obwohl alles geklappt hat.

### 5.4 Offene Weiterleitung in `?weiter=`

**Symptom** — Kein Fehler, sondern ein Fund beim Durchlesen: Der
Anmelde-Parameter `?weiter=` hätte auf eine fremde Domain zeigen können.

**Lösung** — Nur Pfade akzeptieren, und keine, die mit `//` beginnen.

---

### 5.5 Die Antwort war gespeichert, die Frage kam wieder

**Symptom** — „Wenn ich eine Frage beantworte, passiert nichts. Es flackert
kurz, dann steht die Frage wieder da."

**Ursache** — Die Antwort war korrekt gespeichert. Nur kommen die Fragen aus
`Result.Skipped`, und bei einer festgeschriebenen Woche ist das die
eingefrorene Momentaufnahme aus `week_plan.skipped` (ADR-0008). `SetFacts`
schreibt das Faktum an den Haushalt und rührt die Woche nicht an — genau wie
`UpdateHousehold`, wo dafür der Knopf „Diese Woche neu rechnen" danebensteht.
Die Fragen hatten keinen solchen Knopf. Das Flackern war `router.refresh()`,
das brav dasselbe neu lud.

**Lösung** — Die Karte klappt nach der Antwort zu einer Bestätigung zusammen
(„Notiert. … — ab nächster Woche."), und darunter steht „Schon diese Woche"
— nur, wenn mindestens ein Ja dabei war, denn ein Nein schaltet nichts frei.
Nicht stillschweigend neu rechnen: Wer auf „Habt ihr Pflanzen? — Ja" tippt,
rechnet nicht damit, dass sich der Samstag umsortiert.

**Nachtrag — der Fix war Kosmetik.** Die Karte klappte zu einer Bestätigung
zusammen, und das war ein Zustand im Browser. Wer die Seite neu lud, bekam
**dieselben zwei Fragen noch einmal** — denn gelesen wurden sie weiter aus
`r.Skipped`, und bei einer festgeschriebenen Woche ist das die eingefrorene
Liste von damals. Die Antwort ändert den Haushalt, nicht die Momentaufnahme.

Behoben durch `planner.OpenFacts`: Die offenen Fragen werden am **aktuellen**
Haushalt gerechnet und stehen als eigenes Feld `Result.Open` neben `Skipped`.
Die beiden meinen verschiedene Zeitpunkte — Skipped erklärt die Woche, wie sie
festgeschrieben wurde, Open beschreibt den Haushalt, wie er jetzt ist. Solange
das dasselbe war, fiel der Unterschied nicht auf.

**Lehre** — **Eine Antwort, die nichts sichtbar bewirkt, sieht aus wie eine
verlorene Antwort.** Der Fehler war nicht im Speichern, sondern im Schweigen
danach.

**Und die zweite, teurere:** Ich habe das Symptom in der Oberfläche behandelt
und die Ursache stehen gelassen — an derselben Stelle, an der ich drei Stunden
später schrieb, was der Server weiß, gehöre nicht nebenher in den Browser.
**Ein Fix, der nur beim Hinsehen hält, ist kein Fix, sondern eine Verzögerung.**

### 5.6 Zustand im Browser, der eine Server-Aktualisierung überlebt

**Symptom** — Nach „Schon diese Woche" kamen zwei neue Fragen. Beantwortet —
und der Knopf kam nicht wieder.

**Ursache** — `router.refresh()` tauscht die Daten aus und lässt die Komponente
stehen. Mein `gerechnet`-Flag war danach noch `true`, und der Knopf hängt an
`!gerechnet`. Zwei frische Fragen liefen in eine Komponente, die sich für
fertig hielt.

**Lösung** — Eine neue Antwort setzt `gerechnet` zurück, das Neurechnen leert
die gemerkten Antworten.

**Und dasselbe ein drittes Mal, eine Stunde später** — „Ich klicke in den
Einstellungen auf *mittel*, es passiert nichts." Es passierte etwas: Der Dienst
rechnete die Stufe in Minuten um und speicherte sie. Nur hielt `PersonTeil` die
Minuten in `useState`, und `useState` nimmt den Anfangswert genau einmal — die
Zeile „Zeit in der Woche" zeigte weiter die Zahl von vor dem Laden. Behoben,
indem der lokale Zustand nur noch einen *Entwurf* hält, solange die Minuten
aufgeklappt sind; zugeklappt kommt die Zahl aus dem Serverstand.

Der erste Versuch war ein wechselnder `key` auf der Komponente. Das funktioniert
und hätte einen getippten, noch nicht gespeicherten Namen mitgerissen, sobald
jemand daneben auf „mittel" tippt. Zurückgenommen: Die grobe Lösung war der
nächste Fehler.

**Lehre** — **Was der Server weiß, wird nicht nebenher im Browser
mitgeführt.** Lokaler Zustand darf einen Entwurf halten, solange jemand tippt,
und eine Bedienfrage beantworten („ist aufgeklappt") — aber keine Kopie
dessen, was gerade gespeichert wurde. Dreimal an einem Abend derselbe Fehler,
und jedes Mal sah es aus, als hätte die App die Eingabe verschluckt: Die Sache
war gespeichert, die Oberfläche zeigte einen Stand, den sie sich vorher
gemerkt hatte.

### 5.7 Die Navigation stand hinter dem Inhalt

**Symptom** — „Ich finde es mühsam, bis nach unten zu scrollen, um die Fragen
oder die Knöpfe zu sehen." Und danach: „Einstellungen, Aufgaben, Anlässe —
die sind immer noch ganz unten."

**Ursache** — Gewachsen, nicht entschieden. Die Bereichslinks waren als
Nachtrag unter den Wochenplan gekommen, als es drei Aufgaben und drei Links
gab. Bei zwei Dutzend Aufgaben mit je einer Knöpfeleiste darunter war die
Navigation der am schwersten erreichbare Teil der Oberfläche.

**Lösung** — Drei Dinge, alle in dieselbe Richtung: die Bereiche als Streifen
unter die Überschrift (auf allen fünf Seiten, er ersetzt dort den
„Zurück“-Link), die Fragen über die Woche statt darunter, und die
Aufgabenzeile klappt ihre Aktionen erst beim Antippen auf — sichtbar bleibt
nur das Häkchen, die häufigste Handlung und die Alternative zur Wischgeste.

**Lehre** — **Was unten steht, wurde nicht entschieden, sondern angehängt.**
Eine lange Hauptseite verschiebt jede spätere Ergänzung weiter aus dem Blick,
und niemand merkt es, solange man die Seite nur in der Entwicklung von oben
liest.

### 5.8 Beschriftung und Eingabefeld klebten aneinander

**Symptom** — „Zimmer" und das Eingabefeld standen ohne Abstand nebeneinander.

**Ursache** — `<label class="block space-y-1.5">` mit einem `<span>` und einem
`<input>` darin. Beide sind inline, stehen also auf derselben Zeile, und
`space-y` setzt einen *oberen* Abstand — der bewirkt in einer Zeile nichts. An
den vier anderen Stellen mit demselben Aufbau fällt es nicht auf, weil dort das
Eingabefeld selbst `block` trägt und ohnehin umbricht.

**Lösung** — `block` an die Beschriftung, an allen fünf Stellen.

**Lehre** — Ein Fehler, der nur an einer Stelle sichtbar ist, steht oft an
fünf. Wenn eine Klassenkombination falsch ist, lohnt `grep` nach genau dieser
Kombination, bevor man die eine Stelle repariert.

### 5.9 Eine Stufe, die nicht angezeigt wurde, weil ich sie nicht nachbauen wollte

**Symptom** — „Die Zeit wird angepasst, aber es wird nicht gezeigt, was gerade
ausgewählt ist."

**Ursache** — Absicht, und trotzdem falsch. Die Knöpfe „wenig / mittel / viel"
setzten eine Stufe, ohne eine anzuzeigen; im Code stand als Begründung, welche
Minuten „mittel" bedeutet, wisse nur der Planer, und das hier nachzubauen wäre
eine zweite Wahrheit. Der erste Teil stimmt. Der Schluss daraus war bequem:
Weil die Anzeige schwierig war, gab es keine.

**Lösung** — Der Planer beantwortet jetzt auch die Rückrichtung (`BudgetOf`:
Minuten → Stufe, leer wenn keine genau passt), die Stufe steht im Vertrag als
`Mitglied.zeit`, und die Oberfläche markiert sie. Passt keine, ist keine
markiert und daneben steht „eigene Minuten" — dann weiß man, warum nichts
leuchtet.

**Lehre** — **Nachbauen wäre eine zweite Wahrheit. Nachfragen ist keine.**
„Das gehört woanders hin" ist ein Argument gegen den Ort einer Rechnung, nie
eines dafür, das Ergebnis wegzulassen.

### 5.10 Drei Befunde, die erst auf dem Telefon sichtbar wurden

**Symptom** — Auf dem iPhone: Die Wochenpfeile waren schwer zu treffen, beim
Wechsel zwischen den Bereichen passierte ein bis zwei Sekunden lang nichts,
und das Datumsfeld lief aus der Karte und saß oben statt mittig.

**Ursachen**, drei verschiedene:

*Die Pfeile* hatte ich mit `size-9` gebaut — 36 Pixel. Die Regel im Projekt
sind 44, und ich habe sie gebrochen, weil es am Rechner kompakter aussah. Am
Rechner trifft eine Maus auch 20 Pixel.

*Die Ladeanzeige* fehlte ganz. Die Seiten sind Server Components: Beim
Antippen holt der Next.js-Server den Inhalt, und über Mobilfunk dauert das.
Dazu kam ein zweiter, feinerer Fehler — `router.refresh()` wurde nirgends
abgewartet, also stand der Knopf nach einer Aktion wieder bereit, während die
neuen Daten noch unterwegs waren.

*Das Datumsfeld* bekommt von Safari eine eigene, vom Inhalt abgeleitete Breite
(`w-full` wird ignoriert) und setzt seinen Wert nach oben statt mittig, weil
ein Datumsfeld anders als ein Textfeld nicht senkrecht zentriert.

**Lösung** — 44 Pixel; `useLinkStatus` für einen Kreisel *im angetippten
Chip* statt eines Balkens am Seitenrand, dazu `useTransition` um jedes
`router.refresh()`; `appearance: none` und `padding-block` für das
Datumsfeld.

**Lehre** — **Eine App für das Telefon wird auf dem Telefon geprüft, nicht im
schmalen Browserfenster.** Zielgrößen, Wartezeiten und die Eigenheiten
nativer Eingabefelder sind genau die drei Dinge, die am Schreibtisch
unsichtbar bleiben — und alle drei entscheiden darüber, ob jemand die App ein
zweites Mal öffnet.

Und als Fortsetzung von 5.6: Erst spiegelte die Oberfläche Serverwissen, dann
behauptete sie, fertig zu sein, bevor der Server es war. **Beide Male log der
Bildschirm über den Zustand.**

### 5.11 Das Etikett behauptete mehr, als die Sache tut

**Symptom** — „Ich wische *Safiya zur Kita bringen* am Montag weg. Das heißt
doch nicht, dass sie die ganze Woche nicht hingeht."

**Ursache** — Der Code war richtig, das Wort war falsch. Gestrichen wird ein
**Termin an einem Tag**; das Etikett hieß „Diese Woche nicht" und behauptete
damit die Reichweite einer ganzen Woche. Bei einer Vorlage, die fünfmal
vorkommt, ist das eine andere Aussage — und die falsche.

**Lösung** — „Diesmal nicht". Und die gestrichene Zeile trägt jetzt den Tag:
*Safiya zur Kita bringen · Montag, 15.09.* Ohne ihn wüsste niemand, welches von
fünf gemeint ist.

**Lehre** — **Die Reichweite einer Handlung gehört in ihr Etikett, und zwar
genau.** Derselbe Fehler wie bei den ausgegrauten Zeilen und bei der
Vorschau-Woche: Die Oberfläche sagt etwas über die Welt, das der Code nicht
deckt. Nur fällt es hier erst auf, wenn die Vorlage mehrfach vorkommt — bei
„Bad putzen" wäre „diese Woche" zufällig richtig gewesen.

### 5.12 „Gilt bei euch nicht" — und dann?

**Symptom** — „Wie schalte ich eine Aufgabe frei, die gerade nicht gilt? Ich
gehe auf Einstellungen und finde nichts."

**Ursachen**, zwei, und beide waren Andeutung statt Auskunft:

*Ein verneintes Faktum war eine Sackgasse.* `Status` gab das Faktum nur bei
„unbekannt" zurück, nicht bei „nein". Die Oberfläche hatte also keine Frage,
die sie noch einmal hätte stellen können: Ein Nein galt für immer. Das ist
derselbe Fehler wie die geratene Betreuungsform aus ADR-0007 — **eine Antwort,
die man nicht zurücknehmen kann, ist eine Behauptung.**

*Die Kontextbedingung wurde nicht genannt.* Die Liste sagte „gilt bei euch
nicht" und dazu eine Aufzählung aller denkbaren Ursachen — Garten, Auto,
Haustiere, Betreuung. Wer „Brotdose vorbereiten" freischalten wollte, las die
Liste und wusste hinterher genauso wenig. Der Planer prüft die Bedingung
ohnehin; er gab sie nur nicht heraus.

**Lösung** — Das Faktum kommt jetzt auch bei „nein" zurück, und die Frage
steht mit einer Zeile darüber wieder da („Ihr habt das mit Nein
beantwortet."). `conditionsMet` liefert die fehlende Bedingung als Satz:
*Braucht ein Kind, das zur Schule geht.*

**Lehre** — **Wer eine Bedingung prüft, kann sie auch nennen.** Ein `bool`
zurückzugeben, wo man den Grund kennt, wirft Auskunft weg, die schon berechnet
ist. Dritte Auflage desselben Musters an einem Tag: erst die grauen Zeilen,
dann „gilt bei euch nicht", jetzt die fehlende Bedingung.

### 5.13 Die Vorabfrage antwortete mit 204 — und ohne PUT

**Symptom** — „Failed to fetch" beim Speichern der Absprache. Im Protokoll des
Dienstes stand genau eine Zeile:

```
level=INFO msg=anfrage methode=OPTIONS pfad=/api/haushalte/…/absprachen/… status=204
```

Also: Vorabfrage beantwortet, alles in Ordnung — und danach nichts mehr.

**Ursache** — `Access-Control-Allow-Methods` stand als Text im Code:
`"GET, POST, PATCH, DELETE, OPTIONS"`. Die Absprache ist die erste PUT-Route
im ganzen Vertrag. Der Browser fragt, bekommt eine Liste ohne PUT, und
schickt die eigentliche Anfrage **gar nicht erst los**. Deshalb steht sie
nirgends im Protokoll.

**Lösung** — PUT ergänzt, die Liste in eine Konstante gehoben, und ein Test
liest `openapi.yaml` und prüft, dass jede dort vorkommende Methode darin
steht.

**Lehre** — Das Tückische ist die Form des Fehlers: Die Antwort, die man
sieht, ist richtig. 204 ist der erwartete Status der Vorabfrage, und wer nur
auf den Status schaut, sucht als Nächstes an der falschen Stelle. Bei „Failed
to fetch" ohne zugehörige Zeile im Protokoll ist die Reihenfolge deshalb:
erst die **Kopfzeilen** der Vorabfrage ansehen, dann alles andere.

Und inhaltlich zum vierten Mal an einem Tag dieselbe Ursache: eine Liste, die
etwas über den Vertrag behauptet und von Hand gepflegt wird (vgl. 3.15, 3.16).

**Nachtrag** — `Access-Control-Max-Age: 86400`. Der Browser merkt sich die
Antwort der Vorabfrage einen Tag lang. Nach der Korrektur kann derselbe Aufruf
also weiter scheitern, obwohl der Dienst richtig antwortet — ein privates
Fenster oder geleerte Caches klären, ob es noch der alte Eintrag ist.

### 5.14 Gespeichert, und dann passierte nichts

**Symptom** — Das Wochenraster für „Von der Kita abholen" für Mo–Fr gesetzt,
„Übernehmen" gedrückt, Meldung „übernommen" — und im Plan stand nichts. Keine
Auskunft, ob es diese Woche noch kommt, ob nächste Woche, ob überhaupt.

**Ursache** — Kein Fehler im Code: Die laufende Woche steht fest, sobald sie
das erste Mal angesehen wurde (ADR-0008). Eine neue Absprache ändert sie also
nicht rückwirkend, sondern gilt ab der nächsten. Das ist richtig — und die
Oberfläche hat es verschwiegen.

**Lösung** — Nach dem Speichern steht jetzt da, was passiert ist („ab nächster
Woche von selbst — diese Woche steht schon fest"), und daneben derselbe Knopf,
den die Fragen schon haben: **Schon diese Woche**. Er ruft dieselbe
Neurechnung auf, die Abgehaktes und Abgegebenes stehen lässt.

**Lehre** — Dieselbe Sorte Lücke wie bei den grauen Zeilen und bei „gilt bei
euch nicht", nur andersherum: Dort behauptete die Oberfläche mehr, als die
Sache trägt, hier hat sie gar nichts gesagt. Beides ist dasselbe Versäumnis —
**eine Handlung ohne sichtbare Wirkung braucht einen Satz, der die Wirkung
nennt.** Und wenn die Wirkung „erst später" ist, gehört der Weg dazu, sie
sofort zu bekommen. Ein Bestätigungswort allein („übernommen") ist keine
Auskunft, sondern eine Quittung.

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

### 6.4 Der `release_command` startete den Dienst statt der Migration

**Symptom** — Deploy bricht nach 5:21 Minuten ab:

```
Running haushalt-api release_command: /migrate
> Waiting for 28692edc409318 to have state: destroyed
✖ Failed: timeout reached waiting for machine's state to change
```

Die Maschine startete, wurde aber nie fertig. Kein Fehler, keine Ausgabe der
Migration, nur ein Zeitüberlauf.

**Ursache** — Eine Zeile im Protokoll der Maschine, und sie sagt alles:

```
INFO Preparing to run: `/server /migrate` as nonroot
```

Das Dockerfile endete auf `ENTRYPOINT ["/server"]`. Fly übergibt den
`release_command` als **COMMAND** — und ein COMMAND wird an ein ENTRYPOINT
**angehängt**, nicht dafür eingesetzt. Aus `/migrate` wurde also
`/server /migrate`: Die Release-Maschine startete den HTTP-Dienst, hörte auf
Port 8080 und hatte keinen Grund, sich jemals zu beenden. Fly wartete auf
`destroyed`, bis die Geduld aufgebraucht war.

**Lösung** — `CMD ["/server"]` statt `ENTRYPOINT`. Damit ersetzt der
`release_command` den Befehl vollständig, und ohne Befehl startet weiterhin
der Dienst.

Zweitens, unabhängig davon: `cmd/server` bricht jetzt mit Exit-Code 2 ab, wenn
ihm ein Argument übergeben wird, das es nicht kennt. Der Dienst hat `/migrate`
kommentarlos geschluckt — und das war der Unterschied zwischen einer Meldung
nach einer Sekunde und einem Zeitüberlauf nach fünf Minuten.

**Lehre** — **Ein Programm, das unbekannte Argumente ignoriert, verschweigt
den Aufrufer-Fehler.** Es fühlt sich nachsichtig an und ist es nicht: Die
Nachsicht verlegt den Fehler von der Stelle, an der er passiert, an eine
Stelle fünf Minuten später, an der er nicht mehr zu erkennen ist.

Und: Bei einem Zeitüberlauf ohne Fehlermeldung lautet die erste Frage nicht
„was ist schiefgelaufen", sondern **„was ist überhaupt gelaufen"**. Die
Antwort stand in der ersten Protokollzeile der Maschine, nicht in der letzten.

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
