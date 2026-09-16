"use client";

import { useSyncExternalStore } from "react";

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
    marke.setAttribute("content", dunkel ? "#10150f" : "#f1f4ec");
  }

  // Die eigene Änderung löst kein Ereignis des Browsers aus — `storage` feuert
  // nur in ANDEREN Tabs. Ohne diese Zeile bliebe der jeweils andere Schalter
  // stehen.
  for (const rueckruf of hoerer) rueckruf();
}

/* ---------------------------------------------------------------- Abonnement
 *
 * Die Farbwahl ist ein Zustand AUSSERHALB von React: Sie steht am
 * `<html>`-Element, im localStorage und in der Einstellung des Geräts. Für
 * genau diesen Fall gibt es `useSyncExternalStore` — und deshalb steht hier
 * kein `useEffect` mit `setState` darin.
 *
 * Der Unterschied ist nicht nur formal. Mit einem Effekt malt React erst mit
 * einem falschen Wert, misst, malt noch einmal — bei jeder Komponente, die
 * mitliest. Mit dem Abonnement fragt React den Wert, wenn es ihn braucht, und
 * erfährt von Änderungen über den Rückruf. Der Linter von React 19 nennt das
 * „cascading renders", und der Hinweis ist berechtigt.
 *
 * Drei Quellen, ein Rückruf: das Gerät (matchMedia), ein anderer Tab
 * (storage) und wir selbst (melde). Ohne die dritte ändert der Knopf oben die
 * Farbe, und die Schalter in den Einstellungen zeigen weiter den alten Stand.
 */

const hoerer = new Set<() => void>();

function abonniere(rueckruf: () => void): () => void {
  hoerer.add(rueckruf);
  const abfrage = window.matchMedia("(prefers-color-scheme: dark)");
  abfrage.addEventListener("change", rueckruf);
  window.addEventListener("storage", rueckruf);
  return () => {
    hoerer.delete(rueckruf);
    abfrage.removeEventListener("change", rueckruf);
    window.removeEventListener("storage", rueckruf);
  };
}

/**
 * Die gewählte Einstellung. `null`, solange auf dem Server gerendert wird —
 * dort gibt es kein Gerät, dessen Einstellung man lesen könnte, und eine
 * geratene Antwort wäre beim ersten Malen falsch.
 */
export function useModus(): Modus | null {
  return useSyncExternalStore(
    abonniere,
    () => lies(),
    () => null,
  );
}

/** Was gerade tatsächlich zu sehen ist. `null` auf dem Server, siehe oben. */
export function useDunkel(): boolean | null {
  return useSyncExternalStore(
    abonniere,
    () => istDunkel(lies()),
    () => null,
  );
}
