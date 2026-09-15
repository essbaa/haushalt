package planner

import (
	"errors"
	"fmt"
	"sort"
	"time"
)

// ErrUnknownHousehold sagt, dass es diesen Haushalt nicht gibt.
//
// Der Fehler steht hier und nicht bei einer der beiden Quellen (Dateien oder
// Datenbank), weil beide ihn brauchen und die HTTP-Schicht ihn prüfen muss,
// ohne zu wissen, woher die Daten kommen. Ein Wachposten-Fehler statt einer
// Zeichenkette: errors.Is entscheidet, nicht ein Textvergleich.
var ErrUnknownHousehold = errors.New("planner: unbekannter haushalt")

// ErrNotAllowed sagt, dass der Aufrufer darf, was er will, nur das nicht.
//
// Anders als ErrUnknownHousehold: Der Aufrufer gehört dazu, ihm fehlt die
// Rolle. Das darf er erfahren — verschwiegen wird nur die Existenz von
// Dingen, die ihn nichts angehen.
var ErrNotAllowed = errors.New("planner: dafür fehlt die berechtigung")

// ErrUnknownInvitation deckt beides ab: es gibt den Code nicht, er ist
// abgelaufen, oder er wurde schon benutzt.
//
// Absichtlich eine Antwort für drei Fälle. Wer einen Code durchprobiert, soll
// nicht erfahren, welcher davon zutrifft.
var ErrUnknownInvitation = errors.New("planner: einladung ungültig")

// ErrUnknownTask deckt beides ab: es gibt die Aufgabe nicht, oder sie gehört
// zu einem Haushalt, in dem der Aufrufer nichts zu suchen hat. Dieselbe
// Auskunft für beides — siehe ErrUnknownHousehold.
var ErrUnknownTask = errors.New("planner: unbekannte aufgabe")

// ErrUnknownMember sagt, dass es diese Person in diesem Haushalt nicht gibt.
//
// Kein 404 für den Aufrufer: Er ist im Haushalt und darf erfahren, dass die
// Kennung nicht dazugehört. Verschwiegen wird nur, was ihn nichts angeht.
var ErrUnknownMember = errors.New("planner: unbekannte person")

// ErrNotEligible sagt, dass diese Person diese Aufgabe nicht übernehmen kann.
//
// Nicht „darf nicht“, sondern „kann nicht“: Alter, Rolle, die
// Verteilungsregel der Vorlage. Vorlieben des Planers — Rotation, Auslastung,
// Aufgabenlänge bei Kindern — erzeugen diesen Fehler ausdrücklich nicht
// (siehe Reassign).
//
// Der Grund ist für Menschen geschrieben und steht in NotEligibleError.Grund
// — ohne technisches Vorwort, damit er unverändert in die Antwort gehen kann.
var ErrNotEligible = errors.New("planner: nicht geeignet")

// NotEligibleError trägt den Grund, den ein Mensch lesen soll.
//
// Ein eigener Typ statt fmt.Errorf("%w: …"), weil die Oberfläche genau diesen
// Satz anzeigt. Würde sie err.Error() nehmen, stünde „planner: nicht geeignet:
// …“ im Formular — die Herkunft des Fehlers geht niemanden etwas an, der
// gerade eine Aufgabe verschieben will.
type NotEligibleError struct {
	Grund string
}

func (e NotEligibleError) Error() string { return ErrNotEligible.Error() + ": " + e.Grund }

// Is macht errors.Is(err, ErrNotEligible) wahr, ohne dass der Sentinel
// eingewickelt werden muss.
func (e NotEligibleError) Is(target error) bool { return target == ErrNotEligible }

