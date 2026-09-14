import Link from "next/link";
import { redirect } from "next/navigation";
import { EinrichtenFormular } from "@/app/components/einrichten-formular";
import { ladeHaushalte, ladeIch } from "@/lib/api";
import { serverToken } from "@/lib/auth-token";

/**
 * Den eigenen Haushalt einrichten.
 *
 * Server Component als Rahmen und als Weiche: Ohne Anmeldung geht es zum
 * Anmelden, mit vorhandenem Haushalt zurück zum Plan. Beides hier und nicht im
 * Formular — eine Weiterleitung, die erst nach dem Laden im Browser passiert,
 * zeigt dem Nutzer für einen Moment eine Seite, die ihn nichts angeht.
 */
export default async function Einrichten() {
  const token = await serverToken();
  if (!token) redirect("/anmelden?weiter=/einrichten");

  const [haushalte, ich] = await Promise.all([ladeHaushalte(token), ladeIch(token)]);
  if (haushalte.some((h) => h.meine_rolle)) redirect("/");

  const name = ich.name?.trim() || "";

  return (
    <main className="mx-auto w-full max-w-lg px-5 py-10">
      <h1 className="mb-2 text-2xl font-extrabold tracking-tight text-balance">
        Kurz einrichten
      </h1>
      <p className="mb-8 max-w-prose text-sm leading-relaxed text-muted text-pretty">
        Drei Fragen, keine Minute. Danach steht dein erster Wochenplan — und du
        kannst alles daran ändern. Was die App nicht weiß, schlägt sie auch
        nicht vor: Ohne Garten kein Rasen, ohne Auto kein TÜV.
      </p>

      <EinrichtenFormular meinName={name} />

      <p className="mt-8 text-sm text-muted">
        Lieber erst umsehen?{" "}
        <Link href="/?haushalt=familie-a" className="text-primary underline">
          Beispielhaushalt ansehen
        </Link>
      </p>
    </main>
  );
}
