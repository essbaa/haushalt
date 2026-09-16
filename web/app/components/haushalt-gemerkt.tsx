"use client";

import { useEffect } from "react";

/**
 * Merkt sich, welcher Haushalt gerade angesehen wird.
 *
 * Schreibt ein Cookie, das die Server Components beim nächsten Aufruf lesen
 * (lib/haushalt-wahl.ts). Damit landet der nächste Start dort, wo der letzte
 * aufgehört hat — und nicht beim ältesten Haushalt, nur weil der zuerst
 * angelegt wurde.
 *
 * Aus dem Browser und nicht vom Server: Ein Cookie darf in Next.js nur in
 * einer Server Action oder einem Route Handler gesetzt werden, nicht beim
 * Rendern. Eine Server Action für eine Gewohnheit dieses Geräts wäre ein
 * Rundgang zum Server für etwas, das der Server gar nicht wissen muss.
 *
 * Kein `Secure`: Die App läuft in der Entwicklung über http, und ein Cookie,
 * das lokal nie ankommt, ist schlimmer als eins ohne Flagge — es enthält eine
 * Kennung, die ohnehin in jeder Adresszeile steht. `SameSite=Lax` genügt
 * hier: Es gibt nichts zu stehlen und nichts auszulösen.
 */
export function HaushaltGemerkt({ id }: { id: string }) {
  useEffect(() => {
    if (!id) return;
    const jahr = 60 * 60 * 24 * 365;
    document.cookie = `haushalt=${encodeURIComponent(id)}; path=/; max-age=${jahr}; samesite=lax`;
  }, [id]);

  return null;
}
