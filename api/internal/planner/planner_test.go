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

// ------------------------------------------------------------------ Ausgleich

// TestZweiterDurchgangGleichtAus beschreibt den Fall aus ADR-0003 in Zahlen.
//
// Drei Aufgaben, zwei Erwachsene, keine Historie — und eine Verteilung, die
// aufgehen kann: zwei Organisationsaufgaben zu je 10 Minuten mit Kopflast 3
// (gewichtet je 10 + 3×15 = 55) und eine Putzaufgabe über 110 Minuten. Zusammen
// 220 gewichtete Minuten, also 110 für jeden.
//
// Der gierige Durchgang allein kommt auf 165 zu 55: Er verteilt die beiden
// kleinen Aufgaben auf beide Personen und legt die große danach auf die, die
// zufällig zuerst dran war. Erst ein Tausch macht daraus 110 zu 110.
func TestZweiterDurchgangGleichtAus(t *testing.T) {
	org := func(id string) TaskTemplate {
		return TaskTemplate{
			ID: id, Title: "Organisation " + id,
			Category: CatAdmin, Kind: KindOrg,
			DurationMin: 10, HeadLoad: HeadLoadHigh,
			Rhythm:       Rhythm{Type: RhythmWindow, EveryDays: 7},
			Distribution: DistAdultsOnly,
			Failure:      FailureSoft,
			Source:       SourceCurated,
		}
	}
	putzen := TaskTemplate{
		ID: "t-grossputz", Title: "Wohnung gründlich putzen",
		Category: CatCleaning, Kind: KindDo,
		DurationMin: 110, HeadLoad: HeadLoadNone,
		Rhythm:       Rhythm{Type: RhythmWindow, EveryDays: 7},
		MinAge:       12,
		Distribution: DistRotate,
		Failure:      FailureSoft,
		Source:       SourceCurated,
	}

	got, err := Plan(basisEingabe(familieMitKleinkind(), org("t-org-a"), org("t-org-b"), putzen))
	if err != nil {
		t.Fatal(err)
	}
	if len(got.Tasks) != 3 {
		t.Fatalf("%d Aufgaben, erwartet 3 (übersprungen: %+v)", len(got.Tasks), got.Skipped)
	}

	for _, l := range got.Balance {
		if l.Weighted != 110 {
			t.Errorf("%s trägt %d gewichtete Minuten, erwartet 110 — die Verteilung geht auf, der Planer findet sie nur nicht",
				l.MemberID, l.Weighted)
		}
	}
}

// TestAusgleichRespektiertRotation stellt sicher, dass der zweite Durchgang
// die wichtigere Regel nicht überfährt. Anna hatte das Bad zuletzt; auch wenn
// ein Tausch die Last gleichmäßiger machen würde, darf es nicht zu ihr
// zurückwandern.
func TestAusgleichRespektiertRotation(t *testing.T) {
	klein := TaskTemplate{
		ID: "t-klein", Title: "Küche wischen",
		Category: CatKitchen, Kind: KindDo,
		DurationMin: 10, HeadLoad: HeadLoadNone,
		Rhythm:       Rhythm{Type: RhythmWindow, EveryDays: 7},
		MinAge:       12,
		Distribution: DistRotate,
		Failure:      FailureSoft,
		Source:       SourceCurated,
	}

	in := basisEingabe(familieMitKleinkind(), bad(), klein)
	in.History = History{LastAssignee: map[string]string{"t-bad": "m-anna"}}

	got, err := Plan(in)
	if err != nil {
		t.Fatal(err)
	}
	for _, task := range got.Tasks {
		if task.TemplateID == "t-bad" && task.AssigneeID == "m-anna" {
			t.Fatalf("das Bad ist zu Anna zurückgetauscht worden — Rotation geht vor Ausgleich")
		}
	}
}

