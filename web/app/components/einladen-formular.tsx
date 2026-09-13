"use client";

import { useState } from "react";
import { postMitToken } from "@/lib/browser-token";

type Einladung = { code: string; rolle: string; gueltig_bis: string; fuer?: string };

/** Jemand, der schon im Plan steht, aber noch kein Konto hat. */
export type Offene = { id: string; name: string; rolle: string };

export function EinladenFormular({
  haushaltId,
  offene,
}: {
  haushaltId: string;
  offene: Offene[];
}) {
  // Vorbelegt mit der ersten Person, die schon im Plan steht: Das ist der
  // häufige Fall. Wer jemand Neues holen will, wählt "neu".
  const [wen, setWen] = useState<string>(offene[0]?.id ?? "neu");
  const [rolle, setRolle] = useState("ausfuehrend");
  const [einladung, setEinladung] = useState<Einladung | null>(null);
  const [fehler, setFehler] = useState<string | null>(null);
  const [laeuft, setLaeuft] = useState(false);

  async function erzeugen() {
    setFehler(null);
    setLaeuft(true);
    try {
      setEinladung(
        await postMitToken<Einladung>(
          `/api/haushalte/${encodeURIComponent(haushaltId)}/einladungen`,
          // Zeigt die Einladung auf jemanden, der schon im Plan steht, kommt
          // die Rolle von dieser Person. Sie mitzuschicken wäre eine zweite
          // Wahrheit — der Server ignoriert sie ohnehin.
          wen === "neu" ? { rolle } : { mitglied: wen },
        ),
      );
    } catch (e) {
      setFehler(e instanceof Error ? e.message : "Das hat nicht funktioniert.");
    } finally {
      setLaeuft(false);
    }
  }

  if (einladung) {
    const link =
      typeof window === "undefined"
        ? ""
        : `${window.location.origin}/beitreten?code=${einladung.code}`;
    return (
      <div className="space-y-4 rounded-lg border border-line bg-surface p-6">
        <div>
          <p className="mb-1 text-sm text-muted">Der Code</p>
          <p className="font-mono text-2xl tracking-[0.2em]">{einladung.code}</p>
        </div>
        <div>
          <p className="mb-1 text-sm text-muted">Oder dieser Link</p>
          <p className="break-all font-mono text-xs">{link}</p>
        </div>
        <p className="text-sm leading-relaxed text-muted">
          {einladung.fuer
            ? `Macht den Empfänger zu ${einladung.fuer}. `
            : "Holt eine neue Person in den Haushalt. "}
          Gilt einmal, bis{" "}
          {new Date(einladung.gueltig_bis).toLocaleDateString("de-DE", {
            day: "2-digit",
            month: "2-digit",
            year: "numeric",
          })}
          . Rolle: {einladung.rolle === "planend" ? "planend" : "ausführend"}.
        </p>
        <button
          type="button"
          onClick={() => setEinladung(null)}
          className="text-sm text-accent underline"
        >
          Noch jemanden einladen
        </button>
      </div>
    );
  }

  return (
    <div className="space-y-6">
      {offene.length > 0 && (
        <fieldset className="space-y-3">
          <legend className="mb-2 text-sm font-medium">Wen lädst du ein?</legend>
          {offene.map((o) => (
            <Wahl
              key={o.id}
              name="wen"
              gewaehlt={wen === o.id}
              waehlen={() => setWen(o.id)}
              titel={o.name}
              erklaerung={`Steht schon im Plan und ${
                o.rolle === "planend" ? "plant mit" : "führt aus"
              }. Der Code verbindet ${o.name} mit einem Konto — Aufgaben, Verlauf und Kapazität bleiben dieselben.`}
            />
          ))}
          <Wahl
            name="wen"
            gewaehlt={wen === "neu"}
            waehlen={() => setWen("neu")}
            titel="Jemand Neues"
            erklaerung="Kommt als zusätzliche Person in den Haushalt und taucht ab dann im Plan auf."
          />
        </fieldset>
      )}

      <fieldset className="space-y-3" hidden={wen !== "neu"}>
        <legend className="mb-2 text-sm font-medium">Welche Rolle?</legend>
        <Wahl
          gewaehlt={rolle === "ausfuehrend"}
          waehlen={() => setRolle("ausfuehrend")}
          titel="Ausführend"
          erklaerung="Übernimmt Aufgaben und sieht den ganzen Wochenplan — aber nicht die Auswertung, wer im Haushalt wie viel trägt. Für Kinder und Jugendliche."
        />
        <Wahl
          gewaehlt={rolle === "planend"}
          waehlen={() => setRolle("planend")}
          titel="Planend"
          erklaerung="Sieht alles, lädt selbst ein und trägt die Kopfarbeit mit. Für die Erwachsenen im Haushalt."
        />
      </fieldset>

      {fehler && (
        <p role="alert" className="text-sm text-clay">
          {fehler}
        </p>
      )}

      <button
        type="button"
        onClick={erzeugen}
        disabled={laeuft}
        className="w-full rounded-md bg-accent px-4 py-2.5 font-medium text-background disabled:opacity-60"
      >
        {laeuft ? "…" : "Code erzeugen"}
      </button>
    </div>
  );
}

function Wahl({
  gewaehlt,
  waehlen,
  titel,
  erklaerung,
  name = "rolle",
}: {
  gewaehlt: boolean;
  waehlen: () => void;
  titel: string;
  erklaerung: string;
  name?: string;
}) {
  return (
    <label
      className={`block cursor-pointer rounded-lg border p-4 transition-colors ${
        gewaehlt ? "border-accent bg-accent/5" : "border-line"
      }`}
    >
      <span className="flex items-baseline gap-2">
        <input
          type="radio"
          name={name}
          checked={gewaehlt}
          onChange={waehlen}
          className="accent-accent"
        />
        <span className="font-medium">{titel}</span>
      </span>
      <span className="mt-1 block pl-6 text-sm leading-relaxed text-muted">
        {erklaerung}
      </span>
    </label>
  );
}
