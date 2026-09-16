import type { Metadata, Viewport } from "next";
import { Plus_Jakarta_Sans } from "next/font/google";
import "./globals.css";

/**
 * Plus Jakarta Sans: freundlich, gut lesbar in kleinen Graden, kein
 * Büro-Ernst. Die App wird zwischen Kita und Feierabend am Handy benutzt —
 * Schrift, die auf Präsentationsfolien gut aussieht, hilft dort nicht.
 *
 * next/font lädt sie beim Bauen herunter und liefert sie von der eigenen
 * Adresse aus. Kein Aufruf zu Google im Browser des Nutzers, keine dritte
 * Partei, die mitbekommt, wer die Seite öffnet.
 */
const jakarta = Plus_Jakarta_Sans({
  subsets: ["latin"],
  variable: "--font-jakarta",
  display: "swap",
});

export const metadata: Metadata = {
  title: "Haushalt",
  // Das Symbol für iOS liegt in public/. Android und der Browser holen sich
  // ihres aus dem Manifest (app/manifest.ts) beziehungsweise aus app/icon.svg.
  appleWebApp: { title: "Haushalt", capable: true, statusBarStyle: "default" },
  icons: { apple: "/apple-icon.png" },
  description:
    "Wochenplanung für berufstätige Eltern — die App verteilt nicht nur Zeit, sondern auch Kopflast.",
};

export const viewport: Viewport = {
  width: "device-width",
  initialScale: 1,
  // Kein maximumScale: Zoom zu verbieten ist auf dem Handy eine
  // Barriere für jeden, der die Schrift größer braucht.
  themeColor: [
    { media: "(prefers-color-scheme: light)", color: "#f6f8f4" },
    { media: "(prefers-color-scheme: dark)", color: "#10150f" },
  ],
};

/**
 * Die gespeicherte Farbwahl, bevor irgendetwas gemalt wird.
 *
 * Ohne dieses Skript malt der Browser erst die Vorgabe und tauscht sie eine
 * Zehntelsekunde später aus — auf einem dunkel gestellten Telefon ein weißer
 * Blitz beim Öffnen der App. Das ist der Grund, warum es ein blockierendes
 * Skript im Kopf sein muss und kein Effekt in React: Ein Effekt läuft nach
 * dem ersten Malen, und genau das erste Malen ist hier das Problem.
 *
 * Es ist mit Absicht winzig und fällt still aus, wenn der Speicher gesperrt
 * ist (privates Fenster). Dann gilt, was das Gerät sagt — die Vorgabe.
 */
const farbwahl = `try{var m=localStorage.getItem("farbmodus");if(m==="hell")document.documentElement.setAttribute("data-theme","light");else if(m==="dunkel")document.documentElement.setAttribute("data-theme","dark");}catch(e){}`;

export default function RootLayout({ children }: LayoutProps<"/">) {
  return (
    <html lang="de" className={`${jakarta.variable} h-full antialiased`} suppressHydrationWarning>
      <head>
        <script dangerouslySetInnerHTML={{ __html: farbwahl }} />
      </head>
      <body className="flex min-h-full flex-col bg-bg text-fg">{children}</body>
    </html>
  );
}
