"use client";

/**
 * Hell, dunkel oder wie das Gerät — die Rechnung dahinter, an einer Stelle.
 *
 * Es gibt zwei Bedienelemente dafür: den Knopf in der Kopfzeile (schnelles
 * Umlegen) und die drei Schalter in den Einstellungen (mit „wie das Gerät").
 * Beide müssten sonst dasselbe wissen — wo der Wert steht, wie er am
 * `<html>` landet, welche Farbe die Systemleiste bekommt. Drei Abschriften
 * davon wären drei Gelegenheiten, sie auseinanderlaufen zu lassen.
 *
 * Das Skript in app/layout.tsx ist die vierte Stelle und bleibt eine eigene:
 * Es läuft, bevor React existiert, und darf nichts importieren. Dafür tut es
 * genau eine Sache — das Attribut setzen — und die steht dort in zwei Zeilen.
 */

export type Modus = "system" | "hell" | "dunkel";

const SCHLUESSEL = "farbmodus";

/** Was gespeichert ist. Ohne Wahl: „wie das Gerät". */
export function lies(): Modus {
  try {
    const wert = window.localStorage.getItem(SCHLUESSEL);
    return wert === "hell" || wert === "dunkel" ? wert : "system";
  } catch {
    // Privates Fenster oder gesperrter Speicher. Kein Grund zu scheitern.
    return "system";
  }
}

/** Was gerade tatsächlich zu sehen ist — die Wahl oder das, was das Gerät sagt. */
export function istDunkel(modus: Modus): boolean {
  if (modus === "dunkel") return true;
  if (modus === "hell") return false;
  return window.matchMedia("(prefers-color-scheme: dark)").matches;
}

/**
 * Setzt die Wahl: am `<html>`, im Speicher und an der Systemleiste.
 *
 * Die Systemleiste gehört dazu, auch wenn sie leicht vergessen wird: Ohne sie
 * steht auf dem iPhone ein heller Streifen über einer dunklen App, und das
 * sieht nach Fehler aus, nicht nach Absicht.
 */
export function waehle(modus: Modus): void {
  const wurzel = document.documentElement;
  if (modus === "system") wurzel.removeAttribute("data-theme");
  else wurzel.setAttribute("data-theme", modus === "hell" ? "light" : "dark");

  try {
    if (modus === "system") window.localStorage.removeItem(SCHLUESSEL);
    else window.localStorage.setItem(SCHLUESSEL, modus);
  } catch {
    // Dann gilt die Wahl für diese Sitzung. Ärgerlich, nicht schlimm.
  }

  const dunkel = istDunkel(modus);
  for (const marke of document.querySelectorAll('meta[name="theme-color"]')) {
    marke.setAttribute("content", dunkel ? "#10150f" : "#f6f8f4");
  }
}
