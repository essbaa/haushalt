import { headers } from "next/headers";

/**
 * Holt serverseitig ein Token für den Go-Dienst.
 *
 * Der Weg ist etwas umständlich und hat einen Grund: Die Sitzung steckt in
 * einem Cookie, das dem Browser gehört. Eine Server Component sieht es nur,
 * weil Next die Kopfzeilen der eingehenden Anfrage bereitstellt — also reichen
 * wir das Cookie an unseren eigenen Anmelde-Endpunkt weiter und bekommen ein
 * signiertes Token zurück, das der Go-Dienst prüfen kann.
 *
 * Kein Token heißt nicht „Fehler", sondern „nicht angemeldet". Dann sieht der
 * Aufrufer die Demo-Haushalte, und das ist gewollt.
 */
export async function serverToken(): Promise<string | null> {
  const cookie = (await headers()).get("cookie");
  if (!cookie) return null;

  const base = process.env.BETTER_AUTH_URL ?? "http://localhost:3000";
  try {
    const res = await fetch(`${base}/api/auth/token`, {
      headers: { cookie },
      cache: "no-store",
    });
    if (!res.ok) return null;
    const daten = (await res.json()) as { token?: string };
    return daten.token ?? null;
  } catch {
    return null;
  }
}
