package planner

import "sort"

// state ist die laufende Bilanz einer Person während der Zuteilung.
type state struct {
	member    Member
	remaining [7]int // freie Minuten je Wochentag, Index 0 = Montag
	count     [7]int // bereits zugeteilte Aufgaben je Wochentag
	weighted  int    // Minuten + Kopflast × HeadLoadMinutes
	capacity  int    // verfügbare Minuten der ganzen Woche
}

// utilization ist die Auslastung in Promille: Last im Verhältnis zu dem, was
// diese Person überhaupt an Zeit hat.
//
// Verglichen wird die Auslastung und nicht die absolute Last. Sonst bekommt
// eine Dreizehnjährige mit 300 verfügbaren Minuten genauso viel aufgeladen wie
// ihre Mutter mit 550 — und ist damit ausgebucht, während die Mutter die halbe
// Woche frei hat.
func (s *state) utilization() int {
	if s.capacity <= 0 {
		return 1 << 30
	}
	return s.weighted * 1000 / s.capacity
}

// assign verteilt die fälligen Aufgaben auf Personen und Tage.
//
// Zwei Regeln greifen ineinander, und ihre Reihenfolge ist die eigentliche
// Aussage des Planers:
//
//	Rotation vor Ausgleich.  Wer eine Aufgabe zuletzt hatte, bekommt sie nicht
//	                         wieder, solange jemand anders sie übernehmen kann.
//	                         Das gilt vor allem für Kopfarbeit — die Rotation
//	                         der Organisationsaufgaben ist die eigentliche
//	                         Entlastung, die Rotation des Staubsaugens ist
//	                         Kosmetik.
//
//	Ausgleich nach Gewicht.  Verglichen wird nicht die Zeit, sondern
//	                         Minuten + Kopflast × HeadLoadMinutes. Ein Planer,
//	                         der nur Minuten zählt, gibt der Person, die alle
//	                         Termine koordiniert, obendrein das Bad.
//
// Der Tag ergibt sich danach: der früheste erlaubte Tag, an dem diese Person
// noch Luft hat. Früh statt spät, damit eine Verschiebung noch Platz findet.
func assign(in Input, cands []candidate) ([]PlannedTask, []Skipped) {
	states := newStates(in.Household)
	// recent hält fest, wer eine Vorlage innerhalb dieser Woche zuletzt bekommen
	// hat. Ohne das würde eine Vorlage, die mehrfach in der Woche fällig ist —
	// Abendessen kochen an drei Abenden — dreimal derselben Person zufallen:
	// Die Historie sagt ja jedes Mal dasselbe.
	recent := map[string]string{}

	var tasks []PlannedTask
	var skipped []Skipped

	for _, c := range cands {
		eligible := eligibleMembers(c.tmpl, in)
		// Gehört der Termin einer bestimmten Person, gibt es nichts zu wählen.
		if c.pinned != "" {
			eligible = onlyMember(eligible, c.pinned)
		}
		if len(eligible) == 0 {
			skipped = append(skipped, Skipped{c.tmpl.ID, c.tmpl.Title, SkipNoOneEligible})
			continue
		}

		previous := in.History.lastAssignee(c.tmpl.ID)
		if who, ok := recent[c.tmpl.ID]; ok {
			previous = who
		}
		order := rankMembers(eligible, states, previous)

		placed := false
		for _, m := range order {
			st := states[m.ID]
			day, ok := earliestFreeDay(st, c, in.Limits)
			if !ok {
				continue
			}
			task := PlannedTask{
				TemplateID:  c.tmpl.ID,
				Title:       c.tmpl.Title,
				Category:    c.tmpl.Category,
				Kind:        c.tmpl.Kind,
				Day:         day,
				Slot:        slotOrDefault(c.tmpl.Slot),
				DurationMin: c.tmpl.DurationMin,
				HeadLoad:    c.tmpl.HeadLoad,
				AssigneeID:  m.ID,
				Failure:     c.tmpl.Failure,
				Deadline:    c.deadline,
				Reason:      reasonFor(c.tmpl, in, m, eligible, previous),
			}
			if c.pinned != "" {
				task.Reason = Reason{Code: ReasonOwn}
			}
			charge(st, day, task, in.Limits)
			// Personengebundene Termine nehmen nicht an der Rotation teil.
			if c.pinned == "" {
				recent[c.tmpl.ID] = m.ID
			}
			tasks = append(tasks, task)
			placed = true
			break
		}
		if !placed {
			skipped = append(skipped, Skipped{c.tmpl.ID, c.tmpl.Title, SkipNoCapacity})
		}
	}
	return tasks, skipped
}

