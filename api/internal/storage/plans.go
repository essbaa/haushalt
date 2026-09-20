package storage

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/zakaria/haushalt/api/internal/library"
	"github.com/zakaria/haushalt/api/internal/planner"
	"github.com/zakaria/haushalt/api/internal/storage/db"
)

// Plans liest Haushalte und Wochenpläne aus der Datenbank.
//
// Dieselbe Schnittstelle, die der Katalog aus dem Repo erfüllt. Für die
// HTTP-Schicht ist der Unterschied nicht sichtbar — sie bekommt Haushalte und
// Pläne und weiß nicht, ob sie aus einer Datei oder aus Postgres kommen.
//
// Gerechnet wird weiterhin im Planer. Dieses Paket holt Daten und reicht sie
// weiter; es entscheidet nichts.
type Plans struct{ db *DB }

// AsPlans gibt die Datenbank als Planquelle aus.
func (d *DB) AsPlans() *Plans { return &Plans{db: d} }

// Households sind die Haushalte, die dieser Aufrufer sehen darf: die
// Demo-Haushalte aus dem Repo und die eigenen.
//
// subject leer heißt nicht angemeldet — dann bleiben es die Demo-Haushalte.
// Ohne diese Trennung stünde jeder fremde Haushalt in der Liste, und das wäre
// kein Schönheitsfehler, sondern ein Datenleck.
func (p *Plans) Households(ctx context.Context, subject string) ([]planner.Household, error) {
	// Wer einen eigenen Haushalt hat, sieht die Beispiele nicht mehr.
	//
	// Sie sind da, damit jemand ohne Konto sehen kann, was die App tut. Wer
	// eingerichtet hat, ist an diesem Punkt vorbei — und zwei fremde Familien
	// neben dem eigenen Haushalt im Umschalter sind keine Einladung, sondern
	// eine Frage, die sich niemand stellen wollte.
	//
	// Über die Adresse bleiben sie erreichbar: Plan prüft den Zugriff selbst,
	// diese Liste füllt nur den Umschalter.
	var zeilen []db.Household
	if subject != "" {
		eigene, err := p.db.ListHouseholdsForAuthUser(ctx, &subject)
		if err != nil {
			return nil, err
		}
		zeilen = eigene
	}
	if len(zeilen) == 0 {
		demos, err := p.db.ListDemoHouseholds(ctx)
		if err != nil {
			return nil, err
		}
		zeilen = demos
	}

	out := make([]planner.Household, 0, len(zeilen))
	gesehen := map[string]bool{}
	for _, z := range zeilen {
		schluessel := formatUUID(z.ID)
		if gesehen[schluessel] {
			continue // ein Demo-Haushalt, in dem der Aufrufer auch Mitglied ist
		}
		gesehen[schluessel] = true

		h, err := p.household(ctx, z)
		if err != nil {
			return nil, err
		}
		out = append(out, h)
	}
	return out, nil
}

