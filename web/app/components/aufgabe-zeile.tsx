"use client";

import { useRouter } from "next/navigation";
import { useId, useState, useTransition } from "react";
import { AufgabeAktionen } from "@/app/components/aufgabe-aktionen";
import { Haken, Kreis, Zurueck } from "@/app/components/icons";
import { Zeichen } from "@/app/components/ui";
import { Wischen } from "@/app/components/wischen";
import type { Aufgabe, Mitglied } from "@/lib/api";
import { begruendung } from "@/lib/begruendung";
import { farbklasse } from "@/lib/personen";
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
  mitglieder,
  kandidaten,
  ich,
  planend,
  haushaltId,
}: {
  aufgabe: Aufgabe;
  namen: Record<string, string>;
  /** Für die Farbe: Sie hängt an der Stelle im Haushalt. */
  mitglieder: Mitglied[];
  kandidaten: { id: string; name: string }[];
  ich: string;
  planend: boolean;
  haushaltId: string;
}) {
  const router = useRouter();
  const [aus, setAus] = useState(false);
  const [aufgeklappt, setAufgeklappt] = useState(false);
  // Das Abgeben-Formular wird von hier gesteuert: Sein Knopf sitzt neben dem
  // Häkchen und muss beides können — aufklappen und gleich fragen.
  const [fragt, setFragt] = useState(false);
  const feld = useId();
  const [laeuft, setLaeuft] = useState(false);
  const [fehler, setFehler] = useState<string | null>(null);
  // router.refresh() holt die Server Components neu und ist nicht abgewartet:
  // Ohne useTransition steht der Knopf wieder bereit, während die neuen Daten
  // noch unterwegs sind — die Zeile sieht fertig aus und ändert sich eine
  // Sekunde später doch noch. `uebergang` hält den Zustand, bis der Server
  // geantwortet hat.
  const [uebergang, starten] = useTransition();
  const beschaeftigt = laeuft || uebergang;

  const meine = ich !== "" && a.zustaendig === ich;
  const offen = a.zustaendig === "";
  const darf = a.id !== undefined && (meine || planend);
  // Abgeben kann nur, wer sie hat. Für planende Personen an fremden Aufgaben
  // ist „Wer macht das?" der richtige Weg — sie wählen, statt zurückzugeben.
  const abgebbar = meine;
  const name = offen ? "Offen" : (namen[a.zustaendig] ?? a.zustaendig);
  const farbe = farbklasse(mitglieder, a.zustaendig);
  const grund = begruendung(a, namen);

  async function vorlage(aktiv: boolean) {
    setLaeuft(true);
    setFehler(null);
    try {
      await patchMitToken<void>(
        `/api/haushalte/${encodeURIComponent(haushaltId)}/vorlagen/${encodeURIComponent(a.vorlage_id)}`,
        { aktiv },
      );
      setAus(!aktiv);
      if (aktiv) starten(() => router.refresh());
    } catch (e) {
      setFehler(e instanceof Error ? e.message : "Das hat nicht funktioniert.");
      setAus(false);
    } finally {
      setLaeuft(false);
    }
  }

  // Streichen statt Abschalten: Das hier gilt nur für DIESEN Termin.
  //
  // Das Etikett hieß zuerst „Diese Woche nicht" und behauptete damit mehr,
  // als die Sache tut: Gestrichen wird eine Aufgabe an einem Tag, nicht die
  // Vorlage für die Woche. Bei „Safiya zur Kita bringen", das fünfmal
  // vorkommt, wäre das die Ansage gewesen, sie gehe die ganze Woche nicht
  // hin.
  //
  // Es ist die Geste geworden und „Brauchen wir nicht" nicht mehr — die
  // unumkehrbare Handlung sollte nicht die sein, die man aus Versehen macht.
  // Ein Wisch, der eine Vorlage für alle künftigen Wochen abschaltet, fällt
  // erst nächste Woche auf, wenn sie fehlt.
  async function streichen() {
    if (!a.id) return;
    try {
      await postMitToken<void>(`/api/aufgaben/${encodeURIComponent(a.id)}/streichen`);
      starten(() => router.refresh());
    } catch (e) {
      setFehler(e instanceof Error ? e.message : "Das hat nicht funktioniert.");
    }
  }

  async function abhaken() {
    if (!a.id) return;
    try {
      await postMitToken<void>(`/api/aufgaben/${encodeURIComponent(a.id)}/erledigt`, {
        erledigt: !a.erledigt,
      });
      starten(() => router.refresh());
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
          disabled={beschaeftigt}
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
        <span className={`leading-snug font-semibold ${a.erledigt ? "line-through" : ""}`}>
          {a.titel}
        </span>
        <span className={`text-sm ${meine ? "font-semibold text-[var(--person)]" : "text-muted"}`}>
          {meine ? "du" : name}
        </span>
      </div>

      {/* Die Kennzahlen als eine Zeile mit Trennpunkten statt drei Blöcken
          mit Lücken. Lücke heißt „hier ist Platz", Punkt heißt „das gehört
          zusammen, ist aber nicht dasselbe" — und genau das trifft zu.
          Leere Teile fallen raus, damit nie ein Punkt am Ende steht. */}
      <div className="mt-0.5 flex flex-wrap items-center gap-x-1.5 gap-y-0.5 text-xs text-subtle">
        <span className="tabular-nums">{a.dauer_min} min</span>
        {a.kopflast >= 2 && (
          <>
            <span aria-hidden="true">·</span>
            {/* Die einzige Zahl, die hier farbig bleibt. Kopflast ist der
                Gedanke, um den es in diesem Produkt geht — sie grau unter
                Minuten und Begründung zu mischen hieße, sie zu verstecken. */}
            <span className="font-semibold text-clay">Kopflast {a.kopflast}</span>
          </>
        )}
        {grund !== "" && (
          <>
            <span aria-hidden="true">·</span>
            <span>{grund}</span>
          </>
        )}
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
      className={`${farbe} flex gap-3 rounded-lg border px-3 py-3 transition-colors ${
        a.erledigt ? "border-transparent opacity-45" : meine ? "zeile-eigen" : "border-transparent"
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
              zustaendig={a.zustaendig}
              kandidaten={kandidaten}
              verteilbar={a.begruendung.code !== "eigene_aufgabe"}
              fragt={fragt}
              setFragt={setFragt}
              streichen={streichen}
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

      {/* Abgeben als Zeichen, direkt neben dem Häkchen.
          Vorher stand es nur aufgeklappt, und damit kostete die zweite
          Handlung des Tages drei Tipper: Zeile antippen, lesen, wählen.
          Sichtbar sind jetzt die beiden, die man wirklich braucht — ich habe
          es getan, und ich kann es nicht. Unter dem Aufklapper bleibt, was
          eine Entscheidung ist: nur diesmal nicht, gar nicht mehr, oder
          jemand anders.

          Nur an eigenen Aufgaben und nur, solange sie offen sind: Ein
          Zeichen, das für die halbe Liste nichts tut, ist schlimmer als
          keins. */}
      {darf && (
        <div className="-mr-1 flex shrink-0 self-start">
          {abgebbar && !a.erledigt && (
            <button
              type="button"
              onClick={() => {
                setAufgeklappt(true);
                setFragt(true);
              }}
              disabled={beschaeftigt}
              aria-label={`${a.titel} abgeben`}
              title="Abgeben"
              className="flex size-11 shrink-0 items-center justify-center rounded-full text-subtle transition-colors hover:bg-surface-2 hover:text-fg disabled:opacity-45"
            >
              <Zurueck className="size-5" />
            </button>
          )}

          <button
            type="button"
            onClick={abhaken}
            disabled={beschaeftigt}
            aria-pressed={a.erledigt ?? false}
            aria-label={a.erledigt ? `${a.titel} wieder öffnen` : `${a.titel} abhaken`}
            className={`flex size-11 shrink-0 items-center justify-center rounded-full transition-colors disabled:opacity-45 ${
              a.erledigt
                ? // Erledigtes wird leise. Ein gefüllter Knopf auf der Zeile,
                  // die niemanden mehr interessiert, zieht den Blick genau
                  // dorthin, wo nichts mehr zu tun ist.
                  "text-primary hover:bg-surface-2"
                : "text-subtle hover:bg-surface-2 hover:text-fg"
            }`}
          >
            {a.erledigt ? <Haken className="size-5" /> : <Kreis className="size-5" />}
          </button>
        </div>
      )}
    </div>
  );

  if (!darf) return inhalt;

  return (
    <Wischen
      rechts={{ text: a.erledigt ? "Wieder öffnen" : "Erledigt", tun: abhaken }}
      links={{ text: "Diesmal nicht", tun: streichen }}
    >
      {inhalt}
    </Wischen>
  );
}
