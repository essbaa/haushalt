package planner

// selectApplicable filtert die Bibliothek auf die Vorlagen, die für diesen
// Haushalt überhaupt gelten. Kein Auto, keine Reifenwechsel-Aufgabe.
//
// Vorlagen, die der Haushalt wiederholt gelöscht hat, fallen hier ebenfalls
// heraus — das ist die einfachste Form der Lernschleife und die einzige, die
// im ersten Durchstich schon nötig ist.
func selectApplicable(in Input) ([]TaskTemplate, []Skipped) {
	var kept []TaskTemplate
	var skipped []Skipped

	for _, t := range in.Templates {
		switch {
		case in.History.isMuted(t.ID):
			skipped = append(skipped, Skipped{t.ID, t.Title, SkipMuted})
		case !conditionsMet(t.AppliesTo, in.Household):
			skipped = append(skipped, Skipped{t.ID, t.Title, SkipNotApplicable})
		case len(eligibleMembers(t, in)) == 0:
			skipped = append(skipped, Skipped{t.ID, t.Title, SkipNoOneEligible})
		default:
			kept = append(kept, t)
		}
	}
	return kept, skipped
}

func conditionsMet(c Conditions, h Household) bool {
	if c.RequiresCar && !h.Context.HasCar {
		return false
	}
	if c.RequiresYard && !h.Context.HasYard {
		return false
	}
	if c.RequiresPet && len(h.Context.Pets) == 0 {
		return false
	}
	if c.RequiresPetKind != "" && !hasPet(h, c.RequiresPetKind) {
		return false
	}
	if c.RequiresHome != "" && h.Context.Home != c.RequiresHome {
		return false
	}
	if c.RequiresChildAged != nil && !hasChildAged(h, *c.RequiresChildAged) {
		return false
	}
	if c.RequiresCare != "" && !hasChildInCare(h, c.RequiresCare) {
		return false
	}
	return true
}

func hasPet(h Household, kind string) bool {
	for _, p := range h.Context.Pets {
		if p == kind {
			return true
		}
	}
	return false
}

func hasChildAged(h Household, r AgeRange) bool {
	for _, m := range h.Members {
		if !m.IsAdult() && r.Contains(m.Age) {
			return true
		}
	}
	return false
}

func hasChildInCare(h Household, care Care) bool {
	for _, m := range h.Members {
		if !m.IsAdult() && m.Care == care {
			return true
		}
	}
	return false
}

// eligibleMembers sind die Personen, die diese Aufgabe übernehmen dürfen.
// Die Reihenfolge folgt der Reihenfolge im Haushalt und ist damit stabil.
func eligibleMembers(t TaskTemplate, in Input) []Member {
	if t.Distribution == DistFixed {
		if id := in.History.fixedTo(t.ID); id != "" {
			for _, m := range in.Household.Members {
				if m.ID == id && m.CanPerform() {
					return []Member{m}
				}
			}
		}
		// Ohne hinterlegte Person verhält sich die Vorlage wie eine rotierende.
	}

	var out []Member
	for _, m := range in.Household.Members {
		if !m.CanPerform() {
			continue
		}
		if t.Distribution == DistAdultsOnly && !m.IsAdult() {
			continue
		}
		if t.Distribution == DistChildrenOnly && m.IsAdult() {
			continue
		}
		if !m.IsAdult() && m.Age < t.MinAge {
			continue
		}
		out = append(out, m)
	}
	return out
}