// Invitation ist ein ausgestellter Code.
//
// For ist der Name der Person, die dieser Code übernimmt — leer, wenn jemand
// dazukommt, den es im Haushalt noch nicht gibt. Der Unterschied gehört in die
// Antwort: „Dieser Link macht Asmae zu Asmae" ist etwas anderes als „dieser
// Link holt irgendjemanden dazu", und nur der Einladende kann das prüfen.
type Invitation struct {
	Code  string
	Until time.Time
	For   string

	// Role ist die Rolle, die der Eingeladene bekommt. Bei einer Einladung
	// für eine vorhandene Person stammt sie von dieser Person und nicht aus
	// der Anfrage — deshalb steht sie in der Antwort und wird nicht dort
	// abgeschrieben, wo der Aufruf herkam.
	Role Role
}

// Input ist alles, was der Planer braucht. Nichts wird nachgeladen.
type Input struct {
	Household Household
	Templates []TaskTemplate
	Week      Week
	History   History
	Limits    Limits
}

// Limits sind die Stellschrauben der Planung. Sie stehen bewusst in der
// Eingabe und nicht als Konstanten im Code: Sie sind Produktentscheidungen,
// keine Implementierungsdetails.
type Limits struct {
	// MaxOrgTasks begrenzt die Organisationsaufgaben pro Woche — die
	// Startdichte aus dem Konzept. Ein Plan, der beim ersten Öffnen erschlägt,
	// wird nicht korrigiert, sondern gelöscht.
	MaxOrgTasks int

	// MaxTasksPerMemberDay verhindert, dass ein Tag zur Liste wird, auch wenn
	// die Minuten rechnerisch passen.
	MaxTasksPerMemberDay int

	// MaxMinutesForChild ist die längste Einzelaufgabe, die jemand unter 18
	// bekommt.
	//
	// Keine Rechenkorrektur, sondern eine Produktaussage. Das Alter in den
	// Vorlagen sagt, ob ein Kind eine Aufgabe *kann* — nicht, wie groß ein
	// einzelner Block sein darf. Bad putzen gilt ab 12; zwei Bäder am Stück
	// sind etwas anderes als eines, und ein Plan, in dem das Kind den längsten
	// Block trägt, wird gelöscht, egal was die Prozentzahl sagt.
	//
	// Eine Vorliebe, keine Bedingung: Bleibt sonst niemand übrig, bekommt das
	// Kind die Aufgabe trotzdem — sie ganz wegfallen zu lassen wäre schlimmer.
	MaxMinutesForChild int

	// MaxHeadLoadPerDay begrenzt die Kopflast einer Person an einem Tag.
	//
	// Die Startdichte begrenzt die Kopfarbeit je Woche und sagt nichts
	// darüber, wie sie über die Woche liegt. Ohne diese zweite Grenze landet
	// alles am Montag: Organisationsaufgaben sind am dringendsten, bekommen
	// also den frühesten Tag — und der Montag trägt die halbe Woche.
	//
	// Vier heißt in der Praxis: ein Termin mit viel Kopflast, oder zwei mit
	// wenig. Harte Fristen sind ausgenommen.
	MaxHeadLoadPerDay int

	// MaxOvershootPermille sagt, ab wann der Ausgleich die Rotation brechen
	// darf: wenn jemand um mehr als diesen Anteil seiner eigenen Kapazität
	// über seinem gerechten Anteil liegt.
	//
	// Die Rotation über die Wochen war im Ausgleich ein Veto, im assign nur
	// eine Vorliebe — und das Veto stand in dem Durchgang, der ausgleichen
	// soll. Bezahlt hat es die Person mit der kleinsten Kapazität: In Familie
	// B lag Mia (13) bei 79 Prozent ihrer Zeit und ihr Vater bei 67, weil ein
	// Tausch daran scheiterte, dass er dieselbe Aufgabe letzte Woche hatte.
	//
	// 25 Promille heißt: etwa eine kleine Aufgabe zu viel. Darunter lohnt die
	// Wiederholung nicht; darüber ist die Woche spürbar schief.
	//
	// Null schaltet den zweiten Durchgang ab — dann bleibt die Rotation
	// unantastbar, wie vorher.
	MaxOvershootPermille int

	// HeadLoadMinutes ist der Preis eines Kopflast-Punktes in Minuten für die
	// Lastrechnung. Der Kern der Zwei-Achsen-Idee: Wer Termine koordiniert,
	// verbringt wenig Zeit und trägt viel. Ein Planer, der nur Minuten zählt,
	// reproduziert genau die Schieflage, die er beheben soll.
	HeadLoadMinutes int
}

