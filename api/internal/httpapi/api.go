package httpapi

import (
	"context"
	"errors"
	"fmt"

	"github.com/zakaria/haushalt/api/internal/auth"
	"github.com/zakaria/haushalt/api/internal/httpapi/openapi"
	"github.com/zakaria/haushalt/api/internal/library"
	"github.com/zakaria/haushalt/api/internal/planner"
)

// Plans ist die Quelle der Wochenpläne — wieder ein Interface an der Stelle,
// wo es gebraucht wird. Heute erfüllt es der Katalog aus dem Repo, ab T5 die
// Datenbank. Dieses Paket merkt den Unterschied nicht.
type Plans interface {
	// Create legt einen Haushalt samt Mitgliedern an und macht den Aufrufer
	// darin planend.
	Create(ctx context.Context, subject string, setup planner.Setup) (planner.Household, error)
	// Households sind die Haushalte, die dieser Aufrufer sehen darf.
	// Leeres subject heißt: nicht angemeldet.
	Households(ctx context.Context, subject string) ([]planner.Household, error)
	// Plan prüft selbst, ob der Aufrufer diesen Haushalt sehen darf. Die
	// Prüfung gehört zur Quelle und nicht in den Handler — ein Handler kann
	// sie vergessen.
	Plan(ctx context.Context, subject, id string, week planner.Week) (planner.Result, planner.Household, error)
	// MemberOf sagt, wer der Aufrufer in diesem Haushalt ist: seine Kennung
	// als Person und seine Rolle. Beides leer heißt: kein Mitglied — bei
	// Demo-Haushalten der Normalfall.
	//
	// Zusammen und nicht in zwei Methoden, weil es eine Zeile in der Datenbank
	// ist. Zwei Abfragen für dieselbe Zeile laufen irgendwann auseinander.
	MemberOf(ctx context.Context, subject, id string) (string, planner.Role, error)
	// Invite erzeugt einen einmalig gültigen Code. Nur planende Personen.
	// Leeres memberID heißt: Es kommt jemand dazu, den es noch nicht gibt.
	Invite(ctx context.Context, subject, id string, role planner.Role, memberID string) (planner.Invitation, error)
	// Accept verbindet die angemeldete Person mit dem Haushalt aus der
	// Einladung.
	Accept(ctx context.Context, subject, name, code string) (planner.Household, error)
	// MarkDone hakt eine Aufgabe ab oder nimmt das Häkchen zurück.
	MarkDone(ctx context.Context, subject, taskID string, done bool) error
	// HandOver gibt eine Aufgabe zurück in den Haushalt und liefert den Namen
	// der Person, die übernimmt — leer, wenn niemand geeignet ist.
	HandOver(ctx context.Context, subject, taskID, reason string) (string, error)
	// Reassign gibt eine Aufgabe von Hand an eine bestimmte Person und
	// liefert deren Namen. Anders als HandOver wählt hier ein Mensch — der
	// Planer prüft nur noch, was diese Person nicht übernehmen kann.
	Reassign(ctx context.Context, subject, taskID, memberID string) (string, error)
	// UpdateHousehold ändert die Einstellungen. Nur planende Personen.
	UpdateHousehold(ctx context.Context, subject, id string, c planner.HouseholdChange) (planner.Household, error)
	// UpdateMember ändert eine Person. Den eigenen Namen und die eigene Zeit
	// darf jeder, alles Weitere die planenden.
	UpdateMember(ctx context.Context, subject, id, memberID string, c planner.MemberChange) (planner.Household, error)
	// Recompute rechnet eine festgeschriebene Woche neu — ohne anzufassen,
	// was schon Spuren hinterlassen hat.
	Recompute(ctx context.Context, subject, id string, week planner.Week) (planner.Result, planner.Household, error)
	// SetFacts trägt ein, was der Haushalt hat und was nicht.
	SetFacts(ctx context.Context, subject, id string, facts map[string]bool) (planner.Household, error)
	// TemplatesFor ist die ganze Bibliothek mit dem Stand dieses Haushalts.
	// Nicht "Templates": Der Katalog hat schon eine Methode dieses Namens für
	// die Bibliothek an sich, und die beiden sind verschiedene Fragen.
	TemplatesFor(ctx context.Context, subject, id string) ([]planner.TemplateState, planner.Household, error)
	// AddOwnTemplate legt eine Aufgabe an, die es in der Bibliothek nicht gibt.
	AddOwnTemplate(ctx context.Context, subject, id string, o planner.OwnTask) (planner.TaskTemplate, error)
	// SetTemplateActive bestellt eine Vorlage ab oder wieder an.
	SetTemplateActive(ctx context.Context, subject, id, templateID string, active bool) error
	// Occasions sind die eingetragenen Anlässe des Haushalts.
	Occasions(ctx context.Context, subject, id string) ([]planner.Occasion, error)
	// AddOccasion trägt einen Anlass ein, RemoveOccasion löscht ihn.
	AddOccasion(ctx context.Context, subject, id string, o planner.Occasion) (planner.Occasion, error)
	RemoveOccasion(ctx context.Context, subject, id, occasionID string) error
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
	facts   *library.Facts
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
	}

	haushalte, err := a.plans.Households(ctx, subject)
	if err != nil {
		return nil, err
	}

	out := openapi.ListHaushalte200JSONResponse{}
	for _, h := range haushalte {
		eintrag := haushaltNachAussen(h)

		// Eine Abfrage je Haushalt. Das sieht nach N+1 aus und ist es auch —
		// nur ist N hier die Zahl der Haushalte eines Menschen plus zwei
		// Beispiele. Wer das optimiert, bevor jemand fünfzig Haushalte hat,
		// tauscht Lesbarkeit gegen nichts.
		if subject != "" {
			ich, rolle, err := a.plans.MemberOf(ctx, subject, h.ID)
			if err != nil {
				return nil, err
			}
			if rolle != "" {
				r := openapi.HaushaltMeineRolle(rolle)
				eintrag.MeineRolle = &r
			}
			if ich != "" {
				eintrag.Ich = &ich
			}
		}
		out = append(out, eintrag)
	}
	return out, nil
}

