/**
 * Die zehn Bereiche als Wörter für Menschen.
 *
 * Steht in lib und nicht in einer Komponente: Die Vorlagenliste ist eine
 * Client Component, und aus einer Datei mit `"use client"` wird jeder Export
 * zur Client-Referenz — auch eine Tabelle. Die aufgeklappte Bilanz wird auf
 * dem Server gerendert und käme sonst nicht heran (siehe lib/begruendung.ts,
 * derselbe Fehler, einmal reicht).
 *
 * Alle zehn aus planner.AllCategories. Eine Abbildung, die nur die bekannten
 * Fälle abdeckt, verrät sich beim ersten unbekannten — hier stand schon
 * einmal „termine" kleingeschrieben als Kennung.
 */
export const bereiche: Record<string, string> = {
  kueche: "Küche",
  waesche: "Wäsche",
  reinigung: "Reinigung",
  kind: "Kinder",
  vorrat: "Vorräte",
  termine: "Termine",
  verwaltung: "Verwaltung",
  wartung: "Wartung",
  sozial: "Soziales",
  aussen: "Draußen",
};
