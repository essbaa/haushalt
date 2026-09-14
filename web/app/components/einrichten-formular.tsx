"use client";

import { useRouter } from "next/navigation";
import { useState } from "react";
import { Auswahl, Feld, Marke } from "@/app/components/ui";
import type { Haushalt, NeuerHaushalt, NeuesMitglied } from "@/lib/api";
import { postMitToken } from "@/lib/browser-token";

type Art = "erwachsen" | "kind" | "betreut";
type Zeit = "wenig" | "mittel" | "viel";

type Person = {
  id: number;
  name: string;
  art: Art;
  plant: boolean;
  geburtsjahr: string;
  zeit: Zeit;
};

const JAHR = new Date().getFullYear();

/**
 * Das Onboarding in drei Schritten.
 *
 * Gefragt wird nur, was den ersten Plan verändert: Wohnform, Personen mit
 * grobem Alter, Zeitbudget. Keine Arbeitszeiten, keine Vorlieben — eine App
 * gegen Mental Load darf beim Einrichten keine erzeugen. Vorlieben lernt sie
 * später aus dem Umverteilen.
 *
 * Der Zustand lebt im Speicher, nicht im Browserspeicher: Ein abgebrochenes
 * Onboarding soll nichts hinterlassen, das später halb ausgefüllt wieder
 * auftaucht.
 */
