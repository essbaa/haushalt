"use client";

import Link from "next/link";
import { useState } from "react";
import { Feld } from "@/app/components/ui";
import { requestPasswordReset } from "@/lib/auth-client";

/**
 * Passwort vergessen — Schritt eins: die Mail anfordern.
 *
 * Die Antwort ist absichtlich immer dieselbe, auch wenn es die Adresse nicht
 * gibt. Sonst wäre dieses Formular eine Auskunft darüber, wer hier ein Konto
 * hat — dieselbe Regel wie bei den Einladungscodes und beim fremden Haushalt:
 * verschwiegen wird die Existenz, nicht die Handlung.
 *
 * Diese Regel gilt aber nur für das Formular. Dass die App **überhaupt** keine
 * Mail verschickt, ist keine Auskunft über eine einzelne Adresse, sondern über
 * die App — das darf und muss dastehen. Ein Formular, das freundlich „die Mail
 * ist unterwegs" antwortet, obwohl nichts unterwegs ist, ist die schlimmere
 * Variante: Es sieht aus wie ein Weg und ist eine Sackgasse, und wer darauf
 * wartet, wartet bis Mittwoch.
 */
export function Anforderung() {
  const [email, setEmail] = useState("");
  const [laeuft, setLaeuft] = useState(false);
  const [geschickt, setGeschickt] = useState(false);

  async function absenden(e: React.FormEvent) {
    e.preventDefault();
    setLaeuft(true);
    await requestPasswordReset({ email, redirectTo: "/passwort-neu" });
    setLaeuft(false);
    setGeschickt(true);
  }

  return (
    <Rahmen>
      <h1 className="mb-2 text-2xl font-extrabold tracking-tight">Passwort vergessen</h1>

      {geschickt ? (
        <>
          <p className="mb-6 text-sm leading-relaxed text-muted text-pretty">
            Wenn es zu <strong className="text-fg">{email}</strong> ein Konto gibt, ist die Mail
            unterwegs. Der Link darin gilt eine Stunde.
          </p>
          <Link
            href="/anmelden"
            className="text-sm font-semibold text-primary underline underline-offset-2"
          >
            Zurück zum Anmelden
          </Link>
        </>
      ) : (
        <>
          <p className="mb-7 text-sm leading-relaxed text-muted text-pretty">
            Wir schicken dir einen Link, mit dem du ein neues setzen kannst.
          </p>
          <form onSubmit={absenden} className="space-y-4">
            <Feld
              beschriftung="E-Mail"
              type="email"
              inputMode="email"
              value={email}
              onChange={(e) => setEmail(e.target.value)}
              autoComplete="email"
              required
            />
            <button
              type="submit"
              disabled={laeuft}
              className="inline-flex min-h-12 w-full items-center justify-center rounded-md bg-primary px-4 text-sm font-semibold text-on-primary transition-colors hover:bg-primary-hover disabled:opacity-45"
            >
              {laeuft ? "Einen Moment …" : "Link schicken"}
            </button>
          </form>
          <Link
            href="/anmelden"
            className="mt-6 inline-flex min-h-11 items-center text-sm font-semibold text-primary underline underline-offset-2"
          >
            Doch nicht
          </Link>
        </>
      )}
    </Rahmen>
  );
}

function Rahmen({ children }: { children: React.ReactNode }) {
  return (
    <main className="mx-auto flex w-full max-w-sm flex-1 flex-col justify-center px-5 py-12">
      <div className="rounded-lg border border-line bg-surface p-6 sm:p-7">{children}</div>
    </main>
  );
}