// DefaultLimits sind die Startwerte. Sie sind geraten und gehören kalibriert,
// sobald echte Haushalte die App benutzen.
func DefaultLimits() Limits {
	return Limits{
		MaxOrgTasks:          3,
		MaxTasksPerMemberDay: 4,
		MaxHeadLoadPerDay:    4,
		MaxMinutesForChild:   45,
		MaxOvershootPermille: 25,
		HeadLoadMinutes:      15,
	}
}

// Result ist der fertige Wochenplan.
type Result struct {
	Week    Week
	Tasks   []PlannedTask
	Balance []MemberLoad
	// Struck sind die Aufgaben, die jemand für diese Woche gestrichen hat.
	//
	// Sie stehen hier und nicht in Tasks: Aus dem Plan sind sie raus, aus der
	// Bilanz auch. Aber sie ganz zu verschweigen hieße, dass eine Zeile
	// wortlos verschwindet — und wer sich vertippt hat, fände keinen Weg
	// zurück. Die Oberfläche zeigt sie leise am Ende des Tages, mit
	// „Doch wieder einplanen" daneben.
	Struck []StruckTask

	// Open sind die Fakten, die der Haushalt HEUTE nicht beantwortet hat —
	// die Grundlage der Fragen in der App.
	//
	// Getrennt von Skipped, weil die beiden verschiedene Zeitpunkte meinen:
	// Skipped erklärt die Woche, wie sie festgeschrieben wurde; Open
	// beschreibt den Haushalt, wie er jetzt ist. Solange beides dasselbe war,
	// fiel der Unterschied nicht auf — bis jemand eine Frage beantwortete und
	// sie nach dem Neuladen wiederkam.
	Open []string

	// Skipped erklärt, was nicht im Plan steht und warum. Ohne dieses Feld ist
	// ein Planer nicht zu debuggen — und die App könnte nicht begründen,
	// warum etwas fehlt.
	Skipped []Skipped
}

// StruckTask ist eine Aufgabe, die für diese Woche gestrichen wurde.
//
// Kennung, Titel und Tag — mehr nicht. Der Tag muss dabei sein, weil
// gestrichen wird, was an EINEM Tag ansteht: „Safiya zur Kita bringen"
// kommt fünfmal vor, und welches der fünf gestrichen ist, sieht man nur am
// Datum. Alles Weitere wäre eine Einladung, sie doch wieder wie eine Aufgabe
// zu behandeln.
type StruckTask struct {
	ID    string
	Title string
	Day   Date
}

// PlannedTask ist eine eingeplante Aufgabe an einem konkreten Tag.
type PlannedTask struct {
	// ID ist die Kennung der festgeschriebenen Aufgabe. Leer, solange die
	// Woche nur gerechnet und nicht geschrieben wurde — etwa beim CLI oder in
	// den Tests. Der Planer selbst vergibt sie nie: Er rechnet, er speichert
	// nicht.
	ID string

	// Done ist wahr, wenn jemand die Aufgabe abgehakt hat. Auch das kommt von
	// außen; für die Rechnung spielt es keine Rolle, ob eine Aufgabe dieser
	// Woche schon erledigt ist.
	Done bool

	TemplateID  string
	Title       string
	Category    Category
	Kind        Kind
	Day         Date
	Slot        Slot
	DurationMin int
	HeadLoad    HeadLoad
	AssigneeID  string
	Failure     Failure
	// Deadline ist gesetzt, wenn die Aufgabe eine echte Frist hat.
	Deadline Date
	// Reason ist maschinenlesbar. Das Sprachmodell formuliert daraus einen
	// Satz für die Oberfläche — es erfindet keine Begründung, es übersetzt
	// eine.
	Reason Reason
}

