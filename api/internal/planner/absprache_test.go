package planner

import (
	"testing"
	"time"
)

// bringen ist die Aufgabe, an der sich die ganze Absprache entscheidet: Wer
// ein Kind um 7:45 in die Kita bringen kann, hängt am Arbeitsplan und nicht
// daran, wer an dem Tag Minuten übrig hat (ADR-0016).
func bringen() TaskTemplate {
	return TaskTemplate{
		ID: "t-bringen", Title: "Zur Kita bringen",
		Category: CatChild, Kind: KindDo,
		DurationMin: 25, HeadLoad: HeadLoadLow,
		Rhythm: Rhythm{
			Type:     RhythmFixed,
			Weekdays: []time.Weekday{time.Monday, time.Tuesday, time.Wednesday},
		},
		Slot:           SlotMorning,
		Distribution:   DistRotate,
		Failure:        FailureHard,
		NeedsAgreement: true,
		Source:         SourceCurated,
	}
}

func TestOhneAbspracheWirdNichtGeraten(t *testing.T) {
	in := basisEingabe(haushaltMitTeenager(), bringen())

	got, err := Plan(in)
	if err != nil {
		t.Fatal(err)
	}
	if len(got.Tasks) != 0 {
		t.Fatalf("%d Aufgaben verteilt, erwartet keine — der Planer kennt den Arbeitsplan nicht", len(got.Tasks))
	}
	if len(got.Skipped) == 0 {
		t.Fatal("nichts übersprungen — dann wäre die Aufgabe wortlos verschwunden")
	}
	for _, s := range got.Skipped {
		if s.Code != SkipNeedsAgreement {
			t.Errorf("Grund %q, erwartet %q", s.Code, SkipNeedsAgreement)
		}
	}
}

func TestAbspracheStichtRotationUndAuslastung(t *testing.T) {
	in := basisEingabe(haushaltMitTeenager(), bringen())
	// Beides spräche gegen m-eltern-1: Er hatte die Aufgabe zuletzt, und die
	// Rotation würde sie ihm nicht wiedergeben. Die Absprache sagt trotzdem
	// ihn — sie weiß etwas, das der Planer nicht wissen kann.
	in.History = History{
		LastAssignee: map[string]string{"t-bringen": "m-eltern-1"},
		AgreedTo: map[string][7]string{
			"t-bringen": {"m-eltern-1", "m-eltern-2", "m-eltern-1", "", "", "", ""},
		},
	}

	got, err := Plan(in)
	if err != nil {
		t.Fatal(err)
	}
	if len(got.Tasks) != 3 {
		t.Fatalf("%d Aufgaben, erwartet 3 (Mo, Di, Mi)", len(got.Tasks))
	}

	will := map[time.Weekday]string{
		time.Monday:    "m-eltern-1",
		time.Tuesday:   "m-eltern-2",
		time.Wednesday: "m-eltern-1",
	}
	for _, a := range got.Tasks {
		wd := a.Day.Weekday()
		if a.AssigneeID != will[wd] {
			t.Errorf("%v: %q, abgesprochen war %q", wd, a.AssigneeID, will[wd])
		}
		if a.Reason.Code != ReasonAgreed {
			t.Errorf("%v: Begründung %q, erwartet %q", wd, a.Reason.Code, ReasonAgreed)
		}
	}
}

// TestAbspracheZaehltInDieBilanz ist der Test, der aus der Absprache mehr
// macht als eine Notiz: Wer dreimal die Woche fährt, hat drei Wege geleistet.
// Sie nicht zu verbuchen, weil der Planer sie nicht verteilt hat, würde genau
// die Arbeit unsichtbar machen, um die es in diesem Produkt geht.
func TestAbspracheZaehltInDieBilanz(t *testing.T) {
	in := basisEingabe(haushaltMitTeenager(), bringen(), bad())
	in.History = History{
		AgreedTo: map[string][7]string{
			"t-bringen": {"m-eltern-1", "m-eltern-1", "m-eltern-1", "", "", "", ""},
		},
	}

	got, err := Plan(in)
	if err != nil {
		t.Fatal(err)
	}

	last := map[string]MemberLoad{}
	for _, b := range got.Balance {
		last[b.MemberID] = b
	}
	if got := last["m-eltern-1"].Minutes; got < 75 {
		t.Errorf("%d Minuten für m-eltern-1, erwartet mindestens 75 (3 × 25)", got)
	}
	if got := last["m-eltern-1"].Tasks; got < 3 {
		t.Errorf("%d Aufgaben für m-eltern-1, erwartet mindestens 3", got)
	}

	// Und die Gegenprobe: Das Bad geht an jemand anderen. Wer die Woche über
	// jeden Morgen fährt, soll abends weniger bekommen — genau dafür wird die
	// Absprache mitverbucht.
	for _, a := range got.Tasks {
		if a.TemplateID == "t-bad" && a.AssigneeID == "m-eltern-1" {
			t.Error("das Bad geht an dieselbe Person, die jeden Morgen fährt")
		}
	}
}

