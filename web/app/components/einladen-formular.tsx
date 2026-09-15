"use client";

import { useState } from "react";
import { postMitToken } from "@/lib/browser-token";

type Einladung = { code: string; rolle: string; gueltig_bis: string; fuer?: string };

/** Jemand, der schon im Plan steht, aber noch kein Konto hat. */
export type Offene = { id: string; name: string; rolle: string };

export function EinladenFormular({
  haushaltId,
  offene,
  zugang,
}: {
  haushaltId: string;
  offene: Offene[];
  /** Der Zugangscode der App — reist im Link mit, damit die eingeladene
   *  Person nicht zwei Zeichenketten abtippen muss. Leer, wenn keiner
   *  eingetragen ist; dann kommt sie gar nicht durch die Registrierung, und
   *  der Hinweis unten sagt das. */
  zugang: string;
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
    // Zwei Codes in einem Link: der Einladungscode für den Haushalt, der
    // Zugangscode für die Registrierung überhaupt. Sie getrennt zu verschicken
    // wäre genauer und in der Praxis eine Nachricht, in der die Hälfte
    // verlorengeht.
    const ziel = `/beitreten?code=${encodeURIComponent(einladung.code)}`;
    const link =
      typeof window === "undefined"
        ? ""
        : zugang === ""
          ? `${window.location.origin}${ziel}`
          : `${window.location.origin}/anmelden?weiter=${encodeURIComponent(ziel)}&zugang=${encodeURIComponent(zugang)}`;
    return (
      <div className="space-y-5 rounded-lg border border-line bg-surface p-6">
        <div className="rounded-md bg-primary-soft px-4 py-5 text-center">
          <p className="text-3xl font-extrabold tracking-[0.25em] text-primary">
            {einladung.code}
          </p>
        </div>
        <div className="space-y-1">
          <p className="text-sm font-semibold">Schick diesen Link</p>
          <p className="break-all text-xs text-muted">{link}</p>
          {zugang === "" ? (
            <p className="text-xs leading-relaxed text-clay text-pretty">
              Achtung: Es ist kein Zugangscode eingetragen (`ZUGANGSCODES`).
              Ohne ihn kann sich niemand registrieren — der Link führt dann ins
              Leere.
            </p>
          ) : (
            <p className="text-xs leading-relaxed text-muted text-pretty">
              Er bringt beides mit: den Zugang zur App und die Einladung in
              euren Haushalt.
            </p>
          )}
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
          className="inline-flex min-h-11 items-center text-sm font-semibold text-primary underline underline-offset-2"
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
          <legend className="mb-2 text-sm font-semibold">Wen lädst du ein?</legend>
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
        <legend className="mb-2 text-sm font-semibold">Welche Rolle?</legend>
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
        <p role="alert" className="rounded-md border border-danger/40 px-3 py-2 text-sm text-danger">
          {fehler}
        </p>
      )}

      <button
        type="button"
        onClick={erzeugen}
        disabled={laeuft}
        className="inline-flex min-h-12 w-full items-center justify-center rounded-md bg-primary px-4 text-sm font-semibold text-on-primary transition-colors hover:bg-primary-hover disabled:opacity-45"
      >
        {laeuft ? "Einen Moment …" : "Code erzeugen"}
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
        gewaehlt
          ? "border-primary bg-primary-soft"
          : "border-line hover:border-line-strong"
      }`}
    >
      <span className="flex items-baseline gap-2">
        <input
          type="radio"
          name={name}
          checked={gewaehlt}
          onChange={waehlen}
          className="accent-[var(--primary)]"
        />
        <span className="font-semibold">{titel}</span>
      </span>
      <span className="mt-1 block pl-6 text-sm leading-relaxed text-muted">
        {erklaerung}
      </span>
    </label>
  );
}
