import Link from "next/link";
import type { ComponentProps, ReactNode } from "react";

/**
 * Die Bausteine der Oberfläche.
 *
 * Vorher trug jede Ansicht ihre eigenen Tailwind-Ketten — dieselbe Schaltfläche
 * an vier Stellen mit vier Abstufungen. Hier steht jede Form einmal; wer sie
 * ändert, ändert sie überall.
 *
 * Alle anfassbaren Flächen sind mindestens 44 Pixel hoch. Das ist keine
 * Vorliebe, sondern die Größe einer Fingerkuppe — darunter trifft man auf dem
 * Handy die Nachbarzeile.
 */

const knopf = {
  basis:
    "inline-flex min-h-11 items-center justify-center gap-2 rounded-md px-4 text-sm font-semibold transition-colors duration-150 disabled:cursor-not-allowed disabled:opacity-45",
  haupt: "bg-primary text-on-primary hover:bg-primary-hover",
  zweit: "border border-line-strong bg-surface text-fg hover:border-primary hover:text-primary",
  leise: "text-muted hover:bg-surface-2 hover:text-fg",
  gefahr: "border border-line-strong bg-surface text-danger hover:border-danger",
};

type KnopfArt = keyof Omit<typeof knopf, "basis">;

export function Knopf({
  art = "zweit",
  className = "",
  ...rest
}: ComponentProps<"button"> & { art?: KnopfArt }) {
  return <button type="button" className={`${knopf.basis} ${knopf[art]} ${className}`} {...rest} />;
}

export function KnopfLink({
  art = "zweit",
  className = "",
  ...rest
}: ComponentProps<typeof Link> & { art?: KnopfArt }) {
  return <Link className={`${knopf.basis} ${knopf[art]} ${className}`} {...rest} />;
}

/** Eine Fläche, die etwas zusammenhält. Rahmen statt Schatten — flach bleibt flach. */
export function Karte({ className = "", ...rest }: ComponentProps<"div">) {
  return <div className={`rounded-lg border border-line bg-surface ${className}`} {...rest} />;
}

/** Überschrift einer Seite, samt Weg zurück. */
export function Kopfzeile({
  titel,
  zurueck,
  rechts,
  children,
}: {
  titel: string;
  zurueck?: { href: string; text: string };
  rechts?: ReactNode;
  children?: ReactNode;
}) {
  return (
    <header className="space-y-3">
      {zurueck && (
        <Link
          href={zurueck.href}
          className="inline-flex min-h-11 items-center gap-1.5 text-sm text-muted transition-colors hover:text-fg"
        >
          <span aria-hidden="true">←</span>
          {zurueck.text}
        </Link>
      )}
      <div className="flex flex-wrap items-start justify-between gap-3">
        <h1 className="text-2xl font-extrabold tracking-tight text-balance sm:text-3xl">{titel}</h1>
        {rechts}
      </div>
      {children && (
        <p className="max-w-prose text-sm leading-relaxed text-muted text-pretty">{children}</p>
      )}
    </header>
  );
}

/**
 * Eine Reihe, aus der genau eines gewählt ist.
 *
 * Als echte Radiogruppe: Ein Kranz gleich aussehender <button> ist für
 * Vorlesesoftware eine Ansammlung von Knöpfen, nicht eine Wahl mit einer
 * Antwort.
 */
export function Auswahl<T extends string>({
  name,
  wert,
  optionen,
  auf,
  aus,
}: {
  name: string;
  wert: T;
  optionen: { wert: T; text: string }[];
  auf: (w: T) => void;
  aus?: boolean;
}) {
  return (
    <div
      role="radiogroup"
      aria-label={name}
      className="inline-flex flex-wrap gap-1.5 rounded-lg bg-surface-2 p-1"
    >
      {optionen.map((o) => {
        const gewaehlt = o.wert === wert;
        return (
          <button
            key={o.wert}
            type="button"
            role="radio"
            aria-checked={gewaehlt}
            disabled={aus}
            onClick={() => auf(o.wert)}
            className={`min-h-10 rounded-md px-3.5 text-sm font-semibold transition-colors duration-150 disabled:opacity-45 ${
              gewaehlt
                ? "bg-surface text-primary shadow-none ring-1 ring-primary/30"
                : "text-muted hover:text-fg"
            }`}
          >
            {o.text}
          </button>
        );
      })}
    </div>
  );
}

