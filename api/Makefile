# Lädt api/.env in die Umgebung, falls vorhanden. So braucht der Go-Code keine
# .env-Bibliothek und verhält sich lokal exakt wie bei Fly.
ifneq (,$(wildcard .env))
include .env
export
endif

VERSION := $(shell git rev-parse --short HEAD 2>/dev/null || echo dev)

.PHONY: run
run: ## Dienst lokal starten
	go run -ldflags="-X main.version=$(VERSION)" ./cmd/server

.PHONY: plan
plan: ## Wochenplan auf der Konsole ausgeben
	go run ./cmd/plan --woche 2026-W38 --uebersprungen

.PHONY: test
test: ## Alle Tests
	go test ./...

.PHONY: check
check: ## Was auch die CI prüft
	gofmt -l . | tee /dev/stderr | (! read)
	go vet ./...
	go test -race ./...

.PHONY: docker-build
docker-build:
	docker build --build-arg VERSION=$(VERSION) -t haushalt-api .

.PHONY: docker-run
docker-run: docker-build ## Image bauen und starten, Umgebung aus .env
	docker run --rm -p 8080:8080 --env-file .env haushalt-api

.PHONY: health
health: ## Gegen den laufenden Dienst prüfen
	curl -sS localhost:8080/healthz | jq .

.PHONY: deploy
deploy: ## Nach Fly deployen, mit Commit-Kennung in der Binärdatei
	fly deploy --build-arg VERSION=$(VERSION)
