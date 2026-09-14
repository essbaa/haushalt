import { AufgabeZeile } from "@/app/components/aufgabe-zeile";
import { Fragen } from "@/app/components/fragen";
import type { Aufgabe, Bilanz, Wochenplan as Plan } from "@/lib/api";
import { heute } from "@/lib/woche";

/**
 * Der Wochenplan.
 *
 * Die Struktur ist die Woche selbst: links eine schmale Spalte mit dem Tag,
 * rechts, was an ihm ansteht. Die Zeit läuft nach unten, heute ist markiert.
 * Kein Kasten je Tag — ein Rahmen um jeden Abschnitt trennt nur optisch und
 * sagt nichts. Die Linie dagegen sagt: hier geht es weiter.
 *
 * Server Component; nur die Knöpfe an jeder Zeile brauchen den Browser.
 */
export function Wochenplan({ plan }: { plan: Plan }) {
  const namen = Object.fromEntries(plan.haushalt.mitglieder.map((m) => [m.id, m.name]));
  // Wer überhaupt Aufgaben übernimmt. Betreute Personen stehen im Haushalt,
  // aber nicht im Plan — sie zur Auswahl anzubieten hieße, eine Antwort
  // anzubieten, die der Dienst sicher ablehnt.
  const kandidaten = plan.haushalt.mitglieder
    .filter((m) => m.rolle !== "betreut")
    .map((m) => ({ id: m.id, name: m.name }));
  const ich = plan.ich ?? "";
  const planend = plan.meine_rolle === "planend";
  const heuteISO = heute();

  const nachTag = new Map<string, Aufgabe[]>();
  for (const a of plan.aufgaben) {
    nachTag.set(a.tag, [...(nachTag.get(a.tag) ?? []), a]);
  }
  const tage = [...nachTag.keys()].sort();
  const meineHeute = (nachTag.get(heuteISO) ?? []).filter(
    (a) => a.zustaendig === ich && !a.erledigt,
  ).length;

  // Vergangene Tage kommen nach unten und zusammengeklappt.
  //
  // Der Reihe nach wäre kalendarisch richtig und im Alltag falsch herum: Am
  // Sonntag scrollt man an sechs erledigten Tagen vorbei, um zu sehen, was
  // jetzt dran ist. Die Frage lautet „was ist heute", nicht „wie war die
  // Woche". Liegt die ganze Woche in der Vergangenheit — jemand sieht sich
  // eine alte an —, wird nichts eingeklappt; dann ist Rückschau der Zweck.
  // Eine Vorlage kann mehrfach übersprungen werden — der Planer meldet je
  // Fälligkeit, und eine Auslöser-Vorlage ist siebenmal fällig. Für die Frage
  // „warum steht das nicht in meinem Plan" zählt die Antwort, nicht wie oft
  // sie zutrifft.
  const gesehen = new Set<string>();
  const uebersprungen = plan.uebersprungen.filter((u) => {
    const schluessel = `${u.vorlage_id}|${u.grund}`;
    if (gesehen.has(schluessel)) return false;
    gesehen.add(schluessel);
    return true;
  });

  const vergangen = tage.filter((t) => t < heuteISO);
  const ab_heute = tage.filter((t) => t >= heuteISO);
  const rueckschau = ab_heute.length === 0 ? [] : vergangen;
  const vorne = ab_heute.length === 0 ? tage : ab_heute;

  return (
    <div className="space-y-10">
      {ich !== "" && (
        <p className="text-lg leading-snug text-pretty">
          {meineHeute === 0
            ? "Heute ist für dich nichts offen."
            : meineHeute === 1
              ? "Heute ist eine Sache für dich offen."
              : `Heute sind ${meineHeute} Sachen für dich offen.`}
        </p>
      )}

      {/* Die Fragen stehen vor der Woche, nicht dahinter.
          Nach ADR-0010 ist eine Frage das, was die App tut, statt zu raten —
          und eine Frage hinter zwei Dutzend Zeilen wird nicht beantwortet.
          Dann rät sie weiter. Beantwortet verschwindet sie; die Kosten sind
          vorübergehend, der Nutzen bleibt. */}
      {plan.fragen && plan.fragen.length > 0 && (
        <Fragen haushaltId={plan.haushalt.id} woche={plan.woche} fragen={plan.fragen} />
      )}

      {tage.length === 0 ? (
        <div className="rounded-lg border border-line bg-surface px-4 py-8 text-center">
          <p className="font-semibold">Diese Woche steht nichts an.</p>
          <p className="mt-1 text-sm text-muted">
            Kein Fehler: Die erste Woche bleibt bewusst leer, statt zu erschlagen.
          </p>
        </div>
      ) : (
        <section aria-label={`Woche ${plan.woche}`}>
          {vorne.map((tag) => (
            <Tag
              key={tag}
              tag={tag}
              heute={tag === heuteISO}
              aufgaben={nachTag.get(tag)!}
              namen={namen}
              kandidaten={kandidaten}
              ich={ich}
              planend={planend}
              haushaltId={plan.haushalt.id}
            />
          ))}

          {rueckschau.length > 0 && (
            <details className="group border-t border-line pt-4">
              <summary className="inline-flex min-h-11 items-center gap-2 text-sm font-semibold text-muted transition-colors hover:text-fg">
                Schon gewesen ({rueckschau.length}{" "}
                {rueckschau.length === 1 ? "Tag" : "Tage"})
                <span aria-hidden="true" className="transition-transform group-open:rotate-90">
                  ›
                </span>
              </summary>
              <div className="mt-4 opacity-70">
                {rueckschau.map((tag) => (
                  <Tag
                    key={tag}
                    tag={tag}
                    heute={false}
                    aufgaben={nachTag.get(tag)!}
                    namen={namen}
                    kandidaten={kandidaten}
                    ich={ich}
                    planend={planend}
                    haushaltId={plan.haushalt.id}
                  />
                ))}
              </div>
            </details>
          )}
        </section>
      )}

      {plan.bilanz ? (
        <Bilanztafel bilanz={plan.bilanz} namen={namen} ich={ich} />
      ) : (
        <p className="text-sm leading-relaxed text-muted">
          Die Auswertung, wer wie viel trägt, sehen die planenden Personen. Du
          siehst den ganzen Plan.
        </p>
      )}

      {uebersprungen.length > 0 && (
        <details className="group border-t border-line pt-4">
          <summary className="inline-flex min-h-11 items-center gap-2 text-sm font-semibold text-muted transition-colors hover:text-fg">
            Nicht im Plan ({uebersprungen.length})
            <span aria-hidden="true" className="transition-transform group-open:rotate-90">
              ›
            </span>
          </summary>
          <ul className="mt-1 space-y-1">
            {uebersprungen.map((u) => (
              <li key={`${u.vorlage_id}-${u.grund}`} className="flex flex-wrap justify-between gap-x-4 py-1 text-sm">
                <span>{u.titel}</span>
                <span className="text-subtle">{grundText[u.grund] ?? u.grund}</span>
              </li>
            ))}
          </ul>
        </details>
      )}
    </div>
  );
}

