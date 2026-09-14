"use client";

import { useRouter } from "next/navigation";
import { useState } from "react";
import { Feld } from "@/app/components/ui";
import type { Anlass } from "@/lib/api";
import { deleteMitToken, postMitToken } from "@/lib/browser-token";
import { tagLesbar } from "@/lib/woche";

const arten: { wert: Anlass["art"]; text: string; jaehrlich: boolean }[] = [
  { wert: "kindergeburtstag", text: "Kindergeburtstag", jaehrlich: true },
  { wert: "geburtstag", text: "Geburtstag", jaehrlich: true },
  { wert: "elternabend", text: "Elternabend", jaehrlich: false },
  { wert: "arzttermin", text: "Arzttermin", jaehrlich: false },
  { wert: "sonstiges", text: "Sonstiges", jaehrlich: false },
];

export function TermineListe({
  haushaltId,
  anlaesse,
  planend,
}: {
  haushaltId: string;
  anlaesse: Anlass[];
  planend: boolean;
}) {
  const router = useRouter();
  const [titel, setTitel] = useState("");
  const [tag, setTag] = useState("");
  const [art, setArt] = useState<Anlass["art"]>("kindergeburtstag");
  const [laeuft, setLaeuft] = useState(false);
  const [fehler, setFehler] = useState<string | null>(null);

  // Die Wiederholung hängt an der Art, nicht an einem eigenen Schalter: Ein
  // Geburtstag kommt jedes Jahr wieder, ein Arzttermin nicht. Danach zu fragen
  // wäre eine Frage, deren Antwort die App schon kennt.
  const jaehrlich = arten.find((a) => a.wert === art)?.jaehrlich ?? false;

  async function eintragen(e: React.FormEvent) {
    e.preventDefault();
    setLaeuft(true);
    setFehler(null);
    try {
      await postMitToken<Anlass>(`/api/haushalte/${encodeURIComponent(haushaltId)}/anlaesse`, {
        titel: titel.trim(),
        tag,
        art,
        jaehrlich,
      });
      setTitel("");
      setTag("");
      router.refresh();
    } catch (e) {
      setFehler(e instanceof Error ? e.message : "Das hat nicht funktioniert.");
    } finally {
      setLaeuft(false);
    }
  }

  async function loeschen(id: string) {
    setLaeuft(true);
    setFehler(null);
    try {
      await deleteMitToken<void>(
        `/api/haushalte/${encodeURIComponent(haushaltId)}/anlaesse/${encodeURIComponent(id)}`,
      );
      router.refresh();
    } catch (e) {
      setFehler(e instanceof Error ? e.message : "Das hat nicht funktioniert.");
    } finally {
      setLaeuft(false);
    }
  }

  return (
    <div className="space-y-8">
      {anlaesse.length === 0 ? (
        <p className="rounded-lg border border-line bg-surface px-4 py-6 text-sm leading-relaxed text-muted">
          Noch nichts eingetragen. Solange hier nichts steht, entstehen auch
          keine Aufgaben mit Vorlauf — kein erfundener Geburtstag im Plan.
        </p>
      ) : (
        <ul className="divide-y divide-line rounded-lg border border-line bg-surface">
          {anlaesse.map((a) => (
            <li key={a.id} className="flex flex-wrap items-center gap-x-3 gap-y-1 px-4 py-3">
              <div className="min-w-0 flex-1">
                <p className="font-semibold">{a.titel}</p>
                <p className="text-xs text-subtle">
                  {tagLesbar(a.tag)}
                  {a.jaehrlich && ", jedes Jahr"}
                  {" — "}
                  {arten.find((x) => x.wert === a.art)?.text ?? a.art}
                </p>
              </div>
              {planend && (
                <button
                  type="button"
                  onClick={() => loeschen(a.id)}
                  disabled={laeuft}
                  className="inline-flex min-h-10 shrink-0 items-center rounded-md px-3 text-sm font-semibold text-subtle transition-colors hover:bg-surface-2 hover:text-danger disabled:opacity-45"
                >
                  Löschen
                </button>
              )}
            </li>
          ))}
        </ul>
      )}

      {planend && (
        <form onSubmit={eintragen} className="space-y-4 rounded-lg border border-line bg-surface p-4 sm:p-5">
          <h2 className="text-lg font-bold tracking-tight">Anlass eintragen</h2>

          <Feld
            beschriftung="Was ist es?"
            value={titel}
            onChange={(e) => setTitel(e.target.value)}
            placeholder="Geburtstag von Mia"
            maxLength={80}
            required
          />

          <Feld
            beschriftung="Wann?"
            type="date"
            value={tag}
            onChange={(e) => setTag(e.target.value)}
            required
          />

          <fieldset className="space-y-2">
            <legend className="text-sm font-semibold">Was für ein Anlass?</legend>
            <div className="flex flex-wrap gap-2">
              {arten.map((o) => (
                <button
                  key={o.wert}
                  type="button"
                  onClick={() => setArt(o.wert)}
                  aria-pressed={art === o.wert}
                  className={`inline-flex min-h-10 items-center rounded-full border px-3.5 text-sm font-semibold transition-colors ${
                    art === o.wert
                      ? "border-primary bg-primary-soft text-primary"
                      : "border-line-strong text-muted hover:text-fg"
                  }`}
                >
                  {o.text}
                </button>
              ))}
            </div>
            <p className="text-xs leading-relaxed text-muted">
              {jaehrlich
                ? "Wiederholt sich jedes Jahr."
                : "Ein einmaliger Termin."}
            </p>
          </fieldset>

          {fehler && (
            <p role="alert" className="rounded-md border border-danger/40 px-3 py-2 text-sm text-danger">
              {fehler}
            </p>
          )}

          <button
            type="submit"
            disabled={laeuft || titel.trim() === "" || tag === ""}
            className="inline-flex min-h-12 w-full items-center justify-center rounded-md bg-primary px-4 text-sm font-semibold text-on-primary transition-colors hover:bg-primary-hover disabled:opacity-45"
          >
            {laeuft ? "Einen Moment …" : "Eintragen"}
          </button>
        </form>
      )}
    </div>
  );
}
