"use client";

import { type Modus, useModus, waehle as setze } from "@/lib/farbmodus";

const wahlen: { wert: Modus; titel: string }[] = [
  { wert: "system", titel: "Wie das Gerät" },
  { wert: "hell", titel: "Hell" },
  { wert: "dunkel", titel: "Dunkel" },
];

/**
 * Hell oder dunkel — oder gar keine Wahl.
 *
 * Beide Fassungen gab es von Anfang an: `:root` trägt die helle, die dunkle
 * kommt über `prefers-color-scheme`. Was fehlte, war die **Wahl**. Wer sein
 * Telefon dunkel stellt, weil abends das Lesen angenehmer ist, will deshalb
 * nicht jede App dunkel; und die Einstellung des Systems ist eine Aussage über
 * das System, keine über diese App.
 *
 * Drei Zustände und nicht zwei. Ein Umschalter kennt nur an und aus und
 * verliert damit genau den Zustand, in dem die meisten Menschen sind: „richte
 * dich nach meinem Gerät". Wer den nicht anbietet, zwingt zu einer
 * Entscheidung, die das Gerät schon getroffen hat — und die App bleibt hell,
 * wenn das Telefon abends umschaltet.
 *
 * Gespeichert wird im Browser, nicht im Haushalt: Es ist eine Gewohnheit
 * dieses Geräts. Dasselbe Argument wie beim zuletzt gewählten Haushalt.
 */
export function Farbmodus() {
  // Kein Effekt, der den gespeicherten Wert nachträgt: Die Farbwahl ist ein
  // Zustand außerhalb von React (lib/farbmodus.ts). `null` heißt „auf dem
  // Server gerendert" — dann ist noch keine der drei Kapseln markiert, weil
  // der Server nicht wissen kann, welche es ist. Eine geratene Markierung
  // wäre beim ersten Malen falsch, und das sieht man.
  const modus = useModus();

  return (
    <fieldset className="space-y-2">
      <legend className="text-sm font-semibold">Aussehen</legend>
      <div className="flex flex-wrap gap-2">
        {wahlen.map((w) => (
          <button
            key={w.wert}
            type="button"
            onClick={() => setze(w.wert)}
            aria-pressed={modus === w.wert}
            className={`inline-flex min-h-10 items-center rounded-full border px-3.5 text-sm font-semibold transition-colors ${
              modus === w.wert
                ? "border-primary bg-primary-soft text-primary"
                : "border-line-strong text-muted hover:text-fg"
            }`}
          >
            {w.titel}
          </button>
        ))}
      </div>
      <p className="text-xs leading-relaxed text-muted text-pretty">
        Gilt auf diesem Gerät. Ohne Wahl folgt die App der Einstellung deines Systems — oben in der
        Leiste legst du schnell um, hier steht der Weg zurück.
      </p>
    </fieldset>
  );
}
