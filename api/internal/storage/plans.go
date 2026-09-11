package storage

import (
	"context"
	"encoding/hex"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/zakaria/haushalt/api/internal/library"
	"github.com/zakaria/haushalt/api/internal/planner"
	"github.com/zakaria/haushalt/api/internal/storage/db"
)

// Plans liest Haushalte und Wochenpläne aus der Datenbank.
//
// Dieselbe Schnittstelle, die der Katalog aus dem Repo erfüllt. Für die
// HTTP-Schicht ist der Unterschied nicht sichtbar — sie bekommt Haushalte und
// Pläne und weiß nicht, ob sie aus einer Datei oder aus Postgres kommen.
//
// Gerechnet wird weiterhin im Planer. Dieses Paket holt Daten und reicht sie
// weiter; es entscheidet nichts.
type Plans struct{ db *DB }

// AsPlans gibt die Datenbank als Planquelle aus.
func (d *DB) AsPlans() *Plans { return &Plans{db: d} }

func (p *Plans) Households(ctx context.Context) ([]planner.Household, error) {
	zeilen, err := p.db.ListHouseholds(ctx)
	if err != nil {
		return nil, err
	}
	out := make([]planner.Household, 0, len(zeilen))
	for _, z := range zeilen {
		h, err := p.household(ctx, z)
		if err != nil {
			return nil, err
		}
		out = append(out, h)
	}
	return out, nil
}

func (p *Plans) Plan(ctx context.Context, id string, week planner.Week) (planner.Result, planner.Household, error) {
	zeile, err := p.lookup(ctx, id)
	if err != nil {
		return planner.Result{}, planner.Household{}, err
	}
	haushalt, err := p.household(ctx, zeile)
	if err != nil {
		return planner.Result{}, planner.Household{}, err
	}
	templates, err := p.templates(ctx, zeile.ID)
	if err != nil {
		return planner.Result{}, planner.Household{}, err
	}
	hist, err := p.history(ctx, zeile.ID)
	if err != nil {
		return planner.Result{}, planner.Household{}, err
	}

	result, err := planner.Plan(planner.Input{
		Household: haushalt,
		Templates: templates,
		Week:      week,
		History:   hist,
		Limits:    planner.DefaultLimits(),
	})
	return result, haushalt, err
}

// lookup findet einen Haushalt über den Slug und, falls das nichts ergibt,
// über die Kennung.
//
// Beides zuzulassen ist Absicht: Demo-Haushalte haben sprechende Adressen
// (/api/haushalte/familie-a/…), echte haben nur ihre UUID. Ein Client muss den
// Unterschied nicht kennen.
func (p *Plans) lookup(ctx context.Context, id string) (db.Household, error) {
	zeile, err := p.db.GetHouseholdBySlug(ctx, &id)
	switch {
	case err == nil:
		return zeile, nil
	case !errors.Is(err, pgx.ErrNoRows):
		return db.Household{}, err
	}

	uid, ok := parseUUID(id)
	if !ok {
		return db.Household{}, fmt.Errorf("%w: %q", planner.ErrUnknownHousehold, id)
	}
	zeile, err = p.db.GetHousehold(ctx, uid)
	if errors.Is(err, pgx.ErrNoRows) {
		return db.Household{}, fmt.Errorf("%w: %q", planner.ErrUnknownHousehold, id)
	}
	return zeile, err
}

func (p *Plans) household(ctx context.Context, z db.Household) (planner.Household, error) {
	mitglieder, err := p.db.ListMembers(ctx, z.ID)
	if err != nil {
		return planner.Household{}, err
	}

	h := planner.Household{
		ID:   publicID(z),
		Name: z.Name,
		Context: planner.Context{
			Home:    planner.Home(z.Home),
			HasCar:  z.HasCar,
			HasYard: z.HasYard,
			Pets:    z.Pets,
		},
	}
	jahr := time.Now().Year()
	for _, m := range mitglieder {
		person := planner.Member{
			ID:   formatUUID(m.ID),
			Name: m.Name,
			Role: planner.Role(m.Role),
		}
		// Das Modell führt Geburtsjahre, der Planer rechnet mit Alter. Die
		// Umrechnung passiert hier und nirgends sonst.
		if m.BirthYear != nil {
			person.Age = jahr - int(*m.BirthYear)
		}
		if m.Care != nil {
			person.Care = planner.Care(*m.Care)
		}
		for i, v := range m.CapacityMinutes {
			if i < len(person.CapacityMinutes) {
				person.CapacityMinutes[i] = int(v)
			}
		}
		h.Members = append(h.Members, person)
	}
	return h, nil
}

