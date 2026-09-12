// Package config liest die Konfiguration aus Umgebungsvariablen.
//
// Alles kommt aus der Umgebung, nichts aus einer Datei. Damit verhält sich der
// Dienst lokal genauso wie bei Fly, wo die Werte als Secrets gesetzt sind —
// und es kann in Produktion keine .env geben, die versehentlich etwas
// überschreibt.
package config

import (
	"errors"
	"fmt"
	"os"
	"strings"
)

type Config struct {
	// Env ist "development" oder "production". Steuert das Log-Format und wie
	// ausführlich Fehler nach außen gehen.
	Env string

	// Port kommt bei Fly aus der Umgebung. Ein fest verdrahteter Wert ergibt
	// ein Deployment, das erfolgreich aussieht und nicht antwortet.
	Port string

	// DatabaseURL darf an T0 leer sein — dann läuft der Dienst ohne Datenbank
	// und sagt das in /healthz auch.
	DatabaseURL string

	// AllowedOrigins sind die Web-Adressen, die den Dienst aufrufen dürfen.
	// Kommagetrennt, damit Vorschau-Domains später dazukommen können.
	AllowedOrigins []string

	// LibraryDir ist das Verzeichnis mit vorlagen.json und beispiele/.
	// Relativ zum Arbeitsverzeichnis: lokal "library", im Container "/library".
	LibraryDir string

	// AuthJWKSURL ist der Schlüsselsatz der Web-App. Leer heißt: Der Dienst
	// kennt keine Anmeldung und behandelt jede Anfrage als anonym.
	AuthJWKSURL string

	// AuthIssuer ist der erwartete Aussteller. Leer heißt: nicht geprüft —
	// in Produktion gehört er gesetzt.
	AuthIssuer string
}

const (
	EnvDevelopment = "development"
	EnvProduction  = "production"
)

func (c Config) IsDevelopment() bool { return c.Env == EnvDevelopment }

// Load liest die Umgebung und prüft sie. Fehler kommen gebündelt zurück, nicht
// einer nach dem anderen — sonst startet man den Dienst dreimal, um drei
// fehlende Variablen zu finden.
func Load() (Config, error) {
	cfg := Config{
		Env:            get("APP_ENV", EnvDevelopment),
		Port:           get("PORT", "8080"),
		DatabaseURL:    os.Getenv("DATABASE_URL"),
		AllowedOrigins: splitAndTrim(get("ALLOWED_ORIGINS", "http://localhost:3000")),
		LibraryDir:     get("LIBRARY_DIR", "library"),
		AuthJWKSURL:    os.Getenv("AUTH_JWKS_URL"),
		AuthIssuer:     os.Getenv("AUTH_ISSUER"),
	}

	var problems []error
	if cfg.Env != EnvDevelopment && cfg.Env != EnvProduction {
		problems = append(problems, fmt.Errorf("APP_ENV ist %q, erlaubt sind %q und %q",
			cfg.Env, EnvDevelopment, EnvProduction))
	}
	if cfg.Port == "" {
		problems = append(problems, errors.New("PORT ist leer"))
	}
	if len(cfg.AllowedOrigins) == 0 {
		problems = append(problems, errors.New("ALLOWED_ORIGINS ist leer — der Browser würde jede Anfrage blockieren"))
	}
	if cfg.Env == EnvProduction && cfg.DatabaseURL == "" {
		problems = append(problems, errors.New("DATABASE_URL fehlt (in Produktion Pflicht)"))
	}

	// errors.Join bündelt mehrere Fehler zu einem. Steht seit Go 1.20 in der
	// Standardbibliothek und erspart eine Fehler-Bibliothek.
	if err := errors.Join(problems...); err != nil {
		return Config{}, fmt.Errorf("konfiguration: %w", err)
	}
	return cfg, nil
}

func get(key, fallback string) string {
	if v := strings.TrimSpace(os.Getenv(key)); v != "" {
		return v
	}
	return fallback
}

func splitAndTrim(s string) []string {
	var out []string
	for _, part := range strings.Split(s, ",") {
		if p := strings.TrimSpace(part); p != "" {
			out = append(out, p)
		}
	}
	return out
}
