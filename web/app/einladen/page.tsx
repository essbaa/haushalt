import { Bereiche } from "@/app/components/bereiche";
import { EinladenFormular, type Offene } from "@/app/components/einladen-formular";
import { ladeHaushalte } from "@/lib/api";
import { serverToken } from "@/lib/auth-token";

/**
 * Jemanden in den Haushalt einladen.
 *
 * Server Component als Rahmen: Sie liest den Haushalt aus der Adresse und holt
 * dessen Mitglieder, damit das Formular anbieten kann, *wen* die Einladung
 * betrifft. Ohne diese Liste legte ein Beitritt immer eine zweite Person
 * gleichen Namens an.
 */
export default async function Einladen({
  searchParams,
}: {
  searchParams: Promise<{ haushalt?: string }>;
}) {
  const { haushalt } = await searchParams;
  const token = await serverToken();

  let offene: Offene[] = [];
  let planend = false;
  if (haushalt && token) {
    const haushalte = await ladeHaushalte(token);
    const gewaehlt = haushalte.find((h) => h.id === haushalt);
    offene = (gewaehlt?.mitglieder ?? [])
      // Wer schon ein Konto hat, braucht keine Einladung; wer betreut wird,
      // meldet sich nicht an — ein Zweijähriger erzeugt Arbeit ohne Konto.
      .filter((m) => !m.hat_zugang && m.rolle !== "betreut")
      .map((m) => ({ id: m.id, name: m.name, rolle: m.rolle }));
    planend = gewaehlt?.meine_rolle === "planend";
  }

  return (
    <main className="mx-auto w-full max-w-md px-5 py-10">
      <Bereiche haushaltId={haushalt} planend={planend} aktiv="einladen" />

      <h1 className="mt-5 mb-2 text-2xl font-extrabold tracking-tight text-balance">
        Jemanden einladen
      </h1>
      <p className="mb-8 max-w-prose text-sm leading-relaxed text-muted text-pretty">
        Du bekommst einen Code, der einmal gilt und nach einer Woche abläuft.
        Wer schon im Plan steht, wird mit seinem Konto verbunden — Aufgaben und
        Verlauf bleiben ihm.
      </p>

      {haushalt ? (
        <EinladenFormular haushaltId={haushalt} offene={offene} />
      ) : (
        <p className="text-sm text-clay">
          Es fehlt der Haushalt in der Adresse. Geh über den Wochenplan hierher.
        </p>
      )}
    </main>
  );
}
