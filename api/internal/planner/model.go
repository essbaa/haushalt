// Package planner berechnet aus einem Haushalt und einer Vorlagen-Bibliothek
// den Wochenplan.
//
// Das Paket ist rein: keine Datenbank, kein Netz, keine Uhr, keine
// Zufallszahlen. Es importiert ausschließlich die Standardbibliothek.
// Dieselbe Eingabe ergibt immer dieselbe Ausgabe — darauf verlassen sich die
// Tests, der Aufrufer und am Ende der Nutzer, der nicht jeden Montag einen
// anderen Plan sehen will.
//
// Alles, was mit der Außenwelt spricht — YAML lesen, Datenbank, Sprachmodell —
// liegt außerhalb. Das Sprachmodell übersetzt Freitext in diese Typen, bevor
// geplant wird, und formuliert aus Reason einen Satz, nachdem geplant wurde.
// Es berührt die Berechnung nicht.
package planner

import "time"

// ---------------------------------------------------------------- Haushalt

// Role beschreibt, welchen Platz eine Person im Haushalt einnimmt.
type Role string

const (
	// RolePlanner plant und führt aus: die Erwachsenen.
	RolePlanner Role = "planend"
	// RoleDoer führt aus, plant aber nicht: Kinder und Jugendliche, oder der
	// Partner, der mit Planung nichts zu tun haben will.
	RoleDoer Role = "ausfuehrend"
	// RoleDependent löst Aufgaben aus, führt selbst keine aus: das Kleinkind.
	// Ohne diese Rolle ließe sich nicht ausdrücken, dass ein Zweijähriger
	// Arbeit erzeugt, ohne welche übernehmen zu können.
	RoleDependent Role = "betreut"
)

// Care sagt, ob und wie ein Kind betreut wird. Steuert einen großen Teil der
// Organisationsaufgaben.
type Care string

const (
	CareNone   Care = "keine"
	CareKita   Care = "kita"
	CareSchool Care = "schule"
)

// Member ist eine Person im Haushalt.
type Member struct {
	ID   string
	Name string
	Role Role

	// Age in Jahren. Bei Erwachsenen darf der Wert grob sein; er entscheidet
	// nur über Altersgrenzen von Aufgaben.
	Age int

	// Care ist nur bei Kindern gesetzt.
	Care Care

	// CapacityMinutes sind die für Haushaltsaufgaben verfügbaren Minuten je
	// Wochentag, Index 0 = Montag. Aus Arbeitszeiten und Betreuungszeiten
	// abgeleitet, bewusst grob: Der Plan soll machbar wirken, nicht exakt sein.
	CapacityMinutes [7]int
}

// IsAdult entscheidet über Aufgaben, die nur Erwachsene übernehmen dürfen.
func (m Member) IsAdult() bool { return m.Age >= 18 }

// CanPerform sagt, ob diese Person überhaupt Aufgaben ausführt.
func (m Member) CanPerform() bool {
	return m.Role == RolePlanner || m.Role == RoleDoer
}

// Household ist der Mandant: alles hängt daran.
type Household struct {
	ID      string
	Name    string
	Members []Member
	Context Context
}

// Context sind die Bedingungen, an denen Vorlagen hängen — Wohnform, Auto,
// Garten, Haustiere. Was hier fehlt, erzeugt keine Aufgaben.
type Context struct {
	Home    Home
	HasCar  bool
	HasYard bool
	Pets    []string // "hund", "katze", …
}

type Home string

const (
	HomeFlat  Home = "wohnung"
	HomeHouse Home = "haus"
)

// Children liefert die Kinder des Haushalts, jüngstes zuerst.
func (h Household) Children() []Member {
	var out []Member
	for _, m := range h.Members {
		if !m.IsAdult() {
			out = append(out, m)
		}
	}
	sortMembersByAge(out)
	return out
}

// ---------------------------------------------------------------- Vorlagen

// Kind trennt Handarbeit von Kopfarbeit. Die wichtigste Unterscheidung im
// ganzen Modell: Danach wird die Startdichte begrenzt und die Rotation
// gewichtet.
type Kind string

