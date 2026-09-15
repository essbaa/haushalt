package planner

import "fmt"

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
		case t.NeedsAgreement && !in.History.hasAgreement(t.ID):
			// Ohne Absprache nicht geplant — und nicht verschwiegen: Daraus
			// wird die Aufforderung, das Raster zu füllen.
			skipped = append(skipped, Skipped{TemplateID: t.ID, Title: t.Title, Code: SkipNeedsAgreement})
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
	if erfuellt, _ := conditionsMet(c, h); !erfuellt {
		return giltNicht, ""
	}
	for _, f := range c.RequiresFacts {
		wert, bekannt := h.Context.Fact(f)
		if !bekannt {
			return unklar, f
		}
		if !wert {
			// Das Faktum wird auch hier zurückgegeben, nicht nur bei
			// „unbekannt". Ein Nein, das man nicht zurücknehmen kann, ist
			// dasselbe wie ein Rateschluss, den man nicht korrigieren kann
			// (ADR-0007) — eine Behauptung. Die Oberfläche braucht das
			// Faktum, um die Frage noch einmal stellen zu können.
			return giltNicht, f
		}
	}
	return gilt, ""
}

// conditionsMet prüft die Bedingungen am Haushalt — und sagt, welche fehlt.
//
// Der zweite Rückgabewert ist ein Satz für Menschen, und er ist der Grund,
// warum diese Funktion nicht mehr nur ein bool liefert. Die Aufgabenliste sagte
// vorher „gilt bei euch nicht" und dazu eine Aufzählung aller denkbaren
// Ursachen — Garten, Auto, Haustiere, Betreuung. Wer „Brotdose vorbereiten"
// freischalten wollte, las eine Liste und wusste hinterher genauso wenig.
//
// Der Planer weiß es genau. Es nicht herauszugeben war dieselbe Andeutung
// statt Auskunft wie die ausgegrauten Zeilen, die dieses Projekt schon einmal
// beseitigt hat.
//
// Die Reihenfolge entscheidet, welche von mehreren fehlenden Bedingungen
// genannt wird. Das ist in Ordnung: Wer die erste erfüllt, bekommt die
// nächste zu sehen.
func conditionsMet(c Conditions, h Household) (bool, string) {
	if c.RequiresCar && !h.Context.HasCar {
		return false, "Braucht ein Auto."
	}
	if c.RequiresYard && !h.Context.HasYard {
		return false, "Braucht einen Garten."
	}
	if c.RequiresPet && len(h.Context.Pets) == 0 {
		return false, "Braucht ein Haustier."
	}
	if c.RequiresPetKind != "" && !hasPet(h, c.RequiresPetKind) {
		return false, fmt.Sprintf("Braucht %s im Haushalt.", tierart(c.RequiresPetKind))
	}
	if c.RequiresHome != "" && h.Context.Home != c.RequiresHome {
		return false, fmt.Sprintf("Gilt nur in %s.", wohnform(c.RequiresHome))
	}
	if c.RequiresChildAged != nil && !hasChildAged(h, *c.RequiresChildAged) {
		return false, "Braucht ein Kind " + altersspanne(*c.RequiresChildAged) + "."
	}
	if c.RequiresCare != "" && !hasChildInCare(h, c.RequiresCare) {
		return false, betreuungssatz(c.RequiresCare)
	}
	return true, ""
}

func tierart(art string) string {
	switch art {
	case "hund":
		return "einen Hund"
	case "katze":
		return "eine Katze"
	}
	return "ein Haustier der Art „" + art + "\u201c"
}

func wohnform(h Home) string {
	if h == HomeHouse {
		return "einem Haus"
	}
	return "einer Wohnung"
}

func altersspanne(a AgeRange) string {
	switch {
	case a.Max == 0:
		return fmt.Sprintf("ab %d Jahren", a.Min)
	case a.Min == 0:
		return fmt.Sprintf("bis %d Jahre", a.Max)
	default:
		return fmt.Sprintf("zwischen %d und %d Jahren", a.Min, a.Max)
	}
}

func betreuungssatz(c Care) string {
	switch c {
	case CareKita:
		return "Braucht ein Kind, das in die Kita geht."
	case CareSchool:
		return "Braucht ein Kind, das zur Schule geht."
	}
	return "Braucht ein betreutes Kind."
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
	// Agreement ist das Wochenraster dieser Vorlage, Index 0 = Montag.
	// Leere Plätze heißen: für diesen Tag ist nichts abgesprochen.
	Agreement [7]string

	// Need ist die fehlende Bedingung als Satz für Menschen: „Braucht ein
	// Kind, das zur Schule geht." Leer, wenn keine fehlt — oder wenn ein
	// verneintes Faktum die Ursache ist; dann steht in Fact, was zu fragen
	// wäre.
	Need string
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
// Der dritte Rückgabewert ist die fehlende Bedingung als Satz für Menschen —
// leer, wenn es keine ist oder wenn ein verneintes Faktum die Ursache war.
func Status(t TaskTemplate, h Household, hist History) (SkipCode, string, string) {
	if hist.isMuted(t.ID) {
		return SkipMuted, "", ""
	}
	// Für die Übersicht zählt „braucht einen Anlass" als eigener Zustand. Im
	// Planer entscheidet dagegen die Fälligkeit: Gibt es einen passenden
	// Anlass, entsteht die Aufgabe ganz normal.
	if t.AppliesTo.RequiresEvent && len(h.Occasions) == 0 {
		return SkipNeedsEvent, "", ""
	}
	// Eine Aufgabe, die abgesprochen gehört, wird ohne Absprache nicht
	// geraten. Der Planer kennt die Kapazität und nicht den Arbeitsplan —
	// er würde mit voller Überzeugung den Falschen einteilen.
	if t.NeedsAgreement && !hist.hasAgreement(t.ID) {
		return SkipNeedsAgreement, "", ""
	}
	if erfuellt, warum := conditionsMet(t.AppliesTo, h); !erfuellt {
		return SkipNotApplicable, "", warum
	}
	gilt, fehlt := applies(t.AppliesTo, h)
	switch gilt {
	case unklar:
		return SkipUnknown, fehlt, ""
	case giltNicht:
		// Kommt das Nein aus einem beantworteten Faktum, steht es hier und die
		// App kann einen Weg zurück anbieten. Kommt es aus dem Kontext (kein
		// Garten, kein Auto), ist es leer — dann führt der Weg über die
		// Einstellungen, und die Oberfläche sagt das auch so.
		return SkipNotApplicable, fehlt, ""
	}
	if len(eligibleMembers(t, Input{Household: h, History: hist})) == 0 {
		return SkipNoOneEligible, "", ""
	}
	return "", "", ""
}
