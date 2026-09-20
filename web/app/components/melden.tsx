"use client";

import { useState } from "react";
import { postMitToken } from "@/lib/browser-token";

const arten = [
  { wert: "fehler", titel: "Stimmt nicht", hilfe: "Etwas ist falsch oder kaputt." },
  { wert: "stoert", titel: "Stört mich", hilfe: "Funktioniert, nervt trotzdem." },
  { wert: "idee", titel: "Idee", hilfe: "Das müsste die App doch …" },
] as const;

/**
 * Etwas melden — aus der App heraus, an der Stelle, an der es auffällt.
 *
 * Der Zettel am Kühlschrank wird abends ausgefüllt, aus dem Gedächtnis, von
 * der Person, die ihn aufgehängt hat. Was jemand am Mittwochmorgen dachte,
 * als der Plan etwas Falsches behauptete, steht abends nicht mehr im selben
 * Wortlaut da — und der Wortlaut ist bei einer Rückmeldung das Ganze.
 *
 * Drei Arten statt zwei. „Fehler" und „Idee" sind die üblichen, und dazwischen
 * fällt das Wichtigste durch: Die App tut, was sie soll, und es nervt
 * trotzdem. Dafür gibt es sonst keine Schublade — und es sind genau die Sätze,
 * wegen derer jemand eine App am Mittwoch zuklappt.
 *
 * Den Kontext sammelt die Oberfläche selbst. Wer erst beschreiben muss, wo er
 * gerade war, meldet nichts.
 */
export function Melden({ haushaltId }: { haushaltId: string }) {
  const [offen, setOffen] = useState(false);
  const [art, setArt] = useState<(typeof arten)[number]["wert"]>("fehler");
  const [text, setText] = useState("");
  const [laeuft, setLaeuft] = useState(false);
  const [fertig, setFertig] = useState(false);
  const [fehler, setFehler] = useState<string | null>(null);

  async function schicken(e: React.FormEvent) {
    e.preventDefault();
    setLaeuft(true);
    setFehler(null);
    try {
      // Direkt aus der Adresszeile statt über usePathname/useSearchParams:
      // Der Kontext wird erst beim Abschicken gebraucht, und dieser Code läuft
      // ohnehin nur im Browser. `useSearchParams` würde die ganze Seite in
      // eine Suspense-Grenze zwingen — viel Bauwerk für eine Zeichenkette.
      const ort = new URL(window.location.href);
      const woche = ort.searchParams.get("woche") ?? "";
      await postMitToken<void>(`/api/haushalte/${encodeURIComponent(haushaltId)}/rueckmeldungen`, {
        art,
        text: text.trim(),
        kontext: [ort.pathname, woche && `Woche ${woche}`].filter(Boolean).join(" · "),
      });
      setText("");
      setOffen(false);
      setFertig(true);
    } catch (e) {
      setFehler(e instanceof Error ? e.message : "Das hat nicht funktioniert.");
    } finally {
      setLaeuft(false);
    }
  }

  if (fertig && !offen) {
    return (
      <div className="mt-10 border-t border-line pt-6">
        <p className="text-sm text-primary">
          Angekommen. Danke — das ist mehr wert als es sich anfühlt.
        </p>
        <button
          type="button"
          onClick={() => {
            setFertig(false);
            setOffen(true);
          }}
          className="mt-1 inline-flex min-h-10 items-center text-sm font-semibold text-muted hover:text-fg"
        >
          Noch etwas melden
        </button>
      </div>
    );
  }

  if (!offen) {
    return (
      <div className="mt-10 border-t border-line pt-6">
        <button
          type="button"
          onClick={() => setOffen(true)}
          className="inline-flex min-h-11 items-center rounded-md px-3 text-sm font-semibold text-muted transition-colors hover:bg-surface-2 hover:text-fg"
        >
          Stimmt etwas nicht?
        </button>
      </div>
    );
  }

  return (
    <form onSubmit={schicken} className="mt-10 space-y-3 border-t border-line pt-6">
      <fieldset className="space-y-2">
        <legend className="text-sm font-semibold">Was ist es?</legend>
        <div className="flex flex-wrap gap-2">
          {arten.map((a) => (
            <button
              key={a.wert}
              type="button"
              onClick={() => setArt(a.wert)}
              aria-pressed={art === a.wert}
              className={`inline-flex min-h-10 items-center rounded-full border px-3.5 text-sm font-semibold transition-colors ${
                art === a.wert
                  ? "border-primary bg-primary-soft text-primary"
                  : "border-line-strong text-muted hover:text-fg"
              }`}
            >
              {a.titel}
            </button>
          ))}
        </div>
        <p className="text-xs text-muted">{arten.find((a) => a.wert === art)?.hilfe}</p>
      </fieldset>

      <label className="block space-y-1.5">
        <span className="block text-sm font-semibold">Sag es so, wie du es denkst</span>
        <textarea
          value={text}
          onChange={(e) => setText(e.target.value)}
          rows={4}
          maxLength={2000}
          autoFocus
          placeholder="Der Plan sagt, ich soll Dienstag kochen, aber dienstags bin ich nie da."
          className="block w-full rounded-md border border-line-strong bg-surface px-3 py-2.5 text-base placeholder:text-subtle"
        />
        {/* Kein „bitte beschreiben Sie die Schritte zur Reproduktion". Ein
            Formular, das nach Form fragt, bekommt weniger Inhalt. */}
        <span className="block text-xs text-muted">
          Ein Satz reicht. Wo du gerade bist, schickt die App von selbst mit.
        </span>
      </label>

      {fehler && (
        <p
          role="alert"
          className="rounded-md border border-danger/40 px-3 py-2 text-sm text-danger"
        >
          {fehler}
        </p>
      )}

      <div className="flex flex-wrap gap-2">
        <button
          type="submit"
          disabled={laeuft || text.trim() === ""}
          className="inline-flex min-h-11 items-center rounded-md bg-primary px-5 text-sm font-semibold text-on-primary transition-colors hover:bg-primary-hover disabled:opacity-45"
        >
          {laeuft ? "Einen Moment …" : "Abschicken"}
        </button>
        <button
          type="button"
          onClick={() => setOffen(false)}
          className="inline-flex min-h-11 items-center rounded-md px-3 text-sm font-semibold text-muted hover:text-fg"
        >
          Abbrechen
        </button>
      </div>
    </form>
  );
}