/** Ein Ja/Nein-Schalter in Form einer Marke, wie man sie antippt. */
export function Marke({
  an,
  text,
  aus,
  klick,
}: {
  an: boolean;
  text: string;
  aus?: boolean;
  klick: () => void;
}) {
  return (
    <button
      type="button"
      onClick={klick}
      disabled={aus}
      aria-pressed={an}
      className={`inline-flex min-h-11 items-center gap-2 rounded-full border px-4 text-sm font-semibold transition-colors duration-150 disabled:opacity-45 ${
        an
          ? "border-primary bg-primary-soft text-primary"
          : "border-line-strong text-muted hover:border-line-strong hover:text-fg"
      }`}
    >
      <span
        aria-hidden="true"
        className={`size-1.5 rounded-full ${an ? "bg-primary" : "bg-line-strong"}`}
      />
      {text}
    </button>
  );
}

/** Ein Feld mit sichtbarer Beschriftung. Platzhalter sind keine Beschriftung:
 *  Sobald jemand tippt, ist die Frage weg. */
export function Feld({
  beschriftung,
  hinweis,
  className = "",
  ...rest
}: ComponentProps<"input"> & { beschriftung: string; hinweis?: string }) {
  return (
    <label className="block space-y-1.5">
      <span className="block text-sm font-semibold">{beschriftung}</span>
      <input
        className={`block min-h-11 w-full rounded-md border border-line-strong bg-surface px-3 text-base text-fg transition-colors placeholder:text-subtle focus:border-primary ${className}`}
        {...rest}
      />
      {hinweis && <span className="block text-xs text-muted">{hinweis}</span>}
    </label>
  );
}

/** Kurze Zusatzinformation an einer Zeile — Dauer, Kopflast, Status. */
export function Merkmal({
  children,
  ton = "still",
}: {
  children: ReactNode;
  ton?: "still" | "warm";
}) {
  return (
    <span
      className={`inline-flex items-center gap-1 rounded-full px-2 py-0.5 text-xs font-semibold ${
        ton === "warm" ? "bg-clay-soft text-clay" : "bg-surface-2 text-muted"
      }`}
    >
      {children}
    </span>
  );
}

/** Die Anfangsbuchstaben einer Person.
 *
 *  Den Ton bringt die Zeile mit: `--person` und `--person-soft` stehen an
 *  einem Elternelement (siehe lib/personen.ts). Ohne eins fällt es auf die
 *  gedeckten Vorgaben zurück und sieht aus wie bisher.
 *
 *  `eigen` füllt die Fläche statt sie zu tönen. Zwei Fragen, zwei Mittel:
 *  Die Farbe sagt WER, die Füllung sagt DU. Farbe allein trägt nie eine
 *  Aussage — der Name steht daneben. */
export function Zeichen({ name, eigen = false }: { name: string; eigen?: boolean }) {
  const kurz = name
    .split(/\s+/)
    .slice(0, 2)
    .map((t) => t[0] ?? "")
    .join("")
    .toUpperCase();
  return (
    <span
      aria-hidden="true"
      className={`inline-flex size-8 shrink-0 items-center justify-center rounded-full text-xs font-bold ${
        eigen
          ? "bg-[var(--person)] text-[var(--person-on)]"
          : "bg-[var(--person-soft)] text-[var(--person)]"
      }`}
    >
      {kurz || "?"}
    </span>
  );
}

/** Ein Hinweis, der etwas erklärt oder warnt. */
export function Hinweis({
  ton = "still",
  children,
}: {
  ton?: "still" | "warm" | "fehler";
  children: ReactNode;
}) {
  const farben = {
    still: "border-line bg-surface-2 text-muted",
    warm: "border-clay/30 bg-clay-soft text-clay",
    fehler: "border-danger/40 bg-surface text-danger",
  };
  return (
    <p
      role={ton === "fehler" ? "alert" : undefined}
      className={`rounded-md border px-4 py-3 text-sm leading-relaxed ${farben[ton]}`}
    >
      {children}
    </p>
  );
}
