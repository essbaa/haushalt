package storage

import (
	"context"
	"time"

	"github.com/zakaria/haushalt/api/internal/planner"
	"github.com/zakaria/haushalt/api/internal/storage/db"
)

// Create legt einen Haushalt samt Mitgliedern an und verbindet den Aufrufer
// mit der ersten Person darin.
//
// In einer Transaktion, und das ist hier nicht die übliche Vorsicht: Ein
// Haushalt ohne Mitglieder wäre nicht bloß unvollständig, sondern unerreichbar
// — der Zugriff läuft immer über member, es gäbe also niemanden mehr, der ihn
// je wieder sehen oder löschen könnte. Eine abgebrochene Einrichtung darf
// keine Leiche hinterlassen.
func (p *Plans) Create(ctx context.Context, subject string, eingabe planner.Setup) (planner.Household, error) {
	if subject == "" {
		return planner.Household{}, planner.ErrNotAllowed
	}

	eingabe = eingabe.Normalized()
	if err := eingabe.Validate(); err != nil {
		return planner.Household{}, err
	}

	tx, err := p.db.Pool.Begin(ctx)
	if err != nil {
		return planner.Household{}, err
	}
	// Nach einem erfolgreichen Commit ist das ein Nichtstun.
	defer func() { _ = tx.Rollback(ctx) }()

	q := p.db.Queries.WithTx(tx)

	haushalt, err := q.CreateHousehold(ctx, db.CreateHouseholdParams{
		Name:     eingabe.Name,
		Home:     string(eingabe.Context.Home),
		HasCar:   eingabe.Context.HasCar,
		HasYard:  eingabe.Context.HasYard,
		Pets:     eingabe.Context.Pets,
		Timezone: eingabe.Timezone,
		// Kein Slug: Der gehört den Beispielhaushalten aus dem Repo. Dieser
		// hier ist echt und wird über seine Kennung angesprochen.
	})
	if err != nil {
		return planner.Household{}, err
	}

	jahr := time.Now().Year()
	for i, m := range eingabe.Members {
		params := db.CreateMemberParams{
			HouseholdID:     haushalt.ID,
			Name:            m.Name,
			Role:            string(m.Role),
			CapacityMinutes: minutenFelder(m.Budget),
		}

		// Nur die erste Person bekommt die Anmeldung. Die übrigen sind
		// zunächst Namen im Plan; wer selbst hineinsehen will, braucht eine
		// Einladung — auch das Kind, das mit am Tisch sitzt.
		if i == 0 {
			params.AuthUserID = &subject
		}

		if m.BirthYear != 0 {
			jahre := int32(m.BirthYear)
			params.BirthYear = &jahre

			if betreuung := planner.CareForAge(jahr - m.BirthYear); betreuung != planner.CareNone {
				wert := string(betreuung)
				params.Care = &wert
			}
		}

		if _, err := q.CreateMember(ctx, params); err != nil {
			return planner.Household{}, err
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return planner.Household{}, err
	}
	return p.household(ctx, haushalt)
}

// minutenFelder übersetzt die Zeitstufe in das, was die Spalte erwartet.
func minutenFelder(b planner.TimeBudget) []int32 {
	minuten := b.Minutes()
	out := make([]int32, len(minuten))
	for i, v := range minuten {
		out[i] = int32(v)
	}
	return out
}
