package planner

import "fmt"

// HouseholdChange ist eine Änderung an den Einstellungen eines Haushalts.
//
// Alles Zeiger: Der Unterschied zwischen „nicht mitgeschickt" und „auf leer
// gesetzt" ist der ganze Sinn einer Einstellungsseite. Ohne ihn müsste jedes
// Formular immer alle Felder senden, und wer zwei Felder gleichzeitig ändert,
// überschreibt das dritte mit dem Stand von vorhin.
type HouseholdChange struct {
	Name     *string
	Home     *Home
	HasCar   *bool
	HasYard  *bool
	Pets     *[]string
	Timezone *string

	// Rooms und Baths skalieren die Putzaufgaben. Korrigierbar, weil man
	// umzieht — und weil beim Einrichten leicht danebengetippt ist.
	Rooms *int
	Baths *int
}

func (c HouseholdChange) Validate() error {
	if c.Name != nil {
		if err := ValidName("der Haushalt", *c.Name, maxSetupNameLen); err != nil {
			return err
		}
	}
	if c.Home != nil && !ValidHome(*c.Home) {
		return fmt.Errorf("%w: %q ist keine Wohnform", ErrInvalidSetup, *c.Home)
	}
	if c.Pets != nil && len(*c.Pets) > 6 {
		return fmt.Errorf("%w: mehr als sechs Haustiere", ErrInvalidSetup)
	}
	if c.Timezone != nil {
		if err := ValidTimezone(*c.Timezone); err != nil {
			return err
		}
	}
	if c.Rooms != nil || c.Baths != nil {
		zimmer, baeder := ReferenceRooms, ReferenceBaths
		if c.Rooms != nil {
			zimmer = *c.Rooms
		}
		if c.Baths != nil {
			baeder = *c.Baths
		}
		if err := ValidSize(zimmer, baeder); err != nil {
			return err
		}
	}
	return nil
}

// MemberChange ist eine Änderung an einer Person.
type MemberChange struct {
	Name *string
	Role *Role

	// BirthYear auf 0 gesetzt heißt „löschen", nicht „Jahr null": Bei einem
	// Erwachsenen ist kein Geburtsjahr die richtige Antwort, und man muss sie
	// eintragen können.
	BirthYear *int

	// Care ist die Betreuungsform. Beim Einrichten geraten, hier korrigierbar
	// — an ihr hängt, welche Aufgaben der Haushalt überhaupt hat.
	Care *Care

	// Minutes ist das selbst gesetzte Zeitbudget je Wochentag, Index 0 =
	// Montag. Die drei Stufen aus dem Onboarding rechnet die Grenze vorher in
	// Minuten um — hier unten gibt es nur noch Minuten.
	Minutes *[7]int
}

func (c MemberChange) Validate() error {
	if c.Name != nil {
		if err := ValidName("die Person", *c.Name, maxMemberNameLen); err != nil {
			return err
		}
	}
	if c.Role != nil && !ValidRole(*c.Role) {
		return fmt.Errorf("%w: %q ist keine Rolle", ErrInvalidSetup, *c.Role)
	}
	if c.BirthYear != nil {
		if err := ValidBirthYear(*c.BirthYear); err != nil {
			return err
		}
	}
	if c.Minutes != nil {
		if err := ValidMinutes(*c.Minutes); err != nil {
			return err
		}
	}
	if c.Care != nil && !ValidCare(*c.Care) {
		return fmt.Errorf("%w: %q ist keine Betreuungsform", ErrInvalidSetup, *c.Care)
	}
	return nil
}

// ValidCare sagt, ob es diese Betreuungsform gibt.
func ValidCare(c Care) bool {
	return c == CareNone || c == CareKita || c == CareSchool
}

// Touched sagt, ob überhaupt etwas geändert werden soll. Eine leere Änderung
// ist kein Fehler, aber auch kein Schreibvorgang.
func (c MemberChange) Touched() bool {
	return c.Name != nil || c.Role != nil || c.BirthYear != nil ||
		c.Minutes != nil || c.Care != nil
}

func (c HouseholdChange) Touched() bool {
	return c.Name != nil || c.Home != nil || c.HasCar != nil ||
		c.HasYard != nil || c.Pets != nil || c.Timezone != nil ||
		c.Rooms != nil || c.Baths != nil
}