func newStates(h Household) map[string]*state {
	out := make(map[string]*state, len(h.Members))
	for _, m := range h.Members {
		if !m.CanPerform() {
			continue
		}
		out[m.ID] = &state{member: m, remaining: m.CapacityMinutes, capacity: weeklyCapacity(m)}
	}
	return out
}

// rankMembers bringt die in Frage kommenden Personen in die Reihenfolge, in
// der sie gefragt werden: erst die, die die Aufgabe zuletzt nicht hatten, dann
// die mit der geringsten Auslastung, zuletzt nach ID — damit das Ergebnis bei
// Gleichstand reproduzierbar bleibt.
func rankMembers(eligible []Member, states map[string]*state, previous string) []Member {
	out := make([]Member, len(eligible))
	copy(out, eligible)
	sort.SliceStable(out, func(i, j int) bool {
		a, b := out[i], out[j]
		if (a.ID == previous) != (b.ID == previous) {
			return b.ID == previous
		}
		la, lb := loadOf(states, a.ID), loadOf(states, b.ID)
		if la != lb {
			return la < lb
		}
		return a.ID < b.ID
	})
	return out
}

// onlyMember reduziert die Liste auf genau eine Person, falls sie dabei ist.
func onlyMember(members []Member, id string) []Member {
	for _, m := range members {
		if m.ID == id {
			return []Member{m}
		}
	}
	return nil
}

func loadOf(states map[string]*state, id string) int {
	if st, ok := states[id]; ok {
		return st.utilization()
	}
	return 1 << 30
}

// earliestFreeDay sucht den ersten der erlaubten Tage, an dem die Person noch
// genug Minuten und noch nicht zu viele Aufgaben hat.
func earliestFreeDay(st *state, c candidate, l Limits) (Date, bool) {
	for _, d := range c.days {
		i := weekdayIndex(d)
		if st.count[i] >= l.MaxTasksPerMemberDay {
			continue
		}
		if st.remaining[i] < c.tmpl.DurationMin {
			continue
		}
		return d, true
	}
	return Date{}, false
}

func charge(st *state, d Date, t PlannedTask, l Limits) {
	i := weekdayIndex(d)
	st.remaining[i] -= t.DurationMin
	st.count[i]++
	st.weighted += t.Weight(l)
}

func reasonFor(t TaskTemplate, in Input, chosen Member, eligible []Member, previous string) Reason {
	if t.Distribution == DistFixed && in.History.fixedTo(t.ID) == chosen.ID {
		return Reason{Code: ReasonFixed}
	}
	if len(eligible) == 1 {
		return Reason{Code: ReasonOnlyOne}
	}
	if previous != "" && previous != chosen.ID {
		return Reason{Code: ReasonRotation, Previous: previous}
	}
	return Reason{Code: ReasonBalance}
}

// weekdayIndex bildet Montag auf 0 ab. Die Standardbibliothek beginnt bei
// Sonntag, der Wochenplan nicht.
func weekdayIndex(d Date) int {
	return (int(d.Weekday()) + 6) % 7
}

func slotOrDefault(s Slot) Slot {
	if s == "" {
		return SlotAny
	}
	return s
}
