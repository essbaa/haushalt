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
		gilt, fehlt := applies(t.AppliesTo, in.Household)
		switch {
		case in.History.isMuted(t.ID):
			skipped = append(skipped, Skipped{TemplateID: t.ID, Title: t.Title, Code: SkipMuted})
		case gilt == unklar:
			// Nicht einplanen, aber auch nicht vergessen: Daraus wird eine
			// Frage. Der Unterschied zu „gilt nicht" ist der ganze Punkt.
			skipped = append(skipped, Skipped{TemplateID: t.ID, Title: t.Title, Code: SkipUnknown, Fact: fehlt})
		case gilt == giltNicht:
			skipped = append(skipped, Skipped{TemplateID: t.ID, Title: t.Title, Code: SkipNotApplicable})
		case len(eligibleMembers(t, in)) == 0:
			skipped = append(skipped, Skipped{TemplateID: t.ID, Title: t.Title, Code: SkipNoOneEligible})
		default:
			kept = append(kept, t)
		}
	}
	return kept, skipped
}

// geltung ist das Ergebnis der Bedingungsprüfung.
//
// Drei Antworten statt zwei, und das ist die eigentliche Korrektur: Vorher gab
// es nur „gilt" und „gilt nicht", und alles Ungefragte galt als vorhanden.
// Deshalb stand Pflanzen gießen im Plan eines Haushalts ohne Pflanzen.
type geltung int

const (
	gilt geltung = iota
	giltNicht
	unklar
)

// applies prüft die Bedingungen einer Vorlage. Der zweite Rückgabewert nennt
// bei unklar das Faktum, das fehlt.
func applies(c Conditions, h Household) (geltung, string) {
	if !conditionsMet(c, h) {
		return giltNicht, ""
	}
	for _, f := range c.RequiresFacts {
		wert, bekannt := h.Context.Fact(f)
		if !bekannt {
			return unklar, f
		}
		if !wert {
			return giltNicht, ""
		}
	}
	return gilt, ""
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

// TemplateState ist eine Vorlage samt der Frage, ob sie für diesen Haushalt
// gilt — und wenn nicht, warum.
//
// Im Planer und nicht im Speicher: „Vorlage plus Grund" ist Fachsprache. Läge
// der Typ bei der Datenbank, müsste die Bibliothek sie importieren, um
// dieselbe Schnittstelle zu erfüllen — und die Datenbank importiert schon die
// Bibliothek.
type TemplateState struct {
	Template TaskTemplate
	Active   bool
	Reason   SkipCode
	Fact     string
}

// Status sagt, warum eine Vorlage im Plan steht oder nicht — leer heißt: sie
// gilt.
//
// Dieselbe Prüfung wie in selectApplicable, nur für eine einzelne Vorlage und
// ohne Fälligkeit. Die Seite „Eure Woche" zeigt damit die ganze Bibliothek und
// daneben, warum etwas fehlt: nicht zuständig, noch ungeklärt, abgewählt.
//
// Es ist derselbe Code und nicht ein zweiter: Eine Liste, die andere Gründe
// nennt als der Planer, ist schlimmer als gar keine.
func Status(t TaskTemplate, h Household, hist History) (SkipCode, string) {
	if hist.isMuted(t.ID) {
		return SkipMuted, ""
	}
	// Für die Übersicht zählt „braucht einen Anlass" als eigener Zustand. Im
	// Planer entscheidet dagegen die Fälligkeit: Gibt es einen passenden
	// Anlass, entsteht die Aufgabe ganz normal.
	if t.AppliesTo.RequiresEvent && len(h.Occasions) == 0 {
		return SkipNeedsEvent, ""
	}
	gilt, fehlt := applies(t.AppliesTo, h)
	switch gilt {
	case unklar:
		return SkipUnknown, fehlt
	case giltNicht:
		return SkipNotApplicable, ""
	}
	if len(eligibleMembers(t, Input{Household: h, History: hist})) == 0 {
		return SkipNoOneEligible, ""
	}
	return "", ""
}