// TestVeralteteAbspracheWirdNichtStillGeloest: Steht im Raster jemand, den es
// im Haushalt nicht (mehr) gibt, teilt der Planer den Tag nicht ersatzweise
// zu. Das ist eine Frage an die Menschen — und sie taucht als übersprungene
// Aufgabe auf, statt still bei jemand anderem zu landen.
func TestVeralteteAbspracheWirdNichtStillGeloest(t *testing.T) {
	in := basisEingabe(haushaltMitTeenager(), bringen())
	in.History = History{
		AgreedTo: map[string][7]string{
			"t-bringen": {"m-ausgezogen", "", "", "", "", "", ""},
		},
	}

	got, err := Plan(in)
	if err != nil {
		t.Fatal(err)
	}
	if len(got.Tasks) != 0 {
		t.Fatalf("%d Aufgaben, erwartet keine — die Absprache ist veraltet", len(got.Tasks))
	}
	if len(got.Skipped) == 0 {
		t.Fatal("nichts übersprungen")
	}
}

// TestAusgleichRuehrtAbspracheNichtAn ist die Regel, die beim ersten Lauf
// gefehlt hat — und der Fehler war der schlimmstmögliche: Der zweite
// Durchgang tauschte den Montag an die andere Person, weil die Zahlen dadurch
// aufgingen, und schrieb als Begründung „zum Ausgleich" daneben. Im Plan
// stand damit jemand, der um 7:45 gar nicht dort sein kann — mitsamt einer
// Behauptung über eine Vereinbarung, die niemand getroffen hat.
func TestAusgleichRuehrtAbspracheNichtAn(t *testing.T) {
	in := basisEingabe(haushaltMitTeenager(), bringen(), bad())
	in.History = History{
		AgreedTo: map[string][7]string{
			"t-bringen": {"m-eltern-1", "m-eltern-1", "m-eltern-1", "", "", "", ""},
		},
	}

	got, err := Plan(in)
	if err != nil {
		t.Fatal(err)
	}
	for _, a := range got.Tasks {
		if a.TemplateID != "t-bringen" {
			continue
		}
		if a.AssigneeID != "m-eltern-1" {
			t.Errorf("%v bei %q, abgesprochen war m-eltern-1", a.Day.Weekday(), a.AssigneeID)
		}
		if a.Reason.Code != ReasonAgreed {
			t.Errorf("%v: Begründung %q, erwartet %q", a.Day.Weekday(), a.Reason.Code, ReasonAgreed)
		}
	}
}

// TestAbgesprochenesGibtDerPlanerNichtWeiter: Abgeben sucht die Person mit den
// meisten freien Minuten — genau die Größe, die hier nicht entscheidet. Die
// Aufgabe steht danach offen da, und das ist die ehrliche Antwort: eine
// sichtbare Lücke statt einer Besetzung, die keine ist.
func TestAbgesprochenesGibtDerPlanerNichtWeiter(t *testing.T) {
	in := basisEingabe(haushaltMitTeenager(), bringen())
	in.History = History{
		AgreedTo: map[string][7]string{
			"t-bringen": {"m-eltern-1", "", "", "", "", "", ""},
		},
	}

	got, err := Plan(in)
	if err != nil {
		t.Fatal(err)
	}
	tasks := mitKennungen(got.Tasks)
	if len(tasks) != 1 {
		t.Fatalf("%d Aufgaben, erwartet 1", len(tasks))
	}
	if _, _, ok := Handover(in, tasks, tasks[0].ID); ok {
		t.Error("der Planer hat eine abgesprochene Aufgabe weitergereicht")
	}
}
