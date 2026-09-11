// Kommando server ist der HTTP-Dienst.
//
// main macht genau vier Dinge: Konfiguration lesen, Abhängigkeiten aufbauen,
// Server starten, sauber beenden. Keine Geschäftslogik, keine Routen — die
// stehen in internal/httpapi.
package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/zakaria/haushalt/api/internal/config"
	"github.com/zakaria/haushalt/api/internal/httpapi"
	"github.com/zakaria/haushalt/api/internal/library"
	"github.com/zakaria/haushalt/api/internal/storage"
)

// version wird beim Bauen gesetzt:
//
//	go build -ldflags="-X main.version=$(git rev-parse --short HEAD)"
//
// Damit sagt /healthz dir, welcher Stand gerade läuft. Klingt nach Kleinigkeit,
// erspart aber jedes Mal die Frage „ist mein Deployment eigentlich durch?".
var version = "dev"

// main bleibt bewusst dünn: Es ruft run auf und übersetzt einen Fehler in
// einen Exit-Code. In Go kann man os.Exit nicht mit defer kombinieren — deshalb
// diese Trennung, sonst laufen Aufräumarbeiten nicht.
func main() {
	if err := run(); err != nil {
		slog.Error("dienst beendet", "fehler", err)
		os.Exit(1)
	}
}

func run() error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}

	log := newLogger(cfg)
	log.Info("starte", "version", version, "env", cfg.Env, "port", cfg.Port)

	// signal.NotifyContext gibt einen Kontext, der bei SIGTERM abgebrochen
	// wird. Fly schickt SIGTERM vor jedem neuen Deployment; ohne das brechen
	// laufende Anfragen mitten im Satz ab.
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGTERM, syscall.SIGINT)
	defer stop()

	// Datenbank verbinden.
	//
	// In Produktion ist eine nicht erreichbare Datenbank ein Startfehler — ein
	// Dienst, der ohne sie hochkommt, liefert stillschweigend falsche Antworten.
	// In der Entwicklung startet er trotzdem: Wenn das Netz gerade Port 5432
	// blockiert, willst du weiter an der Oberfläche arbeiten können. /healthz
	// sagt in dem Fall die Wahrheit und antwortet mit 503.
	var (
		db   httpapi.Pinger
		pool *storage.DB
	)
	switch {
	case cfg.DatabaseURL == "":
		log.Warn("DATABASE_URL ist leer — dienst läuft ohne datenbank")
	default:
		p, err := storage.Open(ctx, cfg.DatabaseURL)
		switch {
		case err == nil:
			defer p.Close()
			pool, db = p, p
			log.Info("datenbank verbunden")
		case cfg.IsDevelopment():
			log.Warn("datenbank nicht erreichbar — dienst startet trotzdem", "fehler", err)
			db = unavailableDB{err}
		default:
			return err
		}
	}

	// Woher die Wochenpläne kommen.
	//
	// Mit Datenbank aus der Datenbank, sonst aus dem Repo. Beide erfüllen
	// dieselbe Schnittstelle, die HTTP-Schicht sieht keinen Unterschied — und
	// solange die Beispielhaushalte importiert sind, sieht auch der Nutzer
	// keinen.
	//
	// Der Rückfall auf die Dateien ist kein Notbehelf, sondern praktisch: Er
	// hält den Dienst ohne Datenbank lauffähig, und die Bibliothek im Repo
	// bleibt die Quelle, aus der importiert wird.
	var plans httpapi.Plans
	if pool != nil {
		plans = pool.AsPlans()
		log.Info("wochenpläne kommen aus der datenbank")
	} else {
		catalog, err := library.LoadCatalog(cfg.LibraryDir)
		if err != nil {
			return err
		}
		plans = catalog
		haushalte, _ := catalog.Households(ctx)
		log.Info("wochenpläne kommen aus dem repo",
			"verzeichnis", cfg.LibraryDir,
			"vorlagen", len(catalog.Templates()),
			"haushalte", len(haushalte))
	}

	srv := &http.Server{
		Addr:    ":" + cfg.Port,
		Handler: httpapi.New(cfg, log, db, version, plans).Handler(),

		// Ohne Zeitgrenzen kann ein einziger langsamer Client eine Verbindung
		// dauerhaft belegen. http.ListenAndServe ohne diese Werte ist der
		// häufigste Fehler in Go-Anleitungen.
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       15 * time.Second,
		WriteTimeout:      30 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	// Der Server läuft in einer eigenen Goroutine, damit main auf das
	// Abbruchsignal warten kann. Ein Fehler beim Zuhören kommt über den Kanal
	// zurück — ein Kanal ist hier schlicht ein Briefkasten für genau einen Wert.
	serverErr := make(chan error, 1)
	go func() {
		log.Info("höre zu", "adresse", srv.Addr)
		// ErrServerClosed ist kein Fehler, sondern die normale Folge von
		// Shutdown. Nur alles andere zählt.
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			serverErr <- err
			return
		}
		serverErr <- nil
	}()

	// select wartet auf das, was zuerst passiert: ein Serverfehler oder das
	// Abbruchsignal.
	select {
	case err := <-serverErr:
		return err
	case <-ctx.Done():
		log.Info("signal empfangen, fahre herunter")
	}

	// Laufende Anfragen bekommen noch 20 Sekunden. Danach wird abgeschnitten.
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		return err
	}
	log.Info("sauber beendet")
	return nil
}

// newLogger schreibt lokal lesbaren Text und in Produktion JSON — das können
// Log-Systeme auswerten.
func newLogger(cfg config.Config) *slog.Logger {
	opts := &slog.HandlerOptions{Level: slog.LevelInfo}
	if cfg.IsDevelopment() {
		opts.Level = slog.LevelDebug
		return slog.New(slog.NewTextHandler(os.Stdout, opts))
	}
	return slog.New(slog.NewJSONHandler(os.Stdout, opts))
}

// unavailableDB steht für eine Datenbank, zu der beim Start keine Verbindung
// zustande kam. Sie erfüllt httpapi.Pinger und gibt denselben Fehler zurück,
// damit /healthz "nicht erreichbar" meldet statt "nicht konfiguriert".
type unavailableDB struct{ err error }

func (u unavailableDB) Ping(context.Context) error { return u.err }
