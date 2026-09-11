// Kommando import schreibt die Bibliothek und die Beispielhaushalte aus dem
// Repo in die Datenbank.
//
// Die Dateien bleiben die Quelle, die Tabellen sind die Kopie. Ein zweiter Lauf
// ändert nichts, was schon stimmt — mit einer Ausnahme, und die ist kein
// Versehen: Das Ereignisprotokoll wird nur beim ersten Mal geschrieben. Es ist
// reines Anhängen und lässt sich nicht bereinigen, also darf der Import es
// nicht zweimal füllen.
//
//	make import
package main

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"time"

	"github.com/jackc/pgx/v5/pgtype"

	"github.com/zakaria/haushalt/api/internal/library"
	"github.com/zakaria/haushalt/api/internal/planner"
	"github.com/zakaria/haushalt/api/internal/storage"
	"github.com/zakaria/haushalt/api/internal/storage/db"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, "fehler:", err)
		os.Exit(1)
	}
}

func run() error {
	var (
		bibliothek = flag.String("bibliothek", "library/vorlagen.json", "Vorlagen-Bibliothek")
		beispiele  = flag.String("beispiele", "library/beispiele", "Verzeichnis mit Beispielhaushalten")
	)
	flag.Parse()

	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		return errors.New("DATABASE_URL ist leer")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	store, err := storage.Open(ctx, dsn)
	if err != nil {
		return err
	}
	defer store.Close()

	templates, err := importTemplates(ctx, store, *bibliothek)
	if err != nil {
		return err
	}

	pfade, err := filepath.Glob(filepath.Join(*beispiele, "*.json"))
	if err != nil {
		return err
	}
	sort.Strings(pfade)
	for _, p := range pfade {
		if err := importHousehold(ctx, store, p, templates); err != nil {
			return fmt.Errorf("%s: %w", p, err)
		}
	}
	return nil
}

