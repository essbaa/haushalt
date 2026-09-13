package planner

import (
	"errors"
	"fmt"
	"strings"
	"time"
)

// ErrInvalidSetup sagt, dass aus dieser Eingabe kein Haushalt wird. Immer
// eingepackt in eine Meldung, die benennt, was fehlt — die Grenze macht daraus
// eine 400, und der Text landet im Formular.
var ErrInvalidSetup = errors.New("planner: eingabe ergibt keinen haushalt")

// Setup ist, was beim Einrichten eines Haushalts gefragt wird — und nur das.
//
// Die Liste ist kurz, weil sie kurz sein muss: Eine App gegen Mental Load darf
// beim Einrichten keine erzeugen. Gefragt wird ausschließlich, was den ersten
// Plan verändert. Alles Weitere (Arbeitszeiten, Vorlieben, Betreuungszeiten)
// lernt die App aus dem Umverteilen, statt es vorher abzufragen.
// Members[0] ist die Person, die einrichtet — das ist Teil des Vertrags und
// nicht bloß Reihenfolge: An ihr hängt die Anmeldung, und sie muss planen
// dürfen, sonst wäre der Haushalt von der ersten Sekunde an verwaist.
type Setup struct {
	Name     string
	Context  Context
	Timezone string
	Members  []SetupMember
}

// SetupMember ist eine Person, wie sie im Onboarding angegeben wird.
type SetupMember struct {
	Name string
	Role Role

	// BirthYear ist 0, wenn nicht gefragt wurde — das ist bei Erwachsenen der
	// Normalfall und bedeutet nicht „unbekannt alt", sondern „erwachsen".
	// Siehe Member.IsAdult.
	BirthYear int

	// Budget ist die grobe Zeitstufe. Leer heißt mittel.
	Budget TimeBudget
}

const (
	maxSetupNameLen   = 80
	maxMemberNameLen  = 40
	maxSetupMembers   = 12
	earliestBirthYear = 1900
)

// Normalized füllt die Lücken, die das Formular offen lassen darf, und räumt
// Leerraum weg. Es prüft nichts — dafür ist Validate da.
//
// Zwei Dinge passieren hier, die keine Kosmetik sind: Betreute Personen
// bekommen das Zeitbudget "keine", egal was der Client schickt (sie erzeugen
// Arbeit, sie übernehmen keine), und eine fehlende Zeitzone wird Europe/Berlin
// statt UTC — ein Haushalt lebt in einer Zeitzone, nicht in der Serverzeit.
func (s Setup) Normalized() Setup {
	s.Name = strings.TrimSpace(s.Name)
	if s.Timezone == "" {
		s.Timezone = "Europe/Berlin"
	}
	if s.Context.Home == "" {
		s.Context.Home = HomeFlat
	}
	if s.Context.Pets == nil {
		s.Context.Pets = []string{}
	}

	out := make([]SetupMember, 0, len(s.Members))
	for _, m := range s.Members {
		m.Name = strings.TrimSpace(m.Name)
		switch {
		case m.Role == RoleDependent:
			m.Budget = BudgetNone
		case m.Budget == "":
			m.Budget = BudgetMedium
		}
		out = append(out, m)
	}
	s.Members = out
	return s
}

// Validate prüft die Eingabe so, wie ein Mensch sie prüfen würde: Ergibt das
// einen Haushalt, den man planen kann?
func (s Setup) Validate() error {
	switch {
	case s.Name == "":
		return fmt.Errorf("%w: der Haushalt braucht einen Namen", ErrInvalidSetup)
	case len([]rune(s.Name)) > maxSetupNameLen:
		return fmt.Errorf("%w: der Name ist länger als %d Zeichen", ErrInvalidSetup, maxSetupNameLen)
	case s.Context.Home != HomeFlat && s.Context.Home != HomeHouse:
		return fmt.Errorf("%w: %q ist keine Wohnform", ErrInvalidSetup, s.Context.Home)
	case len(s.Members) == 0:
		return fmt.Errorf("%w: ohne Personen gibt es nichts zu verteilen", ErrInvalidSetup)
	case len(s.Members) > maxSetupMembers:
		return fmt.Errorf("%w: mehr als %d Personen sind kein Haushalt mehr", ErrInvalidSetup, maxSetupMembers)
	}

	if _, err := time.LoadLocation(s.Timezone); err != nil {
		return fmt.Errorf("%w: %q ist keine Zeitzone", ErrInvalidSetup, s.Timezone)
	}

	if s.Members[0].Role != RolePlanner {
		return fmt.Errorf("%w: wer den Haushalt einrichtet, muss darin planen", ErrInvalidSetup)
	}

	jetzt := time.Now().Year()
	namen := map[string]bool{}

	for _, m := range s.Members {
		switch {
		case m.Name == "":
			return fmt.Errorf("%w: eine Person ohne Namen", ErrInvalidSetup)
		case len([]rune(m.Name)) > maxMemberNameLen:
			return fmt.Errorf("%w: der Name %q ist zu lang", ErrInvalidSetup, m.Name)
		case namen[strings.ToLower(m.Name)]:
			// Zwei gleiche Namen im Plan sind kein Datenfehler, sondern ein
			// Bedienfehler: Niemand weiß dann, wer den Müll rausbringt.
			return fmt.Errorf("%w: %q kommt zweimal vor", ErrInvalidSetup, m.Name)
		case m.Role != RolePlanner && m.Role != RoleDoer && m.Role != RoleDependent:
			return fmt.Errorf("%w: %q ist keine Rolle", ErrInvalidSetup, m.Role)
		case !m.Budget.Known():
			return fmt.Errorf("%w: %q ist keine Zeitstufe", ErrInvalidSetup, m.Budget)
		}
		namen[strings.ToLower(m.Name)] = true

		// Betreute Personen brauchen ein Alter, weil daran hängt, welche
		// Aufgaben der Haushalt überhaupt hat: Wechselkleidung für die Kita
		// und Zahnkontrolle für ein Kind sind etwas anderes als die Begleitung
		// eines pflegebedürftigen Elternteils.
		if m.Role == RoleDependent && m.BirthYear == 0 {
			return fmt.Errorf("%w: bei %q fehlt das Geburtsjahr — daran hängen die Aufgaben", ErrInvalidSetup, m.Name)
		}
		if m.BirthYear != 0 && (m.BirthYear < earliestBirthYear || m.BirthYear > jetzt) {
			return fmt.Errorf("%w: %d ist kein Geburtsjahr", ErrInvalidSetup, m.BirthYear)
		}
	}
	return nil
}

// CareForAge rät die Betreuungsform aus dem Alter, statt sie abzufragen.
//
// Eine Frage weniger im Onboarding, und in Deutschland trifft die Regel
// meistens: Krippe bleibt außen vor, Kindergarten ab drei, Schule ab sechs.
// Sie ist ein Vorschlag, keine Vorschrift — sobald es die Einstellungen gibt,
// gehört die Betreuungsform dorthin, denn an ihr hängen die
// Organisationsaufgaben.
func CareForAge(age int) Care {
	switch {
	case age >= 18 || age <= 0:
		return CareNone
	case age >= 6:
		return CareSchool
	case age >= 3:
		return CareKita
	default:
		return CareNone
	}
}
