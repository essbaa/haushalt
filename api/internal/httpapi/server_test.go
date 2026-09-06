package httpapi

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/zakaria/haushalt/api/internal/config"
)

// fakePinger ist die Attrappe für die Datenbank. Sie erfüllt das Pinger-
// Interface, weil sie die eine Methode hat — mehr braucht Go nicht. Kein
// "implements", keine Registrierung.
type fakePinger struct{ err error }

func (f fakePinger) PingContext(context.Context) error { return f.err }

func testServer(db Pinger) *Server {
	cfg := config.Config{
		Env:            config.EnvDevelopment,
		Port:           "8080",
		AllowedOrigins: []string{"http://localhost:3000"},
	}
	// Logs im Test ins Nichts schreiben, sonst rauscht die Ausgabe voll.
	log := slog.New(slog.NewTextHandler(io.Discard, nil))
	return New(cfg, log, db, "test")
}

func TestHealth(t *testing.T) {
	tests := []struct {
		name       string
		db         Pinger
		wantStatus int
		wantDB     string
	}{
		{
			name:       "ohne Datenbank laeuft der Dienst trotzdem",
			db:         nil,
			wantStatus: http.StatusOK,
			wantDB:     "nicht konfiguriert",
		},
		{
			name:       "mit erreichbarer Datenbank",
			db:         fakePinger{},
			wantStatus: http.StatusOK,
			wantDB:     "erreichbar",
		},
		{
			name:       "kaputte Datenbank ergibt 503",
			db:         fakePinger{err: errors.New("verbindung abgelehnt")},
			wantStatus: http.StatusServiceUnavailable,
			wantDB:     "nicht erreichbar",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// httptest ruft den Handler direkt auf — kein echter Server, kein
			// Port, keine Wartezeit. Deshalb laufen Go-HTTP-Tests in
			// Millisekunden.
			req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
			rec := httptest.NewRecorder()

			testServer(tc.db).Handler().ServeHTTP(rec, req)

			if rec.Code != tc.wantStatus {
				t.Fatalf("Status = %d, erwartet %d", rec.Code, tc.wantStatus)
			}

			var got healthResponse
			if err := json.NewDecoder(rec.Body).Decode(&got); err != nil {
				t.Fatalf("Antwort ist kein gültiges JSON: %v", err)
			}
			if got.Database != tc.wantDB {
				t.Errorf("database = %q, erwartet %q", got.Database, tc.wantDB)
			}
		})
	}
}

func TestCORS(t *testing.T) {
	tests := []struct {
		name       string
		origin     string
		wantHeader string
	}{
		{"erlaubte Adresse bekommt den Header", "http://localhost:3000", "http://localhost:3000"},
		{"fremde Adresse bekommt ihn nicht", "https://boese.example", ""},
		{"ohne Origin passiert nichts", "", ""},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
			if tc.origin != "" {
				req.Header.Set("Origin", tc.origin)
			}
			rec := httptest.NewRecorder()

			testServer(nil).Handler().ServeHTTP(rec, req)

			if got := rec.Header().Get("Access-Control-Allow-Origin"); got != tc.wantHeader {
				t.Errorf("Allow-Origin = %q, erwartet %q", got, tc.wantHeader)
			}
		})
	}
}

func TestPreflight(t *testing.T) {
	req := httptest.NewRequest(http.MethodOptions, "/api/version", nil)
	req.Header.Set("Origin", "http://localhost:3000")
	rec := httptest.NewRecorder()

	testServer(nil).Handler().ServeHTTP(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Fatalf("Vorabfrage = %d, erwartet %d", rec.Code, http.StatusNoContent)
	}
}

func TestUnbekannterPfad(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/gibtsnicht", nil)
	rec := httptest.NewRecorder()

	testServer(nil).Handler().ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("Status = %d, erwartet 404", rec.Code)
	}
}
