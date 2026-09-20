import type { Aufgabe } from "@/lib/api";

/**
 * Warum diese Aufgabe bei dieser Person liegt, als Satzteil.
 *
 * Steht hier und nicht in aufgabe-zeile.tsx, und das ist kein Ordnungssinn:
 * Jene Datei trägt `"use client"`, und damit wird **jeder** Export daraus zur
 * Client-Referenz — auch eine reine Funktion ohne Zustand. Die aufgeklappte
 * Bilanz wird auf dem Server gerendert und bekam beim Aufruf:
 *
 *   Attempted to call begruendung() from the server but begruendung is on
 *   the client.
 *
 * Weder `tsc` noch `eslint` sehen das; es ist eine Grenze zur Laufzeit und
 * keine des Typsystems. Eine Funktion, die beide Seiten brauchen, gehört
 * deshalb in ein Modul ohne Direktive — dann darf sie jeder.
 *
 * Einmal geschrieben, weil Plan und Bilanz dieselbe Auskunft geben müssen.
 * Zwei Übersetzungen desselben Codes wären zwei Gelegenheiten, dem Menschen
 * an zwei Stellen Verschiedenes über dieselbe Aufgabe zu erzählen.
 */
export function begruendung(a: Aufgabe, namen: Record<string, string>): string {
  const zuletzt = a.begruendung.zuletzt_bei;
  switch (a.begruendung.code) {
    case "rotation":
      return zuletzt ? `zuletzt bei ${namen[zuletzt] ?? zuletzt}` : "Rotation";
    case "ausgleich":
      return "zum Ausgleich";
    case "feste_person":
      return "feste Zuständigkeit";
    case "einzige_moeglichkeit":
      return "einzige Möglichkeit";
    case "frist":
      return "wegen der Frist";
    case "eigene_aufgabe":
      return "eigene Aufgabe";
    case "von_hand":
      return zuletzt ? `von Hand, vorher ${namen[zuletzt] ?? zuletzt}` : "von Hand verteilt";
    case "absprache":
      return "so abgesprochen";
    default:
      return "";
  }
}