// CreateHaushalt ist das Onboarding.
//
// Vorher entstand der Haushalt als Nebenwirkung des ersten GET auf
// /api/haushalte: idempotent, bequem und trotzdem falsch. Ein Lesezugriff, der
// schreibt, ist für jeden Zwischenspeicher eine Lüge, und er nimmt dem Nutzer
// die einzige Gelegenheit, zu sagen, wie sein Haushalt aussieht.
func (a api) CreateHaushalt(ctx context.Context, r openapi.CreateHaushaltRequestObject) (openapi.CreateHaushaltResponseObject, error) {
	id, ok := auth.From(ctx)
	if !ok {
		return openapi.CreateHaushalt401JSONResponse{
			Fehler: "dafür musst du angemeldet sein",
		}, nil
	}
	if r.Body == nil {
		return openapi.CreateHaushalt400JSONResponse{Fehler: "leere Anfrage"}, nil
	}

	haushalt, err := a.plans.Create(ctx, id.Subject, setupNachInnen(*r.Body))
	switch {
	case errors.Is(err, planner.ErrInvalidSetup):
		// Die Meldung ist für Menschen geschrieben und landet im Formular.
		return openapi.CreateHaushalt400JSONResponse{Fehler: err.Error()}, nil
	case errors.Is(err, planner.ErrNotAllowed):
		return openapi.CreateHaushalt401JSONResponse{Fehler: "dafür musst du angemeldet sein"}, nil
	case err != nil:
		return nil, err
	}

	eintrag := haushaltNachAussen(haushalt)
	rolle := openapi.HaushaltMeineRolle(planner.RolePlanner)
	eintrag.MeineRolle = &rolle
	return openapi.CreateHaushalt201JSONResponse(eintrag), nil
}

