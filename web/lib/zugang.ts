import "server-only";
import { timingSafeEqual } from "node:crypto";

/**
 * Wer sich überhaupt registrieren darf.
 *
 * Die App steht offen im Netz, und das ist bis auf Weiteres nicht gewollt: In
 * dieser Datenbank stehen Namen und Alter von Kindern, und das erste, was ein
 * offenes Registrierungsformular anzieht, sind keine Familien.
 *
 * Deshalb ein Zugangscode. Nicht als Sicherheitsmaßnahme im engeren Sinn — er
 * steht in einer Umgebungsvariablen und wird per Nachricht weitergegeben, wer
 * ihn hat, hat ihn. Er ist ein **Türschloss, kein Tresor**: Er hält
 * Vorbeikommende draußen und macht aus „jeder kann" ein „wer eingeladen ist".
 * Gegen jemanden, der es auf diese App abgesehen hat, hilft er nicht, und
 * dagegen hilft auch nichts anderes, was in einer halben Stunde zu bauen wäre.
 *
 * **Fehlt die Variable, darf niemand.** Das ist die unbequemere Vorgabe und
 * die richtige: Die andere Richtung — fehlt sie, darf jeder — heißt, dass ein
 * vergessener Eintrag beim nächsten Deploy die Tür stillschweigend aufmacht.
 * Ein Fehler, der sofort und laut auffällt, ist besser als einer, den man erst
 * bemerkt, wenn Fremde in der Datenbank stehen.
 */

/** Die eingetragenen Codes, normalisiert. */
function codes(): string[] {
  return (process.env.ZUGANGSCODES ?? "")
    .split(",")
    .map((c) => c.trim().toUpperCase())
    .filter((c) => c.length > 0);
}

/**
 * Vergleicht ohne Zeitunterschied.
 *
 * Bei einem Code, der in einer Umgebungsvariablen steht und mündlich
 * weitergegeben wird, ist das eher Handwerk als Notwendigkeit. Aber ein
 * Vergleich mit `===` an einer Anmeldestelle ist genau die Zeile, die drei
 * Jahre später jemand kopiert, wenn es dann um ein echtes Geheimnis geht.
 */
function gleich(a: string, b: string): boolean {
  const x = Buffer.from(a);
  const y = Buffer.from(b);
  if (x.length !== y.length) return false;
  return timingSafeEqual(x, y);
}

/** Ob mit diesem Code ein Konto angelegt werden darf. */
export function zugangErlaubt(code: string | undefined | null): boolean {
  const gegeben = (code ?? "").trim().toUpperCase();
  if (gegeben === "") return false;
  return codes().some((erlaubt) => gleich(erlaubt, gegeben));
}

/** Ob überhaupt ein Code eingetragen ist — für die Meldung an den Betreiber. */
export function zugangEingerichtet(): boolean {
  return codes().length > 0;
}

/**
 * Der Code, den ein Einladungslink mitbringen darf.
 *
 * Wer eine Einladung in den Haushalt bekommt, ist eingeladen — ihm daneben
 * noch eine zweite Zeichenkette zu schicken, die er in ein zweites Feld
 * tippen muss, ist eine Hürde ohne Gegenwert. Der Link trägt den ersten
 * eingetragenen Code mit, das Formular füllt ihn vor.
 *
 * Damit ist klar, was der Code leistet und was nicht: Er hängt an der
 * Einladung, nicht am Menschen. Sobald daraus mehr werden soll — Codes mit
 * Ablauf, Codes je Person, Codes, die sich verbrauchen —, gehört er in die
 * Datenbank und nicht mehr hierher.
 */
export function ersterCode(): string {
  return codes()[0] ?? "";
}
