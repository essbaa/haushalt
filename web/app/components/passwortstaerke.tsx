"use client";

import { staerke, type Klassen } from "@/lib/passwort";

/**
 * Die Leiste unter dem Passwortfeld.
 *
 * Vier Abschnitte statt eines Balkens mit Prozent: „73 % stark" wäre eine
 * Genauigkeit, die niemand hat. Vier Stufen sind eine Aussage, die man
 * verstehen kann.
 *
 * Die Zeichenklassen stehen als Merkmale daneben — erfüllt oder nicht, ohne
 * Häkchen und Kreuz in Rot. Sie sind ein Rat: Wer ein langes Passwort ohne
 * Sonderzeichen nimmt, bekommt trotzdem „sehr stark", und das ist richtig so
 * (siehe lib/passwort.ts).
 *
 * `aria-live="polite"`: Die Bewertung ändert sich beim Tippen. Eine
 * Bildschirmleseausgabe soll sie mitbekommen, aber nicht bei jedem Buchstaben
 * dazwischenreden.
 */
export function Passwortstaerke({
  passwort,
  umfeld = [],
}: {
  passwort: string;
  umfeld?: string[];
}) {
  if (passwort.length === 0) return null;

  const { stufe, text, klassen, einwand } = staerke(passwort, umfeld);

  const farbe = stufe <= 1 ? "bg-danger" : stufe === 2 ? "bg-clay" : "bg-primary";
  const schrift = stufe <= 1 ? "text-danger" : stufe === 2 ? "text-clay" : "text-primary";

  return (
    <div className="space-y-1.5">
      <div className="flex items-center gap-2">
        <div className="flex h-1.5 flex-1 gap-1" aria-hidden="true">
          {[1, 2, 3, 4].map((i) => (
            <div
              key={i}
              className={`flex-1 rounded-full transition-colors ${i <= stufe ? farbe : "bg-surface-2"}`}
            />
          ))}
        </div>
        <span aria-live="polite" className={`text-xs font-semibold ${schrift}`}>
          {text}
        </span>
      </div>

      {einwand ? (
        <p className={`text-xs ${schrift}`}>{einwand}</p>
      ) : (
        <ul className="flex flex-wrap gap-x-3 gap-y-1 text-xs text-subtle">
          <Merkmal erfuellt={passwort.length >= 12}>12 Zeichen</Merkmal>
          <Merkmal erfuellt={klassen.klein && klassen.gross}>Groß und klein</Merkmal>
          <Merkmal erfuellt={klassen.ziffer}>Ziffer</Merkmal>
          <Merkmal erfuellt={klassen.sonder}>Sonderzeichen</Merkmal>
        </ul>
      )}
    </div>
  );
}

function Merkmal({ erfuellt, children }: { erfuellt: boolean; children: string }) {
  return (
    <li className={erfuellt ? "font-semibold text-primary" : ""}>
      <span aria-hidden="true">{erfuellt ? "✓" : "·"}</span> {children}
      <span className="sr-only">{erfuellt ? " — erfüllt" : " — fehlt"}</span>
    </li>
  );
}

export type { Klassen };