// TestAusgleichHaeuftKeineTermineDerselbenVorlageAn ist die Regressionsprüfung
// zu einem Fehler, den die Tests nicht gefunden haben — er stand im Wochenplan
// auf der Konsole: Anna kochte an allen drei Abenden, weil der Ausgleich das
// zusammengeführt hatte, was die Zuteilung getrennt hatte.
//
// Der Aufbau hier ist so gewählt, dass der Tausch verlockend ist: Anna trägt
// 180 gewichtete Minuten, Ben 70. Zöge das Kochen vom Mittwoch zu Anna und der
// Großputz zu Ben, stünde es 140 zu 110 — deutlich gleichmäßiger. Trotzdem darf
// es nicht passieren.
func TestAusgleichHaeuftKeineTermineDerselbenVorlageAn(t *testing.T) {
	kochen := TaskTemplate{
		ID: "t-kochen", Title: "Abendessen kochen",
		Category: CatKitchen, Kind: KindDo,
		DurationMin: 40, HeadLoad: HeadLoadMid,
		Rhythm:       Rhythm{Type: RhythmFixed, Weekdays: []time.Weekday{time.Monday, time.Wednesday}},
		MinAge:       12,
		Distribution: DistRotate,
		Failure:      FailureSoft,
		Source:       SourceCurated,
	}
	gross := TaskTemplate{
		ID: "t-grossputz", Title: "Wohnung gründlich putzen",
		Category: CatCleaning, Kind: KindDo,
		DurationMin: 110, HeadLoad: HeadLoadNone,
		Rhythm:       Rhythm{Type: RhythmWindow, EveryDays: 7},
		MinAge:       12,
		Distribution: DistRotate,
		Failure:      FailureSoft,
		Source:       SourceCurated,
	}

	in := basisEingabe(familieMitKleinkind(), kochen, gross)
	tasks := []PlannedTask{
		{TemplateID: "t-kochen", Title: kochen.Title, Kind: KindDo, Day: MustDate("2026-09-14"),
			DurationMin: 40, HeadLoad: HeadLoadMid, AssigneeID: "m-anna"},
		{TemplateID: "t-kochen", Title: kochen.Title, Kind: KindDo, Day: MustDate("2026-09-16"),
			DurationMin: 40, HeadLoad: HeadLoadMid, AssigneeID: "m-ben"},
		{TemplateID: "t-grossputz", Title: gross.Title, Kind: KindDo, Day: MustDate("2026-09-19"),
			DurationMin: 110, HeadLoad: HeadLoadNone, AssigneeID: "m-anna"},
	}

	got := rebalance(in, tasks)

	proPerson := map[string]int{}
	for _, task := range got {
		if task.TemplateID == "t-kochen" {
			proPerson[task.AssigneeID]++
		}
	}
	for id, n := range proPerson {
		if n > 1 {
			t.Fatalf("%s kocht %d von 2 Abenden — der Ausgleich hat die Rotation innerhalb der Woche überfahren", id, n)
		}
	}
}

// ------------------------------------------------------------- Auslastung

// erwachseneUndTeenager: zwei Personen mit sehr unterschiedlicher Kapazität —
// 660 Minuten die Woche gegen 330. Der Zuschnitt, an dem sich zeigt, ob der
// Planer Belastung oder nur Minuten vergleicht.
func erwachseneUndTeenager() Household {
	return Household{
		ID: "hh-3",
		Members: []Member{
			{ID: "m-erw", Name: "Erwachsene", Role: RolePlanner, Age: 40,
				CapacityMinutes: [7]int{60, 60, 60, 60, 60, 180, 180}},
			{ID: "m-teen", Name: "Teenager", Role: RoleDoer, Age: 14, Care: CareSchool,
				CapacityMinutes: [7]int{30, 30, 30, 30, 30, 90, 90}},
		},
		Context: Context{Home: HomeFlat},
	}
}

// TestAuslastungSchlaegtAbsoluteMinuten prüft die Kennzahl selbst: Von zwei
// Plänen mit derselben Gesamtlast muss der gewinnen, der sie im Verhältnis zur
// verfügbaren Zeit verteilt — nicht der, der gleich viele Minuten vergibt.
func TestAuslastungSchlaegtAbsoluteMinuten(t *testing.T) {
	in := Input{Household: erwachseneUndTeenager(), Week: Week{2026, 38}, Limits: DefaultLimits()}
	tag := MustDate("2026-09-14")

	aufgabe := func(id, wer string) PlannedTask {
		return PlannedTask{TemplateID: id, Kind: KindDo, Day: tag, DurationMin: 30, AssigneeID: wer}
	}

	// 180 Minuten insgesamt, Kapazität 660 zu 330 — der gerechte Anteil ist
	// also 120 zu 60, beide bei 18 Prozent.
	nachAuslastung := []PlannedTask{
		aufgabe("t-1", "m-erw"), aufgabe("t-2", "m-erw"),
		aufgabe("t-3", "m-erw"), aufgabe("t-4", "m-erw"),
		aufgabe("t-5", "m-teen"), aufgabe("t-6", "m-teen"),
	}
	// Gleich viele Minuten für beide: 90 zu 90 — und damit 14 Prozent gegen
	// 27 Prozent. Genau der Plan, den die alte Kennzahl bevorzugte.
	nachMinuten := []PlannedTask{
		aufgabe("t-1", "m-erw"), aufgabe("t-2", "m-erw"), aufgabe("t-3", "m-erw"),
		aufgabe("t-4", "m-teen"), aufgabe("t-5", "m-teen"), aufgabe("t-6", "m-teen"),
	}

	if imbalanceOf(in, nachAuslastung) >= imbalanceOf(in, nachMinuten) {
		t.Fatalf("gleiche Minuten (%d) gelten als mindestens so fair wie gleiche Auslastung (%d)",
			imbalanceOf(in, nachMinuten), imbalanceOf(in, nachAuslastung))
	}
}

