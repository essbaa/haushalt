package storage

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/zakaria/haushalt/api/internal/planner"
	"github.com/zakaria/haushalt/api/internal/storage/db"
)

// UpdateHousehold ändert die Einstellungen eines Haushalts.
//
// Nur planende Personen. Was hier steht, entscheidet, welche Aufgaben es im
// Haushalt überhaupt gibt — „kein Garten" löscht das Rasenmähen für alle.
func (p *Plans) UpdateHousehold(ctx context.Context, subject, id string, c planner.HouseholdChange) (planner.Household, error) {
	zeile, err := p.alsPlanende(ctx, subject, id)
	if err != nil {
		return planner.Household{}, err
	}
	if err := c.Validate(); err != nil {
		return planner.Household{}, err
	}
	if !c.Touched() {
		return p.household(ctx, zeile)
	}

	params := db.UpdateHouseholdParams{ID: zeile.ID}
	if c.Name != nil {
		params.Name = c.Name
	}
	if c.Home != nil {
		wert := string(*c.Home)
		params.Home = &wert
	}
	if c.HasCar != nil {
		params.HasCar = c.HasCar
	}
	if c.HasYard != nil {
		params.HasYard = c.HasYard
	}
	if c.Pets != nil {
		// Nicht-nil, aber leer: Genau so trägt man „keine Haustiere mehr" ein.
		// Ein nil-Slice wäre für COALESCE dasselbe wie „nicht mitgeschickt".
		haustiere := *c.Pets
		if haustiere == nil {
			haustiere = []string{}
		}
		params.Pets = haustiere
	}
	if c.Timezone != nil {
		params.Timezone = c.Timezone
	}
	if c.Rooms != nil {
		zimmer := int32(*c.Rooms)
		params.Rooms = &zimmer
	}
	if c.Baths != nil {
		baeder := int32(*c.Baths)
		params.Baths = &baeder
	}

	neu, err := p.db.UpdateHousehold(ctx, params)
	if err != nil {
		return planner.Household{}, err
	}
	return p.household(ctx, neu)
}

// UpdateMember ändert eine Person.
//
// Zwei Rechte, nicht eines: Den eigenen Namen und die eigene Zeit setzt jeder
// selbst — wie viel Zeit jemand hat, weiß nur er. Rolle und Geburtsjahr sind
// Sache der planenden Personen, denn daran hängt, was jemand im Haushalt darf
// und welche Aufgaben er bekommen kann.
func (p *Plans) UpdateMember(ctx context.Context, subject, id, mitgliedID string, c planner.MemberChange) (planner.Household, error) {
	zeile, err := p.lookup(ctx, id)
	if err != nil {
		return planner.Household{}, err
	}
	if err := p.mayAccess(ctx, subject, zeile); err != nil {
		return planner.Household{}, err
	}
	if err := c.Validate(); err != nil {
		return planner.Household{}, err
	}

	ich, rolle, err := p.MemberOf(ctx, subject, id)
	if err != nil {
		return planner.Household{}, err
	}
	if ich == "" {
		return planner.Household{}, fmt.Errorf("%w: %q", planner.ErrUnknownHousehold, id)
	}
	planend := rolle == planner.RolePlanner

	if !planend && ich != mitgliedID {
		return planner.Household{}, fmt.Errorf("%w: andere Personen ändern die planenden", planner.ErrNotAllowed)
	}
	if !planend && (c.Role != nil || c.BirthYear != nil) {
		return planner.Household{}, fmt.Errorf("%w: Rolle und Geburtsjahr ändern die planenden", planner.ErrNotAllowed)
	}
	if !c.Touched() {
		return p.household(ctx, zeile)
	}

	kennung, ok := parseUUID(mitgliedID)
	if !ok {
		return planner.Household{}, fmt.Errorf("%w: %q ist keine person", planner.ErrNotAllowed, mitgliedID)
	}

	params := db.UpdateMemberParams{ID: kennung, HouseholdID: zeile.ID}
	if c.Name != nil {
		params.Name = c.Name
	}
	if c.Role != nil {
		wert := string(*c.Role)
		params.Role = &wert
	}
	if c.Minutes != nil {
		minuten := *c.Minutes
		params.CapacityMinutes = int32Felder(minuten[:])
	}
	// Ausdrücklich gesetzte Betreuungsform gewinnt gegen die geratene. Steht
	// sie in derselben Anfrage wie ein neues Geburtsjahr, gilt trotzdem die
	// Angabe — geraten wird nur, wo niemand etwas gesagt hat.
	if c.Care != nil {
		params.SetCare = true
		if *c.Care != planner.CareNone {
			wert := string(*c.Care)
			params.Care = &wert
		}
	}
	if c.BirthYear != nil {
		params.SetBirthYear = true
		if *c.BirthYear != 0 {
			jahr := int32(*c.BirthYear)
			params.BirthYear = &jahr
		}

		// Die Betreuungsform hängt am Alter und wird mitgezogen — außer
		// jemand hat sie ausdrücklich gesetzt. Sie bleibt geraten (ADR-0007),
		// nur eben nach dem neuen Jahr.
		if c.Care == nil {
			params.SetCare = true
			alter := 0
			if *c.BirthYear != 0 {
				alter = time.Now().Year() - *c.BirthYear
			}
			if betreuung := planner.CareForAge(alter); betreuung != planner.CareNone {
				wert := string(betreuung)
				params.Care = &wert
			}
		}
	}

	if _, err := p.db.UpdateMember(ctx, params); err != nil {
		return planner.Household{}, err
	}
	return p.household(ctx, zeile)
}