// Weight ist die Last der Aufgabe in gewichteten Minuten.
func (t PlannedTask) Weight(l Limits) int {
	return t.DurationMin + int(t.HeadLoad)*l.HeadLoadMinutes
}

// ReasonCode benennt die Regel, die zur Zuteilung geführt hat.
type ReasonCode string

const (
	ReasonRotation ReasonCode = "rotation"  // war zuletzt bei jemand anderem
	ReasonBalance  ReasonCode = "ausgleich" // hatte die geringste Last
	ReasonFixed    ReasonCode = "feste_person"
	ReasonOnlyOne  ReasonCode = "einzige_moeglichkeit"
	ReasonDeadline ReasonCode = "frist"
	ReasonOwn      ReasonCode = "eigene_aufgabe" // gehört dieser Person selbst
	ReasonManual   ReasonCode = "von_hand"       // ein Mensch hat umverteilt
	ReasonAgreed   ReasonCode = "absprache"      // steht so im Wochenraster
)

// AllReasonCodes ist die vollständige Liste, aus demselben Grund wie
// AllSkipCodes: Der Vertrag führt dieselben Wörter noch einmal, und zwei
// Listen von Hand laufen auseinander.
var AllReasonCodes = []ReasonCode{
	ReasonRotation, ReasonBalance, ReasonFixed, ReasonOnlyOne,
	ReasonDeadline, ReasonOwn, ReasonManual, ReasonAgreed,
}

type Reason struct {
	Code ReasonCode
	// Previous ist bei ReasonRotation die Person, die die Aufgabe zuletzt hatte.
	Previous string
}

// MemberLoad ist die Wochenlast einer Person, getrennt nach den beiden Achsen.
type MemberLoad struct {
	MemberID string
	Minutes  int
	HeadLoad int
	// Weighted ist Minuten + Kopflast × HeadLoadMinutes. Der Wert, nach dem
	// verteilt wird.
	Weighted int
	Tasks    int
	// Capacity sind die für Haushaltsaufgaben verfügbaren Minuten der ganzen
	// Woche. Ohne diese Zahl ist Weighted nicht zu deuten: 280 Minuten sind
	// für einen Erwachsenen die halbe Woche und für eine Dreizehnjährige
	// alles, was sie hat.
	Capacity int
}

// Utilization ist der Anteil der verfügbaren Zeit, der verplant ist, in
// Prozent.
//
// Gerechnet wird mit Minutes, nicht mit Weighted: Kopflast kostet Aufmerksamkeit,
// aber keine Uhrzeit, und wird deshalb auch nicht gegen die Kapazität
// verrechnet. Eine Zahl mit Kopflast durch eine Zahl ohne zu teilen ergibt
// Werte über 100 Prozent und damit eine Prozentangabe, die keine ist.
//
// Verteilt wird trotzdem nach der gewichteten Last im Verhältnis zur
// Kapazität — das ist eine Vergleichsgröße zwischen Personen und keine
// Prozentangabe. Sie bleibt deshalb im Paket.
func (l MemberLoad) Utilization() int {
	if l.Capacity <= 0 {
		return 0
	}
	return l.Minutes * 100 / l.Capacity
}

// SkipCode sagt, warum eine Vorlage nicht im Plan steht.
type SkipCode string

const (
	SkipNotApplicable  SkipCode = "gilt_nicht"        // Bedingungen nicht erfüllt
	SkipMuted          SkipCode = "abgewaehlt"        // Haushalt hat sie wiederholt gelöscht
	SkipNotDue         SkipCode = "nicht_faellig"     // diese Woche nicht dran
	SkipDensity        SkipCode = "startdichte"       // bewusst zurückgehalten
	SkipNoCapacity     SkipCode = "keine_kapazitaet"  // niemand hat Zeit
	SkipNoOneEligible  SkipCode = "niemand_geeignet"  // Alters- oder Rollenregel
	SkipUnknown        SkipCode = "unbekannt"         // eine Voraussetzung ist ungeklärt
	SkipNeedsEvent     SkipCode = "braucht_termin"    // entsteht nur aus einem Anlass
	SkipNeedsAgreement SkipCode = "braucht_absprache" // wer wann, ist nicht abgesprochen
)