const (
	KindDo  Kind = "ausfuehrung"
	KindOrg Kind = "organisation"
)

// HeadLoad ist der Planungs- und Erinnerungsaufwand einer Aufgabe, unabhängig
// von ihrer Dauer. 0 = einfach tun, 3 = im Kopf behalten, entscheiden,
// koordinieren.
type HeadLoad int

const (
	HeadLoadNone HeadLoad = 0
	HeadLoadLow  HeadLoad = 1
	HeadLoadMid  HeadLoad = 2
	HeadLoadHigh HeadLoad = 3
)

type Category string

const (
	CatKitchen     Category = "kueche"
	CatLaundry     Category = "waesche"
	CatCleaning    Category = "reinigung"
	CatChild       Category = "kind"
	CatSupplies    Category = "vorrat"
	CatAppointment Category = "termine"
	CatAdmin       Category = "verwaltung"
	CatMaintain    Category = "wartung"
	CatSocial      Category = "sozial"
	CatOutdoor     Category = "aussen"
)

// Slot ist das bevorzugte Zeitfenster innerhalb eines Tages.
type Slot string

const (
	SlotAny     Slot = "egal"
	SlotMorning Slot = "morgens"
	SlotEvening Slot = "abends"
)

// Distribution sagt, wie eine Aufgabe verteilt werden darf.
type Distribution string

const (
	DistRotate     Distribution = "rotiert"
	DistFixed      Distribution = "feste_person"
	DistAdultsOnly Distribution = "nur_erwachsene"
	// DistChildrenOnly ist das Gegenstück zu DistAdultsOnly: Aufgaben, die
	// ausdrücklich den Kindern gehören. Ohne diesen Wert landet „eigenes
	// Zimmer aufräumen" bei einem Achtundvierzigjährigen, sobald seine
	// Auslastung gerade die niedrigste ist.
	DistChildrenOnly Distribution = "nur_kinder"
)

// Failure beschreibt, was passiert, wenn die Aufgabe liegen bleibt. Steuert,
// ob sie verschoben werden darf oder mit Vorrang eingeplant wird.
type Failure string

const (
	// FailureSoft: Es passiert nichts, das Bad ist eine Woche länger unsauber.
	FailureSoft Failure = "weich"
	// FailureHard: Die Frist verstreicht, der Platz ist weg, die Mahnung kommt.
	FailureHard Failure = "hart"
)

// Source hält fest, woher eine Vorlage stammt. Kuratiertes wird anders
// gewichtet als das, was ein Haushalt selbst angelegt hat.
type Source string

const (
	SourceCurated   Source = "kuratiert"
	SourceHousehold Source = "haushalt"
	SourceLearned   Source = "gelernt"
)

// RhythmType ist einer der fünf Wiederholungstypen. "Täglich oder wöchentlich"
// reicht für einen echten Haushalt nicht.
type RhythmType string

const (
	// RhythmFixed liegt auf festen Wochentagen: Müll dienstags.
	RhythmFixed RhythmType = "fest"
	// RhythmWindow muss in einem Zeitraum passieren, egal wann: Bad putzen.
	RhythmWindow RhythmType = "fenster"
	// RhythmTrigger startet an einem Zustand: Wäschekorb voll. Bis Signale es
	// besser wissen, wird der Abstand geschätzt.
	RhythmTrigger RhythmType = "ausloeser"
	// RhythmSeason kehrt jährlich wieder, mit Vorlauf: Ferienbetreuung.
	RhythmSeason RhythmType = "saison"
	// RhythmPhase kommt und geht mit dem Alter des Kindes. Die Kadenz
	// entspricht RhythmWindow; die Phase selbst steht in Conditions.
	RhythmPhase RhythmType = "phase"
)

