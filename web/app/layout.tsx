import type { Metadata } from "next";
import "./globals.css";

/**
 * Schriften: bewusst die Systemschriften des Geräts, kein next/font/google.
 *
 * Das spart einen Netzzugriff beim Bauen — die CI braucht damit keinen Zugang
 * zu fonts.googleapis.com — und auf macOS und iOS sieht die Systemschrift
 * ohnehin gut aus. Wenn du lieber Geist willst: next/font/google importieren
 * und die Variablen wieder an <html> hängen, dann greifen die Tailwind-Tokens
 * in globals.css automatisch.
 */
export const metadata: Metadata = {
  title: "Haushalt",
  description:
    "Wochenplanung für berufstätige Eltern — die App verteilt nicht nur Zeit, sondern auch Kopflast.",
};

export default function RootLayout({ children }: LayoutProps<"/">) {
  return (
    <html lang="de" className="h-full antialiased">
      <body className="min-h-full flex flex-col">{children}</body>
    </html>
  );
}
