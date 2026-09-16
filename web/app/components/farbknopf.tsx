"use client";

import { Mond, Sonne } from "@/app/components/icons";
import { useDunkel, waehle } from "@/lib/farbmodus";

/**
 * Der schnelle Griff: hell oder dunkel, ein Tipp.
 *
 * Bewusst **zwei** Zustände hier und drei in den Einstellungen. Ein
 * Bedienelement in der Kopfzeile, das im Kreis durch drei Zustände schaltet,
 * ist nicht zu durchschauen: Man tippt, etwas ändert sich, man tippt noch
 * einmal, etwas anderes ändert sich — und „wie das Gerät" sieht in dem Moment
 * genauso aus wie die ausdrückliche Wahl, in dem es dasselbe ergibt. Wer
 * einmal hier tippt, hat sich entschieden; wer die Entscheidung zurücknehmen
 * will, findet „Wie das Gerät" in den Einstellungen.
 *
 * Das Zeichen zeigt, **was passiert**, nicht was gerade ist: Im Dunkeln steht
 * die Sonne, weil sie das Ziel des Tippens ist. Beide Lesarten sind
 * verbreitet und beide falsch für jemanden — deshalb steht die Wahrheit in
 * `aria-label` und im Titel, und die kann man nicht missverstehen.
 */
export function Farbknopf({ className = "" }: { className?: string }) {
  // Kein Effekt, der nach dem ersten Malen den Zustand nachträgt: Die
  // Farbwahl ist ein Zustand außerhalb von React, und `useSyncExternalStore`
  // ist dafür gebaut (siehe lib/farbmodus.ts). `null` heißt „auf dem Server
  // gerendert" — dann steht hier ein Platz in derselben Größe, damit die
  // Kopfzeile nicht eine Zehntelsekunde später zur Seite rückt.
  const dunkel = useDunkel();

  if (dunkel === null) {
    return <span aria-hidden="true" className={`block size-10 ${className}`} />;
  }

  return (
    <button
      type="button"
      onClick={() => waehle(dunkel ? "hell" : "dunkel")}
      aria-label={dunkel ? "Helles Aussehen einschalten" : "Dunkles Aussehen einschalten"}
      title={dunkel ? "Hell" : "Dunkel"}
      className={`inline-flex size-10 items-center justify-center rounded-full text-muted transition-colors hover:bg-surface-2 hover:text-fg ${className}`}
    >
      {dunkel ? <Sonne className="size-5" /> : <Mond className="size-5" />}
    </button>
  );
}
