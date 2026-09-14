import Link from "next/link";
import { redirect } from "next/navigation";
import { Bereiche } from "@/app/components/bereiche";
import { Uhr } from "@/app/components/icons";
import { Laedt } from "@/app/components/laedt";
import { Sitzung } from "@/app/components/sitzung";
import { Wochenplan } from "@/app/components/wochenplan";
import { ApiError, ladeHaushalte, ladePlan, type Haushalt, type Wochenplan as Plan } from "@/lib/api";
import { serverToken } from "@/lib/auth-token";
import { aktuelleWoche, istWoche, wocheVersetzt, wochenSpanne } from "@/lib/woche";

/**
 * Die Startseite: der Wochenplan.
 *
 * Server Component — der Plan wird auf dem Next.js-Server geholt und fertig
 * ausgeliefert. Kein Ladezustand, kein Springen, lesbar auch ohne JavaScript.
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

  const token = await serverToken();

  try {
    haushalte = await ladeHaushalte(token);
    const gewaehlt = haushalte.find((h) => h.id === params.haushalt)?.id ?? haushalte[0]?.id;
    if (gewaehlt) plan = await ladePlan(gewaehlt, woche, token);
  } catch (e) {
    fehler =
      e instanceof ApiError
        ? { text: e.message, hinweis: e.hint }
        : { text: "Der Wochenplan konnte nicht geladen werden." };
  }

  // Angemeldet, aber noch ohne eigenen Haushalt: Zeit fürs Einrichten.
  // Außerhalb des try, weil redirect() intern eine Ausnahme wirft — im catch
  // oben würde sie als Ladefehler enden.
  if (token && !params.haushalt && !haushalte.some((h) => h.meine_rolle)) {
    redirect("/einrichten");
  }

  const eigener = plan?.meine_rolle !== undefined;
  const jetzt = aktuelleWoche();
  const gewaehlterHaushalt = plan?.haushalt.id ?? params.haushalt;

  return (
    <>
      <header className="sticky top-0 z-10 border-b border-line bg-bg/90 backdrop-blur">
        <div className="mx-auto flex w-full max-w-2xl items-center justify-between gap-3 px-5 py-3">
          <Link href="/" className="text-base font-extrabold tracking-tight">
            Haushalt
          </Link>
          <Sitzung />
        </div>
      </header>

      <main className="mx-auto w-full max-w-2xl flex-1 px-5 py-8">
        {haushalte.length > 1 && (
          <nav aria-label="Haushalt wählen" className="mb-7 flex flex-wrap gap-2">
            {haushalte.map((h) => {
              const aktiv = h.id === plan?.haushalt.id;
              return (
                <Link
                  key={h.id}
                  href={`/?haushalt=${h.id}`}
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

        <div className="mb-5 space-y-2">
          <h1 className="text-3xl font-extrabold tracking-tight text-balance">
            {plan?.haushalt.name ?? "Der Wochenplan"}
          </h1>

          {/* Bis hierher gab es keinen Weg zur nächsten Woche: `?woche=` wurde
              gelesen, aber von keinem Link gesetzt. Ein Wochenplan, der nur
              diese eine Woche zeigt, ist nach sieben Tagen zu Ende. */}
          <div className="-ml-2 flex flex-wrap items-center gap-1">
            <Link
              href={adresse(gewaehlterHaushalt, wocheVersetzt(woche, -1))}
              aria-label="Woche davor"
              className="inline-flex size-11 items-center justify-center rounded-full text-lg text-muted transition-colors hover:bg-surface-2 hover:text-fg"
            >
              <Laedt>
                <span aria-hidden="true">‹</span>
              </Laedt>
            </Link>
            <span className="text-sm tabular-nums text-muted">{wochenSpanne(woche)}</span>
            <Link
              href={adresse(gewaehlterHaushalt, wocheVersetzt(woche, 1))}
              aria-label="Woche danach"
              className="inline-flex size-11 items-center justify-center rounded-full text-lg text-muted transition-colors hover:bg-surface-2 hover:text-fg"
            >
              <Laedt>
                <span aria-hidden="true">›</span>
              </Laedt>
            </Link>
            {woche !== jetzt && (
              <Link
                href={adresse(gewaehlterHaushalt, jetzt)}
                className="ml-1 inline-flex min-h-9 items-center rounded-full px-3 text-sm font-semibold text-primary transition-colors hover:bg-primary-soft"
              >
                <span className="inline-flex items-center gap-2">
                  Diese Woche
                  <Laedt />
                </span>
              </Link>
            )}
          </div>

        </div>

        {/* Eine künftige Woche wird nicht festgeschrieben (ADR-0008,
            storage.Plan). Das stand zuerst als graue Kleinschrift unter den
            Pfeilen — und wurde übersehen. Eine Fußnote ist keine Ansage: Der
            Unterschied zwischen „Plan" und „Vermutung" ist das Wichtigste auf
            dieser Seite, solange er gilt.
            Er erklärt gleich mit, warum hier nichts abzuhaken ist — ohne
            festgeschriebene Woche gibt es keine Aufgabenkennungen. Eine
            Oberfläche, die etwas weglässt, muss sagen warum. */}
        {woche > jetzt && (
          <div className="mb-8 flex gap-3 rounded-lg border border-clay/40 bg-clay-soft px-4 py-3.5">
            <Uhr className="mt-0.5 size-5 shrink-0 text-clay" />
            <div className="space-y-1">
              <p className="font-bold text-clay text-pretty">
                Vorschau — diese Woche steht noch nicht fest.
              </p>
              <p className="text-sm leading-relaxed text-muted text-pretty">
                Festgeschrieben wird sie am Montag. Bis dahin ändert jede
                Einstellung sie noch, und abhaken oder umverteilen geht erst
                dann.
              </p>
            </div>
          </div>
        )}

        {eigener && plan && (
          <div className="mb-8">
            <Bereiche
              haushaltId={plan.haushalt.id}
              planend={plan.meine_rolle === "planend"}
              aktiv="plan"
            />
          </div>
        )}

        {fehler && (
          <div className="mb-8 rounded-lg border border-line bg-surface px-4 py-3">
            <p className="font-semibold">{fehler.text}</p>
            {fehler.hinweis && <p className="mt-1 text-sm text-muted">{fehler.hinweis}</p>}
          </div>
        )}

        {plan && <Wochenplan plan={plan} />}

        {!token && (
          <p className="mt-10 border-t border-line pt-6 text-sm leading-relaxed text-muted">
            Das ist ein Beispielhaushalt.{" "}
            <Link href="/anmelden" className="font-semibold text-primary underline underline-offset-2">
              Melde dich an
            </Link>
            , um einen eigenen anzulegen.
          </p>
        )}
      </main>
    </>
  );
}

/** Die Adresse des Wochenplans — mit Haushalt, wenn einer gewählt ist. */
function adresse(haushalt: string | undefined, woche: string): string {
  const teile = new URLSearchParams();
  if (haushalt) teile.set("haushalt", haushalt);
  teile.set("woche", woche);
  return `/?${teile.toString()}`;
}
