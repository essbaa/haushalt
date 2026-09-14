package planner

import "testing"

func TestDurationFor(t *testing.T) {
	saugen := TaskTemplate{ID: "t-saugen", DurationMin: 25, ScalesWith: ScaleRooms}
	bad := TaskTemplate{ID: "t-bad", DurationMin: 35, ScalesWith: ScaleBaths}
	muell := TaskTemplate{ID: "t-muell", DurationMin: 5}

	t.Run("der Bezugshaushalt behält die Zahl aus der Bibliothek", func(t *testing.T) {
		c := Context{Rooms: ReferenceRooms, Baths: ReferenceBaths}
		if got := saugen.DurationFor(c); got != 25 {
			t.Errorf("%d min statt 25 bei drei Zimmern", got)
		}
	})

	t.Run("mehr Zimmer, länger", func(t *testing.T) {
		if got := saugen.DurationFor(Context{Rooms: 5, Baths: 1}); got <= 25 {
			t.Errorf("fünf Zimmer ergeben %d min, nicht mehr als drei", got)
		}
	})

	t.Run("weniger Zimmer, kürzer", func(t *testing.T) {
		if got := saugen.DurationFor(Context{Rooms: 2, Baths: 1}); got >= 25 {
			t.Errorf("zwei Zimmer ergeben %d min, nicht weniger als drei", got)
		}
	})

	t.Run("der Deckel hält beide Enden", func(t *testing.T) {
		klein := saugen.DurationFor(Context{Rooms: 1, Baths: 1})
		gross := saugen.DurationFor(Context{Rooms: 12, Baths: 1})
		if klein < 15 {
			t.Errorf("eine Einzimmerwohnung ergibt %d min — unter dem Deckel", klein)
		}
		if gross > 50 {
			t.Errorf("ein Zwölfzimmerhaus ergibt %d min — über dem Deckel", gross)
		}
	})

	t.Run("zwei Bäder sind zwei Aufgaben, keine längere", func(t *testing.T) {
		// Der Unterschied ist der Punkt: Man putzt Bad A und Bad B, womöglich
		// an verschiedenen Tagen. Als eine Aufgabe von 70 Minuten gerechnet
		// landete sie bei einer Dreizehnjährigen.
		if got := bad.DurationFor(Context{Rooms: 3, Baths: 2}); got != 35 {
			t.Errorf("die Dauer wurde auf %d min gestreckt statt die Anzahl zu erhöhen", got)
		}
		if got := bad.TimesIn(Context{Rooms: 3, Baths: 2}); got != 2 {
			t.Errorf("zwei Bäder ergeben %d Aufgaben statt zwei", got)
		}
		if got := bad.TimesIn(Context{Rooms: 3, Baths: 0}); got != 0 {
			t.Errorf("ohne Bad ergeben sich %d Aufgaben statt keiner", got)
		}
		if got := saugen.TimesIn(Context{Rooms: 5, Baths: 2}); got != 1 {
			t.Errorf("Staubsaugen ergibt %d Aufgaben — fünf Zimmer sind eine längere, keine fünf", got)
		}
	})

	t.Run("was nicht skaliert, bleibt", func(t *testing.T) {
		if got := muell.DurationFor(Context{Rooms: 9, Baths: 3}); got != 5 {
			t.Errorf("Müll rausbringen dauert %d min statt 5 — es skaliert nicht", got)
		}
	})

	t.Run("ohne Angabe passiert nichts", func(t *testing.T) {
		if got := saugen.DurationFor(Context{}); got != 25 {
			t.Errorf("ohne Größenangabe %d min statt der Bibliothekszahl", got)
		}
	})
}

func TestDurationInNachKopfzahl(t *testing.T) {
	betten := TaskTemplate{ID: "t-betten", DurationMin: 20, ScalesWith: ScalePeople}

	haushalt := func(n int) Household {
		h := Household{Context: Context{Rooms: 3, Baths: 1}}
		for i := 0; i < n; i++ {
			h.Members = append(h.Members, Member{ID: string(rune('a' + i))})
		}
		return h
	}

	// Vier Personen haben vier Betten, ob in drei oder fünf Zimmern. Zimmer
	// sind ein guter Maßstab für Fläche und ein schlechter für Kopfzahl.
	drei := betten.DurationIn(haushalt(ReferencePeople))
	fuenf := betten.DurationIn(haushalt(5))
	if drei != 20 {
		t.Errorf("der Bezugshaushalt ergibt %d min statt 20", drei)
	}
	if fuenf <= drei {
		t.Errorf("fünf Personen ergeben %d min, nicht mehr als drei (%d)", fuenf, drei)
	}

	// Ohne Mitglieder bleibt die Zahl aus der Bibliothek stehen: Eine einzelne
	// Vorlage anzuzeigen darf nicht davon abhängen, ob gerade jemand da ist.
	if got := betten.DurationFor(Context{Rooms: 3, Baths: 1}); got != 20 {
		t.Errorf("ohne Haushalt %d min statt der Bibliothekszahl", got)
	}
}
