"use client";

import { useRouter } from "next/navigation";
import { useState, useTransition } from "react";
import type { Haushalt, Mitglied } from "@/lib/api";
import { Auswahl, Marke } from "@/app/components/ui";
import { patchMitToken, postMitToken } from "@/lib/browser-token";

const TAGE = ["Mo", "Di", "Mi", "Do", "Fr", "Sa", "So"];

/**
 * Einstellungen: was das Onboarding geraten hat, hier korrigieren.
 *
 * Die laufende Woche steht fest (ADR-0008). Eine Änderung wirkt deshalb erst
 * ab der nächsten — und das muss dastehen, sonst wirkt die Einstellung kaputt:
 * Wer „kein Garten" einträgt und das Rasenmähen bis Sonntag im Plan sieht,
 * glaubt nicht der Erklärung, sondern dem Plan.
 */
export function EinstellungenFormular({
  haushalt,
  planend,
  woche,
}: {
  haushalt: Haushalt;
  planend: boolean;
  woche: string;
}) {
  const router = useRouter();
  const [stand, setStand] = useState(haushalt);
  const [fehler, setFehler] = useState<string | null>(null);
  const [laeuft, setLaeuft] = useState(false);
  const [geaendert, setGeaendert] = useState(false);
  const [gerechnet, setGerechnet] = useState(false);
  // router.refresh() holt die Server Components neu und ist nicht abgewartet:
  // Ohne useTransition steht der Knopf wieder bereit, während die neuen Daten
  // noch unterwegs sind — die Zeile sieht fertig aus und ändert sich eine
  // Sekunde später doch noch. `uebergang` hält den Zustand, bis der Server
  // geantwortet hat.
  const [uebergang, starten] = useTransition();
  const beschaeftigt = laeuft || uebergang;

  async function schicke(tun: () => Promise<Haushalt>) {
    setLaeuft(true);
    setFehler(null);
    try {
      setStand(await tun());
      setGeaendert(true);
      setGerechnet(false);
      starten(() => router.refresh());
    } catch (e) {
      setFehler(e instanceof Error ? e.message : "Das hat nicht funktioniert.");
    } finally {
      setLaeuft(false);
    }
  }

  async function neuRechnen() {
    setLaeuft(true);
    setFehler(null);
    try {
      await postMitToken(
        `/api/haushalte/${encodeURIComponent(stand.id)}/plan/${encodeURIComponent(woche)}/neu`,
      );
      setGerechnet(true);
      setGeaendert(false);
      starten(() => router.refresh());
    } catch (e) {
      setFehler(e instanceof Error ? e.message : "Das hat nicht funktioniert.");
    } finally {
      setLaeuft(false);
    }
  }

  return (
    <div className="space-y-10">
      {planend && (
        <HaushaltTeil
          stand={stand}
          laeuft={beschaeftigt}
          speichern={(aenderung) =>
            schicke(() =>
              patchMitToken<Haushalt>(`/api/haushalte/${encodeURIComponent(stand.id)}`, aenderung),
            )
          }
        />
      )}

      <section className="space-y-4">
        <h2 className="text-lg font-bold tracking-tight">Personen</h2>
        {stand.mitglieder.map((m) => (
          <PersonTeil
            key={m.id}
            person={m}
            planend={planend}
            ichSelbst={m.id === stand.ich}
            laeuft={beschaeftigt}
            speichern={(aenderung) =>
              schicke(() =>
                patchMitToken<Haushalt>(
                  `/api/haushalte/${encodeURIComponent(stand.id)}/mitglieder/${encodeURIComponent(m.id)}`,
                  aenderung,
                ),
              )
            }
          />
        ))}
      </section>

      {fehler && (
        <p className="rounded-md border border-danger/40 px-3 py-2 text-sm text-danger">{fehler}</p>
      )}

      {gerechnet && (
        <p className="rounded-lg border border-line bg-surface p-4 text-sm leading-relaxed text-muted">
          Die Woche ist neu gerechnet. Was schon abgehakt oder abgegeben war, steht unverändert da —
          nur die offenen Vorschläge wurden ersetzt.
        </p>
      )}

      {geaendert && (
        <div className="space-y-3 rounded-lg border border-line bg-surface p-4 sm:p-5">
          <p className="text-sm leading-relaxed text-muted">
            Gespeichert. Wirksam wird es ab der nächsten Woche: Der laufende Plan steht fest, damit
            er sich niemandem unter den Händen ändert.
          </p>
          {planend && (
            <>
              <button
                type="button"
                onClick={neuRechnen}
                disabled={laeuft}
                className="rounded-md border border-line px-4 py-2 text-sm transition-colors hover:border-primary hover:text-primary disabled:opacity-50"
              >
                Diese Woche neu rechnen
              </button>
              <p className="text-xs leading-relaxed text-muted">
                Erledigtes und Abgegebenes bleibt, wie es ist. Ersetzt werden nur die Aufgaben, an
                denen noch nichts passiert ist.
              </p>
            </>
          )}
        </div>
      )}
    </div>
  );
}