export function EinrichtenFormular({ meinName }: { meinName: string }) {
  const router = useRouter();
  const [schritt, setSchritt] = useState(1);
  const [laeuft, setLaeuft] = useState(false);
  const [fehler, setFehler] = useState<string | null>(null);

  const [name, setName] = useState(meinName ? `Haushalt von ${meinName}` : "Mein Haushalt");
  const [zimmer, setZimmer] = useState("3");
  const [baeder, setBaeder] = useState("1");
  const [garten, setGarten] = useState(false);
  const [auto, setAuto] = useState(false);
  const [haustiere, setHaustiere] = useState<string[]>([]);

  const [personen, setPersonen] = useState<Person[]>([
    { id: 0, name: meinName || "Ich", art: "erwachsen", plant: true, geburtsjahr: "", zeit: "mittel" },
  ]);

  function aendern(id: number, teil: Partial<Person>) {
    setPersonen((alt) => alt.map((p) => (p.id === id ? { ...p, ...teil } : p)));
  }

  function haustier(art: string) {
    setHaustiere((alt) => (alt.includes(art) ? alt.filter((h) => h !== art) : [...alt, art]));
  }

  // Wer ausführt, bekommt im dritten Schritt eine Zeitstufe. Betreute Personen
  // erzeugen Arbeit und übernehmen keine — sie werden nicht gefragt.
  const ausfuehrende = personen.filter((p) => p.art !== "betreut");

  async function anlegen() {
    setFehler(null);
    setLaeuft(true);
    try {
      const mitglieder: NeuesMitglied[] = personen.map((p, i) => {
        const m: NeuesMitglied = {
          name: p.name.trim(),
          // Index 0 ist die einrichtende Person. Ihre Rolle wird nicht aus dem
          // Formular abgeleitet, sondern gesetzt — siehe planner.Setup.
          rolle:
            i === 0 ? "planend" : p.art === "betreut" ? "betreut" : p.plant ? "planend" : "ausfuehrend",
        };
        if (p.art !== "erwachsen" && p.geburtsjahr) m.geburtsjahr = Number(p.geburtsjahr);
        m.zeit = p.art === "betreut" ? "keine" : p.zeit;
        return m;
      });

      const rumpf: NeuerHaushalt = {
        name: name.trim(),
        zimmer: Number(zimmer) || 3,
        baeder: Number(baeder) || 1,
        garten,
        auto,
        haustiere,
        mitglieder,
      };

      const haushalt = await postMitToken<Haushalt>("/api/haushalte", rumpf);
      // push statt window.location: Der React Compiler verbietet die
      // Zuweisung, und refresh sorgt dafür, dass die Server Components den
      // neuen Haushalt auch sehen.
      router.push(`/?haushalt=${encodeURIComponent(haushalt.id)}`);
      router.refresh();
    } catch (e) {
      setFehler(e instanceof Error ? e.message : "Das hat nicht funktioniert.");
      setLaeuft(false);
    }
  }

  const weiterGesperrt =
    (schritt === 1 && name.trim() === "") ||
    (schritt === 2 &&
      (personen.some((p) => p.name.trim() === "") ||
        personen.some((p, i) => i > 0 && p.art !== "erwachsen" && !p.geburtsjahr)));

  return (
    <div className="space-y-8">
      <div className="space-y-2">
        <p className="text-sm font-semibold text-muted">
          Schritt {schritt} von 3
          <span className="ml-2 font-normal text-subtle">
            {["Wohnen", "Wer gehört dazu", "Zeit"][schritt - 1]}
          </span>
        </p>
        <div className="flex gap-1.5" aria-hidden="true">
          {[1, 2, 3].map((i) => (
            <span
              key={i}
              className={`h-1 flex-1 rounded-full ${i <= schritt ? "bg-primary" : "bg-surface-3"}`}
            />
          ))}
        </div>
      </div>

      {schritt === 1 && (
        <section className="space-y-6">
          <Feld
            beschriftung="Wie soll der Haushalt heißen?"
            value={name}
            onChange={(e) => setName(e.target.value)}
            maxLength={80}
          />

          {/* Zwei Zahlen, die etwas bewirken: Sie filtern keine Aufgaben,
              sie skalieren die Dauern. Genau das unterschied sie von der
              Frage nach der Wohnform, die hier stand und nichts tat. */}
          <div className="flex flex-wrap gap-4">
            <Feld
              beschriftung="Wie viele Zimmer?"
              type="number"
              inputMode="numeric"
              min={1}
              max={15}
              value={zimmer}
              onChange={(e) => setZimmer(e.target.value)}
              className="w-28"
            />
            <Feld
              beschriftung="Wie viele Bäder?"
              type="number"
              inputMode="numeric"
              min={0}
              max={5}
              value={baeder}
              onChange={(e) => setBaeder(e.target.value)}
              className="w-28"
            />
          </div>
          <p className="text-xs leading-relaxed text-muted">
            Danach richtet sich, wie lange Putzaufgaben dauern — in fünf Zimmern
            länger als in zwei, und zwei Bäder sind doppelte Arbeit.
          </p>

          <div className="space-y-2">
            <p className="text-sm text-muted">Gibt es …</p>
            <div className="flex flex-wrap gap-2">
              <Marke an={garten} text="Garten" klick={() => setGarten(!garten)} />
              <Marke an={auto} text="Auto" klick={() => setAuto(!auto)} />
              <Marke an={haustiere.includes("hund")} text="Hund" klick={() => haustier("hund")} />
              <Marke an={haustiere.includes("katze")} text="Katze" klick={() => haustier("katze")} />
            </div>
            <p className="text-xs leading-relaxed text-muted">
              Jedes Ja bringt Aufgaben mit, jedes Nein spart sie. Ändern kannst
              du das später.
            </p>
          </div>
        </section>
      )}

      {schritt === 2 && (
        <section className="space-y-5">
          <p className="text-sm text-muted">Wer gehört zum Haushalt?</p>

          {personen.map((p, i) => (
            <div key={p.id} className="space-y-3 rounded-lg border border-line bg-surface p-4">
              <div className="flex items-center gap-3">
                <input
                  value={p.name}
                  onChange={(e) => aendern(p.id, { name: e.target.value })}
                  placeholder="Name"
                  maxLength={40}
                  className="min-w-0 flex-1 rounded-md border border-line bg-transparent px-3 py-2"
                />
                {i > 0 && (
                  <button
                    type="button"
                    onClick={() => setPersonen((alt) => alt.filter((x) => x.id !== p.id))}
                    className="text-sm text-muted underline"
                  >
                    entfernen
                  </button>
                )}
              </div>

              {/*
                Die erste Person richtet ein — sie plant, und daran gibt es
                nichts zu wählen. Vorher stand hier ein ausgegrautes Häkchen:
                Es zeigte die Regel an, statt sie durchzusetzen, und der
                Zustand dahinter konnte beim Umschalten der Art auf false
                kippen. Der Server wies das dann ab, das Formular wusste von
                nichts. Eine Regel, die man nur anzeigt, ist keine.
              */}
              {i === 0 ? (
                <p className="text-sm text-muted">
                  Du richtest ein und planst mit — du verteilst, lädst ein und
                  siehst die Bilanz.
                </p>
              ) : (
                <Auswahl
                  name={`Art von ${p.name || "dieser Person"}`}
                  wert={p.art}
                  auf={(art) => aendern(p.id, { art, plant: art === "erwachsen" && p.plant })}
                  optionen={[
                    { wert: "erwachsen", text: "Erwachsen" },
                    { wert: "kind", text: "Kind" },
                    { wert: "betreut", text: "Wird betreut" },
                  ]}
                />
              )}

              {i > 0 && p.art === "erwachsen" && (
                <label className="flex items-center gap-2 text-sm text-muted">
                  <input
                    type="checkbox"
                    checked={p.plant}
                    onChange={(e) => aendern(p.id, { plant: e.target.checked })}
                  />
                  plant mit — verteilt, lädt ein und sieht die Bilanz
                </label>
              )}

              {i > 0 && p.art !== "erwachsen" && (
                <label className="flex items-center gap-2 text-sm text-muted">
                  Geburtsjahr
                  <input
                    type="number"
                    inputMode="numeric"
                    min={1900}
                    max={JAHR}
                    value={p.geburtsjahr}
                    onChange={(e) => aendern(p.id, { geburtsjahr: e.target.value })}
                    className="w-24 rounded-md border border-line bg-transparent px-2 py-1"
                  />
                  <span className="text-xs">daran hängen die Aufgaben</span>
                </label>
              )}
            </div>
          ))}

          {personen.length < 12 && (
            <button
              type="button"
              onClick={() =>
                setPersonen((alt) => [
                  ...alt,
                  {
                    id: Math.max(...alt.map((p) => p.id)) + 1,
                    name: "",
                    art: "erwachsen",
                    plant: false,
                    geburtsjahr: "",
                    zeit: "mittel",
                  },
                ])
              }
              className="inline-flex min-h-11 items-center rounded-md border border-line-strong px-4 text-sm font-semibold text-muted transition-colors hover:border-primary hover:text-primary"
            >
              Person hinzufügen
            </button>
          )}
        </section>
      )}

      {schritt === 3 && (
        <section className="space-y-5">
          <p className="text-sm leading-relaxed text-muted">
            Wie viel Zeit bleibt im Alltag für den Haushalt? Grob reicht —
            niemand weiß, wie viele Minuten er dienstags hat. Die Verteilung
            richtet sich danach: Wer weniger Zeit hat, bekommt weniger.
          </p>

          {ausfuehrende.map((p) => (
            <div key={p.id} className="space-y-2 rounded-lg border border-line bg-surface p-4">
              <p className="font-medium">{p.name || "Ohne Namen"}</p>
              <Auswahl
                name={`Zeit von ${p.name || "dieser Person"}`}
                wert={p.zeit}
                auf={(zeit) => aendern(p.id, { zeit })}
                optionen={[
                  { wert: "wenig", text: "Wenig" },
                  { wert: "mittel", text: "Mittel" },
                  { wert: "viel", text: "Viel" },
                ]}
              />
            </div>
          ))}
        </section>
      )}

      {fehler && <p className="text-sm text-clay">{fehler}</p>}

      <div className="flex items-center justify-between gap-3 border-t border-line pt-6">
        <button
          type="button"
          onClick={() => setSchritt((s) => s - 1)}
          disabled={schritt === 1 || laeuft}
          className="text-sm text-muted underline disabled:invisible"
        >
          Zurück
        </button>

        {schritt < 3 ? (
          <button
            type="button"
            onClick={() => setSchritt((s) => s + 1)}
            disabled={weiterGesperrt}
            className="inline-flex min-h-12 items-center rounded-md bg-primary px-6 text-sm font-semibold text-on-primary transition-colors hover:bg-primary-hover disabled:opacity-45"
          >
            Weiter
          </button>
        ) : (
          <button
            type="button"
            onClick={anlegen}
            disabled={laeuft}
            className="inline-flex min-h-12 items-center rounded-md bg-primary px-6 text-sm font-semibold text-on-primary transition-colors hover:bg-primary-hover disabled:opacity-45"
          >
            {laeuft ? "Wird angelegt …" : "Haushalt anlegen"}
          </button>
        )}
      </div>
    </div>
  );
}