func (p *Plans) templates(ctx context.Context, haushalt pgtype.UUID) ([]planner.TaskTemplate, error) {
	zeilen, err := p.db.ListTemplatesForHousehold(ctx, haushalt)
	if err != nil {
		return nil, err
	}
	out := make([]planner.TaskTemplate, 0, len(zeilen))
	for _, z := range zeilen {
		// Derselbe Parser wie beim Einlesen der Datei. Ein zweiter wäre die
		// sicherste Art, dass beide Wege auseinanderlaufen.
		t, err := library.ParseTemplate(z.Definition)
		if err != nil {
			return nil, fmt.Errorf("vorlage %q: %w", z.ID, err)
		}
		out = append(out, t)
	}
	return out, nil
}

// history verdichtet, was der Planer über die Vergangenheit wissen muss.
//
// Drei Abfragen, drei Quellen: erledigt kommt aus dem Protokoll, zuletzt
// zugeteilt aus den Zuteilungen, abgewählt aus den Lernwerten. Der Planer
// selbst liest nichts davon — er bekommt das Ergebnis.
func (p *Plans) history(ctx context.Context, haushalt pgtype.UUID) (planner.History, error) {
	hist := planner.History{
		LastDone:     map[string]planner.Date{},
		LastAssignee: map[string]string{},
		Muted:        map[string]bool{},
		FixedTo:      map[string]string{},
	}

	erledigt, err := p.db.LastDonePerTemplate(ctx, haushalt)
	if err != nil {
		return hist, err
	}
	for _, z := range erledigt {
		if z.LastDone.Valid {
			hist.LastDone[z.TemplateID] = planner.DateOf(z.LastDone.Time)
		}
	}

	zugeteilt, err := p.db.LastAssigneePerTemplate(ctx, haushalt)
	if err != nil {
		return hist, err
	}
	for _, z := range zugeteilt {
		if z.MemberID.Valid {
			hist.LastAssignee[z.TemplateID] = formatUUID(z.MemberID)
		}
	}

	signale, err := p.db.ListSignals(ctx, haushalt)
	if err != nil {
		return hist, err
	}
	for _, z := range signale {
		if z.Kind == "abgewaehlt" && z.Value != 0 {
			hist.Muted[z.TemplateID] = true
		}
	}
	return hist, nil
}

// ------------------------------------------------------------------ UUID

// publicID ist der Slug, wenn es einen gibt, sonst die Kennung.
func publicID(z db.Household) string {
	if z.Slug != nil && *z.Slug != "" {
		return *z.Slug
	}
	return formatUUID(z.ID)
}

// formatUUID schreibt die sechzehn Bytes in die übliche Form.
//
// pgtype.UUID hat dafür keine Methode, und eine Abhängigkeit nur für sechs
// Zeilen Hexadezimal wäre übertrieben.
func formatUUID(u pgtype.UUID) string {
	if !u.Valid {
		return ""
	}
	h := hex.EncodeToString(u.Bytes[:])
	return h[0:8] + "-" + h[8:12] + "-" + h[12:16] + "-" + h[16:20] + "-" + h[20:32]
}

func parseUUID(s string) (pgtype.UUID, bool) {
	var u pgtype.UUID
	roh := make([]byte, 0, 32)
	for i := 0; i < len(s); i++ {
		if s[i] != '-' {
			roh = append(roh, s[i])
		}
	}
	if len(roh) != 32 {
		return u, false
	}
	b, err := hex.DecodeString(string(roh))
	if err != nil {
		return u, false
	}
	copy(u.Bytes[:], b)
	u.Valid = true
	return u, true
}
