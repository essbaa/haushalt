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
	// UpdateHousehold ändert die Einstellungen. Nur planende Personen.
	UpdateHousehold(ctx context.Context, subject, id string, c planner.HouseholdChange) (planner.Household, error)
	// UpdateMember ändert eine Person. Den eigenen Namen und die eigene Zeit
	// darf jeder, alles Weitere die planenden.
	UpdateMember(ctx context.Context, subject, id, memberID string, c planner.MemberChange) (planner.Household, error)
	// Recompute rechnet eine festgeschriebene Woche neu — ohne anzufassen,
	// was schon Spuren hinterlassen hat.
	Recompute(ctx context.Context, subject, id string, week planner.Week) (planner.Result, planner.Household, error)
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
	s := planner.Setup{
		Name: b.Name,
		Context: planner.Context{
			Home: planner.Home(b.Wohnform),
		},
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

	return openapi.GetWochenplan200JSONResponse(planNachAussen(household, result, ich, rolle, mitBilanz)), nil
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

func planNachAussen(h planner.Household, r planner.Result, ich string, rolle planner.Role, mitBilanz bool) openapi.Wochenplan {
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
		planNachAussen(haushalt, result, ich, rolle, rolle == planner.RolePlanner),
	), nil
}
