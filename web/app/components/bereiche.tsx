import Link from "next/link";
import { Fahne, Kalender, Liste, Zahnrad } from "@/app/components/icons";
import { Laedt } from "@/app/components/laedt";

/**
 * Die Bereiche der App.
 *
 * Auf dem Telefon eine feste Leiste am unteren Rand, mit Zeichen und kleiner
 * Beschriftung — dort, wo der Daumen ohnehin liegt, und dort, wo jede App
 * ihre Navigation hat. Ab sm wandert sie zurück in den Textfluss und sieht
 * aus wie vorher: eine Reihe Kapseln über dem Inhalt.
 *
 * Das ist eine Kehrtwende, und sie gehört benannt: Hier stand, ein fester
 * Balken koste auf einem Telefon dauerhaft 44 Pixel, und der Grund zum
 * Wechseln entstehe beim Ankommen und nicht mitten im Scrollen. Beides stimmt
 * weiterhin. Was dagegen schwerer wiegt: Der Wochenplan ist lang, und oben
 * angeheftete Kapseln scrollen weg. Wer unten in der Woche steht und auf
 * „Aufgaben" will, scrollt erst zurück. Die 44 Pixel kosten einmal Platz, das
 * Zurückscrollen kostet jedes Mal.
 *
 * „Einladen" ist dabei herausgefallen und steht jetzt in den Einstellungen.
 * Es ist eine Handlung und kein Ort — und eine Leiste mit fünf Einträgen auf
 * einem schmalen Telefon ist eine Leiste, in der man danebentippt.
 *
 * Server Component: nur Links, kein Zustand.
 */
export type Bereich = "plan" | "vorlagen" | "termine" | "einstellungen" | "einladen";

const bereiche: {
  id: Bereich;
  titel: string;
  pfad: string;
  nurPlanend?: boolean;
  Zeichen: (p: { className?: string }) => React.ReactElement;
}[] = [
  { id: "plan", titel: "Plan", pfad: "/", Zeichen: Kalender },
  { id: "vorlagen", titel: "Aufgaben", pfad: "/vorlagen", nurPlanend: true, Zeichen: Liste },
  { id: "termine", titel: "Anlässe", pfad: "/termine", nurPlanend: true, Zeichen: Fahne },
  { id: "einstellungen", titel: "Einstellungen", pfad: "/einstellungen", Zeichen: Zahnrad },
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
  // Einladen hat keinen eigenen Platz mehr in der Leiste. Auf dieser Seite
  // leuchtet der Bereich, aus dem man kommt — eine Leiste ohne Markierung
  // sieht aus, als wäre man nirgends.
  const hier = aktiv === "einladen" ? "einstellungen" : aktiv;

  return (
    <nav
      aria-label="Bereiche"
      className={[
        // Telefon: fest unten, über allem, mit Platz für die Hausgeste.
        "fixed inset-x-0 bottom-0 z-20 flex border-t border-line bg-bg/95 backdrop-blur",
        "pb-[env(safe-area-inset-bottom,0px)]",
        // Ab sm: zurück in den Fluss, als Reihe Kapseln.
        "sm:static sm:z-auto sm:flex-wrap sm:gap-2 sm:border-0 sm:bg-transparent sm:pb-1 sm:backdrop-blur-none",
      ].join(" ")}
    >
      {sichtbar.map((b) => {
        const dort = b.id === hier;
        const Zeichen = b.Zeichen;
        return (
          <Link
            key={b.id}
            href={haushaltId ? `${b.pfad}?haushalt=${encodeURIComponent(haushaltId)}` : b.pfad}
            aria-current={dort ? "page" : undefined}
            className={[
              "flex min-h-14 flex-1 flex-col items-center justify-center gap-0.5 text-center transition-colors",
              "sm:min-h-10 sm:flex-none sm:flex-row sm:gap-2 sm:rounded-full sm:border sm:px-4",
              dort
                ? "text-primary sm:border-primary sm:bg-primary-soft"
                : "text-muted hover:text-fg sm:border-line-strong sm:hover:border-primary sm:hover:text-primary",
            ].join(" ")}
          >
            {/* Der Kreisel ersetzt das Zeichen, solange die Seite lädt. Die
                Rückmeldung erscheint damit dort, wo der Finger war — und die
                Leiste springt nicht, weil beide gleich groß sind. */}
            <span className="flex size-6 items-center justify-center sm:size-4">
              <Laedt>
                <Zeichen className="size-6 sm:size-4" />
              </Laedt>
            </span>
            <span className="text-[0.65rem] leading-none font-semibold sm:text-sm">{b.titel}</span>
          </Link>
        );
      })}
    </nav>
  );
}
