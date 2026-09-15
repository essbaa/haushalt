package storage

import (
	"context"
	"encoding/json"

	"github.com/jackc/pgx/v5/pgtype"

	"github.com/zakaria/haushalt/api/internal/planner"
	"github.com/zakaria/haushalt/api/internal/storage/db"
)

// ApplyAgreement trägt das Wochenraster einer Vorlage in die laufende Woche
// ein — ab heute.
//
// Die Woche steht fest, sobald sie das erste Mal angesehen wurde (ADR-0008).
// Eine neue Absprache gilt deshalb ab der nächsten Woche von selbst, und für
// „ab sofort" braucht es einen Handgriff. Der naheliegende wäre das vorhandene
// Neurechnen — und es ist das falsche Werkzeug:
//
// Neurechnen verwirft alles, woran nichts hängt, und verteilt die ganze Woche
// neu. Wer am Mittwoch die Kita-Absprache einträgt, bekäme also nebenbei eine
// neue Antwort auf die Frage, wer am Freitag das Bad putzt. Das ist nicht,
// wonach gefragt wurde — und ein Plan, der sich an Stellen ändert, die man
// gar nicht angefasst hat, ist genau der Plan, dem man nicht mehr glaubt.
//
// Also eng: nur diese Vorlage, nur ab heute, und nur dort, wo noch nichts
// geschehen ist. Was Spuren hat, bleibt stehen — dieselbe Regel wie beim
// Neurechnen. Vergangene Tage bleiben unberührt: Ein Kita-Termin von gestern
// früh ist keine Verabredung mehr, die man noch treffen könnte.
//
// Zurück kommt, wie viele Termine entstanden sind. Null ist eine gültige
// Antwort — an den Tagen ist entweder schon etwas eingetragen, oder das
// Raster trifft in dieser Woche keinen fälligen Tag mehr. Die Oberfläche
// braucht die Zahl, um genau das sagen zu können, statt „fertig" zu melden
// und nichts zu zeigen.
func (p *Plans) ApplyAgreement(ctx context.Context, subject, id, vorlageID string) (int, error) {
	zeile, err := p.alsPlanende(ctx, subject, id)
	if err != nil {
		return 0, err
	}

	haushalt, err := p.household(ctx, zeile)
	if err != nil {
		return 0, err
	}
	vorlagen, err := p.templates(ctx, zeile.ID)
	if err != nil {
		return 0, err
	}
	hist, err := p.history(ctx, zeile.ID)
	if err != nil {
		return 0, err
	}

	woche := aktuelleWoche(zeile.Timezone)
	heute := heuteIn(zeile.Timezone)

	// Gerechnet wird der ganze Plan, benutzt wird eine Vorlage daraus. Die
	// Alternative wäre, Fälligkeit und Rhythmus hier noch einmal auszurechnen
	// — eine zweite Vorstellung davon, wann eine Aufgabe ansteht, und die
	// zweite ist immer die, die niemand pflegt.
	frisch, err := planner.Plan(planner.Input{
		Household: haushalt,
		Templates: vorlagen,
		Week:      woche,
		History:   hist,
		Limits:    planner.DefaultLimits(),
	})
	if err != nil {
		return 0, err
	}

	tx, err := p.db.Pool.Begin(ctx)
	if err != nil {
		return 0, err
	}
	defer func() { _ = tx.Rollback(ctx) }()
	q := p.db.Queries.WithTx(tx)

	if err := q.DeleteUntouchedTemplateTasksFrom(ctx, db.DeleteUntouchedTemplateTasksFromParams{
		HouseholdID: zeile.ID,
		ISOWeek:     woche.String(),
		TemplateID:  vorlageID,
		Day:         tag(heute),
	}); err != nil {
		return 0, err
	}

	tage, err := q.ListTemplateTaskDays(ctx, db.ListTemplateTaskDaysParams{
		HouseholdID: zeile.ID,
		ISOWeek:     woche.String(),
		TemplateID:  vorlageID,
	})
	if err != nil {
		return 0, err
	}
	belegt := make(map[string]bool, len(tage))
	for _, t := range tage {
		belegt[datum(t).String()] = true
	}

	anzahl := 0
	for _, t := range frisch.Tasks {
		if t.TemplateID != vorlageID || t.Day.Before(heute) || belegt[t.Day.String()] {
			continue
		}
		if err := schreibeAufgabe(ctx, q, zeile.ID, woche, t); err != nil {
			return 0, err
		}
		anzahl++
	}

	// Die Liste „nicht im Plan" ist gespeichert und nicht gerechnet. Bleibt
	// der Eintrag „braucht Absprache" stehen, während die Termine schon im
	// Plan sind, widerspricht sich die Seite selbst — und der Widerspruch
	// steht ausgerechnet an der Stelle, an der man gerade nachgesehen hat.
	//
	// Entfernt wird nur, was diese Vorlage betrifft. Die übrigen Gründe
	// erklären die Woche, wie sie festgeschrieben wurde, und die hat sich
	// nicht geändert.
	if anzahl > 0 {
		if err := ohneAbspracheHinweis(ctx, q, zeile.ID, woche, vorlageID); err != nil {
			return 0, err
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return 0, err
	}
	return anzahl, nil
}

// ohneAbspracheHinweis streicht die Übersprungen-Einträge dieser Vorlage aus
// der festgeschriebenen Woche.
func ohneAbspracheHinweis(
	ctx context.Context,
	q *db.Queries,
	haushalt pgtype.UUID,
	woche planner.Week,
	vorlageID string,
) error {
	marke, err := q.GetWrittenWeek(ctx, db.GetWrittenWeekParams{
		HouseholdID: haushalt,
		ISOWeek:     woche.String(),
	})
	if err != nil {
		return err
	}
	if len(marke.Skipped) == 0 {
		return nil
	}

	var alt []planner.Skipped
	if err := json.Unmarshal(marke.Skipped, &alt); err != nil {
		return err
	}
	neu := make([]planner.Skipped, 0, len(alt))
	for _, s := range alt {
		if s.TemplateID == vorlageID && s.Code == planner.SkipNeedsAgreement {
			continue
		}
		neu = append(neu, s)
	}
	if len(neu) == len(alt) {
		return nil
	}

	roh, err := json.Marshal(neu)
	if err != nil {
		return err
	}
	return q.UpdateWeekSkipped(ctx, db.UpdateWeekSkippedParams{
		HouseholdID: haushalt,
		ISOWeek:     woche.String(),
		Skipped:     roh,
	})
}
