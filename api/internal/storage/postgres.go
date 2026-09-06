// Package storage kapselt die Datenbankverbindung.
//
// An T0 macht es nur eines: die Verbindung aufbauen und beantworten, ob sie
// steht. Tabellen und Abfragen kommen an T5 dazu.
package storage

import (
	"context"
	"database/sql"
	"fmt"
	"time"
)

// DB umschließt den Verbindungspool aus der Standardbibliothek.
//
// sql.DB ist kein einzelner Verbindungs-Handle, sondern ein Pool. Man baut ihn
// einmal beim Start und reicht ihn herum — nicht pro Anfrage einen neuen.
type DB struct {
	*sql.DB
}

// Open baut den Pool auf und prüft ihn sofort.
//
// sql.Open verbindet nichts — es prüft nur den Verbindungsstring. Erst der Ping
// stellt wirklich eine Verbindung her. Wer das nicht weiß, wundert sich, warum
// ein falsches Passwort erst bei der ersten Abfrage auffällt.
//
// Der Treiber wird nicht hier eingebunden, sondern per leerem Import in
// cmd/server/main.go registriert:
//
//	go get github.com/jackc/pgx/v5
//	import _ "github.com/jackc/pgx/v5/stdlib"
func Open(ctx context.Context, dsn string) (*DB, error) {
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		return nil, fmt.Errorf("storage: verbindungsstring: %w", err)
	}

	// Neon schließt inaktive Verbindungen. Ohne Obergrenzen sammelt der Pool
	// tote Verbindungen an, und die erste Abfrage nach einer Pause schlägt fehl.
	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(30 * time.Minute)
	db.SetConnMaxIdleTime(5 * time.Minute)

	// 15 Sekunden, nicht 5: Neon fährt inaktive Datenbanken herunter, der
	// erste Verbindungsaufbau danach weckt sie und dauert länger.
	pingCtx, cancel := context.WithTimeout(ctx, 15*time.Second)
	defer cancel()

	if err := db.PingContext(pingCtx); err != nil {
		// Aufräumen, bevor der Fehler nach oben geht — sonst bleibt ein
		// halboffener Pool zurück.
		_ = db.Close()
		return nil, fmt.Errorf("storage: datenbank nicht erreichbar: %w", err)
	}
	return &DB{db}, nil
}
