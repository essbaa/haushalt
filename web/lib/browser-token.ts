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
export async function postMitToken<T>(pfad: string, rumpf?: unknown): Promise<T> {
  const t = await token();

  const res = await fetch(`${API_BASE}${pfad}`, {
    method: "POST",
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
  return (await res.json()) as T;
}
