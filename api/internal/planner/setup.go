package planner

import (
	"errors"
	"fmt"
	"sort"
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
	if s.Context.Rooms == 0 {
		s.Context.Rooms = ReferenceRooms
	}
	if s.Context.Baths == 0 {
		s.Context.Baths = ReferenceBaths
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
	if err := ValidName("der Haushalt", s.Name, maxSetupNameLen); err != nil {
		return err
	}
	switch {
	case !ValidHome(s.Context.Home):
		return fmt.Errorf("%w: %q ist keine Wohnform", ErrInvalidSetup, s.Context.Home)
	case len(s.Members) == 0:
		return fmt.Errorf("%w: ohne Personen gibt es nichts zu verteilen", ErrInvalidSetup)
	case len(s.Members) > maxSetupMembers:
		return fmt.Errorf("%w: mehr als %d Personen sind kein Haushalt mehr", ErrInvalidSetup, maxSetupMembers)
	}

	if err := ValidTimezone(s.Timezone); err != nil {
		return err
	}
	if err := ValidSize(s.Context.Rooms, s.Context.Baths); err != nil {
		return err
	}

	if s.Members[0].Role != RolePlanner {
		return fmt.Errorf("%w: wer den Haushalt einrichtet, muss darin planen", ErrInvalidSetup)
	}

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
		case !ValidRole(m.Role):
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
		if err := ValidBirthYear(m.BirthYear); err != nil {
			return err
		}
	}
	return nil
}

// CareForAge rät die Betreuungsform aus dem Alter, statt sie abzufragen.
//
// Eine Frage weniger im Onboarding — und ein Rateschluss, der danebenliegen
// darf, weil die Einstellungen ihn korrigieren. Ohne diese Korrekturmöglichkeit
// wäre er eine Behauptung.
//
// Die Grenzen folgen dem deutschen Alltag: Der Rechtsanspruch auf Betreuung
// gilt ab dem vollendeten ersten Lebensjahr, Krippe ab eins ist der Normalfall.
// Die erste Fassung begann bei drei ("Kindergarten") und übersah damit jedes
// Krippenkind — aufgefallen an einer Zweijährigen, die längst in die Kita geht.
func CareForAge(age int) Care {
	switch {
	case age >= 18 || age <= 0:
		return CareNone
	case age >= 6:
		return CareSchool
	default:
		return CareKita
	}
}

// Die Prüfungen unten stehen einzeln, weil sie zweimal gebraucht werden: beim
// Einrichten für einen ganzen Haushalt und in den Einstellungen für ein
// einzelnes Feld. Zwei Kopien derselben Regel laufen auseinander, und die
// Fassung in den Einstellungen wäre die laxere — dort fällt es später auf.

// ValidName prüft einen Namen, wie ein Mensch ihn prüfen würde.
//
// was benennt, wessen Name gemeint ist („der Haushalt", „die Person"). Das ist
// kein Schmuck: Die Meldung landet unverändert im Formular, und „der Name
// fehlt" hilft niemandem, der zwei Namensfelder vor sich hat.
func ValidName(was, s string, max int) error {
	switch {
	case strings.TrimSpace(s) == "":
		return fmt.Errorf("%w: %s braucht einen Namen", ErrInvalidSetup, was)
	case len([]rune(s)) > max:
		return fmt.Errorf("%w: der Name ist länger als %d Zeichen", ErrInvalidSetup, max)
	}
	return nil
}

// ValidBirthYear lässt 0 zu — das heißt „nicht gefragt" und ist bei
// Erwachsenen die richtige Antwort (siehe Member.IsAdult).
func ValidBirthYear(y int) error {
	if y == 0 {
		return nil
	}
	if y < earliestBirthYear || y > time.Now().Year() {
		return fmt.Errorf("%w: %d ist kein Geburtsjahr", ErrInvalidSetup, y)
	}
	return nil
}

// ValidTimezone prüft, ob es diese Zeitzone auf diesem System gibt. Daran
// hängt, wann ein Tag beginnt und endet (ADR-0002).
func ValidTimezone(name string) error {
	if _, err := time.LoadLocation(name); err != nil {
		return fmt.Errorf("%w: %q ist keine Zeitzone", ErrInvalidSetup, name)
	}
	return nil
}

// ValidSize prüft Zimmer und Bäder.
//
// Die Obergrenzen sind großzügig und trotzdem da: Eine 40 im Zimmerfeld ist
// ein Tippfehler, und ohne Grenze verschöbe er stillschweigend jede Dauer im
// Haushalt an den Deckel.
func ValidSize(zimmer, baeder int) error {
	if zimmer < 1 || zimmer > 15 {
		return fmt.Errorf("%w: %d Zimmer sind keine Wohnung", ErrInvalidSetup, zimmer)
	}
	if baeder < 0 || baeder > 5 {
		return fmt.Errorf("%w: %d Bäder sind zu viele", ErrInvalidSetup, baeder)
	}
	return nil
}

// ValidRole sagt, ob es diese Rolle gibt.
func ValidRole(r Role) bool {
	return r == RolePlanner || r == RoleDoer || r == RoleDependent
}

// ValidHome sagt, ob es diese Wohnform gibt.
func ValidHome(h Home) bool { return h == HomeFlat || h == HomeHouse }

// MaxDailyMinutes ist die Obergrenze je Tag: acht Stunden.
//
// Nicht, weil niemand mehr im Haushalt tut — sondern weil ein Tippfehler wie
// 6000 sonst still die ganze Verteilung kippt. Wer wirklich mehr braucht,
// stößt an eine Zahl und nicht an einen falschen Plan.
const MaxDailyMinutes = 480

// ValidMinutes prüft ein selbst gesetztes Zeitbudget.
//
// Null an allen Tagen ist erlaubt: Genau das ist eine betreute Person, und
// genau das ist jemand, der diese Woche nichts übernehmen kann.
func ValidMinutes(m [7]int) error {
	for i, v := range m {
		if v < 0 || v > MaxDailyMinutes {
			return fmt.Errorf("%w: %d Minuten an Tag %d sind keine Kapazität", ErrInvalidSetup, v, i+1)
		}
	}
	return nil
}

// OpenQuestions sind die Fakten, deren Antwort den Plan am meisten verändern
// würde — höchstens so viele, wie angefragt.
//
// Die Reihenfolge entscheidet, was gefragt wird: erst das Faktum, an dem die
// meisten Vorlagen hängen, bei Gleichstand alphabetisch, damit zwei Aufrufe
// dasselbe ergeben.
//
// Die Obergrenze ist derselbe Gedanke wie die Startdichte: Wer beim ersten
// Öffnen vierzehn Fragen sieht, beantwortet keine. Zwei mit sichtbarem Nutzen
// werden beantwortet.
func OpenQuestions(skipped []Skipped, max int) []string {
	zaehler := map[string]int{}
	for _, s := range skipped {
		if s.Code == SkipUnknown && s.Fact != "" {
			zaehler[s.Fact]++
		}
	}
	fakten := make([]string, 0, len(zaehler))
	for f := range zaehler {
		fakten = append(fakten, f)
	}
	sort.Slice(fakten, func(i, j int) bool {
		if zaehler[fakten[i]] != zaehler[fakten[j]] {
			return zaehler[fakten[i]] > zaehler[fakten[j]]
		}
		return fakten[i] < fakten[j]
	})
	if max > 0 && len(fakten) > max {
		fakten = fakten[:max]
	}
	return fakten
}
