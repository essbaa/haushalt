package planner

import "time"

// Rückmeldungen — was der Haushalt über die App sagt.
//
// Die Typen stehen in diesem Paket, obwohl der Planer sie nie anfasst. Der
// Grund ist der praktische: Hier liegt das gemeinsame Vokabular von Speicher
// und Schnittstelle, und eine Rückmeldung ist ein Begriff, den beide
// brauchen. Ein eigenes Paket für drei Typen wäre sauberer im Diagramm und
// teurer in jeder Datei, die es importiert.
//
// Warum es das überhaupt gibt: Ein Zettel am Kühlschrank wird abends
// ausgefüllt, aus dem Gedächtnis, von der Person, die ihn aufgehängt hat. Der
// Satz, den jemand am Mittwochmorgen gedacht hat, steht dort am Abend nicht
// mehr im selben Wortlaut — und der Wortlaut ist das, was bei einer
// Rückmeldung zählt.

// FeedbackKind ist die Art einer Rückmeldung.
//
// Drei und nicht zwei: „Fehler" und „Idee" sind die üblichen, und dazwischen
// fällt das Wichtigste durch. „Stört mich" ist kein Fehler — die App tut, was
// sie soll — und keine Idee, weil niemand eine Lösung anzubieten hat. Es ist
// die Kategorie, in der die Sätze landen, die man sonst nirgends meldet und
// wegen derer man eine App am Mittwoch zuklappt.
type FeedbackKind string

const (
	FeedbackBug   FeedbackKind = "fehler"
	FeedbackIdea  FeedbackKind = "idee"
	FeedbackNoise FeedbackKind = "stoert"
)

// AllFeedbackKinds ist die vollständige Liste — dieselbe Vorsorge wie bei
// AllSkipCodes: Das Wort steht auch im Vertrag und in einer CHECK-Bedingung.
var AllFeedbackKinds = []FeedbackKind{FeedbackBug, FeedbackIdea, FeedbackNoise}

// Feedback ist eine gemeldete Rückmeldung.
type Feedback struct {
	ID   string
	Kind FeedbackKind
	Text string
	// Context ist, woher die Meldung kam: Seite, Woche, Fassung. Gesammelt von
	// der Oberfläche, nicht vom Menschen getippt — wer erst beschreiben muss,
	// wo er war, meldet nichts.
	Context string
	// Who ist der Name der meldenden Person. Leer, wenn sie den Haushalt
	// verlassen hat: Die Rückmeldung bleibt wahr, auch wenn niemand mehr
	// danebensteht.
	Who string
	At  time.Time
}
