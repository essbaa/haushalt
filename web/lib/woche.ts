/**
 * Kalenderwochen und Kalendertage.
 *
 * Der Dienst liefert Tage als reine Zeichenkette „2026-09-19" — ein
 * Kalendertag ohne Uhrzeit und ohne Zone (ADR-0002). Genau hier liegt die
 * Falle: `new Date("2026-09-19")` liest die Zeichenkette als Mitternacht UTC,
 * und westlich von Greenwich zeigt der Browser dann den Vortag an.
 *
 * Deshalb wird hier nirgends mit der lokalen Zeitzone gerechnet. Tage werden
 * als UTC gelesen und als UTC formatiert; nur die Frage „welche Woche ist
 * gerade?" schaut auf Europe/Berlin, weil sie von der Uhr abhängt.
 */

/** Der heutige Kalendertag in Berlin, als UTC-Mitternacht. */
function heuteInBerlin(jetzt: Date): Date {
  const teile = new Intl.DateTimeFormat("en-CA", {
    timeZone: "Europe/Berlin",
    year: "numeric",
    month: "2-digit",
    day: "2-digit",
  }).format(jetzt);
  const [jahr, monat, tag] = teile.split("-").map(Number);
  return new Date(Date.UTC(jahr, monat - 1, tag));
}

/**
 * Die laufende ISO-Kalenderwoche, z. B. „2026-W38".
 *
 * Nach ISO 8601 beginnt die Woche am Montag, und das Jahr der Woche ist das
 * Jahr ihres Donnerstags — deshalb kann der 1. Januar noch zur letzten Woche
 * des Vorjahres gehören.
 */
export function aktuelleWoche(jetzt: Date = new Date()): string {
  const d = heuteInBerlin(jetzt);
  const wochentag = (d.getUTCDay() + 6) % 7; // Montag = 0
  d.setUTCDate(d.getUTCDate() - wochentag + 3); // auf den Donnerstag
  const jahr = d.getUTCFullYear();
  const januar4 = Date.UTC(jahr, 0, 4);
  const versatz = (new Date(januar4).getUTCDay() + 6) % 7;
  const woche =
    1 + Math.round(((d.getTime() - januar4) / 86_400_000 - 3 + versatz) / 7);
  return `${jahr}-W${String(woche).padStart(2, "0")}`;
}

/** Prüft das Format, bevor eine Woche in die URL des Dienstes wandert. */
export function istWoche(wert: string): boolean {
  return /^\d{4}-W\d{2}$/.test(wert);
}

const wochentage = ["Montag", "Dienstag", "Mittwoch", "Donnerstag", "Freitag", "Samstag", "Sonntag"];

/** „Samstag, 19.09." aus „2026-09-19". */
export function tagLesbar(iso: string): string {
  const [jahr, monat, tag] = iso.split("-").map(Number);
  const d = new Date(Date.UTC(jahr, monat - 1, tag));
  const name = wochentage[(d.getUTCDay() + 6) % 7];
  return `${name}, ${String(tag).padStart(2, "0")}.${String(monat).padStart(2, "0")}.`;
}

/** Montag und Sonntag einer Woche als „14.09. bis 20.09.". */
export function wochenSpanne(woche: string): string {
  const [jahr, nummer] = woche.split("-W").map(Number);
  const januar4 = new Date(Date.UTC(jahr, 0, 4));
  const versatz = (januar4.getUTCDay() + 6) % 7;
  const montag = new Date(januar4);
  montag.setUTCDate(januar4.getUTCDate() - versatz + (nummer - 1) * 7);
  const sonntag = new Date(montag);
  sonntag.setUTCDate(montag.getUTCDate() + 6);
  const kurz = (d: Date) =>
    `${String(d.getUTCDate()).padStart(2, "0")}.${String(d.getUTCMonth() + 1).padStart(2, "0")}.`;
  return `${kurz(montag)} bis ${kurz(sonntag)}`;
}
