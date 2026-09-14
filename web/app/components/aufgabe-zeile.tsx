"use client";

import { useRouter } from "next/navigation";
import { useId, useState } from "react";
import { AufgabeAktionen } from "@/app/components/aufgabe-aktionen";
import { Haken, Kreis } from "@/app/components/icons";
import { Zeichen } from "@/app/components/ui";
import { Wischen } from "@/app/components/wischen";
import type { Aufgabe } from "@/lib/api";
import { patchMitToken, postMitToken } from "@/lib/browser-token";

/**
 * Eine Aufgabe im Wochenplan.
 *
 * Die Zeile ist kompakt und klappt beim Antippen auf. Vorher stand unter jeder
 * eine Leiste mit bis zu vier Knöpfen — bei zwei Dutzend Aufgaben macht das
 * eine Seite, durch die man scrollt, statt sie zu lesen. Sichtbar bleibt das
 * Häkchen rechts: die häufigste Handlung des Tages und zugleich die
 * Alternative zur Wischgeste, die WCAG 2.2 verlangt.
 *
 * Nach rechts wischen hakt ab, nach links schaltet die Vorlage dauerhaft aus —
 * „brauchen wir nicht". Der Impuls dazu entsteht hier, beim Lesen des Plans,
 * nicht auf einer Einstellungsseite; ein Umweg dorthin würde nicht gegangen.
 *
 * Weil Ausschalten dauerhaft ist, bleibt die Zeile danach als „Rückgängig"
 * stehen. Ein Bestätigungsdialog wäre der übliche Weg und der falsche: Er
 * unterbricht genau die Geste, die schnell sein soll.
 */
