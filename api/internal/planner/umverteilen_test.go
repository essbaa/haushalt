package planner

import (
	"errors"
	"strings"
	"testing"
)

// mitKennungen gibt den Aufgaben eines gerechneten Plans Kennungen. Der Planer
// vergibt selbst keine — er rechnet, er speichert nicht.
func mitKennungen(tasks []PlannedTask) []PlannedTask {
	out := make([]PlannedTask, len(tasks))
	copy(out, tasks)
	for i := range out {
		out[i].ID = "a-" + out[i].TemplateID
	}
	return out
}

func TestUmverteilenGehtAnJedeGeeignetePerson(t *testing.T) {
	in := basisEingabe(haushaltMitTeenager(), bad())
	got, err := Plan(in)
	if err != nil {
		t.Fatal(err)
	}
	tasks := mitKennungen(got.Tasks)

	grund, err := Reassign(in, tasks, "a-t-bad", "m-teen")
	if err != nil {
		t.Fatalf("das Bad darf an den Teenager: %v", err)
	}
	if grund.Code != ReasonManual {
		t.Errorf("Begründung %q, erwartet %q", grund.Code, ReasonManual)
	}
	if grund.Previous != tasks[0].AssigneeID {
		t.Errorf("Vorgänger %q, erwartet %q", grund.Previous, tasks[0].AssigneeID)
	}
}

// TestUmverteilenStichtDieVorliebenDesPlaners ist der eigentliche Test dieser
// Datei: Ein Mensch, der umverteilt, muss gegen die Annahmen des Planers
// gewinnen können. Sonst ist `manual = true` keine Rückmeldung, sondern eine
// Genehmigung.
func TestUmverteilenStichtDieVorliebenDesPlaners(t *testing.T) {
	in := basisEingabe(haushaltMitTeenager(), bad())
	// Der Teenager hatte das Bad letzte Woche — die Rotation würde es ihm
	// nicht wiedergeben.
	in.History = History{LastAssignee: map[string]string{"t-bad": "m-teen"}}

	got, err := Plan(in)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := Reassign(in, mitKennungen(got.Tasks), "a-t-bad", "m-teen"); err != nil {
		t.Fatalf("die Rotation darf eine Umverteilung von Hand nicht verhindern: %v", err)
	}
}

func TestUmverteilenPrueftDasAlter(t *testing.T) {
	abAchtzehn := bad()
	abAchtzehn.MinAge = 18

	in := basisEingabe(haushaltMitTeenager(), abAchtzehn)
	got, err := Plan(in)
	if err != nil {
		t.Fatal(err)
	}

	_, err = Reassign(in, mitKennungen(got.Tasks), "a-t-bad", "m-teen")
	if !errors.Is(err, ErrNotEligible) {
		t.Fatalf("Fehler %v, erwartet ErrNotEligible", err)
	}
	if !strings.Contains(err.Error(), "18") {
		t.Errorf("die Meldung nennt die Altersgrenze nicht: %q", err)
	}
}

func TestUmverteilenKenntKeineFremdePerson(t *testing.T) {
	in := basisEingabe(haushaltMitTeenager(), bad())
	got, err := Plan(in)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := Reassign(in, mitKennungen(got.Tasks), "a-t-bad", "m-gibt-es-nicht"); !errors.Is(err, ErrUnknownMember) {
		t.Fatalf("Fehler %v, erwartet ErrUnknownMember", err)
	}
}

func TestUmverteilenLaesstEigeneAufgabenInRuhe(t *testing.T) {
	zimmer := TaskTemplate{
		ID: "t-zimmer", Title: "Eigenes Zimmer aufräumen",
		Category: CatCleaning, Kind: KindDo,
		DurationMin: 25, HeadLoad: HeadLoadLow,
		Rhythm:       Rhythm{Type: RhythmWindow, EveryDays: 7},
		MinAge:       6,
		PerPerson:    true,
		Distribution: DistRotate,
		Failure:      FailureSoft,
		Source:       SourceCurated,
	}

	in := basisEingabe(haushaltMitTeenager(), zimmer)
	got, err := Plan(in)
	if err != nil {
		t.Fatal(err)
	}
	if len(got.Tasks) == 0 {
		t.Fatalf("keine Aufgaben (übersprungen: %+v)", got.Skipped)
	}
	tasks := mitKennungen(got.Tasks)
	ziel := "m-eltern-1"
	if tasks[0].AssigneeID == ziel {
		ziel = "m-eltern-2"
	}
	if _, err := Reassign(in, tasks, tasks[0].ID, ziel); !errors.Is(err, ErrNotEligible) {
		t.Fatalf("Fehler %v, erwartet ErrNotEligible — eine eigene Aufgabe wandert nicht", err)
	}
}