// Recompute rechnet eine festgeschriebene Woche neu.
//
// Was Spuren hinterlassen hat, bleibt: Aufgaben mit einem Ereignis — abgehakt,
// abgegeben, wieder geöffnet — werden nicht angefasst. Alles andere war ein
// Vorschlag und wird ersetzt.
//
// Das ist keine Vorsicht, sondern das, was die Datenbank zulässt:
// event.task_instance_id ist ON DELETE SET NULL, und der Trigger
// event_kein_update verbietet jedes UPDATE auf event. Eine Aufgabe mit
// Ereignis ließe sich gar nicht löschen. Aus einer Zusage von T5 wird damit
// eine Produktregel — das Protokoll bestimmt, was Bestand hat.
func (p *Plans) Recompute(ctx context.Context, subject, id string, woche planner.Week) (planner.Result, planner.Household, error) {
	zeile, err := p.alsPlanende(ctx, subject, id)
	if err != nil {
		return planner.Result{}, planner.Household{}, err
	}

	haushalt, err := p.household(ctx, zeile)
	if err != nil {
		return planner.Result{}, planner.Household{}, err
	}
	vorlagen, err := p.templates(ctx, zeile.ID)
	if err != nil {
		return planner.Result{}, planner.Household{}, err
	}
	hist, err := p.history(ctx, zeile.ID)
	if err != nil {
		return planner.Result{}, planner.Household{}, err
	}

	frisch, err := planner.Plan(planner.Input{
		Household: haushalt,
		Templates: vorlagen,
		Week:      woche,
		History:   hist,
		Limits:    planner.DefaultLimits(),
	})
	if err != nil {
		return planner.Result{}, haushalt, err
	}

	uebersprungen, err := json.Marshal(frisch.Skipped)
	if err != nil {
		return planner.Result{}, haushalt, err
	}

	tx, err := p.db.Pool.Begin(ctx)
	if err != nil {
		return planner.Result{}, haushalt, err
	}
	defer func() { _ = tx.Rollback(ctx) }()
	q := p.db.Queries.WithTx(tx)

	if err := q.DeleteUntouchedWeekTasks(ctx, db.DeleteUntouchedWeekTasksParams{
		HouseholdID: zeile.ID,
		ISOWeek:     woche.String(),
	}); err != nil {
		return planner.Result{}, haushalt, err
	}

	// Was die Löschung überlebt hat, steht weiter im Plan. Der neue Lauf darf
	// dieselbe Vorlage am selben Tag nicht noch einmal hinlegen — sonst stünde
	// „Bad putzen" zweimal da, einmal abgehakt und einmal offen.
	geblieben, err := q.ListWeekTaskKeys(ctx, db.ListWeekTaskKeysParams{
		HouseholdID: zeile.ID,
		ISOWeek:     woche.String(),
	})
	if err != nil {
		return planner.Result{}, haushalt, err
	}
	belegt := make(map[string]bool, len(geblieben))
	for _, g := range geblieben {
		belegt[g.TemplateID+"|"+datum(g.Day).String()] = true
	}

	for _, t := range frisch.Tasks {
		if belegt[t.TemplateID+"|"+t.Day.String()] {
			continue
		}
		if err := schreibeAufgabe(ctx, q, zeile.ID, woche, t); err != nil {
			return planner.Result{}, haushalt, err
		}
	}

	if err := q.UpdateWeekSkipped(ctx, db.UpdateWeekSkippedParams{
		HouseholdID: zeile.ID,
		ISOWeek:     woche.String(),
		Skipped:     uebersprungen,
	}); err != nil {
		return planner.Result{}, haushalt, err
	}
	if err := tx.Commit(ctx); err != nil {
		return planner.Result{}, haushalt, err
	}

	r, err := p.geschriebeneWoche(ctx, haushalt, zeile.ID, woche, vorlagen)
	return r, haushalt, err
}

