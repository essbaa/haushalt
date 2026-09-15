"use client";

import Link from "next/link";
import { useRouter } from "next/navigation";
import { useState } from "react";
import { AbspracheRaster } from "@/app/components/absprache-raster";
import { EigeneAufgabe } from "@/app/components/eigene-aufgabe";
import type { Mitglied, VorlagenStand } from "@/lib/api";
import { patchMitToken } from "@/lib/browser-token";

const gruende: Record<string, string> = {
  gilt_nicht: "gilt bei euch nicht",
  abgewaehlt: "abgeschaltet",
  nicht_faellig: "nicht fällig",
  startdichte: "zurückgehalten",
  keine_kapazitaet: "keine Zeit frei",
  niemand_geeignet: "niemand geeignet",
  unbekannt: "ungeklärt",
  braucht_termin: "braucht Termin",
  braucht_absprache: "braucht Absprache",
};

// Alle zehn Bereiche aus planner.AllCategories. Eine Abbildung, die nur die
// bekannten Fälle abdeckt, verrät sich beim ersten unbekannten — hier stand
// schon einmal „termine" kleingeschrieben als Kennung.
const bereiche: Record<string, string> = {
  kueche: "Küche",
  waesche: "Wäsche",
  reinigung: "Reinigung",
  kind: "Kinder",
  vorrat: "Vorräte",
  termine: "Termine",
  verwaltung: "Verwaltung",
  wartung: "Wartung",
  sozial: "Soziales",
  aussen: "Draußen",
};

export function VorlagenListe({
  haushaltId,
  vorlagen,
  mitglieder,
  planend,
}: {
  haushaltId: string;
  vorlagen: VorlagenStand[];
  mitglieder: Mitglied[];
  planend: boolean;
}) {
  const router = useRouter();
  const [laeuft, setLaeuft] = useState<string | null>(null);
  const [fehler, setFehler] = useState<string | null>(null);

  // Zwei Zwecke, die sich widersprechen: Wer nachsehen will, was es gibt,
  // braucht alles; wer abschalten will, braucht das Geltende. Ein Filter
  // bedient beide, ohne eine zweite Seite.
  const [filter, setFilter] = useState<"alle" | "gilt" | "unklar" | "aus">("alle");

  const passt = (v: VorlagenStand) => {
    switch (filter) {
      case "gilt":
        return v.aktiv;
      case "unklar":
        return v.grund === "unbekannt";
      case "aus":
        return v.grund === "abgewaehlt";
      default:
        return true;
    }
  };

  async function antworten(faktum: string, wert: boolean) {
    setLaeuft(faktum);
    setFehler(null);
    try {
      await patchMitToken(`/api/haushalte/${encodeURIComponent(haushaltId)}/fakten`, {
        [faktum]: wert,
      });
      router.refresh();
    } catch (e) {
      setFehler(e instanceof Error ? e.message : "Das hat nicht funktioniert.");
    } finally {
      setLaeuft(null);
    }
  }

  async function umschalten(id: string, aktiv: boolean) {
    setLaeuft(id);
    setFehler(null);
    try {
      await patchMitToken<void>(
        `/api/haushalte/${encodeURIComponent(haushaltId)}/vorlagen/${encodeURIComponent(id)}`,
        { aktiv },
      );
      router.refresh();
    } catch (e) {
      setFehler(e instanceof Error ? e.message : "Das hat nicht funktioniert.");
    } finally {
      setLaeuft(null);
    }
  }

  const nachBereich = new Map<string, VorlagenStand[]>();
  for (const v of vorlagen.filter(passt)) {
    nachBereich.set(v.kategorie, [...(nachBereich.get(v.kategorie) ?? []), v]);
  }

  const aktive = vorlagen.filter((v) => v.aktiv).length;
  const zaehle = (grund: string) => vorlagen.filter((v) => v.grund === grund).length;

  return (
    <div className="space-y-5">
      <p className="text-sm leading-relaxed text-muted">
        <span className="font-bold text-fg">{aktive}</span> von {vorlagen.length}{" "}
        gelten bei euch. Bei den übrigen steht, warum nicht.
      </p>

      <div className="flex flex-wrap gap-2">
        {(
          [
            ["alle", "Alle", vorlagen.length],
            ["gilt", "Gilt bei euch", aktive],
            ["unklar", "Ungeklärt", zaehle("unbekannt")],
            ["aus", "Abgeschaltet", zaehle("abgewaehlt")],
          ] as const
        ).map(([wert, text, anzahl]) => (
          <button
            key={wert}
            type="button"
            onClick={() => setFilter(wert)}
            aria-pressed={filter === wert}
            className={`inline-flex min-h-10 items-center gap-1.5 rounded-full border px-3.5 text-sm font-semibold transition-colors ${
              filter === wert
                ? "border-primary bg-primary-soft text-primary"
                : "border-line-strong text-muted hover:text-fg"
            }`}
          >
            {text}
            <span className="text-xs opacity-70">{anzahl}</span>
          </button>
        ))}
      </div>

      {fehler && (
        <p role="alert" className="rounded-md border border-danger/40 px-3 py-2 text-sm text-danger">
          {fehler}
        </p>
      )}

      {[...nachBereich.keys()].sort().map((bereich) => (
        <Bereich
          key={bereich}
          name={bereiche[bereich] ?? bereich}
          kategorie={bereich}
          haushaltId={haushaltId}
          mitglieder={mitglieder}
          planend={planend}
          liste={[...nachBereich.get(bereich)!].sort((a, b) =>
            a.aktiv === b.aktiv ? a.titel.localeCompare(b.titel) : a.aktiv ? -1 : 1,
          )}
          laeuft={laeuft}
          umschalten={umschalten}
          antworten={antworten}
        />
      ))}
    </div>
  );
}

