"use client";

import { useRouter } from "next/navigation";
import { useState } from "react";
import type { Frage } from "@/lib/api";
import { patchMitToken } from "@/lib/browser-token";

/**
 * Die offenen Fragen zum Haushalt.
 *
 * Der Gegenentwurf zum Onboarding-Fragebogen: höchstens zwei Fragen, jede mit
 * dem Nutzen daneben, jede in einem Tipp beantwortet. Wer nicht weiß, wofür er
 * antwortet, antwortet nicht — deshalb steht unter jeder Frage, was ein Ja
 * bringt.
 *
 * „Später" gibt es nicht: Die Frage kommt von selbst wieder, solange die
 * Antwort fehlt. Ein Knopf, der nur wegräumt, wäre ein dritter Zustand, den
 * niemand pflegt.
 */
export function Fragen({ haushaltId, fragen }: { haushaltId: string; fragen: Frage[] }) {
  const router = useRouter();
  const [laeuft, setLaeuft] = useState<string | null>(null);
  const [fehler, setFehler] = useState<string | null>(null);

  async function antworten(faktum: string, wert: boolean) {
    setLaeuft(faktum);
    setFehler(null);
    try {
      await patchMitToken(`/api/haushalte/${encodeURIComponent(haushaltId)}/fakten`, {
        [faktum]: wert,
      });
      router.refresh();
    } catch (e) {
      setFehler(e instanceof Error ? e.message : "Das hat nicht funktioniert.");
    } finally {
      setLaeuft(null);
    }
  }

  return (
    <section className="space-y-3 border-t border-line pt-6">
      <h2 className="text-lg font-bold tracking-tight">
        {fragen.length === 1 ? "Eine Frage" : "Zwei Fragen"}
      </h2>
      <p className="text-sm leading-relaxed text-muted">
        Was die App nicht weiß, plant sie nicht ein. Jede Antwort gilt dauerhaft.
      </p>

      <ul className="space-y-3">
        {fragen.map((f) => (
          <li key={f.faktum} className="rounded-lg border border-line bg-surface p-4">
            <p className="font-semibold text-pretty">{f.frage}</p>
            <p className="mt-1 text-sm leading-relaxed text-muted text-pretty">{f.dann}</p>
            <div className="mt-3 flex flex-wrap gap-2">
              <button
                type="button"
                onClick={() => antworten(f.faktum, true)}
                disabled={laeuft !== null}
                className="inline-flex min-h-11 items-center rounded-md bg-primary px-5 text-sm font-semibold text-on-primary transition-colors hover:bg-primary-hover disabled:opacity-45"
              >
                {laeuft === f.faktum ? "Einen Moment …" : "Ja"}
              </button>
              <button
                type="button"
                onClick={() => antworten(f.faktum, false)}
                disabled={laeuft !== null}
                className="inline-flex min-h-11 items-center rounded-md border border-line-strong px-5 text-sm font-semibold text-muted transition-colors hover:border-line-strong hover:text-fg disabled:opacity-45"
              >
                Nein
              </button>
            </div>
          </li>
        ))}
      </ul>

      {fehler && (
        <p role="alert" className="text-sm text-danger">
          {fehler}
        </p>
      )}
    </section>
  );
}