const wochentage = ["Mo", "Di", "Mi", "Do", "Fr", "Sa", "So"];

function Tag({
  tag,
  heute,
  aufgaben,
  namen,
  kandidaten,
  ich,
  planend,
  haushaltId,
}: {
  tag: string;
  heute: boolean;
  aufgaben: Aufgabe[];
  namen: Record<string, string>;
  kandidaten: { id: string; name: string }[];
  ich: string;
  planend: boolean;
  haushaltId: string;
}) {
  const [, monat, nummer] = tag.split("-").map(Number);
  const d = new Date(Date.UTC(Number(tag.slice(0, 4)), monat - 1, nummer));
  const kurz = wochentage[(d.getUTCDay() + 6) % 7];

  return (
    <div className="flex gap-4 sm:gap-6">
      {/* Die Spalte trägt den Tag und die Linie, an der die Woche hängt. */}
      <div className="flex w-12 shrink-0 flex-col items-center sm:w-16">
        <div
          className={`flex size-11 flex-col items-center justify-center rounded-full leading-none ${
            heute ? "bg-primary text-on-primary" : "text-muted"
          }`}
        >
          <span className="text-[0.7rem] font-semibold">{kurz}</span>
          <span className="text-base font-extrabold tabular-nums">{nummer}</span>
        </div>
        <div className="mt-1 w-px flex-1 bg-line" aria-hidden="true" />
      </div>

      <div className="min-w-0 flex-1 space-y-1 pb-7">
        {heute && <p className="pb-1 text-sm font-semibold text-primary">Heute</p>}
        {aufgaben.map((a, i) => (
          <AufgabeZeile
            key={a.id ?? `${a.vorlage_id}-${i}`}
            aufgabe={a}
            namen={namen}
            kandidaten={kandidaten}
            ich={ich}
            planend={planend}
            haushaltId={haushaltId}
          />
        ))}
      </div>
    </div>
  );
}