// Rhythm bündelt Typ und Parameter. Nicht jedes Feld gilt für jeden Typ; was
// nicht passt, bleibt der Nullwert.
type Rhythm struct {
	Type RhythmType

	// Weekdays gilt für RhythmFixed.
	Weekdays []time.Weekday

	// EveryDays gilt für RhythmWindow, RhythmTrigger und RhythmPhase:
	// gewünschter Abstand zwischen zwei Erledigungen.
	EveryDays int

	// Months gilt für RhythmSeason: die Monate, in denen die Aufgabe erledigt
	// sein muss.
	Months []time.Month
}

// AgeRange ist eine Altersspanne in Jahren. Max == 0 heißt "nach oben offen".
type AgeRange struct {
	Min int
	Max int
}

func (r AgeRange) Contains(age int) bool {
	if age < r.Min {
		return false
	}
	return r.Max == 0 || age <= r.Max
}

// Conditions sind die Bedingungen, unter denen eine Vorlage für einen Haushalt
// überhaupt gilt.
type Conditions struct {
	// RequiresChildAged ist nil, wenn kein Kind nötig ist.
	RequiresChildAged *AgeRange
	// RequiresCare ist leer, wenn die Betreuungsform egal ist.
	RequiresCare Care
	RequiresCar  bool
	RequiresYard bool
	RequiresPet  bool
	// RequiresPetKind grenzt auf eine Tierart ein ("hund", "katze"). Leer
	// heißt: irgendein Tier genügt. Ohne dieses Feld steht das Katzenklo im
	// Plan eines Haushalts mit Hund.
	RequiresPetKind string
	// RequiresHome ist leer, wenn die Wohnform egal ist.
	RequiresHome Home
}

// TaskTemplate ist eine Vorlage aus der Bibliothek — die vierzehn Felder aus
// dem Konzept. Vorlagen liegen als YAML im Repo und werden vor dem Planen
// eingelesen; der Planer selbst sieht nur diese Struktur.
type TaskTemplate struct {
	ID       string
	Title    string
	Category Category
	Kind     Kind

	DurationMin int
	HeadLoad    HeadLoad

	Rhythm Rhythm
	// LeadDays ist der Vorlauf: wie viele Tage vorher die Aufgabe auftauchen
	// muss. Bei der Ferienbetreuung ist der Vorlauf die ganze Leistung.
	LeadDays int
	Slot     Slot

	// MinAge gilt für Aufgaben, die auch Kinder übernehmen dürfen.
	MinAge       int
	Distribution Distribution

	// PerPerson macht aus einem Termin eine Aufgabe je berechtigter Person,
	// fest an sie gebunden. Für alles, was jedem selbst gehört: das eigene
	// Zimmer, die eigene Wäsche, die eigene Schultasche. Rotation ergibt dort
	// keinen Sinn — niemand räumt abwechselnd das Zimmer eines anderen auf.
	PerPerson bool

	AppliesTo Conditions

	// ChainNext sind Folgeaufgaben: waschen → aufhängen → zusammenlegen.
	ChainNext []string

	Failure Failure
	Source  Source
}

// ---------------------------------------------------------------- Historie

// History ist das, was der Planer über die Vergangenheit wissen muss. Sie wird
// vom Aufrufer aus dem Ereignisprotokoll verdichtet — der Planer liest keine
// Datenbank.
type History struct {
	// LastDone ist das Datum der letzten Erledigung je Vorlage.
	LastDone map[string]Date
	// LastAssignee ist die Person, die eine Vorlage zuletzt hatte. Grundlage
	// der Rotation.
	LastAssignee map[string]string
	// Muted sind Vorlagen, die der Haushalt wiederholt gelöscht hat. Sie
	// tauchen nicht mehr auf, bis jemand sie wieder aktiviert.
	Muted map[string]bool
	// FixedTo bindet eine Vorlage dauerhaft an eine Person
	// (Distribution == DistFixed).
	FixedTo map[string]string
}

func (h History) lastDone(templateID string) Date  { return h.LastDone[templateID] }
func (h History) lastAssignee(id string) string    { return h.LastAssignee[id] }
func (h History) isMuted(templateID string) bool   { return h.Muted[templateID] }
func (h History) fixedTo(templateID string) string { return h.FixedTo[templateID] }
