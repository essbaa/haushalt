package planner

import (
	"fmt"
	"sort"
)

// Occasion ist ein Anlass im Haushalt: ein Geburtstag, ein Elternabend, ein
// Arzttermin.
//
// Der Grund für diesen Typ steht in ADR-0010: „Geschenk für Kindergeburtstag
// besorgen" trug den Rhythmus „Auslöser, alle 45 Tage" und erfand damit alle
// anderthalb Monate einen Geburtstag. Eine Aufgabe ohne Anlass ist keine
// Erinnerung, sondern eine Behauptung. Hier ist der Anlass.
type Occasion struct {
	ID    string
	Title string
	Date  Date

	// Kind verbindet den Anlass mit den Vorlagen, die daran hängen —
	// "kindergeburtstag", "elternabend", "arzttermin".
	Kind string

	// Yearly gilt für Geburtstage: derselbe Tag, jedes Jahr. Ohne das müsste
	// man jeden Geburtstag jährlich neu eintragen, und genau das würde
	// niemand tun.
	Yearly bool
}

// occurrenceIn liefert das Datum dieses Anlasses im gegebenen Jahr — bei
// einmaligen Terminen nur, wenn sie in dieses Jahr fallen.
//
// Der 29. Februar wird in Nicht-Schaltjahren zum 1. März. Das ist eine
// Festlegung und keine Wahrheit; sie steht hier, damit sie nicht an drei
// Stellen verschieden getroffen wird.
func (o Occasion) occurrenceIn(jahr int) (Date, bool) {
	if !o.Yearly {
		if o.Date.Year != jahr {
			return Date{}, false
		}
		return o.Date, true
	}
	d := Date{Year: jahr, Month: o.Date.Month, Day: o.Date.Day}
	if d.time().Day() != o.Date.Day {
		return Date{Year: jahr, Month: 3, Day: 1}, true
	}
	return d, true
}

// dueFromOccasions findet die Fälligkeiten einer anlassgebundenen Vorlage.
//
// Fällig ist sie im Fenster [Anlass − Vorlauf, Anlass]. Schneidet dieses
// Fenster die Woche, entsteht ein Kandidat mit dem Anlass als Frist — ein
// Geschenk nach dem Geburtstag ist kein Geschenk.
//
// Ohne passenden Anlass entsteht nichts. Das ist der ganze Punkt.
func dueFromOccasions(t TaskTemplate, anlaesse []Occasion, monday, sunday Date) []candidate {
	vorlauf := t.LeadDays
	if vorlauf <= 0 {
		vorlauf = 7
	}

	var out []candidate
	for _, o := range anlaesse {
		if o.Kind != t.AppliesTo.RequiresOccasion {
			continue
		}
		// Zwei Jahre prüfen: Ein Fenster über den Jahreswechsel gehört zum
		// Geburtstag im Januar, liegt aber im Dezember.
		for _, jahr := range []int{monday.Year, sunday.Year, sunday.Year + 1} {
			tag, ok := o.occurrenceIn(jahr)
			if !ok {
				continue
			}
			von := tag.AddDays(-vorlauf)
			if tag.Before(monday) || von.After(sunday) {
				continue
			}
			start, ende := von, tag
			if start.Before(monday) {
				start = monday
			}
			if ende.After(sunday) {
				ende = sunday
			}
			out = append(out, candidate{
				tmpl:     t,
				days:     daysBetween(start, ende),
				deadline: tag,
			})
		}
	}

	sort.SliceStable(out, func(i, j int) bool {
		return out[i].deadline.Before(out[j].deadline)
	})
	return dedupe(out)
}

// dedupe entfernt Kandidaten, die durch die Jahresschleife doppelt entstanden
// sind — derselbe Anlass, dasselbe Fenster.
func dedupe(cs []candidate) []candidate {
	gesehen := map[string]bool{}
	out := cs[:0]
	for _, c := range cs {
		schluessel := c.tmpl.ID + "|" + c.deadline.String()
		if gesehen[schluessel] {
			continue
		}
		gesehen[schluessel] = true
		out = append(out, c)
	}
	return out
}

// Kinds sind die Arten von Anlässen, die Vorlagen kennen.
//
// Eine feste Liste und kein freier Text: Ein Anlass, dessen Art zu keiner
// Vorlage passt, erzeugt nie eine Aufgabe — der Nutzer hätte etwas eingetragen
// und nie erfahren, warum nichts passiert.
var Kinds = []string{"kindergeburtstag", "geburtstag", "elternabend", "arzttermin", "sonstiges"}

// Validate prüft einen Anlass so, wie ein Mensch ihn prüfen würde.
func (o Occasion) Validate() error {
	if err := ValidName("der Anlass", o.Title, 80); err != nil {
		return err
	}
	if o.Date.IsZero() {
		return fmt.Errorf("%w: der Anlass braucht ein Datum", ErrInvalidSetup)
	}
	for _, k := range Kinds {
		if o.Kind == k {
			return nil
		}
	}
	return fmt.Errorf("%w: %q ist keine Art von Anlass", ErrInvalidSetup, o.Kind)
}
