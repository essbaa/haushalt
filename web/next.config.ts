import path from "node:path";
import type { NextConfig } from "next";

const nextConfig: NextConfig = {
  turbopack: {
    // Die Projektwurzel ausdrücklich festlegen.
    //
    // Ohne diese Zeile sucht Turbopack sie selbst, indem es nach oben nach
    // einer package-lock.json schaut — und findet dabei auch eine, die
    // versehentlich im Home-Verzeichnis gelandet ist. Dann zeigt der Alias
    // "@/" auf das falsche Verzeichnis und jeder Import schlägt fehl, mit
    // einer Meldung, die nach einem Tippfehler aussieht.
    root: path.join(__dirname),
  },
};

export default nextConfig;
