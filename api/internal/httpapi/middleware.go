package httpapi

import (
	"log/slog"
	"net/http"
	"slices"
	"time"
)

// middleware ist eine Funktion, die einen Handler nimmt und einen Handler
// zurückgibt.
//
// Wenn dir das aus React bekannt vorkommt: Es ist dasselbe Muster wie eine
// Higher-Order-Komponente. Kein Framework nötig, kein Register, keine
// Reihenfolgen-Magie — man verschachtelt sie einfach.
type middleware func(http.Handler) http.Handler

// logging schreibt eine Zeile pro Anfrage: Methode, Pfad, Status, Dauer.
func logging(log *slog.Logger) middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			rec := &statusRecorder{ResponseWriter: w, status: http.StatusOK}

			next.ServeHTTP(rec, r)

			log.Info("anfrage",
				"methode", r.Method,
				"pfad", r.URL.Path,
				"status", rec.status,
				"dauer_ms", time.Since(start).Milliseconds(),
			)
		})
	}
}

// statusRecorder merkt sich den Statuscode.
//
// http.ResponseWriter verrät nicht, welcher Status geschrieben wurde. Man legt
// deshalb eine eigene Struktur darum, die die Schnittstelle einbettet und nur
// die eine Methode überschreibt, die einen interessiert. Das ist in Go der
// übliche Weg — Einbetten statt Vererben.
type statusRecorder struct {
	http.ResponseWriter
	status int
}

func (r *statusRecorder) WriteHeader(code int) {
	r.status = code
	r.ResponseWriter.WriteHeader(code)
}

// recoverPanic fängt einen Absturz in einem Handler ab.
//
// Ohne das reißt eine einzelne kaputte Anfrage den ganzen Prozess mit. Go
// beendet bei einer Panik das Programm, nicht nur die Goroutine — anders als
// eine Ausnahme in JavaScript, die nur den einen Aufruf zerlegt.
func recoverPanic(log *slog.Logger) middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if rec := recover(); rec != nil {
					log.Error("panik im handler", "pfad", r.URL.Path, "grund", rec)
					http.Error(w, `{"error":"interner fehler"}`, http.StatusInternalServerError)
				}
			}()
			next.ServeHTTP(w, r)
		})
	}
}

// cors erlaubt genau die Adressen aus der Konfiguration.
//
// Der erste echte Fehler zwischen den beiden Diensten kommt fast immer von
// hier. Zwei Dinge sind wichtig: Die erlaubte Adresse steht in einer Variablen,
// nicht im Code — du wirst sie mehrfach ändern. Und ein "*" ist nur so lange
// harmlos, wie keine Anmeldedaten mitgeschickt werden; sobald Cookies oder
// Tokens im Spiel sind, lehnt der Browser "*" ab.
func cors(allowed []string) middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			origin := r.Header.Get("Origin")

			if origin != "" && slices.Contains(allowed, origin) {
				w.Header().Set("Access-Control-Allow-Origin", origin)
				w.Header().Set("Access-Control-Allow-Credentials", "true")
				w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PATCH, DELETE, OPTIONS")
				w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
				w.Header().Set("Access-Control-Max-Age", "86400")
				// Antworten unterscheiden sich je nach Origin — sonst liefert
				// ein Zwischenspeicher die Antwort für die falsche Adresse aus.
				w.Header().Add("Vary", "Origin")
			}

			// Der Browser fragt vor der eigentlichen Anfrage per OPTIONS nach.
			// Diese Vorabfrage darf nicht in den Router laufen.
			if r.Method == http.MethodOptions {
				w.WriteHeader(http.StatusNoContent)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
