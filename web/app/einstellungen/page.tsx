import Link from "next/link";
import { redirect } from "next/navigation";
import { Bereiche } from "@/app/components/bereiche";
import { EinstellungenFormular } from "@/app/components/einstellungen-formular";
import { PersonPlus } from "@/app/components/icons";
import { Laedt } from "@/app/components/laedt";
import { Melden } from "@/app/components/melden";
import { RueckmeldungenListe } from "@/app/components/rueckmeldungen-liste";
import { ladeHaushalte, ladeRueckmeldungen, type Haushalt, type Rueckmeldung } from "@/lib/api";
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

  // Gemeldetes sehen nur die planenden Personen. Fehlschlagen darf das nicht
  // die Seite: Die Einstellungen sind der Ort, an dem man etwas repariert —
  // ausgerechnet hier an einer Nebensache zu scheitern wäre die schlechteste
  // aller Stellen.
  let gemeldet: Rueckmeldung[] = [];
  if (gewaehlt.meine_rolle === "planend") {
    try {
      gemeldet = await ladeRueckmeldungen(gewaehlt.id, token);
    } catch {
      gemeldet = [];
    }
  }

  return (
    <main className="mx-auto w-full max-w-lg px-5 py-10">
      <Bereiche
        haushaltId={gewaehlt.id}
        planend={gewaehlt.meine_rolle === "planend"}
        aktiv="einstellungen"
      />

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

      {gewaehlt.meine_rolle === "planend" && (
        <section className="mt-10 border-t border-line pt-6">
          <h2 className="text-lg font-bold tracking-tight">Was gemeldet wurde</h2>
          <p className="mt-1 mb-4 max-w-prose text-sm leading-relaxed text-muted text-pretty">
            Unten auf jeder Seite steht &bdquo;Stimmt etwas nicht?&ldquo;. Das darf jeder im
            Haushalt benutzen — und in der Probewoche ist es die wichtigste
            Zeile der App.
          </p>
          <RueckmeldungenListe liste={gemeldet} />
        </section>
      )}

      {/* Einladen hat die Bereichsleiste verlassen: Es ist eine Handlung und
          kein Ort, und man tut es ein-, zweimal im Leben eines Haushalts.
          Hier steht es richtig — bei den anderen Dingen, die man einmal
          einstellt. */}
      {gewaehlt.meine_rolle === "planend" && (
        <section className="mt-10 border-t border-line pt-6">
          <h2 className="text-lg font-bold tracking-tight">Wer noch dazugehört</h2>
          <p className="mt-1 mb-4 max-w-prose text-sm leading-relaxed text-muted text-pretty">
            Jede Person im Haushalt kann einen eigenen Zugang bekommen und sieht
            dann ihren Teil des Plans — und kann abhaken, abgeben, widersprechen.
          </p>
          <Link
            href={`/einladen?haushalt=${encodeURIComponent(gewaehlt.id)}`}
            className="inline-flex min-h-11 items-center gap-2 rounded-md border border-line-strong px-4 text-sm font-semibold transition-colors hover:border-primary hover:text-primary"
          >
            <PersonPlus className="size-4" />
            Jemanden einladen
            <Laedt />
          </Link>
        </section>
      )}

      <Melden haushaltId={gewaehlt.id} />
    </main>
  );
}
