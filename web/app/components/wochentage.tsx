"use client";

import { useRouter } from "next/navigation";
import { useState, useTransition } from "react";
import { putMitToken } from "@/lib/browser-token";

const tage = ["Mo", "Di", "Mi", "Do", "Fr", "Sa", "So"];
const lang = ["Montag", "Dienstag", "Mittwoch", "Donnerstag", "Freitag", "Samstag", "Sonntag"];

/**
 * An welchen Wochentagen diese Aufgabe liegt.
 *
 * Der Fall, der sie nötig macht: Müll rausbringen steht in der Bibliothek auf
 * Dienstag, weil irgendein Tag dastehen musste. Bei euch kommt die Tonne
 * donnerstags. Bisher gab es dagegen genau ein Mittel — die Vorlage
 * abbestellen und von Hand neu anlegen, womit die Zahlen der Bibliothek
 * verloren gehen und die Aufgabe als geschätzte „eigene" dasteht. Ein Tag,
 * den man nicht ändern kann, ist nicht einmal falsch, sondern jede Woche.
 *
 * **Die Wahl ändert die Häufigkeit mit**, und das steht auch da. Wer bei
 * „alle 14 Tage" den Samstag ankreuzt, bekommt jeden Samstag — nicht jeden
 * zweiten. Das ist die Aussage und keine Nebenwirkung; sie zu verschweigen
 * hieße, den Menschen einen Plan zu geben, der doppelt so voll ist wie der,
 * den er gesetzt hat.
 *
 * Kein Tag angekreuzt heißt: zurück zur Vorgabe. Deshalb gibt es keinen
 * eigenen „Zurücksetzen"-Knopf neben dem Speichern — er wäre ein zweiter Weg
 * zu demselben Zustand, und zwei Wege zu einem Zustand sind die Gelegenheit,
 * sie auseinanderlaufen zu lassen. Der Satz darunter sagt es stattdessen.
 */
export function Wochentage({
  haushaltId,
  vorlageId,
  titel,
  gesetzt,
  eigen,
  planend,
}: {
  haushaltId: string;
  vorlageId: string;
  titel: string;
  /** Sieben Werte, Index 0 = Montag. */
  gesetzt: boolean[];
  /** Die Tage stammen von diesem Haushalt, nicht aus der Bibliothek. */
  eigen: boolean;
  planend: boolean;
}) {
  const router = useRouter();
  const [offen, setOffen] = useState(false);
  const [entwurf, setEntwurf] = useState<boolean[]>(gesetzt);
  const [laeuft, setLaeuft] = useState(false);
  const [fehler, setFehler] = useState<string | null>(null);
  const [, starten] = useTransition();

  const gewaehlt = entwurf.filter(Boolean).length;

  async function speichern() {
    setLaeuft(true);
    setFehler(null);
    try {
      await putMitToken<void>(
        `/api/haushalte/${encodeURIComponent(haushaltId)}/wochentage/${encodeURIComponent(vorlageId)}`,
        { wochentage: entwurf },
      );
      setOffen(false);
      starten(() => router.refresh());
    } catch (e) {
      setFehler(e instanceof Error ? e.message : "Das hat nicht funktioniert.");
    } finally {
      setLaeuft(false);
    }
  }

  // Was gerade gilt, als Satz. Für Leute ohne Planungsrecht ist das alles —
  // sie sollen sehen, woran sie sind, und nichts verstellen können.
  const stand = gesetzt.some(Boolean)
    ? tage.filter((_, i) => gesetzt[i]).join(", ")
    : "kein fester Tag";

  if (!planend) {
    return (
      <p className="w-full text-xs text-subtle">
        {stand}
        {eigen && " — von euch festgelegt"}
      </p>
    );
  }

  if (!offen) {
    return (
      <div className="flex w-full flex-wrap items-center gap-x-3 gap-y-1">
        <p className="text-xs text-subtle">
          {stand}
          {eigen && " — von euch festgelegt"}
        </p>
        <button
          type="button"
          onClick={() => {
            setEntwurf(gesetzt);
            setOffen(true);
          }}
          className="inline-flex min-h-9 items-center text-xs font-semibold text-primary underline underline-offset-2"
        >
          Tag ändern
        </button>
      </div>
    );
  }

  return (
    <div className="w-full space-y-3 rounded-lg border border-line bg-surface-2 p-4">
      <p className="text-sm font-semibold text-pretty">
        An welchen Tagen steht &bdquo;{titel}&ldquo; an?
      </p>

      <div className="flex flex-wrap gap-1.5">
        {tage.map((t, i) => (
          <button
            key={t}
            type="button"
            aria-pressed={entwurf[i]}
            aria-label={lang[i]}
            onClick={() => setEntwurf((alt) => alt.map((w, j) => (i === j ? !w : w)))}
            className={`inline-flex size-11 items-center justify-center rounded-full border text-sm font-bold transition-colors ${
              entwurf[i]
                ? "border-primary bg-primary text-on-primary"
                : "border-line-strong text-muted hover:text-fg"
            }`}
          >
            {t}
          </button>
        ))}
      </div>

      <p className="text-xs leading-relaxed text-muted text-pretty">
        {gewaehlt === 0
          ? "Kein Tag gewählt: Dann gilt wieder der Rhythmus aus der Bibliothek, und der Planer sucht sich selbst einen Tag."
          : gewaehlt === 1
            ? "Steht ab dann jede Woche an diesem Tag — auch wenn die Aufgabe vorher seltener dran war."
            : `Steht ab dann jede Woche an diesen ${gewaehlt} Tagen — auch wenn die Aufgabe vorher seltener dran war.`}
      </p>

      {fehler && (
        <p role="alert" className="text-xs font-semibold text-danger">
          {fehler}
        </p>
      )}

      <div className="flex flex-wrap items-center gap-2">
        <button
          type="button"
          onClick={speichern}
          disabled={laeuft}
          className="inline-flex min-h-10 items-center rounded-md bg-primary px-3.5 text-sm font-semibold text-on-primary transition-colors hover:bg-primary-hover disabled:opacity-45"
        >
          {laeuft ? "Einen Moment …" : "Übernehmen"}
        </button>
        <button
          type="button"
          onClick={() => setOffen(false)}
          className="inline-flex min-h-10 items-center px-2 text-sm font-semibold text-muted transition-colors hover:text-fg"
        >
          Abbrechen
        </button>
      </div>
    </div>
  );
}
