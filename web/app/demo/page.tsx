import Link from "next/link";
import { Uhr } from "@/app/components/icons";
import { Wochenplan } from "@/app/components/wochenplan";
import {
  ApiError,
  ladeHaushalte,
  ladePlan,
  type Haushalt,
  type Wochenplan as Plan,
} from "@/lib/api";
import { aktuelleWoche, istWoche, wochenSpanne } from "@/lib/woche";

/**
 * Ein echter Plan, ohne Konto.
 *
 * Die Beispielhaushalte standen vorher auf der Startseite — und damit sah
 * jeder, der die Adresse zum ersten Mal öffnete, den Wochenplan von
 * „Familie A" und musste sich selbst zusammenreimen, was das soll. Aus dem
 * stärksten Argument des Produkts war ein Rätsel geworden.
 *
 * Hier stehen sie richtig: einen Klick hinter einem Satz, der sagt, was man
 * sieht. Gerechnet wird mit demselben Planer wie für echte Haushalte — die
 * Zahlen sind nicht gestellt, die Familien schon.
 *
 * Ohne Token: Der Dienst liefert dann die Haushalte mit Slug, und mehr gibt
 * es hier auch nicht zu sehen. Abhaken oder umverteilen geht nicht, weil
 * gerechnete Wochen keine Aufgabenkennungen haben (ADR-0008) — das steht im
 * Streifen oben, statt dass jemand vergeblich tippt.
 */
export default async function Demo({
  searchParams,
}: {
  searchParams: Promise<{ haushalt?: string; woche?: string }>;
}) {
  const params = await searchParams;
  const woche = params.woche && istWoche(params.woche) ? params.woche : aktuelleWoche();

  let haushalte: Haushalt[] = [];
  let plan: Plan | null = null;
  let fehler: string | null = null;

  try {
    haushalte = await ladeHaushalte();
    const gewaehlt = haushalte.find((h) => h.id === params.haushalt)?.id ?? haushalte[0]?.id;
    if (gewaehlt) plan = await ladePlan(gewaehlt, woche, null);
  } catch (e) {
    fehler = e instanceof ApiError ? e.message : "Der Beispielplan ist gerade nicht erreichbar.";
  }

  return (
    <main className="mx-auto w-full max-w-2xl px-5 py-8">
      <div className="mb-8 flex gap-3 rounded-lg border border-clay/40 bg-clay-soft px-4 py-3.5">
        <Uhr className="mt-0.5 size-5 shrink-0 text-clay" aria-hidden="true" />
        <div className="space-y-1">
          <p className="font-bold text-clay text-pretty">
            Beispielhaushalt — echte Rechnung, erfundene Familie.
          </p>
          <p className="text-sm leading-relaxed text-muted text-pretty">
            Dieser Plan kommt aus demselben Planer wie jeder andere. Abhaken und umverteilen geht
            hier nicht: Dafür müsste die Woche einem Haushalt gehören.{" "}
            <Link href="/" className="font-semibold text-primary underline underline-offset-2">
              Zurück zur Übersicht
            </Link>
          </p>
        </div>
      </div>

      {haushalte.length > 1 && (
        <nav aria-label="Beispielhaushalt wählen" className="mb-7 flex flex-wrap gap-2">
          {haushalte.map((h) => {
            const aktiv = h.id === plan?.haushalt.id;
            return (
              <Link
                key={h.id}
                href={`/demo?haushalt=${encodeURIComponent(h.id)}`}
                aria-current={aktiv ? "page" : undefined}
                className={`inline-flex min-h-10 items-center rounded-full border px-4 text-sm font-semibold transition-colors ${
                  aktiv
                    ? "border-primary bg-primary-soft text-primary"
                    : "border-line-strong text-muted hover:text-fg"
                }`}
              >
                {h.name}
              </Link>
            );
          })}
        </nav>
      )}

      <h1 className="text-3xl font-extrabold tracking-tight text-balance">
        {plan?.haushalt.name ?? "Beispielhaushalt"}
      </h1>
      <p className="mt-1 mb-8 text-sm tabular-nums text-muted">{wochenSpanne(woche)}</p>

      {fehler && (
        <p className="mb-8 rounded-lg border border-line bg-surface px-4 py-3 font-semibold">
          {fehler}
        </p>
      )}

      {plan && <Wochenplan plan={plan} />}

      <div className="mt-12 border-t border-line pt-6">
        <p className="text-sm leading-relaxed text-muted text-pretty">
          So sieht eine Woche aus, wenn sie gerechnet wird statt ausgehandelt. Für einen eigenen
          Haushalt brauchst du einen Zugangscode — die App ist in Erprobung.
        </p>
        <Link
          href="/anmelden"
          className="mt-4 inline-flex min-h-12 items-center rounded-md bg-primary px-6 text-sm font-semibold text-on-primary transition-colors hover:bg-primary-hover"
        >
          Konto anlegen
        </Link>
      </div>
    </main>
  );
}
