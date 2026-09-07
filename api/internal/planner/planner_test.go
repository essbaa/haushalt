package planner

import (
	"testing"
	"time"
)

// Ein Haushalt, wie er im Konzept steht: zwei berufstätige Erwachsene, ein
// Kind in der Kita. Kapazität grob: werktags eine knappe Stunde, am Wochenende
// mehr.
func familieMitKleinkind() Household {
	werktags := [7]int{60, 60, 60, 60, 60, 180, 180}
	return Household{
		ID:   "hh-1",
		Name: "Familie A",
		Members: []Member{
			{ID: "m-anna", Name: "Anna", Role: RolePlanner, Age: 36, CapacityMinutes: werktags},
			{ID: "m-ben", Name: "Ben", Role: RolePlanner, Age: 38, CapacityMinutes: werktags},
			{ID: "m-lina", Name: "Lina", Role: RoleDependent, Age: 3, Care: CareKita},
		},
		Context: Context{Home: HomeFlat},
	}
}

func haushaltMitTeenager() Household {
	erw := [7]int{60, 60, 60, 60, 60, 180, 180}
	teen := [7]int{30, 30, 30, 30, 30, 90, 90}
	return Household{
		ID: "hh-2",
		Members: []Member{
			{ID: "m-eltern-1", Role: RolePlanner, Age: 44, CapacityMinutes: erw},
			{ID: "m-eltern-2", Role: RolePlanner, Age: 45, CapacityMinutes: erw},
			{ID: "m-teen", Role: RoleDoer, Age: 15, Care: CareSchool, CapacityMinutes: teen},
		},
		Context: Context{Home: HomeHouse, HasCar: true, HasYard: true},
	}
}

func spuelmaschine() TaskTemplate {
	return TaskTemplate{
		ID: "t-spuelmaschine", Title: "Spülmaschine ausräumen",
		Category: CatKitchen, Kind: KindDo,
		DurationMin: 8, HeadLoad: HeadLoadNone,
		Rhythm:       Rhythm{Type: RhythmTrigger, EveryDays: 2},
		MinAge:       8,
		Distribution: DistRotate,
		Failure:      FailureSoft,
		Source:       SourceCurated,
	}
}

func bad() TaskTemplate {
	return TaskTemplate{
		ID: "t-bad", Title: "Bad putzen",
		Category: CatCleaning, Kind: KindDo,
		DurationMin: 35, HeadLoad: HeadLoadNone,
		Rhythm:       Rhythm{Type: RhythmWindow, EveryDays: 7},
		MinAge:       12,
		Distribution: DistRotate,
		Failure:      FailureSoft,
		Source:       SourceCurated,
	}
}

func zahnarzt() TaskTemplate {
	return TaskTemplate{
		ID: "t-zahnarzt", Title: "Zahnkontrolle für alle buchen",
		Category: CatAppointment, Kind: KindOrg,
		DurationMin: 8, HeadLoad: HeadLoadHigh,
		Rhythm:       Rhythm{Type: RhythmWindow, EveryDays: 180},
		Distribution: DistAdultsOnly,
		Failure:      FailureSoft,
		Source:       SourceCurated,
	}
}

func reifenwechsel() TaskTemplate {
	return TaskTemplate{
		ID: "t-reifen", Title: "Reifenwechsel-Termin machen",
		Category: CatMaintain, Kind: KindOrg,
		DurationMin: 10, HeadLoad: HeadLoadMid,
		Rhythm:       Rhythm{Type: RhythmSeason, Months: []time.Month{time.October}},
		LeadDays:     21,
		Distribution: DistAdultsOnly,
		AppliesTo:    Conditions{RequiresCar: true},
		Failure:      FailureSoft,
		Source:       SourceCurated,
	}
}

func wechselkleidung() TaskTemplate {
	drei := AgeRange{Min: 1, Max: 6}
	return TaskTemplate{
		ID: "t-wechselkleidung", Title: "Wechselkleidung in der Kita prüfen",
		Category: CatChild, Kind: KindOrg,
		DurationMin: 10, HeadLoad: HeadLoadMid,
		Rhythm:       Rhythm{Type: RhythmWindow, EveryDays: 30},
		Distribution: DistAdultsOnly,
		AppliesTo:    Conditions{RequiresChildAged: &drei, RequiresCare: CareKita},
		Failure:      FailureSoft,
		Source:       SourceCurated,
	}
}

func basisEingabe(h Household, ts ...TaskTemplate) Input {
	return Input{
		Household: h,
		Templates: ts,
		Week:      Week{Year: 2026, Week: 38},
		History:   History{},
		Limits:    DefaultLimits(),
	}
}

// ------------------------------------------------------------------ Auswahl