export function AufgabeZeile({
  aufgabe: a,
  namen,
  kandidaten,
  ich,
  planend,
  haushaltId,
}: {
  aufgabe: Aufgabe;
  namen: Record<string, string>;
  kandidaten: { id: string; name: string }[];
  ich: string;
  planend: boolean;
  haushaltId: string;
}) {
  const router = useRouter();
  const [aus, setAus] = useState(false);
  const [aufgeklappt, setAufgeklappt] = useState(false);
  const feld = useId();
  const [laeuft, setLaeuft] = useState(false);
  const [fehler, setFehler] = useState<string | null>(null);

  const meine = ich !== "" && a.zustaendig === ich;
  const offen = a.zustaendig === "";
  const darf = a.id !== undefined && (meine || planend);
  const name = offen ? "Offen" : (namen[a.zustaendig] ?? a.zustaendig);

  async function vorlage(aktiv: boolean) {
    setLaeuft(true);
    setFehler(null);
    try {
      await patchMitToken<void>(
        `/api/haushalte/${encodeURIComponent(haushaltId)}/vorlagen/${encodeURIComponent(a.vorlage_id)}`,
        { aktiv },
      );
      setAus(!aktiv);
      if (aktiv) router.refresh();
    } catch (e) {
      setFehler(e instanceof Error ? e.message : "Das hat nicht funktioniert.");
      setAus(false);
    } finally {
      setLaeuft(false);
    }
  }

  async function abhaken() {
    if (!a.id) return;
    try {
      await postMitToken<void>(`/api/aufgaben/${encodeURIComponent(a.id)}/erledigt`, {
        erledigt: !a.erledigt,
      });
      router.refresh();
    } catch (e) {
      setFehler(e instanceof Error ? e.message : "Das hat nicht funktioniert.");
    }
  }

  // Ausgeschaltet: die Zeile bleibt als Rückweg stehen, bis die Seite neu
  // geladen wird. Wer sich vertan hat, muss nicht suchen gehen.
  if (aus) {
    return (
      <div className="flex flex-wrap items-center gap-3 rounded-lg border border-line bg-surface-2 px-3 py-3">
        <p className="min-w-0 flex-1 text-sm text-muted">
          <span className="font-semibold text-fg">{a.titel}</span> kommt nicht mehr vor.
        </p>
        <button
          type="button"
          onClick={() => vorlage(true)}
          disabled={laeuft}
          className="inline-flex min-h-10 items-center rounded-full border border-line-strong px-3.5 text-sm font-semibold transition-colors hover:border-primary hover:text-primary disabled:opacity-45"
        >
          Rückgängig
        </button>
      </div>
    );
  }

  // Titel, Person und die Kennzahlen — und, wo es Aktionen gibt, der
  // Aufklapper. Ein Knopf um den ganzen Textblock statt eines Pfeils am Rand:
  // Die Fläche, die man ohnehin ansieht, ist die, die man antippt.
  const kopftext = (
    <>
      <div className="flex flex-wrap items-baseline gap-x-2">
        <span className={`font-semibold ${a.erledigt ? "line-through" : ""}`}>{a.titel}</span>
        <span className={`text-sm ${meine ? "font-semibold text-primary" : "text-muted"}`}>
          {meine ? "du" : name}
        </span>
      </div>

      <div className="mt-1 flex flex-wrap items-center gap-x-2.5 gap-y-1 text-xs text-subtle">
        <span className="tabular-nums">{a.dauer_min} min</span>
        {a.kopflast >= 2 && <span className="font-semibold text-clay">Kopflast {a.kopflast}</span>}
        <span>{begruendung(a, namen)}</span>
      </div>
    </>
  );

  const kopfzeile = darf ? (
    <button
      type="button"
      onClick={() => setAufgeklappt((o) => !o)}
      aria-expanded={aufgeklappt}
      aria-controls={feld}
      className="block w-full min-w-0 cursor-pointer text-left"
    >
      {kopftext}
    </button>
  ) : (
    <div className="min-w-0">{kopftext}</div>
  );

  const inhalt = (
    <div
      className={`flex gap-3 rounded-lg px-3 py-2.5 ${
        a.erledigt
          ? "border border-transparent opacity-50"
          : meine
            ? "border border-primary/35 bg-primary-soft/60"
            : "border border-transparent"
      }`}
    >
      <Zeichen name={offen ? "?" : name} eigen={meine} />

      <div className="min-w-0 flex-1">
        {kopfzeile}

        {/* Immer im Baum, nur versteckt: aria-controls muss auf etwas zeigen,
            das es gibt — und eine Meldung wie „Mia macht das jetzt" überlebt
            so das Zuklappen. */}
        {darf && (
          <div id={feld} hidden={!aufgeklappt}>
            <AufgabeAktionen
              aufgabeId={a.id!}
              erledigt={a.erledigt ?? false}
              abgebbar={meine}
              zustaendig={a.zustaendig}
              kandidaten={kandidaten}
              verteilbar={a.begruendung.code !== "eigene_aufgabe"}
              abschaltbar={planend}
              abschalten={() => vorlage(false)}
            />
          </div>
        )}

        {fehler && (
          <p role="alert" className="text-sm text-danger">
            {fehler}
          </p>
        )}
      </div>

      {darf && (
        <button
          type="button"
          onClick={abhaken}
          aria-pressed={a.erledigt ?? false}
          aria-label={a.erledigt ? `${a.titel} wieder öffnen` : `${a.titel} abhaken`}
          className={`-mr-1 flex size-11 shrink-0 items-center justify-center self-start rounded-full transition-colors ${
            a.erledigt
              ? // Erledigtes wird leise. Ein gefüllter Knopf auf der Zeile, die
                // niemanden mehr interessiert, zieht den Blick genau dorthin,
                // wo nichts mehr zu tun ist.
                "text-primary hover:bg-surface-2"
              : "text-subtle hover:bg-surface-2 hover:text-fg"
          }`}
        >
          {a.erledigt ? <Haken className="size-5" /> : <Kreis className="size-5" />}
        </button>
      )}
    </div>
  );

  if (!darf) return inhalt;

  return (
    <Wischen
      rechts={{ text: a.erledigt ? "Wieder öffnen" : "Erledigt", tun: abhaken }}
      links={planend ? { text: "Brauchen wir nicht", tun: () => vorlage(false) } : undefined}
    >
      {inhalt}
    </Wischen>
  );
}

function begruendung(a: Aufgabe, namen: Record<string, string>): string {
  const zuletzt = a.begruendung.zuletzt_bei;
  switch (a.begruendung.code) {
    case "rotation":
      return zuletzt ? `zuletzt bei ${namen[zuletzt] ?? zuletzt}` : "Rotation";
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
    case "von_hand":
      return zuletzt ? `von Hand, vorher ${namen[zuletzt] ?? zuletzt}` : "von Hand verteilt";
    default:
      return "";
  }
}
