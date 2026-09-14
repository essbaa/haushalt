"use client";

import { useRef, useState, type PointerEvent, type ReactNode } from "react";

/**
 * Eine Zeile, die sich zur Seite schieben lässt.
 *
 * Nach links kommt die stille Aktion, nach rechts die freundliche — wie man es
 * vom Telefon kennt. Drei Dinge machen daraus etwas, das man benutzen kann,
 * ohne sich zu ärgern:
 *
 *  1. `touch-action: pan-y` — senkrechtes Scrollen bleibt Sache des Browsers.
 *     Ohne das kämpft jede Wischgeste gegen die Liste.
 *  2. Die Geste startet erst, wenn sie deutlich waagerecht ist, und nie am
 *     linken Rand: Dort gehört sie dem Browser (Zurück-Geste).
 *  3. Ausgelöst wird erst über einer Schwelle. Darunter federt die Zeile
 *     zurück — ein Wisch, der versehentlich löscht, wird kein zweites Mal
 *     gewagt.
 *
 * Die Geste ist eine Abkürzung, nie der einzige Weg: WCAG 2.2 verlangt für
 * jede Zieh-Bewegung eine Alternative mit einem Zeiger, und die Knöpfe in der
 * Zeile sind sie.
 */
export function Wischen({
  links,
  rechts,
  children,
}: {
  links?: { text: string; ton?: "still" | "warm"; tun: () => void };
  rechts?: { text: string; tun: () => void };
  children: ReactNode;
}) {
  const [dx, setDx] = useState(0);
  const [zieht, setZieht] = useState(false);
  const start = useRef<{ x: number; y: number } | null>(null);
  const achse = useRef<"offen" | "waagerecht" | "senkrecht">("offen");

  const schwelle = 96;

  function runter(e: PointerEvent<HTMLDivElement>) {
    // Der linke Rand gehört der Zurück-Geste des Browsers.
    if (e.clientX < 24) return;
    start.current = { x: e.clientX, y: e.clientY };
    achse.current = "offen";
  }

  function bewegt(e: PointerEvent<HTMLDivElement>) {
    if (!start.current) return;
    const dX = e.clientX - start.current.x;
    const dY = e.clientY - start.current.y;

    if (achse.current === "offen") {
      if (Math.abs(dX) < 10 && Math.abs(dY) < 10) return;
      achse.current = Math.abs(dX) > Math.abs(dY) ? "waagerecht" : "senkrecht";
      if (achse.current === "waagerecht") {
        setZieht(true);
        e.currentTarget.setPointerCapture(e.pointerId);
      }
    }
    if (achse.current !== "waagerecht") return;

    // Nur in die Richtungen, für die es eine Aktion gibt.
    if ((dX < 0 && !links) || (dX > 0 && !rechts)) return;
    setDx(dX);
  }

  function hoch() {
    if (achse.current === "waagerecht") {
      if (dx <= -schwelle && links) links.tun();
      else if (dx >= schwelle && rechts) rechts.tun();
    }
    start.current = null;
    achse.current = "offen";
    setZieht(false);
    setDx(0);
  }

  const weit = Math.abs(dx) >= schwelle;
  const nachLinks = dx < 0;
  const aktion = nachLinks ? links : rechts;

  return (
    <div className="relative overflow-hidden rounded-lg">
      {/* Was unter der Zeile zum Vorschein kommt. Die Beschriftung sagt, was
          beim Loslassen passiert — eine farbige Fläche allein sagt es nicht. */}
      {dx !== 0 && aktion && (
        <div
          aria-hidden="true"
          className={`absolute inset-0 flex items-center px-4 text-sm font-semibold ${
            nachLinks ? "justify-end" : "justify-start"
          } ${
            nachLinks
              ? weit
                ? "bg-clay-soft text-clay"
                : "bg-surface-2 text-muted"
              : weit
                ? "bg-primary text-on-primary"
                : "bg-surface-2 text-muted"
          }`}
        >
          {aktion.text}
        </div>
      )}

      <div
        onPointerDown={runter}
        onPointerMove={bewegt}
        onPointerUp={hoch}
        onPointerCancel={hoch}
        style={{
          transform: `translateX(${dx}px)`,
          touchAction: "pan-y",
          transition: zieht ? "none" : "transform 180ms ease-out",
        }}
        className="relative bg-bg"
      >
        {children}
      </div>
    </div>
  );
}
