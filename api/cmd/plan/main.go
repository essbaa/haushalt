// Kommando plan berechnet einen Wochenplan und schreibt ihn auf die Konsole.
//
//	go run ./cmd/plan --haushalt library/beispiele/familie-a.json --woche 2026-W38
//
// Der schnellste Weg, den Planer anzusehen, ohne App, Datenbank oder Anmeldung.
// Auch der Aufruf, der ins README gehört: Er zeigt den Kern in fünf Zeilen.
package main

import (
	"flag"
	"fmt"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/zakaria/haushalt/api/internal/library"
	"github.com/zakaria/haushalt/api/internal/planner"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, "fehler:", err)
		os.Exit(1)
	}
}

func run() error {
	var (
		householdPath = flag.String("haushalt", "library/beispiele/familie-a.json", "Haushaltsdatei")
		templatePath  = flag.String("bibliothek", "library/vorlagen.json", "Vorlagen-Bibliothek")
		weekFlag      = flag.String("woche", "", "Kalenderwoche, z. B. 2026-W38 (Vorgabe: laufende Woche)")
		showSkipped   = flag.Bool("uebersprungen", false, "auch zeigen, was nicht im Plan steht")
	)
	flag.Parse()

	household, history, err := library.LoadHousehold(*householdPath)
	if err != nil {
		return err
	}
	templates, err := library.LoadTemplates(*templatePath)
	if err != nil {
		return err
	}

	week := planner.WeekOf(planner.DateOf(time.Now()))
	if *weekFlag != "" {
		if week, err = planner.ParseWeek(*weekFlag); err != nil {
			return err
		}
	}

	result, err := planner.Plan(planner.Input{
		Household: household,
		Templates: templates,
		Week:      week,
		History:   history,
		Limits:    planner.DefaultLimits(),
	})
	if err != nil {
		return err
	}

	print(household, result, *showSkipped)
	return nil
}

func print(h planner.Household, r planner.Result, withSkipped bool) {
	monday := r.Week.Monday()
	fmt.Printf("\n%s · Kalenderwoche %s (%s bis %s)\n\n",
		h.Name, r.Week, ddmm(monday), ddmm(monday.AddDays(6)))

	names := map[string]string{}
	for _, m := range h.Members {
		names[m.ID] = firstWord(m.Name, m.ID)
	}

	byDay := map[planner.Date][]planner.PlannedTask{}
	for _, t := range r.Tasks {
		byDay[t.Day] = append(byDay[t.Day], t)
	}

	for _, day := range r.Week.Days() {
		tasks := byDay[day]
		fmt.Printf("%-10s %s\n", weekdayName(day), ddmm(day))
		if len(tasks) == 0 {
			fmt.Printf("  %s\n", dim("nichts geplant"))
		}
		for _, t := range tasks {
			mark := " "
			if t.Kind == planner.KindOrg {
				mark = "*"
			}
			fmt.Printf("  %s %-9s %-38s %3d min  Kopflast %d  %s\n",
				mark, names[t.AssigneeID], t.Title, t.DurationMin, t.HeadLoad, reason(t, names))
		}
		fmt.Println()
	}

	fmt.Println("Bilanz")
	for _, l := range r.Balance {
		fmt.Printf("  %-9s %3d min · Kopflast %2d · gewichtet %3d · Zeit %3d %% · %d Aufgaben\n",
			names[l.MemberID], l.Minutes, l.HeadLoad, l.Weighted, l.Utilization(), l.Tasks)
	}
	fmt.Printf("\n  %s\n", dim("* = Organisationsaufgabe. Gewichtet = Minuten + Kopflast × 15. Zeit = verplante Minuten im Verhältnis zur verfügbaren Zeit."))
	fmt.Printf("  %s\n", dim("Verteilt wird nach gewichteter Last im Verhältnis zur Kapazität — wer weniger Zeit hat, trägt weniger."))

	if withSkipped && len(r.Skipped) > 0 {
		fmt.Println("\nNicht im Plan")
		byCode := map[planner.SkipCode][]string{}
		var codes []planner.SkipCode
		for _, s := range r.Skipped {
			if _, ok := byCode[s.Code]; !ok {
				codes = append(codes, s.Code)
			}
			byCode[s.Code] = append(byCode[s.Code], s.Title)
		}
		sort.Slice(codes, func(i, j int) bool { return codes[i] < codes[j] })
		for _, c := range codes {
			fmt.Printf("  %-16s %s\n", c, strings.Join(byCode[c], ", "))
		}
	}
	fmt.Println()
}

func reason(t planner.PlannedTask, names map[string]string) string {
	switch t.Reason.Code {
	case planner.ReasonRotation:
		return dim("zuletzt bei " + names[t.Reason.Previous])
	case planner.ReasonBalance:
		return dim("geringste Last")
	case planner.ReasonOwn:
		return dim("eigene Aufgabe")
	case planner.ReasonFixed:
		return dim("feste Person")
	case planner.ReasonOnlyOne:
		return dim("einzige Möglichkeit")
	}
	return ""
}

func weekdayName(d planner.Date) string {
	return [...]string{"Montag", "Dienstag", "Mittwoch", "Donnerstag", "Freitag", "Samstag", "Sonntag"}[(int(d.Weekday())+6)%7]
}

func ddmm(d planner.Date) string { return fmt.Sprintf("%02d.%02d.", d.Day, int(d.Month)) }

func dim(s string) string { return "\x1b[2m" + s + "\x1b[0m" }

func firstWord(name, fallback string) string {
	if name == "" {
		return fallback
	}
	if i := strings.IndexByte(name, ' '); i > 0 {
		return name[:i]
	}
	return name
}
