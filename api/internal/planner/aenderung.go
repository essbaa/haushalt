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
	return nil
}

// Touched sagt, ob überhaupt etwas geändert werden soll. Eine leere Änderung
// ist kein Fehler, aber auch kein Schreibvorgang.
func (c MemberChange) Touched() bool {
	return c.Name != nil || c.Role != nil || c.BirthYear != nil || c.Minutes != nil
}

func (c HouseholdChange) Touched() bool {
	return c.Name != nil || c.Home != nil || c.HasCar != nil ||
		c.HasYard != nil || c.Pets != nil || c.Timezone != nil
}
