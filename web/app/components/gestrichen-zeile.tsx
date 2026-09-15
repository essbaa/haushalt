"use client";

import { useRouter } from "next/navigation";
import { useState, useTransition } from "react";
import { postMitToken } from "@/lib/browser-token";
import { tagLesbar } from "@/lib/woche";

/**
 * Eine gestrichene Aufgabe mit dem Weg zurück.
 *
 * Der Knopf heißt „Doch wieder einplanen" und nicht „Rückgängig": Er beschreibt
 * das Ergebnis, nicht die Bedienung. Wer ihn eine Woche später liest, weiß
 * sonst nicht mehr, was da rückgängig gemacht wird.
 */
export function GestrichenZeile({
  id,
  titel,
  tag,
}: {
  id: string;
  titel: string;
  tag: string;
}) {
  const router = useRouter();
  const [uebergang, starten] = useTransition();
  const [fehler, setFehler] = useState<string | null>(null);

  async function zurueck() {
    setFehler(null);
    try {
      await postMitToken<void>(`/api/aufgaben/${encodeURIComponent(id)}/streichen`, {
        gestrichen: false,
      });
      starten(() => router.refresh());
    } catch (e) {
      setFehler(e instanceof Error ? e.message : "Das hat nicht funktioniert.");
    }
  }

  return (
    <li className="flex flex-wrap items-center gap-x-3 gap-y-1 rounded-lg px-3 py-2">
      <span className="min-w-0 flex-1 text-sm text-muted">
        <span className="line-through">{titel}</span>{" "}
        <span className="text-subtle">· {tagLesbar(tag)}</span>
      </span>
      <button
        type="button"
        onClick={zurueck}
        disabled={uebergang}
        className="inline-flex min-h-9 items-center rounded-full px-3 text-sm font-semibold text-primary transition-colors hover:bg-primary-soft disabled:opacity-45"
      >
        Doch wieder einplanen
      </button>
      {fehler && (
        <p role="alert" className="w-full text-sm text-danger">
          {fehler}
        </p>
      )}
    </li>
  );
}
