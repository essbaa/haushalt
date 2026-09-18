import { Anmeldeformular } from "@/app/anmelden/formular";
import { mailErreichtAlle } from "@/lib/mail";

/**
 * Die Anmeldeseite.
 *
 * Eine Server Component über einer Client Component, und sie beantwortet drei
 * Fragen, die das erste Bild bestimmen — bevor der Browser irgendetwas malt.
 *
 *  1. **Kann diese App Mail zustellen?** Nur der Server weiß das; der
 *     Bildschirm nach der Registrierung sagt je nachdem etwas anderes.
 *  2. **Kommt hier jemand mit einer Einladung an?** Dann ist „Konto anlegen"
 *     das Richtige und nicht „Anmelden".
 *  3. **Wohin danach?**
 */
export default async function Seite({
  searchParams,
}: {
  searchParams: Promise<{ zugang?: string; weiter?: string }>;
}) {
  const { zugang, weiter } = await searchParams;
  const ziel = geprueftesZiel(weiter);

  return (
    <Anmeldeformular
      mailAn={mailErreichtAlle()}
      zugang={zugang ?? ""}
      // Ein Zugangscode im Link oder ein Ziel, das jemand ohne Konto gar
      // nicht erreichen kann: Beides heißt „hier kommt eine eingeladene
      // Person an".
      zuerstAnlegen={Boolean(zugang) || ziel.startsWith("/beitreten")}
      ziel={ziel}
    />
  );
}

/**
 * Wohin nach dem Anmelden.
 *
 * Aus der Adresse gelesen, damit ein Einladungslink nicht verloren geht:
 * /anmelden?weiter=/beitreten?code=… führt hinterher dorthin zurück.
 *
 * Nur Pfade, und keine, die mit zwei Schrägstrichen beginnen. Sonst wäre
 * ?weiter=//fremde.example eine offene Weiterleitung — der klassische Weg,
 * eine vertrauenswürdige Anmeldeseite als Sprungbrett zu missbrauchen.
 *
 * Steht seit der Teilung hier und nicht im Formular: Die Prüfung gehört
 * dorthin, wo der Wert das erste Mal angefasst wird, und das ist jetzt der
 * Server.
 */
function geprueftesZiel(weiter: string | undefined): string {
  if (!weiter || !weiter.startsWith("/") || weiter.startsWith("//")) return "/";
  return weiter;
}
