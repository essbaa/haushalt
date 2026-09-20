package storage

import (
	"testing"
	"time"
)

// Die Gegenrichtung zu wochentageNachAussen im Vertrag: aus dem Index der
// Datenbank wird ein Wochentag für den Planer. Beide Richtungen stehen in
// verschiedenen Paketen, und genau deshalb steht hier dieselbe Tabelle noch
// einmal — ein Test, der beide Seiten aus derselben Quelle rechnet, prüft
// nur, dass sie zueinander passen, nicht dass sie stimmen.
func TestIndexNullIstMontag(t *testing.T) {
	erwartet := []time.Weekday{
		time.Monday,
		time.Tuesday,
		time.Wednesday,
		time.Thursday,
		time.Friday,
		time.Saturday,
		time.Sunday,
	}
	for i, tag := range erwartet {
		if got := tagAusIndex(i); got != tag {
			t.Errorf("Index %d ergibt %v, erwartet %v", i, got, tag)
		}
	}
}
