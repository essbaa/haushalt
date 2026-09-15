package storage

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"

	"github.com/zakaria/haushalt/api/internal/library"
	"github.com/zakaria/haushalt/api/internal/planner"
	"github.com/zakaria/haushalt/api/internal/storage/db"
)

// signalAbgewaehlt ist die Art des Lernwerts, mit dem ein Haushalt eine
// Vorlage dauerhaft abbestellt.
const signalAbgewaehlt = "abgewaehlt"

// TemplatesFor listet die ganze Bibliothek mit dem Stand dieses Haushalts.
//
// Die Seite dahinter ist die Antwort auf einen berechtigten Einwand: Ein Plan,
// dem man nicht widersprechen kann, ist eine Behauptung. Hier steht alles, was
// die App kennt, was davon bei euch gilt und was nicht — und jede Zeile lässt
// sich umdrehen.
func (p *Plans) TemplatesFor(ctx context.Context, subject, id string) ([]planner.TemplateState, planner.Household, error) {
	zeile, err := p.lookup(ctx, id)
	if err != nil {
		return nil, planner.Household{}, err
	}
	if err := p.mayAccess(ctx, subject, zeile); err != nil {
		return nil, planner.Household{}, err
	}

	haushalt, err := p.household(ctx, zeile)
	if err != nil {
		return nil, planner.Household{}, err
	}
	vorlagen, err := p.templates(ctx, zeile.ID)
	if err != nil {
		return nil, planner.Household{}, err
	}
	hist, err := p.history(ctx, zeile.ID)
	if err != nil {
		return nil, planner.Household{}, err
	}

	// Dieselbe Umrechnung wie im Planer. Stünde hier die Zahl aus der
	// Bibliothek, zeigte diese Seite 25 Minuten Staubsaugen und der Plan 42 —
	// und der Nutzer hätte zwei Wahrheiten vor sich.
	vorlagen = planner.ScaleTemplates(vorlagen, haushalt)

	out := make([]planner.TemplateState, 0, len(vorlagen))
	for _, t := range vorlagen {
		grund, faktum, fehlt := planner.Status(t, haushalt, hist)
		out = append(out, planner.TemplateState{
			Template:  t,
			Active:    grund == "",
			Reason:    grund,
			Fact:      faktum,
			Need:      fehlt,
			Agreement: hist.Agreement(t.ID),
		})
	}
	return out, haushalt, nil
}

// SetTemplateActive bestellt eine Vorlage ab oder wieder an.
//
// Abbestellen schreibt einen Lernwert und ein Ereignis: Der Lernwert steuert
// den Planer, das Ereignis erzählt später, wann und von wem. Wieder anbestellen
// löscht den Lernwert — „nie etwas gesagt" und „wieder erlaubt" sollen im
// Planer gleich wirken.
//
// Nur planende Personen: Eine abbestellte Vorlage verschwindet für alle.
func (p *Plans) SetTemplateActive(ctx context.Context, subject, id, vorlageID string, aktiv bool) error {
	zeile, err := p.alsPlanende(ctx, subject, id)
	if err != nil {
		return err
	}

	mitglied, err := p.db.GetMemberInHousehold(ctx, db.GetMemberInHouseholdParams{
		HouseholdID: zeile.ID,
		AuthUserID:  &subject,
	})
	if err != nil {
		return err
	}

	if aktiv {
		err = p.db.DeleteSignal(ctx, db.DeleteSignalParams{
			HouseholdID: zeile.ID,
			TemplateID:  vorlageID,
			Kind:        signalAbgewaehlt,
		})
	} else {
		err = p.db.UpsertSignal(ctx, db.UpsertSignalParams{
			HouseholdID: zeile.ID,
			TemplateID:  vorlageID,
			Kind:        signalAbgewaehlt,
			Value:       1,
		})
	}
	if err != nil {
		return err
	}

	nutzlast, err := json.Marshal(map[string]any{"vorlage": vorlageID, "aktiv": aktiv})
	if err != nil {
		return err
	}
	// „geloescht" ist die Art, die das Protokoll für Abbestellen kennt. Auch
	// das Wiederanbestellen läuft darüber — mit aktiv=true in der Nutzlast,
	// damit die Zeitreihe vollständig bleibt.
	_, err = p.db.AppendEvent(ctx, db.AppendEventParams{
		HouseholdID: zeile.ID,
		MemberID:    mitglied.ID,
		Kind:        "geloescht",
		Payload:     nutzlast,
	})
	return err
}

