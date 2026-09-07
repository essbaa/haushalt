package planner

import (
	"sort"
	"time"
)

// candidate ist eine Vorlage, die diese Woche dran ist, zusammen mit den Tagen,
// an denen sie liegen darf. Ein Zwischenschritt, der das Paket nicht verlässt.
type candidate struct {
	tmpl TaskTemplate
	// days sind die möglichen Tage in bevorzugter Reihenfolge.
	days []Date
	// deadline ist der letzte mögliche Tag; Nullwert heißt: keine echte Frist.
	deadline Date
	// pinned bindet den Termin an eine Person (PerPerson-Vorlagen). Leer
	// heißt: die Zuteilung entscheidet.
	pinned string
}

// urgency ist die Sortierschlüsselzahl: je kleiner, desto früher wird
// zugeteilt. Harte Fristen zuerst, dann hohe Kopflast, dann lange Aufgaben —
// große Brocken zuerst zu platzieren füllt die Woche besser aus.
func (c candidate) urgency(week Week) int {
	score := 0
	if c.tmpl.Failure == FailureHard {
		score -= 1000
	}
	if !c.deadline.IsZero() {
		score += week.Monday().DaysUntil(c.deadline) * 10
	} else {
		score += 70
	}
	score -= int(c.tmpl.HeadLoad) * 20
	score -= c.tmpl.DurationMin / 10
	return score
}

// selectDue entscheidet, welche der geltenden Vorlagen in dieser Woche
// auftauchen — und an welchen Tagen sie liegen dürfen.
func selectDue(in Input, templates []TaskTemplate) ([]candidate, []Skipped) {
	var out []candidate
	var skipped []Skipped

	monday := in.Week.Monday()
	sunday := monday.AddDays(6)
	days := in.Week.Days()

	for _, t := range templates {
		last := in.History.lastDone(t.ID)
		var cs []candidate

		switch t.Rhythm.Type {
		case RhythmFixed:
			cs = dueFixed(t, days)
		case RhythmWindow, RhythmTrigger, RhythmPhase:
			cs = dueEvery(t, last, monday, sunday)
		case RhythmSeason:
			cs = dueSeason(t, last, monday, sunday)
		}

		if len(cs) == 0 {
			skipped = append(skipped, Skipped{t.ID, t.Title, SkipNotDue})
			continue
		}
		if t.PerPerson {
			cs = perPerson(cs, eligibleMembers(t, in))
		}
		out = append(out, cs...)
	}

	sort.SliceStable(out, func(i, j int) bool {
		ui, uj := out[i].urgency(in.Week), out[j].urgency(in.Week)
		if ui != uj {
			return ui < uj
		}
		return out[i].tmpl.ID < out[j].tmpl.ID
	})
	return out, skipped
}

// perPerson macht aus jedem Termin eine Aufgabe je Person, fest an sie
// gebunden. Aus einem „Zimmer aufräumen" werden zwei, wenn zwei Kinder im
// Haushalt leben — und keines der beiden räumt das Zimmer des anderen.
//
// Das vervielfacht die Last bewusst: Zwei Kinder bedeuten zwei Aufgaben, nicht
// eine, die zwischen ihnen wandert. In der Bilanz taucht sie bei beiden auf,
// und das ist richtig so.
func perPerson(cs []candidate, members []Member) []candidate {
	if len(members) == 0 {
		return nil
	}
	out := make([]candidate, 0, len(cs)*len(members))
	for _, c := range cs {
		for _, m := range members {
			c.pinned = m.ID
			out = append(out, c)
		}
	}
	return out
}

// dueFixed: feste Wochentage. Jeder passende Tag ergibt eine eigene Aufgabe —
// Abendessen kochen ist an fünf Werktagen fünfmal zu tun, nicht einmal.
func dueFixed(t TaskTemplate, days [7]Date) []candidate {
	var out []candidate
	for _, d := range days {
		for _, wd := range t.Rhythm.Weekdays {
			if d.Weekday() == wd {
				out = append(out, candidate{tmpl: t, days: []Date{d}, deadline: d})
			}
		}
	}
	return out
}

