import Link from "next/link";
import { redirect } from "next/navigation";
import { Bereiche } from "@/app/components/bereiche";
import { Sitzung } from "@/app/components/sitzung";
import { Wochenplan } from "@/app/components/wochenplan";
import { ApiError, ladeHaushalte, ladePlan, type Haushalt, type Wochenplan as Plan } from "@/lib/api";
import { serverToken } from "@/lib/auth-token";
import { aktuelleWoche, istWoche, wochenSpanne } from "@/lib/woche";

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

        <div className="mb-5 space-y-1">
          <h1 className="text-3xl font-extrabold tracking-tight text-balance">
            {plan?.haushalt.name ?? "Der Wochenplan"}
          </h1>
          <p className="text-sm text-muted">{wochenSpanne(woche)}</p>
        </div>

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