// Plan liefert den Wochenplan — gerechnet, wenn es ihn noch nicht gibt, und
// gelesen, sobald er steht.
//
// Festgeschrieben wird nur für Mitglieder. Die öffentlichen Beispielhaushalte
// bleiben eine Rechnung: Ein Besucher, der sich die Demo ansieht, soll keine
// Zeilen erzeugen, und ihre Pläne sollen mit der Bibliothek mitwachsen statt
// im Januar einzufrieren.
func (p *Plans) Plan(ctx context.Context, subject, id string, week planner.Week) (planner.Result, planner.Household, error) {
	zeile, err := p.lookup(ctx, id)
	if err != nil {
		return planner.Result{}, planner.Household{}, err
	}
	if err := p.mayAccess(ctx, subject, zeile); err != nil {
		return planner.Result{}, planner.Household{}, err
	}
	haushalt, err := p.household(ctx, zeile)
	if err != nil {
		return planner.Result{}, planner.Household{}, err
	}
	vorlagen, err := p.templates(ctx, zeile.ID)
	if err != nil {
		return planner.Result{}, planner.Household{}, err
	}

	// Steht die Woche schon, ist sie die Wahrheit — auch wenn eine neue
	// Rechnung heute etwas anderes ergäbe. Genau das ist der Sinn: Wer am
	// Montag gelesen hat, dass er den Müll rausbringt, soll das am Mittwoch
	// noch so vorfinden.
	_, err = p.db.GetWrittenWeek(ctx, db.GetWrittenWeekParams{
		HouseholdID: zeile.ID,
		ISOWeek:     week.String(),
	})
	switch {
	case err == nil:
		r, err := p.geschriebeneWoche(ctx, haushalt, zeile.ID, week, vorlagen)
		if err == nil {
			// Die offenen Fragen kommen vom Haushalt, wie er JETZT ist, und
			// nicht aus der eingefrorenen Liste der Woche. Sonst kommt eine
			// beantwortete Frage nach dem Neuladen wieder — die Antwort hat
			// den Haushalt geändert, nicht die Momentaufnahme.
			hist, fehler := p.history(ctx, zeile.ID)
			if fehler != nil {
				return planner.Result{}, haushalt, fehler
			}
			r.Open = planner.OpenFacts(vorlagen, haushalt, hist, 0)
		}
		return r, haushalt, err
	case !errors.Is(err, pgx.ErrNoRows):
		return planner.Result{}, planner.Household{}, err
	}

	hist, err := p.history(ctx, zeile.ID)
	if err != nil {
		return planner.Result{}, planner.Household{}, err
	}
	result, err := planner.Plan(planner.Input{
		Household: haushalt,
		Templates: vorlagen,
		Week:      week,
		History:   hist,
		Limits:    planner.DefaultLimits(),
	})
	if err != nil {
		return planner.Result{}, haushalt, err
	}
	result.Open = planner.OpenFacts(vorlagen, haushalt, hist, 0)

	dabei, err := p.istMitglied(ctx, subject, zeile.ID)
	if err != nil {
		return planner.Result{}, haushalt, err
	}
	if !dabei {
		return result, haushalt, nil
	}

	// Festgeschrieben wird nur die laufende Woche.
	//
	// ADR-0008 begründet das Festschreiben damit, dass der Plan sich unter
	// einem nicht ändern soll: Wer am Montag gelesen hat, dass er den Müll
	// rausbringt, findet das am Mittwoch noch so vor. Dieses Versprechen gilt
	// der Woche, in der jemand lebt — und nur ihr.
	//
	// Nach vorn: Ein Blick verspricht nichts. Die kommende Woche einzufrieren
	// wäre ein Nebeneffekt des Hinsehens — der erste Klick auf „nächste
	// Woche" nägelte sie fest, und ein danach eingetragener Anlass käme nie
	// an.
	//
	// Nach hinten: Eine vergangene Woche, die nie geschrieben wurde, hat
	// niemand gesehen. Sie jetzt zu schreiben erfände Vergangenheit — mit
	// Zuteilungen, die nie jemand hatte, und einer Rotation, die daraus
	// weiterrechnet. Bereits geschriebene Wochen sind oben abgefangen; hier
	// landen nur die leeren.
	//
	// Beides wurde erst nötig, als es Pfeile zum Blättern gab. Vorher rief
	// niemand eine andere Woche auf, und die Regel „erstes Ansehen" und die
	// Regel „laufende Woche" sahen gleich aus.
	if week != aktuelleWoche(zeile.Timezone) {
		return result, haushalt, nil
	}

	if err := p.festschreiben(ctx, zeile.ID, week, result); err != nil {
		return planner.Result{}, haushalt, err
	}
	offen := result.Open
	// Noch einmal lesen statt das Gerechnete zurückzugeben: Erst jetzt haben
	// die Aufgaben Kennungen, und im Wettlauf zweier erster Aufrufe steht hier
	// das, was der Schnellere geschrieben hat.
	r, err := p.geschriebeneWoche(ctx, haushalt, zeile.ID, week, vorlagen)
	r.Open = offen
	return r, haushalt, err
}

// istMitglied sagt, ob diese Anmeldung zu einer Person in diesem Haushalt
// gehört. Leeres subject heißt nein — nicht angemeldet ist kein Mitglied.
func (p *Plans) istMitglied(ctx context.Context, subject string, haushalt pgtype.UUID) (bool, error) {
	if subject == "" {
		return false, nil
	}
	_, err := p.db.RoleInHousehold(ctx, db.RoleInHouseholdParams{
		HouseholdID: haushalt,
		AuthUserID:  &subject,
	})
	if errors.Is(err, pgx.ErrNoRows) {
		return false, nil
	}
	return err == nil, err
}

