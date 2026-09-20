"use client";

import { useRouter } from "next/navigation";
import { useState } from "react";
import { Feld } from "@/app/components/ui";
import type { VorlagenStand } from "@/lib/api";
import { postMitToken } from "@/lib/browser-token";

const rhythmen = [
  { tage: 1, text: "täglich" },
  { tage: 7, text: "wöchentlich" },
  { tage: 14, text: "alle zwei Wochen" },
  { tage: 30, text: "monatlich" },
];

/**
 * Eine Aufgabe anlegen, die es in der Bibliothek nicht gibt.
 *
 * Fünf Fragen, und keine davon verlangt Kenntnis des Modells. Vor allem die
 * Kopflast wird nicht als Zahl abgefragt: „Kopflast 0 bis 3" würde geraten,
 * „muss jemand daran denken?" wird beantwortet.
 */
export function EigeneAufgabe({
  haushaltId,
  kategorie,
  offen,
  setOffen,
}: {
  haushaltId: string;
  kategorie: string;
  /** Von außen gesteuert, damit das „+" in der Bereichszeile beides kann:
   *  den Bereich aufklappen und gleich das Formular öffnen. */
  offen: boolean;
  setOffen: (o: boolean) => void;
}) {
  const router = useRouter();
  const [titel, setTitel] = useState("");
  const [dauer, setDauer] = useState("20");
  const [tage, setTage] = useState(7);
  const [denken, setDenken] = useState(false);
  const [klaeren, setKlaeren] = useState(false);
  const [nurErwachsene, setNurErwachsene] = useState(false);
  const [absprache, setAbsprache] = useState(false);
  const [laeuft, setLaeuft] = useState(false);
  const [fehler, setFehler] = useState<string | null>(null);

  async function anlegen(e: React.FormEvent) {
    e.preventDefault();
    setLaeuft(true);
    setFehler(null);
    try {
      await postMitToken<VorlagenStand>(
        `/api/haushalte/${encodeURIComponent(haushaltId)}/vorlagen`,
        {
          titel: titel.trim(),
          kategorie,
          dauer_min: Number(dauer),
          alle_tage: tage,
          denken,
          klaeren,
          nur_erwachsene: nurErwachsene,
          braucht_absprache: absprache,
        },
      );
      setTitel("");
      setOffen(false);
      router.refresh();
    } catch (e) {
      setFehler(e instanceof Error ? e.message : "Das hat nicht funktioniert.");
    } finally {
      setLaeuft(false);
    }
  }

  if (!offen) {
    return (
      <button
        type="button"
        onClick={() => setOffen(true)}
        className="inline-flex min-h-11 items-center gap-1.5 rounded-md px-3 text-sm font-semibold text-primary transition-colors hover:bg-primary-soft"
      >
        <span aria-hidden="true">+</span>
        Eigene Aufgabe
      </button>
    );
  }

  return (
    <form onSubmit={anlegen} className="space-y-4 rounded-lg border border-line bg-surface p-4">
      <Feld
        beschriftung="Was ist zu tun?"
        value={titel}
        onChange={(e) => setTitel(e.target.value)}
        placeholder="Medikamente sortieren"
        maxLength={80}
        autoFocus
        required
      />

      <Feld
        beschriftung="Wie lange dauert es ungefähr?"
        type="number"
        inputMode="numeric"
        min={1}
        max={480}
        value={dauer}
        onChange={(e) => setDauer(e.target.value)}
        hinweis="In Minuten. Die Zahl trägt die Verteilung mit — lieber großzügig schätzen als zu knapp."
        required
      />

      <fieldset className="space-y-2">
        <legend className="text-sm font-semibold">Wie oft?</legend>
        <div className="flex flex-wrap gap-2">
          {rhythmen.map((r) => (
            <button
              key={r.tage}
              type="button"
              onClick={() => setTage(r.tage)}
              aria-pressed={tage === r.tage}
              className={`inline-flex min-h-10 items-center rounded-full border px-3.5 text-sm font-semibold transition-colors ${
                tage === r.tage
                  ? "border-primary bg-primary-soft text-primary"
                  : "border-line-strong text-muted hover:text-fg"
              }`}
            >
              {r.text}
            </button>
          ))}
        </div>
      </fieldset>

      {/* Die Kopflast — die wichtigste Zahl im Produkt und die einzige, die
          niemand von außen versteht. Deshalb zwei Fragen statt einer Skala. */}
      <fieldset className="space-y-2">
        <legend className="text-sm font-semibold">Wie viel Kopf kostet es?</legend>
        <label className="flex items-start gap-2.5 text-sm">
          <input
            type="checkbox"
            checked={denken}
            onChange={(e) => setDenken(e.target.checked)}
            className="mt-1"
          />
          <span>
            Jemand muss daran <strong>denken</strong> — man sieht es nicht von selbst.
          </span>
        </label>
        <label className="flex items-start gap-2.5 text-sm">
          <input
            type="checkbox"
            checked={klaeren}
            onChange={(e) => setKlaeren(e.target.checked)}
            className="mt-1"
          />
          <span>
            Man muss erst etwas <strong>klären</strong> — einen Termin machen, nachsehen, jemanden
            fragen.
          </span>
        </label>
      </fieldset>

      <label className="flex items-center gap-2.5 text-sm">
        <input
          type="checkbox"
          checked={nurErwachsene}
          onChange={(e) => setNurErwachsene(e.target.checked)}
        />
        Nur Erwachsene können das übernehmen
      </label>

      {/* Die Frage, die entscheidet, ob der Planer diese Aufgabe überhaupt
          verteilt. Er verteilt nach Kapazität — nach verfügbarer Zeit. Wo
          nicht die Zeit entscheidet, sondern der Arbeitsplan, würde er mit
          voller Überzeugung den Falschen einteilen (ADR-0016). */}
      <label className="flex items-start gap-2.5 text-sm">
        <input
          type="checkbox"
          checked={absprache}
          onChange={(e) => setAbsprache(e.target.checked)}
          className="mt-1"
        />
        <span>
          Das müssen wir <strong>absprechen</strong> — wer das kann, hängt an den Arbeitszeiten und
          nicht daran, wer Zeit übrig hat. Ihr tragt dann selbst ein, wer an welchem Tag dran ist.
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
          disabled={laeuft || titel.trim() === ""}
          className="inline-flex min-h-11 items-center rounded-md bg-primary px-5 text-sm font-semibold text-on-primary transition-colors hover:bg-primary-hover disabled:opacity-45"
        >
          {laeuft ? "Einen Moment …" : "Anlegen"}
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
