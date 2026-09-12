import Link from "next/link";
import { ApiStatus } from "@/app/components/api-status";
import { Sitzung } from "@/app/components/sitzung";
import { Wochenplan } from "@/app/components/wochenplan";
import { ApiError, ladeHaushalte, ladePlan, type Haushalt, type Wochenplan as Plan } from "@/lib/api";
import { aktuelleWoche, istWoche } from "@/lib/woche";

/**
 * Startseite: der Wochenplan eines Beispielhaushalts.
 *
 * Server Component — der Plan wird auf dem Next.js-Server geholt und fertig
 * ausgeliefert. Kein Ladezustand, kein Springen, lesbar auch ohne JavaScript.
 *
 * Die Statuszeile ganz unten bleibt dagegen eine Client Component: Sie fragt
 * den Dienst aus dem Browser und misst damit die CORS-Grenze, die der
 * Server-Abruf gerade nicht berührt.
 */
export default async function Page({
  searchParams,
}: {
  searchParams: Promise<{ haushalt?: string; woche?: string }>;
}) {
  const params = await searchParams;
  const woche = params.woche && istWoche(params.woche) ? params.woche : aktuelleWoche();

  let haushalte: Haushalt[] = [];
  let plan: Plan | null = null;
  let fehler: { text: string; hinweis?: string } | null = null;

  try {
    haushalte = await ladeHaushalte();
    const gewaehlt =
      haushalte.find((h) => h.id === params.haushalt)?.id ?? haushalte[0]?.id;
    if (gewaehlt) plan = await ladePlan(gewaehlt, woche);
  } catch (e) {
    // Der Dienst schläft (Fly fährt bei Ruhe herunter) oder ist kaputt. Beides
    // gehört gesagt, nicht in eine leere Seite verwandelt.
    fehler =
      e instanceof ApiError
        ? { text: e.message, hinweis: e.hint }
        : { text: "Der Wochenplan konnte nicht geladen werden." };
  }

  return (
    <main className="mx-auto flex w-full max-w-3xl flex-1 flex-col gap-10 px-6 py-12">
      <header className="space-y-3">
        <div className="flex flex-wrap items-baseline justify-between gap-3">
          <p className="font-mono text-xs uppercase tracking-widest text-muted">
            Haushalt als Team
          </p>
          <Sitzung />
        </div>
        <h1 className="text-4xl font-semibold tracking-tight text-balance">
          Der Wochenplan
        </h1>
        <p className="text-lg leading-relaxed text-muted text-pretty">
          Verteilt wird nicht nur Zeit, sondern auch Kopflast — die Arbeit, an
          die jemand denken muss. Wer weniger Zeit hat, trägt weniger.
        </p>
      </header>

      {haushalte.length > 1 && (
        <nav className="flex flex-wrap gap-2" aria-label="Haushalt wählen">
          {haushalte.map((h) => {
            const aktiv = h.id === plan?.haushalt.id;
            return (
              <Link
                key={h.id}
                href={`/?haushalt=${h.id}`}
                aria-current={aktiv ? "page" : undefined}
                className={`rounded-full border px-4 py-1.5 text-sm transition-colors ${
                  aktiv
                    ? "border-accent bg-accent/10 text-accent"
                    : "border-line text-muted hover:text-foreground"
                }`}
              >
                {h.name}
                <span className="ml-2 text-xs text-muted">
                  {h.mitglieder.length} Personen
                </span>
              </Link>
            );
          })}
        </nav>
      )}

      {fehler && (
        <div className="rounded-lg border border-line bg-surface p-5">
          <p className="font-medium">{fehler.text}</p>
          {fehler.hinweis && (
            <p className="mt-1 text-sm text-muted">{fehler.hinweis}</p>
          )}
        </div>
      )}

      {plan && <Wochenplan plan={plan} />}

      <section className="space-y-3 border-t border-line pt-8">
        <h2 className="font-mono text-xs uppercase tracking-widest text-muted">
          Verbindung
        </h2>
        <ApiStatus />
        <p className="text-sm leading-relaxed text-muted">
          Der Plan oben kommt vom Go-Dienst, berechnet in einem reinen,
          deterministischen Paket. Diese Zeile fragt denselben Dienst noch
          einmal — aus dem Browser, über die CORS-Grenze hinweg.
        </p>
      </section>
    </main>
  );
}