function HaushaltTeil({
  stand,
  laeuft,
  speichern,
}: {
  stand: Haushalt;
  laeuft: boolean;
  speichern: (a: Record<string, unknown>) => void;
}) {
  const [name, setName] = useState(stand.name);
  const haustiere = stand.haustiere ?? [];

  function tier(art: string) {
    const neu = haustiere.includes(art) ? haustiere.filter((h) => h !== art) : [...haustiere, art];
    speichern({ haustiere: neu });
  }

  return (
    <section className="space-y-5">
      <h2 className="text-lg font-bold tracking-tight">Der Haushalt</h2>

      <div className="flex flex-wrap items-center gap-2">
        <input
          value={name}
          onChange={(e) => setName(e.target.value)}
          maxLength={80}
          className="min-h-11 min-w-0 flex-1 rounded-md border border-line-strong bg-surface px-3 text-base"
        />
        <button
          type="button"
          onClick={() => speichern({ name })}
          disabled={laeuft || name.trim() === "" || name === stand.name}
          className="inline-flex min-h-11 items-center rounded-md border border-line-strong px-4 text-sm font-semibold transition-colors hover:border-primary hover:text-primary disabled:opacity-40"
        >
          Namen speichern
        </button>
      </div>

      {/* Die Beschriftung braucht `block`. Ohne das stehen span und input als
          zwei Inline-Elemente auf derselben Zeile und kleben aneinander —
          `space-y` setzt nur einen oberen Abstand, und der bewirkt in einer
          Zeile nichts. An den anderen Stellen fällt es nicht auf, weil dort
          das Eingabefeld selbst `block` trägt. */}
      <div className="flex flex-wrap items-end gap-4">
        <label className="block space-y-1.5">
          <span className="block text-sm font-semibold">Zimmer</span>
          <input
            type="number"
            inputMode="numeric"
            min={1}
            max={15}
            defaultValue={stand.zimmer ?? 3}
            onBlur={(e) => {
              const wert = Number(e.target.value);
              if (wert && wert !== stand.zimmer) speichern({ zimmer: wert });
            }}
            className="min-h-11 w-24 rounded-md border border-line-strong bg-surface px-3 text-base"
          />
        </label>
        <label className="block space-y-1.5">
          <span className="block text-sm font-semibold">Bäder</span>
          <input
            type="number"
            inputMode="numeric"
            min={0}
            max={5}
            defaultValue={stand.baeder ?? 1}
            onBlur={(e) => {
              const wert = Number(e.target.value);
              if (wert !== stand.baeder) speichern({ baeder: wert });
            }}
            className="min-h-11 w-24 rounded-md border border-line-strong bg-surface px-3 text-base"
          />
        </label>
      </div>
      <p className="text-xs leading-relaxed text-muted">
        Danach richtet sich, wie lange Putzaufgaben dauern. Ändert ihr das, gilt es ab der nächsten
        Woche.
      </p>

      <div className="flex flex-wrap gap-2">
        <Marke
          an={stand.garten === true}
          text="Garten"
          aus={laeuft}
          klick={() => speichern({ garten: !stand.garten })}
        />
        <Marke
          an={stand.auto === true}
          text="Auto"
          aus={laeuft}
          klick={() => speichern({ auto: !stand.auto })}
        />
        <Marke
          an={haustiere.includes("hund")}
          text="Hund"
          aus={laeuft}
          klick={() => tier("hund")}
        />
        <Marke
          an={haustiere.includes("katze")}
          text="Katze"
          aus={laeuft}
          klick={() => tier("katze")}
        />
      </div>

      <p className="text-xs leading-relaxed text-muted">
        Diese vier entscheiden, welche Aufgaben es im Haushalt überhaupt gibt. Jedes Nein spart
        Aufgaben, jedes Ja bringt welche mit — ohne Garten kein Rasen, ohne Auto kein TÜV.
      </p>
    </section>
  );
}

