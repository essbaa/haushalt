// Package library liest Haushalte und Vorlagen aus Dateien und übersetzt sie
// in die Typen des Planers.
//
// Diese Trennung ist Absicht: Alles, was mit Dateien, Formaten und fehlerhaften
// Eingaben zu tun hat, liegt hier. Das Paket planner bekommt nur noch geprüfte
// Strukturen und bleibt damit rein.
//
// Format: Im Durchstich JSON, weil es ohne Abhängigkeit auskommt. Der Wechsel
// auf YAML ist zwei Zeilen — die Struktur-Tags stehen schon dran:
//
//	import "gopkg.in/yaml.v3"
//	yaml.Unmarshal(raw, &doc)   statt   json.Unmarshal(raw, &doc)
package library

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/zakaria/haushalt/api/internal/planner"
)

// ------------------------------------------------------------------ Vorlagen

type templateDoc struct {
	Templates []templateDTO `json:"vorlagen" yaml:"vorlagen"`
}

type templateDTO struct {
	ID          string       `json:"id" yaml:"id"`
	Title       string       `json:"titel" yaml:"titel"`
	Category    string       `json:"kategorie" yaml:"kategorie"`
	Kind        string       `json:"art" yaml:"art"`
	DurationMin int          `json:"dauer_min" yaml:"dauer_min"`
	HeadLoad    int          `json:"kopflast" yaml:"kopflast"`
	Rhythm      rhythmDTO    `json:"rhythmus" yaml:"rhythmus"`
	LeadDays    int          `json:"vorlauf_tage" yaml:"vorlauf_tage"`
	Slot        string       `json:"zeitfenster" yaml:"zeitfenster"`
	MinAge      int          `json:"ab_alter" yaml:"ab_alter"`
	Dist        string       `json:"verteilung" yaml:"verteilung"`
	PerPerson   bool         `json:"je_person" yaml:"je_person"`
	AppliesTo   conditionDTO `json:"gilt_fuer" yaml:"gilt_fuer"`
	ChainNext   []string     `json:"kette" yaml:"kette"`
	Failure     string       `json:"ausfall" yaml:"ausfall"`
	Source      string       `json:"quelle" yaml:"quelle"`
}

type rhythmDTO struct {
	Type      string   `json:"typ" yaml:"typ"`
	Weekdays  []string `json:"wochentage" yaml:"wochentage"`
	EveryDays int      `json:"alle_tage" yaml:"alle_tage"`
	Months    []int    `json:"monate" yaml:"monate"`
}

type conditionDTO struct {
	ChildAgeMin *int   `json:"kind_alter_min" yaml:"kind_alter_min"`
	ChildAgeMax *int   `json:"kind_alter_max" yaml:"kind_alter_max"`
	Care        string `json:"betreuung" yaml:"betreuung"`
	Car         bool   `json:"auto" yaml:"auto"`
	Yard        bool   `json:"garten" yaml:"garten"`
	Pet         bool   `json:"haustier" yaml:"haustier"`
	PetKind     string `json:"haustier_art" yaml:"haustier_art"`
	Home        string `json:"wohnform" yaml:"wohnform"`
}

// LoadTemplates liest die Bibliothek. Fehler nennen immer die betroffene
// Vorlage — die CI meldet damit nicht "ungültige Datei", sondern welche Zeile
// jemand kaputtgemacht hat.
func LoadTemplates(path string) ([]planner.TaskTemplate, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var doc templateDoc
	if err := json.Unmarshal(raw, &doc); err != nil {
		return nil, fmt.Errorf("%s: %w", path, err)
	}
	if len(doc.Templates) == 0 {
		return nil, fmt.Errorf("%s: enthält keine Vorlagen", path)
	}

	seen := map[string]bool{}
	out := make([]planner.TaskTemplate, 0, len(doc.Templates))
	for i, d := range doc.Templates {
		t, err := d.toTemplate()
		if err != nil {
			return nil, fmt.Errorf("%s, Vorlage %d (%q): %w", path, i+1, d.ID, err)
		}
		if seen[t.ID] {
			return nil, fmt.Errorf("%s: die ID %q kommt doppelt vor", path, t.ID)
		}
		seen[t.ID] = true
		out = append(out, t)
	}
	return out, nil
}

