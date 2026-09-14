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
    <main className="mx-auto w-full max-w-lg px-5 py-10">
      <Link href={`/?haushalt=${gewaehlt.id}`} className="inline-flex min-h-11 items-center gap-1.5 text-sm text-muted transition-colors hover:text-fg">
        <span aria-hidden="true">←</span>
        Zurück zum Wochenplan
      </Link>

      <h1 className="mt-5 mb-2 text-2xl font-extrabold tracking-tight text-balance">
        Einstellungen
      </h1>
      <p className="mb-8 max-w-prose text-sm leading-relaxed text-muted text-pretty">
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
