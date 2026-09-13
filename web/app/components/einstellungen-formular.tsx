"use client";

import { useRouter } from "next/navigation";
import { useState } from "react";
import type { Haushalt, Mitglied } from "@/lib/api";
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

  async function schicke(tun: () => Promise<Haushalt>) {
    setLaeuft(true);
    setFehler(null);
    try {
      setStand(await tun());
      setGeaendert(true);
      setGerechnet(false);
      router.refresh();
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
      router.refresh();
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
          laeuft={laeuft}
          speichern={(aenderung) =>
            schicke(() =>
              patchMitToken<Haushalt>(
                `/api/haushalte/${encodeURIComponent(stand.id)}`,
                aenderung,
              ),
            )
          }
        />
      )}

      <section className="space-y-4">
        <h2 className="text-sm font-medium">Personen</h2>
        {stand.mitglieder.map((m) => (
          <PersonTeil
            key={m.id}
            person={m}
            planend={planend}
            ichSelbst={m.id === stand.ich}
            laeuft={laeuft}
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

      {fehler && <p className="text-sm text-clay">{fehler}</p>}

      {gerechnet && (
        <p className="rounded-lg border border-line bg-surface p-4 text-sm leading-relaxed text-muted">
          Die Woche ist neu gerechnet. Was schon abgehakt oder abgegeben war,
          steht unverändert da — nur die offenen Vorschläge wurden ersetzt.
        </p>
      )}

      {geaendert && (
        <div className="space-y-3 rounded-lg border border-line bg-surface p-4">
          <p className="text-sm leading-relaxed text-muted">
            Gespeichert. Wirksam wird es ab der nächsten Woche: Der laufende
            Plan steht fest, damit er sich niemandem unter den Händen ändert.
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
                Erledigtes und Abgegebenes bleibt, wie es ist. Ersetzt werden
                nur die Aufgaben, an denen noch nichts passiert ist.
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
    const neu = haustiere.includes(art)
      ? haustiere.filter((h) => h !== art)
      : [...haustiere, art];
    speichern({ haustiere: neu });
  }

  return (
    <section className="space-y-5">
      <h2 className="text-sm font-medium">Der Haushalt</h2>

      <div className="flex flex-wrap items-center gap-2">
        <input
          value={name}
          onChange={(e) => setName(e.target.value)}
          maxLength={80}
          className="min-w-0 flex-1 rounded-md border border-line bg-surface px-3 py-2"
        />
        <button
          type="button"
          onClick={() => speichern({ name })}
          disabled={laeuft || name.trim() === "" || name === stand.name}
          className="rounded-md border border-line px-3 py-2 text-sm transition-colors hover:border-primary hover:text-primary disabled:opacity-40"
        >
          Namen speichern
        </button>
      </div>

      <div className="flex flex-wrap gap-2">
        {(["wohnung", "haus"] as const).map((w) => (
          <Knopf
            key={w}
            an={stand.wohnform === w}
            text={w === "wohnung" ? "Wohnung" : "Haus"}
            laeuft={laeuft}
            klick={() => speichern({ wohnform: w })}
          />
        ))}
        <Knopf
          an={stand.garten === true}
          text="Garten"
          laeuft={laeuft}
          klick={() => speichern({ garten: !stand.garten })}
        />
        <Knopf
          an={stand.auto === true}
          text="Auto"
          laeuft={laeuft}
          klick={() => speichern({ auto: !stand.auto })}
        />
        <Knopf
          an={haustiere.includes("hund")}
          text="Hund"
          laeuft={laeuft}
          klick={() => tier("hund")}
        />
        <Knopf
          an={haustiere.includes("katze")}
          text="Katze"
          laeuft={laeuft}
          klick={() => tier("katze")}
        />
      </div>

      <p className="text-xs leading-relaxed text-muted">
        Diese fünf entscheiden, welche Aufgaben es im Haushalt überhaupt gibt.
        Jedes Nein spart Aufgaben, jedes Ja bringt welche mit — ohne Garten kein
        Rasen, ohne Auto kein TÜV.
      </p>
    </section>
  );
}

/** Ein Schalter, der sofort speichert. Kein „Übernehmen", kein Zustand, der
 *  im Formular auf den Server wartet: Was hier steht, ist immer das, was der
 *  Dienst zuletzt bestätigt hat. */
function Knopf({
  an,
  text,
  laeuft,
  klick,
}: {
  an: boolean;
  text: string;
  laeuft: boolean;
  klick: () => void;
}) {
  return (
    <button
      type="button"
      onClick={klick}
      disabled={laeuft}
      aria-pressed={an}
      className={`rounded-md border px-3 py-1.5 text-sm transition-colors disabled:opacity-50 ${
        an
          ? "border-primary bg-primary-soft text-primary"
          : "border-line text-muted hover:text-fg"
      }`}
    >
      {text}
    </button>
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
  const [minuten, setMinuten] = useState<number[]>(person.minuten ?? new Array(7).fill(0));
  const [offen, setOffen] = useState(false);

  // Die eigene Zeit setzt jeder selbst, alles Weitere die planenden. Dieselbe
  // Regel steht im Dienst; hier steuert sie nur, was man anfassen kann.
  const darfNamen = planend || ichSelbst;
  const darfZeit = planend || ichSelbst;

  function stufe(wert: string) {
    speichern({ zeit: wert });
  }

  const summe = minuten.reduce((a, b) => a + b, 0);

  return (
    <div className="space-y-3 rounded-lg border border-line bg-surface p-4">
      <div className="flex flex-wrap items-center gap-2">
        <input
          value={name}
          onChange={(e) => setName(e.target.value)}
          disabled={!darfNamen}
          maxLength={40}
          className="min-w-0 flex-1 rounded-md border border-line bg-transparent px-3 py-2 disabled:opacity-60"
        />
        <span className="font-mono text-xs text-muted">{person.rolle}</span>
        {darfNamen && name !== person.name && name.trim() !== "" && (
          <button
            type="button"
            onClick={() => speichern({ name })}
            disabled={laeuft}
            className="text-xs text-primary underline"
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
            className="w-24 rounded-md border border-line bg-transparent px-2 py-1"
          />
          {jahr !== (person.geburtsjahr?.toString() ?? "") && (
            <button
              type="button"
              onClick={() => speichern({ geburtsjahr: Number(jahr) || 0 })}
              disabled={laeuft}
              className="text-xs text-primary underline"
            >
              speichern
            </button>
          )}
          <span className="text-xs">daran hängen die Aufgaben</span>
        </label>
      )}

      {darfZeit && person.rolle !== "betreut" && (
        <div className="space-y-2">
          <div className="flex flex-wrap items-center gap-2">
            <span className="text-sm text-muted">Zeit</span>
            {["wenig", "mittel", "viel"].map((w) => (
              <button
                key={w}
                type="button"
                onClick={() => stufe(w)}
                disabled={laeuft}
                className="rounded-md border border-line px-3 py-1 text-sm text-muted transition-colors hover:border-primary hover:text-primary disabled:opacity-50"
              >
                {w}
              </button>
            ))}
            <span className="font-mono text-xs text-muted">
              zurzeit {Math.round(summe / 60)} h in der Woche
            </span>
          </div>

          <button
            type="button"
            onClick={() => setOffen((o) => !o)}
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
                        setMinuten((alt) =>
                          alt.map((v, k) => (k === i ? Number(e.target.value) : v)),
                        )
                      }
                      className="w-16 rounded-md border border-line bg-transparent px-2 py-1"
                    />
                  </label>
                ))}
              </div>
              <button
                type="button"
                onClick={() => speichern({ minuten })}
                disabled={laeuft}
                className="rounded-md border border-line px-3 py-1 text-sm transition-colors hover:border-primary hover:text-primary disabled:opacity-50"
              >
                Minuten speichern
              </button>
              <p className="text-xs leading-relaxed text-muted">
                Diese Zahlen tragen die gesamte Verteilung: Wer weniger Zeit
                hat, bekommt weniger. Null an einem Tag heißt, dass an dem Tag
                nichts eingeplant wird.
              </p>
            </div>
          )}
        </div>
      )}
    </div>
  );
}