// setupNachInnen übersetzt die Anfrage in die Fachsprache. Geprüft wird hier
// nichts: Das tut planner.Setup, und zwar an einer Stelle statt an zweien.
func setupNachInnen(b openapi.NeuerHaushalt) planner.Setup {
	s := planner.Setup{Name: b.Name}
	// Fehlt die Wohnform, setzt Normalized() sie auf "wohnung". Das Onboarding
	// fragt nicht mehr danach: Keine Vorlage hängt daran, und eine Frage, die
	// nichts bewirkt, gehört nicht in die ersten neunzig Sekunden.
	if b.Wohnform != nil {
		s.Context.Home = planner.Home(*b.Wohnform)
	}
	if b.Zeitzone != nil {
		s.Timezone = *b.Zeitzone
	}
	if b.Garten != nil {
		s.Context.HasYard = *b.Garten
	}
	if b.Auto != nil {
		s.Context.HasCar = *b.Auto
	}
	if b.Haustiere != nil {
		s.Context.Pets = *b.Haustiere
	}
	if b.Zimmer != nil {
		s.Context.Rooms = *b.Zimmer
	}
	if b.Baeder != nil {
		s.Context.Baths = *b.Baeder
	}
	for _, m := range b.Mitglieder {
		person := planner.SetupMember{Name: m.Name, Role: planner.Role(m.Rolle)}
		if m.Geburtsjahr != nil {
			person.BirthYear = *m.Geburtsjahr
		}
		if m.Zeit != nil {
			person.Budget = planner.TimeBudget(*m.Zeit)
		}
		s.Members = append(s.Members, person)
	}
	return s
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
	ich, rolle, err := a.plans.MemberOf(ctx, subject, r.HaushaltId)
	if err != nil {
		return nil, err
	}
	mitBilanz := rolle == planner.RolePlanner || rolle == ""

	return openapi.GetWochenplan200JSONResponse(planNachAussen(household, result, ich, rolle, mitBilanz, a.facts)), nil
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

	var rolle planner.Role
	if r.Body.Rolle != nil {
		rolle = planner.Role(*r.Body.Rolle)
	}
	var mitglied string
	if r.Body.Mitglied != nil {
		mitglied = *r.Body.Mitglied
	}

	einladung, err := a.plans.Invite(ctx, id.Subject, r.HaushaltId, rolle, mitglied)
	switch {
	case errors.Is(err, planner.ErrUnknownHousehold):
		return openapi.CreateEinladung404JSONResponse{
			Fehler: fmt.Sprintf("den Haushalt %q gibt es nicht", r.HaushaltId),
		}, nil
	case errors.Is(err, planner.ErrNotAllowed):
		return openapi.CreateEinladung403JSONResponse{
			Fehler: "diese Einladung darfst du nicht ausstellen",
		}, nil
	case err != nil:
		return nil, err
	}

	antwort := openapi.CreateEinladung201JSONResponse{
		Code:       einladung.Code,
		Rolle:      openapi.EinladungRolle(einladung.Role),
		GueltigBis: einladung.Until,
	}
	if einladung.For != "" {
		antwort.Fuer = &einladung.For
	}
	return antwort, nil
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

func planNachAussen(h planner.Household, r planner.Result, ich string, rolle planner.Role, mitBilanz bool, katalog *library.Facts) openapi.Wochenplan {
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
	if ich != "" {
		plan.Ich = &ich
	}
	if mitBilanz {
		plan.Bilanz = &[]openapi.Bilanz{}
	}

	// Fragen nur für Mitglieder — und nur zwei. Wer beim Öffnen vierzehn
	// Fragen sieht, beantwortet keine; zwei mit sichtbarem Nutzen werden
	// beantwortet. Dieselbe Überlegung wie bei der Startdichte.
	if rolle == planner.RolePlanner {
		fragen := []openapi.Frage{}
		for _, id := range planner.OpenQuestions(r.Skipped, 2) {
			f, ok := katalog.Get(id)
			if !ok {
				// Ein Faktum ohne Frage wird nicht gefragt. Es bleibt
				// unbekannt, und die Vorlage bleibt draußen — richtig herum.
				continue
			}
			fragen = append(fragen, openapi.Frage{Faktum: f.ID, Frage: f.Question, Dann: f.Benefit})
		}
		if len(fragen) > 0 {
			plan.Fragen = &fragen
		}
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
		if t.ID != "" {
			kennung := t.ID
			aufgabe.Id = &kennung
			erledigt := t.Done
			aufgabe.Erledigt = &erledigt
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
	garten, auto := h.Context.HasYard, h.Context.HasCar
	zimmer, baeder := h.Context.Rooms, h.Context.Baths
	haustiere := h.Context.Pets
	if haustiere == nil {
		haustiere = []string{}
	}
	wohnform := openapi.HaushaltWohnform(h.Context.Home)

	out := openapi.Haushalt{
		Id:         h.ID,
		Name:       h.Name,
		Mitglieder: []openapi.Mitglied{},
		Wohnform:   &wohnform,
		Garten:     &garten,
		Auto:       &auto,
		Haustiere:  &haustiere,
		Zimmer:     &zimmer,
		Baeder:     &baeder,
	}
	for _, m := range h.Members {
		zugang := m.HasAccess
		minuten := make([]int, 0, len(m.CapacityMinutes))
		for _, v := range m.CapacityMinutes {
			minuten = append(minuten, v)
		}
		mitglied := openapi.Mitglied{
			Id:        m.ID,
			Name:      m.Name,
			Rolle:     openapi.MitgliedRolle(m.Role),
			HatZugang: &zugang,
			Minuten:   &minuten,
		}
		if m.BirthYear != 0 {
			jahr := m.BirthYear
			mitglied.Geburtsjahr = &jahr
		}
		// Die Stufe wird gerechnet, nicht gespeichert: Gespeichert sind die
		// Minuten. Passt keine Stufe genau, bleibt das Feld leer — dann hat
		// jemand von Hand gesetzt, und das ist ein eigener Zustand.
		if stufe := planner.BudgetOf(m.CapacityMinutes); stufe != "" {
			zeit := openapi.MitgliedZeit(stufe)
			mitglied.Zeit = &zeit
		}
		if m.Care != "" {
			betreuung := openapi.MitgliedBetreuung(m.Care)
			mitglied.Betreuung = &betreuung
		}
		out.Mitglieder = append(out.Mitglieder, mitglied)
	}
	return out
}

// SetErledigt hakt eine Aufgabe ab — oder nimmt das Häkchen zurück.
func (a api) SetErledigt(ctx context.Context, r openapi.SetErledigtRequestObject) (openapi.SetErledigtResponseObject, error) {
	id, ok := auth.From(ctx)
	if !ok {
		return openapi.SetErledigt403JSONResponse{Fehler: "dafür musst du angemeldet sein"}, nil
	}

	// Ohne Rumpf gilt „erledigt". Das ist der Fall, den es hundertmal am Tag
	// gibt; das Zurücknehmen ist die Ausnahme und darf das Feld kosten.
	erledigt := true
	if r.Body != nil && r.Body.Erledigt != nil {
		erledigt = *r.Body.Erledigt
	}

	switch err := a.plans.MarkDone(ctx, id.Subject, r.AufgabeId, erledigt); {
	case errors.Is(err, planner.ErrUnknownTask):
		return openapi.SetErledigt404JSONResponse{Fehler: "diese Aufgabe gibt es nicht"}, nil
	case errors.Is(err, planner.ErrNotAllowed):
		return openapi.SetErledigt403JSONResponse{Fehler: "das ist nicht deine Aufgabe"}, nil
	case err != nil:
		return nil, err
	}
	return openapi.SetErledigt204Response{}, nil
}

// AufgabeAbgeben gibt eine Aufgabe zurück in den Haushalt.
func (a api) AufgabeAbgeben(ctx context.Context, r openapi.AufgabeAbgebenRequestObject) (openapi.AufgabeAbgebenResponseObject, error) {
	id, ok := auth.From(ctx)
	if !ok {
		return openapi.AufgabeAbgeben403JSONResponse{Fehler: "dafür musst du angemeldet sein"}, nil
	}

	var grund string
	if r.Body != nil && r.Body.Grund != nil {
		grund = *r.Body.Grund
	}

	name, err := a.plans.HandOver(ctx, id.Subject, r.AufgabeId, grund)
	switch {
	case errors.Is(err, planner.ErrUnknownTask):
		return openapi.AufgabeAbgeben404JSONResponse{Fehler: "diese Aufgabe gibt es nicht"}, nil
	case errors.Is(err, planner.ErrNotAllowed):
		return openapi.AufgabeAbgeben403JSONResponse{Fehler: "das ist nicht deine Aufgabe"}, nil
	case err != nil:
		return nil, err
	}

	antwort := openapi.AufgabeAbgeben200JSONResponse{}
	if name != "" {
		antwort.Uebernimmt = &name
	}
	return antwort, nil
}

// AufgabeZuteilen gibt eine Aufgabe von Hand an eine bestimmte Person.
func (a api) AufgabeZuteilen(ctx context.Context, r openapi.AufgabeZuteilenRequestObject) (openapi.AufgabeZuteilenResponseObject, error) {
	id, ok := auth.From(ctx)
	if !ok {
		return openapi.AufgabeZuteilen403JSONResponse{Fehler: "dafür musst du angemeldet sein"}, nil
	}
	if r.Body == nil || r.Body.Mitglied == "" {
		return openapi.AufgabeZuteilen400JSONResponse{Fehler: "es fehlt die Person, die übernehmen soll"}, nil
	}

	name, err := a.plans.Reassign(ctx, id.Subject, r.AufgabeId, r.Body.Mitglied)

	// Die Meldung zu ErrNotEligible ist für Menschen geschrieben und steht im
	// Fehler selbst — „Bad putzen gilt ab 12 Jahren“ hilft, „nicht geeignet“
	// nicht. Deshalb errors.As und nicht errors.Is.
	var ungeeignet planner.NotEligibleError
	switch {
	case errors.Is(err, planner.ErrUnknownTask):
		return openapi.AufgabeZuteilen404JSONResponse{Fehler: "diese Aufgabe gibt es nicht"}, nil
	case errors.Is(err, planner.ErrNotAllowed):
		return openapi.AufgabeZuteilen403JSONResponse{Fehler: "das ist nicht deine Aufgabe"}, nil
	case errors.Is(err, planner.ErrUnknownMember):
		return openapi.AufgabeZuteilen400JSONResponse{Fehler: "diese Person gehört nicht zum Haushalt"}, nil
	case errors.As(err, &ungeeignet):
		return openapi.AufgabeZuteilen400JSONResponse{Fehler: ungeeignet.Grund}, nil
	case err != nil:
		return nil, err
	}
	return openapi.AufgabeZuteilen200JSONResponse{Zustaendig: name}, nil
}

// UpdateHaushalt ändert die Einstellungen eines Haushalts.
func (a api) UpdateHaushalt(ctx context.Context, r openapi.UpdateHaushaltRequestObject) (openapi.UpdateHaushaltResponseObject, error) {
	id, ok := auth.From(ctx)
	if !ok {
		return openapi.UpdateHaushalt403JSONResponse{Fehler: "dafür musst du angemeldet sein"}, nil
	}
	if r.Body == nil {
		return openapi.UpdateHaushalt400JSONResponse{Fehler: "leere Anfrage"}, nil
	}

	c := planner.HouseholdChange{
		Name:     r.Body.Name,
		HasCar:   r.Body.Auto,
		HasYard:  r.Body.Garten,
		Pets:     r.Body.Haustiere,
		Timezone: r.Body.Zeitzone,
		Rooms:    r.Body.Zimmer,
		Baths:    r.Body.Baeder,
	}
	if r.Body.Wohnform != nil {
		wohnform := planner.Home(*r.Body.Wohnform)
		c.Home = &wohnform
	}

	haushalt, err := a.plans.UpdateHousehold(ctx, id.Subject, r.HaushaltId, c)
	switch {
	case errors.Is(err, planner.ErrInvalidSetup):
		return openapi.UpdateHaushalt400JSONResponse{Fehler: err.Error()}, nil
	case errors.Is(err, planner.ErrUnknownHousehold):
		return openapi.UpdateHaushalt404JSONResponse{
			Fehler: fmt.Sprintf("den Haushalt %q gibt es nicht", r.HaushaltId),
		}, nil
	case errors.Is(err, planner.ErrNotAllowed):
		return openapi.UpdateHaushalt403JSONResponse{Fehler: "das dürfen die planenden Personen"}, nil
	case err != nil:
		return nil, err
	}
	return openapi.UpdateHaushalt200JSONResponse(haushaltNachAussen(haushalt)), nil
}

// UpdateMitglied ändert eine Person im Haushalt.
func (a api) UpdateMitglied(ctx context.Context, r openapi.UpdateMitgliedRequestObject) (openapi.UpdateMitgliedResponseObject, error) {
	id, ok := auth.From(ctx)
	if !ok {
		return openapi.UpdateMitglied403JSONResponse{Fehler: "dafür musst du angemeldet sein"}, nil
	}
	if r.Body == nil {
		return openapi.UpdateMitglied400JSONResponse{Fehler: "leere Anfrage"}, nil
	}

	c := planner.MemberChange{Name: r.Body.Name, BirthYear: r.Body.Geburtsjahr}
	if r.Body.Betreuung != nil {
		betreuung := planner.Care(*r.Body.Betreuung)
		c.Care = &betreuung
	}
	if r.Body.Rolle != nil {
		rolle := planner.Role(*r.Body.Rolle)
		c.Role = &rolle
	}
	// Die Stufe wird hier in Minuten übersetzt und nicht im Browser: Sonst
	// gäbe es zwei Vorstellungen davon, was „mittel" heißt, und die zweite
	// wäre die, die niemand pflegt.
	if r.Body.Zeit != nil {
		minuten := planner.TimeBudget(*r.Body.Zeit).Minutes()
		c.Minutes = &minuten
	}
	if r.Body.Minuten != nil {
		var minuten [7]int
		for i, v := range *r.Body.Minuten {
			if i < len(minuten) {
				minuten[i] = v
			}
		}
		c.Minutes = &minuten
	}

	haushalt, err := a.plans.UpdateMember(ctx, id.Subject, r.HaushaltId, r.MitgliedId, c)
	switch {
	case errors.Is(err, planner.ErrInvalidSetup):
		return openapi.UpdateMitglied400JSONResponse{Fehler: err.Error()}, nil
	case errors.Is(err, planner.ErrUnknownHousehold):
		return openapi.UpdateMitglied404JSONResponse{
			Fehler: fmt.Sprintf("den Haushalt %q gibt es nicht", r.HaushaltId),
		}, nil
	case errors.Is(err, planner.ErrNotAllowed):
		return openapi.UpdateMitglied403JSONResponse{Fehler: err.Error()}, nil
	case err != nil:
		return nil, err
	}
	return openapi.UpdateMitglied200JSONResponse(haushaltNachAussen(haushalt)), nil
}

// WocheNeuRechnen ersetzt den Vorschlag und lässt stehen, was passiert ist.
func (a api) WocheNeuRechnen(ctx context.Context, r openapi.WocheNeuRechnenRequestObject) (openapi.WocheNeuRechnenResponseObject, error) {
	id, ok := auth.From(ctx)
	if !ok {
		return openapi.WocheNeuRechnen403JSONResponse{Fehler: "dafür musst du angemeldet sein"}, nil
	}
	week, err := planner.ParseWeek(r.Woche)
	if err != nil {
		return openapi.WocheNeuRechnen400JSONResponse{Fehler: err.Error()}, nil
	}

	result, haushalt, err := a.plans.Recompute(ctx, id.Subject, r.HaushaltId, week)
	switch {
	case errors.Is(err, planner.ErrUnknownHousehold):
		return openapi.WocheNeuRechnen404JSONResponse{
			Fehler: fmt.Sprintf("den Haushalt %q gibt es nicht", r.HaushaltId),
		}, nil
	case errors.Is(err, planner.ErrNotAllowed):
		return openapi.WocheNeuRechnen403JSONResponse{Fehler: "das dürfen die planenden Personen"}, nil
	case err != nil:
		return nil, err
	}

	ich, rolle, err := a.plans.MemberOf(ctx, id.Subject, r.HaushaltId)
	if err != nil {
		return nil, err
	}
	return openapi.WocheNeuRechnen200JSONResponse(
		planNachAussen(haushalt, result, ich, rolle, rolle == planner.RolePlanner, a.facts),
	), nil
}

// SetFakten beantwortet eine Frage zum Haushalt.
func (a api) SetFakten(ctx context.Context, r openapi.SetFaktenRequestObject) (openapi.SetFaktenResponseObject, error) {
	id, ok := auth.From(ctx)
	if !ok {
		return openapi.SetFakten403JSONResponse{Fehler: "dafür musst du angemeldet sein"}, nil
	}
	if r.Body == nil || len(*r.Body) == 0 {
		return openapi.SetFakten400JSONResponse{Fehler: "keine Antwort dabei"}, nil
	}

	haushalt, err := a.plans.SetFacts(ctx, id.Subject, r.HaushaltId, *r.Body)
	switch {
	case errors.Is(err, planner.ErrUnknownHousehold):
		return openapi.SetFakten404JSONResponse{
			Fehler: fmt.Sprintf("den Haushalt %q gibt es nicht", r.HaushaltId),
		}, nil
	case errors.Is(err, planner.ErrNotAllowed):
		return openapi.SetFakten403JSONResponse{Fehler: "das dürfen die planenden Personen"}, nil
	case err != nil:
		return nil, err
	}
	return openapi.SetFakten200JSONResponse(haushaltNachAussen(haushalt)), nil
}

// ListVorlagen zeigt die ganze Bibliothek mit dem Stand dieses Haushalts.
func (a api) ListVorlagen(ctx context.Context, r openapi.ListVorlagenRequestObject) (openapi.ListVorlagenResponseObject, error) {
	var subject string
	if id, ok := auth.From(ctx); ok {
		subject = id.Subject
	}

	stand, _, err := a.plans.TemplatesFor(ctx, subject, r.HaushaltId)
	switch {
	case errors.Is(err, planner.ErrUnknownHousehold):
		return openapi.ListVorlagen404JSONResponse{
			Fehler: fmt.Sprintf("den Haushalt %q gibt es nicht", r.HaushaltId),
		}, nil
	case err != nil:
		return nil, err
	}

	out := openapi.ListVorlagen200JSONResponse{}
	for _, v := range stand {
		eintrag := openapi.VorlagenStand{
			Id:        v.Template.ID,
			Titel:     v.Template.Title,
			Kategorie: openapi.Kategorie(v.Template.Category),
			Art:       openapi.VorlagenStandArt(v.Template.Kind),
			DauerMin:  v.Template.DurationMin,
			Kopflast:  int(v.Template.HeadLoad),
			Aktiv:     v.Active,
		}
		if v.Template.Source == planner.SourceHousehold {
			eigene := true
			eintrag.Eigene = &eigene
		}
		if v.Reason != "" {
			grund := openapi.VorlagenStandGrund(v.Reason)
			eintrag.Grund = &grund
		}
		if v.Fact != "" {
			faktum := v.Fact
			eintrag.Faktum = &faktum
			if f, ok := a.facts.Get(v.Fact); ok {
				eintrag.Frage = &f.Question
				eintrag.Dann = &f.Benefit
			}
		}
		out = append(out, eintrag)
	}
	return out, nil
}

// SetVorlage bestellt eine Vorlage ab oder wieder an.
func (a api) SetVorlage(ctx context.Context, r openapi.SetVorlageRequestObject) (openapi.SetVorlageResponseObject, error) {
	id, ok := auth.From(ctx)
	if !ok {
		return openapi.SetVorlage403JSONResponse{Fehler: "dafür musst du angemeldet sein"}, nil
	}
	if r.Body == nil {
		return openapi.SetVorlage403JSONResponse{Fehler: "leere Anfrage"}, nil
	}

	switch err := a.plans.SetTemplateActive(ctx, id.Subject, r.HaushaltId, r.VorlageId, r.Body.Aktiv); {
	case errors.Is(err, planner.ErrUnknownHousehold):
		return openapi.SetVorlage404JSONResponse{
			Fehler: fmt.Sprintf("den Haushalt %q gibt es nicht", r.HaushaltId),
		}, nil
	case errors.Is(err, planner.ErrNotAllowed):
		return openapi.SetVorlage403JSONResponse{Fehler: "das dürfen die planenden Personen"}, nil
	case err != nil:
		return nil, err
	}
	return openapi.SetVorlage204Response{}, nil
}

func anlassNachAussen(o planner.Occasion) openapi.Anlass {
	return openapi.Anlass{
		Id:        o.ID,
		Titel:     o.Title,
		Tag:       o.Date.String(),
		Art:       openapi.AnlassArt(o.Kind),
		Jaehrlich: o.Yearly,
	}
}

// ListAnlaesse zeigt die eingetragenen Anlässe.
func (a api) ListAnlaesse(ctx context.Context, r openapi.ListAnlaesseRequestObject) (openapi.ListAnlaesseResponseObject, error) {
	var subject string
	if id, ok := auth.From(ctx); ok {
		subject = id.Subject
	}

	anlaesse, err := a.plans.Occasions(ctx, subject, r.HaushaltId)
	switch {
	case errors.Is(err, planner.ErrUnknownHousehold):
		return openapi.ListAnlaesse404JSONResponse{
			Fehler: fmt.Sprintf("den Haushalt %q gibt es nicht", r.HaushaltId),
		}, nil
	case err != nil:
		return nil, err
	}

	out := openapi.ListAnlaesse200JSONResponse{}
	for _, o := range anlaesse {
		out = append(out, anlassNachAussen(o))
	}
	return out, nil
}

// CreateAnlass trägt einen Anlass ein.
func (a api) CreateAnlass(ctx context.Context, r openapi.CreateAnlassRequestObject) (openapi.CreateAnlassResponseObject, error) {
	id, ok := auth.From(ctx)
	if !ok {
		return openapi.CreateAnlass403JSONResponse{Fehler: "dafür musst du angemeldet sein"}, nil
	}
	if r.Body == nil {
		return openapi.CreateAnlass400JSONResponse{Fehler: "leere Anfrage"}, nil
	}

	tag, err := planner.ParseDate(r.Body.Tag)
	if err != nil {
		return openapi.CreateAnlass400JSONResponse{Fehler: err.Error()}, nil
	}

	o := planner.Occasion{Title: r.Body.Titel, Date: tag, Kind: string(r.Body.Art)}
	if r.Body.Jaehrlich != nil {
		o.Yearly = *r.Body.Jaehrlich
	}

	neu, err := a.plans.AddOccasion(ctx, id.Subject, r.HaushaltId, o)
	switch {
	case errors.Is(err, planner.ErrInvalidSetup):
		return openapi.CreateAnlass400JSONResponse{Fehler: err.Error()}, nil
	case errors.Is(err, planner.ErrUnknownHousehold):
		return openapi.CreateAnlass404JSONResponse{
			Fehler: fmt.Sprintf("den Haushalt %q gibt es nicht", r.HaushaltId),
		}, nil
	case errors.Is(err, planner.ErrNotAllowed):
		return openapi.CreateAnlass403JSONResponse{Fehler: "das dürfen die planenden Personen"}, nil
	case err != nil:
		return nil, err
	}
	return openapi.CreateAnlass201JSONResponse(anlassNachAussen(neu)), nil
}

// DeleteAnlass löscht einen Anlass.
func (a api) DeleteAnlass(ctx context.Context, r openapi.DeleteAnlassRequestObject) (openapi.DeleteAnlassResponseObject, error) {
	id, ok := auth.From(ctx)
	if !ok {
		return openapi.DeleteAnlass403JSONResponse{Fehler: "dafür musst du angemeldet sein"}, nil
	}

	switch err := a.plans.RemoveOccasion(ctx, id.Subject, r.HaushaltId, r.AnlassId); {
	case errors.Is(err, planner.ErrUnknownHousehold):
		return openapi.DeleteAnlass404JSONResponse{
			Fehler: fmt.Sprintf("den Haushalt %q gibt es nicht", r.HaushaltId),
		}, nil
	case errors.Is(err, planner.ErrNotAllowed):
		return openapi.DeleteAnlass403JSONResponse{Fehler: "das dürfen die planenden Personen"}, nil
	case err != nil:
		return nil, err
	}
	return openapi.DeleteAnlass204Response{}, nil
}

// CreateEigeneVorlage legt eine Aufgabe an, die es in der Bibliothek nicht gibt.
func (a api) CreateEigeneVorlage(ctx context.Context, r openapi.CreateEigeneVorlageRequestObject) (openapi.CreateEigeneVorlageResponseObject, error) {
	id, ok := auth.From(ctx)
	if !ok {
		return openapi.CreateEigeneVorlage403JSONResponse{Fehler: "dafür musst du angemeldet sein"}, nil
	}
	if r.Body == nil {
		return openapi.CreateEigeneVorlage400JSONResponse{Fehler: "leere Anfrage"}, nil
	}

	o := planner.OwnTask{
		Title:       r.Body.Titel,
		Category:    planner.Category(r.Body.Kategorie),
		DurationMin: r.Body.DauerMin,
		EveryDays:   int(r.Body.AlleTage),
	}
	if r.Body.Denken != nil {
		o.Remember = *r.Body.Denken
	}
	if r.Body.Klaeren != nil {
		o.Arrange = *r.Body.Klaeren
	}
	if r.Body.NurErwachsene != nil {
		o.AdultsOnly = *r.Body.NurErwachsene
	}

	vorlage, err := a.plans.AddOwnTemplate(ctx, id.Subject, r.HaushaltId, o)
	switch {
	case errors.Is(err, planner.ErrInvalidSetup):
		return openapi.CreateEigeneVorlage400JSONResponse{Fehler: err.Error()}, nil
	case errors.Is(err, planner.ErrUnknownHousehold):
		return openapi.CreateEigeneVorlage404JSONResponse{
			Fehler: fmt.Sprintf("den Haushalt %q gibt es nicht", r.HaushaltId),
		}, nil
	case errors.Is(err, planner.ErrNotAllowed):
		return openapi.CreateEigeneVorlage403JSONResponse{Fehler: "das dürfen die planenden Personen"}, nil
	case err != nil:
		return nil, err
	}

	eigene := true
	return openapi.CreateEigeneVorlage201JSONResponse{
		Id:        vorlage.ID,
		Titel:     vorlage.Title,
		Kategorie: openapi.Kategorie(vorlage.Category),
		Art:       openapi.VorlagenStandArt(vorlage.Kind),
		DauerMin:  vorlage.DurationMin,
		Kopflast:  int(vorlage.HeadLoad),
		Aktiv:     true,
		Eigene:    &eigene,
	}, nil
}
