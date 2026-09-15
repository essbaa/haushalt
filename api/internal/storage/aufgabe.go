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

// Reassign gibt eine Aufgabe von Hand an eine bestimmte Person.
//
// Der Unterschied zu HandOver ist die Richtung: Dort sucht der Planer jemanden
// nach seinen Regeln, hier hat ein Mensch schon gewählt. Der Planer prüft
// deshalb nur noch, was diese Person nicht übernehmen *kann* — Alter, Rolle,
// Verteilungsregel, eigene Aufgabe — und nicht, was er selbst bevorzugt hätte
// (planner.Reassign, ADR-0012).
//
// `manual = true` setzt SetAssignment ohnehin. Das ist keine Nebenwirkung,
// sondern der Zweck der Übung: Jede Korrektur von Hand sagt, wo der Planer
// danebenlag, und diese Spalte ist die einzige Stelle, an der das später
// nachzulesen ist.
//
// Zurückgegeben wird der Name der Person, die jetzt zuständig ist.
func (p *Plans) Reassign(ctx context.Context, subject, aufgabeID, mitgliedID string) (string, error) {
	aufgabe, err := p.zustaendig(ctx, subject, aufgabeID)
	if err != nil {
		return "", err
	}
	ziel, ok := parseUUID(mitgliedID)
	if !ok {
		return "", planner.ErrUnknownMember
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

	grund, err := planner.Reassign(planner.Input{
		Household: haushalt,
		Templates: vorlagen,
		Week:      woche,
		History:   hist,
		Limits:    planner.DefaultLimits(),
	}, stand.Tasks, formatUUID(aufgabe.ID), formatUUID(ziel))
	if err != nil {
		return "", err
	}

	name := ""
	for _, m := range haushalt.Members {
		if m.ID == formatUUID(ziel) {
			name = m.Name
		}
	}

	// Schon dort. Kein Fehler und kein Ereignis — ein Protokolleintrag, der
	// nichts festhält, macht die Auswertung später schlechter, nicht besser.
	if aufgabe.AssigneeID.Valid && aufgabe.AssigneeID == ziel {
		return name, nil
	}

	tx, err := p.db.Pool.Begin(ctx)
	if err != nil {
		return "", err
	}
	defer func() { _ = tx.Rollback(ctx) }()
	q := p.db.Queries.WithTx(tx)

	// Erst weg, dann neu — es gibt genau eine Zuteilung je Aufgabe, und eine
	// abgegebene hat gar keine.
	if err := q.ClearAssignment(ctx, aufgabe.ID); err != nil {
		return "", err
	}
	if err := q.SetAssignment(ctx, db.SetAssignmentParams{
		TaskInstanceID: aufgabe.ID,
		MemberID:       ziel,
		ReasonCode:     string(grund.Code),
		ReasonPrevious: aufgabe.AssigneeID,
	}); err != nil {
		return "", err
	}

	nutzlast, err := json.Marshal(map[string]string{
		"von": formatUUID(aufgabe.AssigneeID),
		"an":  formatUUID(ziel),
	})
	if err != nil {
		return "", err
	}
	if _, err := q.AppendEvent(ctx, db.AppendEventParams{
		HouseholdID:    aufgabe.HouseholdID,
		MemberID:       aufgabe.MemberID,
		TaskInstanceID: aufgabe.ID,
		Kind:           "umverteilt",
		Payload:        nutzlast,
	}); err != nil {
		return "", err
	}

	if err := tx.Commit(ctx); err != nil {
		return "", err
	}
	return name, nil
}

// Strike streicht eine Aufgabe für diese Woche — oder plant sie wieder ein.
//
// Der Unterschied zu „Brauchen wir nicht" ist die Reichweite, und er ist der
// ganze Grund für diese Funktion: Das Abschalten einer Vorlage gilt für alle
// künftigen Wochen, das hier nur für diese eine. Bis es das gab, war der
// einzige Ausweg aus einer Aufgabe eine Entscheidung über die Zukunft — und
// ein Fehlgriff nahm dem Haushalt eine Vorlage weg, die er eigentlich wollte.
//
// Geschrieben wird ein Ereignis, keine Spalte. Wie bei „erledigt": Nur das
// Protokoll kann erzählen, dass jemand es sich anders überlegt hat.
func (p *Plans) Strike(ctx context.Context, subject, aufgabeID string, gestrichen bool) error {
	aufgabe, err := p.zustaendig(ctx, subject, aufgabeID)
	if err != nil {
		return err
	}

	art := "gestrichen"
	if !gestrichen {
		art = "wieder_eingeplant"
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
