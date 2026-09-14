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

	// BirthYear ist das, woraus Age gerechnet wurde — 0 heißt „nicht
	// gefragt". Der Planer liest es nie; er rechnet mit Age. Es steht hier,
	// damit die Einstellungen zeigen können, was tatsächlich gespeichert ist,
	// statt aus dem Alter ein Jahr zurückzurechnen und dabei jeden zu
	// verjüngen, der dieses Jahr noch Geburtstag hat.
	BirthYear int

	// HasAccess sagt, ob diese Person sich anmelden kann. Der Planer schaut
	// nie hin — für ihn ist jede Person gleich, ob sie die App je geöffnet
	// hat oder nicht. Die Oberfläche braucht es: Sie entscheidet, wen man noch
	// einladen kann, und macht sichtbar, wer bisher nur im Plan steht.
	HasAccess bool

	// CapacityMinutes sind die für Haushaltsaufgaben verfügbaren Minuten je
	// Wochentag, Index 0 = Montag. Aus Arbeitszeiten und Betreuungszeiten
	// abgeleitet, bewusst grob: Der Plan soll machbar wirken, nicht exakt sein.
	CapacityMinutes [7]int
}

// IsAdult entscheidet über Aufgaben, die nur Erwachsene übernehmen dürfen.
//
// Die Rolle zählt mit, nicht nur das Alter: RolePlanner ist per Definition
// erwachsen (siehe dort — „plant und führt aus: die Erwachsenen"). Das ist
// nicht Bequemlichkeit, sondern nötig, damit ein frisch registrierter Mensch
// einen brauchbaren Plan bekommt, bevor er sein Geburtsjahr eingetragen hat.
// Ohne diese Zeile wäre er mit Alter 0 ein Kind und bekäme keine einzige
// Aufgabe, die Erwachsene voraussetzt.
//
// Ein unbekanntes Alter heißt ebenfalls erwachsen, und das ist die
// unangenehmere Hälfte der Regel. Der Grund steht im Beitrittspfad: Wer einer
// Einladung als ausführende Person folgt, bekommt kein Geburtsjahr — wir
// fragen es nicht ab und erfinden es nicht. Mit der alten Regel war dieser
// Mensch mit Alter 0 ein Kind: Er bekam keine Aufgabe, die Erwachsene
// voraussetzt, und zählte obendrein als Kind des Haushalts, was darüber
// entscheidet, welche Vorlagen überhaupt fällig werden. Genau der Fall aus dem
// Produktkonzept — der planungsunwillige Partner — war damit falsch geplant.
//
// Die Gegenrichtung kostet weniger: Kinder bekommen im Onboarding immer ein
// Geburtsjahr, betreute Personen müssen eines haben. Ein Kind ohne Jahr kann
// also nur über einen Import ohne Altersangabe entstehen — und dort ist die
// fehlende Angabe der Fehler, nicht diese Zeile.
func (m Member) IsAdult() bool {
	switch {
	case m.Role == RolePlanner:
		return true
	case m.Age <= 0:
		return true
	default:
		return m.Age >= 18
	}
}

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

	// Occasions sind die eingetragenen Anlässe: Geburtstage, Elternabend,
	// Arzttermine. Sie gehören zum Haushalt wie seine Mitglieder — und ohne
	// sie entstehen anlassgebundene Aufgaben nicht. Die App erfindet keinen
	// Geburtstag (ADR-0010).
	Occasions []Occasion
}

// Context sind die Bedingungen, an denen Vorlagen hängen — Wohnform, Auto,
// Garten, Haustiere. Was hier fehlt, erzeugt keine Aufgaben.
type Context struct {
	Home    Home
	HasCar  bool
	HasYard bool
	Pets    []string // "hund", "katze", …

	// Rooms und Baths sind die Größe des Haushalts. Sie filtern nichts, sie
	// skalieren: Putzaufgaben dauern in fünf Zimmern länger als in zwei, und
	// wer zwei Bäder hat, putzt zwei.
	//
	// Anders als die Wohnform, die abgefragt wurde und nichts bewirkte, ändern
	// diese beiden Zahlen die Dauern — und die Dauer trägt die Verteilung.
	Rooms int
	Baths int

	// Facts ist alles, was ein Haushalt haben kann und wonach beim Einrichten
	// niemand gefragt hat: Pflanzen, Spülmaschine, Keller, Fahrrad.
	//
	// Drei Zustände, nicht zwei. Nicht enthalten heißt **unbekannt**, und
	// unbekannt ist nicht dasselbe wie nein: Eine Aufgabe, die ein unbekanntes
	// Faktum voraussetzt, wird nicht eingeplant — aber sie wird zur Frage.
	//
	// Genau diese Unterscheidung fehlte. Vorher galt jede Bedingung als
	// erfüllt, und die App plante Pflanzen gießen in einen Haushalt ohne
	// Pflanzen. Was die App nicht weiß, darf sie nicht behaupten.
	Facts map[string]bool
}

// Fact liefert den Wert eines Faktums und ob es überhaupt bekannt ist.
func (c Context) Fact(name string) (wert bool, bekannt bool) {
	v, ok := c.Facts[name]
	return v, ok
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

// AllCategories ist die vollständige Liste. Sie steht hier und nicht in der
// Oberfläche: Eine zweite Liste in TypeScript wäre beim ersten neuen Bereich
// unvollständig — genau das ist auf der Seite „Eure Woche" schon passiert.
var AllCategories = []Category{
	CatKitchen, CatLaundry, CatCleaning, CatChild, CatSupplies,
	CatAppointment, CatAdmin, CatMaintain, CatSocial, CatOutdoor,
}

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

	// RequiresFacts sind Fakten, die wahr sein müssen. Ist eines davon
	// unbekannt, wird die Vorlage nicht eingeplant und stattdessen zur Frage.
	RequiresFacts []string

	// RequiresOccasion nennt die Art des Anlasses, an dem diese Aufgabe hängt
	// — "kindergeburtstag", "elternabend". Leer bei allen anderen.
	RequiresOccasion string

	// RequiresEvent heißt: Diese Aufgabe entsteht aus einem Anlass, nicht aus
	// einem Zeitraum — ein Geburtstag, ein Elternabend, ein Termin.
	//
	// Solange es in der App keine Termine gibt, wird sie nie eingeplant. Das
	// ist Absicht und die Lehre aus „Geschenk für Kindergeburtstag besorgen":
	// Die Vorlage trug den Rhythmus „Auslöser, alle 45 Tage" und erfand damit
	// alle anderthalb Monate einen Geburtstag. Eine Aufgabe ohne Anlass ist
	// keine Erinnerung, sondern eine Behauptung.
	RequiresEvent bool
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

	// ScalesWith sagt, ob die Dauer mit der Größe des Haushalts wächst.
	// Leer heißt: feste Dauer. Müll rausbringen dauert in zwei wie in fünf
	// Zimmern gleich lang.
	ScalesWith Scale

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