func (d templateDTO) toTemplate() (planner.TaskTemplate, error) {
	var t planner.TaskTemplate

	if d.ID == "" || d.Title == "" {
		return t, fmt.Errorf("id und titel sind Pflicht")
	}
	if d.DurationMin <= 0 {
		return t, fmt.Errorf("dauer_min muss größer als 0 sein")
	}
	if d.HeadLoad < 0 || d.HeadLoad > 3 {
		return t, fmt.Errorf("kopflast %d liegt außerhalb von 0..3", d.HeadLoad)
	}

	kind, err := parseKind(d.Kind)
	if err != nil {
		return t, err
	}
	rhythm, err := parseRhythm(d.Rhythm)
	if err != nil {
		return t, err
	}
	dist, err := parseDistribution(d.Dist)
	if err != nil {
		return t, err
	}
	cond, err := d.AppliesTo.toConditions()
	if err != nil {
		return t, err
	}

	return planner.TaskTemplate{
		ID:           d.ID,
		Title:        d.Title,
		Category:     planner.Category(d.Category),
		Kind:         kind,
		DurationMin:  d.DurationMin,
		HeadLoad:     planner.HeadLoad(d.HeadLoad),
		Rhythm:       rhythm,
		LeadDays:     d.LeadDays,
		Slot:         parseSlot(d.Slot),
		MinAge:       d.MinAge,
		Distribution: dist,
		PerPerson:    d.PerPerson,
		AppliesTo:    cond,
		ChainNext:    d.ChainNext,
		Failure:      parseFailure(d.Failure),
		Source:       parseSource(d.Source),
	}, nil
}

func (c conditionDTO) toConditions() (planner.Conditions, error) {
	out := planner.Conditions{
		RequiresCar:     c.Car,
		RequiresYard:    c.Yard,
		RequiresPet:     c.Pet,
		RequiresPetKind: c.PetKind,
	}
	if c.Care != "" {
		care, err := parseCare(c.Care)
		if err != nil {
			return out, err
		}
		out.RequiresCare = care
	}
	if c.Home != "" {
		switch c.Home {
		case "wohnung":
			out.RequiresHome = planner.HomeFlat
		case "haus":
			out.RequiresHome = planner.HomeHouse
		default:
			return out, fmt.Errorf("unbekannte wohnform %q", c.Home)
		}
	}
	if c.ChildAgeMin != nil || c.ChildAgeMax != nil {
		r := planner.AgeRange{}
		if c.ChildAgeMin != nil {
			r.Min = *c.ChildAgeMin
		}
		if c.ChildAgeMax != nil {
			r.Max = *c.ChildAgeMax
		}
		out.RequiresChildAged = &r
	}
	return out, nil
}

// ----------------------------------------------------------------- Haushalt

type householdDoc struct {
	ID      string      `json:"id" yaml:"id"`
	Name    string      `json:"name" yaml:"name"`
	Home    string      `json:"wohnform" yaml:"wohnform"`
	Car     bool        `json:"auto" yaml:"auto"`
	Yard    bool        `json:"garten" yaml:"garten"`
	Pets    []string    `json:"haustiere" yaml:"haustiere"`
	Members []memberDTO `json:"mitglieder" yaml:"mitglieder"`
	History historyDTO  `json:"historie" yaml:"historie"`
}

type memberDTO struct {
	ID       string `json:"id" yaml:"id"`
	Name     string `json:"name" yaml:"name"`
	Role     string `json:"rolle" yaml:"rolle"`
	Age      int    `json:"alter" yaml:"alter"`
	Care     string `json:"betreuung" yaml:"betreuung"`
	Capacity []int  `json:"kapazitaet_minuten" yaml:"kapazitaet_minuten"`
}