// lookup findet einen Haushalt über den Slug und, falls das nichts ergibt,
// über die Kennung.
//
// Beides zuzulassen ist Absicht: Demo-Haushalte haben sprechende Adressen
// (/api/haushalte/familie-a/…), echte haben nur ihre UUID. Ein Client muss den
// Unterschied nicht kennen.
func (p *Plans) lookup(ctx context.Context, id string) (db.Household, error) {
	zeile, err := p.db.GetHouseholdBySlug(ctx, &id)
	switch {
	case err == nil:
		return zeile, nil
	case !errors.Is(err, pgx.ErrNoRows):
		return db.Household{}, err
	}

	uid, ok := parseUUID(id)
	if !ok {
		return db.Household{}, fmt.Errorf("%w: %q", planner.ErrUnknownHousehold, id)
	}
	zeile, err = p.db.GetHousehold(ctx, uid)
	if errors.Is(err, pgx.ErrNoRows) {
		return db.Household{}, fmt.Errorf("%w: %q", planner.ErrUnknownHousehold, id)
	}
	return zeile, err
}

// mayAccess entscheidet, ob dieser Aufrufer diesen Haushalt sehen darf.
//
// Fremder Haushalt ergibt „unbekannt", nicht „verboten". Das ist Absicht: Ein
// 403 verrät, dass es diesen Haushalt gibt — ein 404 nicht. Bei einer App, in
// der Haushalte über sprechende Adressen erreichbar sind, ist das der
// Unterschied zwischen „du darfst nicht" und „dich betrifft das nicht".
//
// Die Prüfung sitzt hier und nicht im Handler. Ein Handler kann sie
// vergessen; diese Funktion kann man nicht umgehen, ohne Plan zu ändern.
func (p *Plans) mayAccess(ctx context.Context, subject string, z db.Household) error {
	// Demo-Haushalte aus dem Repo sind öffentlich — das ist der Zugang ohne
	// Anmeldung, den der Bauplan für die Demo vorsieht.
	if z.Slug != nil && *z.Slug != "" {
		return nil
	}
	if subject == "" {
		return fmt.Errorf("%w: %q", planner.ErrUnknownHousehold, publicID(z))
	}
	dabei, err := p.db.IsMemberOf(ctx, db.IsMemberOfParams{
		HouseholdID: z.ID,
		AuthUserID:  &subject,
	})
	if err != nil {
		return err
	}
	if !dabei {
		return fmt.Errorf("%w: %q", planner.ErrUnknownHousehold, publicID(z))
	}
	return nil
}

func (p *Plans) household(ctx context.Context, z db.Household) (planner.Household, error) {
	mitglieder, err := p.db.ListMembers(ctx, z.ID)
	if err != nil {
		return planner.Household{}, err
	}

	h := planner.Household{
		ID:   publicID(z),
		Name: z.Name,
		Context: planner.Context{
			Home:    planner.Home(z.Home),
			HasCar:  z.HasCar,
			HasYard: z.HasYard,
			Pets:    z.Pets,
			Facts:   fakten(z.Facts),
			Rooms:   int(z.Rooms),
			Baths:   int(z.Baths),
		},
	}
	anlaesse, err := p.db.ListOccasions(ctx, z.ID)
	if err != nil {
		return planner.Household{}, err
	}
	for _, a := range anlaesse {
		h.Occasions = append(h.Occasions, planner.Occasion{
			ID:     formatUUID(a.ID),
			Title:  a.Title,
			Date:   datum(a.Day),
			Kind:   a.Kind,
			Yearly: a.Yearly,
		})
	}

	jahr := time.Now().Year()
	for _, m := range mitglieder {
		person := planner.Member{
			ID:        formatUUID(m.ID),
			Name:      m.Name,
			Role:      planner.Role(m.Role),
			HasAccess: m.AuthUserID != nil,
		}
		// Das Modell führt Geburtsjahre, der Planer rechnet mit Alter. Die
		// Umrechnung passiert hier und nirgends sonst.
		if m.BirthYear != nil {
			person.BirthYear = int(*m.BirthYear)
			person.Age = jahr - person.BirthYear
		}
		if m.Care != nil {
			person.Care = planner.Care(*m.Care)
		}
		for i, v := range m.CapacityMinutes {
			if i < len(person.CapacityMinutes) {
				person.CapacityMinutes[i] = int(v)
			}
		}
		h.Members = append(h.Members, person)
	}
	return h, nil
}

