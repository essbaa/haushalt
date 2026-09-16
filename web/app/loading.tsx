import { Bereiche } from "@/app/components/bereiche";
import { Balken, TagSkelett } from "@/app/components/skelett";

/**
 * Das Gerüst des Wochenplans, während er geholt wird.
 *
 * Next.js zeigt diese Datei, sobald jemand hierher navigiert, und tauscht sie
 * gegen die Seite, sobald der Server geantwortet hat. Vorher blieb die alte
 * Seite stehen und im angetippten Knopf drehte sich ein Kreisel — man wartete
 * dort, wo man herkam, statt dort, wo man hinwollte.
 *
 * Drei Tage und zwei Zeilen: ungefähr die Höhe einer echten Woche. Genau zu
 * treffen ist unmöglich, und das ist in Ordnung — was zählt, ist, dass nichts
 * springt, wenn die Daten eintreffen.
 */
export default function Laden() {
  return (
    <>
      <header className="sticky top-0 z-10 border-b border-line bg-bg/90 backdrop-blur">
        <div className="mx-auto flex w-full max-w-2xl items-center justify-between gap-3 px-5 py-3">
          <span className="text-base font-extrabold tracking-tight">Haushalt</span>
          <Balken breit="w-20" hoch="h-8" />
        </div>
      </header>

      <main className="mx-auto w-full max-w-2xl flex-1 px-5 py-8">
        <div className="mb-5 space-y-3">
          <Balken breit="w-64" hoch="h-9" />
          <Balken breit="w-44" hoch="h-6" />
        </div>

        <div className="mb-8 max-sm:mb-0">
          <Bereiche planend aktiv="plan" skelett />
        </div>

        <div className="space-y-6">
          <Balken breit="w-56" hoch="h-6" />
          <TagSkelett erster zeilen={2} />
          <TagSkelett zeilen={3} />
          <TagSkelett zeilen={2} />
        </div>
      </main>
    </>
  );
}
