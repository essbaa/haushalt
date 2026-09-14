package planner

// rebalance ist der zweite Durchgang über den fertigen Plan: Er tauscht die
// zuständigen Personen zwischen zwei Aufgaben, solange sich die Verteilung
// dadurch verbessert.
//
// Warum es ihn braucht: assign entscheidet gierig — jede Aufgabe geht an die
// Person, die in genau diesem Moment am wenigsten trägt, und was einmal
// zugeteilt ist, wird nie wieder angefasst. Drei Aufgaben genügen, damit das
// schiefgeht: Zwei kleine Organisationsaufgaben landen bei beiden Personen,
// die große Putzaufgabe danach bei der, die zufällig zuerst dran war. 165 zu
// 55, wo 110 zu 110 möglich wäre.
//
// Fünf Regeln bleiben unantastbar; sie stehen ausformuliert in swapAllowed.
// Die wichtigste davon: Rotation geht vor Ausgleich — auch vor diesem
// Durchgang. Eine Aufgabe wandert nie zu der Person zurück, die sie zuletzt
// hatte, und niemand sammelt zwei Termine derselben Vorlage ein, selbst wenn
// die Zahlen dadurch schöner würden.
//
// Die Tage bleiben, wie der erste Durchgang sie gesetzt hat. Getauscht werden
// nur die Personen.
//
// Zwei Durchgänge, und der Unterschied ist die Rotation über die Wochen.
//
// Sie war hier lange ein Veto: Eine Aufgabe durfte nie zu der Person zurück,
// die sie zuletzt hatte, egal wie schief die Woche dadurch blieb. Im assign
// ist dieselbe Regel nur eine Vorliebe — rankMembers sortiert die Person von
// letzter Woche nach hinten und nimmt sie trotzdem, wenn sonst niemand kann.
// Dieselbe Regel in zwei Stärken, und die härtere stand ausgerechnet in dem
// Durchgang, der ausgleichen soll.
//
// Bezahlt hat das die Person mit der kleinsten Kapazität. Gemessen an Familie
// B: Die dreizehnjährige Mia lag bei 79 Prozent ihrer verfügbaren Zeit, ihr
// Vater bei 67 — weil der Tausch "Böden wischen gegen Wäsche zusammenlegen"
// die Kennzahl um ein Viertel verbessert hätte und daran scheiterte, dass
// Tarek die Böden letzte Woche gewischt hatte. Mia wischte die Böden, weil
// ihr Vater sie zuletzt gewischt hatte.
//
// Also wie überall sonst im Planer: erst alle Tausche mit Rotation, dann,
// wenn immer noch jemand deutlich über seinem Anteil liegt, auch die ohne.
// Eine Vorliebe, keine Bedingung — dasselbe Muster wie earliestFreeDay und
// nichtZuLangFuerKinder.
//
// Siehe docs/adr/0003-ausgleich-als-zweiter-durchgang.md und
// docs/adr/0004-fairness-nach-kapazitaet.md.
func rebalance(in Input, tasks []PlannedTask) []PlannedTask {
	if len(tasks) < 2 {
		return tasks
	}
	tasks = ausgleichen(in, tasks, true)
	if in.Limits.MaxOvershootPermille > 0 {
		tasks = ausgleichen(in, tasks, false)
	}
	return tasks
}

// ausgleichen ist ein Durchgang über den Plan. strengeRotation sagt, ob die
// Rotation über die Wochen dabei unantastbar ist.
func ausgleichen(in Input, tasks []PlannedTask, strengeRotation bool) []PlannedTask {

	templates := make(map[string]TaskTemplate, len(in.Templates))
	for _, t := range in.Templates {
		templates[t.ID] = t
	}

	current := imbalanceOf(in, tasks)

	// Nur echte Verbesserungen werden angenommen, und die Kennzahl ist eine
	// ganze Zahl nach unten begrenzt — der Durchgang endet also von selbst.
	// maxPasses ist trotzdem da: eine Schleife ohne Obergrenze im Kern eines
	// Dienstes ist eine Wette, die man nicht eingehen muss.
	const maxPasses = 8

	for pass := 0; pass < maxPasses; pass++ {
		// Der zweite Durchgang bricht die Rotation nur, solange es etwas
		// bringt: Sobald niemand mehr deutlich über seinem Anteil liegt,
		// hört er auf. Ohne diese Schranke würde er eine Wiederholung auch
		// für eine Verbesserung im Promillebereich in Kauf nehmen — und aus
		// "Rotation vor Ausgleich" wäre unbemerkt das Gegenteil geworden.
		if !strengeRotation && ueberschussPromille(in, tasks) <= in.Limits.MaxOvershootPermille {
			break
		}
		improved := false
		for i := range tasks {
			for j := i + 1; j < len(tasks); j++ {
				if !swapAllowed(in, templates, tasks, i, j, strengeRotation) {
					continue
				}
				trial := swapAssignees(tasks, i, j)
				if !feasible(in, trial) {
					continue
				}
				next := imbalanceOf(in, trial)
				if next >= current {
					continue
				}
				tasks, current, improved = trial, next, true
			}
		}
		if !improved {
			break
		}
	}
	return tasks
}

