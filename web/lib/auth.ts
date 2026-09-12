import { betterAuth } from "better-auth";
import { jwt } from "better-auth/plugins";
import { Pool } from "pg";

/**
 * Die Anmeldung läuft in der Next.js-App, nicht im Go-Dienst.
 *
 * Better Auth ist eine Bibliothek, kein Dienst: Nutzer und Sitzungen liegen in
 * unserer eigenen Datenbank — demselben Postgres, in dem auch die Haushalte
 * stehen. Für eine App mit Namen und Alter von Kindern ist das die stärkste
 * Antwort, die es gibt: Es gibt keinen Dritten, dem etwas zu erklären wäre.
 *
 * Der Preis steht in ADR-0006: Diese App ist damit der Autorisierungsserver
 * und kein dünner Client mehr. Sie schreibt selbst in die Datenbank, und die
 * Anmeldung hängt an ihrer Verfügbarkeit.
 *
 * WEM WELCHE TABELLE GEHÖRT — die Regel, ohne die zwei Migrationswerkzeuge auf
 * einer Datenbank Ärger machen:
 *
 *   user, session, account, verification   →  Better Auth, über sein CLI
 *   alles andere                           →  goose, in api/db/migrations
 *
 * Keines der beiden fasst die Tabellen des anderen an. Die einzige Verbindung
 * ist die Spalte member.auth_user_id im Fachmodell.
 */
export const auth = betterAuth({
  // Kysely mit dem pg-Pool. Kein ORM: Better Auth bringt sein eigenes Schema
  // mit, und wir wollen es nicht in unsere Modelle mischen.
  database: new Pool({ connectionString: process.env.DATABASE_URL }),

  emailAndPassword: {
    enabled: true,
  },

  plugins: [
    /**
     * Das JWT-Plugin ist der Grund, warum diese Wahl mit einem Go-Backend
     * überhaupt funktioniert.
     *
     * Es stellt zwei Endpunkte bereit: /api/auth/token liefert ein
     * signiertes Token, /api/auth/jwks die öffentlichen Schlüssel dazu. Der
     * Go-Dienst prüft die Signatur damit LOKAL — ohne Rückfrage, ohne
     * Datenbankabfrage, ohne diese App überhaupt zu erreichen.
     *
     * Die Dokumentation sagt ausdrücklich, dass das Token kein Sitzungsersatz
     * ist, sondern für Fremddienste gedacht. Der Go-Dienst ist dieser
     * Fremddienst; im Browser bleibt die Sitzung ein Cookie.
     */
    jwt(),
  ],
});
