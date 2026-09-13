# Lädt api/.env in die Umgebung, falls vorhanden. So braucht der Go-Code keine
# .env-Bibliothek und verhält sich lokal exakt wie bei Fly.
ifneq (,$(wildcard .env))
include .env
export
endif

# Der Go-Proxy wird hier unbedingt gesetzt, nicht nur als Vorgabe.
#
# Auf diesem Rechner steht in den Shell-Profilen noch der Athens-Proxy des
# früheren Arbeitgebers, und eine echte Umgebungsvariable schlägt `go env -w`.
# Ein Repo, das sich nur bauen lässt, wenn vorher jemand von Hand die richtige
# Variable exportiert hat, ist kaputt — auch wenn die Ursache woanders liegt.
GOPROXY := https://proxy.golang.org,direct
export GOPROXY

VERSION := $(shell git rev-parse --short HEAD 2>/dev/null || echo dev)

# Die Go-Version aus der go.mod, auf Haupt- und Nebennummer gekürzt: aus
# "go 1.26.0" wird "1.26". Damit gibt es eine Wahrheit statt zweier, die
# auseinanderlaufen können.
GO_VERSION := $(shell awk '/^go /{split($$2,v,"."); print v[1]"."v[2]; exit}' go.mod)

.PHONY: run
run: ## Dienst lokal starten
	go run -ldflags="-X main.version=$(VERSION)" ./cmd/server

# Beispielhaushalt für `make plan`. Überschreibbar:
#   make plan HAUSHALT=library/beispiele/familie-b.json
HAUSHALT ?= library/beispiele/familie-a.json

.PHONY: plan
plan: ## Wochenplan auf der Konsole ausgeben
	go run ./cmd/plan --haushalt $(HAUSHALT) --woche 2026-W38 --uebersprungen

.PHONY: test
test: ## Alle Tests
	go test ./...

.PHONY: generate
generate: ## Typen und Server-Rümpfe aus ../openapi.yaml erzeugen
	go tool oapi-codegen -config oapi.yaml ../openapi.yaml

.PHONY: import
import: ## Bibliothek und Beispielhaushalte in die Datenbank schreiben
	go run ./cmd/import

.PHONY: migrate-up
migrate-up: ## Migrationen einspielen
	go tool goose -dir db/migrations postgres "$(DATABASE_URL)" up

.PHONY: migrate-down
migrate-down: ## Letzte Migration zurücknehmen
	go tool goose -dir db/migrations postgres "$(DATABASE_URL)" down

.PHONY: migrate-status
migrate-status: ## Was ist eingespielt
	go tool goose -dir db/migrations postgres "$(DATABASE_URL)" status

.PHONY: sqlc
sqlc: ## Typsichere Abfragen aus db/query erzeugen
	go tool sqlc generate

.PHONY: verify-generate
verify-generate: generate ## Prüfen, dass der erzeugte Code zur Spezifikation passt
	git diff --exit-code -- internal/httpapi/openapi

.PHONY: check
check: ## Was auch die CI prüft
	gofmt -l . | tee /dev/stderr | (! read)
	go vet ./...
	go test -race ./...

# --wait wartet, bis der Health-Check in der compose.yaml grün ist.
#
# Ohne das kehrt `docker compose up -d` zurück, sobald der Container läuft —
# Postgres nimmt da noch keine Verbindungen an, und der nächste Befehl
# scheitert an etwas, das nach einem Fehler im Programm aussieht.
.PHONY: db-up
db-up: ## Entwicklungsdatenbank starten (Postgres im Container)
	docker compose up -d --wait db

.PHONY: db-down
db-down: ## Entwicklungsdatenbank anhalten — die Daten bleiben im Volume
	docker compose down

.PHONY: db-reset
db-reset: ## Entwicklungsdatenbank wegwerfen, neu anlegen, Migrationen und Beispieldaten
	docker compose down -v
	docker compose up -d --wait db
	$(MAKE) migrate-up
	$(MAKE) import

.PHONY: db-shell
db-shell: ## psql in der Entwicklungsdatenbank
	docker compose exec db psql -U haushalt -d haushalt

.PHONY: docker-build
docker-build:
	docker build --build-arg VERSION=$(VERSION) --build-arg GO_VERSION=$(GO_VERSION) -t haushalt-api .

.PHONY: docker-run
docker-run: docker-build ## Image bauen und starten, Umgebung aus .env
	docker run --rm -p 8080:8080 --env-file .env haushalt-api

.PHONY: health
health: ## Gegen den laufenden Dienst prüfen
	curl -sS localhost:8080/healthz | jq .

.PHONY: deploy
deploy: ## Nach Fly deployen, mit Commit-Kennung in der Binärdatei
	fly deploy --build-arg VERSION=$(VERSION) --build-arg GO_VERSION=$(GO_VERSION)
