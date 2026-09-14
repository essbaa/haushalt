"use client";

import { useState, type ComponentProps } from "react";

import { Auge, AugeZu } from "@/app/components/icons";

/**
 * Ein Passwortfeld mit Auge zum Einblenden.
 *
 * Statt eines zweiten Feldes „Passwort wiederholen". Der Grund ist nicht
 * Kürze: Menschen machen denselben Tippfehler zweimal, und dann fängt ein
 * zweites Feld gar nichts — es bestätigt den Fehler nur. Sehen zu können, was
 * man getippt hat, fängt ihn wirklich.
 *
 * Auf dem Telefon kommt dazu, dass jedes Feld weniger ein Feld weniger über
 * der eingeblendeten Tastatur ist.
 *
 * Der Knopf ist kein Zustand des Feldes, sondern ein Schalter mit
 * `aria-pressed` — eine Bildschirmleseausgabe soll sagen können, ob das
 * Passwort gerade sichtbar ist.
 */
export function Passwortfeld({
  beschriftung,
  hinweis,
  ...rest
}: ComponentProps<"input"> & { beschriftung: string; hinweis?: string }) {
  const [sichtbar, setSichtbar] = useState(false);

  return (
    <label className="block space-y-1.5">
      <span className="block text-sm font-semibold">{beschriftung}</span>
      <span className="relative block">
        <input
          {...rest}
          type={sichtbar ? "text" : "password"}
          className="block min-h-11 w-full rounded-md border border-line-strong bg-surface py-2.5 pl-3 pr-12 text-base text-fg transition-colors placeholder:text-subtle focus:border-primary"
        />
        <button
          type="button"
          onClick={() => setSichtbar((v) => !v)}
          aria-pressed={sichtbar}
          aria-label={sichtbar ? "Passwort verbergen" : "Passwort anzeigen"}
          // Kein tabIndex={-1}: Wer mit der Tastatur arbeitet, braucht den
          // Schalter genauso — gerade dann.
          className="absolute inset-y-0 right-0 flex w-12 items-center justify-center rounded-r-md text-muted transition-colors hover:text-fg"
        >
          {sichtbar ? <AugeZu className="size-5" /> : <Auge className="size-5" />}
        </button>
      </span>
      {hinweis && <span className="block text-xs text-muted">{hinweis}</span>}
    </label>
  );
}