// TestKopflastEntscheidetMit prüft die zweite Hälfte von ADR-0004: Bei
// identischer gewichteter Last muss der Plan gewinnen, der die Kopfarbeit
// teilt. Unter der alten Regel entschied die Kopflast nur bei exakter
// Gleichheit — was praktisch nie eintrat.
func TestKopflastEntscheidetMit(t *testing.T) {
	in := Input{Household: familieMitKleinkind(), Week: Week{2026, 38}, Limits: DefaultLimits()}
	tag := MustDate("2026-09-14")

	// Vier Aufgaben, jede gewichtet 55: zwei kurze mit Kopflast 3, zwei lange
	// ohne. Wie man sie auch aufteilt, jede Person kommt auf 110.
	kopf := func(id, wer string) PlannedTask {
		return PlannedTask{TemplateID: id, Kind: KindOrg, Day: tag, DurationMin: 10, HeadLoad: HeadLoadHigh, AssigneeID: wer}
	}
	hand := func(id, wer string) PlannedTask {
		return PlannedTask{TemplateID: id, Kind: KindDo, Day: tag, DurationMin: 55, AssigneeID: wer}
	}

	geteilt := []PlannedTask{
		kopf("t-k1", "m-anna"), hand("t-h1", "m-anna"),
		kopf("t-k2", "m-ben"), hand("t-h2", "m-ben"),
	}
	gebuendelt := []PlannedTask{
		kopf("t-k1", "m-anna"), kopf("t-k2", "m-anna"),
		hand("t-h1", "m-ben"), hand("t-h2", "m-ben"),
	}

	if imbalanceOf(in, geteilt) >= imbalanceOf(in, gebuendelt) {
		t.Fatalf("die gesamte Kopfarbeit bei einer Person (%d) gilt als ebenso fair wie geteilte (%d)",
			imbalanceOf(in, gebuendelt), imbalanceOf(in, geteilt))
	}
}

// TestJugendlicheTragenAnteiligWeniger prüft dasselbe durch den ganzen Planer:
// Sechs gleich große Aufgaben, zwei Personen mit doppelt so großem
// Kapazitätsunterschied — am Ende müssen beide ähnlich ausgelastet sein, nicht
// gleich viele Minuten haben.
func TestJugendlicheTragenAnteiligWeniger(t *testing.T) {
	var vorlagen []TaskTemplate
	for _, id := range []string{"t-1", "t-2", "t-3", "t-4", "t-5", "t-6"} {
		vorlagen = append(vorlagen, TaskTemplate{
			ID: id, Title: "Aufgabe " + id,
			Category: CatCleaning, Kind: KindDo,
			DurationMin: 30, HeadLoad: HeadLoadNone,
			Rhythm:       Rhythm{Type: RhythmWindow, EveryDays: 7},
			MinAge:       12,
			Distribution: DistRotate,
			Failure:      FailureSoft,
			Source:       SourceCurated,
		})
	}

	got, err := Plan(basisEingabe(erwachseneUndTeenager(), vorlagen...))
	if err != nil {
		t.Fatal(err)
	}
	if len(got.Tasks) != 6 {
		t.Fatalf("%d Aufgaben, erwartet 6 (übersprungen: %+v)", len(got.Tasks), got.Skipped)
	}

	last := map[string]MemberLoad{}
	for _, l := range got.Balance {
		last[l.MemberID] = l
	}
	erw, teen := last["m-erw"], last["m-teen"]

	if abstand(erw.Utilization(), teen.Utilization()) > 5 {
		t.Errorf("Auslastung %d %% gegen %d %% — der Unterschied gehört unter 5 Punkte",
			erw.Utilization(), teen.Utilization())
	}
	if teen.Weighted >= erw.Weighted {
		t.Errorf("der Teenager trägt %d gewichtete Minuten, die Erwachsene %d — bei halber Kapazität gehört das umgekehrt",
			teen.Weighted, erw.Weighted)
	}
}

func abstand(a, b int) int {
	if a > b {
		return a - b
	}
	return b - a
}

// --------------------------------------------------- Eigene Aufgaben, Tiere

