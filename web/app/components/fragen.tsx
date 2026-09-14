"use client";

import { useRouter } from "next/navigation";
import { useId, useState, useTransition } from "react";
import type { Frage } from "@/lib/api";
import { patchMitToken, postMitToken } from "@/lib/browser-token";

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
 *
 * Nach der Antwort bleibt die Karte stehen und bestätigt sie. Der Grund ist
 * nicht Höflichkeit: Eine festgeschriebene Woche ändert sich von einer
 * Antwort nicht (ADR-0008), die neue Aufgabe kommt also erst nächste Woche.
 * Ohne diesen Satz sah es aus, als wäre die Antwort verlorengegangen — die
 * Frage wurde gespeichert, die Seite lud neu und zeigte dieselbe Frage.
 *
 * Wer nicht warten will, bekommt den Knopf daneben. Die Woche selbst
 * umzusortieren, ohne dass jemand danach gefragt hat, wäre der andere Fehler.
 *
 * Zugeklappt, weil der Abschnitt über der Woche steht: Zwei Karten sind auf
 * einem Telefon fast der ganze Bildschirm, und der Plan ist der Grund, warum
 * jemand die App öffnet. Eine Zeile mit Anzahl und Einsatz ist immer noch
 * ungleich sichtbarer als zwei Karten hinter zwei Dutzend Aufgaben — und sie
 * kostet die Woche nichts.
 */
