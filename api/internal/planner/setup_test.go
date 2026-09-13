package planner

import (
	"errors"
	"strings"
	"testing"
)

func gueltig() Setup {
	return Setup{
		Name:    "Familie Bauer",
		Context: Context{Home: HomeFlat},
		Members: []SetupMember{
			{Name: "Zakaria", Role: RolePlanner, Budget: BudgetMedium},
			{Name: "Lina", Role: RoleDependent, BirthYear: 2022, Budget: BudgetNone},
		},
	}
}

func TestSetupValidate(t *testing.T) {
	faelle := []struct {
		name    string
		aendern func(*Setup)
		fehlt   string // Teil der erwarteten Meldung; leer = muss durchgehen
	}{
		{name: "der Normalfall"},
		{name: "ohne Namen", aendern: func(s *Setup) { s.Name = "" }, fehlt: "Namen"},
		{name: "ohne Personen", aendern: func(s *Setup) { s.Members = nil }, fehlt: "Personen"},
		{
			name:    "ohne planende Person",
			aendern: func(s *Setup) { s.Members[0].Role = RoleDoer },
			fehlt:   "planen",
		},
		{
			name:    "zwei gleiche Namen",
			aendern: func(s *Setup) { s.Members[1].Name = "zakaria" },
			fehlt:   "zweimal",
		},
		{
			name:    "erfundene Wohnform",
			aendern: func(s *Setup) { s.Context.Home = "hausboot" },
			fehlt:   "Wohnform",
		},
		{
			name:    "betreut ohne Geburtsjahr",
			aendern: func(s *Setup) { s.Members[1].BirthYear = 0 },
			fehlt:   "Geburtsjahr",
		},
		{
			name:    "Geburtsjahr in der Zukunft",
			aendern: func(s *Setup) { s.Members[1].BirthYear = 3000 },
			fehlt:   "Geburtsjahr",
		},
		{
			name:    "erfundene Zeitzone",
			aendern: func(s *Setup) { s.Timezone = "Europe/Bockenheim" },
			fehlt:   "Zeitzone",
		},
		{
			name: "dreizehn Personen",
			aendern: func(s *Setup) {
				for range 12 {
					s.Members = append(s.Members, SetupMember{Name: strings.Repeat("x", len(s.Members)+1), Role: RoleDoer})
				}
			},
			fehlt: "Personen",
		},
		{
			// Ein Erwachsener, der nur ausführt: der planungsunwillige Partner
			// aus dem Produktkonzept. Er hat kein Geburtsjahr und ist trotzdem
			// erwachsen.
			name: "erwachsen und nur ausführend",
			aendern: func(s *Setup) {
				s.Members = append(s.Members, SetupMember{Name: "Sam", Role: RoleDoer, Budget: BudgetLow})
			},
		},
	}

	for _, f := range faelle {
		t.Run(f.name, func(t *testing.T) {
			s := gueltig()
			if f.aendern != nil {
				f.aendern(&s)
			}
			err := s.Normalized().Validate()

			if f.fehlt == "" {
				if err != nil {
					t.Fatalf("sollte durchgehen, kam: %v", err)
				}
				return
			}
			if err == nil {
				t.Fatal("sollte scheitern, ging durch")
			}
			if !errors.Is(err, ErrInvalidSetup) {
				t.Errorf("kein ErrInvalidSetup: %v", err)
			}
			if !strings.Contains(err.Error(), f.fehlt) {
				t.Errorf("Meldung nennt %q nicht: %v", f.fehlt, err)
			}
		})
	}
}

func TestSetupNormalized(t *testing.T) {
	s := Setup{
		Name:    "  Familie Bauer  ",
		Members: []SetupMember{{Name: " Zakaria ", Role: RolePlanner}, {Name: "Lina", Role: RoleDependent, Budget: BudgetHigh}},
	}.Normalized()

	if s.Name != "Familie Bauer" {
		t.Errorf("Name nicht getrimmt: %q", s.Name)
	}
	if s.Timezone != "Europe/Berlin" {
		t.Errorf("Zeitzone %q statt Europe/Berlin", s.Timezone)
	}
	if s.Context.Home != HomeFlat {
		t.Errorf("Wohnform %q statt der Vorgabe", s.Context.Home)
	}
	if s.Members[0].Budget != BudgetMedium {
		t.Errorf("fehlendes Budget wurde %q statt mittel", s.Members[0].Budget)
	}
	// Betreute übernehmen nichts — egal, was der Client behauptet.
	if s.Members[1].Budget != BudgetNone {
		t.Errorf("betreute Person behielt Budget %q", s.Members[1].Budget)
	}
}

func TestIsAdultOhneAlter(t *testing.T) {
	// Der Beitritt über eine Einladung setzt kein Geburtsjahr. Diese Person
	// darf nicht als Kind gelten — sonst bekommt sie keine Erwachsenenaufgabe
	// und zählt obendrein als Kind des Haushalts.
	if !(Member{Name: "Sam", Role: RoleDoer}).IsAdult() {
		t.Error("ausführende Person ohne Alter gilt als Kind")
	}
	if (Member{Name: "Mia", Role: RoleDoer, Age: 13}).IsAdult() {
		t.Error("13-Jährige gilt als erwachsen")
	}
}
