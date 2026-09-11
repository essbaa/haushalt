// Package storage kapselt die Datenbankverbindung.
package storage

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/zakaria/haushalt/api/internal/storage/db"
)

// DB ist der Verbindungspool samt der erzeugten Abfragen.
//
// Eingebettet, nicht als Feld: Damit liegen alle sqlc-Funktionen direkt auf
// *DB — storage.Queries wäre ein Umweg, den niemand braucht.
type DB struct {
	*pgxpool.Pool
	*db.Queries
}

// Open baut den Pool auf und prüft ihn sofort.
//
// pgxpool statt database/sql: Die von sqlc erzeugten Funktionen sprechen
// pgx-Signaturen (pgx.Rows statt sql.Rows), und der Umweg über den
// stdlib-Adapter würde genau die Typen kosten, wegen derer wir sqlc benutzen —
// etwa Postgres-Felder wie int[] oder jsonb ohne Handarbeit.
//
// ParseConfig statt New(ctx, dsn): Nur so lassen sich die Poolgrenzen setzen,
// bevor die erste Verbindung entsteht.
func Open(ctx context.Context, dsn string) (*DB, error) {
	cfg, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		return nil, fmt.Errorf("storage: verbindungsstring: %w", err)
	}

	// Neon schließt inaktive Verbindungen. Ohne Obergrenzen sammelt der Pool
	// tote Verbindungen an, und die erste Abfrage nach einer Pause schlägt fehl.
	cfg.MaxConns = 10
	cfg.MinConns = 0
	cfg.MaxConnLifetime = 30 * time.Minute
	cfg.MaxConnIdleTime = 5 * time.Minute

	pool, err := pgxpool.NewWithConfig(ctx, cfg)
	if err != nil {
		return nil, fmt.Errorf("storage: pool: %w", err)
	}

	// 15 Sekunden, nicht 5: Neon fährt inaktive Datenbanken herunter, der
	// erste Verbindungsaufbau danach weckt sie und dauert länger.
	pingCtx, cancel := context.WithTimeout(ctx, 15*time.Second)
	defer cancel()

	if err := pool.Ping(pingCtx); err != nil {
		// Aufräumen, bevor der Fehler nach oben geht — sonst bleibt ein
		// halboffener Pool zurück.
		pool.Close()
		return nil, fmt.Errorf("storage: datenbank nicht erreichbar: %w", err)
	}

	return &DB{Pool: pool, Queries: db.New(pool)}, nil
}
