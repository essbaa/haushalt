package planner

// Scale sagt, ob eine Vorlage mit der Größe des Haushalts wächst.
//
// Der Unterschied zu allen anderen Bedingungen: Die entscheiden, *welche*
// Aufgaben es gibt. Diese entscheidet, *wie lange* sie dauern — und die Dauer
// trägt die Fairnessrechnung. Bisher bekam jeder Haushalt dieselben Zahlen:
// Staubsaugen 25 Minuten in der Zweizimmerwohnung wie im Fünfzimmerhaus.
type Scale string

const (
	ScaleNone  Scale = ""
	ScaleRooms Scale = "zimmer"
	// ScaleBaths ist die Ausnahme: Es skaliert keine Dauer, sondern die
	// Anzahl. Zwei Bäder sind nicht eine doppelt so lange Aufgabe, sondern
	// zwei Aufgaben — man putzt Bad A und Bad B, womöglich an verschiedenen
	// Tagen und von verschiedenen Personen.
	//
	// Der Unterschied ist sauber benennbar: Bäder sind eine Anzahl, Zimmer und
	// Personen sind ein Ausmaß. Als Ausmaß gerechnet entstand eine
	// Einzelaufgabe von 70 Minuten, und die landete bei einer Dreizehnjährigen.
	ScaleBaths Scale = "baeder"

	// ScalePeople gilt für alles, was sich nach Kopfzahl richtet: Betten
	// beziehen, Wäsche, Bettwäsche. Vier Personen haben vier Betten, ob in
	// drei oder fünf Zimmern — Zimmer sind ein guter Maßstab für Fläche und
	// ein schlechter für Kopfzahl.
	ScalePeople Scale = "personen"
)

const (
	// ReferenceRooms und ReferenceBaths sind der Haushalt, für den die
	// Bibliothek kuratiert ist: drei Zimmer, ein Bad. Alle Dauern in
	// vorlagen.json beziehen sich darauf.
	ReferenceRooms  = 3
	ReferenceBaths  = 1
	ReferencePeople = 3

	// Der Deckel ist kein Detail. Ohne ihn bekäme eine Einzimmerwohnung acht
	// Minuten Staubsaugen und ein Zehnzimmerhaus dreiundachtzig — beides
	// Zahlen, die niemand ernst nimmt, und der Plan verliert seine
	// Glaubwürdigkeit an einer Rechnung.
	minFactorPermille = 600
	maxFactorPermille = 2000
)

// DurationFor ist die Dauer dieser Vorlage in diesem Haushalt.
//
// Gerechnet in Promille statt mit Fließkomma: Der Planer ist deterministisch,
// und zwei Läufe mit derselben Eingabe sollen bis auf die Minute dasselbe
// ergeben — auch auf einer anderen Maschine.
func (t TaskTemplate) DurationFor(c Context) int { return t.durationWith(c, 0) }

// DurationIn ist die Dauer in einem Haushalt — mit Kopfzahl, wo sie zählt.
//
// Zwei Wege, weil nur der Haushalt weiß, wie viele Menschen darin leben, der
// Kontext aber nicht. Wer nur den Kontext hat, bekommt die Zahl ohne
// Personenmaßstab; das ist genau, was die Anzeige einer einzelnen Vorlage
// braucht.
func (t TaskTemplate) DurationIn(h Household) int {
	return t.durationWith(h.Context, len(h.Members))
}

func (t TaskTemplate) durationWith(c Context, personen int) int {
	faktor := 1000
	switch t.ScalesWith {
	case ScaleRooms:
		faktor = anteil(c.Rooms, ReferenceRooms)
	case ScaleBaths:
		// Die Dauer bleibt. Die Vervielfachung passiert bei der Fälligkeit
		// (siehe timesFor in due.go).
		return t.DurationMin
	case ScalePeople:
		if personen == 0 {
			return t.DurationMin
		}
		faktor = anteil(personen, ReferencePeople)
	default:
		return t.DurationMin
	}

	minuten := (t.DurationMin*faktor + 500) / 1000
	if minuten < 1 {
		return 1
	}
	return minuten
}

func anteil(ist, bezug int) int {
	if ist <= 0 || bezug <= 0 {
		return 1000
	}
	f := ist * 1000 / bezug
	if f < minFactorPermille {
		return minFactorPermille
	}
	if f > maxFactorPermille {
		return maxFactorPermille
	}
	return f
}

// TimesIn sagt, wie oft diese Vorlage je Fälligkeit anfällt.
//
// Für alles außer Bädern ist das eins. Null heißt: gar nicht — ein Haushalt
// ohne Bad putzt keines.
func (t TaskTemplate) TimesIn(c Context) int {
	if t.ScalesWith != ScaleBaths {
		return 1
	}
	if c.Baths <= 0 {
		return 0
	}
	if c.Baths > 5 {
		return 5
	}
	return c.Baths
}

// ScaleTemplates rechnet die Dauern einmal auf diesen Haushalt um.
//
// Einmal ganz vorne und nicht an jeder Stelle, die eine Dauer braucht: Sonst
// rechnet die Zuteilung mit der einen Zahl, die Kapazitätsprüfung mit der
// anderen und die Anzeige mit einer dritten.
func ScaleTemplates(ts []TaskTemplate, h Household) []TaskTemplate {
	out := make([]TaskTemplate, len(ts))
	for i, t := range ts {
		t.DurationMin = t.DurationIn(h)
		out[i] = t
	}
	return out
}
