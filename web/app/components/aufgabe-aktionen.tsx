"use client";

import { useRouter } from "next/navigation";
import { useState } from "react";
import { Haken, Kreis, Zurueck } from "@/app/components/icons";
import { postMitToken } from "@/lib/browser-token";

/**
 * Abhaken und Abgeben — die zwei Handgriffe, die einen Plan von einem
 * Dienstplan unterscheiden.
 *
 * Aus dem Produktkonzept: „Ausführende brauchen Handlungsmacht, nicht nur
 * Pflichten. Eine reine Empfangsliste wird gelöscht." Abgeben ist die kleinste
 * Form davon, und es darf keine Verhandlung sein — ein Tipp, ein Grund,
 * fertig.
 */
export function AufgabeAktionen({
  aufgabeId,
  erledigt,
  abgebbar,
  eigene,
}: {
  aufgabeId: string;
  erledigt: boolean;
  abgebbar: boolean;
  /** Die eigene Aufgabe wird angefasst, fremde nur ausnahmsweise — deshalb
   *  steht dort ein leiser Knopf statt zweier auffälliger. */
  eigene: boolean;
}) {
  const router = useRouter();
  const [laeuft, setLaeuft] = useState(false);
  const [fragt, setFragt] = useState(false);
  const [grund, setGrund] = useState("");
  const [meldung, setMeldung] = useState<string | null>(null);
  const [fehler, setFehler] = useState<string | null>(null);

  async function tun(arbeit: () => Promise<void>) {
    setLaeuft(true);
    setFehler(null);
    try {
      await arbeit();
      router.refresh();
    } catch (e) {
      setFehler(e instanceof Error ? e.message : "Das hat nicht funktioniert.");
    } finally {
      setLaeuft(false);
    }
  }

  const haken = () =>
    tun(async () => {
      await postMitToken<void>(`/api/aufgaben/${encodeURIComponent(aufgabeId)}/erledigt`, {
        erledigt: !erledigt,
      });
      setMeldung(null);
    });

  const abgeben = () =>
    tun(async () => {
      const antwort = await postMitToken<{ uebernimmt?: string }>(
        `/api/aufgaben/${encodeURIComponent(aufgabeId)}/abgeben`,
        { grund: grund.trim() || undefined },
      );
      setFragt(false);
      setGrund("");
      setMeldung(
        antwort.uebernimmt
          ? `${antwort.uebernimmt} übernimmt.`
          : "Steht jetzt offen — niemand sonst kann sie diese Woche.",
      );
    });

  return (
    <div className="space-y-2 pt-1">
      <div className="flex flex-wrap items-center gap-2">
        <button
          type="button"
          onClick={haken}
          disabled={laeuft}
          aria-pressed={erledigt}
          className={`-ml-1 inline-flex min-h-9 items-center gap-1.5 rounded-full px-3 text-sm font-semibold transition-colors disabled:opacity-45 ${
            erledigt
              ? // Erledigtes wird leise. Ein gefüllter Knopf auf der Zeile,
                // die niemanden mehr interessiert, zieht den Blick genau
                // dorthin, wo nichts mehr zu tun ist.
                "text-primary hover:bg-surface-2"
              : eigene
                ? "border border-line-strong text-fg hover:border-primary hover:text-primary"
                : "text-subtle hover:bg-surface-2 hover:text-fg"
          }`}
        >
          {erledigt ? <Haken className="size-4" /> : <Kreis className="size-4" />}
          {erledigt ? "Erledigt" : "Abhaken"}
        </button>

        {abgebbar && !erledigt && !fragt && (
          <button
            type="button"
            onClick={() => setFragt(true)}
            disabled={laeuft}
            className="inline-flex min-h-9 items-center gap-1.5 rounded-full px-3 text-sm font-semibold text-muted transition-colors hover:bg-surface-2 hover:text-fg disabled:opacity-45"
          >
            <Zurueck className="size-4" />
            Abgeben
          </button>
        )}
      </div>

      {fragt && (
        <div className="space-y-2 rounded-md border border-line bg-surface-2 p-3">
          <label className="block space-y-1.5">
            <span className="text-sm font-semibold">Warum gibst du sie ab?</span>
            <input
              value={grund}
              onChange={(e) => setGrund(e.target.value)}
              placeholder="Freiwillig"
              maxLength={200}
              autoFocus
              className="block min-h-11 w-full rounded-md border border-line-strong bg-surface px-3 text-base placeholder:text-subtle"
            />
          </label>
          <div className="flex flex-wrap gap-2">
            <button
              type="button"
              onClick={abgeben}
              disabled={laeuft}
              className="inline-flex min-h-10 items-center rounded-md bg-primary px-4 text-sm font-semibold text-on-primary transition-colors hover:bg-primary-hover disabled:opacity-45"
            >
              Zurückgeben
            </button>
            <button
              type="button"
              onClick={() => setFragt(false)}
              className="inline-flex min-h-10 items-center rounded-md px-3 text-sm font-semibold text-muted hover:text-fg"
            >
              Doch nicht
            </button>
          </div>
        </div>
      )}

      {meldung && <p className="text-sm text-primary">{meldung}</p>}
      {fehler && (
        <p role="alert" className="text-sm text-danger">
          {fehler}
        </p>
      )}
    </div>
  );
}
