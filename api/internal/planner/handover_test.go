package planner

import "testing"

func haushaltZuDritt() Household {
	return Household{
		ID:   "h",
		Name: "Test",
		Members: []Member{
			{ID: "a", Name: "Anna", Role: RolePlanner, Age: 38, CapacityMinutes: [7]int{60, 60, 60, 60, 60, 120, 120}},
			{ID: "b", Name: "Ben", Role: RolePlanner, Age: 40, CapacityMinutes: [7]int{60, 60, 60, 60, 60, 120, 120}},
			{ID: "c", Name: "Mia", Role: RoleDoer, Age: 13, CapacityMinutes: [7]int{30, 30, 30, 30, 30, 90, 60}},
		},
	}
}

func eingabe(h Household, ts ...TaskTemplate) Input {
	return Input{Household: h, Templates: ts, Limits: DefaultLimits()}
}

func TestHandover(t *testing.T) {
	kochen := TaskTemplate{ID: "kochen", Title: "Kochen", MinAge: 14, DurationMin: 40}
	saugen := TaskTemplate{ID: "saugen", Title: "Saugen", MinAge: 10, DurationMin: 25}

	t.Run("nimmt die Person mit der geringsten Last", func(t *testing.T) {
		tasks := []PlannedTask{
			{ID: "1", TemplateID: "kochen", AssigneeID: "a", DurationMin: 40},
			{ID: "2", TemplateID: "saugen", AssigneeID: "b", DurationMin: 25},
			{ID: "3", TemplateID: "saugen", AssigneeID: "b", DurationMin: 25},
		}
		// Anna gibt ab. Ben trägt schon 50 Minuten, Mia nichts — aber Mia ist
		// 13 und die Vorlage verlangt 14.
		wer, grund, ok := Handover(eingabe(haushaltZuDritt(), kochen, saugen), tasks, "1")
		if !ok {
			t.Fatal("niemand gefunden")
		}
		if wer != "b" {
			t.Errorf("übernimmt %q statt Ben", wer)
		}
		if grund.Code != ReasonBalance || grund.Previous != "a" {
			t.Errorf("Begründung %+v nennt nicht den Ausgleich und die Vorgängerin", grund)
		}
	})

	t.Run("meidet, wer dieselbe Vorlage diese Woche schon hat", func(t *testing.T) {
		tasks := []PlannedTask{
			{ID: "1", TemplateID: "saugen", AssigneeID: "a", DurationMin: 25},
			{ID: "2", TemplateID: "saugen", AssigneeID: "b", DurationMin: 25},
		}
		// Mia hat mehr Last im Verhältnis zu ihrer Kapazität als Ben? Nein —
		// sie hat gar keine. Ben hält aber schon dasselbe, also geht es an Mia.
		wer, _, ok := Handover(eingabe(haushaltZuDritt(), saugen), tasks, "1")
		if !ok || wer != "c" {
			t.Errorf("übernimmt %q statt Mia (Ben saugt diese Woche schon)", wer)
		}
	})

	t.Run("eigene Aufgaben gibt man nicht ab", func(t *testing.T) {
		bett := TaskTemplate{ID: "bett", Title: "Bett beziehen", PerPerson: true, DurationMin: 20}
		tasks := []PlannedTask{{ID: "1", TemplateID: "bett", AssigneeID: "a", DurationMin: 20}}
		if _, _, ok := Handover(eingabe(haushaltZuDritt(), bett), tasks, "1"); ok {
			t.Error("eine personengebundene Aufgabe wurde weitergereicht")
		}
	})

	t.Run("niemand geeignet ist kein Fehler", func(t *testing.T) {
		nurErwachsen := TaskTemplate{ID: "steuer", Title: "Steuer", Distribution: DistAdultsOnly, DurationMin: 60}
		h := haushaltZuDritt()
		h.Members = h.Members[:1] // nur Anna
		tasks := []PlannedTask{{ID: "1", TemplateID: "steuer", AssigneeID: "a", DurationMin: 60}}
		if _, _, ok := Handover(eingabe(h, nurErwachsen), tasks, "1"); ok {
			t.Error("es wurde jemand gefunden, den es nicht gibt")
		}
	})

	t.Run("zwei Aufrufe ergeben dasselbe", func(t *testing.T) {
		tasks := []PlannedTask{{ID: "1", TemplateID: "saugen", AssigneeID: "a", DurationMin: 25}}
		in := eingabe(haushaltZuDritt(), saugen)
		erst, _, _ := Handover(in, tasks, "1")
		zweit, _, _ := Handover(in, tasks, "1")
		if erst != zweit {
			t.Errorf("%q beim ersten, %q beim zweiten Aufruf", erst, zweit)
		}
	})
}
