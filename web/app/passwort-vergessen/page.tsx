import Link from "next/link";
import { Anforderung } from "@/app/passwort-vergessen/formular";
import { mailErreichtAlle } from "@/lib/mail";

/**
 * Passwort vergessen — oder eben: noch nicht.
 *
 * Der Server weiß, ob die App zustellen kann; der Browser darf es nicht
 * wissen (`RESEND_API_KEY` ist ein Geheimnis). Also entscheidet die Seite,
 * und das Formular kommt nur, wenn es etwas bewirkt.
 */
export default function Seite() {
  if (mailErreichtAlle()) return <Anforderung />;

  return (
    <main className="mx-auto flex w-full max-w-sm flex-1 flex-col justify-center px-5 py-12">
      <div className="rounded-lg border border-line bg-surface p-6 sm:p-7">
        <h1 className="mb-2 text-2xl font-extrabold tracking-tight">Das geht noch nicht</h1>
        <p className="mb-6 text-sm leading-relaxed text-muted text-pretty">
          Die App kann noch keine Mails verschicken — ihr fehlt eine eigene Adresse, von der aus sie
          das dürfte. Solange das so ist, gibt es keinen Link zum Zurücksetzen.
        </p>
        <p className="mb-7 text-sm leading-relaxed text-muted text-pretty">
          Sag der Person Bescheid, die euren Haushalt eingerichtet hat. Sie kann dein Passwort von
          Hand neu setzen.
        </p>
        <Link
          href="/anmelden"
          className="inline-flex min-h-11 items-center text-sm font-semibold text-primary underline underline-offset-2"
        >
          Zurück zum Anmelden
        </Link>
      </div>
    </main>
  );
}
