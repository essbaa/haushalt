package httpapi

import (
	"os"
	"regexp"
	"strings"
	"testing"

	"github.com/zakaria/haushalt/api/internal/httpapi/openapi"
	"github.com/zakaria/haushalt/api/internal/planner"
)

// Der Planer und der Vertrag führen dieselben geschlossenen Wortlisten — die
// Gründe fürs Überspringen sogar zweimal, für die Vorlagenliste und für den
// Wochenplan. Drei Listen von Hand laufen auseinander, und zwar leise: Die
// Liste in `Uebersprungen.grund` kannte `unbekannt` und `braucht_termin`
// monatelang nicht, obwohl der Planer beide meldet. Aufgefallen ist es erst,
// als die Oberfläche einen der Werte zum ersten Mal vergleichen wollte statt
// ihn nur nachzuschlagen — bis dahin log der Vertrag, ohne dass etwas brach.
//
// Diese beiden Tests sind billig und fangen genau diesen Fehler beim nächsten
// Mal in `make check` statt im Browser. Dieselbe Sorte Absicherung wie das
// `git diff --exit-code` über die erzeugten Dateien in der CI: Der Vertrag ist
// nur dann die eine Wahrheit, wenn etwas nachrechnet, dass er es ist.
func TestJederUebersprunggrundStehtImVertrag(t *testing.T) {
	for _, code := range planner.AllSkipCodes {
		if !openapi.VorlagenStandGrund(code).Valid() {
			t.Errorf("VorlagenStand.grund kennt %q nicht", code)
		}
		if !openapi.UebersprungenGrund(code).Valid() {
			t.Errorf("Uebersprungen.grund kennt %q nicht", code)
		}
	}
}

func TestJedeBegruendungStehtImVertrag(t *testing.T) {
	for _, code := range planner.AllReasonCodes {
		if !openapi.BegruendungCode(code).Valid() {
			t.Errorf("Begruendung.code kennt %q nicht", code)
		}
	}
}

// TestJedeMethodeStehtInDerVorabfrage liest die Spezifikation und prüft, dass
// CORS jede Methode erlaubt, die es darin gibt.
//
// Der Fehler, der dazu geführt hat, ist der unangenehmste seiner Art: Die
// erste PUT-Route kam dazu, die Vorabfrage antwortete mit 204 — und ohne PUT
// in der Liste. Der Browser brach dann mit „Failed to fetch" ab, und im
// Protokoll des Dienstes stand nur die OPTIONS-Zeile mit Status 204. Alles
// sah richtig aus. Die eigentliche Anfrage ist nie losgeschickt worden.
//
// Bewusst über die YAML-Datei und nicht über den erzeugten Code: Die Methoden
// stehen dort als Wort da, und ein grobes Muster genügt. Ein Test, der dafür
// einen Parser bräuchte, wäre teurer als der Fehler, den er findet.
func TestJedeMethodeStehtInDerVorabfrage(t *testing.T) {
	roh, err := os.ReadFile("../../../openapi.yaml")
	if err != nil {
		t.Fatalf("Spezifikation nicht lesbar: %v", err)
	}

	// Vier Leerzeichen Einzug, dann die Methode: So stehen sie unter einem
	// Pfad und nirgends sonst.
	muster := regexp.MustCompile(`(?m)^    (get|put|post|patch|delete):`)
	treffer := muster.FindAllStringSubmatch(string(roh), -1)
	if len(treffer) == 0 {
		t.Fatal("keine einzige Methode gefunden — das Muster passt nicht mehr")
	}

	for _, m := range treffer {
		methode := strings.ToUpper(m[1])
		if !strings.Contains(erlaubteMethoden, methode) {
			t.Errorf("die Vorabfrage erlaubt %s nicht: %q", methode, erlaubteMethoden)
		}
	}
}
