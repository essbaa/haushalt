package storage

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"

	"github.com/zakaria/haushalt/api/internal/planner"
	"github.com/zakaria/haushalt/api/internal/storage/db"
)

// MarkDone hakt eine Aufgabe ab — oder nimmt das Häkchen zurück.
//
// Geschrieben wird ein Ereignis, keine Spalte auf der Aufgabe. Der Unterschied
// ist nicht Geschmack: Eine Spalte kennt nur den letzten Zustand, das
// Protokoll kennt den Weg dahin. Die Rotation der Folgewoche liest genau
// diesen Weg, und „am Dienstag erledigt" ist etwas anderes als „irgendwann".
func (p *Plans) MarkDone(ctx context.Context, subject, aufgabeID string, erledigt bool) error {
	aufgabe, err := p.zustaendig(ctx, subject, aufgabeID)
	if err != nil {
		return err
	}

	art := "erledigt"
	if !erledigt {
		art = "wieder_geoeffnet"
	}
	_, err = p.db.AppendEvent(ctx, db.AppendEventParams{
		HouseholdID:    aufgabe.HouseholdID,
		MemberID:       aufgabe.MemberID,
		TaskInstanceID: aufgabe.ID,
		Kind:           art,
		Payload:        []byte(`{}`),
	})
	return err
}

// HandOver gibt eine Aufgabe zurück in den Haushalt.
//
// Zurückgegeben wird der Name der Person, die übernimmt — leer, wenn niemand
// geeignet ist. Das ist kein Fehler: Dann steht die Aufgabe offen da, sichtbar
// für alle. Eine Aufgabe, die beim Abgeben verschwindet, wäre Löschen mit
// besserem Gewissen.
func (p *Plans) HandOver(ctx context.Context, subject, aufgabeID, grund string) (string, error) {
	aufgabe, err := p.zustaendig(ctx, subject, aufgabeID)
	if err != nil {
		return "", err
	}
	if !aufgabe.AssigneeID.Valid {
		// Schon abgegeben. Zweimal abgeben ist kein Fehler, aber auch kein
		// Ereignis.
		return "", nil
	}

	zeile, err := p.db.GetHousehold(ctx, aufgabe.HouseholdID)
	if err != nil {
		return "", err
	}
	haushalt, err := p.household(ctx, zeile)
	if err != nil {
		return "", err
	}
	vorlagen, err := p.templates(ctx, zeile.ID)
	if err != nil {
		return "", err
	}
	hist, err := p.history(ctx, zeile.ID)
	if err != nil {
		return "", err
	}
	woche, err := planner.ParseWeek(aufgabe.ISOWeek)
	if err != nil {
		return "", err
	}
	stand, err := p.geschriebeneWoche(ctx, haushalt, zeile.ID, woche, vorlagen)
	if err != nil {
		return "", err
	}

	nachfolger, begruendung, gefunden := planner.Handover(planner.Input{
		Household: haushalt,
		Templates: vorlagen,
		Week:      woche,
		History:   hist,
		Limits:    planner.DefaultLimits(),
	}, stand.Tasks, formatUUID(aufgabe.ID))

	tx, err := p.db.Pool.Begin(ctx)
	if err != nil {
		return "", err
	}
	defer func() { _ = tx.Rollback(ctx) }()
	q := p.db.Queries.WithTx(tx)

	// Erst weg, dann neu: Es gibt genau eine Zuteilung je Aufgabe, und der
	// Zwischenzustand „niemand" ist der gültige Zustand einer abgegebenen
	// Aufgabe.
	if err := q.ClearAssignment(ctx, aufgabe.ID); err != nil {
		return "", err
	}

	name := ""
	if gefunden {
		neu, ok := parseUUID(nachfolger)
		if !ok {
			return "", fmt.Errorf("abgabe: %q ist keine person", nachfolger)
		}
		var vorher = aufgabe.AssigneeID
		if err := q.SetAssignment(ctx, db.SetAssignmentParams{
			TaskInstanceID: aufgabe.ID,
			MemberID:       neu,
			ReasonCode:     string(begruendung.Code),
			ReasonPrevious: vorher,
		}); err != nil {
			return "", err
		}
		for _, m := range haushalt.Members {
			if m.ID == nachfolger {
				name = m.Name
			}
		}
	}

	nutzlast, err := json.Marshal(map[string]string{
		"grund": grund,
		"von":   formatUUID(aufgabe.AssigneeID),
		"an":    nachfolger,
	})
	if err != nil {
		return "", err
	}
	if _, err := q.AppendEvent(ctx, db.AppendEventParams{
		HouseholdID:    aufgabe.HouseholdID,
		MemberID:       aufgabe.MemberID,
		TaskInstanceID: aufgabe.ID,
		Kind:           "abgegeben",
		Payload:        nutzlast,
	}); err != nil {
		return "", err
	}

	if err := tx.Commit(ctx); err != nil {
		return "", err
	}
	return name, nil
}

// zustaendig holt die Aufgabe und prüft in einem Zug, ob dieser Aufrufer sie
// anfassen darf.
//
// Erlaubt ist es der zuständigen Person und jeder planenden. Die erste, weil
// es ihre Aufgabe ist; die zweite, weil jemand den Überblick behalten muss,
// wenn ein Kind sein Handy nicht anfasst.
//
// Eine Aufgabe aus einem fremden Haushalt gibt es nicht — nicht „verboten",
// sondern nicht vorhanden. Dieselbe Regel wie beim Haushalt selbst: Ein 403
// verriete, dass es sie gibt.
func (p *Plans) zustaendig(ctx context.Context, subject, aufgabeID string) (db.GetTaskForMemberRow, error) {
	if subject == "" {
		return db.GetTaskForMemberRow{}, planner.ErrNotAllowed
	}
	kennung, ok := parseUUID(aufgabeID)
	if !ok {
		return db.GetTaskForMemberRow{}, planner.ErrUnknownTask
	}

	zeile, err := p.db.GetTaskForMember(ctx, db.GetTaskForMemberParams{
		ID:         kennung,
		AuthUserID: &subject,
	})
	if errors.Is(err, pgx.ErrNoRows) {
		return db.GetTaskForMemberRow{}, planner.ErrUnknownTask
	}
	if err != nil {
		return db.GetTaskForMemberRow{}, err
	}

	eigene := zeile.AssigneeID.Valid && zeile.AssigneeID == zeile.MemberID
	if !eigene && planner.Role(zeile.MemberRole) != planner.RolePlanner {
		return db.GetTaskForMemberRow{}, fmt.Errorf("%w: das ist nicht deine aufgabe", planner.ErrNotAllowed)
	}
	return zeile, nil
}
