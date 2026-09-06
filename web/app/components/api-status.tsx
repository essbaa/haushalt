"use client";

import { useEffect, useState } from "react";
import { API_BASE, ApiError, fetchHealth, type Health } from "@/lib/api";

type State =
  | { phase: "laedt" }
  | { phase: "erreichbar"; health: Health }
  | { phase: "gestoert"; health: Health }
  | { phase: "fehler"; message: string; hint?: string };

const TIMEOUT_MS = 12_000;

export function ApiStatus() {
  const [state, setState] = useState<State>({ phase: "laedt" });
  // Erhöhen löst den Effekt erneut aus. Der Zustand wird im Klick-Handler
  // zurückgesetzt und nicht im Effekt — React verbietet zu Recht, im Effekt
  // synchron setState aufzurufen.
  const [versuch, setVersuch] = useState(0);

  useEffect(() => {
    const controller = new AbortController();
    const timer = setTimeout(() => controller.abort(), TIMEOUT_MS);

    fetchHealth(controller.signal)
      .then((health) =>
        setState(
          health.status === "ok"
            ? { phase: "erreichbar", health }
            : { phase: "gestoert", health },
        ),
      )
      .catch((err: unknown) => {
        if (err instanceof DOMException && err.name === "AbortError") return;
        if (err instanceof ApiError) {
          setState({ phase: "fehler", message: err.message, hint: err.hint });
          return;
        }
        setState({
          phase: "fehler",
          message: err instanceof Error ? err.message : "Unbekannter Fehler",
        });
      })
      .finally(() => clearTimeout(timer));

    // Aufräumen bricht die laufende Anfrage ab. Nötig, weil React im
    // Entwicklungsmodus jeden Effekt zweimal ausführt.
    return () => {
      clearTimeout(timer);
      controller.abort();
    };
  }, [versuch]);

  function erneutPruefen() {
    setState({ phase: "laedt" });
    setVersuch((n) => n + 1);
  }

  return (
    <div className="rounded-lg border border-line bg-surface p-5 shadow-sm">
      <div
        className="flex items-center gap-3"
        aria-live="polite"
        aria-busy={state.phase === "laedt"}
      >
        <Dot phase={state.phase} />
        <span className="font-medium">{label(state)}</span>
        {state.phase !== "laedt" && (
          <button
            type="button"
            onClick={erneutPruefen}
            className="ml-auto rounded border border-line px-2.5 py-1 text-xs text-muted transition-colors hover:text-foreground focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-accent"
          >
            Erneut prüfen
          </button>
        )}
      </div>

      <Details state={state} />
    </div>
  );
}

function label(state: State): string {
  switch (state.phase) {
    case "laedt":
      return "Frage den Dienst …";
    case "erreichbar":
      return "API erreichbar";
    case "gestoert":
      return "API antwortet, meldet aber ein Problem";
    case "fehler":
      return "API nicht erreichbar";
  }
}

function Dot({ phase }: { phase: State["phase"] }) {
  const color =
    phase === "erreichbar"
      ? "bg-emerald-600"
      : phase === "gestoert"
        ? "bg-amber-500"
        : phase === "fehler"
          ? "bg-rose-600"
          : "bg-muted animate-pulse motion-reduce:animate-none";
  return <span aria-hidden className={`size-2.5 shrink-0 rounded-full ${color}`} />;
}

function Details({ state }: { state: State }) {
  if (state.phase === "laedt") {
    return (
      <p className="mt-3 font-mono text-xs text-muted">
        {API_BASE || "keine Adresse gesetzt"}
      </p>
    );
  }

  if (state.phase === "fehler") {
    return (
      <div className="mt-3 space-y-1 text-sm">
        <p className="text-rose-700 dark:text-rose-400">{state.message}</p>
        {state.hint && <p className="text-muted">{state.hint}</p>}
      </div>
    );
  }

  const { health } = state;
  return (
    <dl className="mt-4 grid grid-cols-[auto_1fr] gap-x-6 gap-y-1.5 font-mono text-xs">
      <Row k="Adresse" v={API_BASE} />
      <Row k="Version" v={health.version} />
      <Row k="Umgebung" v={health.env} />
      <Row k="Datenbank" v={health.database} />
    </dl>
  );
}

function Row({ k, v }: { k: string; v: string }) {
  return (
    <>
      <dt className="text-muted">{k}</dt>
      <dd className="truncate">{v}</dd>
    </>
  );
}
