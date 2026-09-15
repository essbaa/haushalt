"use client";

import { API_BASE, ApiError } from "./api";

/**
 * Holt im Browser ein Token für den Go-Dienst.
 *
 * Der Aufruf geht an die eigene Adresse, das Sitzungs-Cookie fährt automatisch
 * mit. Erst der zweite Aufruf — der an den Go-Dienst — überquert eine Grenze
 * und braucht deshalb das Token im Kopf.
 */
async function token(): Promise<string> {
  const res = await fetch("/api/auth/token", { cache: "no-store" });
  if (!res.ok) {
    throw new ApiError("Du bist nicht angemeldet.");
  }
  const daten = (await res.json()) as { token?: string };
  if (!daten.token) {
    throw new ApiError("Du bist nicht angemeldet.");
  }
  return daten.token;
}

/** Ein POST an den Go-Dienst, angemeldet. */
export function postMitToken<T>(pfad: string, rumpf?: unknown): Promise<T> {
  return sendeMitToken<T>("POST", pfad, rumpf);
}

/** Ein PATCH — eine Änderung, bei der alles Weggelassene stehen bleibt. */
export function patchMitToken<T>(pfad: string, rumpf?: unknown): Promise<T> {
  return sendeMitToken<T>("PATCH", pfad, rumpf);
}

/** Ein PUT — der ganze Zustand auf einmal, nicht ein Stück davon. */
export function putMitToken<T>(pfad: string, rumpf?: unknown): Promise<T> {
  return sendeMitToken<T>("PUT", pfad, rumpf);
}

/** Ein DELETE. */
export function deleteMitToken<T>(pfad: string): Promise<T> {
  return sendeMitToken<T>("DELETE", pfad);
}

async function sendeMitToken<T>(methode: string, pfad: string, rumpf?: unknown): Promise<T> {
  const t = await token();

  const res = await fetch(`${API_BASE}${pfad}`, {
    method: methode,
    headers: {
      Authorization: `Bearer ${t}`,
      ...(rumpf ? { "Content-Type": "application/json" } : {}),
    },
    body: rumpf ? JSON.stringify(rumpf) : undefined,
  });

  if (!res.ok) {
    let text = `Der Dienst antwortet mit ${res.status}`;
    try {
      const fehler = (await res.json()) as { fehler?: string };
      if (fehler?.fehler) text = fehler.fehler;
    } catch {
      /* keine JSON-Antwort */
    }
    throw new ApiError(text);
  }

  // 204 heißt „hat geklappt, gibt nichts zu sagen" — etwa beim Abhaken. Ein
  // res.json() darauf scheitert mit „Unexpected end of JSON input", und der
  // Nutzer sieht einen Fehler, obwohl alles gut ging. Deshalb wird hier die
  // leere Antwort geprüft und nicht geraten.
  if (res.status === 204) {
    return undefined as T;
  }
  return (await res.json()) as T;
}
