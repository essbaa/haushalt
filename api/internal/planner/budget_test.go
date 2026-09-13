package planner

import "testing"

func TestTimeBudgetMinutes(t *testing.T) {
	t.Run("Wochenende ist nie knapper als der Werktag", func(t *testing.T) {
		for _, b := range []TimeBudget{BudgetLow, BudgetMedium, BudgetHigh} {
			m := b.Minutes()
			if m[5] < m[0] || m[6] < m[0] {
				t.Errorf("%s: Wochenende %d/%d unter Werktag %d", b, m[5], m[6], m[0])
			}
		}
	})

	t.Run("Stufen sind geordnet", func(t *testing.T) {
		summe := func(b TimeBudget) int {
			s := 0
			for _, v := range b.Minutes() {
				s += v
			}
			return s
		}
		if !(summe(BudgetNone) < summe(BudgetLow) &&
			summe(BudgetLow) < summe(BudgetMedium) &&
			summe(BudgetMedium) < summe(BudgetHigh)) {
			t.Error("die Stufen steigen nicht durchgehend an")
		}
	})

	t.Run("betreut bedeutet null", func(t *testing.T) {
		for i, v := range BudgetNone.Minutes() {
			if v != 0 {
				t.Fatalf("Tag %d: %d statt 0", i, v)
			}
		}
	})

	t.Run("unbekannte Stufe fällt auf mittel zurück", func(t *testing.T) {
		if TimeBudget("weiß nicht").Minutes() != BudgetMedium.Minutes() {
			t.Error("unbekannte Stufe ergibt nicht die mittlere")
		}
		if TimeBudget("weiß nicht").Known() {
			t.Error("unbekannte Stufe gilt als bekannt")
		}
	})
}
