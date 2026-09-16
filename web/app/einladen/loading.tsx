import { Balken } from "@/app/components/skelett";

/**
 * Platzhalter für die Formularseiten.
 *
 * Er sagt fachlich nichts und ist genau deshalb hier: `app/loading.tsx` gilt
 * auch für jede Unterseite ohne eigenes — ohne diese Datei blitzte beim
 * Anmelden das Gerüst des Wochenplans auf, und der Platzhalter behauptete ein
 * Ziel, das gar nicht angesteuert wurde.
 */
export default function Laden() {
  return (
    <main className="mx-auto flex w-full max-w-sm flex-1 flex-col justify-center px-5 py-16">
      <div className="space-y-4 rounded-lg border border-line bg-surface p-6">
        <Balken breit="w-40" hoch="h-7" />
        <Balken breit="w-full" hoch="h-4" />
        <Balken breit="w-full" hoch="h-11" />
        <Balken breit="w-full" hoch="h-11" />
        <Balken breit="w-full" hoch="h-12" />
      </div>
    </main>
  );
}
