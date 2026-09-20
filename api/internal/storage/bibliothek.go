package storage

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"os"

	"github.com/zakaria/haushalt/api/internal/library"
	"github.com/zakaria/haushalt/api/internal/planner"
	"github.com/zakaria/haushalt/api/internal/storage/db"
)

// ImportLibrary schreibt die Vorlagen aus der Datei in die Datenbank und gibt
// sie geparst zurück — die Historie braucht später Dauer, Kopflast und
// Zeitfenster.
//
// Stand hier nicht, sondern in cmd/import — und damit an einem Ort, den das
// Image nicht enthält. Die Folge war eine Regel im Kopf: „nach dem Deploy noch
// make import laufen lassen, mit der Produktions-URL". Einmal vergessen, und
// in der Produktion standen die Migrationen, aber die Kita-Vorlagen fehlten.
// Eine Regel, an die man sich erinnern muss, ist keine Regel (dasselbe
// Argument wie bei den Migrationen selbst, cmd/migrate).
//
// Jetzt ruft cmd/migrate das hier auf, und das ist der release_command. Die
// Bibliothek reist damit im selben Image wie der Dienst, der sie braucht.
//
// In die Spalte definition geht der Abschnitt der Datei UNVERÄNDERT, nicht
// das, was der Parser daraus gemacht hat. Sonst wäre der Umweg durch Go eine
// stille Übersetzung: Felder, die eine spätere Version kennt, wären beim
// Import verloren.
//
// Ein zweiter Lauf ändert nichts, was schon stimmt: UpsertCuratedTemplate
// schreibt dieselbe Zeile noch einmal. Deshalb darf der Import bei jedem
// Deploy laufen.
func (d *DB) ImportLibrary(ctx context.Context, pfad string) (map[string]planner.TaskTemplate, error) {
	roh, err := os.ReadFile(pfad)
	if err != nil {
		return nil, err
	}
	var doc struct {
		Vorlagen []json.RawMessage `json:"vorlagen"`
	}
	if err := json.Unmarshal(roh, &doc); err != nil {
		return nil, fmt.Errorf("%s: %w", pfad, err)
	}

	// Der Stand der Bibliothek als Prüfsumme der Datei. Eine Versionsnummer
	// von Hand zu pflegen hieße, sie irgendwann zu vergessen.
	summe := sha256.Sum256(roh)
	version := hex.EncodeToString(summe[:])[:12]

	for _, eintrag := range doc.Vorlagen {
		var kopf struct {
			ID string `json:"id"`
		}
		if err := json.Unmarshal(eintrag, &kopf); err != nil {
			return nil, err
		}
		if kopf.ID == "" {
			return nil, errors.New("eine Vorlage ohne id")
		}
		if err := d.UpsertCuratedTemplate(ctx, db.UpsertCuratedTemplateParams{
			ID:         kopf.ID,
			Version:    version,
			Definition: eintrag,
		}); err != nil {
			return nil, err
		}
	}

	geparst, err := library.LoadTemplates(pfad)
	if err != nil {
		return nil, err
	}
	nachID := make(map[string]planner.TaskTemplate, len(geparst))
	for _, t := range geparst {
		nachID[t.ID] = t
	}
	return nachID, nil
}