function PersonTeil({
  person,
  planend,
  ichSelbst,
  laeuft,
  speichern,
}: {
  person: Mitglied;
  planend: boolean;
  ichSelbst: boolean;
  laeuft: boolean;
  speichern: (a: Record<string, unknown>) => void;
}) {
  const [name, setName] = useState(person.name);
  const [jahr, setJahr] = useState(person.geburtsjahr?.toString() ?? "");
  // Die Minuten liegen nur im Zustand, solange sie bearbeitet werden.
  //
  // Vorher hielt `useState` sie dauerhaft — und `useState` nimmt den
  // Anfangswert genau einmal. Nach dem Speichern kam der neue Haushalt
  // zurück, die Prop änderte sich, der Zustand nicht: Ein Klick auf „mittel"
  // war längst gespeichert, während die Zeile weiter die alte Stundenzahl
  // zeigte. „Nichts passiert" war die einzig mögliche Schlussfolgerung.
  //
  // Zugeklappt kommt die Zahl deshalb aus dem Serverstand, und beim Aufklappen
  // wird der Entwurf frisch daraus gefüllt. Das ist die Regel dahinter: Was
  // der Server weiß, wird nicht nebenher im Browser mitgeführt.
  const [entwurf, setEntwurf] = useState<number[]>([]);
  const [offen, setOffen] = useState(false);

  const gespeichert = person.minuten ?? new Array(7).fill(0);
  const minuten = offen ? entwurf : gespeichert;

  // Die eigene Zeit setzt jeder selbst, alles Weitere die planenden. Dieselbe
  // Regel steht im Dienst; hier steuert sie nur, was man anfassen kann.
  const darfNamen = planend || ichSelbst;
  const darfZeit = planend || ichSelbst;

  function stufe(wert: string) {
    speichern({ zeit: wert });
  }

  const summe = minuten.reduce((a, b) => a + b, 0);

  return (
    <div
      className={`space-y-3 rounded-lg border p-4 sm:p-5 ${
        ichSelbst ? "border-primary/35 bg-primary-soft/40" : "border-line bg-surface"
      }`}
    >
      {ichSelbst && <p className="text-sm font-semibold text-primary">Du</p>}
      <div className="flex flex-wrap items-center gap-2">
        <input
          value={name}
          onChange={(e) => setName(e.target.value)}
          disabled={!darfNamen}
          maxLength={40}
          aria-label={ichSelbst ? "Dein Name" : `Name von ${person.name}`}
          className="min-h-11 min-w-0 flex-1 rounded-md border border-line-strong bg-transparent px-3 text-base disabled:opacity-60"
        />
        <span className="inline-flex items-center rounded-full bg-surface-2 px-2.5 py-1 text-xs font-semibold text-muted">
          {rollenText[person.rolle] ?? person.rolle}
        </span>
        {darfNamen && name !== person.name && name.trim() !== "" && (
          <button
            type="button"
            onClick={() => speichern({ name })}
            disabled={laeuft}
            className="inline-flex min-h-11 items-center text-sm font-semibold text-primary underline underline-offset-2"
          >
            Namen speichern
          </button>
        )}
      </div>

      {planend && person.rolle !== "planend" && (
        <label className="flex flex-wrap items-center gap-2 text-sm text-muted">
          Geburtsjahr
          <input
            type="number"
            inputMode="numeric"
            min={1900}
            max={new Date().getFullYear()}
            value={jahr}
            onChange={(e) => setJahr(e.target.value)}
            className="min-h-11 w-24 rounded-md border border-line-strong bg-transparent px-3 text-base"
          />
          {jahr !== (person.geburtsjahr?.toString() ?? "") && (
            <button
              type="button"
              onClick={() => speichern({ geburtsjahr: Number(jahr) || 0 })}
              disabled={laeuft}
              className="inline-flex min-h-11 items-center text-sm font-semibold text-primary underline underline-offset-2"
            >
              speichern
            </button>
          )}
          <span className="text-xs">daran hängen die Aufgaben</span>
        </label>
      )}

      {/* Betreuung: beim Einrichten aus dem Alter geraten, hier korrigierbar.
          Genau hier lag der Fehler — ein Zweijähriges Kita-Kind galt als
          „keine Betreuung", und damit fielen alle Kita-Aufgaben weg. */}
      {planend && person.rolle !== "planend" && (
        <div className="space-y-2">
          <p className="text-sm text-muted">Betreuung</p>
          <Auswahl
            name={`Betreuung von ${person.name}`}
            wert={person.betreuung ?? "keine"}
            aus={laeuft}
            auf={(b) => speichern({ betreuung: b })}
            optionen={[
              { wert: "keine", text: "keine" },
              { wert: "kita", text: "Kita" },
              { wert: "schule", text: "Schule" },
            ]}
          />
        </div>
      )}

      {darfZeit && person.rolle !== "betreut" && (
        <div className="space-y-2">
          <div className="space-y-2">
            <p className="flex flex-wrap items-center gap-x-2 gap-y-1 text-sm">
              <span>
                <span className="font-semibold">{stunden(summe)}</span>{" "}
                <span className="text-muted">Zeit in der Woche</span>
              </span>
              {!person.zeit && (
                <span className="inline-flex items-center rounded-full bg-surface-2 px-2.5 py-1 text-xs font-semibold text-muted">
                  eigene Minuten
                </span>
              )}
            </p>
            {/* Welche Stufe gerade gilt, rechnet der Planer aus und schickt es
                mit (`zeit`, siehe BudgetOf). Die Zuordnung hier nachzubauen
                wäre eine zweite Wahrheit gewesen — sie nicht anzuzeigen war
                aber auch keine Lösung: Ein Knopfpaar ohne Zustand lässt den
                Nutzer raten, was gerade eingestellt ist.
                Passt keine Stufe, ist keine markiert und daneben steht, warum. */}
            <div className="flex flex-wrap items-center gap-2">
              {["wenig", "mittel", "viel"].map((w) => {
                const aktiv = person.zeit === w;
                return (
                  <button
                    key={w}
                    type="button"
                    onClick={() => stufe(w)}
                    disabled={laeuft}
                    aria-pressed={aktiv}
                    className={`inline-flex min-h-10 items-center rounded-md border px-3.5 text-sm font-semibold transition-colors disabled:opacity-45 ${
                      aktiv
                        ? "border-primary bg-primary-soft text-primary"
                        : "border-line-strong hover:border-primary hover:text-primary"
                    }`}
                  >
                    {w}
                  </button>
                );
              })}
            </div>
          </div>

          <button
            type="button"
            onClick={() => {
              if (!offen) setEntwurf(gespeichert);
              setOffen((o) => !o);
            }}
            className="text-xs text-muted underline"
          >
            {offen ? "Minuten zuklappen" : "Minuten je Tag"}
          </button>

          {offen && (
            <div className="space-y-2">
              <div className="flex flex-wrap gap-2">
                {TAGE.map((t, i) => (
                  <label key={t} className="flex flex-col gap-1 text-xs text-muted">
                    {t}
                    <input
                      type="number"
                      min={0}
                      max={480}
                      value={minuten[i] ?? 0}
                      onChange={(e) =>
                        setEntwurf((alt) =>
                          alt.map((v, k) => (k === i ? Number(e.target.value) : v)),
                        )
                      }
                      className="min-h-11 w-16 rounded-md border border-line-strong bg-transparent px-2 text-center text-base tabular-nums"
                    />
                  </label>
                ))}
              </div>
              <button
                type="button"
                onClick={() => speichern({ minuten })}
                disabled={laeuft}
                className="inline-flex min-h-10 items-center rounded-md border border-line-strong px-3.5 text-sm font-semibold transition-colors hover:border-primary hover:text-primary disabled:opacity-45"
              >
                Minuten speichern
              </button>
              <p className="text-xs leading-relaxed text-muted">
                Diese Zahlen tragen die gesamte Verteilung: Wer weniger Zeit hat, bekommt weniger.
                Null an einem Tag heißt, dass an dem Tag nichts eingeplant wird.
              </p>
            </div>
          )}
        </div>
      )}
    </div>
  );
}

const rollenText: Record<string, string> = {
  planend: "plant mit",
  ausfuehrend: "führt aus",
  betreut: "wird betreut",
};

/** „9 Stunden" oder „8,5 Stunden" — halbe Stunden werden nicht gerundet, sonst
 *  stimmt die Zahl nicht mit dem überein, was jemand gerade eingestellt hat. */
function stunden(minuten: number): string {
  const wert = minuten / 60;
  const text = Number.isInteger(wert) ? String(wert) : wert.toFixed(1).replace(".", ",");
  return `${text} ${wert === 1 ? "Stunde" : "Stunden"}`;
}
