"use client";

import { useRouter } from "next/navigation";
import { useState, useTransition } from "react";
import { Personen } from "@/app/components/icons";
import { postMitToken } from "@/lib/browser-token";

/**
 * Was mit einer Aufgabe sonst noch geht: streichen, abschalten, umverteilen.
 *
 * Die beiden häufigen Handgriffe sitzen in der Zeile selbst und sind immer
 * sichtbar — abhaken und abgeben, beide als Zeichen neben dem Text. Sie sind
 * zugleich die Alternative zu den Wischgesten, die WCAG 2.2 verlangt. Was
 * seltener gebraucht wird, steht hier und erscheint erst, wenn jemand die
 * Zeile antippt.
 *
 * Das Formular fürs Abgeben steht trotzdem hier: Es braucht Platz für eine
 * Zeile Text, und der ist neben dem Häkchen nicht. Aufgeklappt wird es von
 * dort — der Knopf oben, das Feld hier.
 *
 * Aus dem Produktkonzept: „Ausführende brauchen Handlungsmacht, nicht nur
 * Pflichten. Eine reine Empfangsliste wird gelöscht." Abgeben ist die kleinste
 * Form davon, und es darf keine Verhandlung sein — ein Tipp, ein Grund,
 * fertig.
 *
 * Der Unterschied zwischen Abgeben und Umverteilen ist der, wer wählt: Beim
 * Abgeben sucht die App jemanden nach ihren Regeln, beim Umverteilen hat ein
 * Mensch schon entschieden. Deshalb widerspricht die App hier nur noch, wenn
 * die Person die Aufgabe nicht übernehmen *kann* — Alter, Rolle, eigene
 * Aufgabe. Rotation und Auslastung sind Annahmen, und Annahmen verlieren gegen
 * jemanden, der es besser weiß.
 */