// AddOwnTemplate legt eine Aufgabe an, die es in der Bibliothek nicht gibt.
//
// Jede Auswahlliste hat ein fehlendes Gericht. Ohne diesen Weg merkt ein
// Haushalt beim ersten Blick, dass seine wichtigste Sache fehlt — die
// Medikamente der Großmutter, der Putzplan der WG, das Vereinstraining.
//
// Nur planende Personen, und die Kennung trägt den Ursprung: Eine eigene
// Aufgabe soll in jeder Liste als solche erkennbar sein. Ihre Zahlen sind
// geschätzt und tragen trotzdem die Fairnessrechnung mit.
func (p *Plans) AddOwnTemplate(ctx context.Context, subject, id string, o planner.OwnTask) (planner.TaskTemplate, error) {
	zeile, err := p.alsPlanende(ctx, subject, id)
	if err != nil {
		return planner.TaskTemplate{}, err
	}
	if err := o.Validate(); err != nil {
		return planner.TaskTemplate{}, err
	}

	kennung, err := eigeneKennung()
	if err != nil {
		return planner.TaskTemplate{}, err
	}
	vorlage := o.Template(kennung)

	roh, err := library.MarshalTemplate(vorlage)
	if err != nil {
		return planner.TaskTemplate{}, err
	}
	if _, err := p.db.CreateOwnTemplate(ctx, db.CreateOwnTemplateParams{
		ID:          kennung,
		HouseholdID: zeile.ID,
		Definition:  roh,
	}); err != nil {
		return planner.TaskTemplate{}, err
	}
	return vorlage, nil
}

// eigeneKennung erzeugt eine Kennung, die nicht mit den kuratierten kollidiert.
//
// Das Präfix ist kein Schmuck: task_template.id ist global eindeutig, und eine
// selbst angelegte Aufgabe darf keine kuratierte überschreiben, wenn beide
// zufällig „t-waesche" heißen.
func eigeneKennung() (string, error) {
	roh := make([]byte, 8)
	if _, err := rand.Read(roh); err != nil {
		return "", err
	}
	return "eigen-" + hex.EncodeToString(roh), nil
}

// SetAgreement schreibt das Wochenraster einer Vorlage — alle sieben Plätze
// auf einmal.
//
// Erst leeren, dann setzen: Sieben Plätze einzeln abzugleichen wäre sieben
// Abfragen und dieselbe Wirkung, und ein halb abgeglichenes Raster wäre ein
// Zustand, den niemand gewollt hätte.
//
// Nur planende Personen. Eine Absprache ist eine Vereinbarung über andere
// Menschen — das ist keine Einstellung, die jeder für sich dreht.
func (p *Plans) SetAgreement(ctx context.Context, subject, id, vorlageID string, raster [7]string) error {
	zeile, err := p.alsPlanende(ctx, subject, id)
	if err != nil {
		return err
	}

	tx, err := p.db.Pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback(ctx) }()
	q := p.db.Queries.WithTx(tx)

	if err := q.ClearAgreementRow(ctx, db.ClearAgreementRowParams{
		HouseholdID: zeile.ID,
		TemplateID:  vorlageID,
	}); err != nil {
		return err
	}

	for tag, wer := range raster {
		if wer == "" {
			continue
		}
		kennung, ok := parseUUID(wer)
		if !ok {
			return fmt.Errorf("%w: %q ist keine person", planner.ErrUnknownMember, wer)
		}
		if err := q.SetAgreement(ctx, db.SetAgreementParams{
			HouseholdID: zeile.ID,
			TemplateID:  vorlageID,
			Weekday:     int32(tag),
			MemberID:    kennung,
		}); err != nil {
			return err
		}
	}
	return tx.Commit(ctx)
}
