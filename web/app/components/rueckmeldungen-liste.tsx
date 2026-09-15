import type { Rueckmeldung } from "@/lib/api";

const arten: Record<string, { wort: string; klassen: string }> = {
  fehler: { wort: "Stimmt nicht", klassen: "bg-clay-soft text-clay" },
  stoert: { wort: "Stört", klassen: "bg-surface-2 text-muted" },
  idee: { wort: "Idee", klassen: "bg-primary-soft text-primary" },
};

/**
 * Was der Haushalt gemeldet hat.
 *
 * Dass die Liste überhaupt sichtbar ist, gehört zur Sache: Ein Formular, das
 * schluckt und nie etwas zeigt, ist schlechter als der Zettel am Kühlschrank
 * — beim Zettel sieht man wenigstens, dass er voll wird. Wer einmal gemeldet
 * hat und nichts passieren sah, meldet kein zweites Mal.
 *
 * Nur für die planenden Personen: In den Sätzen steht Kritik an anderen
 * Menschen im selben Haushalt, und die gehört nicht an jede Pinnwand.
 *
 * Server Component — nur Text.
 */
export function RueckmeldungenListe({ liste }: { liste: Rueckmeldung[] }) {
  if (liste.length === 0) {
    return (
      <p className="text-sm leading-relaxed text-muted text-pretty">
        Noch nichts gemeldet. Unten auf jeder Seite steht &bdquo;Stimmt etwas
        nicht?&ldquo;
        — das darf jeder im Haushalt benutzen, nicht nur du.
      </p>
    );
  }

  return (
    <ul className="space-y-3">
      {liste.map((r) => {
        const art = arten[r.art] ?? { wort: r.art, klassen: "bg-surface-2 text-muted" };
        return (
          <li key={r.id} className="rounded-lg border border-line bg-surface px-4 py-3">
            <div className="flex flex-wrap items-center gap-x-2 gap-y-1">
              <span
                className={`inline-flex items-center rounded-full px-2.5 py-0.5 text-xs font-semibold ${art.klassen}`}
              >
                {art.wort}
              </span>
              {r.wer && <span className="text-sm font-semibold">{r.wer}</span>}
              <span className="text-xs text-subtle tabular-nums">
                {new Date(r.gemeldet_am).toLocaleDateString("de-DE", {
                  day: "2-digit",
                  month: "2-digit",
                  hour: "2-digit",
                  minute: "2-digit",
                })}
              </span>
            </div>
            {/* Der Wortlaut bleibt, wie er getippt wurde — Zeilenumbrüche
                inklusive. Eine Rückmeldung umzuformulieren heißt, sie zu
                verlieren. */}
            <p className="mt-1.5 text-sm leading-relaxed whitespace-pre-wrap text-pretty">
              {r.text}
            </p>
            {r.kontext && <p className="mt-1 text-xs text-subtle">{r.kontext}</p>}
          </li>
        );
      })}
    </ul>
  );
}
