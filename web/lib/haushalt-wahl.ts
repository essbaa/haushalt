import { cookies } from "next/headers";
import type { Haushalt } from "@/lib/api";

/**
 * Welchen Haushalt die Seite zeigt, wenn keiner in der Adresse steht.
 *
 * Bis hierher war das schlicht der erste aus der Liste — und die ist nach
 * `created_at` sortiert, also der **älteste**. Das hat tagelang einen Fehler
 * vorgetäuscht, den es nie gab: Beantwortete Fragen kamen nach dem Neuladen
 * wieder. In Wahrheit wurden sie im echten Haushalt beantwortet und im
 * ältesten wieder gestellt — einem Übungshaushalt vom Vortag, in dem nichts
 * beantwortet war. Die Katalogreihenfolge ist überall dieselbe, also standen
 * dort dieselben zwei Fragen oben.
 *
 * Auf dem Telefon war es am schlimmsten: Das Symbol auf dem Startbildschirm
 * öffnet `/`, also bei jedem Start den ältesten Haushalt.
 *
 * Jetzt zählt, wo jemand zuletzt war. Ein Cookie, kein Eintrag in der
 * Datenbank: Es ist eine Gewohnheit dieses Geräts und keine Eigenschaft des
 * Haushalts — wer am Telefon den einen und am Rechner den anderen ansieht,
 * hat nicht zwei Wahrheiten, sondern zwei Geräte.
 *
 * Die Reihenfolge ist: Adresse schlägt Cookie schlägt Liste. Ein Link mit
 * `?haushalt=` meint genau diesen Haushalt und darf nie gegen eine
 * gespeicherte Gewohnheit verlieren.
 */
export const HAUSHALT_COOKIE = "haushalt";

export async function waehleHaushalt<T extends Haushalt>(
  haushalte: T[],
  ausDerAdresse?: string,
): Promise<T | undefined> {
  if (haushalte.length === 0) return undefined;

  const ausAdresse = haushalte.find((h) => h.id === ausDerAdresse);
  if (ausAdresse) return ausAdresse;

  // Der gemerkte Haushalt gilt nur, solange er noch einer ist: Wer aus einem
  // Haushalt ausgeschieden ist, soll nicht auf einer leeren Seite landen,
  // sondern beim nächsten, den er hat.
  const gemerkt = (await cookies()).get(HAUSHALT_COOKIE)?.value;
  const ausCookie = haushalte.find((h) => h.id === gemerkt);
  if (ausCookie) return ausCookie;

  return haushalte[0];
}
