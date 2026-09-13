# Mehrstufig: gebaut wird in einem vollen Go-Image, ausgeliefert wird ein
# Zielimage, das nur die fertige Binärdatei enthält. Ergebnis: rund 15 MB statt
# rund 900 MB — und ein Container ohne Shell, in dem es wenig zu übernehmen gibt.

# Die Go-Version kommt aus der go.mod und wird beim Bauen hereingereicht
# (siehe Makefile und .github/workflows/api.yml).
#
# Vorher stand sie hier fest, und zweimal ist genau das passiert: go.mod zog
# auf eine neue Version, der Bau lief mit der alten weiter und brach mit
# "go.mod requires go >= X" ab — beim Deployen, nicht beim Testen. Eine Regel,
# an die man sich erinnern muss, ist keine Regel.
ARG GO_VERSION=1.26
FROM golang:${GO_VERSION}-alpine AS build
WORKDIR /src

# Erst nur die Modul-Dateien kopieren und die Abhängigkeiten holen. Dieser
# Schritt wird zwischengespeichert und läuft nur neu, wenn sich go.mod ändert —
# nicht bei jeder Codeänderung.
COPY go.mod go.sum* ./
RUN go mod download

COPY . .

# CGO_ENABLED=0 baut ohne C-Bibliotheken, damit die Binärdatei in einem Image
# ohne libc läuft. -trimpath entfernt lokale Pfade, -s -w die Debug-Symbole.
ARG VERSION=dev
RUN CGO_ENABLED=0 GOOS=linux go build \
    -trimpath \
    -ldflags="-s -w -X main.version=${VERSION}" \
    -o /out/server ./cmd/server

# distroless: kein Paketmanager, keine Shell, kein root.
FROM gcr.io/distroless/static-debian12:nonroot
WORKDIR /
COPY --from=build /out/server /server

# Die Vorlagen und die Beispielhaushalte sind Daten, keine Binärdatei — sie
# müssen mit ins Image. WORKDIR ist /, deshalb findet der Dienst sie unter dem
# Vorgabepfad "library" ohne weitere Konfiguration.
COPY --from=build /src/library /library

# Nur Dokumentation — Fly nutzt die PORT-Variable, nicht diese Zeile.
EXPOSE 8080
USER nonroot:nonroot
ENTRYPOINT ["/server"]
