import { betterAuth } from "better-auth";
import { jwt } from "better-auth/plugins";
import { Pool } from "pg";

import { appAdresse, schicke } from "@/lib/mail";

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

    // Acht Zeichen, keine erzwungenen Sonderzeichen. Regeln wie „mindestens
    // eine Ziffer und ein Satzzeichen" erzeugen `Passwort1!` und sonst nichts;
    // Länge ist die einzige Anforderung, die wirklich etwas bringt.
    minPasswordLength: 8,

    // Ohne das heißt ein vergessenes Passwort: verlorener Haushalt. Kein
    // Komfortmangel — Datenverlust.
    sendResetPassword: async ({ user, url }) => {
      await schicke({
        an: user.email,
        betreff: "Passwort zurücksetzen",
        text: [
          `Hallo${user.name ? " " + user.name : ""},`,
          "",
          "du hast ein neues Passwort angefordert. Über diesen Link setzt du es:",
          "",
          url,
          "",
          "Der Link gilt eine Stunde. Warst du das nicht, kannst du diese Mail",
          "ignorieren — dein Passwort bleibt, wie es ist.",
        ].join("\n"),
      });
    },
  },

  // Die Bestätigung der Adresse.
  //
  // Sie hängt am Zurücksetzen: Ohne bestätigte Adresse ginge die Rücksetzmail
  // an einen Tippfehler, und der Mensch wartet auf etwas, das nie kommt.
  //
  // `requireEmailVerification` steht bewusst NICHT auf true. Solange keine
  // Domain verifiziert ist, verschickt diese App gar nichts — eine Anmeldung,
  // die eine Bestätigung verlangt, die niemand bekommen kann, sperrt alle aus.
  // Der Auslöser zum Umlegen ist benannt: sobald Resend mit eigener Domain
  // läuft und eine Testmail ankommt.
  emailVerification: {
    sendOnSignUp: true,
    autoSignInAfterVerification: true,
    sendVerificationEmail: async ({ user, url }) => {
      await schicke({
        an: user.email,
        betreff: "E-Mail-Adresse bestätigen",
        text: [
          `Hallo${user.name ? " " + user.name : ""},`,
          "",
          "willkommen. Bestätige kurz deine Adresse:",
          "",
          url,
          "",
          "Danach können wir dir ein neues Passwort schicken, falls du es",
          "vergisst. Ohne bestätigte Adresse geht das nicht.",
        ].join("\n"),
      });
    },
  },

  // Better Auth bringt eine Begrenzung mit; sie ist nur standardmäßig aus.
  // Ein Anmeldeformular ohne Begrenzung lädt zum Durchprobieren ein, und das
  // kostet uns nichts außer dieser Zeile.
  rateLimit: {
    enabled: true,
  },

  // Die Basisadresse für die Links in den Mails oben.
  baseURL: appAdresse(),

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
