package storage

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/zakaria/haushalt/api/internal/planner"
	"github.com/zakaria/haushalt/api/internal/storage/db"
)

// festschreiben schreibt eine gerechnete Woche in die Datenbank.
//
// Solange ein Plan nur gerechnet wird, hat keine Aufgabe eine Kennung — und
// ohne Kennung kann niemand etwas abhaken oder abgeben. Ab dem ersten
// Ansehen steht die Woche deshalb als Zeilen da.
//
// Der zweite Grund wiegt schwerer als der erste: Ein Plan, der sich beim
// Neuladen ändert, ist kein Plan. Der Planer ist zwar deterministisch, aber
// seine Eingabe ist es nicht — der Verlauf wächst, Gleichstände brechen nach
// Mitglieds-Kennung, und morgen ist ein anderer Tag. Wer am Montag gelesen
// hat, dass er den Müll rausbringt, soll das am Mittwoch noch so vorfinden.
//
// Die ganze Woche in einer Transaktion: Eine halb geschriebene Woche wäre
// schlimmer als gar keine, weil sie nicht mehr nachgerechnet wird.
func (p *Plans) festschreiben(ctx context.Context, haushalt pgtype.UUID, woche planner.Week, r planner.Result) error {
	uebersprungen, err := json.Marshal(r.Skipped)
	if err != nil {
		return err
	}

	tx, err := p.db.Pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback(ctx) }()

	q := p.db.Queries.WithTx(tx)

	// Zuerst die Marke. Kommt keine Zeile zurück, war jemand anderes schneller
	// — dann wird hier nichts geschrieben und der Aufrufer liest, was der
	// erste hinterlassen hat.
	if _, err := q.MarkWeekWritten(ctx, db.MarkWeekWrittenParams{
		HouseholdID: haushalt,
		ISOWeek:     woche.String(),
		Skipped:     uebersprungen,
	}); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil
		}
		return err
	}

	for _, t := range r.Tasks {
		if err := schreibeAufgabe(ctx, q, haushalt, woche, t); err != nil {
			return err
		}
	}

	return tx.Commit(ctx)
}

// schreibeAufgabe legt eine gerechnete Aufgabe samt Zuteilung an.
//
// Ausgelagert, weil zwei Wege sie brauchen: das erste Festschreiben und das
// Neurechnen einer Woche. Zwei Kopien dieser zwanzig Zeilen wären zwei
// Vorstellungen davon, was eine Aufgabe in der Datenbank ist.
func schreibeAufgabe(ctx context.Context, q *db.Queries, haushalt pgtype.UUID, woche planner.Week, t planner.PlannedTask) error {
	kennung, err := q.InsertTaskInstance(ctx, db.InsertTaskInstanceParams{
		HouseholdID: haushalt,
		TemplateID:  t.TemplateID,
		ISOWeek:     woche.String(),
		Day:         tag(t.Day),
		Slot:        string(slotOderEgal(t.Slot)),
		DurationMin: int32(t.DurationMin),
		HeadLoad:    int32(t.HeadLoad),
		Deadline:    tag(t.Deadline),
	})
	if err != nil {
		return fmt.Errorf("aufgabe %q: %w", t.TemplateID, err)
	}

	if t.AssigneeID == "" {
		return nil // niemand zuständig — ein gültiger Zustand
	}
	mitglied, ok := parseUUID(t.AssigneeID)
	if !ok {
		return fmt.Errorf("aufgabe %q: %q ist keine person", t.TemplateID, t.AssigneeID)
	}
	var vorher pgtype.UUID
	if t.Reason.Previous != "" {
		if v, ok := parseUUID(t.Reason.Previous); ok {
			vorher = v
		}
	}
	if err := q.InsertAssignment(ctx, db.InsertAssignmentParams{
		TaskInstanceID: kennung,
		MemberID:       mitglied,
		ReasonCode:     string(t.Reason.Code),
		ReasonPrevious: vorher,
		Manual:         false,
	}); err != nil {
		return fmt.Errorf("zuteilung %q: %w", t.TemplateID, err)
	}
	return nil
}

