package storage

import (
	"testing"
	"time"

	"github.com/zakaria/haushalt/api/internal/planner"
)

// haushaltMitZweien legt einen Haushalt an und gibt seine Kennung zurück.
func haushaltMitZweien(t *testing.T, subject string) string {
	t.Helper()
	leeren(t)

	h, err := testDB.AsPlans().Create(t.Context(), subject, planner.Setup{
		Name:     "Testhaushalt",
		Timezone: "Europe/Berlin",
		Context:  planner.Context{Home: planner.HomeFlat},
		Members: []planner.SetupMember{
			{Name: "Amir", Role: planner.RolePlanner},
			{Name: "Noor", Role: planner.RoleDoer},
		},
	})
	if err != nil {
		t.Fatalf("Haushalt anlegen: %v", err)
	}
	return h.ID
}

// vorlage sucht eine Vorlage im Stand dieses Haushalts.
func vorlage(t *testing.T, subject, haushalt, id string) planner.TemplateState {
	t.Helper()
	stand, _, err := testDB.AsPlans().TemplatesFor(t.Context(), subject, haushalt)
	if err != nil {
		t.Fatalf("TemplatesFor: %v", err)
	}
	for _, v := range stand {
		if v.Template.ID == id {
			return v
		}
	}
	t.Fatalf("Vorlage %q steht nicht im Stand des Haushalts", id)
	return planner.TemplateState{}
}

// Der Fall, der diese Tabelle nötig gemacht hat: Müll steht in der Bibliothek
// auf Dienstag, bei diesem Haushalt kommt die Tonne donnerstags.
func TestEigenerWochentagSchlaegtDieBibliothek(t *testing.T) {
	const subject = "auth-amir"
	haushalt := haushaltMitZweien(t, subject)
	plans := testDB.AsPlans()

	vorher := vorlage(t, subject, haushalt, "t-muell")
	if vorher.Template.Rhythm.Type != planner.RhythmFixed {
		t.Fatalf("t-muell hat den Rhythmus %q, erwartet %q",
			vorher.Template.Rhythm.Type, planner.RhythmFixed)
	}
	if len(vorher.Template.Rhythm.Weekdays) != 1 || vorher.Template.Rhythm.Weekdays[0] != time.Tuesday {
		t.Fatalf("t-muell steht auf %v, erwartet [Dienstag]", vorher.Template.Rhythm.Weekdays)
	}
	if vorher.WeekdaysOwn {
		t.Error("ohne eigene Festlegung gilt WeekdaysOwn als eigen")
	}

	// Donnerstag, Index 3.
	var donnerstag [7]bool
	donnerstag[3] = true
	if err := plans.SetWeekdays(t.Context(), subject, haushalt, "t-muell", donnerstag); err != nil {
		t.Fatalf("SetWeekdays: %v", err)
	}

	nachher := vorlage(t, subject, haushalt, "t-muell")
	if len(nachher.Template.Rhythm.Weekdays) != 1 || nachher.Template.Rhythm.Weekdays[0] != time.Thursday {
		t.Errorf("t-muell steht auf %v, erwartet [Donnerstag]", nachher.Template.Rhythm.Weekdays)
	}
	if !nachher.WeekdaysOwn {
		t.Error("die eigene Festlegung wird nicht als eigen gemeldet")
	}

	// Und der Plan zieht mit. Das ist der eigentliche Punkt: Die Liste zu
	// ändern, ohne dass die Woche folgt, wäre die teurere Hälfte des Fehlers.
	woche := planner.WeekOf(planner.DateOf(time.Date(2026, 10, 5, 0, 0, 0, 0, time.UTC)))
	ergebnis, _, err := plans.Plan(t.Context(), subject, haushalt, woche)
	if err != nil {
		t.Fatalf("Plan: %v", err)
	}
	gefunden := false
	for _, a := range ergebnis.Tasks {
		if a.TemplateID != "t-muell" {
			continue
		}
		gefunden = true
		if a.Day.Weekday() != time.Thursday {
			t.Errorf("Müll steht am %v, erwartet Donnerstag", a.Day.Weekday())
		}
	}
	if !gefunden {
		t.Error("Müll steht gar nicht im Plan")
	}
}

// Kein Tag angekreuzt heißt: zurück zur Bibliothek. Es gibt keinen dritten
// Zustand (ADR-0017).
func TestKeinTagSetztZurueck(t *testing.T) {
	const subject = "auth-amir"
	haushalt := haushaltMitZweien(t, subject)
	plans := testDB.AsPlans()

	var donnerstag [7]bool
	donnerstag[3] = true
	if err := plans.SetWeekdays(t.Context(), subject, haushalt, "t-muell", donnerstag); err != nil {
		t.Fatalf("SetWeekdays: %v", err)
	}
	if err := plans.SetWeekdays(t.Context(), subject, haushalt, "t-muell", [7]bool{}); err != nil {
		t.Fatalf("SetWeekdays leer: %v", err)
	}

	zurueck := vorlage(t, subject, haushalt, "t-muell")
	if len(zurueck.Template.Rhythm.Weekdays) != 1 || zurueck.Template.Rhythm.Weekdays[0] != time.Tuesday {
		t.Errorf("t-muell steht auf %v, erwartet wieder [Dienstag]", zurueck.Template.Rhythm.Weekdays)
	}
	if zurueck.WeekdaysOwn {
		t.Error("nach dem Zurücksetzen gilt die Vorgabe noch als eigen")
	}
}

// Die Grenze steht im Dienst und nicht nur im Bildschirm: Beim Auslöser ist
// ein Wochentag keine Antwort, sondern eine andere Frage.
func TestAusloeserBekommtKeinenWochentag(t *testing.T) {
	const subject = "auth-amir"
	haushalt := haushaltMitZweien(t, subject)
	plans := testDB.AsPlans()

	// Irgendeine Vorlage mit Auslöser-Rhythmus aus der echten Bibliothek.
	stand, _, err := plans.TemplatesFor(t.Context(), subject, haushalt)
	if err != nil {
		t.Fatalf("TemplatesFor: %v", err)
	}
	var ausloeser string
	for _, v := range stand {
		if v.Template.Rhythm.Type == planner.RhythmTrigger {
			ausloeser = v.Template.ID
			break
		}
	}
	if ausloeser == "" {
		t.Skip("die Bibliothek enthält keine Auslöser-Vorlage mehr")
	}

	var samstag [7]bool
	samstag[5] = true
	err = plans.SetWeekdays(t.Context(), subject, haushalt, ausloeser, samstag)
	if err == nil {
		t.Fatalf("%q bekam einen Wochentag, obwohl sie einen Auslöser hat", ausloeser)
	}
}

// Wer nur ausführt, legt keine Tage fest — der Tag gilt für alle.
func TestNurPlanendeLegenTageFest(t *testing.T) {
	const planend = "auth-amir"
	haushalt := haushaltMitZweien(t, planend)

	var donnerstag [7]bool
	donnerstag[3] = true
	err := testDB.AsPlans().SetWeekdays(t.Context(), "auth-fremd", haushalt, "t-muell", donnerstag)
	if err == nil {
		t.Fatal("ein fremder Aufrufer durfte den Tag festlegen")
	}
}
