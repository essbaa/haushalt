import type { authClient } from "@/lib/auth-client";

/**
 * Deutsche Sätze für die Fehler des Anmeldedienstes.
 *
 * Vorher stand auf der Anmeldeseite `antwort.error.message` — und das ist
 * englisch: „Invalid email or password.". Der Kommentar dazu war ehrlich
 * („die echte Meldung ist besser als eine erfundene"), aber die Rechnung geht
 * nicht auf: Deutsche Überschrift, deutscher Knopf, englischer Fehler. Der
 * Moment, in dem etwas schiefgeht, ist genau der, in dem eine App am
 * wenigsten fremd klingen darf — und für die meisten Menschen ist die
 * Anmeldeseite der erste Bildschirm überhaupt.
 *
 * Übersetzt wird über den **Code**, nicht über den Text. Der Text ist die
 * Ausgabe einer fremden Bibliothek und ändert sich mit jeder Version; der
 * Code ist ihre Zusage. Und weil nur die Fälle in der Tabelle stehen, die wir
 * kennen, bleibt der alte Kommentar für alles andere wahr: Was wir nicht
 * übersetzt haben, kommt im Original durch. Ein englischer Satz ist dann ein
 * Hinweis auf eine Lücke — eine erfundene deutsche Entsprechung wäre keiner.
 *
 * Der Typ kommt aus dem Client selbst (`$ERROR_CODES`). Ein Tippfehler in
 * einem Schlüssel ist damit ein Übersetzungsfehler, kein toter Eintrag, der
 * jahrelang niemandem auffällt — dieselbe Aufgabe wie die Vertragstests im
 * Dienst, hier vom Compiler erledigt statt von einem Test.
 */

type Fehlercode = keyof typeof authClient.$ERROR_CODES;

/** Was der Client im Fehlerfall zurückgibt. */
export type Anmeldefehler = {
  code?: string;
  message?: string;
  status?: number;
};

/**
 * Zwei Fälle bekommen bewusst **denselben** Satz wie „stimmt nicht":
 * `USER_NOT_FOUND` und `CREDENTIAL_ACCOUNT_NOT_FOUND`. Beide verraten sonst,
 * ob es zu einer Adresse ein Konto gibt — und wer das abfragen kann, hat eine
 * Liste eurer Mitglieder. Die englischen Originale verraten es; die
 * Übersetzung ist die Gelegenheit, das zuzumachen, statt es mitzuübersetzen.
 * Dieselbe Regel gilt schon beim Passwort-Vergessen und bei Einladungscodes.
 */
const stimmtNicht = "E-Mail-Adresse oder Passwort stimmt nicht.";

const saetze: Partial<Record<Fehlercode, string>> = {
  INVALID_EMAIL_OR_PASSWORD: stimmtNicht,
  INVALID_PASSWORD: stimmtNicht,
  USER_NOT_FOUND: stimmtNicht,
  CREDENTIAL_ACCOUNT_NOT_FOUND: stimmtNicht,

  INVALID_EMAIL: "Das sieht nicht wie eine E-Mail-Adresse aus.",
  USER_ALREADY_EXISTS: "Zu dieser Adresse gibt es schon ein Konto. Melde dich damit an.",
  USER_ALREADY_EXISTS_USE_ANOTHER_EMAIL:
    "Zu dieser Adresse gibt es schon ein Konto. Melde dich damit an.",
  EMAIL_NOT_VERIFIED:
    "Diese Adresse ist noch nicht bestätigt. In deinem Postfach liegt eine Mail mit dem Link.",
  EMAIL_ALREADY_VERIFIED: "Diese Adresse ist schon bestätigt. Du kannst dich anmelden.",

  PASSWORD_TOO_SHORT: "Das Passwort ist zu kurz.",
  PASSWORD_TOO_LONG: "Das Passwort ist zu lang.",

  INVALID_TOKEN: "Dieser Link stimmt nicht. Fordere einen neuen an.",
  TOKEN_EXPIRED: "Dieser Link ist abgelaufen. Fordere einen neuen an.",
  SESSION_EXPIRED: "Deine Anmeldung ist abgelaufen. Melde dich noch einmal an.",

  FAILED_TO_CREATE_USER: "Das Konto konnte nicht angelegt werden. Versuch es gleich noch einmal.",
  FAILED_TO_CREATE_SESSION: "Die Anmeldung hat nicht geklappt. Versuch es gleich noch einmal.",
  VALIDATION_ERROR: "Da fehlt noch eine Angabe.",
  MISSING_FIELD: "Da fehlt noch eine Angabe.",
};

/**
 * Der Satz, der auf dem Bildschirm steht.
 *
 * Die Reihenfolge ist die Reihenfolge der Verlässlichkeit: erst der Fall ohne
 * Code, dann unsere Übersetzung, dann der Originaltext, dann der Rückfall.
 */
export function anmeldeFehler(fehler: Anmeldefehler, rueckfall: string): string {
  // Zu viele Versuche kommt **ohne** Code zurück — die Bremse des Dienstes
  // sitzt vor den Endpunkten und kennt deren Fehlerliste nicht, sie schickt
  // nur 429 und einen englischen Satz. Ohne diesen Zweig wäre ausgerechnet
  // der häufigste Fehler einer Probewoche (dreimal das falsche Passwort) der
  // eine, der englisch bleibt.
  if (fehler.status === 429) {
    return "Zu viele Versuche. Warte einen Moment und probier es noch einmal.";
  }

  const satz = fehler.code ? saetze[fehler.code as Fehlercode] : undefined;
  return satz ?? fehler.message ?? rueckfall;
}
