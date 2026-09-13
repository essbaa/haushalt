import Link from "next/link";
import { redirect } from "next/navigation";
import { EinstellungenFormular } from "@/app/components/einstellungen-formular";
import { ladeHaushalte, type Haushalt } from "@/lib/api";
import { serverToken } from "@/lib/auth-token";
import { aktuelleWoche } from "@/lib/woche";

/**
 * Was das Onboarding geraten hat, hier korrigieren.
 *
 * Server Component als Rahmen und Weiche: Ohne Anmeldung zum Anmelden, ohne
 * eigenen Haushalt ins Onboarding. Welche Rolle der Aufrufer hat, entscheidet
 * nicht diese Seite, sondern der Dienst — hier steuert sie nur, was angezeigt
 * wird.
 */
export default async function Einstellungen({
  searchParams,
}: {
  searchParams: Promise<{ haushalt?: string }>;
}) {
  const token = await serverToken();
  if (!token) redirect("/anmelden?weiter=/einstellungen");

  const haushalte: Haushalt[] = await ladeHaushalte(token);
  const eigene = haushalte.filter((h) => h.meine_rolle);
  if (eigene.length === 0) redirect("/einrichten");

  const { haushalt } = await searchParams;
  const gewaehlt = eigene.find((h) => h.id === haushalt) ?? eigene[0];

  return (
    <main className="mx-auto w-full max-w-lg px-6 py-16">
      <Link href={`/?haushalt=${gewaehlt.id}`} className="text-sm text-muted underline">
        Zurück zum Wochenplan
      </Link>

      <h1 className="mt-6 mb-1 text-2xl font-semibold tracking-tight">
        Einstellungen
      </h1>
      <p className="mb-8 text-sm leading-relaxed text-muted">
        Beim Einrichten hat die App einiges geraten — Kapazität, Betreuungsform,
        Wohnform. Hier steht, was sie angenommen hat, und hier änderst du es.
      </p>

      <EinstellungenFormular
        haushalt={gewaehlt}
        planend={gewaehlt.meine_rolle === "planend"}
        woche={aktuelleWoche()}
      />
    </main>
  );
}
