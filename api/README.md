# api — der Go-Dienst

Alles aus dem Block „Go-Dienst" der T0-Checkliste, lauffähig.

```bash
cp .env.example .env
make run          # startet auf :8080
make health       # {"status":"ok", …}
make test
make docker-run   # baut das Image und startet es
```

Der Vertrag liegt eine Ebene höher: [`../openapi.yaml`](../openapi.yaml).
Er ist die Wahrheit — Typen und Server-Rümpfe entstehen daraus:

```bash
go get -tool github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen@latest  # einmalig
make generate     # erzeugt internal/httpapi/openapi/openapi.gen.go
```

Der erzeugte Code liegt im Repo, damit ein Bau keinen Generator braucht. Dass
er zur Spezifikation passt, prüft die CI mit `git diff --exit-code`.

```
api/
├─ cmd/
│  ├─ server/main.go       Start, Herunterfahren, Verdrahtung — sonst nichts
│  └─ plan/main.go         Kommandozeilenwerkzeug für den Planer
├─ internal/
│  ├─ config/              Umgebungsvariablen lesen und prüfen
│  ├─ httpapi/             Routen, Middleware, JSON
│  ├─ storage/             Datenbankverbindung
│  ├─ planner/             der reine Kern
│  └─ library/             Vorlagen einlesen
├─ library/                die Vorlagen als Daten
├─ Dockerfile
└─ Makefile
```

`internal/` ist keine Konvention, sondern eine Sprachregel: Pakete darunter
kann niemand von außerhalb dieses Moduls importieren. Praktisch, wenn das
Repo später öffentlich ist.

## Acht Dinge, die aus der React-Welt ungewohnt sind

**1. Es gibt kein Framework.** `net/http` in der Standardbibliothek *ist* der
Webserver. Kein Express, kein Nest. Und seit Go 1.22 kann der `ServeMux`
Methoden und Pfad-Parameter — `mux.HandleFunc("GET /api/haushalte/{id}", …)`.
Ältere Anleitungen empfehlen deshalb chi oder gorilla/mux; für diesen Umfang
brauchst du beides nicht.

**2. Fehler sind Rückgabewerte, keine Ausnahmen.** `if err != nil { return … }`
steht überall, und das ist gewollt: Jede Fehlerquelle ist im Code sichtbar,
statt irgendwo im Aufrufstapel abgefangen zu werden. `fmt.Errorf("…: %w", err)`
verpackt einen Fehler mit Kontext, `errors.Is` prüft ihn später wieder aus.

**3. Middleware ist eine Funktion, die einen Handler nimmt und einen Handler
zurückgibt.** Genau das Muster, das du als Higher-Order-Komponente kennst.
Siehe `middleware.go` — kein Register, keine Reihenfolgen-Magie, man
verschachtelt sie einfach.

**4. `context.Context` wird durchgereicht.** Er trägt Abbruch und Zeitlimit.
Praktisch heißt das: `r.Context()` in den Handler geben, weiter an die
Datenbank, und wenn der Browser die Verbindung schließt, bricht die Abfrage
mit ab. Faustregel: Ist `ctx` der erste Parameter einer Funktion, ist das kein
Zufall, sondern die Konvention.

**5. Interfaces werden dort definiert, wo sie gebraucht werden.** `Pinger` in
`server.go` hat eine Methode und steht im HTTP-Paket, nicht im Datenbank-Paket.
Der Test übergibt eine Attrappe, die Produktion die echte Datenbank, und keiner
von beiden muss vom anderen wissen. Kein `implements`, keine Registrierung —
wer die Methode hat, erfüllt das Interface.

**6. Einbetten statt Vererben.** `statusRecorder` in `middleware.go` legt sich
um einen `http.ResponseWriter`, übernimmt alle seine Methoden und überschreibt
nur die eine, die interessiert.

**7. Eine Panik beendet den ganzen Prozess**, nicht nur den einen Aufruf —
anders als eine Ausnahme in JavaScript. Deshalb die `recoverPanic`-Middleware.
Und deshalb benutzt man `panic` nicht für normale Fehler.

**8. Goroutinen und Kanäle sind hier nur zwei Zeilen.** In `main.go` läuft der
Server in einer eigenen Goroutine, damit die Hauptfunktion auf das
Abbruchsignal warten kann. Der Kanal ist dabei nichts weiter als ein
Briefkasten für genau einen Wert.

## Fallen, in die man an T0 tappt

- **Zeitgrenzen am Server.** `http.ListenAndServe(":8080", handler)` ohne
  `ReadHeaderTimeout` und Co. steht in fast jeder Anleitung und ist falsch: Ein
  einziger langsamer Client belegt sonst dauerhaft eine Verbindung. Siehe die
  `http.Server`-Struktur in `main.go`.
- **`os.Exit` und `defer` vertragen sich nicht.** Deshalb ist `main` dünn und
  die eigentliche Arbeit steht in `run() error` — sonst laufen die
  Aufräumarbeiten nie.
- **`sql.Open` verbindet nichts.** Es prüft nur den Verbindungsstring. Erst
  `PingContext` stellt wirklich eine Verbindung her; ohne das fällt ein
  falsches Passwort erst bei der ersten Abfrage auf.
- **Reihenfolge beim Antworten:** erst Header setzen, dann `WriteHeader`, dann
  schreiben. Vertauscht bekommt der Client den falschen Statuscode und du ein
  „superfluous WriteHeader call" im Log.
- **`PORT` aus der Umgebung lesen.** Fly gibt den Port vor. Ein fest
  verdrahtetes 8080 ergibt ein Deployment, das erfolgreich aussieht und nicht
  antwortet.

## Datenbank anschließen (Checklisten-Punkt „Go verbindet sich beim Start")

```bash
go get github.com/jackc/pgx/v5
```

Dann in `cmd/server/main.go` den auskommentierten leeren Import aktivieren:

```go
_ "github.com/jackc/pgx/v5/stdlib"
```

Ein leerer Import lädt ein Paket nur wegen seiner Nebenwirkung — hier
registriert es den Treiber unter dem Namen `"pgx"`, den `storage.Open`
benutzt. Danach `DATABASE_URL` in `.env` setzen; `/healthz` meldet ab dann
`"database": "erreichbar"` und antwortet mit 503, wenn die Verbindung steht.

## Abnahme für diesen Block

```bash
make test                                   # grün
make run                                    # läuft
curl -i localhost:8080/healthz              # 200 + JSON
curl -i -H "Origin: http://localhost:3000" localhost:8080/healthz   # Allow-Origin gesetzt
curl -i -H "Origin: https://fremd.example" localhost:8080/healthz   # kein Allow-Origin
make docker-run                             # läuft auch im Container
```

Wenn es im Container läuft, liegt ein Fehler in der Cloud an der Konfiguration
— nicht am Code. Das ist der Grund für diesen letzten Schritt.