// AllSkipCodes ist die vollständige Liste — dieselbe Überlegung wie bei
// AllCategories, nur teurer erkauft.
//
// Dieses Vokabular steht an drei Stellen: hier, und im Vertrag gleich zweimal
// (VorlagenStand.grund und Uebersprungen.grund). Die zweite war monatelang
// unvollständig, ohne dass es auffiel: Sie kannte `unbekannt` und
// `braucht_termin` nicht, obwohl der Planer beide meldet. Der Vertrag
// behauptete also etwas über die Antwort, was nicht stimmte, und der Fehler
// tauchte erst auf, als die Oberfläche einen der Werte zum ersten Mal
// vergleichen wollte.
//
// Ein Test in httpapi hält die drei Listen jetzt zusammen.
var AllSkipCodes = []SkipCode{
	SkipNotApplicable, SkipMuted, SkipNotDue, SkipDensity, SkipNoCapacity,
	SkipNoOneEligible, SkipUnknown, SkipNeedsEvent, SkipNeedsAgreement,
}

type Skipped struct {
	TemplateID string
	Title      string
	Code       SkipCode

	// Fact ist bei SkipUnknown das Faktum, das fehlt. Daraus entstehen die
	// Fragen, die die App stellt — höchstens zwei pro Woche, jede mit dem
	// Nutzen daneben.
	Fact string
}

// Plan berechnet den Wochenplan.
//
// Die Funktion ist rein: Sie liest nichts, schreibt nichts, fragt keine Uhr und
// würfelt nicht. Zwei Aufrufe mit derselben Eingabe liefern dasselbe Ergebnis,
// Feld für Feld und in derselben Reihenfolge.
//
// Der Ablauf in fünf Schritten, jeder in einer eigenen Datei:
//
//	auswaehlen  – welche Vorlagen gelten für diesen Haushalt   (select.go)
//	faellig     – welche davon sind diese Woche dran           (due.go)
//	verdichten  – wie viele davon zeigen wir wirklich          (density.go)
//	zuteilen    – wer macht was an welchem Tag                 (assign.go)
//	ausgleichen – tauschen, solange es gleichmäßiger wird      (rebalance.go)
func Plan(in Input) (Result, error) {
	if err := in.validate(); err != nil {
		return Result{}, err
	}

	// Die Dauern werden einmal auf diesen Haushalt umgerechnet, bevor
	// irgendetwas damit rechnet: Zuteilung, Kapazitätsprüfung und Anzeige
	// sollen dieselbe Zahl sehen (siehe ScaleTemplates).
	in.Templates = ScaleTemplates(in.Templates, in.Household)

	var skipped []Skipped

	applicable, s1 := selectApplicable(in)
	skipped = append(skipped, s1...)

	due, s2 := selectDue(in, applicable)
	skipped = append(skipped, s2...)

	kept, s3 := capDensity(in, due)
	skipped = append(skipped, s3...)

	tasks, s4 := assign(in, kept)
	skipped = append(skipped, s4...)

	tasks = rebalance(in, tasks)

	sortTasks(tasks)
	sortSkipped(skipped)

	return Result{
		Week:    in.Week,
		Tasks:   tasks,
		Balance: balance(in, tasks),
		Skipped: skipped,
	}, nil
}

