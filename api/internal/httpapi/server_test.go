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
	"github.com/zakaria/haushalt/api/internal/planner"
)

// Die vier Methoden unten braucht die Attrappe nur, um die Schnittstelle zu
// erfüllen. Einladungen und Rollen haben eigene Tests gegen die Datenbank.

// fakePinger ist die Attrappe für die Datenbank. Sie erfüllt das Pinger-
// Interface, weil sie die eine Methode hat — mehr braucht Go nicht. Kein
// "implements", keine Registrierung.
type fakePinger struct{ err error }

func (f fakePinger) Ping(context.Context) error { return f.err }

// fakePlans ist die Attrappe für die Planquelle. Ein Haushalt, ein Plan, kein
// Dateisystem — die HTTP-Schicht wird geprüft, nicht der Planer.
type fakePlans struct{}

func (fakePlans) Create(context.Context, string, planner.Setup) (planner.Household, error) {
	return planner.Household{}, planner.ErrNotAllowed
}

func (fakePlans) RoleOf(context.Context, string, string) (planner.Role, error) { return "", nil }

func (fakePlans) Invite(context.Context, string, string, planner.Role, string) (planner.Invitation, error) {
	return planner.Invitation{}, planner.ErrNotAllowed
}

func (fakePlans) Accept(context.Context, string, string, string) (planner.Household, error) {
	return planner.Household{}, planner.ErrUnknownInvitation
}

func (fakePlans) Households(context.Context, string) ([]planner.Household, error) {
	return []planner.Household{testHaushalt()}, nil
}

func (fakePlans) Plan(_ context.Context, _, id string, week planner.Week) (planner.Result, planner.Household, error) {
	if id != "familie-a" {
		return planner.Result{}, planner.Household{}, planner.ErrUnknownHousehold
	}
	return planner.Result{
		Week: week,
		Tasks: []planner.PlannedTask{{
			TemplateID: "t-bad", Title: "Bad putzen",
			Category: planner.CatCleaning, Kind: planner.KindDo,
			Day: planner.MustDate("2026-09-19"), Slot: planner.SlotAny,
			DurationMin: 35, AssigneeID: "m-ben",
			Reason: planner.Reason{Code: planner.ReasonRotation, Previous: "m-anna"},
		}},
		Balance: []planner.MemberLoad{
			{MemberID: "m-anna", Minutes: 60, Weighted: 60, Capacity: 600, Tasks: 1},
			{MemberID: "m-ben", Minutes: 35, Weighted: 35, Capacity: 600, Tasks: 1},
		},
	}, testHaushalt(), nil
}

func testHaushalt() planner.Household {
	return planner.Household{
		ID: "familie-a", Name: "Familie A",
		Members: []planner.Member{
			{ID: "m-anna", Name: "Anna", Role: planner.RolePlanner, Age: 36},
			{ID: "m-ben", Name: "Ben", Role: planner.RolePlanner, Age: 38},
		},
	}
}

func testServer(db Pinger) *Server {
	cfg := config.Config{
		Env:            config.EnvDevelopment,
		Port:           "8080",
		AllowedOrigins: []string{"http://localhost:3000"},
	}
	// Logs im Test ins Nichts schreiben, sonst rauscht die Ausgabe voll.
	log := slog.New(slog.NewTextHandler(io.Discard, nil))
	return New(cfg, log, db, "test", fakePlans{}, nil)
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

// ---------------------------------------------------------- Vertrag

func TestWochenplan(t *testing.T) {
	tests := []struct {
		name string
		pfad string
		code int
	}{
		{"vorhandener Haushalt", "/api/haushalte/familie-a/plan/2026-W38", http.StatusOK},
		{"unbekannter Haushalt", "/api/haushalte/familie-z/plan/2026-W38", http.StatusNotFound},
		{"unlesbare Woche", "/api/haushalte/familie-a/plan/2026-W99", http.StatusBadRequest},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			rec := httptest.NewRecorder()
			testServer(nil).Handler().ServeHTTP(rec, httptest.NewRequest(http.MethodGet, tc.pfad, nil))

			if rec.Code != tc.code {
				t.Fatalf("Status %d, erwartet %d — Rumpf: %s", rec.Code, tc.code, rec.Body.String())
			}
			if got := rec.Header().Get("Content-Type"); got != "application/json" {
				t.Errorf("Content-Type %q, erwartet application/json", got)
			}
			// Auch Fehler sind Teil des Vertrags und müssen JSON sein.
			var beliebig map[string]any
			if err := json.Unmarshal(rec.Body.Bytes(), &beliebig); err != nil {
				t.Fatalf("Antwort ist kein JSON-Objekt: %v — %s", err, rec.Body.String())
			}
		})
	}
}

