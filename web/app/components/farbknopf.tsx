"use client";

import { useEffect, useState } from "react";
import { Mond, Sonne } from "@/app/components/icons";
import { istDunkel, lies, waehle } from "@/lib/farbmodus";

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
  // Vor dem Einhängen weiß der Server nichts über dieses Gerät. Bis dahin ein
  // Platz in derselben Größe — sonst rückt die Kopfzeile eine Zehntelsekunde
  // nach dem Laden zur Seite.
  const [dunkel, setDunkel] = useState<boolean | null>(null);

  useEffect(() => {
    const modus = lies();
    setDunkel(istDunkel(modus));

    // Wer „wie das Gerät" eingestellt hat, soll das Zeichen mitwandern sehen,
    // wenn das Telefon abends umschaltet — ohne die Seite neu zu laden.
    const abfrage = window.matchMedia("(prefers-color-scheme: dark)");
    const beiWechsel = () => {
      if (lies() === "system") setDunkel(abfrage.matches);
    };
    abfrage.addEventListener("change", beiWechsel);
    return () => abfrage.removeEventListener("change", beiWechsel);
  }, []);

  if (dunkel === null) {
    return <span aria-hidden="true" className={`block size-10 ${className}`} />;
  }

  return (
    <button
      type="button"
      onClick={() => {
        waehle(dunkel ? "hell" : "dunkel");
        setDunkel(!dunkel);
      }}
      aria-label={dunkel ? "Helles Aussehen einschalten" : "Dunkles Aussehen einschalten"}
      title={dunkel ? "Hell" : "Dunkel"}
      className={`inline-flex size-10 items-center justify-center rounded-full text-muted transition-colors hover:bg-surface-2 hover:text-fg ${className}`}
    >
      {dunkel ? <Sonne className="size-5" /> : <Mond className="size-5" />}
    </button>
  );
}
