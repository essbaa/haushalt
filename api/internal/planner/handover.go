package planner

import "sort"

// Handover wählt die Person, die eine abgegebene Aufgabe übernimmt.
//
// Der Produktgrund steht im Konzept: „Ausführende brauchen Handlungsmacht,
// nicht nur Pflichten. Eine reine Empfangsliste wird gelöscht." Abgeben ist
// die kleinste Form davon — und es darf nicht bedeuten, dass die Aufgabe
// liegen bleibt, sonst ist es Löschen mit besserem Gewissen.
//
// Gewählt wird nach denselben Regeln wie beim Zuteilen, in dieser Reihenfolge:
//
//  1. geeignet — Alter, Rolle und Verteilungsregel der Vorlage,
//  2. hat dieselbe Vorlage diese Woche noch nicht (Rotation innerhalb der
//     Woche, dieselbe Regel, die der Ausgleich verletzt hatte),
//  3. am wenigsten ausgelastet im Verhältnis zur eigenen Kapazität,
//  4. bei Gleichstand die kleinere Kennung — damit zwei Aufrufe dasselbe
//     ergeben.
//
// Der zweite Rückgabewert ist die Begründung für die neue Zuteilung, der
// dritte sagt, ob überhaupt jemand gefunden wurde. Niemand gefunden ist kein
// Fehler: Dann steht die Aufgabe ohne Zuständige da, und das ist ein gültiger
// Zustand — sichtbar, offen, von jemandem zu holen.
func Handover(in Input, tasks []PlannedTask, taskID string) (string, Reason, bool) {
	var aufgabe PlannedTask
	gefunden := false
	for _, t := range tasks {
		if t.ID == taskID {
			aufgabe, gefunden = t, true
			break
		}
	}
	if !gefunden {
		return "", Reason{}, false
	}

	var vorlage TaskTemplate
	for _, v := range in.Templates {
		if v.ID == aufgabe.TemplateID {
			vorlage = v
			break
		}
	}
	if vorlage.ID == "" {
		return "", Reason{}, false
	}

	// Eine Aufgabe, die einer Person selbst gehört, gibt man nicht ab: „Dein
	// Bett beziehen" bei jemand anderem ist eine andere Aufgabe, nicht
	// dieselbe in anderen Händen (ADR-0005).
	if vorlage.PerPerson || aufgabe.Reason.Code == ReasonOwn {
		return "", Reason{}, false
	}

	// Abgesprochenes gibt der Planer nicht weiter. Wer einspringen kann, hängt
	// am Arbeitsplan und nicht an freien Minuten (ADR-0016) — die Aufgabe
	// bliebe formal besetzt und wäre in Wahrheit unbesetzt.
	//
	// Sie steht danach offen da, und das ist hier die ehrliche Antwort: eine
	// sichtbare Lücke, die ein Mensch schließt. „Wer macht das?" verteilt sie
	// gezielt (ADR-0013), und für den einen Tag reicht das.
	if vorlage.NeedsAgreement {
		return "", Reason{}, false
	}

	last := BalanceOf(in.Household, tasks, in.Limits)
	auslastung := make(map[string]int, len(last))
	for _, b := range last {
		if b.Capacity > 0 {
			auslastung[b.MemberID] = b.Weighted * 1000 / b.Capacity
		}
	}

	type kandidat struct {
		id      string
		belegt  bool // hält diese Vorlage schon in dieser Woche
		last    int
		reihung string
	}
	var kandidaten []kandidat
	for _, m := range eligibleMembers(vorlage, in) {
		if m.ID == aufgabe.AssigneeID {
			continue
		}
		kandidaten = append(kandidaten, kandidat{
			id:      m.ID,
			belegt:  holdsOther(tasks, aufgabe.TemplateID, m.ID, -1),
			last:    auslastung[m.ID],
			reihung: m.ID,
		})
	}
	if len(kandidaten) == 0 {
		return "", Reason{}, false
	}

	sort.Slice(kandidaten, func(i, j int) bool {
		a, b := kandidaten[i], kandidaten[j]
		if a.belegt != b.belegt {
			return !a.belegt
		}
		if a.last != b.last {
			return a.last < b.last
		}
		return a.reihung < b.reihung
	})

	return kandidaten[0].id, Reason{Code: ReasonBalance, Previous: aufgabe.AssigneeID}, true
}