// geschriebeneWoche liest eine festgeschriebene Woche zurück.
//
// Das Ergebnis sieht aus wie ein gerechnetes — dasselbe planner.Result, damit
// die HTTP-Schicht nicht unterscheiden muss, ob eine Woche gerade entstanden
// ist oder seit Montag steht. Was aus der Vorlage kommt (Titel, Kategorie,
// Art), wird dabei nachgeschlagen statt mitgespeichert: Ändert sich der Titel
// einer Vorlage, soll er sich überall ändern.
func (p *Plans) geschriebeneWoche(
	ctx context.Context,
	haushalt planner.Household,
	kennung pgtype.UUID,
	woche planner.Week,
	vorlagen []planner.TaskTemplate,
) (planner.Result, error) {
	marke, err := p.db.GetWrittenWeek(ctx, db.GetWrittenWeekParams{
		HouseholdID: kennung,
		ISOWeek:     woche.String(),
	})
	if err != nil {
		return planner.Result{}, err
	}

	zeilen, err := p.db.GetWeekPlan(ctx, db.GetWeekPlanParams{
		HouseholdID: kennung,
		ISOWeek:     woche.String(),
	})
	if err != nil {
		return planner.Result{}, err
	}

	erledigt, err := p.erledigte(ctx, kennung, woche)
	if err != nil {
		return planner.Result{}, err
	}

	nachID := make(map[string]planner.TaskTemplate, len(vorlagen))
	for _, v := range vorlagen {
		nachID[v.ID] = v
	}

	out := planner.Result{Week: woche}
	for _, z := range zeilen {
		v := nachID[z.TemplateID]
		t := planner.PlannedTask{
			ID:          formatUUID(z.ID),
			TemplateID:  z.TemplateID,
			Title:       v.Title,
			Category:    v.Category,
			Kind:        v.Kind,
			Failure:     v.Failure,
			Day:         datum(z.Day),
			Slot:        planner.Slot(z.Slot),
			DurationMin: int(z.DurationMin),
			HeadLoad:    planner.HeadLoad(z.HeadLoad),
			Deadline:    datum(z.Deadline),
		}
		t.Done = erledigt[t.ID]
		if z.MemberID.Valid {
			t.AssigneeID = formatUUID(z.MemberID)
		}
		if z.ReasonCode != nil {
			t.Reason.Code = planner.ReasonCode(*z.ReasonCode)
		}
		if z.ReasonPrevious.Valid {
			t.Reason.Previous = formatUUID(z.ReasonPrevious)
		}
		out.Tasks = append(out.Tasks, t)
	}

	// Die Bilanz wird gerechnet, nicht gespeichert: Sie ist eine Sicht auf die
	// Aufgaben, und eine gespeicherte Summe neben den Posten läuft irgendwann
	// auseinander.
	out.Balance = planner.BalanceOf(haushalt, out.Tasks, planner.DefaultLimits())

	if len(marke.Skipped) > 0 {
		if err := json.Unmarshal(marke.Skipped, &out.Skipped); err != nil {
			return planner.Result{}, fmt.Errorf("übersprungene aufgaben der woche %s: %w", woche, err)
		}
	}
	return out, nil
}

// erledigte sagt je Aufgabe, ob sie zuletzt abgehakt oder wieder geöffnet
// wurde.
func (p *Plans) erledigte(ctx context.Context, haushalt pgtype.UUID, woche planner.Week) (map[string]bool, error) {
	zeilen, err := p.db.ListDoneTasks(ctx, db.ListDoneTasksParams{
		HouseholdID: haushalt,
		ISOWeek:     woche.String(),
	})
	if err != nil {
		return nil, err
	}
	out := make(map[string]bool, len(zeilen))
	for _, z := range zeilen {
		if z.TaskInstanceID.Valid {
			out[formatUUID(z.TaskInstanceID)] = z.Kind == "erledigt"
		}
	}
	return out, nil
}

func slotOderEgal(s planner.Slot) planner.Slot {
	if s == "" {
		return planner.SlotAny
	}
	return s
}
