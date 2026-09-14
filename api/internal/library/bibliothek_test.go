package library

import (
	"slices"
	"testing"

	"github.com/zakaria/haushalt/api/internal/planner"
)

// Die Bibliothek gegen sich selbst geprüft.
//
// Anlass war eine Frage, die nichts bewirkte: „Wohnung oder Haus?" stand als
// Pflichtschritt im Onboarding, und keine einzige Vorlage hing daran. Die
// Regel dagegen war längst aufgeschrieben — der Fakten-Loader weist eine Frage
// ohne erkennbaren Nutzen zurück. Sie galt nur für die Fakten, nicht für die
// Bibliothek als Ganzes.
//
// Deshalb hier: Jede Frage muss etwas bewirken, und alles, was gebraucht wird,
// muss fragbar sein. Beide Richtungen, sonst fällt jeweils die andere durch.
func bibliothek(t *testing.T) ([]planner.TaskTemplate, *Facts) {
	t.Helper()
	vorlagen, err := LoadTemplates("../../library/vorlagen.json")
	if err != nil {
		t.Fatalf("Bibliothek: %v", err)
	}
	fakten, err := LoadFacts("../../library/fakten.json")
	if err != nil {
		t.Fatalf("Fakten: %v", err)
	}
	return vorlagen, fakten
}

func TestJedeFrageBewirktEtwas(t *testing.T) {
	vorlagen, fakten := bibliothek(t)

	gebraucht := map[string]bool{}
	for _, v := range vorlagen {
		for _, f := range v.AppliesTo.RequiresFacts {
			gebraucht[f] = true
		}
	}

	for id := range fakten.nachID {
		if !gebraucht[id] {
			t.Errorf("das Faktum %q wird erfragt, aber keine Vorlage hängt daran — eine Frage, die nichts bewirkt", id)
		}
	}
}

func TestJedeVoraussetzungIstFragbar(t *testing.T) {
	vorlagen, fakten := bibliothek(t)

	for _, v := range vorlagen {
		for _, f := range v.AppliesTo.RequiresFacts {
			if _, ok := fakten.Get(f); !ok {
				t.Errorf("%q setzt %q voraus, aber es gibt keine Frage dazu — die Vorlage bliebe für immer unsichtbar",
					v.ID, f)
			}
		}
	}
}

func TestAnlassartenSindBekannt(t *testing.T) {
	vorlagen, _ := bibliothek(t)

	for _, v := range vorlagen {
		art := v.AppliesTo.RequiresOccasion
		if art == "" {
			continue
		}
		if !slices.Contains(planner.Kinds, art) {
			t.Errorf("%q hängt an der Anlassart %q, die niemand eintragen kann", v.ID, art)
		}
	}
}

func TestAnlassgebundeneVorlagenHabenEineArt(t *testing.T) {
	vorlagen, _ := bibliothek(t)

	for _, v := range vorlagen {
		if v.AppliesTo.RequiresEvent && v.AppliesTo.RequiresOccasion == "" {
			t.Errorf("%q braucht einen Anlass, sagt aber nicht welchen — sie entstünde nie", v.ID)
		}
	}
}
