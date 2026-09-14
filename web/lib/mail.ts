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

/** Die Adresse der App, für Links in Mails. */
export function appAdresse(): string {
  return process.env.BETTER_AUTH_URL ?? process.env.NEXT_PUBLIC_APP_URL ?? "http://localhost:3000";
}
