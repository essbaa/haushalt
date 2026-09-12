package httpapi

import (
	"context"
	"errors"
	"fmt"

	"github.com/zakaria/haushalt/api/internal/auth"
	"github.com/zakaria/haushalt/api/internal/httpapi/openapi"
	"github.com/zakaria/haushalt/api/internal/planner"
)

// Plans ist die Quelle der Wochenpläne — wieder ein Interface an der Stelle,
// wo es gebraucht wird. Heute erfüllt es der Katalog aus dem Repo, ab T5 die
// Datenbank. Dieses Paket merkt den Unterschied nicht.
type Plans interface {
	Households(ctx context.Context) ([]planner.Household, error)
	Plan(ctx context.Context, id string, week planner.Week) (planner.Result, planner.Household, error)
}

// api erfüllt openapi.StrictServerInterface.
//
// „Strict" heißt: Die Parameter kommen geparst herein, und zurück geht ein
// Antworttyp, den der Generator aus der Spezifikation gebaut hat. Einen
// Statuscode, der nicht in openapi.yaml steht, kann dieser Code gar nicht
// erzeugen — er würde nicht übersetzen.
type api struct {
	plans   Plans
	version string
}

func (a api) GetVersion(context.Context, openapi.GetVersionRequestObject) (openapi.GetVersionResponseObject, error) {
	return openapi.GetVersion200JSONResponse{Version: a.version}, nil
}

// GetIch beantwortet, was das Token beweist.
//
// Ohne Token ist das „niemand", und das ist eine gültige Antwort mit Status
// 200 — kein 401. Die Frage lautet „wer fragt?", nicht „darf ich das?".
func (a api) GetIch(ctx context.Context, _ openapi.GetIchRequestObject) (openapi.GetIchResponseObject, error) {
	id, ok := auth.From(ctx)
	if !ok {
		return openapi.GetIch200JSONResponse{Angemeldet: false}, nil
	}
	antwort := openapi.GetIch200JSONResponse{Angemeldet: true, Subject: &id.Subject}
	if id.Name != "" {
		antwort.Name = &id.Name
	}
	if id.Email != "" {
		antwort.Email = &id.Email
	}
	return antwort, nil
}

func (a api) ListHaushalte(ctx context.Context, _ openapi.ListHaushalteRequestObject) (openapi.ListHaushalteResponseObject, error) {
	haushalte, err := a.plans.Households(ctx)
	if err != nil {
		return nil, err
	}
	out := openapi.ListHaushalte200JSONResponse{}
	for _, h := range haushalte {
		out = append(out, haushaltNachAussen(h))
	}
	return out, nil
}

func (a api) GetWochenplan(ctx context.Context, r openapi.GetWochenplanRequestObject) (openapi.GetWochenplanResponseObject, error) {
	week, err := planner.ParseWeek(r.Woche)
	if err != nil {
		return openapi.GetWochenplan400JSONResponse{Fehler: err.Error()}, nil
	}

	result, household, err := a.plans.Plan(ctx, r.HaushaltId, week)
	switch {
	case errors.Is(err, planner.ErrUnknownHousehold):
		return openapi.GetWochenplan404JSONResponse{
			Fehler: fmt.Sprintf("den Haushalt %q gibt es nicht", r.HaushaltId),
		}, nil
	case err != nil:
		// Alles andere ist ein Fehler auf unserer Seite. Ihn hier
		// zurückzugeben statt zu verkleiden, ist Absicht: Die
		// Fehler-Middleware macht daraus eine 500 und schreibt ihn ins Log.
		return nil, err
	}

	return openapi.GetWochenplan200JSONResponse(planNachAussen(household, result)), nil
}

// ------------------------------------------------------------ Übersetzung
//
// Die Typen des Planers und die Typen der Spezifikation sind mit Absicht nicht
// dieselben. Der Kern soll sich ändern dürfen, ohne dass der Vertrag bricht —
// und der Vertrag soll Felder tragen dürfen (Auslastung in Prozent), die im
// Kern nichts zu suchen haben. Der Preis ist diese Übersetzung; sie steht an
// einer Stelle und ist langweilig, und das ist genau richtig.

func planNachAussen(h planner.Household, r planner.Result) openapi.Wochenplan {
	plan := openapi.Wochenplan{
		Woche:    r.Week.String(),
		Haushalt: haushaltNachAussen(h),
		// Nicht-nil, damit leere Listen als [] und nicht als null im JSON
		// stehen. Ein Client, der `.map()` darauf aufruft, dankt es.
		Aufgaben:      []openapi.Aufgabe{},
		Bilanz:        []openapi.Bilanz{},
		Uebersprungen: []openapi.Uebersprungen{},
	}

	for _, t := range r.Tasks {
		aufgabe := openapi.Aufgabe{
			VorlageId:   t.TemplateID,
			Titel:       t.Title,
			Kategorie:   openapi.Kategorie(t.Category),
			Art:         openapi.AufgabeArt(t.Kind),
			Tag:         t.Day.String(),
			Zeitfenster: openapi.AufgabeZeitfenster(t.Slot),
			DauerMin:    t.DurationMin,
			Kopflast:    int(t.HeadLoad),
			Zustaendig:  t.AssigneeID,
			Begruendung: openapi.Begruendung{Code: openapi.BegruendungCode(t.Reason.Code)},
		}
		if !t.Deadline.IsZero() {
			frist := t.Deadline.String()
			aufgabe.Frist = &frist
		}
		if t.Reason.Previous != "" {
			zuletzt := t.Reason.Previous
			aufgabe.Begruendung.ZuletztBei = &zuletzt
		}
		plan.Aufgaben = append(plan.Aufgaben, aufgabe)
	}

	for _, l := range r.Balance {
		plan.Bilanz = append(plan.Bilanz, openapi.Bilanz{
			MitgliedId:        l.MemberID,
			Minuten:           l.Minutes,
			Kopflast:          l.HeadLoad,
			Gewichtet:         l.Weighted,
			KapazitaetMinuten: l.Capacity,
			AuslastungProzent: l.Utilization(),
			Aufgaben:          l.Tasks,
		})
	}

	for _, s := range r.Skipped {
		plan.Uebersprungen = append(plan.Uebersprungen, openapi.Uebersprungen{
			VorlageId: s.TemplateID,
			Titel:     s.Title,
			Grund:     openapi.UebersprungenGrund(s.Code),
		})
	}

	return plan
}

func haushaltNachAussen(h planner.Household) openapi.Haushalt {
	out := openapi.Haushalt{Id: h.ID, Name: h.Name, Mitglieder: []openapi.Mitglied{}}
	for _, m := range h.Members {
		out.Mitglieder = append(out.Mitglieder, openapi.Mitglied{
			Id:    m.ID,
			Name:  m.Name,
			Rolle: openapi.MitgliedRolle(m.Role),
		})
	}
	return out
}
