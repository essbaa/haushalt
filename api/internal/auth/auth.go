// Package auth prüft die Token, die die Web-App ausstellt.
//
// Dieses Paket stellt nichts aus und kennt keine Passwörter. Es beantwortet
// eine einzige Frage: Gehört dieses Token zu einer angemeldeten Person, und zu
// welcher? Die Antwort steht danach im Kontext der Anfrage.
//
// Geprüft wird lokal. Der Schlüsselsatz (JWKS) wird einmal geholt und
// zwischengespeichert; danach kostet jede Prüfung eine Signaturrechnung und
// keinen Netzaufruf. Das ist der Grund, warum die Anmeldung in einer anderen
// Sprache und einem anderen Prozess liegen darf: Der Go-Dienst braucht die
// Web-App zur Laufzeit nicht.
//
// Siehe docs/adr/0006-anmeldung-better-auth.md.
package auth

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/MicahParks/keyfunc/v3"
	"github.com/golang-jwt/jwt/v5"
)

// Identity ist, was ein gültiges Token beweist — und nicht mehr.
//
// Absichtlich schmal: Es steht nichts darin, was der Aussteller nicht
// signiert hat. Welche Person im Haushalt dahintersteckt, ist eine Frage an
// die Datenbank und nicht an das Token.
type Identity struct {
	// Subject ist die Nutzerkennung des Anmeldedienstes — dieselbe, die im
	// Fachmodell in member.auth_user_id steht.
	Subject string
	Email   string
	Name    string
}

// Verifier prüft Token gegen den Schlüsselsatz der Web-App.
type Verifier struct {
	keyfunc jwt.Keyfunc
	options []jwt.ParserOption
}

// NewVerifier holt den Schlüsselsatz und hält ihn aktuell.
//
// Der erste Aufruf geht über das Netz; danach erneuert keyfunc im Hintergrund.
// Schlägt der erste Aufruf fehl, startet der Dienst nicht: Ein Dienst, der
// Anmeldung verspricht und keine Schlüssel hat, würde jede Anfrage mit 401
// beantworten und dabei aussehen, als lägen die Passwörter falsch.
func NewVerifier(ctx context.Context, jwksURL, issuer string) (*Verifier, error) {
	if jwksURL == "" {
		return nil, errors.New("auth: JWKS-Adresse fehlt")
	}
	k, err := keyfunc.NewDefaultCtx(ctx, []string{jwksURL})
	if err != nil {
		return nil, fmt.Errorf("auth: schlüsselsatz von %s: %w", jwksURL, err)
	}

	opts := []jwt.ParserOption{
		// Den Algorithmus festnageln. Ohne diese Zeile entscheidet der Kopf
		// des Tokens mit, wie es geprüft wird — der klassische Weg, eine
		// Signaturprüfung wirkungslos zu machen.
		jwt.WithValidMethods([]string{"EdDSA"}),
		jwt.WithExpirationRequired(),
		// Eine Sekunde Nachsicht für auseinanderlaufende Uhren.
		jwt.WithLeeway(time.Second),
	}
	if issuer != "" {
		opts = append(opts, jwt.WithIssuer(issuer))
	}

	return &Verifier{keyfunc: k.Keyfunc, options: opts}, nil
}

// Verify prüft ein Token und gibt zurück, wer es beweist.
func (v *Verifier) Verify(raw string) (Identity, error) {
	var claims jwt.MapClaims
	token, err := jwt.ParseWithClaims(raw, &claims, v.keyfunc, v.options...)
	if err != nil {
		return Identity{}, err
	}
	if !token.Valid {
		return Identity{}, errors.New("auth: token ungültig")
	}

	subject, err := claims.GetSubject()
	if err != nil || subject == "" {
		return Identity{}, errors.New("auth: token ohne subject")
	}
	return Identity{
		Subject: subject,
		Email:   text(claims["email"]),
		Name:    text(claims["name"]),
	}, nil
}

func text(v any) string {
	s, _ := v.(string)
	return s
}

// ------------------------------------------------------------- Middleware

type schluessel struct{}

// Middleware legt die Identität in den Kontext, wenn ein gültiges Token
// mitkommt — und lässt die Anfrage sonst unverändert weiterlaufen.
//
// Absichtlich freiwillig statt zwingend. Die Demo-Haushalte sollen ohne
// Anmeldung sichtbar bleiben; wer welche Anfrage nur angemeldet beantwortet,
// entscheidet der Handler und nicht die Middleware. Eine Middleware, die
// pauschal 401 wirft, macht aus jeder neuen öffentlichen Route eine
// Fehlersuche.
//
// Ein ungültiges Token wird dabei wie keines behandelt. Das ist eine
// Entscheidung und keine Nachlässigkeit: Der Aufrufer erfährt beim Zugriff auf
// geschützte Daten, dass er nicht angemeldet ist — an genau einer Stelle,
// statt an jeder.
func Middleware(v *Verifier) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if v != nil {
				if raw, ok := bearer(r); ok {
					if id, err := v.Verify(raw); err == nil {
						r = r.WithContext(context.WithValue(r.Context(), schluessel{}, id))
					}
				}
			}
			next.ServeHTTP(w, r)
		})
	}
}

// From holt die Identität aus dem Kontext. Das zweite Ergebnis ist false, wenn
// die Anfrage nicht angemeldet ist.
func From(ctx context.Context) (Identity, bool) {
	id, ok := ctx.Value(schluessel{}).(Identity)
	return id, ok
}

func bearer(r *http.Request) (string, bool) {
	kopf := r.Header.Get("Authorization")
	const vorsilbe = "Bearer "
	if len(kopf) <= len(vorsilbe) || !strings.EqualFold(kopf[:len(vorsilbe)], vorsilbe) {
		return "", false
	}
	return strings.TrimSpace(kopf[len(vorsilbe):]), true
}