/**
 * Ein Bereich, auf- und zuklappbar.
 *
 * Kein <details>: Dort ist jeder Klick im <summary> ein Klick aufs Aufklappen,
 * und ein zweiter Knopf darin kämpft dagegen. Mit eigenem Zustand kann „+"
 * beides — den Bereich öffnen und gleich das Formular.
 */
function Bereich({
  name,
  kategorie,
  haushaltId,
  mitglieder,
  planend,
  liste,
  laeuft,
  umschalten,
  antworten,
}: {
  name: string;
  kategorie: string;
  haushaltId: string;
  mitglieder: Mitglied[];
  planend: boolean;
  liste: VorlagenStand[];
  laeuft: string | null;
  umschalten: (id: string, aktiv: boolean) => void;
  antworten: (faktum: string, wert: boolean) => void;
}) {
  const [offen, setOffen] = useState(false);
  const [neu, setNeu] = useState(false);
  const inhaltId = `bereich-${kategorie}`;

  return (
    <section className="overflow-hidden rounded-lg border border-line bg-surface">
      <div className="flex items-center gap-1 pr-2">
        <button
          type="button"
          onClick={() => setOffen((o) => !o)}
          aria-expanded={offen}
          aria-controls={inhaltId}
          className="flex min-h-14 flex-1 items-center gap-2 px-4 text-left text-base font-bold tracking-tight"
        >
          <span
            aria-hidden="true"
            className={`text-muted transition-transform ${offen ? "rotate-90" : ""}`}
          >
            ›
          </span>
          {name}
          <span className="ml-auto text-sm font-semibold text-muted">
            {liste.filter((v) => v.aktiv).length} von {liste.length}
          </span>
        </button>

        {planend && (
          <button
            type="button"
            onClick={() => {
              setOffen(true);
              setNeu(true);
            }}
            title={`Eigene Aufgabe in ${name}`}
            className="inline-flex size-11 shrink-0 items-center justify-center rounded-md text-xl font-bold text-primary transition-colors hover:bg-primary-soft"
          >
            <span aria-hidden="true">+</span>
            <span className="sr-only">Eigene Aufgabe in {name} anlegen</span>
          </button>
        )}
      </div>

      {offen && (
        <div id={inhaltId} className="border-t border-line">
          {planend && (
            <div className="border-b border-line px-3 py-2">
              <EigeneAufgabe
                haushaltId={haushaltId}
                kategorie={kategorie}
                offen={neu}
                setOffen={setNeu}
              />
            </div>
          )}

          <ul className="divide-y divide-line">
            {liste.map((v) => {
              const abgeschaltet = v.grund === "abgewaehlt";
              const knopf = v.aktiv
                ? { text: "Abschalten", nach: false }
                : abgeschaltet
                  ? { text: "Anschalten", nach: true }
                  : null;
              const frage =
                v.frage && v.faktum
                  ? { faktum: v.faktum, frage: v.frage, dann: v.dann ?? "" }
                  : undefined;
              // „Gilt bei euch nicht" hat zwei Ursachen mit zwei Rückwegen.
              // Hängt es an einem Faktum, das ihr verneint habt, steht die
              // Frage hier noch einmal — ein Nein, das man nicht zurücknehmen
              // kann, ist dasselbe wie ein Rateschluss, den man nicht
              // korrigieren kann (ADR-0007). Hängt es am Kontext (kein Garten,
              // kein Auto), führt der Weg über die Einstellungen.
              const verneint = v.grund === "gilt_nicht" && frage !== undefined;
              const ausKontext = v.grund === "gilt_nicht" && frage === undefined;

              return (
                <li key={v.id} className="flex flex-wrap items-center gap-x-3 gap-y-2 px-4 py-3">
                  <div className="min-w-0 flex-1">
                    <p className={`font-semibold ${v.aktiv ? "" : "text-muted"}`}>
                      {v.titel}
                      {v.eigene && (
                        <span className="ml-2 text-xs font-semibold text-subtle">
                          selbst angelegt
                        </span>
                      )}
                    </p>
                    {/* Die Häufigkeit stand hier nicht, und damit fehlte die
                        Zahl, die über den Aufwand entscheidet: „Abendessen
                        kochen, 40 Minuten" ist etwas anderes, wenn es dreimal
                        die Woche dran ist. Wer abschalten will, was zu viel
                        ist, muss das sehen können. */}
                    <p className="text-xs text-subtle">
                      {v.haeufigkeit && (
                        <span className="font-semibold text-muted">{v.haeufigkeit}</span>
                      )}
                      {v.haeufigkeit && " · "}
                      {v.dauer_min} min
                      {v.kopflast >= 2 && `, Kopflast ${v.kopflast}`}
                    </p>
                  </div>

                  {/* Der Zustand steht als Wort da, nicht als Grauton. Wer
                      raten muss, warum eine Zeile blass ist, hat keine
                      Auskunft bekommen, sondern eine Andeutung. */}
                  {v.grund && (
                    <span
                      className={`inline-flex shrink-0 items-center rounded-full px-2.5 py-1 text-xs font-semibold ${
                        abgeschaltet ? "bg-clay-soft text-clay" : "bg-surface-2 text-muted"
                      }`}
                    >
                      {gruende[v.grund] ?? v.grund}
                    </span>
                  )}

                  {planend && frage && (
                    <div className="w-full space-y-2 rounded-md bg-surface-2 p-3">
                      {verneint && (
                        <p className="text-xs font-semibold text-clay">
                          Ihr habt das mit Nein beantwortet. Ändert sich das,
                          antwortet hier neu.
                        </p>
                      )}
                      <p className="text-sm font-semibold text-pretty">{frage.frage}</p>
                      <p className="text-xs leading-relaxed text-muted text-pretty">{frage.dann}</p>
                      <div className="flex flex-wrap gap-2">
                        <button
                          type="button"
                          onClick={() => antworten(frage.faktum, true)}
                          disabled={laeuft !== null}
                          className="inline-flex min-h-10 items-center rounded-md bg-primary px-4 text-sm font-semibold text-on-primary transition-colors hover:bg-primary-hover disabled:opacity-45"
                        >
                          Ja
                        </button>
                        <button
                          type="button"
                          onClick={() => antworten(frage.faktum, false)}
                          disabled={laeuft !== null}
                          className="inline-flex min-h-10 items-center rounded-md border border-line-strong px-4 text-sm font-semibold text-muted transition-colors hover:text-fg disabled:opacity-45"
                        >
                          Nein
                        </button>
                      </div>
                    </div>
                  )}

                  {/* Die fehlende Bedingung im Klartext statt einer
                      Aufzählung aller denkbaren. Der Planer prüft sie ohnehin
                      — sie nicht zu nennen war eine Andeutung statt einer
                      Auskunft. */}
                  {ausKontext && (
                    <p className="w-full text-xs leading-relaxed text-muted text-pretty">
                      <span className="font-semibold text-fg">
                        {v.voraussetzung ?? "Hängt an euren Angaben zum Haushalt."}
                      </span>{" "}
                      Sobald das zutrifft, taucht sie von selbst wieder auf.
                      {planend && (
                        <>
                          {" "}
                          <Link
                            href={`/einstellungen?haushalt=${encodeURIComponent(haushaltId)}`}
                            className="font-semibold text-primary underline underline-offset-2"
                          >
                            In den Einstellungen ändern
                          </Link>
                        </>
                      )}
                    </p>
                  )}

                  {v.braucht_absprache && (
                    <AbspracheRaster
                      haushaltId={haushaltId}
                      vorlageId={v.id}
                      titel={v.titel}
                      mitglieder={mitglieder}
                      raster={v.absprache ?? ["", "", "", "", "", "", ""]}
                      planend={planend}
                    />
                  )}

                  {v.grund === "braucht_termin" && (
                    <p className="w-full text-xs leading-relaxed text-muted">
                      Entsteht aus einem Anlass, nicht aus einem Zeitraum. Trag
                      einen Termin ein, dann taucht sie rechtzeitig davor im Plan
                      auf — erfinden wird die App keinen.
                    </p>
                  )}

                  {planend && knopf && (
                    <button
                      type="button"
                      onClick={() => umschalten(v.id, knopf.nach)}
                      disabled={laeuft !== null}
                      className={`inline-flex min-h-10 shrink-0 items-center rounded-md px-3 text-sm font-semibold transition-colors disabled:opacity-45 ${
                        knopf.nach
                          ? "text-primary hover:bg-primary-soft"
                          : "text-subtle hover:bg-surface-2 hover:text-danger"
                      }`}
                    >
                      {knopf.text}
                    </button>
                  )}
                </li>
              );
            })}
          </ul>
        </div>
      )}
    </section>
  );
}
