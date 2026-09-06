package planner

// capDensity begrenzt die Organisationsaufgaben auf Limits.MaxOrgTasks.
//
// Das ist keine technische Grenze, sondern die wichtigste Produktentscheidung
// im Planer: Lieber zwanzig richtige Aufgaben als sechzig vollständige. Ein
// Plan, der beim ersten Öffnen erschlägt, wird nicht korrigiert, sondern
// gelöscht — und bestätigt genau das Gefühl, gegen das die App antritt.
//
// Ausführungsaufgaben bleiben unangetastet: Sie passieren ohnehin, sie stehen
// nur jetzt auch im Plan. Was Last erzeugt, ist die Kopfarbeit.
//
// Aufgaben mit harter Frist werden nie zurückgehalten. Eine verpasste
// Anmeldefrist ist kein Beitrag zur Entlastung.
func capDensity(in Input, cands []candidate) ([]candidate, []Skipped) {
	limit := in.Limits.MaxOrgTasks
	kept := make([]candidate, 0, len(cands))
	var skipped []Skipped
	org := 0

	// cands kommt bereits nach Dringlichkeit sortiert aus selectDue; die
	// wichtigsten Organisationsaufgaben überleben damit die Kürzung.
	for _, c := range cands {
		if c.tmpl.Kind != KindOrg {
			kept = append(kept, c)
			continue
		}
		if c.tmpl.Failure == FailureHard {
			kept = append(kept, c)
			org++
			continue
		}
		if org < limit {
			kept = append(kept, c)
			org++
			continue
		}
		skipped = append(skipped, Skipped{c.tmpl.ID, c.tmpl.Title, SkipDensity})
	}
	return kept, skipped
}
