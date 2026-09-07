import type { Aufgabe, Bilanz, Wochenplan as Plan } from "@/lib/api";
import { tagLesbar, wochenSpanne } from "@/lib/woche";

/**
 * Der Wochenplan als Ansicht. Server Component, kein JavaScript im Browser.
 *
 * Bewusst kein Rot und kein Grün als einziger Träger einer Aussage: Kopflast
 * und Auslastung stehen als Zahl da, die Balken sind Beiwerk.
 */
export function Wochenplan({ plan }: { plan: Plan }) {
  const namen = new Map(plan.haushalt.mitglieder.map((m) => [m.id, m.name]));
  const nachTag = new Map<string, Aufgabe[]>();
  for (const a of plan.aufgaben) {
    nachTag.set(a.tag, [...(nachTag.get(a.tag) ?? []), a]);
  }
  const tage = [...nachTag.keys()].sort();

  return (
    <div className="space-y-10">
      <div className="flex flex-wrap items-baseline justify-between gap-2 border-b border-line pb-3">
        <h2 className="text-xl font-semibold">{plan.haushalt.name}</h2>
        <p className="font-mono text-xs uppercase tracking-widest text-muted">
          {plan.woche} · {wochenSpanne(plan.woche)}
        </p>
      </div>

      {tage.length === 0 ? (
        <p className="text-muted">Diese Woche steht nichts an.</p>
      ) : (
        <div className="space-y-8">
          {tage.map((tag) => (
            <section key={tag}>
              <h3 className="mb-2 text-sm font-medium">{tagLesbar(tag)}</h3>
              <ul className="divide-y divide-line rounded-lg border border-line bg-surface">
                {nachTag.get(tag)!.map((a, i) => (
                  <AufgabeZeile key={`${a.vorlage_id}-${i}`} aufgabe={a} namen={namen} />
                ))}
              </ul>
            </section>
          ))}
        </div>
      )}

      <BilanzTafel bilanz={plan.bilanz} namen={namen} />

      {plan.uebersprungen.length > 0 && (
        <details className="rounded-lg border border-line bg-surface p-4">
          <summary className="cursor-pointer text-sm font-medium">
            Nicht im Plan ({plan.uebersprungen.length})
          </summary>
          <p className="mt-2 text-sm text-muted">
            Jede Vorlage, die nicht eingeplant wurde, kommt mit einem Grund
            zurück. &bdquo;Warum steht das nicht in meinem Plan?&ldquo; ist die
            erste Frage, die ein Haushalt stellt.
          </p>
          <ul className="mt-3 space-y-1 text-sm">
            {plan.uebersprungen.map((u) => (
              <li key={u.vorlage_id} className="flex justify-between gap-4">
                <span>{u.titel}</span>
                <span className="shrink-0 font-mono text-xs text-muted">
                  {grundText[u.grund] ?? u.grund}
                </span>
              </li>
            ))}
          </ul>
        </details>
      )}
    </div>
  );
}

function AufgabeZeile({
  aufgabe: a,
  namen,
}: {
  aufgabe: Aufgabe;
  namen: Map<string, string>;
}) {
  return (
    <li className="flex flex-wrap items-baseline gap-x-3 gap-y-1 px-4 py-3">
      {a.art === "organisation" && (
        <span
          title="Organisationsaufgabe — Kopfarbeit"
          className="font-mono text-xs text-accent"
        >
          ○
        </span>
      )}
      <span className="min-w-32 font-medium">{namen.get(a.zustaendig) ?? a.zustaendig}</span>
      <span className="flex-1">{a.titel}</span>
      <span className="font-mono text-xs text-muted">
        {a.dauer_min} min
        {a.kopflast > 0 && ` · Kopflast ${a.kopflast}`}
      </span>
      <span className="w-full font-mono text-xs text-muted sm:w-auto sm:min-w-40 sm:text-right">
        {begruendungText(a, namen)}
      </span>
    </li>
  );
}

function BilanzTafel({
  bilanz,
  namen,
}: {
  bilanz: Bilanz[];
  namen: Map<string, string>;
}) {
  return (
    <section className="space-y-3">
      <h3 className="font-mono text-xs uppercase tracking-widest text-muted">
        Bilanz
      </h3>
      <ul className="space-y-3">
        {bilanz.map((b) => (
          <li key={b.mitglied_id} className="space-y-1">
            <div className="flex flex-wrap items-baseline justify-between gap-2 text-sm">
              <span className="font-medium">
                {namen.get(b.mitglied_id) ?? b.mitglied_id}
              </span>
              <span className="font-mono text-xs text-muted">
                {b.minuten} min · Kopflast {b.kopflast} · {b.aufgaben} Aufgaben ·{" "}
                <strong className="font-medium text-foreground">
                  {b.auslastung_prozent} %
                </strong>{" "}
                der verfügbaren Zeit
              </span>
            </div>
            <div
              className="h-1.5 overflow-hidden rounded-full bg-line"
              role="presentation"
            >
              <div
                className="h-full rounded-full bg-accent"
                style={{ width: `${Math.min(100, b.auslastung_prozent)}%` }}
              />
            </div>
          </li>
        ))}
      </ul>
      <p className="text-sm leading-relaxed text-muted">
        Verglichen wird die Auslastung, nicht die Minuten: Wer halb so viel Zeit
        hat, soll halb so viel tragen. Die Kopflast zählt dabei mit, kostet aber
        keine Uhrzeit — deshalb steht sie hier getrennt.
      </p>
    </section>
  );
}

const grundText: Record<string, string> = {
  gilt_nicht: "gilt für euch nicht",
  abgewaehlt: "abgewählt",
  nicht_faellig: "diese Woche nicht dran",
  startdichte: "bewusst zurückgehalten",
  keine_kapazitaet: "niemand hat Zeit",
  niemand_geeignet: "niemand geeignet",
};

function begruendungText(a: Aufgabe, namen: Map<string, string>): string {
  const zuletzt = a.begruendung.zuletzt_bei;
  switch (a.begruendung.code) {
    case "rotation":
      return zuletzt ? `zuletzt bei ${namen.get(zuletzt) ?? zuletzt}` : "Rotation";
    case "ausgleich":
      return "zum Ausgleich";
    case "feste_person":
      return "feste Zuständigkeit";
    case "einzige_moeglichkeit":
      return "einzige Möglichkeit";
    case "frist":
      return "wegen der Frist";
    case "eigene_aufgabe":
      return "eigene Aufgabe";
    default:
      return "";
  }
}
