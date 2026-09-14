import type { MetadataRoute } from "next";

/**
 * Das Web-App-Manifest — damit die App auf den Startbildschirm passt.
 *
 * Ohne das ist sie ein Lesezeichen im Browser, und ein Haushaltsplaner, den
 * man über die Lesezeichenleiste erreicht, wird zweimal geöffnet. Mit dem
 * Manifest bekommt sie ein Symbol neben den anderen Apps, startet ohne
 * Adressleiste und sieht aus, als gehörte sie dorthin.
 *
 * `display: "standalone"` und nicht `fullscreen`: Die Statusleiste mit Uhrzeit
 * und Akku gehört dem Telefon, nicht uns.
 *
 * Die Farben hier sind die hellen. Anders als in `viewport.themeColor` kennt
 * das Manifest keine Medienabfrage — der dunkle Modus wird über die
 * Metadaten im Layout bedient, das Manifest liefert den Startbildschirm beim
 * allerersten Öffnen.
 */
export default function manifest(): MetadataRoute.Manifest {
  return {
    name: "Haushalt — Wochenplan",
    short_name: "Haushalt",
    description:
      "Wochenplanung für berufstätige Eltern — die App verteilt nicht nur Zeit, sondern auch Kopflast.",
    lang: "de",
    start_url: "/",
    display: "standalone",
    background_color: "#f6f8f4",
    theme_color: "#2c6a55",
    icons: [
      { src: "/icon-192.png", sizes: "192x192", type: "image/png" },
      { src: "/icon-512.png", sizes: "512x512", type: "image/png" },
      // Android schneidet bis zu 20 % weg, um das Symbol in seine Form zu
      // bringen. Diese Fassung ist deshalb randlos gefüllt und das Zeichen
      // sitzt kleiner in der Mitte.
      { src: "/icon-maskable-512.png", sizes: "512x512", type: "image/png", purpose: "maskable" },
    ],
  };
}