// OwnTask ist eine Aufgabe, die ein Haushalt selbst anlegt.
//
// Bewusst wenige Felder. Eine eigene Aufgabe anzulegen ist Dateneingabe, und
// Dateneingabe ist das, wogegen dieses Produkt antritt — sie ist der Ausweg
// für das, was in der Bibliothek fehlt, nicht der Weg.
//
// Die Kopflast wird nicht als Zahl abgefragt, sondern aus zwei Fragen
// abgeleitet, die man beantworten kann, ohne das Modell zu kennen. „Kopflast 0
// bis 3" würde geraten; „muss jemand daran denken?" wird beantwortet.
type OwnTask struct {
	Title    string
	Category Category

	// DurationMin ist die geschätzte Dauer. Sie trägt die Fairnessrechnung
	// mit — eine zu niedrige Zahl verschiebt die Bilanz des ganzen Haushalts,
	// ohne dass es jemand merkt.
	DurationMin int

	// Remember: Muss jemand daran denken, oder sieht man es?
	Remember bool
	// Arrange: Muss man erst etwas klären — Termin machen, nachsehen, jemanden
	// fragen?
	Arrange bool

	// EveryDays ist der gewünschte Abstand: 1, 7, 14, 30.
	EveryDays int

	// AdultsOnly schränkt auf Erwachsene ein.
	AdultsOnly bool

	// NeedsAgreement: Muss festgelegt werden, wer wann? Dann verteilt der
	// Planer nicht, sondern fordert ein Wochenraster ein — für alles, was am
	// Arbeitsplan hängt statt an freier Zeit.
	NeedsAgreement bool
}

// HeadLoad leitet die Kopflast aus den beiden Fragen ab.
//
// Daran denken zählt einfach, etwas klären zählt doppelt: Ein Termin, den man
// vereinbaren muss, kostet mehr Aufmerksamkeit als eine Aufgabe, an die man
// sich nur erinnern muss. Ergibt 0, 1, 2 oder 3 — dieselbe Skala wie in der
// Bibliothek.
func (o OwnTask) HeadLoad() HeadLoad {
	wert := 0
	if o.Remember {
		wert++
	}
	if o.Arrange {
		wert += 2
	}
	return HeadLoad(wert)
}

// Kind ergibt sich aus der Kopflast: Was Kopfarbeit kostet, ist Organisation.
//
// Die Unterscheidung steuert die Startdichte, und sie abzufragen wäre eine
// Frage nach einer Einteilung, die aus der vorigen Antwort schon folgt.
func (o OwnTask) Kind() Kind {
	if o.HeadLoad() >= 2 {
		return KindOrg
	}
	return KindDo
}

// Validate prüft eine eigene Aufgabe.
func (o OwnTask) Validate() error {
	if err := ValidName("die Aufgabe", o.Title, 80); err != nil {
		return err
	}
	if o.DurationMin < 1 || o.DurationMin > 480 {
		return fmt.Errorf("%w: %d Minuten sind keine Dauer", ErrInvalidSetup, o.DurationMin)
	}
	switch o.EveryDays {
	case 1, 7, 14, 30:
	default:
		return fmt.Errorf("%w: %d Tage sind kein Rhythmus", ErrInvalidSetup, o.EveryDays)
	}
	for _, c := range AllCategories {
		if o.Category == c {
			return nil
		}
	}
	return fmt.Errorf("%w: %q ist kein Bereich", ErrInvalidSetup, o.Category)
}

// Template macht aus der Eingabe eine vollwertige Vorlage.
func (o OwnTask) Template(id string) TaskTemplate {
	dist := DistRotate
	if o.AdultsOnly {
		dist = DistAdultsOnly
	}
	return TaskTemplate{
		ID:             id,
		Title:          o.Title,
		Category:       o.Category,
		Kind:           o.Kind(),
		DurationMin:    o.DurationMin,
		HeadLoad:       o.HeadLoad(),
		Rhythm:         Rhythm{Type: RhythmWindow, EveryDays: o.EveryDays},
		Distribution:   dist,
		NeedsAgreement: o.NeedsAgreement,
		// Weich: Eine selbst angelegte Aufgabe bekommt keine harte Frist. Wer
		// eine braucht, trägt einen Anlass ein — dafür gibt es Termine.
		Failure: FailureSoft,
		Source:  SourceHousehold,
	}
}
