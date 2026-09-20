package httpapi

import (
	"testing"
	"time"

	"github.com/zakaria/haushalt/api/internal/planner"
)

// Die Woche fängt zweimal an verschiedenen Tagen an: In Go ist time.Sunday
// die 0, im Vertrag und in der Datenbank ist es Montag. Genau hier entsteht
// der Fehler, den niemand sieht — die Aufgabe steht einen Tag daneben, und
// das sieht aus wie eine Entscheidung des Planers.
func TestMontagIstNull(t *testing.T) {
	faelle := []struct {
		tag   time.Weekday
		index int
	}{
		{time.Monday, 0},
		{time.Tuesday, 1},
		{time.Wednesday, 2},
		{time.Thursday, 3},
		{time.Friday, 4},
		{time.Saturday, 5},
		{time.Sunday, 6},
	}

	for _, f := range faelle {
		tage := wochentageNachAussen(planner.Rhythm{
			Type:     planner.RhythmFixed,
			Weekdays: []time.Weekday{f.tag},
		})
		if len(tage) != 7 {
			t.Fatalf("%v: %d Werte, erwartet 7", f.tag, len(tage))
		}
		for i, an := range tage {
			if an != (i == f.index) {
				t.Errorf("%v: Index %d ist %v, erwartet %v", f.tag, i, an, i == f.index)
			}
		}
	}
}

// Mehrere Tage an einer Vorlage — Abendessen kochen steht montags, mittwochs
// und freitags.
func TestMehrereTage(t *testing.T) {
	tage := wochentageNachAussen(planner.Rhythm{
		Type:     planner.RhythmFixed,
		Weekdays: []time.Weekday{time.Monday, time.Wednesday, time.Friday},
	})
	erwartet := []bool{true, false, true, false, true, false, false}
	for i := range erwartet {
		if tage[i] != erwartet[i] {
			t.Errorf("Index %d ist %v, erwartet %v", i, tage[i], erwartet[i])
		}
	}
}

// Ein Rhythmus ohne feste Tage hat keine — und liefert trotzdem sieben Werte.
//
// Der Vertrag sagt „minItems 7"; ein nil-Feld wäre dort ein `null`, und die
// Oberfläche müsste den Fall kennen. Sieben falsche Werte sind dieselbe
// Auskunft ohne Sonderfall.
func TestOhneFesteTageSiebenmalFalsch(t *testing.T) {
	for _, art := range []planner.RhythmType{
		planner.RhythmWindow,
		planner.RhythmTrigger,
		planner.RhythmSeason,
		planner.RhythmPhase,
	} {
		tage := wochentageNachAussen(planner.Rhythm{Type: art, EveryDays: 7})
		if len(tage) != 7 {
			t.Fatalf("%s: %d Werte, erwartet 7", art, len(tage))
		}
		for i, an := range tage {
			if an {
				t.Errorf("%s: Index %d ist gesetzt", art, i)
			}
		}
	}
}
