import Link from "next/link";
import { Laedt } from "@/app/components/laedt";

/**
 * Die Bereiche der App.
 *
 * Vorher standen diese Links unter dem Wochenplan — also hinter zwei Dutzend
 * Aufgaben, in einer App, deren Hauptseite lang ist. Damit war die Navigation
 * der am schwersten erreichbare Teil der Oberfläche, und das ist sie genau
 * verkehrt herum.
 *
 * Nicht klebend: Ein zweiter fester Balken kostet auf einem Telefon dauerhaft
 * 44 Pixel, und der Grund, den Bereich zu wechseln, entsteht beim Ankommen,
 * nicht mitten im Scrollen. Auf den Unterseiten ersetzt der Streifen den
 * „Zurück zum Wochenplan"-Link — der Plan ist dort der erste Eintrag und
 * damit derselbe Weg, nur ohne zweite Schreibweise für dieselbe Sache.
 *
 * Server Component: nur Links, kein Zustand.
 */
export type Bereich = "plan" | "vorlagen" | "termine" | "einstellungen" | "einladen";

const bereiche: { id: Bereich; titel: string; pfad: string; nurPlanend?: boolean }[] = [
  { id: "plan", titel: "Plan", pfad: "/" },
  { id: "vorlagen", titel: "Aufgaben", pfad: "/vorlagen", nurPlanend: true },
  { id: "termine", titel: "Anlässe", pfad: "/termine", nurPlanend: true },
  { id: "einstellungen", titel: "Einstellungen", pfad: "/einstellungen" },
  { id: "einladen", titel: "Einladen", pfad: "/einladen", nurPlanend: true },
];

export function Bereiche({
  haushaltId,
  planend,
  aktiv,
}: {
  /** Fehlt, solange kein Haushalt gewählt ist — dann führen die Links
   *  ohne Anhängsel, und die Seite sucht sich selbst einen. */
  haushaltId?: string;
  planend: boolean;
  aktiv: Bereich;
}) {
  const sichtbar = bereiche.filter((b) => planend || !b.nurPlanend);

  return (
    <nav
      aria-label="Bereiche"
      // Auf dem Telefon waagerecht scrollbar: Fünf Einträge brächen dort in zwei
      // Zeilen um, und zwei Zeilen Navigation sehen aus wie Inhalt. -mx-5/px-5
      // lässt den Streifen am Rand auslaufen, damit sichtbar ist, dass da noch
      // etwas kommt.
      //
      // Ab sm wird umgebrochen statt gescrollt. Die Unterseiten sind schmaler
      // als der Plan (max-w-lg gegen max-w-2xl); ohne Umbruch steht der letzte
      // Eintrag dort angeschnitten am Rand, während daneben das halbe Fenster
      // leer ist — das sieht kaputt aus, nicht scrollbar.
      className="-mx-5 flex gap-2 overflow-x-auto px-5 pb-1 [scrollbar-width:none] sm:flex-wrap sm:overflow-visible [&::-webkit-scrollbar]:hidden"
    >
      {sichtbar.map((b) => {
        const hier = b.id === aktiv;
        return (
          <Link
            key={b.id}
            href={haushaltId ? `${b.pfad}?haushalt=${encodeURIComponent(haushaltId)}` : b.pfad}
            aria-current={hier ? "page" : undefined}
            className={`inline-flex min-h-10 shrink-0 items-center rounded-full border px-4 text-sm font-semibold transition-colors ${
              hier
                ? "border-primary bg-primary-soft text-primary"
                : "border-line-strong text-muted hover:border-primary hover:text-primary"
            }`}
          >
            <span className="inline-flex items-center gap-2">
              {b.titel}
              <Laedt />
            </span>
          </Link>
        );
      })}
    </nav>
  );
}