// dueEvery: Fenster, Auslöser und Phase teilen sich dieselbe Rechnung. Der
// nächste Termin ergibt sich aus der letzten Erledigung plus dem gewünschten
// Abstand; wer noch nie erledigt hat, ist ab Montag dran.
//
// Ist der Abstand kürzer als eine Woche, gibt es mehrere Termine: Wäsche alle
// drei Tage ist von Montag bis Sonntag dreimal fällig, nicht einmal.
//
// Jeder Termin bekommt sein eigenes Fenster — vom Termin bis zum Tag vor dem
// nächsten, der letzte bis Sonntag. Diese Trennung ist der eigentliche Punkt:
// Ohne sie sucht die Zuteilung für jeden Termin denselben frühesten freien
// Tag, und drei Ladungen Wäsche landen alle am Montag.
//
// Was liegen geblieben ist, wird nicht nachgeholt. Wer sechs Wochen nicht
// gewaschen hat, bekommt diese Woche drei Termine, nicht vierzehn. Das ist
// eine Produktentscheidung und keine Rechnung: Ein Plan, der Versäumtes
// aufstapelt, wird gelöscht statt abgearbeitet.
func dueEvery(t TaskTemplate, last, monday, sunday Date) []candidate {
	every := t.Rhythm.EveryDays
	if every <= 0 {
		every = 7
	}

	next := monday
	if !last.IsZero() {
		next = last.AddDays(every)
	}
	if next.After(sunday) {
		return nil
	}
	// Überfällig: Der Termin liegt in der Vergangenheit. Die Woche beginnt neu,
	// die verpassten Termine dazwischen entstehen nicht.
	if next.Before(monday) {
		next = monday
	}

	var out []candidate
	for start := next; !start.After(sunday); start = start.AddDays(every) {
		end := start.AddDays(every - 1)
		if end.After(sunday) {
			end = sunday
		}
		c := candidate{tmpl: t, days: daysBetween(start, end)}
		if t.Failure == FailureHard {
			c.deadline = end
		}
		out = append(out, c)
	}
	return out
}

// daysBetween sind alle Tage von from bis to, beide eingeschlossen. Liegt to
// vor from, ist das Ergebnis leer.
func daysBetween(from, to Date) []Date {
	var out []Date
	for d := from; !d.After(to); d = d.AddDays(1) {
		out = append(out, d)
	}
	return out
}

// dueSeason: die Vorlage muss in einem bestimmten Monat erledigt sein und
// taucht LeadDays vorher auf. Der Vorlauf ist hier die eigentliche Leistung —
// im Juli an die Ferienbetreuung zu denken ist das Problem, nicht das
// Anmelden selbst.
func dueSeason(t TaskTemplate, last, monday, sunday Date) []candidate {
	target := nextMonthStart(t.Rhythm.Months, monday)
	if target.IsZero() {
		return nil
	}
	// Innerhalb eines Jahres vor dem Ziel schon erledigt: dann ist gut.
	if !last.IsZero() && last.DaysUntil(target) < 300 {
		return nil
	}
	appearFrom := target.AddDays(-t.LeadDays)
	if appearFrom.After(sunday) {
		return nil
	}
	start := appearFrom
	if start.Before(monday) {
		start = monday
	}
	last3 := target
	if last3.After(sunday) {
		last3 = sunday
	}
	days := daysBetween(start, last3)
	if len(days) == 0 {
		return nil
	}
	return []candidate{{tmpl: t, days: days, deadline: target}}
}

// nextMonthStart ist der erste Tag des nächsten gelisteten Monats ab from.
func nextMonthStart(months []time.Month, from Date) Date {
	best := Date{}
	for _, m := range months {
		for _, year := range []int{from.Year, from.Year + 1} {
			d := Date{Year: year, Month: m, Day: 1}
			if d.Before(from) {
				continue
			}
			if best.IsZero() || d.Before(best) {
				best = d
			}
		}
	}
	return best
}
