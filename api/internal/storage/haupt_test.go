package storage

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"testing"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"
)

// Tests gegen eine echte Datenbank — und der Grund dafür ist unangenehm
// konkret.
//
// Der Planer ist gründlich geprüft: Tabellentests, reine Funktionen, -race.
// Diese Schicht hier hatte **keinen einzigen Test**, und jeder teure Fehler
// dieses Projekts saß genau in ihr:
//
//   - Beantwortete Fragen kamen wieder. Drei Tage, drei falsche Hypothesen.
//     Ursache: fünf Haushalte, und die Startseite nahm haushalte[0], den
//     ältesten (Stolperstelle 5.15).
//   - Nach einem Deploy fehlten in der Produktion die Kita-Vorlagen. Die
//     Migrationen liefen, der Import nicht.
//
// Beides wäre hier aufgefallen. Attrappen hätten es nicht gefunden: Der
// Fehler steckte jeweils in dem, was die Datenbank wirklich zurückgibt —
// Reihenfolge, Sichtbarkeit, was importiert wurde —, und eine Attrappe gibt
// zurück, was der Autor des Tests erwartet hat.
//
// Ohne TEST_DATABASE_URL wird übersprungen statt zu scheitern. Ein Test, der
// auf einem frisch geklonten Rechner rot ist, wird abgeschaltet und nicht
// repariert.
var testDB *DB

func TestMain(m *testing.M) {
	url := os.Getenv("TEST_DATABASE_URL")
	if url == "" {
		fmt.Fprintln(os.Stderr, "storage: TEST_DATABASE_URL fehlt — Tests übersprungen")
		os.Exit(0)
	}

	if err := vorbereiten(url); err != nil {
		fmt.Fprintln(os.Stderr, "storage:", err)
		os.Exit(1)
	}
	code := m.Run()
	testDB.Close()
	os.Exit(code)
}

func vorbereiten(url string) error {
	// Migrationen über database/sql, wie in cmd/migrate: goose spricht diese
	// Schnittstelle. Derselbe Weg wie in der Produktion — ein Schema, das der
	// Test anders aufbaut als der Betrieb, prüft den Betrieb nicht.
	sqlDB, err := sql.Open("pgx", url)
	if err != nil {
		return fmt.Errorf("verbindung: %w", err)
	}
	defer func() { _ = sqlDB.Close() }()

	if err := sqlDB.Ping(); err != nil {
		return fmt.Errorf("datenbank nicht erreichbar: %w", err)
	}
	if err := goose.SetDialect("postgres"); err != nil {
		return err
	}
	if err := goose.Up(sqlDB, "../../db/migrations"); err != nil {
		return fmt.Errorf("migrationen: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	testDB, err = Open(ctx, url)
	if err != nil {
		return err
	}

	// Dieselbe Bibliothek wie im Betrieb, über denselben Weg. Damit prüft
	// jeder Lauf nebenbei, dass vorlagen.json überhaupt importierbar ist.
	if _, err := testDB.ImportLibrary(ctx, "../../library/vorlagen.json"); err != nil {
		return fmt.Errorf("bibliothek: %w", err)
	}
	return nil
}

// leeren räumt zwischen den Tests auf.
//
// TRUNCATE ... CASCADE statt DELETE: Es setzt auch alles zurück, was per
// Fremdschlüssel daran hängt, und darauf soll sich kein Test verlassen
// müssen. task_template bleibt stehen, weil dort die kuratierte Bibliothek
// liegt — die gehört keinem Haushalt und wird einmal importiert.
func leeren(t *testing.T) {
	t.Helper()
	ctx := t.Context()
	if _, err := testDB.Pool.Exec(ctx,
		`TRUNCATE household, member, task_instance, assignment, event, signal,
		          invitation, week_plan, occasion, agreement, feedback,
		          template_weekday CASCADE`); err != nil {
		t.Fatalf("leeren: %v", err)
	}
	// Eigene Vorlagen gehören einem Haushalt und müssen mit ihm verschwinden;
	// die kuratierten bleiben.
	if _, err := testDB.Pool.Exec(ctx,
		`DELETE FROM task_template WHERE household_id IS NOT NULL`); err != nil {
		t.Fatalf("leeren: %v", err)
	}
}
