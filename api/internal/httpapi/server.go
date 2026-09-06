// Package httpapi ist die HTTP-Schicht: Routen, Middleware, JSON.
//
// Hier steht keine Geschäftslogik. Der Planer weiß nichts von HTTP, und dieses
// Paket weiß nichts davon, wie ein Wochenplan zustande kommt.
package httpapi

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/zakaria/haushalt/api/internal/config"
)

// Pinger ist alles, was sagen kann, ob es erreichbar ist.
//
// Ein Interface mit genau einer Methode, definiert dort, wo es *gebraucht*
// wird — nicht dort, wo es implementiert wird. Das ist in Go die Regel und der
// Grund, warum man hier selten Abhängigkeitsinjektion braucht: Der Test übergibt
// eine Attrappe, die Produktion die echte Datenbank, und keiner der beiden muss
// vom anderen wissen.
type Pinger interface {
	PingContext(ctx context.Context) error
}

type Server struct {
	cfg config.Config
	log *slog.Logger
	// db darf nil sein. An T0 gibt es noch keine Datenbank, und der Dienst
	// soll trotzdem laufen und das ehrlich melden.
	db Pinger
	// version wird beim Bauen hineingereicht (siehe Makefile).
	version string
}

func New(cfg config.Config, log *slog.Logger, db Pinger, version string) *Server {
	return &Server{cfg: cfg, log: log, db: db, version: version}
}

// Handler baut den Router und legt die Middleware darum.
//
// Seit Go 1.22 kann der ServeMux aus der Standardbibliothek Methoden und
// Pfad-Parameter: "GET /api/haushalte/{id}". Ältere Anleitungen empfehlen
// deshalb oft chi oder gorilla/mux — für diesen Umfang braucht es das nicht.
func (s *Server) Handler() http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /healthz", s.handleHealth)
	mux.HandleFunc("GET /api/version", s.handleVersion)

	// Die Reihenfolge ist von außen nach innen zu lesen: Eine Anfrage läuft
	// erst durch recoverPanic, dann durch logging, dann durch cors, dann in
	// den Mux.
	return recoverPanic(s.log)(
		logging(s.log)(
			cors(s.cfg.AllowedOrigins)(mux),
		),
	)
}

type healthResponse struct {
	Status   string `json:"status"`
	Version  string `json:"version"`
	Env      string `json:"env"`
	Database string `json:"database"`
}

// handleHealth sagt die Wahrheit, auch wenn sie unangenehm ist.
//
// Ein Health-Endpunkt, der immer "ok" antwortet, ist wertlos. Dieser meldet,
// ob die Datenbank wirklich erreichbar ist — und antwortet mit 503, wenn nicht,
// damit Fly einen kaputten Container nicht in den Verkehr nimmt.
func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	res := healthResponse{
		Status:   "ok",
		Version:  s.version,
		Env:      s.cfg.Env,
		Database: "nicht konfiguriert",
	}
	code := http.StatusOK

	if s.db != nil {
		// r.Context() wird abgebrochen, wenn der Client die Verbindung
		// schließt. Den Kontext weiterzureichen ist in Go die Regel, nicht
		// die Ausnahme.
		if err := s.db.PingContext(r.Context()); err != nil {
			s.log.Error("datenbank nicht erreichbar", "fehler", err)
			res.Status = "degraded"
			res.Database = "nicht erreichbar"
			code = http.StatusServiceUnavailable
		} else {
			res.Database = "erreichbar"
		}
	}

	writeJSON(w, s.log, code, res)
}

func (s *Server) handleVersion(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, s.log, http.StatusOK, map[string]string{"version": s.version})
}

// writeJSON ist der einzige Ort, an dem eine Antwort geschrieben wird.
//
// Wichtig ist die Reihenfolge: erst Header setzen, dann WriteHeader, dann der
// Rumpf. Wer das vertauscht, bekommt ein "superfluous WriteHeader call" im Log
// und einen falschen Statuscode beim Client.
func writeJSON(w http.ResponseWriter, log *slog.Logger, code int, payload any) {
	body, err := json.Marshal(payload)
	if err != nil {
		log.Error("antwort konnte nicht serialisiert werden", "fehler", err)
		http.Error(w, `{"error":"interner fehler"}`, http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(code)
	if _, err := w.Write(body); err != nil {
		// Der Client ist weg. Nichts, was man noch reparieren könnte.
		log.Debug("antwort konnte nicht gesendet werden", "fehler", err)
	}
}