// importTemplates schreibt die Vorlagen und gibt sie geparst zurück — die
// Historie braucht später Dauer, Kopflast und Zeitfenster.
//
// In die Spalte definition geht der Abschnitt der Datei UNVERÄNDERT, nicht
// das, was der Parser daraus gemacht hat. Sonst wäre der Umweg durch Go eine
// stille Übersetzung: Felder, die eine spätere Version kennt, wären beim
// Import verloren.
func importTemplates(ctx context.Context, store *storage.DB, pfad string) (map[string]planner.TaskTemplate, error) {
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
		if err := store.UpsertCuratedTemplate(ctx, db.UpsertCuratedTemplateParams{
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
	fmt.Printf("Bibliothek %s: %d Vorlagen\n", version, len(geparst))
	return nachID, nil
}

func importHousehold(ctx context.Context, store *storage.DB, pfad string, templates map[string]planner.TaskTemplate) error {
	h, hist, err := library.LoadHousehold(pfad)
	if err != nil {
		return err
	}
	slug := trimExt(filepath.Base(pfad))

	haushalt, err := store.UpsertDemoHousehold(ctx, db.UpsertDemoHouseholdParams{
		Slug:     &slug,
		Name:     h.Name,
		Home:     string(h.Context.Home),
		HasCar:   h.Context.HasCar,
		HasYard:  h.Context.HasYard,
		Pets:     h.Context.Pets,
		Timezone: "Europe/Berlin",
	})
	if err != nil {
		return err
	}

	// Kennung aus der Datei ("m-anna") auf die Kennung in der Datenbank.
	mitglieder := map[string]pgtype.UUID{}
	for _, m := range h.Members {
		zeile, err := store.UpsertMember(ctx, db.UpsertMemberParams{
			HouseholdID: haushalt.ID,
			Name:        m.Name,
			Role:        string(m.Role),
			// Das Modell führt Geburtsjahre, die Beispieldatei ein Alter.
			// Für Demodaten ist die Umrechnung genau genug; bei echten
			// Haushalten wird das Jahr erfasst und nie das volle Datum.
			BirthYear:       int32p(time.Now().Year() - m.Age),
			Care:            textp(string(m.Care)),
			CapacityMinutes: int32s(m.CapacityMinutes[:]),
		})
		if err != nil {
			return err
		}
		mitglieder[m.ID] = zeile.ID
	}

	// Abgewählte Vorlagen sind ein Lernwert, keine eigene Tabelle.
	for id := range hist.Muted {
		if err := store.UpsertSignal(ctx, db.UpsertSignalParams{
			HouseholdID: haushalt.ID,
			TemplateID:  id,
			Kind:        "abgewaehlt",
			Value:       1,
		}); err != nil {
			return err
		}
	}

	vorhandene, err := store.CountEvents(ctx, haushalt.ID)
	if err != nil {
		return err
	}
	if vorhandene > 0 {
		fmt.Printf("%-12s %d Mitglieder · Historie bleibt (%d Ereignisse)\n",
			slug, len(h.Members), vorhandene)
		return nil
	}

	n, err := importHistory(ctx, store, haushalt.ID, hist, mitglieder, templates)
	if err != nil {
		return err
	}
	fmt.Printf("%-12s %d Mitglieder · %d Ereignisse angelegt\n", slug, len(h.Members), n)
	return nil
}

// importHistory macht aus den beiden Verdichtungen in der Beispieldatei das,
// woraus sie im Betrieb entstehen: Aufgaben, Zuteilungen und Protokoll.
//
// Der Umweg ist Absicht. Der Planer liest seine Historie über zwei Abfragen
// auf genau diese Tabellen — würde der Import die Verdichtung direkt
// hinschreiben, wäre der Demo-Haushalt der einzige, bei dem die Abfragen nie
// benutzt werden.
func importHistory(
	ctx context.Context,
	store *storage.DB,
	haushalt pgtype.UUID,
	hist planner.History,
	mitglieder map[string]pgtype.UUID,
	templates map[string]planner.TaskTemplate,
) (int, error) {
	// Alle Vorlagen, über die die Datei etwas sagt — in stabiler Reihenfolge,
	// damit zwei Importe dieselbe Datenbank ergeben.
	var ids []string
	gesehen := map[string]bool{}
	for id := range hist.LastDone {
		if !gesehen[id] {
			ids, gesehen[id] = append(ids, id), true
		}
	}
	for id := range hist.LastAssignee {
		if !gesehen[id] {
			ids, gesehen[id] = append(ids, id), true
		}
	}
	sort.Strings(ids)

	ereignisse := 0
	for _, id := range ids {
		tmpl, ok := templates[id]
		if !ok {
			continue // Vorlage aus der Bibliothek verschwunden
		}

		tag, erledigt := hist.LastDone[id]
		if !erledigt {
			// Nur bekannt, wer sie hatte: eine Woche zurückdatieren, damit
			// die Rotation greift, ohne Erledigung zu behaupten.
			tag = planner.DateOf(time.Now().AddDate(0, 0, -7))
		}

		aufgabe, err := store.InsertTaskInstance(ctx, db.InsertTaskInstanceParams{
			HouseholdID: haushalt,
			TemplateID:  id,
			ISOWeek:     planner.WeekOf(tag).String(),
			Day:         date(tag),
			Slot:        string(slotOrDefault(tmpl.Slot)),
			DurationMin: int32(tmpl.DurationMin),
			HeadLoad:    int32(tmpl.HeadLoad),
		})
		if err != nil {
			return ereignisse, fmt.Errorf("aufgabe %s: %w", id, err)
		}

		var wer pgtype.UUID
		if name, ok := hist.LastAssignee[id]; ok {
			if mid, ok := mitglieder[name]; ok {
				wer = mid
				if err := store.InsertAssignment(ctx, db.InsertAssignmentParams{
					TaskInstanceID: aufgabe,
					MemberID:       mid,
					ReasonCode:     "importiert",
					Manual:         false,
				}); err != nil {
					return ereignisse, fmt.Errorf("zuteilung %s: %w", id, err)
				}
			}
		}

		if !erledigt {
			continue
		}
		if _, err := store.AppendHistoricEvent(ctx, db.AppendHistoricEventParams{
			HouseholdID:    haushalt,
			MemberID:       wer,
			TaskInstanceID: aufgabe,
			Kind:           "erledigt",
			Payload:        []byte(`{"quelle":"import"}`),
			OccurredAt:     mittags(tag),
		}); err != nil {
			return ereignisse, fmt.Errorf("ereignis %s: %w", id, err)
		}
		ereignisse++
	}
	return ereignisse, nil
}

// ------------------------------------------------------------- Kleinkram

func trimExt(name string) string { return name[:len(name)-len(filepath.Ext(name))] }

// Nullbare Text- und Zahlenspalten kommen bei sqlc als Zeiger heraus
// (emit_pointers_for_null_types): nil ist NULL. UUID und Datum haben dagegen
// ihre eigene Valid-Markierung und bleiben Werte.

// textp gibt nil für die leere Zeichenkette — "kein Wert" und "leerer Wert"
// sind in der Datenbank verschiedene Dinge, und care ohne Angabe ist NULL.
func textp(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

func int32p(v int) *int32 {
	n := int32(v)
	return &n
}

// mittags macht aus einem Kalendertag einen Zeitpunkt: 12 Uhr UTC.
//
// Nicht Mitternacht. Der Planer liest das Datum später über
// `occurred_at AT TIME ZONE <zeitzone>`, und Mitternacht UTC rutscht in jeder
// Zone westlich von Greenwich auf den Vortag. Die Mittagsstunde überlebt jede
// Umrechnung.
func mittags(d planner.Date) pgtype.Timestamptz {
	return pgtype.Timestamptz{
		Time:  time.Date(d.Year, d.Month, d.Day, 12, 0, 0, 0, time.UTC),
		Valid: true,
	}
}

func date(d planner.Date) pgtype.Date {
	return pgtype.Date{
		Time:  time.Date(d.Year, d.Month, d.Day, 0, 0, 0, 0, time.UTC),
		Valid: true,
	}
}

func int32s(in []int) []int32 {
	out := make([]int32, len(in))
	for i, v := range in {
		out[i] = int32(v)
	}
	return out
}

func slotOrDefault(s planner.Slot) planner.Slot {
	if s == "" {
		return planner.SlotAny
	}
	return s
}