func (in Input) validate() error {
	if len(in.Household.Members) == 0 {
		return errors.New("planner: der Haushalt hat keine Mitglieder")
	}
	if in.Week.Week < 1 || in.Week.Week > 53 {
		return fmt.Errorf("planner: %s ist keine gültige Kalenderwoche", in.Week)
	}
	seen := make(map[string]bool, len(in.Household.Members))
	for _, m := range in.Household.Members {
		if m.ID == "" {
			return errors.New("planner: ein Mitglied ohne ID")
		}
		if seen[m.ID] {
			return fmt.Errorf("planner: die Mitglieds-ID %q kommt doppelt vor", m.ID)
		}
		seen[m.ID] = true
	}
	ids := make(map[string]bool, len(in.Templates))
	for _, t := range in.Templates {
		if t.ID == "" {
			return errors.New("planner: eine Vorlage ohne ID")
		}
		if ids[t.ID] {
			return fmt.Errorf("planner: die Vorlagen-ID %q kommt doppelt vor", t.ID)
		}
		ids[t.ID] = true
	}
	if in.Limits == (Limits{}) {
		return errors.New("planner: Limits fehlen, nutze DefaultLimits()")
	}
	return nil
}

// BalanceOf ist die Bilanz zu einem beliebigen Satz Aufgaben.
//
// Nötig für die festgeschriebene Woche: Die wird aus der Datenbank gelesen,
// nicht gerechnet, und hat deshalb keinen Input. Die Bilanz selbst wird
// trotzdem gerechnet und nie gespeichert — eine Summe neben den Posten läuft
// irgendwann auseinander.
func BalanceOf(h Household, tasks []PlannedTask, l Limits) []MemberLoad {
	return balance(Input{Household: h, Limits: l}, tasks)
}

// balance summiert die Last je Person. Personen ohne Aufgaben erscheinen mit
// Null — sonst sieht eine leere Woche aus wie eine faire.
func balance(in Input, tasks []PlannedTask) []MemberLoad {
	byID := map[string]*MemberLoad{}
	var order []string
	for _, m := range in.Household.Members {
		if !m.CanPerform() {
			continue
		}
		byID[m.ID] = &MemberLoad{MemberID: m.ID, Capacity: weeklyCapacity(m)}
		order = append(order, m.ID)
	}
	for _, t := range tasks {
		l, ok := byID[t.AssigneeID]
		if !ok {
			continue
		}
		l.Minutes += t.DurationMin
		l.HeadLoad += int(t.HeadLoad)
		l.Weighted += t.Weight(in.Limits)
		l.Tasks++
	}
	out := make([]MemberLoad, 0, len(order))
	for _, id := range order {
		out = append(out, *byID[id])
	}
	return out
}

// weeklyCapacity summiert die verfügbaren Minuten der Woche.
func weeklyCapacity(m Member) int {
	total := 0
	for _, min := range m.CapacityMinutes {
		total += min
	}
	return total
}

func sortTasks(tasks []PlannedTask) {
	sort.SliceStable(tasks, func(i, j int) bool {
		a, b := tasks[i], tasks[j]
		if a.Day != b.Day {
			return a.Day.Before(b.Day)
		}
		if a.Slot != b.Slot {
			return slotOrder(a.Slot) < slotOrder(b.Slot)
		}
		if a.AssigneeID != b.AssigneeID {
			return a.AssigneeID < b.AssigneeID
		}
		return a.TemplateID < b.TemplateID
	})
}

func sortSkipped(s []Skipped) {
	sort.SliceStable(s, func(i, j int) bool {
		if s[i].Code != s[j].Code {
			return s[i].Code < s[j].Code
		}
		return s[i].TemplateID < s[j].TemplateID
	})
}

func slotOrder(s Slot) int {
	switch s {
	case SlotMorning:
		return 0
	case SlotAny:
		return 1
	case SlotEvening:
		return 2
	}
	return 3
}

func sortMembersByAge(ms []Member) {
	sort.SliceStable(ms, func(i, j int) bool {
		if ms[i].Age != ms[j].Age {
			return ms[i].Age < ms[j].Age
		}
		return ms[i].ID < ms[j].ID
	})
}
