package storage

import (
	"context"
	"fmt"

	"github.com/zakaria/haushalt/api/internal/planner"
	"github.com/zakaria/haushalt/api/internal/storage/db"
)

// Occasions sind die eingetragenen Anlässe eines Haushalts.
func (p *Plans) Occasions(ctx context.Context, subject, id string) ([]planner.Occasion, error) {
	zeile, err := p.lookup(ctx, id)
	if err != nil {
		return nil, err
	}
	if err := p.mayAccess(ctx, subject, zeile); err != nil {
		return nil, err
	}
	haushalt, err := p.household(ctx, zeile)
	if err != nil {
		return nil, err
	}
	return haushalt.Occasions, nil
}

// AddOccasion trägt einen Anlass ein.
//
// Nur planende Personen: Ein Anlass erzeugt Aufgaben für andere.
func (p *Plans) AddOccasion(ctx context.Context, subject, id string, o planner.Occasion) (planner.Occasion, error) {
	zeile, err := p.alsPlanende(ctx, subject, id)
	if err != nil {
		return planner.Occasion{}, err
	}
	if err := o.Validate(); err != nil {
		return planner.Occasion{}, err
	}

	neu, err := p.db.CreateOccasion(ctx, db.CreateOccasionParams{
		HouseholdID: zeile.ID,
		Title:       o.Title,
		Day:         tag(o.Date),
		Kind:        o.Kind,
		Yearly:      o.Yearly,
	})
	if err != nil {
		return planner.Occasion{}, err
	}
	return planner.Occasion{
		ID:     formatUUID(neu.ID),
		Title:  neu.Title,
		Date:   datum(neu.Day),
		Kind:   neu.Kind,
		Yearly: neu.Yearly,
	}, nil
}

// RemoveOccasion löscht einen Anlass.
func (p *Plans) RemoveOccasion(ctx context.Context, subject, id, anlassID string) error {
	zeile, err := p.alsPlanende(ctx, subject, id)
	if err != nil {
		return err
	}
	kennung, ok := parseUUID(anlassID)
	if !ok {
		return fmt.Errorf("%w: %q ist kein anlass", planner.ErrNotAllowed, anlassID)
	}
	return p.db.DeleteOccasion(ctx, db.DeleteOccasionParams{
		ID:          kennung,
		HouseholdID: zeile.ID,
	})
}
