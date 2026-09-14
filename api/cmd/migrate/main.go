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
package main

import (
	"database/sql"
	"flag"
	"log"
	"os"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"
)

func main() {
	dir := flag.String("dir", "/migrations", "Verzeichnis mit den .sql-Dateien")
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
}