/**
 * Die Bilanz als Gegenüberstellung.
 *
 * Kein Fortschrittsbalken je Person: Das Versprechen lautet nicht „schaffe
 * hundert Prozent", sondern „wer halb so viel Zeit hat, trägt halb so viel".
 * Also stehen alle auf derselben Skala, und was man vergleicht, ist die Länge
 * nebeneinander — die Zahl daneben sagt, wovon.
 */
function Bilanztafel({
  bilanz,
  namen,
  ich,
}: {
  bilanz: Bilanz[];
  namen: Record<string, string>;
  ich: string;
}) {
  const spitze = Math.max(1, ...bilanz.map((b) => b.auslastung_prozent));

  return (
    <section className="space-y-4 border-t border-line pt-6">
      <h2 className="text-lg font-bold tracking-tight">Wer trägt wie viel</h2>

      <ul className="space-y-4">
        {bilanz.map((b) => {
          const name = namen[b.mitglied_id] ?? b.mitglied_id;
          const eigen = b.mitglied_id === ich;
          return (
            <li key={b.mitglied_id} className="grid grid-cols-[5.5rem_1fr_3rem] items-center gap-3">
              <span className={`truncate text-sm ${eigen ? "font-bold text-primary" : "font-medium"}`}>
                {eigen ? "Du" : name}
              </span>
              <span className="h-3 overflow-hidden rounded-full bg-surface-3">
                <span
                  className={`block h-full rounded-full ${eigen ? "bg-primary" : "bg-line-strong"}`}
                  style={{ width: `${Math.max(3, (b.auslastung_prozent / spitze) * 100)}%` }}
                />
              </span>
              <span className="text-right text-sm font-bold tabular-nums">
                {b.auslastung_prozent}%
              </span>
            </li>
          );
        })}
      </ul>

      <p className="text-xs leading-relaxed text-muted">
        Angezeigt ist der Anteil der eigenen verfügbaren Zeit, nicht die Minuten.
        Kopflast zählt in der Verteilung mit, kostet aber keine Uhrzeit und steht
        deshalb nicht in dieser Zahl.
      </p>

      <ul className="grid gap-x-6 gap-y-1 text-xs text-muted sm:grid-cols-2">
        {bilanz.map((b) => (
          <li key={b.mitglied_id} className="flex justify-between gap-2">
            <span>{namen[b.mitglied_id] ?? b.mitglied_id}</span>
            <span className="tabular-nums">
              {b.minuten} min, Kopflast {b.kopflast}, {b.aufgaben}{" "}
              {b.aufgaben === 1 ? "Aufgabe" : "Aufgaben"}
            </span>
          </li>
        ))}
      </ul>
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
  unbekannt: "noch ungeklärt",
  braucht_termin: "braucht einen Termin",
};