export function Fragen({
  haushaltId,
  woche,
  fragen,
}: {
  haushaltId: string;
  /** Die angezeigte Woche — gebraucht für „Schon diese Woche". */
  woche: string;
  fragen: Frage[];
}) {
  const router = useRouter();
  const feld = useId();
  const [auf, setAuf] = useState(false);
  const [laeuft, setLaeuft] = useState<string | null>(null);
  const [beantwortet, setBeantwortet] = useState<Record<string, boolean>>({});
  const [gerechnet, setGerechnet] = useState(false);
  const [fehler, setFehler] = useState<string | null>(null);
  // router.refresh() holt die Server Components neu und ist nicht abgewartet:
  // Ohne useTransition steht der Knopf wieder bereit, während die neuen Daten
  // noch unterwegs sind — die Zeile sieht fertig aus und ändert sich eine
  // Sekunde später doch noch. `uebergang` hält den Zustand, bis der Server
  // geantwortet hat.
  const [uebergang, starten] = useTransition();

  async function antworten(faktum: string, wert: boolean) {
    setLaeuft(faktum);
    setFehler(null);
    try {
      await patchMitToken(`/api/haushalte/${encodeURIComponent(haushaltId)}/fakten`, {
        [faktum]: wert,
      });
      setBeantwortet((b) => ({ ...b, [faktum]: wert }));
      // Eine neue Antwort hebt die letzte Neurechnung auf: Die Woche kennt
      // sie noch nicht, der Knopf gehört also wieder her.
      setGerechnet(false);
    } catch (e) {
      setFehler(e instanceof Error ? e.message : "Das hat nicht funktioniert.");
    } finally {
      setLaeuft(null);
    }
  }

  async function neuRechnen() {
    setLaeuft("woche");
    setFehler(null);
    try {
      await postMitToken(
        `/api/haushalte/${encodeURIComponent(haushaltId)}/plan/${encodeURIComponent(woche)}/neu`,
      );
      // Nach dem Neurechnen kommt eine frische Liste — die alten Antworten
      // sind darin nicht mehr offen. Bleiben sie stehen, rendert die neue
      // Frage als Bestätigung der alten.
      setBeantwortet({});
      setGerechnet(true);
      starten(() => router.refresh());
    } catch (e) {
      setFehler(e instanceof Error ? e.message : "Das hat nicht funktioniert.");
    } finally {
      setLaeuft(null);
    }
  }

  const offeneAntworten = Object.values(beantwortet).some((w) => w);
  const offeneFragen = fragen.filter((f) => beantwortet[f.faktum] === undefined).length;

  return (
    <section className="rounded-lg border border-line bg-surface px-4 py-1">
      <h2>
        <button
          type="button"
          onClick={() => setAuf((a) => !a)}
          aria-expanded={auf}
          aria-controls={feld}
          className="flex min-h-12 w-full cursor-pointer items-center gap-3 text-left"
        >
          <span className="font-bold tracking-tight">
            {offeneFragen === 0
              ? "Beantwortet"
              : offeneFragen === 1
                ? "Eine Frage offen"
                : `${offeneFragen} Fragen offen`}
          </span>
          <span className="min-w-0 flex-1 truncate text-sm text-muted">
            Was die App nicht weiß, plant sie nicht ein.
          </span>
          <span
            aria-hidden="true"
            className={`text-muted transition-transform ${auf ? "rotate-90" : ""}`}
          >
            ›
          </span>
        </button>
      </h2>

      <div id={feld} hidden={!auf} className="space-y-3 pb-4">
        <p className="text-sm leading-relaxed text-muted">Jede Antwort gilt dauerhaft.</p>

        <ul className="space-y-3">
          {fragen.map((f) => {
            const antwort = beantwortet[f.faktum];

            if (antwort !== undefined) {
              return (
                <li key={f.faktum} className="rounded-lg border border-line bg-surface-2 p-4">
                  <p className="text-sm leading-relaxed text-pretty">
                    <span className="font-semibold">Notiert.</span>{" "}
                    {antwort
                      ? `${f.dann} — ab nächster Woche.`
                      : "Danach fragen wir nicht mehr."}
                  </p>
                </li>
              );
            }

            return (
              <li key={f.faktum} className="rounded-lg border border-line bg-surface p-4">
                <p className="font-semibold text-pretty">{f.frage}</p>
                <p className="mt-1 text-sm leading-relaxed text-muted text-pretty">{f.dann}</p>
                <div className="mt-3 flex flex-wrap gap-2">
                  <button
                    type="button"
                    onClick={() => antworten(f.faktum, true)}
                    disabled={laeuft !== null || uebergang}
                    className="inline-flex min-h-11 items-center rounded-md bg-primary px-5 text-sm font-semibold text-on-primary transition-colors hover:bg-primary-hover disabled:opacity-45"
                  >
                    {laeuft === f.faktum ? "Einen Moment …" : "Ja"}
                  </button>
                  <button
                    type="button"
                    onClick={() => antworten(f.faktum, false)}
                    disabled={laeuft !== null || uebergang}
                    className="inline-flex min-h-11 items-center rounded-md border border-line-strong px-5 text-sm font-semibold text-muted transition-colors hover:border-line-strong hover:text-fg disabled:opacity-45"
                  >
                    Nein
                  </button>
                </div>
              </li>
            );
          })}
        </ul>

        {offeneAntworten && !gerechnet && (
          <div className="flex flex-wrap items-center gap-x-3 gap-y-2">
            <button
              type="button"
              onClick={neuRechnen}
              disabled={laeuft !== null || uebergang}
              className="inline-flex min-h-11 items-center rounded-md border border-line-strong px-4 text-sm font-semibold transition-colors hover:border-primary hover:text-primary disabled:opacity-45"
            >
              {laeuft === "woche" ? "Einen Moment …" : "Schon diese Woche"}
            </button>
            <p className="text-sm text-muted text-pretty">
              Rechnet den Rest der Woche neu. Abgehaktes und Abgegebenes bleibt stehen.
            </p>
          </div>
        )}

        {gerechnet && (
          <p className="text-sm text-primary">Die Woche ist neu gerechnet.</p>
        )}

        {fehler && (
          <p role="alert" className="text-sm text-danger">
            {fehler}
          </p>
        )}
      </div>
    </section>
  );
}
