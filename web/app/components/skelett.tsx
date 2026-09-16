/**
 * Platzhalter, solange die Daten unterwegs sind.
 *
 * Der Weg dahin war vorher ein anderer: Beim Antippen eines Bereichs drehte
 * sich ein Kreisel im angetippten Knopf, und die alte Seite blieb stehen, bis
 * der Server fertig war. Das war ehrlich — aber es ist die falsche Ehrlichkeit.
 * Der Mensch wartet auf eine Seite und sieht dabei die vorige.
 *
 * Mit `loading.tsx` erzeugt Next.js an dieser Stelle eine Suspense-Grenze:
 * Das Gerüst der neuen Seite steht **sofort**, der Inhalt kommt nach. Das
 * ändert an der Wartezeit nichts und alles am Gefühl — man ist schon dort und
 * sieht, was gleich kommt, statt festzuhängen.
 *
 * Zwei Regeln, damit daraus kein Selbstzweck wird:
 *
 *  1. **Das Skelett muss aussehen wie das, was kommt.** Gleiche Höhen, gleiche
 *     Abstände, gleiche Anzahl. Sonst springt das Bild beim Eintreffen der
 *     Daten, und der Platzhalter hat den Ruck erzeugt, den er verhindern
 *     sollte (CLS).
 *  2. **Es zeigt nie einen Inhalt, den es nicht kennt.** Keine erfundenen
 *     Namen, keine Zahlen — graue Flächen. Ein Platzhalter, der so tut, als
 *     wüsste er schon etwas, ist eine Behauptung wie jede andere.
 *
 * `animate-pulse` hört von selbst auf, wenn jemand Bewegung abbestellt hat:
 * Die Regel in globals.css setzt jede Animationsdauer auf null, und übrig
 * bleibt eine ruhige graue Fläche. Die Auskunft ist der Zweck, die Bewegung
 * nur ihre Form (ADR-0009).
 */

/** Ein grauer Balken. `breit` ist eine Tailwind-Breitenklasse. */
export function Balken({ breit = "w-full", hoch = "h-4" }: { breit?: string; hoch?: string }) {
  return <span className={`block animate-pulse rounded-md bg-surface-3 ${breit} ${hoch}`} />;
}

/** Die Kopfzeile einer Unterseite: Überschrift und Einleitung. */
export function KopfSkelett() {
  return (
    <div className="mt-5 mb-8 space-y-3">
      <Balken breit="w-48" hoch="h-8" />
      <Balken breit="w-full max-w-prose" hoch="h-4" />
      <Balken breit="w-2/3 max-w-prose" hoch="h-4" />
    </div>
  );
}

/**
 * Eine Aufgabenzeile, so hoch wie die echte.
 *
 * Die Maße stammen aus aufgabe-zeile.tsx: Zeichen 32 Pixel, Titel und
 * Kennzahlen darunter, rechts die beiden Knöpfe. Wer sie dort ändert, sollte
 * hier nachziehen — dieselbe Pflicht wie bei jedem Platzhalter, und der Grund,
 * warum es nicht mehr davon gibt als nötig.
 */
export function ZeileSkelett() {
  return (
    <div className="flex gap-3 rounded-lg px-3 py-3">
      <span className="size-8 shrink-0 animate-pulse rounded-full bg-surface-3" />
      <div className="min-w-0 flex-1 space-y-2 py-0.5">
        <Balken breit="w-2/3" hoch="h-4" />
        <Balken breit="w-1/3" hoch="h-3" />
      </div>
      <span className="size-11 shrink-0" />
    </div>
  );
}

/** Ein Tag im Wochenplan: schmale Spalte mit Marke, daneben zwei Zeilen. */
export function TagSkelett({ erster = false, zeilen = 2 }: { erster?: boolean; zeilen?: number }) {
  return (
    <div className={`flex gap-4 sm:gap-6 ${erster ? "" : "pt-4"}`}>
      <div className="flex w-12 shrink-0 flex-col items-center sm:w-16">
        <span className="size-11 animate-pulse rounded-full bg-surface-3" />
        <span className="mt-1 w-px flex-1 bg-line" />
      </div>
      <div
        className={`min-w-0 flex-1 space-y-1.5 pb-8 ${erster ? "" : "border-t border-line pt-3.5"}`}
      >
        {Array.from({ length: zeilen }, (_, i) => (
          <ZeileSkelett key={i} />
        ))}
      </div>
    </div>
  );
}

/** Ein Kasten in einer Liste — Vorlagen, Anlässe, Einstellungen. */
export function KastenSkelett({ hoch = "h-14" }: { hoch?: string }) {
  return <div className={`animate-pulse rounded-lg border border-line bg-surface ${hoch}`} />;
}