type historyDTO struct {
	LastDone     map[string]string `json:"zuletzt_erledigt" yaml:"zuletzt_erledigt"`
	LastAssignee map[string]string `json:"zuletzt_zugeteilt" yaml:"zuletzt_zugeteilt"`
	Muted        []string          `json:"abgewaehlt" yaml:"abgewaehlt"`
	FixedTo      map[string]string `json:"feste_person" yaml:"feste_person"`
}

// LoadHousehold liest einen Haushalt samt seiner verdichteten Historie.
func LoadHousehold(path string) (planner.Household, planner.History, error) {
	var h planner.Household
	var hist planner.History

	raw, err := os.ReadFile(path)
	if err != nil {
		return h, hist, err
	}
	var doc householdDoc
	if err := json.Unmarshal(raw, &doc); err != nil {
		return h, hist, fmt.Errorf("%s: %w", path, err)
	}
	if len(doc.Members) == 0 {
		return h, hist, fmt.Errorf("%s: der Haushalt hat keine Mitglieder", path)
	}

	h = planner.Household{
		ID:   doc.ID,
		Name: doc.Name,
		Context: planner.Context{
			Home:    planner.Home(doc.Home),
			HasCar:  doc.Car,
			HasYard: doc.Yard,
			Pets:    doc.Pets,
		},
	}
	for i, m := range doc.Members {
		member, err := m.toMember()
		if err != nil {
			return h, hist, fmt.Errorf("%s, Mitglied %d (%q): %w", path, i+1, m.ID, err)
		}
		h.Members = append(h.Members, member)
	}

	hist, err = doc.History.toHistory()
	if err != nil {
		return h, hist, fmt.Errorf("%s, historie: %w", path, err)
	}
	return h, hist, nil
}

func (m memberDTO) toMember() (planner.Member, error) {
	var out planner.Member
	if m.ID == "" {
		return out, fmt.Errorf("id ist Pflicht")
	}
	role, err := parseRole(m.Role)
	if err != nil {
		return out, err
	}
	out = planner.Member{ID: m.ID, Name: m.Name, Role: role, Age: m.Age}
	if m.Care != "" {
		care, err := parseCare(m.Care)
		if err != nil {
			return out, err
		}
		out.Care = care
	}
	switch len(m.Capacity) {
	case 0:
		// Betreute Personen brauchen keine Kapazität.
		if role != planner.RoleDependent {
			return out, fmt.Errorf("kapazitaet_minuten fehlt (7 Werte, Montag zuerst)")
		}
	case 7:
		copy(out.CapacityMinutes[:], m.Capacity)
	default:
		return out, fmt.Errorf("kapazitaet_minuten hat %d Werte, erwartet werden 7", len(m.Capacity))
	}
	return out, nil
}

func (d historyDTO) toHistory() (planner.History, error) {
	out := planner.History{
		LastDone:     map[string]planner.Date{},
		LastAssignee: map[string]string{},
		Muted:        map[string]bool{},
		FixedTo:      map[string]string{},
	}
	for id, s := range d.LastDone {
		t, err := time.Parse("2006-01-02", s)
		if err != nil {
			return out, fmt.Errorf("zuletzt_erledigt[%s] = %q ist kein Datum (2026-09-14)", id, s)
		}
		out.LastDone[id] = planner.DateOf(t)
	}
	for id, who := range d.LastAssignee {
		out.LastAssignee[id] = who
	}
	for _, id := range d.Muted {
		out.Muted[id] = true
	}
	for id, who := range d.FixedTo {
		out.FixedTo[id] = who
	}
	return out, nil
}

// -------------------------------------------------------------- Übersetzung

func parseKind(s string) (planner.Kind, error) {
	switch s {
	case "ausfuehrung", "ausführung":
		return planner.KindDo, nil
	case "organisation":
		return planner.KindOrg, nil
	}
	return "", fmt.Errorf("art %q ist weder ausfuehrung noch organisation", s)
}