// swapAssignees liefert eine Kopie mit vertauschten Zuständigkeiten.
//
// Beide Aufgaben bekommen dabei die Begründung „Ausgleich" — und das ist
// keine Nachlässigkeit: Nach einem Tausch stimmt die ursprüngliche Begründung
// nicht mehr. Sie stehen zu lassen hieße, die App behauptet „Rotation,
// zuletzt bei Anna", wo in Wahrheit die Lastrechnung entschieden hat. Eine
// falsche Begründung ist schlimmer als eine unscharfe.
func swapAssignees(tasks []PlannedTask, i, j int) []PlannedTask {
	out := make([]PlannedTask, len(tasks))
	copy(out, tasks)
	out[i].AssigneeID, out[j].AssigneeID = tasks[j].AssigneeID, tasks[i].AssigneeID
	out[i].Reason = Reason{Code: ReasonBalance}
	out[j].Reason = Reason{Code: ReasonBalance}
	return out
}

// swapAllowed prüft die Regeln, die der Ausgleich nicht verletzen darf.
func swapAllowed(in Input, templates map[string]TaskTemplate, tasks []PlannedTask, i, j int, strengeRotation bool) bool {
	a, b := tasks[i], tasks[j]

	if a.AssigneeID == b.AssigneeID {
		return false
	}
	ta, ok := templates[a.TemplateID]
	if !ok {
		return false
	}
	tb, ok := templates[b.TemplateID]
	if !ok {
		return false
	}

	// Feste Zuordnungen bleiben, wo sie sind — und was einer Person selbst
	// gehört, erst recht.
	if ta.Distribution == DistFixed || tb.Distribution == DistFixed {
		return false
	}
	if ta.PerPerson || tb.PerPerson {
		return false
	}

	// Eignung: Alters- und Rollenregeln gelten nach dem Tausch wie davor.
	if !eligibleFor(in, ta, b.AssigneeID) || !eligibleFor(in, tb, a.AssigneeID) {
		return false
	}

	// Rotation über die Wochen: nicht zurück an die Person, die zuletzt dran
	// war — außer im zweiten Durchgang, wenn die Woche sonst schief bleibt.
	if strengeRotation {
		if in.History.lastAssignee(ta.ID) == b.AssigneeID {
			return false
		}
		if in.History.lastAssignee(tb.ID) == a.AssigneeID {
			return false
		}
	}

	// Rotation innerhalb der Woche: Niemand bekommt durch den Ausgleich einen
	// zweiten Termin derselben Vorlage. Ohne diese Regel führt der Durchgang
	// zusammen, was assign bewusst getrennt hat — dreimal Abendessen kochen
	// bei derselben Person, weil die Zahlen dadurch aufgehen. Genau die
	// Schieflage, gegen die der Planer antritt.
	//
	// Der eigene Tauschpartner zählt nicht mit: Er verlässt diese Person ja.
	if holdsOther(tasks, a.TemplateID, b.AssigneeID, j) {
		return false
	}
	if holdsOther(tasks, b.TemplateID, a.AssigneeID, i) {
		return false
	}

	return true
}

// holdsOther sagt, ob memberID noch eine andere Aufgabe dieser Vorlage hat —
// die Aufgabe an der Stelle except ausgenommen.
func holdsOther(tasks []PlannedTask, templateID, memberID string, except int) bool {
	for k, t := range tasks {
		if k == except {
			continue
		}
		if t.TemplateID == templateID && t.AssigneeID == memberID {
			return true
		}
	}
	return false
}

func eligibleFor(in Input, t TaskTemplate, memberID string) bool {
	for _, m := range eligibleMembers(t, in) {
		if m.ID == memberID {
			return true
		}
	}
	return false
}

// feasible rechnet die Kapazität für einen kompletten Plan neu durch: Minuten
// je Person und Tag, und die Höchstzahl Aufgaben je Person und Tag.
//
// Bewusst über den ganzen Plan statt inkrementell. Es ist ein paar Rechnungen
// teurer und dafür offensichtlich richtig — bei zwanzig Aufgaben pro Woche ein
// Tausch, den man ohne Nachdenken macht.
func feasible(in Input, tasks []PlannedTask) bool {
	states := newStates(in.Household)
	for _, t := range tasks {
		st, ok := states[t.AssigneeID]
		if !ok {
			return false
		}
		i := weekdayIndex(t.Day)
		if st.count[i] >= in.Limits.MaxTasksPerMemberDay {
			return false
		}
		if st.remaining[i] < t.DurationMin {
			return false
		}
		// MaxHeadLoadPerDay wird hier bewusst NICHT geprüft.
		//
		// Die Grenze gehört in die Zuteilung, die den Tag wählen kann. Der
		// Ausgleich tauscht nur Zuständige und lässt den Tag stehen (siehe
		// ADR-0003, offener Punkt „verschiebt keine Tage"). Eine
		// Tagesgrenze in einem Durchgang zu prüfen, der den Tag nicht ändern
		// kann, nimmt Möglichkeiten weg, ohne je eine bessere zu finden — und
		// der Preis dafür ist eine unfaire Woche.
		//
		// Zwischen „eine Person trägt das Dreifache" und „ein voller Montag"
		// ist die Antwort nicht offen: Fairness ist das Versprechen, der
		// ruhige Montag ist die Kür. Sobald der Ausgleich Tage verschieben
		// kann, gehört die Prüfung hierher zurück.
		charge(st, t.Day, t, in.Limits)
	}
	return true
}

