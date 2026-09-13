/**
 * Zugriff auf den Go-Dienst.
 *
 * Die Adresse kommt aus NEXT_PUBLIC_API_URL. Alles mit diesem Präfix wird beim
 * Bauen fest in das Browser-Bundle geschrieben — es ist damit öffentlich
 * sichtbar. Hier steht deshalb nur eine URL und niemals ein Schlüssel.
 *
 * Wichtig für Vercel: Der Wert wird zur BAUZEIT eingesetzt, nicht zur Laufzeit.
 * Wer ihn im Dashboard ändert, muss neu deployen, sonst zeigt die Seite
 * weiterhin die alte Adresse.
 */
import type { components } from "./api-types";

export const API_BASE = process.env.NEXT_PUBLIC_API_URL ?? "";

/**
 * Die Typen des Vertrags — erzeugt aus ../openapi.yaml, nicht getippt.
 *
 * `npm run generate` schreibt lib/api-types.ts neu; die CI prüft mit
 * `git diff --exit-code`, dass die Datei zur Spezifikation passt. Wer im
 * Go-Dienst ein Feld umbenennt, ohne die Spezifikation zu ändern, merkt es
 * spätestens hier — und wer beides ändert, bekommt hier den Übersetzungsfehler
 * statt eines leeren Feldes im Browser.
 */
export type Wochenplan = components["schemas"]["Wochenplan"];
export type Haushalt = components["schemas"]["Haushalt"];
export type Aufgabe = components["schemas"]["Aufgabe"];
export type Bilanz = components["schemas"]["Bilanz"];
export type Uebersprungen = components["schemas"]["Uebersprungen"];
export type Ich = components["schemas"]["Ich"];
export type NeuerHaushalt = components["schemas"]["NeuerHaushalt"];
export type NeuesMitglied = components["schemas"]["NeuesMitglied"];

/** Antwort von GET /healthz — spiegelt healthResponse im Go-Dienst. */
export type Health = {
  status: string;
  version: string;
  env: string;
  database: string;
};

export class ApiError extends Error {
  constructor(
    message: string,
    readonly hint?: string,
  ) {
    super(message);
    this.name = "ApiError";
  }
}

/**
 * Fragt den Go-Dienst.
 *
 * Bewusst vom Browser aus und nicht auf dem Next.js-Server: Nur so läuft die
 * Anfrage wirklich über die Sprachgrenze, mit CORS und allem. Ein Abruf auf
 * dem Server würde CORS gar nicht berühren und die Hälfte dessen verstecken,
 * was diese Seite beweisen soll.
 */
export async function fetchHealth(signal?: AbortSignal): Promise<Health> {
  if (!API_BASE) {
    throw new ApiError(
      "NEXT_PUBLIC_API_URL ist nicht gesetzt",
      "In web/.env.local eintragen und den Entwicklungsserver neu starten.",
    );
  }

  let res: Response;
  try {
    res = await fetch(`${API_BASE}/healthz`, { signal, cache: "no-store" });
  } catch (cause) {
    // fetch wirft bei Netzfehlern UND bei blockiertem CORS — der Browser
    // verrät den Grund aus Sicherheitsgründen nicht. Deshalb der Hinweis.
    if (cause instanceof DOMException && cause.name === "AbortError") throw cause;
    throw new ApiError(
      `${API_BASE} ist nicht erreichbar`,
      "Läuft der Dienst? Und steht diese Adresse in ALLOWED_ORIGINS?",
    );
  }

  if (!res.ok && res.status !== 503) {
    throw new ApiError(`Der Dienst antwortet mit ${res.status}`);
  }

  return (await res.json()) as Health;
}

/**
 * Ruft den Dienst auf dem Next.js-Server ab, nicht im Browser.
 *
 * Anders als die Statuszeile: Der Wochenplan ist Inhalt, und Inhalt soll da
 * sein, bevor die Seite ankommt — ohne Ladezustand, ohne Springen, und lesbar
 * auch ohne JavaScript. Die Statuszeile bleibt trotzdem im Browser, weil sie
 * genau das beweisen soll, was hier bewusst umgangen wird: die CORS-Grenze.
 *
 * Zwischengespeichert wird nur, was öffentlich ist — siehe unten. Angemeldete
 * Anfragen tragen ein Token und werden nie gespeichert.
 */
async function hole<T>(pfad: string, token?: string | null): Promise<T> {
  if (!API_BASE) {
    throw new ApiError(
      "NEXT_PUBLIC_API_URL ist nicht gesetzt",
      "In web/.env.local eintragen und den Entwicklungsserver neu starten.",
    );
  }

  // Mit Token ist die Antwort persönlich und darf NICHT zwischengespeichert
  // werden — sonst bekäme der nächste Besucher den Plan des vorigen zu sehen.
  // Ohne Token ist es die öffentliche Demo, und fünf Minuten Puffer ersparen
  // dem schlafenden Fly-Dienst jeden Aufruf.
  const wie: RequestInit = token
    ? { headers: { Authorization: `Bearer ${token}` }, cache: "no-store" }
    : { next: { revalidate: 300 } };

  let res: Response;
  try {
    res = await fetch(`${API_BASE}${pfad}`, wie);
  } catch {
    throw new ApiError(`${API_BASE} ist nicht erreichbar`, "Läuft der Go-Dienst?");
  }

  if (!res.ok) {
    // Der Vertrag sagt zu, dass auch Fehler JSON sind. Falls nicht, ist der
    // Statuscode immer noch besser als eine leere Meldung.
    let text = `Der Dienst antwortet mit ${res.status}`;
    try {
      const fehler = (await res.json()) as components["schemas"]["Fehler"];
      if (fehler?.fehler) text = fehler.fehler;
    } catch {
      /* keine JSON-Antwort — dann bleibt der Statuscode die Auskunft */
    }
    throw new ApiError(text);
  }

  return (await res.json()) as T;
}

/** Wer fragt — soweit das Token es belegt. Ohne Token: niemand. */
export function ladeIch(token?: string | null): Promise<Ich> {
  return hole<Ich>("/api/ich", token);
}

export function ladeHaushalte(token?: string | null): Promise<Haushalt[]> {
  return hole<Haushalt[]>("/api/haushalte", token);
}

export function ladePlan(
  haushaltId: string,
  woche: string,
  token?: string | null,
): Promise<Wochenplan> {
  return hole<Wochenplan>(
    `/api/haushalte/${encodeURIComponent(haushaltId)}/plan/${encodeURIComponent(woche)}`,
    token,
  );
}