func TestSelectApplicable(t *testing.T) {
	tests := []struct {
		name      string
		household Household
		template  TaskTemplate
		wantKept  bool
		wantCode  SkipCode
	}{
		{
			name:      "Autoaufgabe ohne Auto faellt heraus",
			household: familieMitKleinkind(),
			template:  reifenwechsel(),
			wantKept:  false,
			wantCode:  SkipNotApplicable,
		},
		{
			name:      "Autoaufgabe mit Auto bleibt",
			household: haushaltMitTeenager(),
			template:  reifenwechsel(),
			wantKept:  true,
		},
		{
			name:      "Kita-Aufgabe gilt nur mit Kind in der Kita",
			household: haushaltMitTeenager(),
			template:  wechselkleidung(),
			wantKept:  false,
			wantCode:  SkipNotApplicable,
		},
		{
			name:      "Kita-Aufgabe gilt mit Kleinkind",
			household: familieMitKleinkind(),
			template:  wechselkleidung(),
			wantKept:  true,
		},
		{
			name:      "Aufgabe ab 12 findet im Kleinkind-Haushalt trotzdem Erwachsene",
			household: familieMitKleinkind(),
			template:  bad(),
			wantKept:  true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			in := basisEingabe(tc.household, tc.template)
			kept, skipped := selectApplicable(in)

			if tc.wantKept {
				if len(kept) != 1 {
					t.Fatalf("Vorlage sollte gelten, wurde aber übersprungen: %+v", skipped)
				}
				return
			}
			if len(kept) != 0 {
				t.Fatalf("Vorlage sollte nicht gelten, blieb aber im Plan")
			}
			if len(skipped) != 1 || skipped[0].Code != tc.wantCode {
				t.Fatalf("Grund = %+v, erwartet %q", skipped, tc.wantCode)
			}
		})
	}
}

// ------------------------------------------------------------ Determinismus

func TestPlanIstDeterministisch(t *testing.T) {
	in := basisEingabe(familieMitKleinkind(),
		spuelmaschine(), bad(), zahnarzt(), wechselkleidung())

	first, err := Plan(in)
	if err != nil {
		t.Fatal(err)
	}
	for i := 0; i < 20; i++ {
		again, err := Plan(in)
		if err != nil {
			t.Fatal(err)
		}
		if len(again.Tasks) != len(first.Tasks) {
			t.Fatalf("Lauf %d: %d Aufgaben statt %d", i, len(again.Tasks), len(first.Tasks))
		}
		for j := range first.Tasks {
			if again.Tasks[j] != first.Tasks[j] {
				t.Fatalf("Lauf %d, Aufgabe %d weicht ab:\n  %+v\n  %+v",
					i, j, first.Tasks[j], again.Tasks[j])
			}
		}
	}
}

// ---------------------------------------------------------------- Rotation

func TestRotationGehtVorAusgleich(t *testing.T) {
	in := basisEingabe(familieMitKleinkind(), bad())
	in.History = History{LastAssignee: map[string]string{"t-bad": "m-anna"}}

	got, err := Plan(in)
	if err != nil {
		t.Fatal(err)
	}
	if len(got.Tasks) != 1 {
		t.Fatalf("erwartet wurde eine Aufgabe, bekommen %d", len(got.Tasks))
	}
	if got.Tasks[0].AssigneeID == "m-anna" {
		t.Fatalf("Anna hatte das Bad zuletzt und bekommt es wieder")
	}
	if got.Tasks[0].Reason.Code != ReasonRotation {
		t.Fatalf("Grund = %q, erwartet %q", got.Tasks[0].Reason.Code, ReasonRotation)
	}
	if got.Tasks[0].Reason.Previous != "m-anna" {
		t.Fatalf("Vorgänger = %q, erwartet m-anna", got.Tasks[0].Reason.Previous)
	}
}

func TestRotationInnerhalbDerWoche(t *testing.T) {
	kochen := TaskTemplate{
		ID: "t-kochen", Title: "Abendessen kochen",
		Category: CatKitchen, Kind: KindDo,
		DurationMin: 40, HeadLoad: HeadLoadMid,
		Rhythm:       Rhythm{Type: RhythmFixed, Weekdays: []time.Weekday{time.Monday, time.Wednesday, time.Friday}},
		Distribution: DistRotate,
		Failure:      FailureHard,
		Source:       SourceCurated,
	}

	got, err := Plan(basisEingabe(familieMitKleinkind(), kochen))
	if err != nil {
		t.Fatal(err)
	}
	if len(got.Tasks) != 3 {
		t.Fatalf("%d Aufgaben, erwartet 3 — feste Wochentage ergeben je einen Termin", len(got.Tasks))
	}

	counts := map[string]int{}
	for _, task := range got.Tasks {
		counts[task.AssigneeID]++
	}
	if len(counts) < 2 {
		t.Fatalf("dreimal kochen landet komplett bei einer Person: %v", counts)
	}
	for id, n := range counts {
		if n > 2 {
			t.Fatalf("%s kocht %d von 3 Abenden", id, n)
		}
	}
}

