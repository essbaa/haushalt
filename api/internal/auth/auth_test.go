package auth

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestBearer(t *testing.T) {
	tests := []struct {
		name string
		kopf string
		want string
		ok   bool
	}{
		{"normal", "Bearer abc.def.ghi", "abc.def.ghi", true},
		{"kleingeschrieben", "bearer abc", "abc", true},
		{"mit Leerraum", "Bearer   abc  ", "abc", true},
		{"leer", "", "", false},
		{"nur das Wort", "Bearer", "", false},
		{"nur das Wort mit Leerzeichen", "Bearer ", "", false},
		{"anderes Verfahren", "Basic abc", "", false},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			r := httptest.NewRequest(http.MethodGet, "/", nil)
			if tc.kopf != "" {
				r.Header.Set("Authorization", tc.kopf)
			}
			got, ok := bearer(r)
			if ok != tc.ok || got != tc.want {
				t.Fatalf("bearer(%q) = %q, %v — erwartet %q, %v", tc.kopf, got, ok, tc.want, tc.ok)
			}
		})
	}
}

// TestMiddlewareOhneVerifier hält die Entscheidung fest, dass der Dienst ohne
// eingerichtete Anmeldung weiterläuft und jede Anfrage als anonym gilt.
// Andernfalls wäre eine fehlende Umgebungsvariable ein Totalausfall.
func TestMiddlewareOhneVerifier(t *testing.T) {
	var angekommen bool
	var angemeldet bool

	h := Middleware(nil)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		angekommen = true
		_, angemeldet = From(r.Context())
	}))

	r := httptest.NewRequest(http.MethodGet, "/", nil)
	r.Header.Set("Authorization", "Bearer irgendwas")
	h.ServeHTTP(httptest.NewRecorder(), r)

	if !angekommen {
		t.Fatal("die Anfrage kam nicht beim Handler an")
	}
	if angemeldet {
		t.Fatal("ohne Verifier darf keine Identität im Kontext stehen")
	}
}

// TestUngueltigesTokenGiltWieKeines: Ein kaputtes Token führt nicht zu 401 in
// der Middleware, sondern zu einer anonymen Anfrage. Wer welche Daten nur
// angemeldet herausgibt, entscheidet der Handler.
func TestUngueltigesTokenGiltWieKeines(t *testing.T) {
	// Ein Verifier ohne Schlüsselfunktion: jwt.ParseWithClaims gibt dann einen
	// Fehler zurück, und genau das ist der Fall, den wir prüfen wollen.
	var angemeldet bool
	h := Middleware(&Verifier{})(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, angemeldet = From(r.Context())
	}))

	r := httptest.NewRequest(http.MethodGet, "/", nil)
	r.Header.Set("Authorization", "Bearer kein.gueltiges.token")
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, r)

	if angemeldet {
		t.Fatal("ein ungültiges Token darf keine Identität ergeben")
	}
	if rec.Code != http.StatusOK {
		t.Fatalf("Status %d — die Middleware darf die Anfrage nicht abweisen", rec.Code)
	}
}
