import Link from "next/link";
import { AufgabeZeile } from "@/app/components/aufgabe-zeile";
import { GestrichenZeile } from "@/app/components/gestrichen-zeile";
import { Fragen } from "@/app/components/fragen";
import type { Aufgabe, Bilanz, Mitglied, Wochenplan as Plan } from "@/lib/api";
import { farbklasse } from "@/lib/personen";
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

  // Der Stand der Woche. Nur für Wochen, die schon laufen: „0 von 13
  // erledigt" über einer Woche, die erst Montag anfängt, ist keine Auskunft,
  // sondern ein Vorwurf.
  const laeuft = tage.length > 0 && tage.some((t) => t <= heuteISO);
  const gesamt = plan.aufgaben.length;
  const fertig = plan.aufgaben.filter((a) => a.erledigt).length;

  return (
    <div className="space-y-10">
      {(laeuft || ich !== "") && tage.length > 0 && (
        <div className="flex items-center gap-4">
          {laeuft && <Fortschritt fertig={fertig} gesamt={gesamt} />}
          <div className="min-w-0 space-y-0.5">
            {ich !== "" && (
              <p className="text-lg leading-snug font-semibold text-pretty">
                {meineHeute === 0
                  ? "Heute ist für dich nichts offen."
                  : meineHeute === 1
                    ? "Heute ist eine Sache für dich offen."
                    : `Heute sind ${meineHeute} Sachen für dich offen.`}
              </p>
            )}
            {laeuft && (
              <p className="text-sm text-muted">
                {fertig === gesamt
                  ? "Die ganze Woche ist abgehakt."
                  : `${fertig} von ${gesamt} erledigt, im ganzen Haushalt.`}
              </p>
            )}
          </div>
        </div>
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
          {vorne.map((tag, i) => (
            <Tag
              key={tag}
              tag={tag}
              erster={i === 0}
              heute={tag === heuteISO}
              aufgaben={nachTag.get(tag)!}
              namen={namen}
              mitglieder={plan.haushalt.mitglieder}
              kandidaten={kandidaten}
              ich={ich}
              planend={planend}
              haushaltId={plan.haushalt.id}
            />
          ))}

          {rueckschau.length > 0 && (
            <details className="group border-t border-line pt-4">
              <summary className="inline-flex min-h-11 items-center gap-2 text-sm font-semibold text-muted transition-colors hover:text-fg">
                Schon gewesen ({rueckschau.length} {rueckschau.length === 1 ? "Tag" : "Tage"})
                <span aria-hidden="true" className="transition-transform group-open:rotate-90">
                  ›
                </span>
              </summary>
              <div className="mt-4 opacity-70">
                {rueckschau.map((tag, i) => (
                  <Tag
                    key={tag}
                    tag={tag}
                    erster={i === 0}
                    heute={false}
                    aufgaben={nachTag.get(tag)!}
                    namen={namen}
                    mitglieder={plan.haushalt.mitglieder}
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

      {plan.gestrichen && plan.gestrichen.length > 0 && <Gestrichene eintraege={plan.gestrichen} />}

      {plan.bilanz ? (
        <Bilanztafel
          bilanz={plan.bilanz}
          namen={namen}
          mitglieder={plan.haushalt.mitglieder}
          ich={ich}
        />
      ) : (
        <p className="text-sm leading-relaxed text-muted">
          Die Auswertung, wer wie viel trägt, sehen die planenden Personen. Du siehst den ganzen
          Plan.
        </p>
      )}

      {uebersprungen.length > 0 && (
        <details className="group border-t border-line pt-4">
          <summary className="inline-flex min-h-11 items-center gap-2 text-sm font-semibold text-muted transition-colors hover:text-fg">
            Nicht im Plan ({uebersprungen.length}) — und warum
            <span aria-hidden="true" className="transition-transform group-open:rotate-90">
              ›
            </span>
          </summary>
          <ul className="mt-1 space-y-1">
            {uebersprungen.map((u) => (
              <li
                key={`${u.vorlage_id}-${u.grund}`}
                className="flex flex-wrap justify-between gap-x-4 py-1 text-sm"
              >
                <span>{u.titel}</span>
                <span className="text-subtle">
                  {grundText[u.grund] ?? u.grund}
                  {/* Der einzige Grund mit einem Rückweg, den nur ihr gehen
                      könnt: Der Planer verteilt diese Aufgabe nicht, weil er
                      es nicht kann. Ohne diesen Weg stünde hier eine
                      Feststellung ohne Tür. */}
                  {u.grund === "braucht_absprache" && planend && (
                    <>
                      {" — "}
                      <Link
                        href={`/vorlagen?haushalt=${encodeURIComponent(plan.haushalt.id)}`}
                        className="font-semibold text-primary underline underline-offset-2"
                      >
                        jetzt absprechen
                      </Link>
                    </>
                  )}
                </span>
              </li>
            ))}
          </ul>
        </details>
      )}
    </div>
  );
}

/**
 * Der Stand der Woche als Ring.
 *
 * Zwei Kreise übereinander: der graue ganz, der grüne anteilig. Gerechnet
 * wird über die Strichlücke — `stroke-dasharray` legt die Länge des Strichs
 * fest, der Rest bleibt Lücke. Kein Bogen, den man aus Winkeln zusammensetzen
 * müsste, und deshalb auch nichts, was bei 0 oder 100 Prozent kippt.
 *
 * Die Zahl steht in der Mitte und nicht darunter: Wer hersieht, will wissen
 * „wie weit", nicht „wie viel Prozent von was". Der Satz daneben trägt die
 * genaue Auskunft; der Ring ist die Abkürzung.
 */
function Fortschritt({ fertig, gesamt }: { fertig: number; gesamt: number }) {
  const anteil = gesamt === 0 ? 0 : fertig / gesamt;
  const r = 20;
  const umfang = 2 * Math.PI * r;

  return (
    <div className="relative size-14 shrink-0" aria-hidden="true">
      <svg viewBox="0 0 48 48" className="size-full -rotate-90">
        <circle cx="24" cy="24" r={r} fill="none" strokeWidth="4" className="stroke-surface-3" />
        <circle
          cx="24"
          cy="24"
          r={r}
          fill="none"
          strokeWidth="4"
          strokeLinecap="round"
          className="stroke-primary transition-[stroke-dasharray] duration-500"
          strokeDasharray={`${umfang * anteil} ${umfang}`}
        />
      </svg>
      <span className="absolute inset-0 flex items-center justify-center text-sm font-extrabold tabular-nums">
        {Math.round(anteil * 100)}
        <span className="text-[0.6rem] font-bold">%</span>
      </span>
    </div>
  );
}

const wochentage = ["Mo", "Di", "Mi", "Do", "Fr", "Sa", "So"];

function Tag({
  tag,
  erster,
  heute,
  aufgaben,
  namen,
  mitglieder,
  kandidaten,
  ich,
  planend,
  haushaltId,
}: {
  tag: string;
  /** Der erste Tag braucht keine Trennlinie über sich. */
  erster: boolean;
  heute: boolean;
  aufgaben: Aufgabe[];
  namen: Record<string, string>;
  mitglieder: Mitglied[];
  kandidaten: { id: string; name: string }[];
  ich: string;
  planend: boolean;
  haushaltId: string;
}) {
  const [, monat, nummer] = tag.split("-").map(Number);
  const d = new Date(Date.UTC(Number(tag.slice(0, 4)), monat - 1, nummer));
  const kurz = wochentage[(d.getUTCDay() + 6) % 7];
  // Samstag und Sonntag stehen leiser da. Nicht aus Gestaltungslaune: Der
  // Blick sucht werktags den nächsten Werktag, und eine Woche, in der alle
  // sieben Marken gleich laut sind, muss man lesen statt zu überfliegen.
  const wochenende = d.getUTCDay() === 0 || d.getUTCDay() === 6;

  return (
    <div className={`flex gap-4 sm:gap-6 ${erster ? "" : "pt-4"}`}>
      {/* Die Spalte trägt den Tag und die Linie, an der die Woche hängt.
          Das Schild klebt beim Scrollen: Bei zehn Samstagsaufgaben stand es
          sonst längst oberhalb des Bildschirms, und man las eine Liste ohne
          zu wissen, zu welchem Tag sie gehört. */}
      <div className="flex w-12 shrink-0 flex-col items-center sm:w-16">
        <div
          className={`sticky top-16 flex size-11 flex-col items-center justify-center rounded-full leading-none ${
            heute
              ? "bg-primary text-on-primary"
              : wochenende
                ? "bg-bg text-subtle"
                : "bg-bg text-muted"
          }`}
        >
          {/* Zwei Zeilen in 44 Pixeln: Ohne feste Zeilenhöhen setzt die
              Schrift ihre eigenen Abstände, und die Marke sitzt sichtbar zu
              hoch. `leading-none` allein genügt nicht — die Versalhöhe von
              „MO" und die Ziffernhöhe sind verschieden, deshalb der halbe
              Pixel Ausgleich unten statt eines Abstands oben. */}
          <span className="text-[0.625rem] leading-[1] font-bold tracking-[0.08em] uppercase">
            {kurz}
          </span>
          <span className="mt-[0.1875rem] text-[1.0625rem] leading-[1] font-extrabold tabular-nums">
            {nummer}
          </span>
        </div>
        <div className="mt-1 w-px flex-1 bg-line" aria-hidden="true" />
      </div>

      {/* Eine Haarlinie über jedem Tag außer dem ersten.
          Kein Kasten — ADR-0009 hat die verworfen, weil ein Rahmen nur
          optisch trennt und nichts sagt. Eine Trennlinie sagt etwas: hier
          fängt ein neuer Tag an. Die senkrechte Linie daneben verbindet, und
          genau das war das Problem: Sie führt durch, wo eine Grenze liegt. */}
      <div
        className={`min-w-0 flex-1 space-y-1.5 pb-8 ${erster ? "" : "border-t border-line pt-3.5"}`}
      >
        {heute && <p className="pb-0.5 text-sm font-semibold text-primary">Heute</p>}
        {aufgaben.map((a, i) => (
          <div key={a.id ?? `${a.vorlage_id}-${i}`}>
            {/* Das Zeitfenster als leise Marke, und nur dort, wo es wechselt.
                Der Dienst sortiert den Tag jetzt danach (morgens, egal,
                abends); ohne die Marke merkt das niemand, und mit einer Marke
                an jeder Zeile stünde sie neunmal da.

                Bewusst klein und grau: Die App plant keine Uhrzeiten, das
                Fenster ist eine Angabe der Vorlage („Kita-Tasche packen:
                abends"). Eine Überschrift würde daraus einen Termin machen,
                den es nicht gibt. */}
            {a.zeitfenster !== "egal" && a.zeitfenster !== aufgaben[i - 1]?.zeitfenster && (
              <p className="px-3 pt-2 pb-1 text-[0.6875rem] font-semibold tracking-wide text-subtle uppercase">
                {a.zeitfenster}
              </p>
            )}
            <AufgabeZeile
              aufgabe={a}
              namen={namen}
              mitglieder={mitglieder}
              kandidaten={kandidaten}
              ich={ich}
              planend={planend}
              haushaltId={haushaltId}
            />
          </div>
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
  mitglieder,
  ich,
}: {
  bilanz: Bilanz[];
  namen: Record<string, string>;
  mitglieder: Mitglied[];
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
            <li
              key={b.mitglied_id}
              className={`${farbklasse(mitglieder, b.mitglied_id)} grid grid-cols-[5.5rem_1fr_3rem] items-center gap-3`}
            >
              <span
                className={`truncate text-sm ${
                  eigen ? "font-bold text-[var(--person)]" : "font-medium"
                }`}
              >
                {eigen ? "Du" : name}
              </span>
              {/* Derselbe Ton wie im Plan. Damit beantwortet der Balken zwei
                  Fragen auf einmal: wie viel — und wer das oben in der Woche
                  war. Vorher war jeder Balken grau außer dem eigenen, und die
                  Zuordnung kostete einen Blick zurück nach oben. */}
              <span className="h-3 overflow-hidden rounded-full bg-surface-3">
                <span
                  className="block h-full rounded-full bg-[var(--person)]"
                  style={{
                    width: `${Math.max(3, (b.auslastung_prozent / spitze) * 100)}%`,
                    opacity: eigen ? 1 : 0.75,
                  }}
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
        Angezeigt ist der Anteil der eigenen verfügbaren Zeit, nicht die Minuten. Kopflast zählt in
        der Verteilung mit, kostet aber keine Uhrzeit und steht deshalb nicht in dieser Zahl.
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
  braucht_absprache: "noch nicht abgesprochen",
};

/**
 * Was jemand für diese Woche gestrichen hat.
 *
 * Leise am Ende und nicht im Plan: Aus der Woche sind diese Termine raus, aus
 * der Bilanz auch. Ganz zu verschweigen wäre aber falsch — eine Zeile, die
 * wortlos verschwindet, lässt jemanden, der danebengewischt hat, ohne Rückweg.
 *
 * Der Tag steht dabei, weil eine Vorlage mehrfach in der Woche vorkommen kann:
 * Ohne ihn wüsste niemand, welches von fünf „Safiya zur Kita bringen"
 * gestrichen ist.
 */
function Gestrichene({ eintraege }: { eintraege: { id: string; titel: string; tag: string }[] }) {
  return (
    <section className="space-y-2 border-t border-line pt-6">
      <h2 className="text-sm font-bold tracking-tight text-muted">Gestrichen</h2>
      <ul className="space-y-1">
        {eintraege.map((g) => (
          <GestrichenZeile key={g.id} id={g.id} titel={g.titel} tag={g.tag} />
        ))}
      </ul>
    </section>
  );
}