// TestHaustierartGrenztEin: Ein Haushalt mit Hund bekommt kein Katzenklo.
// Vorher war die Bedingung ein Ja/Nein — der Haushalt wusste, dass es ein Hund
// ist, die Vorlage konnte aber nicht danach fragen.
func TestHaustierartGrenztEin(t *testing.T) {
	tier := func(id, titel, art string) TaskTemplate {
		return TaskTemplate{
			ID: id, Title: titel,
			Category: CatCleaning, Kind: KindDo,
			DurationMin: 10, HeadLoad: HeadLoadNone,
			Rhythm:       Rhythm{Type: RhythmWindow, EveryDays: 7},
			MinAge:       10,
			Distribution: DistRotate,
			AppliesTo:    Conditions{RequiresPetKind: art},
			Failure:      FailureSoft,
			Source:       SourceCurated,
		}
	}

	mitHund := familieMitKleinkind()
	mitHund.Context.Pets = []string{"hund"}

	got, err := Plan(basisEingabe(mitHund,
		tier("t-katzenklo", "Katzenklo säubern", "katze"),
		tier("t-gassi", "Mit dem Hund rausgehen", "hund")))
	if err != nil {
		t.Fatal(err)
	}

	if len(got.Tasks) != 1 || got.Tasks[0].TemplateID != "t-gassi" {
		t.Fatalf("im Plan steht %+v, erwartet wurde allein t-gassi", got.Tasks)
	}
	var grund SkipCode
	for _, s := range got.Skipped {
		if s.TemplateID == "t-katzenklo" {
			grund = s.Code
		}
	}
	if grund != SkipNotApplicable {
		t.Errorf("das Katzenklo fällt mit %q heraus, erwartet %q", grund, SkipNotApplicable)
	}
}

// TestEigeneAufgabeGehoertJedemSelbst prüft die PerPerson-Vervielfachung:
// Zwei Kinder ergeben zwei Aufgaben, je eine fest an das eigene Kind gebunden
// — und keine bei den Erwachsenen.
func TestEigeneAufgabeGehoertJedemSelbst(t *testing.T) {
	zimmer := TaskTemplate{
		ID: "t-zimmer", Title: "Eigenes Zimmer aufräumen",
		Category: CatCleaning, Kind: KindDo,
		DurationMin: 25, HeadLoad: HeadLoadLow,
		Rhythm:       Rhythm{Type: RhythmWindow, EveryDays: 7},
		MinAge:       8,
		Distribution: DistChildrenOnly,
		PerPerson:    true,
		Failure:      FailureSoft,
		Source:       SourceCurated,
	}

	h := Household{
		ID: "hh-4",
		Members: []Member{
			{ID: "m-eltern", Role: RolePlanner, Age: 44,
				CapacityMinutes: [7]int{60, 60, 60, 60, 60, 180, 180}},
			{ID: "m-kind-1", Role: RoleDoer, Age: 15, Care: CareSchool,
				CapacityMinutes: [7]int{30, 30, 30, 30, 30, 90, 90}},
			{ID: "m-kind-2", Role: RoleDoer, Age: 12, Care: CareSchool,
				CapacityMinutes: [7]int{25, 25, 25, 25, 25, 80, 80}},
		},
		Context: Context{Home: HomeFlat},
	}

	got, err := Plan(basisEingabe(h, zimmer))
	if err != nil {
		t.Fatal(err)
	}
	if len(got.Tasks) != 2 {
		t.Fatalf("%d Aufgaben, erwartet 2 — je Kind eine (übersprungen: %+v)", len(got.Tasks), got.Skipped)
	}

	wer := map[string]int{}
	for _, task := range got.Tasks {
		wer[task.AssigneeID]++
		if task.Reason.Code != ReasonOwn {
			t.Errorf("Begründung %q, erwartet %q", task.Reason.Code, ReasonOwn)
		}
	}
	if wer["m-kind-1"] != 1 || wer["m-kind-2"] != 1 {
		t.Errorf("Verteilung %v — jedes Kind räumt genau sein eigenes Zimmer auf", wer)
	}
	if wer["m-eltern"] != 0 {
		t.Errorf("das Elternteil räumt das Zimmer eines Kindes auf: %v", wer)
	}

	// Und im Haushalt mit einem Dreijährigen gibt es die Aufgabe gar nicht.
	ohne, err := Plan(basisEingabe(familieMitKleinkind(), zimmer))
	if err != nil {
		t.Fatal(err)
	}
	if len(ohne.Tasks) != 0 {
		t.Errorf("%d Aufgaben, erwartet 0 — ein Dreijähriger räumt sein Zimmer nicht selbst auf", len(ohne.Tasks))
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
