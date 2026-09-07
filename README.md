# haushalt

**Web:** https://haushalt-omega.vercel.app · **API:** https://haushalt-api.fly.dev/healthz

Ein Wochenplaner für Haushalte, der nicht fragt „wer macht das?", sondern einen
Vorschlag macht und ihn begründet. Er verteilt wiederkehrende Aufgaben über zwei
Achsen — **Dauer** in Minuten und **Kopflast** von 0 bis 3 — weil die eigentliche
Last selten im Spülen steckt, sondern im Daran-denken. Wer plant, wer ausführt
und wer nur beteiligt ist, sind getrennte Rollen. Aufgaben rotieren, bevor sie
ausgeglichen werden.

Funktioniert für eine Person allein ebenso wie für einen Haushalt mit
Jugendlichen, die selbst Aufgaben übernehmen.

Status: **T0 abgeschlossen** — beide Dienste sind live, ein Push auf `main`
aktualisiert sie. Inhaltlich beginnt es jetzt: der Planer-Kern.

## Aufbau

```
api/     Go 1.25, net/http — der Dienst und der Planer-Kern
web/     Next.js 16, React 19, Tailwind 4 — die Oberfläche
docs/    Entscheidungen (ADR)
```

Ein Repo, zwei Dienste, zwei getrennte Auslieferungswege: Die CI filtert nach
Pfad, eine Änderung an `web/` startet keine Go-Tests und löst kein
Fly-Deployment aus. Warum überhaupt zwei Dienste, steht in
[ADR-0001](docs/adr/0001-go-neben-nextjs.md).

Das Herzstück ist `api/internal/planner` — ein reines Paket ohne I/O: Eingabe
rein, Wochenplan raus, keine Datenbank, keine Uhr, keine Zufallszahlen.
Dadurch ist der Kern ohne Infrastruktur testbar, und jede nicht eingeplante
Vorlage kommt mit einem maschinenlesbaren Grund zurück statt kommentarlos zu
verschwinden.

## Lokal starten

```bash
# Backend
cd api && cp .env.example .env    # DATABASE_URL eintragen, oder leer lassen
make run                          # :8080
make health                       # {"status":"ok", …}
make check                        # gofmt, vet, test -race — dasselbe wie die CI

# Frontend, in einer zweiten Sitzung
cd web && cp .env.example .env.local
npm install && npm run dev        # :3000
```

Ohne `DATABASE_URL` läuft der Dienst und meldet in `/healthz`
`"database": "nicht konfiguriert"`. Ist eine URL gesetzt, aber die Datenbank
nicht erreichbar, startet er im Entwicklungsmodus trotzdem und antwortet mit
503 — in Produktion bricht er ab.

Der Wochenplaner lässt sich auch ohne laufenden Dienst ausprobieren:

```bash
cd api && make plan
```

## Betrieb

| | |
|---|---|
| API | Fly.io, Region `fra`, distroless-Image (~15 MB) |
| Web | Vercel, Wurzelverzeichnis `web/` |
| Datenbank | Neon PostgreSQL, Frankfurt |
| CI | GitHub Actions — Tests immer, Deploy nur auf `main` nach grünen Tests |

`/healthz` gibt die Commit-Kennung des laufenden Stands zurück. Die Frage „ist
mein Deployment eigentlich durch?" beantwortet damit der Dienst und nicht das
Bauchgefühl.

## Entscheidungen

Nicht offensichtliche Entscheidungen stehen als [ADR](docs/adr/) im Repo —
mit Kontext, Nachteilen und den verworfenen Alternativen.
