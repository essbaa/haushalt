import Link from "next/link";
import { Farbknopf } from "@/app/components/farbknopf";
import { Sitzung } from "@/app/components/sitzung";

/**
 * Die Kopfzeile der angemeldeten Seiten.
 *
 * Sie gab es nur auf dem Wochenplan. Die Unterseiten — Aufgaben, Anlässe,
 * Einstellungen, Einladen — begannen mit der Bereichsleiste und einer
 * Überschrift, ohne Name, ohne Sitzung, ohne Weg zurück außer über die
 * Leiste.
 *
 * Das fiel nicht auf, solange nichts darin stand. Beim Farbknopf fiel es
 * sofort auf: **Ein Bedienelement, das auf einer von fünf Seiten existiert,
 * ist schlechter als eins, das an einer einzigen erklärten Stelle steht.**
 * Also zuerst die Kopfzeile überall, dann der Knopf darin.
 *
 * Klebend, aber schmal: Auf dem Telefon liegt die Navigation unten, hier oben
 * stehen nur Ort und Gerät. Zusammen sind das 56 Pixel, die nicht scrollen —
 * dafür ist der Weg zum Plan von jeder Seite aus derselbe.
 */
export function Kopfzeile({ breit = false }: { breit?: boolean }) {
  return (
    <header className="sticky top-0 z-10 border-b border-line bg-bg/90 backdrop-blur">
      <div
        className={`mx-auto flex w-full items-center justify-between gap-3 px-5 py-2.5 ${
          breit ? "max-w-2xl" : "max-w-lg"
        }`}
      >
        <Link href="/" className="text-base font-extrabold tracking-tight">
          Haushalt
        </Link>
        <div className="flex items-center gap-1">
          <Farbknopf />
          <Sitzung />
        </div>
      </div>
    </header>
  );
}
