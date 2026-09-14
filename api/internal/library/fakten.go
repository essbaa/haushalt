package library

import (
	"encoding/json"
	"fmt"
	"os"
)

// Fact ist ein Ding, das ein Haushalt hat oder nicht hat — samt der Frage, mit
// der die App danach fragt.
//
// Frage und Nutzen stehen als Daten im Repo, nicht im Code: Es ist Inhalt, und
// wer ihn ändert, soll dafür kein Go schreiben müssen.
type Fact struct {
	ID       string `json:"id"`
	Question string `json:"frage"`
	Benefit  string `json:"dann"`
}

// Facts ist der Fragenkatalog, nach Kennung nachschlagbar.
type Facts struct {
	nachID map[string]Fact
}

type factDoc struct {
	Facts []Fact `json:"fakten"`
}

// LoadFacts liest library/fakten.json.
func LoadFacts(path string) (*Facts, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var doc factDoc
	if err := json.Unmarshal(raw, &doc); err != nil {
		return nil, fmt.Errorf("%s: %w", path, err)
	}

	out := &Facts{nachID: make(map[string]Fact, len(doc.Facts))}
	for _, f := range doc.Facts {
		switch {
		case f.ID == "":
			return nil, fmt.Errorf("%s: ein Faktum ohne Kennung", path)
		case f.Question == "":
			return nil, fmt.Errorf("%s: %q hat keine Frage", path, f.ID)
		case f.Benefit == "":
			// Eine Frage ohne erkennbaren Nutzen wird nicht beantwortet. Das
			// ist keine Formalie: Genau daran scheitert jedes Onboarding, das
			// abfragt, bevor es etwas gibt.
			return nil, fmt.Errorf("%s: %q sagt nicht, was die Antwort bringt", path, f.ID)
		}
		out.nachID[f.ID] = f
	}
	return out, nil
}

// Get schlägt ein Faktum nach. Unbekannt heißt: Es gibt keine Frage dazu, und
// dann wird auch keine gestellt.
func (f *Facts) Get(id string) (Fact, bool) {
	if f == nil {
		return Fact{}, false
	}
	fa, ok := f.nachID[id]
	return fa, ok
}

// Known sagt, wie viele Fakten der Katalog kennt.
func (f *Facts) Known() int {
	if f == nil {
		return 0
	}
	return len(f.nachID)
}
