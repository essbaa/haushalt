package planner

import "fmt"

// Reassign prüft, ob eine Aufgabe von Hand an eine bestimmte Person gehen
// darf, und liefert die Begründung für die neue Zuteilung.
//
// Der Unterschied zu Handover ist nicht technisch, sondern fachlich: Handover
// *sucht* jemanden nach den Regeln des Planers. Hier hat ein Mensch bereits
// entschieden, und die Frage ist nur noch, ob der Planer widersprechen muss.
//
// Er widerspricht selten. Die Regeln des Planers zerfallen in zwei Sorten, und
// die Unterscheidung ist dieselbe wie überall sonst im Paket (ADR-0012):
//
//	Bedingungen  gelten auch hier. Wer die Aufgabe nicht übernehmen *kann*,
//	             übernimmt sie nicht: Alter, Rolle, die Verteilungsregel der
//	             Vorlage, und dass eine eigene Aufgabe der Person gehört.
//
//	Vorlieben    gelten hier nicht. Rotation, Auslastung, Kopflast je Tag, die
//	             Obergrenze für Einzelaufgaben von Kindern — das alles sind
//	             Annahmen des Planers darüber, was gut wäre. Ein Mensch, der
//	             etwas anderes weiß, sticht sie.
//
// Das ist der ganze Sinn der Sache. `manual = true` in der Zuteilung ist die
// wertvollste Rückmeldung, die dieses Produkt bekommt: Jede Korrektur sagt,
// wo der Planer danebenlag. Eine Korrektur, die der Planer erst genehmigen
// muss, sagt gar nichts.
func Reassign(in Input, tasks []PlannedTask, taskID, memberID string) (Reason, error) {
	var aufgabe PlannedTask
	gefunden := false
	for _, t := range tasks {
		if t.ID == taskID {
			aufgabe, gefunden = t, true
			break
		}
	}
	if !gefunden {
		return Reason{}, ErrUnknownTask
	}

	var vorlage TaskTemplate
	for _, v := range in.Templates {
		if v.ID == aufgabe.TemplateID {
			vorlage = v
			break
		}
	}
	if vorlage.ID == "" {
		return Reason{}, ErrUnknownTask
	}

	// Eine Aufgabe, die einer Person selbst gehört, wandert nicht: „Dein Bett
	// beziehen" bei jemand anderem ist eine andere Aufgabe, nicht dieselbe in
	// anderen Händen (ADR-0005).
	if vorlage.PerPerson || aufgabe.Reason.Code == ReasonOwn {
		return Reason{}, NotEligibleError{Grund: fmt.Sprintf("%q gehört der Person selbst", vorlage.Title)}
	}

	var person Member
	for _, m := range in.Household.Members {
		if m.ID == memberID {
			person = m
			break
		}
	}
	if person.ID == "" {
		return Reason{}, ErrUnknownMember
	}

	// Ab hier: die Bedingungen der Vorlage, einzeln geprüft statt über
	// eligibleMembers. Der Grund ist die Meldung — „diese Person kann das
	// nicht" hilft niemandem, „Mia ist 13, die Vorlage gilt ab 16" schon.
	if !person.CanPerform() {
		return Reason{}, NotEligibleError{Grund: fmt.Sprintf("%s führt keine Aufgaben aus", wer(person))}
	}
	if vorlage.Distribution == DistAdultsOnly && !person.IsAdult() {
		return Reason{}, NotEligibleError{Grund: fmt.Sprintf("%q ist für Erwachsene", vorlage.Title)}
	}
	if vorlage.Distribution == DistChildrenOnly && person.IsAdult() {
		return Reason{}, NotEligibleError{Grund: fmt.Sprintf("%q ist für Kinder", vorlage.Title)}
	}
	if !person.IsAdult() && person.Age < vorlage.MinAge {
		return Reason{}, NotEligibleError{Grund: fmt.Sprintf("%q gilt ab %d Jahren", vorlage.Title, vorlage.MinAge)}
	}

	return Reason{Code: ReasonManual, Previous: aufgabe.AssigneeID}, nil
}

// wer liefert den Namen, wenn es einen gibt, und sonst die Kennung. Die
// Meldung geht an einen Menschen; eine UUID darin wäre eine Ausrede.
func wer(m Member) string {
	if m.Name != "" {
		return m.Name
	}
	return m.ID
}