// alsPlanende holt den Haushalt und besteht darauf, dass der Aufrufer darin
// plant. Fremder Haushalt ergibt „gibt es nicht", fehlende Rolle „nicht
// erlaubt" — derselbe Unterschied wie überall sonst.
func (p *Plans) alsPlanende(ctx context.Context, subject, id string) (db.Household, error) {
	zeile, err := p.lookup(ctx, id)
	if err != nil {
		return db.Household{}, err
	}
	if err := p.mayAccess(ctx, subject, zeile); err != nil {
		return db.Household{}, err
	}
	_, rolle, err := p.MemberOf(ctx, subject, id)
	if err != nil {
		return db.Household{}, err
	}
	switch rolle {
	case "":
		return db.Household{}, fmt.Errorf("%w: %q", planner.ErrUnknownHousehold, id)
	case planner.RolePlanner:
		return zeile, nil
	default:
		return db.Household{}, fmt.Errorf("%w: das dürfen die planenden Personen", planner.ErrNotAllowed)
	}
}

func int32Felder(in []int) []int32 {
	out := make([]int32, len(in))
	for i, v := range in {
		out[i] = int32(v)
	}
	return out
}

// SetFacts trägt Antworten auf die Fragen zum Haushalt ein.
//
// Zusammengeführt, nicht ersetzt — die Abfrage benutzt dafür den
// JSONB-Operator ||. Ein Formular schickt eine Antwort, nicht den ganzen
// Wissensstand.
//
// Nur planende Personen: Ein Nein nimmt Aufgaben aus dem Plan aller.
func (p *Plans) SetFacts(ctx context.Context, subject, id string, fakten map[string]bool) (planner.Household, error) {
	zeile, err := p.alsPlanende(ctx, subject, id)
	if err != nil {
		return planner.Household{}, err
	}
	if len(fakten) == 0 {
		return p.household(ctx, zeile)
	}

	roh, err := json.Marshal(fakten)
	if err != nil {
		return planner.Household{}, err
	}
	neu, err := p.db.SetFacts(ctx, db.SetFactsParams{ID: zeile.ID, Column2: roh})
	if err != nil {
		return planner.Household{}, err
	}
	return p.household(ctx, neu)
}
