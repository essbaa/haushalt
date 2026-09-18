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
      <h1 className="mb-2 text-2xl font-extrabold tracking-tight text-balance">Kurz einrichten</h1>
      <p className="mb-6 max-w-prose text-sm leading-relaxed text-muted text-pretty">
        Drei Fragen, keine Minute. Danach steht dein erster Wochenplan — und du kannst alles daran
        ändern. Was die App nicht weiß, schlägt sie auch nicht vor: Ohne Garten kein Rasen, ohne
        Auto kein TÜV.
      </p>

      {/* Der zweite Weg, und er steht **vor** dem Formular.
          Hierher kommt jeder ohne Haushalt — auch jemand, der gerade eingeladen
          wurde und nach der Registrierung die Seite neu lädt, statt auf
          „Weiter" zu tippen. Ohne diesen Kasten bietet der Bildschirm dann nur
          eine Handlung an, und es ist die falsche: Sie legt einen zweiten,
          leeren Haushalt an, die Einladung bleibt offen und ist nirgends mehr
          zu sehen. Ein Bildschirm, der nur einen Weg zeigt, behauptet damit,
          es gebe nur einen. */}
      <div className="mb-8 flex flex-wrap items-center justify-between gap-x-4 gap-y-2 rounded-lg border border-line bg-surface px-4 py-3.5">
        <p className="text-sm font-semibold text-pretty">Du bist eingeladen worden?</p>
        <Link
          href="/beitreten"
          className="inline-flex min-h-11 items-center text-sm font-semibold text-primary underline underline-offset-2"
        >
          Code eingeben
        </Link>
      </div>

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
