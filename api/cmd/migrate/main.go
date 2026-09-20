// Command migrate spielt die Migrationen ein und beendet sich.
//
// Warum eine eigene Binärdatei und nicht ein Schritt im Dienst: Ein Dienst,
// der beim Start migriert, migriert bei jedem Neustart und bei jeder zweiten
// Maschine gleichzeitig. Und warum nicht von Hand: Die CI deployt beim Push,
// die Migration machte bisher ein Mensch hinterher — dazwischen lief in der
// Produktion ein neuer Dienst vor einem alten Schema. Genau der Zustand, der
// lokal dreimal ein 500 war.
//
// Fly ruft das hier als `release_command` auf: in einer eigenen Maschine, mit
// denselben Secrets, VOR dem Umschalten auf die neue Version. Scheitert es,
// geht die neue Version gar nicht erst live.
//
// Die Regel wandert damit aus dem Kopf in die Konfiguration — eine Regel, an
// die man sich erinnern muss, ist keine Regel.
//
// Eine Einschränkung, die man kennen muss: Der Befehl läuft, während die ALTE
// Version noch bedient. Migrationen müssen deshalb zum alten Schema passen —
// hinzufügen ja, wegnehmen erst im übernächsten Schritt.
//
// Seit dem 20. September spielt er außerdem die **Vorlagen-Bibliothek** ein.
// Aus demselben Grund: Sie lag als Datei im Repo, kam aber nur dann in die
// Produktion, wenn ein Mensch nach dem Deploy `make import` mit der
// Produktions-URL laufen ließ. Einmal vergessen, und in der Produktion standen
// die Migrationen, während die Kita-Vorlagen fehlten — sichtbar erst daran,
// dass ein Haushalt einen halb leeren Plan bekam. Die Bibliothek ist das
// Produkt; sie gehört in den Release und nicht in eine Shell.
//
// Der Import ist idempotent (Upsert je Vorlage) und darf deshalb bei jedem
// Deploy laufen. Die Beispielhaushalte bleiben draußen: Die sind Saatgut fürs
// lokale Arbeiten, nicht Teil des Produkts. Dafür gibt es weiterhin
// `make import`.
package main

import (
	"context"
	"database/sql"
	"flag"
	"log"
	"os"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"

	"github.com/zakaria/haushalt/api/internal/storage"
)

func main() {
	dir := flag.String("dir", "/migrations", "Verzeichnis mit den .sql-Dateien")
	bibliothek := flag.String("bibliothek", "/library/vorlagen.json",
		"Vorlagen-Bibliothek; leer überspringt den Import")
	flag.Parse()

	url := os.Getenv("DATABASE_URL")
	if url == "" {
		log.Fatal("migrate: DATABASE_URL fehlt")
	}

	db, err := sql.Open("pgx", url)
	if err != nil {
		log.Fatalf("migrate: verbindung: %v", err)
	}
	defer func() { _ = db.Close() }()

	if err := db.Ping(); err != nil {
		log.Fatalf("migrate: datenbank nicht erreichbar: %v", err)
	}
	if err := goose.SetDialect("postgres"); err != nil {
		log.Fatalf("migrate: %v", err)
	}
	if err := goose.Up(db, *dir); err != nil {
		log.Fatalf("migrate: %v", err)
	}

	if *bibliothek == "" {
		return
	}

	// Erst nach den Migrationen: Eine neue Vorlage kann eine Spalte brauchen,
	// die es vorher nicht gab.
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	store, err := storage.Open(ctx, url)
	if err != nil {
		log.Fatalf("migrate: bibliothek: %v", err)
	}
	defer store.Close()

	vorlagen, err := store.ImportLibrary(ctx, *bibliothek)
	if err != nil {
		log.Fatalf("migrate: bibliothek: %v", err)
	}
	log.Printf("migrate: %d Vorlagen eingelesen", len(vorlagen))
}