// --------------------------------------------------------------- Startdichte

func TestStartdichteBegrenztKopfarbeitNichtHandarbeit(t *testing.T) {
	orgVorlagen := []TaskTemplate{}
	for i, titel := range []string{"A", "B", "C", "D", "E", "F"} {
		v := zahnarzt()
		v.ID = "t-org-" + titel
		v.Title = "Organisation " + titel
		v.Rhythm = Rhythm{Type: RhythmWindow, EveryDays: 7 + i}
		orgVorlagen = append(orgVorlagen, v)
	}
	templates := append(orgVorlagen, spuelmaschine(), bad())

	in := basisEingabe(familieMitKleinkind(), templates...)
	in.Limits.MaxOrgTasks = 3

	got, err := Plan(in)
	if err != nil {
		t.Fatal(err)
	}

	org, do := 0, 0
	for _, task := range got.Tasks {
		switch task.Kind {
		case KindOrg:
			org++
		case KindDo:
			do++
		}
	}
	if org > 3 {
		t.Fatalf("%d Organisationsaufgaben im Plan, erlaubt sind 3", org)
	}
	// Fünf: Bad einmal, Spülmaschine alle zwei Tage — also viermal in der
	// Woche. Vor der Fenster-Reparatur waren es zwei, weil jede Vorlage
	// höchstens einen Termin ergab.
	if do != 5 {
		t.Fatalf("%d Ausführungsaufgaben im Plan, erwartet 5 — die Dichtegrenze darf sie nicht treffen", do)
	}
}

// -------------------------------------------------------- Mehrfach fällig

// waesche ist die Vorlage, an der sich zeigt, ob der Planer Abstände wirklich
// versteht: alle drei Tage, also mehr als einmal pro Woche.
func waesche() TaskTemplate {
	return TaskTemplate{
		ID: "t-waesche", Title: "Wäsche waschen",
		Category: CatLaundry, Kind: KindDo,
		DurationMin: 15, HeadLoad: HeadLoadLow,
		Rhythm:       Rhythm{Type: RhythmWindow, EveryDays: 3},
		MinAge:       12,
		Distribution: DistRotate,
		Failure:      FailureSoft,
		Source:       SourceCurated,
	}
}

// TestFensterVorlageWirdMehrfachFaellig prüft die Regel selbst: Eine Vorlage
// mit einem Abstand kürzer als eine Woche ist mehrfach fällig, und jeder
// Termin bekommt ein eigenes Fenster — sonst landen zwei Ladungen Wäsche am
// selben Tag.
//
// Die Fenster dürfen sich nicht überlappen; sie reichen jeweils bis zum Tag
// vor dem nächsten Termin, das letzte bis Sonntag.
func TestFensterVorlageWirdMehrfachFaellig(t *testing.T) {
	monday := MustDate("2026-09-14")
	sunday := MustDate("2026-09-20")

	tests := []struct {
		name  string
		every int
		last  string
		want  [][2]string // je Termin: erster und letzter erlaubter Tag
	}{
		{
			name:  "alle drei Tage, noch nie erledigt",
			every: 3,
			want:  [][2]string{{"2026-09-14", "2026-09-16"}, {"2026-09-17", "2026-09-19"}, {"2026-09-20", "2026-09-20"}},
		},
		{
			name:  "alle zwei Tage",
			every: 2,
			want:  [][2]string{{"2026-09-14", "2026-09-15"}, {"2026-09-16", "2026-09-17"}, {"2026-09-18", "2026-09-19"}, {"2026-09-20", "2026-09-20"}},
		},
		{
			name:  "woechentlich bleibt ein einziger Termin",
			every: 7,
			want:  [][2]string{{"2026-09-14", "2026-09-20"}},
		},
		{
			name:  "am Freitag zuletzt erledigt",
			every: 3,
			last:  "2026-09-11",
			want:  [][2]string{{"2026-09-14", "2026-09-16"}, {"2026-09-17", "2026-09-19"}, {"2026-09-20", "2026-09-20"}},
		},
		{
			// Sechs Wochen nicht gewaschen: Der Planer holt nicht nach, was
			// liegen geblieben ist. Er beginnt am Montag neu. Ein Plan, der
			// Versäumtes aufstapelt, wird gelöscht statt abgearbeitet.
			name:  "laengst ueberfaellig haeuft sich nicht an",
			every: 3,
			last:  "2026-08-01",
			want:  [][2]string{{"2026-09-14", "2026-09-16"}, {"2026-09-17", "2026-09-19"}, {"2026-09-20", "2026-09-20"}},
		},
		{
			name:  "erst naechste Woche wieder dran",
			every: 30,
			last:  "2026-09-13",
			want:  nil,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			v := waesche()
			v.Rhythm.EveryDays = tc.every
			var last Date
			if tc.last != "" {
				last = MustDate(tc.last)
			}

			got := dueEvery(v, last, monday, sunday)

			if len(got) != len(tc.want) {
				t.Fatalf("%d Termine, erwartet %d", len(got), len(tc.want))
			}
			for i, c := range got {
				if len(c.days) == 0 {
					t.Fatalf("Termin %d hat kein einziges erlaubtes Datum", i)
				}
				erster, letzter := c.days[0].String(), c.days[len(c.days)-1].String()
				if erster != tc.want[i][0] || letzter != tc.want[i][1] {
					t.Errorf("Termin %d: Fenster %s..%s, erwartet %s..%s",
						i, erster, letzter, tc.want[i][0], tc.want[i][1])
				}
			}
		})
	}
}

