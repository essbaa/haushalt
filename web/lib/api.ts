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
export const API_BASE = process.env.NEXT_PUBLIC_API_URL ?? "";

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
