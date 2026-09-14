package planner

// TimeBudget ist die grob geschätzte Zeit, die jemand für den Haushalt hat.
//
// Drei Stufen statt sieben Zahlen, und zwar nicht aus Bequemlichkeit: Niemand
// weiß, wie viele Minuten Haushalt er dienstags hat. Eine erfundene Zahl sieht
// nur präziser aus als eine ehrliche Stufe — tragen muss sie dieselbe
// Verteilung.
//
// Die Werte sind Startwerte, keine Wahrheit. Sobald es eine Einstellungsseite
// gibt, darf jede Person ihre Minuten direkt setzen; die Stufen bleiben dann
// das, was das Onboarding daraus macht.
type TimeBudget string

const (
	BudgetNone   TimeBudget = "keine"  // betreute Kinder: erzeugen Arbeit, übernehmen keine
	BudgetLow    TimeBudget = "wenig"  // Schicht, Pendeln, kleines Kind
	BudgetMedium TimeBudget = "mittel" // der Normalfall
	BudgetHigh   TimeBudget = "viel"   // Teilzeit, Elternzeit, Ruhestand
)

// budgets bildet die Stufe auf Minuten je Wochentag ab, Index 0 = Montag.
//
// Werktag und Wochenende unterscheiden sich bewusst deutlich: Der Unterschied
// ist der Grund, warum am Samstag das Bad geputzt wird und am Mittwoch nicht.
var budgets = map[TimeBudget][7]int{
	BudgetNone:   {0, 0, 0, 0, 0, 0, 0},
	BudgetLow:    {30, 30, 30, 30, 30, 90, 60},
	BudgetMedium: {60, 60, 60, 60, 60, 120, 120},
	BudgetHigh:   {120, 120, 120, 120, 120, 180, 150},
}

// Minutes sind die Minuten je Wochentag, Index 0 = Montag.
//
// Eine unbekannte Stufe ergibt die mittlere. Ein Haushalt mit Nullkapazität
// hätte einen leeren Plan, und ein leerer Plan erklärt niemandem, was die App
// tut — das wäre die teuerste Art, streng zu sein.
func (b TimeBudget) Minutes() [7]int {
	if m, ok := budgets[b]; ok {
		return m
	}
	return budgets[BudgetMedium]
}

// Known sagt, ob die Stufe eine der vier bekannten ist. Für die Prüfung an der
// Grenze, wo ein Tippfehler noch eine Fehlermeldung wert ist.
func (b TimeBudget) Known() bool {
	_, ok := budgets[b]
	return ok
}

// BudgetOf ist die Rückrichtung: Welche Stufe gehört zu diesen Minuten?
//
// Leer, wenn keine passt — dann hat jemand die Minuten von Hand gesetzt, und
// das ist ein eigener, gültiger Zustand. „mittel" zurückzugeben, weil es am
// nächsten liegt, wäre eine Behauptung über eine Absicht, die niemand geäußert
// hat.
//
// Die Funktion steht hier und nicht in der Oberfläche. Welche Minuten „mittel"
// bedeutet, weiß genau eine Stelle im System; die Anzeige fragt nach, statt es
// nachzubauen. Eine nachgebaute Zuordnung wäre eine zweite Wahrheit — und die
// zweite ist immer die, die niemand pflegt.
//
// Die Reihenfolge ist fest, damit zwei Aufrufe dasselbe ergeben: Bei einer
// Karte wäre sie es nicht.
func BudgetOf(minuten [7]int) TimeBudget {
	for _, b := range []TimeBudget{BudgetNone, BudgetLow, BudgetMedium, BudgetHigh} {
		if budgets[b] == minuten {
			return b
		}
	}
	return ""
}