// TestWaescheLandetAnDreiVerschiedenenTagen prüft dasselbe eine Ebene höher:
// Was im Plan ankommt, muss auch verteilt sein — drei Termine an drei Tagen,
// nicht dreimal am Montag, und nicht alles bei derselben Person.
func TestWaescheLandetAnDreiVerschiedenenTagen(t *testing.T) {
	got, err := Plan(basisEingabe(familieMitKleinkind(), waesche()))
	if err != nil {
		t.Fatal(err)
	}
	if len(got.Tasks) != 3 {
		t.Fatalf("%d Aufgaben, erwartet 3 — alle drei Tage heißt dreimal in dieser Woche (übersprungen: %+v)",
			len(got.Tasks), got.Skipped)
	}

	tage := map[Date]int{}
	personen := map[string]int{}
	for _, task := range got.Tasks {
		tage[task.Day]++
		personen[task.AssigneeID]++
	}
	if len(tage) != 3 {
		t.Errorf("die drei Ladungen liegen auf %d Tagen: %v", len(tage), tage)
	}
	if len(personen) < 2 {
		t.Errorf("dreimal Wäsche landet komplett bei einer Person: %v", personen)
	}
}

// ------------------------------------------------------------------ Kopflast

func TestKopflastZaehltInDerBilanz(t *testing.T) {
	l := DefaultLimits()
	kurzUndSchwer := PlannedTask{DurationMin: 8, HeadLoad: HeadLoadHigh}
	langUndLeicht := PlannedTask{DurationMin: 35, HeadLoad: HeadLoadNone}

	if kurzUndSchwer.Weight(l) <= langUndLeicht.Weight(l) {
		t.Fatalf("acht Minuten mit Kopflast 3 (%d) sollten schwerer wiegen als 35 Minuten ohne (%d)",
			kurzUndSchwer.Weight(l), langUndLeicht.Weight(l))
	}
}

// -------------------------------------------------------------------- Vorlauf

func TestSaisonaufgabeErscheintMitVorlauf(t *testing.T) {
	// Reifenwechsel: Ziel ist der 1. Oktober, Vorlauf 21 Tage — die Aufgabe
	// muss also ab dem 10. September auftauchen, nicht erst im Oktober.
	tests := []struct {
		week Week
		want bool
	}{
		{Week{2026, 36}, false}, // 31.08.–06.09.
		{Week{2026, 38}, true},  // 14.09.–20.09., innerhalb des Vorlaufs
		{Week{2026, 40}, true},  // 28.09.–04.10., Ziel liegt in der Woche
	}

	for _, tc := range tests {
		t.Run(tc.week.String(), func(t *testing.T) {
			in := basisEingabe(haushaltMitTeenager(), reifenwechsel())
			in.Week = tc.week

			got, err := Plan(in)
			if err != nil {
				t.Fatal(err)
			}
			drin := len(got.Tasks) == 1
			if drin != tc.want {
				t.Fatalf("im Plan = %v, erwartet %v (übersprungen: %+v)", drin, tc.want, got.Skipped)
			}
		})
	}
}

// ------------------------------------------------------------------- Wochen

func TestWeekMonday(t *testing.T) {
	tests := []struct {
		week Week
		want string
	}{
		{Week{2026, 1}, "2025-12-29"},
		{Week{2026, 38}, "2026-09-14"},
		{Week{2026, 40}, "2026-09-28"},
	}
	for _, tc := range tests {
		if got := tc.week.Monday().String(); got != tc.want {
			t.Errorf("%s: Montag = %s, erwartet %s", tc.week, got, tc.want)
		}
	}
}

func TestValidateMeldetDoppelteIDs(t *testing.T) {
	in := basisEingabe(familieMitKleinkind(), spuelmaschine(), spuelmaschine())
	if _, err := Plan(in); err == nil {
		t.Fatal("doppelte Vorlagen-ID sollte einen Fehler ergeben")
	}
}
