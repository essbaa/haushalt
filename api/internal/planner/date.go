package planner

import (
	"fmt"
	"time"
)

// Date ist ein Kalendertag ohne Uhrzeit und ohne Zeitzone.
//
// Der Planer rechnet bewusst nicht mit time.Time: Ein Wochenplan kennt Tage,
// keine Zeitpunkte. Das erspart Zeitzonen- und Sommerzeitfehler an einer
// Stelle, an der sie nur schaden.
type Date struct {
	Year  int
	Month time.Month
	Day   int
}

// DateOf schneidet die Uhrzeit von t ab.
func DateOf(t time.Time) Date {
	y, m, d := t.Date()
	return Date{Year: y, Month: m, Day: d}
}

// MustDate baut ein Datum aus "2026-09-14". Panik bei falschem Format —
// nur für Tests und Konstanten gedacht.
func MustDate(s string) Date {
	t, err := time.Parse("2006-01-02", s)
	if err != nil {
		panic(err)
	}
	return DateOf(t)
}

func (d Date) time() time.Time {
	return time.Date(d.Year, d.Month, d.Day, 0, 0, 0, 0, time.UTC)
}

func (d Date) String() string { return d.time().Format("2006-01-02") }

func (d Date) IsZero() bool { return d == Date{} }

func (d Date) Weekday() time.Weekday { return d.time().Weekday() }

func (d Date) AddDays(n int) Date { return DateOf(d.time().AddDate(0, 0, n)) }

func (d Date) Before(o Date) bool { return d.time().Before(o.time()) }

func (d Date) After(o Date) bool { return d.time().After(o.time()) }

// DaysUntil ist die Anzahl Tage von d bis o; negativ, wenn o vor d liegt.
func (d Date) DaysUntil(o Date) int {
	return int(o.time().Sub(d.time()).Hours() / 24)
}

// IsWeekend ist true für Samstag und Sonntag.
func (d Date) IsWeekend() bool {
	w := d.Weekday()
	return w == time.Saturday || w == time.Sunday
}

// Week ist eine ISO-8601-Kalenderwoche.
type Week struct {
	Year int
	Week int
}

// ParseWeek liest "2026-W38".
func ParseWeek(s string) (Week, error) {
	var w Week
	if _, err := fmt.Sscanf(s, "%d-W%d", &w.Year, &w.Week); err != nil {
		return Week{}, fmt.Errorf("woche %q: erwartet wird das Format 2026-W38", s)
	}
	if w.Week < 1 || w.Week > 53 {
		return Week{}, fmt.Errorf("woche %q: %d liegt außerhalb von 1..53", s, w.Week)
	}
	return w, nil
}

func WeekOf(d Date) Week {
	y, n := d.time().ISOWeek()
	return Week{Year: y, Week: n}
}

func (w Week) String() string { return fmt.Sprintf("%d-W%02d", w.Year, w.Week) }

// Monday ist der erste Tag der Woche.
func (w Week) Monday() Date {
	// Der 4. Januar liegt nach ISO 8601 immer in Kalenderwoche 1.
	jan4 := time.Date(w.Year, time.January, 4, 0, 0, 0, 0, time.UTC)
	offset := (int(jan4.Weekday()) + 6) % 7 // Montag = 0
	mondayOfWeek1 := jan4.AddDate(0, 0, -offset)
	return DateOf(mondayOfWeek1.AddDate(0, 0, (w.Week-1)*7))
}

// Days sind Montag bis Sonntag der Woche.
func (w Week) Days() [7]Date {
	var days [7]Date
	monday := w.Monday()
	for i := range days {
		days[i] = monday.AddDays(i)
	}
	return days
}

// Contains sagt, ob d in dieser Woche liegt.
func (w Week) Contains(d Date) bool {
	return WeekOf(d) == w
}
