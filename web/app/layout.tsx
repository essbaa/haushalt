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

export default function RootLayout({ children }: LayoutProps<"/">) {
  return (
    <html lang="de" className={`${jakarta.variable} h-full antialiased`}>
      <body className="flex min-h-full flex-col bg-bg text-fg">{children}</body>
    </html>
  );
}
