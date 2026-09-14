import { redirect } from "next/navigation";
import { Bereiche } from "@/app/components/bereiche";
import { TermineListe } from "@/app/components/termine-liste";
import { ladeAnlaesse, ladeHaushalte } from "@/lib/api";
import { serverToken } from "@/lib/auth-token";

/**
 * Anlässe: Geburtstage, Elternabend, Termine.
 *
 * Der Grund für diese Seite steht in ADR-0010. „Geschenk für Kindergeburtstag
 * besorgen" hatte vorher den Rhythmus „alle 45 Tage" und erfand damit
 * regelmäßig einen Geburtstag. Was hier steht, ist echt — und nur daran hängen
 * die Aufgaben mit Vorlauf.
 */
export default async function Termine({
  searchParams,
}: {
  searchParams: Promise<{ haushalt?: string }>;
}) {
  const token = await serverToken();
  if (!token) redirect("/anmelden?weiter=/termine");

  const haushalte = await ladeHaushalte(token);
  const eigene = haushalte.filter((h) => h.meine_rolle);
  if (eigene.length === 0) redirect("/einrichten");

  const { haushalt } = await searchParams;
  const gewaehlt = eigene.find((h) => h.id === haushalt) ?? eigene[0];
  const anlaesse = await ladeAnlaesse(gewaehlt.id, token);

  return (
    <main className="mx-auto w-full max-w-lg px-5 py-10">
      <Bereiche
        haushaltId={gewaehlt.id}
        planend={gewaehlt.meine_rolle === "planend"}
        aktiv="termine"
      />

      <h1 className="mt-5 mb-2 text-2xl font-extrabold tracking-tight text-balance">
        Anlässe
      </h1>
      <p className="mb-8 max-w-prose text-sm leading-relaxed text-muted text-pretty">
        Geburtstage, Elternabende, Termine. Manche Aufgaben entstehen nur daraus
        — ein Geschenk braucht einen Geburtstag. Erfinden wird die App keinen.
      </p>

      <TermineListe
        haushaltId={gewaehlt.id}
        anlaesse={anlaesse}
        planend={gewaehlt.meine_rolle === "planend"}
      />
    </main>
  );
}
