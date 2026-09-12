package httpapi

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/zakaria/haushalt/api/internal/auth"
	"github.com/zakaria/haushalt/api/internal/httpapi/openapi"
	"github.com/zakaria/haushalt/api/internal/planner"
)

// Plans ist die Quelle der Wochenpläne — wieder ein Interface an der Stelle,
// wo es gebraucht wird. Heute erfüllt es der Katalog aus dem Repo, ab T5 die
// Datenbank. Dieses Paket merkt den Unterschied nicht.
type Plans interface {
	// Arrive verbindet eine angemeldete Person mit einer Person im Haushalt
	// und legt beim ersten Mal beides an.
	Arrive(ctx context.Context, subject, name string) error
	// Households sind die Haushalte, die dieser Aufrufer sehen darf.
	// Leeres subject heißt: nicht angemeldet.
	Households(ctx context.Context, subject string) ([]planner.Household, error)
	// Plan prüft selbst, ob der Aufrufer diesen Haushalt sehen darf. Die
	// Prüfung gehört zur Quelle und nicht in den Handler — ein Handler kann
	// sie vergessen.
	Plan(ctx context.Context, subject, id string, week planner.Week) (planner.Result, planner.Household, error)
	// RoleOf ist die Rolle des Aufrufers in diesem Haushalt. Leer heißt: kein
	// Mitglied — bei Demo-Haushalten der Normalfall.
	RoleOf(ctx context.Context, subject, id string) (planner.Role, error)
	// Invite erzeugt einen einmalig gültigen Code. Nur planende Personen.
	Invite(ctx context.Context, subject, id string, role planner.Role) (string, time.Time, error)
	// Accept verbindet die angemeldete Person mit dem Haushalt aus der
	// Einladung.
	Accept(ctx context.Context, subject, name, code string) (planner.Household, error)
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
	var subject string
	if id, ok := auth.From(ctx); ok {
		subject = id.Subject
		// Hier entsteht beim ersten angemeldeten Zugriff der eigene Haushalt.
		//
		// Ein GET, der schreibt — das ist die unschöne Seite. Dafür passiert
		// es genau dort, wo jemand zum ersten Mal nach seinen Haushalten
		// fragt, es ist idempotent, und die Alternative wäre ein
		// Einrichtungsschritt, den man vergessen kann. Mit dem richtigen
		// Onboarding (T6) fällt es weg.
		if err := a.plans.Arrive(ctx, id.Subject, id.Name); err != nil {
			return nil, err
		}
	}

	haushalte, err := a.plans.Households(ctx, subject)
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

	var subject string
	if id, ok := auth.From(ctx); ok {
		subject = id.Subject
	}

	result, household, err := a.plans.Plan(ctx, subject, r.HaushaltId, week)
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

	// Die Bilanz ist der einzige Teil, der an die Rolle gebunden ist: Wer
	// ausführt, sieht den ganzen Plan, aber nicht die Auswertung darüber, wer
	// im Haushalt wie viel trägt.
	//
	// Bei Demo-Haushalten ist die Rolle leer und die Bilanz sichtbar — sie
	// haben nichts zu verbergen, und sie sind der Teil, der das Produkt
	// erklärt.
	rolle, err := a.plans.RoleOf(ctx, subject, r.HaushaltId)
	if err != nil {
		return nil, err
	}
	mitBilanz := rolle == planner.RolePlanner || rolle == ""

	return openapi.GetWochenplan200JSONResponse(planNachAussen(household, result, rolle, mitBilanz)), nil
}

// CreateEinladung erzeugt einen Code, mit dem jemand in den Haushalt kommt.
func (a api) CreateEinladung(ctx context.Context, r openapi.CreateEinladungRequestObject) (openapi.CreateEinladungResponseObject, error) {
	id, ok := auth.From(ctx)
	if !ok {
		return openapi.CreateEinladung403JSONResponse{Fehler: "dafür muss man angemeldet sein"}, nil
	}
	if r.Body == nil {
		return openapi.CreateEinladung403JSONResponse{Fehler: "es fehlt die rolle"}, nil
	}

	code, bis, err := a.plans.Invite(ctx, id.Subject, r.HaushaltId, planner.Role(r.Body.Rolle))
	switch {
	case errors.Is(err, planner.ErrUnknownHousehold):
		return openapi.CreateEinladung404JSONResponse{
			Fehler: fmt.Sprintf("den Haushalt %q gibt es nicht", r.HaushaltId),
		}, nil
	case errors.Is(err, planner.ErrNotAllowed):
		return openapi.CreateEinladung403JSONResponse{
			Fehler: "nur planende Personen dürfen einladen",
		}, nil
	case err != nil:
		return nil, err
	}

	return openapi.CreateEinladung201JSONResponse{
		Code:       code,
		Rolle:      openapi.EinladungRolle(r.Body.Rolle),
		GueltigBis: bis,
	}, nil
}

// AcceptEinladung verbindet die angemeldete Person mit dem Haushalt.
func (a api) AcceptEinladung(ctx context.Context, r openapi.AcceptEinladungRequestObject) (openapi.AcceptEinladungResponseObject, error) {
	id, ok := auth.From(ctx)
	if !ok {
		return openapi.AcceptEinladung401JSONResponse{Fehler: "dafür muss man angemeldet sein"}, nil
	}

	haushalt, err := a.plans.Accept(ctx, id.Subject, id.Name, r.Code)
	switch {
	case errors.Is(err, planner.ErrUnknownInvitation):
		// Eine Antwort für drei Fälle: gibt es nicht, abgelaufen, verbraucht.
		// Wer Codes durchprobiert, soll nicht erfahren, welcher zutrifft.
		return openapi.AcceptEinladung404JSONResponse{
			Fehler: "diese Einladung gilt nicht mehr",
		}, nil
	case errors.Is(err, planner.ErrNotAllowed):
		return openapi.AcceptEinladung401JSONResponse{
			Fehler: "dafür muss man angemeldet sein",
		}, nil
	case err != nil:
		return nil, err
	}

	return openapi.AcceptEinladung200JSONResponse(haushaltNachAussen(haushalt)), nil
}

// ------------------------------------------------------------ Übersetzung
//
// Die Typen des Planers und die Typen der Spezifikation sind mit Absicht nicht
// dieselben. Der Kern soll sich ändern dürfen, ohne dass der Vertrag bricht —
// und der Vertrag soll Felder tragen dürfen (Auslastung in Prozent), die im
// Kern nichts zu suchen haben. Der Preis ist diese Übersetzung; sie steht an
// einer Stelle und ist langweilig, und das ist genau richtig.

func planNachAussen(h planner.Household, r planner.Result, rolle planner.Role, mitBilanz bool) openapi.Wochenplan {
	plan := openapi.Wochenplan{
		Woche:    r.Week.String(),
		Haushalt: haushaltNachAussen(h),
		// Nicht-nil, damit leere Listen als [] und nicht als null im JSON
		// stehen. Ein Client, der `.map()` darauf aufruft, dankt es.
		Aufgaben:      []openapi.Aufgabe{},
		Uebersprungen: []openapi.Uebersprungen{},
	}
	if rolle != "" {
		r := openapi.WochenplanMeineRolle(rolle)
		plan.MeineRolle = &r
	}
	if mitBilanz {
		plan.Bilanz = &[]openapi.Bilanz{}
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
		if !mitBilanz {
			break
		}
		*plan.Bilanz = append(*plan.Bilanz, openapi.Bilanz{
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