export function AufgabeAktionen({
  aufgabeId,
  erledigt,
  zustaendig,
  kandidaten,
  verteilbar,
  fragt,
  setFragt,
  streichen,
  abschaltbar,
  abschalten,
}: {
  aufgabeId: string;
  erledigt: boolean;
  /** Wer die Aufgabe gerade hat. Leer, wenn sie offen steht. */
  zustaendig: string;
  /** Alle, die überhaupt Aufgaben übernehmen — betreute Personen stehen
   *  nicht darin. Eine Auswahl anzubieten, die der Dienst sicher ablehnt,
   *  wäre eine Falle. */
  kandidaten: { id: string; name: string }[];
  /** Eigene Aufgaben wandern nicht: „Dein Bett beziehen" bei jemand anderem
   *  ist eine andere Aufgabe, nicht dieselbe in anderen Händen (ADR-0005). */
  verteilbar: boolean;
  /** Ob nach dem Grund fürs Abgeben gefragt wird. Von außen gesteuert: Der
   *  Knopf dafür sitzt in der Zeile selbst, direkt neben dem Häkchen, und
   *  muss beides können — die Zeile aufklappen und gleich das Feld öffnen.
   *  Dasselbe Muster wie „+" bei den eigenen Aufgaben. */
  fragt: boolean;
  setFragt: (f: boolean) => void;
  /** „Diesmal nicht" — die Alternative zur Wischgeste nach links.
   *  WCAG 2.2 verlangt für jede Zieh-Bewegung einen Weg mit einem Zeiger. */
  streichen?: () => void;
  /** „Brauchen wir nicht" — schaltet die Vorlage dauerhaft ab, für alle
   *  künftigen Wochen. Steht nur hier und ist keine Geste mehr. */
  abschaltbar?: boolean;
  abschalten?: () => void;
}) {
  const router = useRouter();
  const [laeuft, setLaeuft] = useState(false);
  const [verteilt, setVerteilt] = useState(false);
  const [grund, setGrund] = useState("");
  const [meldung, setMeldung] = useState<string | null>(null);
  const [fehler, setFehler] = useState<string | null>(null);
  // router.refresh() holt die Server Components neu und ist nicht abgewartet:
  // Ohne useTransition steht der Knopf wieder bereit, während die neuen Daten
  // noch unterwegs sind — die Zeile sieht fertig aus und ändert sich eine
  // Sekunde später doch noch. `uebergang` hält den Zustand, bis der Server
  // geantwortet hat.
  const [uebergang, starten] = useTransition();
  const beschaeftigt = laeuft || uebergang;

  async function tun(arbeit: () => Promise<void>) {
    setLaeuft(true);
    setFehler(null);
    try {
      await arbeit();
      starten(() => router.refresh());
    } catch (e) {
      setFehler(e instanceof Error ? e.message : "Das hat nicht funktioniert.");
    } finally {
      setLaeuft(false);
    }
  }

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

  const zuteilen = (mitglied: string) =>
    tun(async () => {
      const antwort = await postMitToken<{ zustaendig: string }>(
        `/api/aufgaben/${encodeURIComponent(aufgabeId)}/zuteilen`,
        { mitglied },
      );
      setVerteilt(false);
      setMeldung(`${antwort.zustaendig} macht das jetzt.`);
    });

  const andere = kandidaten.filter((m) => m.id !== zustaendig);

  return (
    <div className="space-y-2 pt-1">
      <div className="flex flex-wrap items-center gap-2">
        {streichen && !fragt && (
          <button
            type="button"
            onClick={streichen}
            disabled={beschaeftigt}
            className="inline-flex min-h-9 items-center rounded-full px-3 text-sm font-semibold text-muted transition-colors hover:bg-surface-2 hover:text-fg disabled:opacity-45"
          >
            Diesmal nicht
          </button>
        )}

        {abschaltbar && abschalten && !fragt && (
          <button
            type="button"
            onClick={abschalten}
            disabled={beschaeftigt}
            className="inline-flex min-h-9 items-center rounded-full px-3 text-sm font-semibold text-subtle transition-colors hover:bg-surface-2 hover:text-fg disabled:opacity-45"
          >
            Brauchen wir nicht
          </button>
        )}

        {verteilbar && !erledigt && !fragt && andere.length > 0 && (
          <button
            type="button"
            onClick={() => setVerteilt((v) => !v)}
            disabled={beschaeftigt}
            aria-expanded={verteilt}
            className="inline-flex min-h-9 items-center gap-1.5 rounded-full px-3 text-sm font-semibold text-muted transition-colors hover:bg-surface-2 hover:text-fg disabled:opacity-45"
          >
            <Personen className="size-4" />
            {zustaendig === "" ? "Übernehmen" : "Wer macht das?"}
          </button>
        )}
      </div>

      {verteilt && (
        <div className="space-y-2 rounded-md border border-line bg-surface-2 p-3">
          <p className="text-sm font-semibold">Wer macht das?</p>
          <div className="flex flex-wrap gap-2">
            {andere.map((m) => (
              <button
                key={m.id}
                type="button"
                onClick={() => zuteilen(m.id)}
                disabled={beschaeftigt}
                className="inline-flex min-h-10 items-center rounded-full border border-line-strong bg-surface px-3.5 text-sm font-semibold transition-colors hover:border-primary hover:text-primary disabled:opacity-45"
              >
                {m.name}
              </button>
            ))}
            <button
              type="button"
              onClick={() => setVerteilt(false)}
              className="inline-flex min-h-10 items-center rounded-md px-3 text-sm font-semibold text-muted hover:text-fg"
            >
              Doch nicht
            </button>
          </div>
        </div>
      )}

      {fragt && (
        <div className="space-y-2 rounded-md border border-line bg-surface-2 p-3">
          <label className="block space-y-1.5">
            <span className="block text-sm font-semibold">Warum gibst du sie ab?</span>
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
              disabled={beschaeftigt}
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
