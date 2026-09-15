package storage

import (
	"context"
	"fmt"
	"slices"
	"strings"

	"github.com/jackc/pgx/v5/pgtype"

	"github.com/zakaria/haushalt/api/internal/planner"
	"github.com/zakaria/haushalt/api/internal/storage/db"
)

// AddFeedback nimmt eine Rückmeldung aus der App entgegen.
//
// **Jedes Mitglied darf melden, nicht nur die planenden.** Das ist keine
// Nachlässigkeit bei der Rechteprüfung, sondern der Zweck: Die ausführenden
// Personen sehen die Stellen, an denen es klemmt, zuerst — sie bekommen den
// Plan, den sie nicht gemacht haben. Eine Meldefunktion, die nur den
// Planenden offensteht, meldet, was der Planende ohnehin weiß.
//
// Wer nicht zum Haushalt gehört, meldet trotzdem nichts: mayAccess gilt. Und
// Demo-Haushalte sind ausgenommen — sonst schreibt jeder Vorbeikommende in
// eine Tabelle, die niemandem gehört.
func (p *Plans) AddFeedback(ctx context.Context, subject, id string, art planner.FeedbackKind, text, kontext string) error {
	if !slices.Contains(planner.AllFeedbackKinds, art) {
		return fmt.Errorf("%w: %q ist keine Art von Rückmeldung", planner.ErrInvalidSetup, art)
	}
	text = strings.TrimSpace(text)
	if text == "" {
		return fmt.Errorf("%w: eine leere Rückmeldung sagt nichts", planner.ErrInvalidSetup)
	}
	if len(text) > 2000 {
		text = text[:2000]
	}
	if len(kontext) > 500 {
		kontext = kontext[:500]
	}

	zeile, err := p.lookup(ctx, id)
	if err != nil {
		return err
	}
	if err := p.mayAccess(ctx, subject, zeile); err != nil {
		return err
	}
	// Ein Demo-Haushalt gehört niemandem; eine Rückmeldung darin hätte keinen
	// Absender und keinen Empfänger.
	if zeile.Slug != nil && *zeile.Slug != "" {
		return fmt.Errorf("%w: in einem Beispielhaushalt gibt es nichts zu melden", planner.ErrNotAllowed)
	}

	mitgliedID, _, err := p.MemberOf(ctx, subject, id)
	if err != nil {
		return err
	}
	var wer pgtype.UUID
	if kennung, ok := parseUUID(mitgliedID); ok {
		wer = kennung
	}

	return p.db.InsertFeedback(ctx, db.InsertFeedbackParams{
		HouseholdID: zeile.ID,
		MemberID:    wer,
		Art:         string(art),
		Text:        text,
		Kontext:     kontext,
	})
}

// Feedback ist, was dieser Haushalt gemeldet hat — für die planenden Personen.
//
// Lesen dürfen nur sie, melden alle. Das ist Absicht und keine Ungleichheit:
// Die Liste enthält Sätze über andere Menschen im selben Haushalt („sie
// bekommt immer die Küche"), und die gehören nicht an jede Pinnwand.
//
// Dass die Liste überhaupt sichtbar ist, gehört zur Sache: Ein Formular, das
// schluckt und nie etwas zeigt, ist schlechter als der Zettel am Kühlschrank
// — beim Zettel sieht man wenigstens, dass er voll wird.
func (p *Plans) Feedback(ctx context.Context, subject, id string) ([]planner.Feedback, error) {
	zeile, err := p.alsPlanende(ctx, subject, id)
	if err != nil {
		return nil, err
	}

	zeilen, err := p.db.ListFeedback(ctx, zeile.ID)
	if err != nil {
		return nil, err
	}

	out := make([]planner.Feedback, 0, len(zeilen))
	for _, z := range zeilen {
		eintrag := planner.Feedback{
			ID:      formatUUID(z.ID),
			Kind:    planner.FeedbackKind(z.Art),
			Text:    z.Text,
			Context: z.Kontext,
			At:      z.CreatedAt.Time,
		}
		if z.Wer != nil {
			eintrag.Who = *z.Wer
		}
		out = append(out, eintrag)
	}
	return out, nil
}
