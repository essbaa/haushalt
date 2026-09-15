"use client";

import { useRouter } from "next/navigation";
import { useState, useTransition } from "react";
import type { Mitglied } from "@/lib/api";
import { postMitToken, putMitToken } from "@/lib/browser-token";

const tage = ["Mo", "Di", "Mi", "Do", "Fr", "Sa", "So"];

/**
 * Wer an welchem Wochentag — das Wochenraster einer Aufgabe.
 *
 * Diese Aufgaben verteilt der Planer nicht. Er verteilt nach Kapazität, also
 * nach verfügbarer Zeit; wer ein Kind um 7:45 in die Kita bringen kann, hängt
 * dagegen am Arbeitsplan, und den kennt die App nicht. Ohne Absprache würde
 * sie mit voller Überzeugung den Falschen einteilen — deshalb wird hier
 * gefragt und nicht geraten (ADR-0016).
 *
 * Das Raster gilt, bis jemand es ändert. „Diese Woche wie letzte" ist damit
 * der Normalfall und braucht keinen Knopf. Eine einzelne Ausnahme — Dienstag
 * hat Amir einen Termin — ist „Wer macht das?" im Plan und betrifft nur den
 * einen Tag.
 */
export function AbspracheRaster({
  haushaltId,
  vorlageId,
  titel,
  mitglieder,
  raster,
  planend,
}: {
  haushaltId: string;
  vorlageId: string;
  titel: string;
  mitglieder: Mitglied[];
  raster: string[];
  planend: boolean;
}) {
  const router = useRouter();
  const [offen, setOffen] = useState(false);
  const [entwurf, setEntwurf] = useState<string[]>(raster);
  const [laeuft, setLaeuft] = useState<string | null>(null);
  const [gespeichert, setGespeichert] = useState(false);
  const [eingetragen, setEingetragen] = useState<number | null>(null);
  const [fehler, setFehler] = useState<string | null>(null);
  const [uebergang, starten] = useTransition();

  // Wer kann abgesprochen werden: alle, die überhaupt etwas übernehmen. Wer
  // nur Arbeit erzeugt (`betreut`), steht nicht zur Wahl.
  const kandidaten = mitglieder.filter((m) => m.rolle !== "betreut");
  const name = (id: string) => kandidaten.find((m) => m.id === id)?.name ?? "";

  const gesetzt = raster.filter((w) => w !== "").length;

  async function speichern() {
    setLaeuft("speichern");
    setFehler(null);
    try {
      await putMitToken<void>(
        `/api/haushalte/${encodeURIComponent(haushaltId)}/absprachen/${encodeURIComponent(vorlageId)}`,
        { wochentage: entwurf },
      );
      setOffen(false);
      setGespeichert(true);
      setEingetragen(null);
      starten(() => router.refresh());
    } catch (e) {
      setFehler(e instanceof Error ? e.message : "Das hat nicht funktioniert.");
    } finally {
      setLaeuft(null);
    }
  }

  // Die laufende Woche steht fest (ADR-0008), eine neue Absprache ändert sie
  // also nicht von selbst. Ohne diesen Weg speichert man ein Raster, sieht im
  // Plan nichts und erfährt nicht, warum — genau das ist beim ersten
  // Ausprobieren passiert.
  //
  // Bewusst nicht das allgemeine Neurechnen: Das verteilt die ganze Woche neu,
  // und wer am Mittwoch eine Kita-Absprache einträgt, bekäme nebenbei eine
  // neue Antwort darauf, wer am Freitag das Bad putzt.
  async function abHeute() {
    setLaeuft("woche");
    setFehler(null);
    try {
      const antwort = await postMitToken<{ eingetragen: number }>(
        `/api/haushalte/${encodeURIComponent(haushaltId)}/absprachen/${encodeURIComponent(vorlageId)}/ab-heute`,
      );
      setEingetragen(antwort.eingetragen);
      starten(() => router.refresh());
    } catch (e) {
      setFehler(e instanceof Error ? e.message : "Das hat nicht funktioniert.");
    } finally {
      setLaeuft(null);
    }
  }

  return (
    <div className="w-full space-y-2 rounded-md bg-surface-2 p-3">
      <p className="text-xs leading-relaxed text-muted text-pretty">
        <span className="font-semibold text-fg">Das sprecht ihr ab.</span> Wer
        das kann, hängt an euren Arbeitszeiten — die kennt die App nicht, und
        raten wäre hier schlimmer als fragen.
      </p>

      {/* Der Stand als Zeile, bevor irgendetwas aufgeklappt ist. Ein Raster,
          das man erst öffnen muss, um zu sehen, ob es leer ist, verschweigt
          genau die Auskunft, wegen der man hinsieht. */}
      {gesetzt === 0 ? (
        <p className="text-sm font-semibold text-clay">Noch nichts abgesprochen.</p>
      ) : (
        <ul className="flex flex-wrap gap-x-3 gap-y-1 text-sm">
          {raster.map((wer, i) => (
            <li key={tage[i]} className={wer === "" ? "text-subtle" : ""}>
              <span className="font-semibold">{tage[i]}</span>{" "}
              {wer === "" ? "—" : name(wer)}
            </li>
          ))}
        </ul>
      )}

      {planend && !offen && (
        <button
          type="button"
          onClick={() => {
            setEntwurf(raster);
            setOffen(true);
          }}
          className="inline-flex min-h-10 items-center rounded-md px-3 text-sm font-semibold text-primary transition-colors hover:bg-primary-soft"
        >
          {gesetzt === 0 ? "Absprechen" : "Absprache ändern"}
        </button>
      )}

      {/* Was das Speichern bewirkt hat — und was nicht. Ein „Übernommen"
          allein ließe offen, warum im Plan nichts steht, und das ist die
          erste Frage, die sich jeder stellt. */}
      {planend && gespeichert && !offen && (
        <div className="space-y-2 border-t border-line pt-2">
          {eingetragen === null ? (
            <>
              <p className="text-sm leading-relaxed text-pretty">
                Übernommen. Ab nächster Woche steht sie von selbst im Plan —{" "}
                <span className="font-semibold">diese Woche steht schon fest</span>{" "}
                und ändert sich nicht von allein.
              </p>
              <div className="flex flex-wrap items-center gap-x-3 gap-y-2">
                <button
                  type="button"
                  onClick={abHeute}
                  disabled={laeuft !== null || uebergang}
                  className="inline-flex min-h-11 items-center rounded-md border border-line-strong px-4 text-sm font-semibold transition-colors hover:border-primary hover:text-primary disabled:opacity-45"
                >
                  {laeuft === "woche" ? "Einen Moment …" : "Ab heute eintragen"}
                </button>
                <p className="text-xs text-muted text-pretty">
                  Trägt die restlichen Tage dieser Woche ein. Nur diese
                  Aufgabe — am übrigen Plan ändert sich nichts.
                </p>
              </div>
            </>
          ) : eingetragen > 0 ? (
            <p className="text-sm text-primary">
              {eingetragen === 1
                ? "Ein Termin steht jetzt im Plan."
                : `${eingetragen} Termine stehen jetzt im Plan.`}
            </p>
          ) : (
            // Null ist eine Antwort und kein Fehlschlag. Ein „fertig" ohne
            // sichtbare Wirkung wäre wieder genau die Quittung, die diese
            // Zeilen vermeiden sollen.
            <p className="text-sm leading-relaxed text-muted text-pretty">
              Für den Rest dieser Woche gab es nichts einzutragen — an den
              verbleibenden Tagen steht die Aufgabe schon, oder sie ist nicht
              mehr fällig. Ab nächster Woche gilt das neue Raster.
            </p>
          )}
        </div>
      )}

      {planend && offen && (
        <div className="space-y-3 pt-1">
          <fieldset className="space-y-1.5">
            <legend className="sr-only">Wochenraster für {titel}</legend>
            {tage.map((tag, i) => (
              <div key={tag} className="flex items-center gap-2">
                <span className="w-8 shrink-0 text-sm font-semibold text-muted">{tag}</span>
                <select
                  value={entwurf[i] ?? ""}
                  onChange={(e) => {
                    const naechster = [...entwurf];
                    naechster[i] = e.target.value;
                    setEntwurf(naechster);
                  }}
                  aria-label={`${tag}: wer macht „${titel}"?`}
                  className="min-h-11 flex-1 rounded-md border border-line-strong bg-surface px-2 text-sm"
                >
                  <option value="">— niemand —</option>
                  {kandidaten.map((m) => (
                    <option key={m.id} value={m.id}>
                      {m.name}
                    </option>
                  ))}
                </select>
              </div>
            ))}
          </fieldset>

          <div className="flex flex-wrap gap-2">
            <button
              type="button"
              onClick={speichern}
              disabled={laeuft !== null || uebergang}
              className="inline-flex min-h-11 items-center rounded-md bg-primary px-5 text-sm font-semibold text-on-primary transition-colors hover:bg-primary-hover disabled:opacity-45"
            >
              {laeuft === "speichern" ? "Einen Moment …" : "Übernehmen"}
            </button>
            <button
              type="button"
              onClick={() => setOffen(false)}
              className="inline-flex min-h-11 items-center rounded-md px-3 text-sm font-semibold text-muted hover:text-fg"
            >
              Abbrechen
            </button>
          </div>
        </div>
      )}

      {fehler && (
        <p role="alert" className="rounded-md border border-danger/40 px-3 py-2 text-sm text-danger">
          {fehler}
        </p>
      )}
    </div>
  );
}
