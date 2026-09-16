import Link from "next/link";
import { Farbknopf } from "@/app/components/farbknopf";

/**
 * Der Kopf der Landingpage.
 *
 * Er fehlte ganz — die Seite begann mit einer Überschrift und hatte keinen
 * Ort, an dem „was ist das hier" und „ich bin schon dabei" stehen. Beides
 * gehört nach oben rechts, weil es dort gesucht wird.
 *
 * Nicht klebend: Die Seite ist kurz, und ein Balken, der beim Lesen
 * mitfährt, nimmt auf dem Telefon dauerhaft Platz für eine Entscheidung, die
 * man einmal trifft.
 */
export function LandingKopf() {
  return (
    <header className="mx-auto flex w-full max-w-5xl items-center justify-between gap-4 px-5 py-5">
      <Link href="/" className="flex items-center gap-2.5">
        {/* Kein Wortbild und kein Logo — beides gibt es noch nicht. Ein
            Quadrat mit dem Anfangsbuchstaben ist ehrlicher als ein
            Platzhalter, der so tut. */}
        <span
          aria-hidden="true"
          className="inline-flex size-8 items-center justify-center rounded-lg bg-primary text-base font-extrabold text-on-primary"
        >
          H
        </span>
        <span className="text-base font-extrabold tracking-tight">Haushalt</span>
      </Link>

      <nav aria-label="Hauptbereiche" className="flex items-center gap-1 sm:gap-2">
        <Farbknopf />
        <Link
          href="/demo"
          className="inline-flex min-h-10 items-center rounded-full px-3 text-sm font-semibold text-muted transition-colors hover:bg-surface-2 hover:text-fg sm:px-4"
        >
          Demo
        </Link>
        <Link
          href="/anmelden"
          className="inline-flex min-h-10 items-center rounded-full border border-line-strong px-3.5 text-sm font-semibold transition-colors hover:border-primary hover:text-primary sm:px-4"
        >
          Anmelden
        </Link>
      </nav>
    </header>
  );
}
