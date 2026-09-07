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
	"github.com/zakaria/haushalt/api/internal/httpapi/openapi"
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
	// plans liefert die Wochenpläne. Heute aus dem Repo, ab T5 aus der
	// Datenbank.
	plans Plans
}

func New(cfg config.Config, log *slog.Logger, db Pinger, version string, plans Plans) *Server {
	return &Server{cfg: cfg, log: log, db: db, version: version, plans: plans}
}

// Handler baut den Router und legt die Middleware darum.
//
// Seit Go 1.22 kann der ServeMux aus der Standardbibliothek Methoden und
// Pfad-Parameter: "GET /api/haushalte/{id}". Ältere Anleitungen empfehlen
// deshalb oft chi oder gorilla/mux — für diesen Umfang braucht es das nicht.
func (s *Server) Handler() http.Handler {
	mux := http.NewServeMux()

	// /healthz gehört der Plattform, nicht dem Vertrag: Fly ruft es auf, kein
	// Client. Deshalb steht es nicht in openapi.yaml und wird hier von Hand
	// eingehängt.
	mux.HandleFunc("GET /healthz", s.handleHealth)

	// Alles andere kommt aus der Spezifikation. NewStrictHandler übersetzt
	// zwischen den erzeugten Antworttypen und dem ResponseWriter; die beiden
	// Fehlerfunktionen sorgen dafür, dass auch Fehler als JSON herauskommen —
	// die Vorgabe des Generators schreibt reinen Text und würde den Vertrag
	// brechen.
	strict := openapi.NewStrictHandlerWithOptions(
		api{plans: s.plans, version: s.version},
		nil,
		openapi.StrictHTTPServerOptions{
			RequestErrorHandlerFunc:  s.badRequest,
			ResponseErrorHandlerFunc: s.serverError,
		},
	)
	handler := openapi.HandlerWithOptions(strict, openapi.StdHTTPServerOptions{
		BaseRouter:       mux,
		ErrorHandlerFunc: s.badRequest,
	})

	// Die Reihenfolge ist von außen nach innen zu lesen: Eine Anfrage läuft
	// erst durch recoverPanic, dann durch logging, dann durch cors, dann in
	// den Mux.
	return recoverPanic(s.log)(
		logging(s.log)(
			cors(s.cfg.AllowedOrigins)(handler),
		),
	)
}

// badRequest beantwortet, was der Client falsch gemacht hat.
func (s *Server) badRequest(w http.ResponseWriter, _ *http.Request, err error) {
	writeJSON(w, s.log, http.StatusBadRequest, openapi.Fehler{Fehler: err.Error()})
}

// serverError beantwortet, was wir falsch gemacht haben.
//
// Der Fehlertext geht ins Log, nicht an den Client: Interne Meldungen können
// Pfade, Abfragen oder Namen enthalten, die niemanden draußen etwas angehen.
func (s *Server) serverError(w http.ResponseWriter, r *http.Request, err error) {
	s.log.Error("anfrage fehlgeschlagen", "pfad", r.URL.Path, "fehler", err)
	writeJSON(w, s.log, http.StatusInternalServerError,
		openapi.Fehler{Fehler: "im dienst ist ein fehler aufgetreten"})
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
