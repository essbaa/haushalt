"use client";

import { useRouter } from "next/navigation";
import { useState } from "react";
import { postMitToken } from "@/lib/browser-token";

/**
 * Abhaken und Abgeben — die zwei Handgriffe, die einen Plan von einem
 * Dienstplan unterscheiden.
 *
 * Aus dem Produktkonzept: „Ausführende brauchen Handlungsmacht, nicht nur
 * Pflichten. Eine reine Empfangsliste wird gelöscht." Abgeben ist die kleinste
 * Form davon, und es darf keine Verhandlung sein — ein Klick, ein Grund,
 * fertig. Wer erst fragen muss, gibt nicht ab, sondern schweigt.
 */
export function AufgabeAktionen({
  aufgabeId,
  erledigt,
  abgebbar,
}: {
  aufgabeId: string;
  erledigt: boolean;
  abgebbar: boolean;
}) {
  const router = useRouter();
  const [laeuft, setLaeuft] = useState(false);
  const [fragt, setFragt] = useState(false);
  const [grund, setGrund] = useState("");
  const [meldung, setMeldung] = useState<string | null>(null);

  async function haken() {
    setLaeuft(true);
    setMeldung(null);
    try {
      await postMitToken<void>(`/api/aufgaben/${encodeURIComponent(aufgabeId)}/erledigt`, {
        erledigt: !erledigt,
      });
      router.refresh();
    } catch (e) {
      setMeldung(e instanceof Error ? e.message : "Das hat nicht funktioniert.");
    } finally {
      setLaeuft(false);
    }
  }

  async function abgeben() {
    setLaeuft(true);
    setMeldung(null);
    try {
      const antwort = await postMitToken<{ uebernimmt?: string }>(
        `/api/aufgaben/${encodeURIComponent(aufgabeId)}/abgeben`,
        { grund: grund.trim() || undefined },
      );
      setFragt(false);
      setGrund("");
      // Wer übernimmt, sagt die Antwort — sonst steht die Aufgabe offen da.
      setMeldung(
        antwort.uebernimmt
          ? `${antwort.uebernimmt} übernimmt.`
          : "Steht jetzt offen im Plan — niemand sonst kann sie diese Woche.",
      );
      router.refresh();
    } catch (e) {
      setMeldung(e instanceof Error ? e.message : "Das hat nicht funktioniert.");
    } finally {
      setLaeuft(false);
    }
  }

  return (
    <span className="flex flex-wrap items-center gap-3">
      <button
        type="button"
        onClick={haken}
        disabled={laeuft}
        aria-pressed={erledigt}
        className={`rounded-md border px-2.5 py-1 font-mono text-xs transition-colors disabled:opacity-50 ${
          erledigt
            ? "border-accent bg-accent/10 text-accent"
            : "border-line text-muted hover:text-foreground"
        }`}
      >
        {erledigt ? "✓ erledigt" : "abhaken"}
      </button>

      {abgebbar && !erledigt && !fragt && (
        <button
          type="button"
          onClick={() => setFragt(true)}
          disabled={laeuft}
          className="text-xs text-muted underline disabled:opacity-50"
        >
          abgeben
        </button>
      )}

      {fragt && (
        <span className="flex w-full flex-wrap items-center gap-2">
          <input
            value={grund}
            onChange={(e) => setGrund(e.target.value)}
            placeholder="Warum? (freiwillig)"
            maxLength={200}
            className="min-w-0 flex-1 rounded-md border border-line bg-transparent px-2 py-1 text-xs"
          />
          <button
            type="button"
            onClick={abgeben}
            disabled={laeuft}
            className="rounded-md border border-line px-2.5 py-1 text-xs hover:border-accent hover:text-accent disabled:opacity-50"
          >
            zurückgeben
          </button>
          <button
            type="button"
            onClick={() => setFragt(false)}
            className="text-xs text-muted underline"
          >
            doch nicht
          </button>
        </span>
      )}

      {meldung && <span className="w-full text-xs text-muted">{meldung}</span>}
    </span>
  );
}