// templates sind die Vorlagen dieses Haushalts, mit seinen eigenen
// Wochentagen darüber.
//
// Fast alles im Dienst will diese Fassung: den Rhythmus, wie er hier gilt.
// Wer die Bibliotheksfassung braucht — und das ist genau eine Stelle, die
// Prüfung beim Festlegen —, nimmt templatesRoh.
func (p *Plans) templates(ctx context.Context, haushalt pgtype.UUID) ([]planner.TaskTemplate, error) {
	roh, err := p.templatesRoh(ctx, haushalt)
	if err != nil {
		return nil, err
	}
	tage, err := p.wochentage(ctx, haushalt)
	if err != nil {
		return nil, err
	}
	for i, t := range roh {
		if wd, ok := tage[t.ID]; ok && len(wd) > 0 {
			// Feste Tage schlagen jeden anderen Rhythmus: Wer „donnerstags"
			// sagt, meint nicht „alle sieben Tage, bevorzugt donnerstags".
			// Damit ändert sich auch die Häufigkeit — aus „alle 14 Tage,
			// samstags" wird wöchentlich. Das ist keine Nebenwirkung, sondern
			// die Aussage; die Oberfläche sagt es beim Setzen dazu.
			roh[i].Rhythm.Type = planner.RhythmFixed
			roh[i].Rhythm.Weekdays = wd
		}
	}
	return roh, nil
}

// templatesRoh liest die Vorlagen so, wie sie in der Datenbank stehen.
func (p *Plans) templatesRoh(ctx context.Context, haushalt pgtype.UUID) ([]planner.TaskTemplate, error) {
	zeilen, err := p.db.ListTemplatesForHousehold(ctx, haushalt)
	if err != nil {
		return nil, err
	}
	out := make([]planner.TaskTemplate, 0, len(zeilen))
	for _, z := range zeilen {
		// Derselbe Parser wie beim Einlesen der Datei. Ein zweiter wäre die
		// sicherste Art, dass beide Wege auseinanderlaufen.
		t, err := library.ParseTemplate(z.Definition)
		if err != nil {
			return nil, fmt.Errorf("vorlage %q: %w", z.ID, err)
		}
		out = append(out, t)
	}
	return out, nil
}

// wochentage sind die vom Haushalt festgelegten Tage, je Vorlage.
//
// Der Index in der Datenbank ist 0 = Montag, wie beim Wochenraster der
// Absprache. Go zählt anders — time.Sunday ist 0 —, und genau diese
// Umrechnung ist die Stelle, an der so etwas schiefgeht. Deshalb steht sie in
// tagAusIndex und sonst nirgends; die Gegenrichtung liegt in
// wochentageNachAussen, weil nur der Vertrag sie braucht.
func (p *Plans) wochentage(ctx context.Context, haushalt pgtype.UUID) (map[string][]time.Weekday, error) {
	zeilen, err := p.db.ListTemplateWeekdays(ctx, haushalt)
	if err != nil {
		return nil, err
	}
	out := map[string][]time.Weekday{}
	for _, z := range zeilen {
		if z.Weekday < 0 || z.Weekday > 6 {
			continue
		}
		out[z.TemplateID] = append(out[z.TemplateID], tagAusIndex(int(z.Weekday)))
	}
	return out, nil
}

// tagAusIndex: 0 = Montag … 6 = Sonntag, wie in der Datenbank.
func tagAusIndex(i int) time.Weekday { return time.Weekday((i + 1) % 7) }

// history verdichtet, was der Planer über die Vergangenheit wissen muss.
//
// Drei Abfragen, drei Quellen: erledigt kommt aus dem Protokoll, zuletzt
// zugeteilt aus den Zuteilungen, abgewählt aus den Lernwerten. Der Planer
// selbst liest nichts davon — er bekommt das Ergebnis.
func (p *Plans) history(ctx context.Context, haushalt pgtype.UUID) (planner.History, error) {
	hist := planner.History{
		LastDone:     map[string]planner.Date{},
		LastAssignee: map[string]string{},
		Muted:        map[string]bool{},
		FixedTo:      map[string]string{},
	}

	erledigt, err := p.db.LastDonePerTemplate(ctx, haushalt)
	if err != nil {
		return hist, err
	}
	for _, z := range erledigt {
		if z.LastDone.Valid {
			hist.LastDone[z.TemplateID] = planner.DateOf(z.LastDone.Time)
		}
	}

	zugeteilt, err := p.db.LastAssigneePerTemplate(ctx, haushalt)
	if err != nil {
		return hist, err
	}
	for _, z := range zugeteilt {
		if z.MemberID.Valid {
			hist.LastAssignee[z.TemplateID] = formatUUID(z.MemberID)
		}
	}

	absprachen, err := p.db.ListAgreements(ctx, haushalt)
	if err != nil {
		return hist, err
	}
	hist.AgreedTo = map[string][7]string{}
	for _, a := range absprachen {
		if !a.MemberID.Valid || a.Weekday < 0 || a.Weekday > 6 {
			continue
		}
		raster := hist.AgreedTo[a.TemplateID]
		raster[a.Weekday] = formatUUID(a.MemberID)
		hist.AgreedTo[a.TemplateID] = raster
	}

	signale, err := p.db.ListSignals(ctx, haushalt)
	if err != nil {
		return hist, err
	}
	for _, z := range signale {
		if z.Kind == "abgewaehlt" && z.Value != 0 {
			hist.Muted[z.TemplateID] = true
		}
	}
	return hist, nil
}