func parseRole(s string) (planner.Role, error) {
	switch s {
	case "planend":
		return planner.RolePlanner, nil
	case "ausfuehrend", "ausführend":
		return planner.RoleDoer, nil
	case "betreut":
		return planner.RoleDependent, nil
	}
	return "", fmt.Errorf("rolle %q ist unbekannt (planend, ausfuehrend, betreut)", s)
}

func parseCare(s string) (planner.Care, error) {
	switch s {
	case "keine":
		return planner.CareNone, nil
	case "kita":
		return planner.CareKita, nil
	case "schule":
		return planner.CareSchool, nil
	}
	return "", fmt.Errorf("betreuung %q ist unbekannt (keine, kita, schule)", s)
}

func parseDistribution(s string) (planner.Distribution, error) {
	switch s {
	case "", "rotiert":
		return planner.DistRotate, nil
	case "feste_person":
		return planner.DistFixed, nil
	case "nur_erwachsene":
		return planner.DistAdultsOnly, nil
	case "nur_kinder":
		return planner.DistChildrenOnly, nil
	}
	return "", fmt.Errorf("verteilung %q ist unbekannt", s)
}

func parseRhythm(d rhythmDTO) (planner.Rhythm, error) {
	var r planner.Rhythm
	switch d.Type {
	case "fest":
		r.Type = planner.RhythmFixed
		if len(d.Weekdays) == 0 {
			return r, fmt.Errorf("rhythmus fest braucht wochentage")
		}
		for _, w := range d.Weekdays {
			wd, err := parseWeekday(w)
			if err != nil {
				return r, err
			}
			r.Weekdays = append(r.Weekdays, wd)
		}
	case "fenster", "ausloeser", "auslöser", "phase":
		switch d.Type {
		case "fenster":
			r.Type = planner.RhythmWindow
		case "phase":
			r.Type = planner.RhythmPhase
		default:
			r.Type = planner.RhythmTrigger
		}
		if d.EveryDays <= 0 {
			return r, fmt.Errorf("rhythmus %s braucht alle_tage größer 0", d.Type)
		}
		r.EveryDays = d.EveryDays
	case "saison":
		r.Type = planner.RhythmSeason
		if len(d.Months) == 0 {
			return r, fmt.Errorf("rhythmus saison braucht monate")
		}
		for _, m := range d.Months {
			if m < 1 || m > 12 {
				return r, fmt.Errorf("monat %d liegt außerhalb von 1..12", m)
			}
			r.Months = append(r.Months, time.Month(m))
		}
	default:
		return r, fmt.Errorf("rhythmus %q ist unbekannt (fest, fenster, ausloeser, saison, phase)", d.Type)
	}
	return r, nil
}

func parseWeekday(s string) (time.Weekday, error) {
	switch strings.ToLower(s) {
	case "mo":
		return time.Monday, nil
	case "di":
		return time.Tuesday, nil
	case "mi":
		return time.Wednesday, nil
	case "do":
		return time.Thursday, nil
	case "fr":
		return time.Friday, nil
	case "sa":
		return time.Saturday, nil
	case "so":
		return time.Sunday, nil
	}
	return 0, fmt.Errorf("wochentag %q ist unbekannt (mo, di, mi, do, fr, sa, so)", s)
}

func parseSlot(s string) planner.Slot {
	switch s {
	case "morgens":
		return planner.SlotMorning
	case "abends":
		return planner.SlotEvening
	}
	return planner.SlotAny
}

func parseFailure(s string) planner.Failure {
	if s == "hart" {
		return planner.FailureHard
	}
	return planner.FailureSoft
}

func parseSource(s string) planner.Source {
	switch s {
	case "haushalt":
		return planner.SourceHousehold
	case "gelernt":
		return planner.SourceLearned
	}
	return planner.SourceCurated
}