// TestWochenplanFelder prüft die Übersetzung: Was der Planer liefert, muss
// unter den Namen aus openapi.yaml herauskommen. Deshalb wird hier gegen die
// rohen JSON-Felder geprüft und nicht gegen die erzeugten Go-Typen — sonst
// würde der Test dieselbe Annahme benutzen, die er absichern soll.
func TestWochenplanFelder(t *testing.T) {
	rec := httptest.NewRecorder()
	testServer(nil).Handler().ServeHTTP(rec,
		httptest.NewRequest(http.MethodGet, "/api/haushalte/familie-a/plan/2026-W38", nil))

	var plan struct {
		Woche    string `json:"woche"`
		Haushalt struct {
			Id         string `json:"id"`
			Name       string `json:"name"`
			Mitglieder []struct {
				Id    string `json:"id"`
				Rolle string `json:"rolle"`
			} `json:"mitglieder"`
		} `json:"haushalt"`
		Aufgaben []struct {
			VorlageId   string `json:"vorlage_id"`
			Tag         string `json:"tag"`
			DauerMin    int    `json:"dauer_min"`
			Zustaendig  string `json:"zustaendig"`
			Begruendung struct {
				Code       string  `json:"code"`
				ZuletztBei *string `json:"zuletzt_bei"`
			} `json:"begruendung"`
		} `json:"aufgaben"`
		Bilanz []struct {
			MitgliedId        string `json:"mitglied_id"`
			AuslastungProzent int    `json:"auslastung_prozent"`
		} `json:"bilanz"`
		Uebersprungen []any `json:"uebersprungen"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &plan); err != nil {
		t.Fatal(err)
	}

	if plan.Woche != "2026-W38" {
		t.Errorf("woche = %q", plan.Woche)
	}
	if plan.Haushalt.Id != "familie-a" || plan.Haushalt.Name != "Familie A" {
		t.Errorf("haushalt = %+v", plan.Haushalt)
	}
	if len(plan.Aufgaben) != 1 {
		t.Fatalf("%d Aufgaben, erwartet 1", len(plan.Aufgaben))
	}
	a := plan.Aufgaben[0]
	if a.VorlageId != "t-bad" || a.Tag != "2026-09-19" || a.DauerMin != 35 || a.Zustaendig != "m-ben" {
		t.Errorf("Aufgabe = %+v", a)
	}
	if a.Begruendung.Code != "rotation" {
		t.Errorf("Begründung = %q, erwartet rotation", a.Begruendung.Code)
	}
	if a.Begruendung.ZuletztBei == nil || *a.Begruendung.ZuletztBei != "m-anna" {
		t.Errorf("zuletzt_bei fehlt oder ist falsch: %+v", a.Begruendung)
	}
	if len(plan.Bilanz) != 2 || plan.Bilanz[0].AuslastungProzent != 10 {
		t.Errorf("Bilanz = %+v", plan.Bilanz)
	}
	// Leere Listen müssen als [] herauskommen, nicht als null.
	if plan.Uebersprungen == nil {
		t.Error("uebersprungen ist null statt einer leeren Liste")
	}
}

func TestHaushalteListe(t *testing.T) {
	rec := httptest.NewRecorder()
	testServer(nil).Handler().ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/api/haushalte", nil))

	if rec.Code != http.StatusOK {
		t.Fatalf("Status %d", rec.Code)
	}
	var liste []struct {
		Id         string `json:"id"`
		Mitglieder []any  `json:"mitglieder"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &liste); err != nil {
		t.Fatal(err)
	}
	if len(liste) != 1 || liste[0].Id != "familie-a" || len(liste[0].Mitglieder) != 2 {
		t.Errorf("Liste = %+v", liste)
	}
}