// fakten liest die JSONB-Spalte. Ein kaputter Inhalt ergibt „nichts bekannt"
// statt eines Fehlers: Der Plan soll dann vorsichtig sein, nicht ausfallen.
func fakten(roh []byte) map[string]bool {
	if len(roh) == 0 {
		return nil
	}
	var out map[string]bool
	if err := json.Unmarshal(roh, &out); err != nil {
		return nil
	}
	return out
}

func zeitpunkt(t time.Time) pgtype.Timestamptz {
	return pgtype.Timestamptz{Time: t, Valid: true}
}

// tag macht aus einem Kalendertag eine Datumsspalte. Mitternacht UTC, weil
// date in Postgres keine Uhrzeit hat und jede erfundene die Datumsgrenze
// verschieben könnte (ADR-0002).
func tag(d planner.Date) pgtype.Date {
	if d.IsZero() {
		return pgtype.Date{}
	}
	return pgtype.Date{
		Time:  time.Date(d.Year, d.Month, d.Day, 0, 0, 0, 0, time.UTC),
		Valid: true,
	}
}

// datum ist die Rückrichtung. Eine leere Spalte ergibt das Nulldatum, und das
// heißt im Planer „nicht gesetzt" — etwa bei einer Aufgabe ohne Frist.
func datum(d pgtype.Date) planner.Date {
	if !d.Valid {
		return planner.Date{}
	}
	return planner.DateOf(d.Time)
}

// ------------------------------------------------------------------ UUID

// publicID ist der Slug, wenn es einen gibt, sonst die Kennung.
func publicID(z db.Household) string {
	if z.Slug != nil && *z.Slug != "" {
		return *z.Slug
	}
	return formatUUID(z.ID)
}

// formatUUID schreibt die sechzehn Bytes in die übliche Form.
//
// pgtype.UUID hat dafür keine Methode, und eine Abhängigkeit nur für sechs
// Zeilen Hexadezimal wäre übertrieben.
func formatUUID(u pgtype.UUID) string {
	if !u.Valid {
		return ""
	}
	h := hex.EncodeToString(u.Bytes[:])
	return h[0:8] + "-" + h[8:12] + "-" + h[12:16] + "-" + h[16:20] + "-" + h[20:32]
}

func parseUUID(s string) (pgtype.UUID, bool) {
	var u pgtype.UUID
	roh := make([]byte, 0, 32)
	for i := 0; i < len(s); i++ {
		if s[i] != '-' {
			roh = append(roh, s[i])
		}
	}
	if len(roh) != 32 {
		return u, false
	}
	b, err := hex.DecodeString(string(roh))
	if err != nil {
		return u, false
	}
	copy(u.Bytes[:], b)
	u.Valid = true
	return u, true
}

// aktuelleWoche ist die laufende Woche in der Zeitzone des Haushalts.
//
// Die Zeitzone gehört hierher und nicht in den Planer: Der rechnet in UTC und
// kennt keine Uhr (ADR-0002). Ob „heute" schon Montag ist, entscheidet sich
// aber dort, wo die Menschen wohnen — für einen Haushalt in Berlin ist es
// sonntags um 23:30 Uhr noch die alte Woche, in UTC schon die neue.
//
// Eine unbekannte Zeitzone ergibt UTC. Falsch zu rechnen ist besser, als
// deshalb gar keinen Plan auszuliefern.
// heuteIn ist der laufende Kalendertag in der Zeitzone des Haushalts.
//
// Dieselbe Rechnung wie in aktuelleWoche, nur einen Schritt kürzer: Für „ab
// heute" zählt der Tag und nicht die Woche. Ein Aufruf kurz nach Mitternacht
// in Frankfurt ist in UTC noch gestern — und träfe damit einen Tag, der
// bereits vorbei ist.
func heuteIn(zone string) planner.Date {
	ort := time.UTC
	if zone != "" {
		if geladen, err := time.LoadLocation(zone); err == nil {
			ort = geladen
		}
	}
	return planner.DateOf(time.Now().In(ort))
}

func aktuelleWoche(zone string) planner.Week {
	ort := time.UTC
	if zone != "" {
		if geladen, err := time.LoadLocation(zone); err == nil {
			ort = geladen
		}
	}
	return planner.WeekOf(planner.DateOf(time.Now().In(ort)))
}
