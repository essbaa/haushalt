package library

import (
	"context"
	"fmt"
	"path/filepath"
	"sort"
	"strings"

	"github.com/zakaria/haushalt/api/internal/planner"
)

// Catalog ist die eingelesene Bibliothek samt Beispielhaushalten.
//
// Bis zum Datenmodell (T5) ist das die Datenquelle des Dienstes. Danach
// kommen die Haushalte aus der Datenbank und die Vorlagen bleiben hier — die
// Bibliothek ist versionierte Repo-Datei, kein Nutzerinhalt.
//
// Einmal beim Start gelesen und danach nur noch gelesen: keine Sperren nötig.
type Catalog struct {
	templates  []planner.TaskTemplate
	households map[string]entry
	order      []string
}

type entry struct {
	household planner.Household
	history   planner.History
}

// LoadCatalog liest dir/vorlagen.json und alle dir/beispiele/*.json.
//
// Der Dateiname ohne Endung ist die Kennung des Haushalts: aus
// beispiele/familie-a.json wird "familie-a". Damit steht die URL in der Datei
// und nicht in einer zweiten Liste, die man vergisst mitzupflegen.
func LoadCatalog(dir string) (*Catalog, error) {
	templates, err := LoadTemplates(filepath.Join(dir, "vorlagen.json"))
	if err != nil {
		return nil, err
	}

	pattern := filepath.Join(dir, "beispiele", "*.json")
	paths, err := filepath.Glob(pattern)
	if err != nil {
		return nil, err
	}
	if len(paths) == 0 {
		return nil, fmt.Errorf("library: keine Beispielhaushalte unter %s", pattern)
	}
	sort.Strings(paths)

	c := &Catalog{templates: templates, households: map[string]entry{}}
	for _, p := range paths {
		h, hist, err := LoadHousehold(p)
		if err != nil {
			return nil, err
		}
		id := strings.TrimSuffix(filepath.Base(p), filepath.Ext(p))
		if _, doppelt := c.households[id]; doppelt {
			return nil, fmt.Errorf("library: die Haushalts-Kennung %q kommt doppelt vor", id)
		}
		// Die Kennung aus dem Dateinamen wird zur Kennung des Haushalts.
		// In der Datenbank ist es später der Slug — dieselbe Rolle, andere
		// Quelle.
		h.ID = id
		c.households[id] = entry{household: h, history: hist}
		c.order = append(c.order, id)
	}
	return c, nil
}

// Templates ist die Vorlagen-Bibliothek.
func (c *Catalog) Templates() []planner.TaskTemplate { return c.templates }

// Households liefert die Haushalte in stabiler Reihenfolge. Household.ID
// trägt dabei die Kennung aus dem Dateinamen — dieselbe, die in der URL steht.
//
// Der Kontext wird nicht gebraucht und steht trotzdem in der Signatur: Die
// zweite Quelle ist eine Datenbank, und die HTTP-Schicht soll die beiden nicht
// unterscheiden können.
func (c *Catalog) Households(context.Context) ([]planner.Household, error) {
	out := make([]planner.Household, 0, len(c.order))
	for _, id := range c.order {
		out = append(out, c.households[id].household)
	}
	return out, nil
}

// Plan berechnet den Wochenplan eines Haushalts.
//
// Der Catalog hält die Daten, die Rechnung macht der Planer — dieses Paket
// entscheidet nichts, es reicht durch.
func (c *Catalog) Plan(_ context.Context, id string, week planner.Week) (planner.Result, planner.Household, error) {
	e, ok := c.households[id]
	if !ok {
		return planner.Result{}, planner.Household{}, fmt.Errorf("%w: %q", planner.ErrUnknownHousehold, id)
	}
	res, err := planner.Plan(planner.Input{
		Household: e.household,
		Templates: c.templates,
		Week:      week,
		History:   e.history,
		Limits:    planner.DefaultLimits(),
	})
	return res, e.household, err
}
