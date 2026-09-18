import "server-only";

/**
 * Mailversand.
 *
 * Eine Funktion, zwei Wege — und der zweite ist der Grund, warum es diese
 * Datei gibt: **Ohne `RESEND_API_KEY` landet die Mail in der Serverkonsole,
 * statt dass etwas scheitert.** Damit lässt sich die Registrierung samt
 * Bestätigung und Zurücksetzen vollständig durchspielen, bevor eine Domain
 * eingerichtet ist. Wenn der Schlüssel da ist, ändert sich keine Zeile
 * anderswo.
 *
 * Kein SDK, nur `fetch`: Resend ist eine HTTP-Schnittstelle mit einem
 * Endpunkt. Ein Paket dafür wäre eine Abhängigkeit, die nichts abnimmt außer
 * dem Tippen von zwölf Zeilen.
 *
 * Warum überhaupt ein eigener Versand und nicht Gmail: An beliebige Adressen
 * zustellen darf nur, wer eine Domain verifiziert hat. Ohne das käme die
 * Bestätigungsmail beim Absender an und sonst nirgends — die Person, die sie
 * braucht, bekäme sie nie.
 */
type Nachricht = {
  an: string;
  betreff: string;
  text: string;
};

export async function schicke({ an, betreff, text }: Nachricht): Promise<void> {
  const schluessel = process.env.RESEND_API_KEY;
  const von = process.env.MAIL_VON;

  if (!schluessel || !von) {
    // Absichtlich auffällig: Wer das im Log sieht, soll wissen, dass die Mail
    // NICHT unterwegs ist — und trotzdem weiterarbeiten können.
    console.info(
      [
        "",
        "┌─ Mail (kein RESEND_API_KEY — nicht verschickt)",
        `│  an:      ${an}`,
        `│  betreff: ${betreff}`,
        "│",
        ...text.split("\n").map((z) => `│  ${z}`),
        "└─",
        "",
      ].join("\n"),
    );
    return;
  }

  const antwort = await fetch("https://api.resend.com/emails", {
    method: "POST",
    headers: {
      authorization: `Bearer ${schluessel}`,
      "content-type": "application/json",
    },
    body: JSON.stringify({ from: von, to: an, subject: betreff, text }),
  });

  if (!antwort.ok) {
    // Werfen und nicht schlucken: Better Auth meldet dem Nutzer dann, dass es
    // nicht geklappt hat. Eine Bestätigungsmail, die still verschwindet, ist
    // schlimmer als ein sichtbarer Fehler — der Mensch wartet sonst auf etwas,
    // das nie kommt.
    throw new Error(`Mailversand fehlgeschlagen (${antwort.status}): ${await antwort.text()}`);
  }
}

/**
 * Erreicht diese App **beliebige** Adressen — oder nur die eigene?
 *
 * Drei Zustände, und der mittlere ist der gefährliche:
 *
 *  1. Kein Schlüssel: `schicke` schreibt ins Log, niemand bekommt Mail.
 *  2. Schlüssel **und** `onboarding@resend.dev`: Resend stellt nur an die
 *     Adresse des eigenen Kontos zu und antwortet für alle anderen mit 403.
 *     Die App verhält sich dann gespalten — der Betreiber bekommt Mail, alle
 *     anderen nicht, und keiner der beiden merkt es.
 *  3. Schlüssel und eigene, verifizierte Domain: alle bekommen Mail.
 *
 * Nur im dritten Fall darf die Oberfläche eine Mail ankündigen. Deshalb steht
 * die Frage hier und nicht als zweite Umgebungsvariable: Ein `MAIL_AUS`-Schalter
 * neben `MAIL_VON` wären zwei Wahrheiten, die auseinanderlaufen können — und
 * zwar genau an dem Tag, an dem die Domain kommt und jemand vergisst, den
 * zweiten Schalter umzulegen. Abgeleitet kann das nicht passieren: Sobald
 * `MAIL_VON` auf die eigene Domain zeigt, stimmen alle Sätze von selbst.
 */
export function mailErreichtAlle(): boolean {
  const von = process.env.MAIL_VON ?? "";
  if (!process.env.RESEND_API_KEY || von === "") return false;
  return !von.includes("onboarding@resend.dev");
}

/** Die Adresse der App, für Links in Mails. */
export function appAdresse(): string {
  return process.env.BETTER_AUTH_URL ?? process.env.NEXT_PUBLIC_APP_URL ?? "http://localhost:3000";
}
