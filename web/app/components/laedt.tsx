"use client";

import { useLinkStatus } from "next/link";
import type { ReactNode } from "react";

/**
 * Zeigt, dass ein Link gerade lädt.
 *
 * Die Seiten sind Server Components: Beim Antippen von „Aufgaben" holt der
 * Next.js-Server den neuen Inhalt, und auf dem Telefon dauert das ein bis zwei
 * Sekunden. Bis hierher passierte in dieser Zeit *nichts* — kein Zeichen, kein
 * Zustand, die alte Seite stand einfach weiter da. Wer nicht weiß, dass etwas
 * läuft, tippt ein zweites Mal.
 *
 * `useLinkStatus` kennt den Zustand genau des Links, in dem diese Komponente
 * steckt. Das ist besser als ein Balken am Seitenrand: Die Rückmeldung
 * erscheint dort, wo der Finger war.
 *
 * WO SIE NOCH STEHT, UND WO NICHT MEHR.
 *
 * Seit es `loading.tsx` gibt, zeigt ein Wechsel des Bereichs sofort das
 * Gerüst der neuen Seite. Dort ist der Kreisel überflüssig geworden und war
 * sogar schädlich: zwei Signale für ein Ereignis, und das schwächere kam
 * zuerst — ein Flackern von hundert Millisekunden.
 *
 * Er bleibt an den Links, die **denselben Abschnitt** ansteuern und nur den
 * Anhang der Adresse ändern: die Wochenpfeile, „Diese Woche", die
 * Haushaltswahl. Next.js zeigt dort keine Suspense-Grenze — die Seite ist ja
 * schon da —, und ohne diesen Kreisel gäbe es wieder das, was auf dem Telefon
 * als Erstes aufgefallen ist: Man tippt, und eine Sekunde lang passiert
 * sichtbar nichts.
 *
 * Mit Kindern ersetzt der Kreisel sie (für Knöpfe, die nur ein Zeichen
 * tragen); ohne Kinder steht er daneben.
 */
export function Laedt({ children }: { children?: ReactNode }) {
  const { pending } = useLinkStatus();

  if (!pending) return children ?? null;

  return (
    <span
      role="status"
      aria-label="lädt"
      // motion-reduce: Wer Bewegung abbestellt hat, bekommt ein Pulsieren
      // statt der Drehung — aber nicht gar nichts. Die Auskunft ist der
      // Zweck, die Animation nur ihre Form (ADR-0009).
      className="inline-block size-4 animate-spin rounded-full border-2 border-current border-t-transparent opacity-70 motion-reduce:animate-pulse motion-reduce:border-t-current"
    />
  );
}