// imbalanceOf misst, wie unfair ein Plan verteilt ist — als eine Zahl, je
// kleiner desto besser.
//
// Zwei Dinge unterscheiden die Rechnung von der naheliegenden:
//
// Der gerechte Anteil richtet sich nach der Kapazität, nicht nach der
// Kopfzahl: Wer halb so viel Zeit hat, soll halb so viel tragen. Ein Plan,
// der allen dieselbe absolute Last gibt, buchte eine Dreizehnjährige zu 93
// Prozent aus, während ihre Mutter bei 50 Prozent stand — beides gemessen an
// einem Wochenplan, der auf dem Papier ausgeglichen aussah.
//
// Die Kopflast geht als eigener Posten ein, nicht nur als Gleichstandsregel.
// In der ersten Fassung entschied sie erst, wenn die gewichtete Last exakt
// gleich war — was praktisch nie vorkommt. Ergebnis: 737 zu 740 gewichtete
// Minuten bei Kopflast 22 zu 15. Beide Abweichungen zählen jetzt zusammen,
// und sie sind vergleichbar, weil ein Kopflastpunkt in HeadLoadMinutes
// umgerechnet wird — derselbe Wechselkurs, mit dem auch Weight rechnet.
//
// Gerechnet wird ganzzahlig: Statt durch die Gesamtkapazität zu teilen, wird
// jede Abweichung mit ihr multipliziert. Der gemeinsame Faktor ändert die
// Reihenfolge zweier Pläne nicht, erspart aber Fließkommazahlen im Kern —
// eine Sorge weniger beim Determinismus.
//
// Siehe docs/adr/0004-fairness-nach-kapazitaet.md.
func imbalanceOf(in Input, tasks []PlannedTask) int64 {
	loads := balance(in, tasks)
	if len(loads) < 2 {
		return 0
	}

	var totalCapacity, totalWeighted, totalHead int64
	for _, l := range loads {
		totalCapacity += int64(l.Capacity)
		totalWeighted += int64(l.Weighted)
		totalHead += int64(l.HeadLoad)
	}
	if totalCapacity == 0 {
		return 0
	}

	minutesPerPoint := int64(in.Limits.HeadLoadMinutes)

	var score int64
	for _, l := range loads {
		// Nicht cap nennen — das verdeckt die eingebaute Funktion.
		c := int64(l.Capacity)

		// Abweichung vom kapazitätsgerechten Anteil, mal totalCapacity.
		dw := int64(l.Weighted)*totalCapacity - totalWeighted*c
		dh := (int64(l.HeadLoad)*totalCapacity - totalHead*c) * minutesPerPoint

		score += dw*dw + dh*dh
	}
	return score
}

// ueberschussPromille sagt, wie weit die am stärksten belastete Person über
// ihrem gerechten Anteil liegt — in Promille ihrer eigenen Kapazität.
//
// Gerechnet wird auf der gewichteten Last, weil danach verteilt wird. Der
// Maßstab ist die eigene Kapazität und nicht der Haushalt: 20 gewichtete
// Minuten zu viel sind für eine Dreizehnjährige etwas anderes als für ihren
// Vater, und genau dieser Unterschied ist der Grund für den zweiten
// Durchgang.
//
// Null, wenn niemand über seinem Anteil liegt — was bei ganzen Aufgaben nur
// vorkommt, wenn alle exakt darauf sitzen.
func ueberschussPromille(in Input, tasks []PlannedTask) int {
	loads := balance(in, tasks)
	if len(loads) < 2 {
		return 0
	}

	var totalCapacity, totalWeighted int64
	for _, l := range loads {
		totalCapacity += int64(l.Capacity)
		totalWeighted += int64(l.Weighted)
	}
	if totalCapacity == 0 {
		return 0
	}
	ziel := totalWeighted * 1000 / totalCapacity

	groesster := 0
	for _, l := range loads {
		if l.Capacity <= 0 {
			continue
		}
		anteil := int64(l.Weighted) * 1000 / int64(l.Capacity)
		if d := int(anteil - ziel); d > groesster {
			groesster = d
		}
	}
	return groesster
}
