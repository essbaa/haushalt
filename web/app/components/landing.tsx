import Link from "next/link";
import { Fahne, Haken, Kreis, Personen, Uhr, Zurueck } from "@/app/components/icons";
import { LandingKopf } from "@/app/components/landing-kopf";

/**
 * Die Seite für alle, die noch nicht angemeldet sind.
 *
 * Vorher standen hier zwei fremde Familien ohne Erklärung — wer die Adresse
 * zum ersten Mal öffnete, sah den Wochenplan von „Familie A" und musste sich
 * selbst zusammenreimen, was das soll. Aus dem stärksten Argument des
 * Produkts (ein echter Plan, sofort, ohne Konto) war ein Rätsel geworden,
 * weil niemand es eingeführt hat. Die Beispielhaushalte sind deshalb nicht
 * weg, sondern **einen Satz weiter**, unter /demo.
 *
 * Zur Gestaltung, nach dem ersten Entwurf: Der sah tot aus, und die Ursache
 * war nicht die Farbwahl, sondern die **Fläche**. Text auf Grund, dazwischen
 * dünne Linien — im Dunkelmodus ist der Grund fast schwarz, und warme Töne in
 * Akzenten wärmen nichts. Jetzt tragen zwei weiche Verlaufsflächen den Kopf
 * der Seite (`.warm-grund`), die Abschnitte wechseln zwischen Grund und
 * Fläche, und die Vorschau liegt als Karte mit Tiefe darüber statt als
 * weiterer Rahmen daneben.
 *
 * Alle Ansichten sind **nachgebaut**, nicht abfotografiert — dieselben
 * Tokens, dieselbe Schrift, dieselben Abstände wie in der App. Das ist scharf
 * auf jedem Bildschirm und im Dunkelmodus richtig. Es ist trotzdem eine
 * Darstellung und kein Beweis, und genau deshalb steht der Weg zum echten
 * Plan gleich daneben: Wer es nachprüfen will, klickt.
 *
 * Server Component — kein Zustand, keine Bewegung. Eine Seite, die erklärt,
 * was eine App gegen Mental Load tut, darf nicht selbst zappeln.
 */
export function Landing() {
  return (
    <>
      <LandingKopf />
      <main>
        <Hero />
        <div className="mx-auto w-full max-w-5xl px-5">
          <Problem />
          <SoGehtEs />
          <Absprache />
          <WasEsNichtTut />
          <Schluss />
        </div>
      </main>
      <Fuss />
    </>
  );
}

/* ------------------------------------------------------------------ Hero */

function Hero() {
  return (
    <section className="warm-grund koerner overflow-hidden pb-14 sm:pb-20">
      <div className="mx-auto grid w-full max-w-5xl gap-12 px-5 pt-6 sm:pt-10 lg:grid-cols-[1.05fr_1fr] lg:items-center lg:gap-14">
        <div>
          <p className="mb-5 inline-flex items-center gap-2 rounded-full border border-clay/30 bg-clay-soft px-3 py-1.5 text-xs font-semibold text-clay">
            <span className="size-1.5 rounded-full bg-clay" aria-hidden="true" />
            In Erprobung — Zugang nur mit Code
          </p>

          <h1 className="text-4xl leading-[1.08] font-extrabold tracking-tight text-balance sm:text-[3.25rem]">
            Der Haushalt verteilt sich nicht von selbst.
            <span className="mt-1 block text-primary">Die Arbeit daran auch nicht.</span>
          </h1>

          <p className="mt-6 max-w-[34rem] text-lg leading-relaxed text-muted text-pretty sm:text-xl">
            Ein Wochenplan für berufstätige Eltern, der nicht nur Zeit verteilt, sondern auch das{" "}
            <strong className="font-semibold text-fg">Daran-Denken</strong>. Wer die Termine im Kopf
            hat, trägt mehr als die Uhr zeigt — diese App rechnet damit.
          </p>

          <div className="mt-9 flex flex-wrap gap-3">
            <Link
              href="/demo"
              className="karte-tief inline-flex min-h-12 items-center rounded-full bg-primary px-7 text-sm font-bold text-on-primary transition-colors hover:bg-primary-hover"
            >
              Einen echten Plan ansehen
            </Link>
            <Link
              href="/anmelden"
              className="inline-flex min-h-12 items-center rounded-full border border-line-strong bg-surface px-6 text-sm font-semibold transition-colors hover:border-primary hover:text-primary"
            >
              Ich habe einen Code
            </Link>
          </div>

          <p className="mt-4 text-xs leading-relaxed text-subtle">
            Ansehen geht ohne Konto. Für einen eigenen Haushalt brauchst du einen Zugangscode.
          </p>
        </div>

        <WochenVorschau />
      </div>
    </section>
  );
}

/* -------------------------------------------------------------- Problem */

function Problem() {
  return (
    <section className="mt-16 sm:mt-24">
      <h2 className="max-w-2xl text-2xl font-extrabold tracking-tight text-balance sm:text-[2rem] sm:leading-tight">
        Die Hälfte der Arbeit steht auf keiner Liste
      </h2>

      <div className="mt-8 grid gap-4 sm:grid-cols-2 sm:gap-6">
        <div className="rounded-2xl border border-line bg-surface p-6">
          <p className="text-xs font-bold tracking-[0.08em] text-subtle uppercase">
            Was jeder sieht
          </p>
          <p className="mt-3 text-xl leading-snug font-bold text-pretty">
            Vierzig Minuten Staubsaugen.
          </p>
          <p className="mt-2.5 text-sm leading-relaxed text-muted text-pretty">
            Dauer lässt sich messen, aufteilen und abhaken. Deshalb landet sie in jedem
            Haushaltsplan — und deshalb sieht jeder Haushaltsplan gerechter aus, als er ist.
          </p>
        </div>

        <div className="rounded-2xl border border-clay/35 bg-clay-soft p-6">
          <p className="text-xs font-bold tracking-[0.08em] text-clay uppercase">
            Was niemand zählt
          </p>
          <p className="mt-3 text-xl leading-snug font-bold text-pretty">
            Daran denken, dass der Zahnarzttermin gemacht werden muss.
          </p>
          <p className="mt-2.5 text-sm leading-relaxed text-muted text-pretty">
            Acht Minuten am Telefon — und drei Wochen im Kopf. Diese Last heißt{" "}
            <strong className="font-semibold text-fg">Kopflast</strong>, sie liegt fast immer bei
            derselben Person, und sie ist der Grund, warum sich „halbe-halbe" nicht halbe-halbe
            anfühlt.
          </p>
        </div>
      </div>
    </section>
  );
}

/* ------------------------------------------------------------ So geht es */

function SoGehtEs() {
  const schritte = [
    {
      nr: "1",
      titel: "Drei Fragen, keine dreißig",
      text: "Wer wohnt hier, wie viel Zeit hat jeder, wie alt sind die Kinder. Den Rest rät die App — und alles Geratene steht sichtbar da und lässt sich ändern.",
    },
    {
      nr: "2",
      titel: "Montags steht der Plan",
      text: "Jede Aufgabe bei einer Person, an einem Tag, mit einem Grund daneben. Wer eine Woche später wieder hineinsieht, findet denselben Plan vor — er ändert sich nicht heimlich.",
    },
    {
      nr: "3",
      titel: "Widersprechen ist eingebaut",
      text: "Abhaken, abgeben, diesmal nicht, jemand anders. Jede Korrektur sticht die Annahmen der App — und wird gezählt, damit sie beim nächsten Mal weniger daneben liegt.",
    },
  ];

  return (
    <section className="warm-grund mt-16 overflow-hidden rounded-3xl border border-line bg-surface px-6 py-12 sm:mt-24 sm:px-10 sm:py-14">
      <h2 className="text-2xl font-extrabold tracking-tight text-balance sm:text-[2rem] sm:leading-tight">
        Wie es läuft
      </h2>

      <ol className="mt-9 grid gap-8 sm:grid-cols-3 sm:gap-6">
        {schritte.map((s) => (
          <li key={s.nr}>
            <span className="inline-flex size-10 items-center justify-center rounded-full bg-primary text-sm font-extrabold text-on-primary">
              {s.nr}
            </span>
            <p className="mt-4 text-lg leading-snug font-bold text-pretty">{s.titel}</p>
            <p className="mt-2 text-sm leading-relaxed text-muted text-pretty">{s.text}</p>
          </li>
        ))}
      </ol>

      <div className="mt-12 grid gap-8 border-t border-line pt-10 lg:grid-cols-2 lg:items-center">
        <Bilanz />
        <div>
          <h3 className="text-xl font-bold tracking-tight text-balance sm:text-2xl">
            Fairness misst sich an der eigenen Zeit
          </h3>
          <p className="mt-3 max-w-prose text-sm leading-relaxed text-muted text-pretty">
            Nicht in Minuten, sondern in Anteilen: Wer zehn Stunden in der Woche hat, soll nicht
            dasselbe tragen wie jemand mit dreißig. Dieselbe absolute Last buchte eine
            Dreizehnjährige zu 93 Prozent aus, während ihre Mutter bei 50 stand — auf dem Papier
            ausgeglichen, im Alltag unfair.
          </p>
          <p className="mt-3 max-w-prose text-sm leading-relaxed text-muted text-pretty">
            Die Kopflast zählt dabei als eigener Posten mit. Ein Planer, der nur Minuten addiert,
            gibt der Person, die ohnehin alle Termine koordiniert, obendrein das Bad.
          </p>
        </div>
      </div>
    </section>
  );
}

/* ------------------------------------------------------------ Absprache */

function Absprache() {
  return (
    <section className="mt-16 grid gap-10 sm:mt-24 lg:grid-cols-2 lg:items-center lg:gap-14">
      <div>
        <h2 className="text-2xl font-extrabold tracking-tight text-balance sm:text-[2rem] sm:leading-tight">
          Und wo sie nichts weiß, fragt sie
        </h2>
        <p className="mt-4 max-w-prose leading-relaxed text-muted text-pretty">
          Wer das Kind um 7:45 in die Kita bringen kann, hängt am Arbeitsplan — nicht daran, wer an
          dem Tag Zeit übrig hat. Die App kennt euren Arbeitsplan nicht. Also verteilt sie diese
          Aufgaben nicht, sondern nimmt eure Absprache entgegen und trägt sie ein.
        </p>
        <p className="mt-3 max-w-prose text-sm leading-relaxed text-muted text-pretty">
          Sie zählt trotzdem voll in die Bilanz. Wer jeden Morgen fährt, bekommt abends weniger —
          und das ist der ganze Punkt.
        </p>
      </div>

      <AbspracheVorschau />
    </section>
  );
}

/* ------------------------------------------------------- Was es nicht tut */

function WasEsNichtTut() {
  const punkte = [
    {
      Zeichen: Uhr,
      titel: "Sie erinnert an nichts",
      text: "Noch nicht. Der Plan steht, und wer nicht hineinsieht, erfährt es nicht. Benachrichtigungen kommen erst, wenn klar ist, wann sie kommen müssten.",
    },
    {
      Zeichen: Fahne,
      titel: "Sie ist kein Kalender",
      text: "Termine trägst du als Anlässe ein, damit die Vorbereitung rechtzeitig auftaucht. Den Familienkalender ersetzt sie nicht.",
    },
    {
      Zeichen: Personen,
      titel: "Sie vergibt keine Punkte",
      text: "Kein Wettbewerb, keine Abzeichen, keine Streaks. Wer Hausarbeit zum Spiel macht, macht sie nicht gerechter.",
    },
  ];

  return (
    <section className="mt-16 sm:mt-24">
      <h2 className="text-2xl font-extrabold tracking-tight text-balance sm:text-[2rem] sm:leading-tight">
        Was sie nicht tut
      </h2>
      <p className="mt-3 max-w-prose leading-relaxed text-muted text-pretty">
        Damit niemand darauf wartet:
      </p>

      <ul className="mt-8 grid gap-4 sm:grid-cols-3 sm:gap-6">
        {punkte.map((p) => {
          const Zeichen = p.Zeichen;
          return (
            <li key={p.titel} className="rounded-2xl border border-line bg-surface p-6">
              <span className="inline-flex size-10 items-center justify-center rounded-xl bg-surface-2">
                <Zeichen className="size-5 text-muted" />
              </span>
              <p className="mt-4 font-bold text-pretty">{p.titel}</p>
              <p className="mt-2 text-sm leading-relaxed text-muted text-pretty">{p.text}</p>
            </li>
          );
        })}
      </ul>
    </section>
  );
}

/* ------------------------------------------------------------- Abschluss */

function Schluss() {
  return (
    <section className="warm-grund koerner mt-16 overflow-hidden rounded-3xl border border-line px-6 py-12 text-center sm:mt-24 sm:px-10 sm:py-16">
      <h2 className="mx-auto max-w-xl text-2xl font-extrabold tracking-tight text-balance sm:text-[2rem] sm:leading-tight">
        Erst ansehen, dann entscheiden
      </h2>
      <p className="mx-auto mt-4 max-w-prose leading-relaxed text-muted text-pretty">
        Der Beispielhaushalt ist ein echter Plan aus derselben Rechnung — gerechnet, nicht gestellt.
        Für einen eigenen Haushalt brauchst du einen Zugangscode: Die App ist in Erprobung, und in
        dieser Datenbank stehen Namen und Alter von Kindern.
      </p>

      <div className="mt-8 flex flex-wrap justify-center gap-3">
        <Link
          href="/demo"
          className="karte-tief inline-flex min-h-12 items-center rounded-full bg-primary px-7 text-sm font-bold text-on-primary transition-colors hover:bg-primary-hover"
        >
          Beispielhaushalt ansehen
        </Link>
        <Link
          href="/anmelden"
          className="inline-flex min-h-12 items-center rounded-full border border-line-strong bg-surface px-6 text-sm font-semibold transition-colors hover:border-primary hover:text-primary"
        >
          Konto anlegen
        </Link>
      </div>
    </section>
  );
}

function Fuss() {
  return (
    <footer className="mx-auto w-full max-w-5xl px-5 py-10">
      <p className="text-xs leading-relaxed text-subtle text-pretty">
        Haushalt — ein Wochenplan gegen Mental Load. Die Ansichten auf dieser Seite sind nachgebaut;
        die Zahlen im Beispielhaushalt sind gerechnet.
      </p>
    </footer>
  );
}

/* ================================================================ Ansichten
 *
 * Nachgebaut aus denselben Bausteinen wie die App. Bewusst kein Bildschirm-
 * foto: Ein Foto ist auf großen Bildschirmen unscharf, steht im falschen
 * Farbmodus da und veraltet bei der nächsten Änderung.
 *
 * Die Zahlen sind erfunden und die Namen auch. Wer echte sehen will, klickt
 * auf „Einen echten Plan ansehen" — deshalb steht der Knopf daneben und nicht
 * irgendwo unten.
 */

const beispiel = [
  {
    tag: "Mo",
    nr: "14",
    heute: true,
    aufgaben: [
      {
        titel: "Zur Kita bringen",
        wer: "Anna",
        farbe: "person-2",
        meta: "25 min · so abgesprochen",
        eigen: true,
        fertig: false,
      },
      {
        titel: "Abendessen kochen",
        wer: "Ben",
        farbe: "person-4",
        meta: "40 min · zuletzt bei Anna",
        eigen: false,
        fertig: true,
      },
    ],
  },
  {
    tag: "Di",
    nr: "15",
    heute: false,
    aufgaben: [
      {
        titel: "Zahnkontrolle buchen",
        wer: "Ben",
        farbe: "person-4",
        meta: "8 min · Kopflast 3",
        eigen: false,
        fertig: false,
      },
      {
        titel: "Bad putzen",
        wer: "Mia",
        farbe: "person-6",
        meta: "35 min · zum Ausgleich",
        eigen: false,
        fertig: false,
      },
    ],
  },
];

function WochenVorschau() {
  return (
    <div
      aria-hidden="true"
      className="karte-tief rounded-2xl border border-line bg-surface p-4 sm:p-5"
    >
      <div className="mb-4 flex items-center justify-between">
        <p className="text-xs font-bold tracking-[0.08em] text-subtle uppercase">Eure Woche</p>
        <span className="rounded-full bg-primary-soft px-2.5 py-1 text-xs font-bold text-primary">
          4 von 13 erledigt
        </span>
      </div>

      {beispiel.map((t, i) => (
        <div key={t.tag} className={`flex gap-4 ${i === 0 ? "" : "pt-4"}`}>
          <div className="flex w-12 shrink-0 flex-col items-center">
            <div
              className={`flex size-11 flex-col items-center justify-center rounded-full ${
                t.heute ? "bg-primary text-on-primary" : "text-muted"
              }`}
            >
              <span className="text-[0.625rem] leading-[1] font-bold tracking-[0.08em] uppercase">
                {t.tag}
              </span>
              <span className="mt-[0.1875rem] text-[1.0625rem] leading-[1] font-extrabold tabular-nums">
                {t.nr}
              </span>
            </div>
            <div className="mt-1 w-px flex-1 bg-line" />
          </div>

          <div
            className={`min-w-0 flex-1 space-y-1.5 pb-6 ${
              i === 0 ? "" : "border-t border-line pt-3.5"
            }`}
          >
            {t.aufgaben.map((a) => (
              <div
                key={a.titel}
                className={`${a.farbe} flex gap-3 rounded-lg border px-3 py-3 ${
                  a.eigen ? "zeile-eigen" : "border-transparent"
                } ${a.fertig ? "opacity-45" : ""}`}
              >
                <span className="inline-flex size-8 shrink-0 items-center justify-center rounded-full bg-[var(--person-soft)] text-xs font-bold text-[var(--person)]">
                  {a.wer[0]}
                </span>
                <div className="min-w-0 flex-1">
                  <div className="flex flex-wrap items-baseline gap-x-2">
                    <span
                      className={`leading-snug font-semibold ${a.fertig ? "line-through" : ""}`}
                    >
                      {a.titel}
                    </span>
                    <span className="text-sm text-[var(--person)]">{a.wer}</span>
                  </div>
                  <p className="mt-0.5 text-xs text-subtle">{a.meta}</p>
                </div>
                <span className="flex size-11 shrink-0 items-center justify-center self-start text-subtle">
                  {a.fertig ? (
                    <Haken className="size-5 text-primary" />
                  ) : (
                    <Kreis className="size-5" />
                  )}
                </span>
              </div>
            ))}
          </div>
        </div>
      ))}
    </div>
  );
}

/**
 * Die Bilanz als waagerechte Balken.
 *
 * Kein Kreisdiagramm: Es geht um den Vergleich dreier Größen, nicht um
 * Anteile an einem Ganzen. Und die Zahl steht an jedem Balken — Farbe allein
 * trägt hier keine Aussage, sonst liest niemand mit einer Rot-Grün-Schwäche
 * die Tafel (WCAG 1.4.1).
 */
function Bilanz() {
  const zeilen = [
    {
      name: "Anna",
      farbe: "person-2",
      prozent: 68,
      minuten: "310 min · Kopflast 9",
    },
    {
      name: "Ben",
      farbe: "person-4",
      prozent: 64,
      minuten: "295 min · Kopflast 8",
    },
    {
      name: "Mia",
      farbe: "person-6",
      prozent: 61,
      minuten: "120 min · Kopflast 2",
    },
  ];

  return (
    <div className="karte-tief rounded-2xl border border-line bg-bg p-6">
      <h3 className="text-base font-bold tracking-tight">Wer trägt wie viel</h3>

      <ul className="mt-5 space-y-4">
        {zeilen.map((z) => (
          <li
            key={z.name}
            className={`${z.farbe} grid grid-cols-[3.5rem_1fr_3rem] items-center gap-3`}
          >
            <span className="truncate text-sm font-medium">{z.name}</span>
            <span className="h-3 overflow-hidden rounded-full bg-surface-3">
              <span
                className="block h-full rounded-full bg-[var(--person)]"
                style={{ width: `${z.prozent}%` }}
              />
            </span>
            <span className="text-right text-sm font-bold tabular-nums">{z.prozent}%</span>
          </li>
        ))}
      </ul>

      <p className="mt-5 text-xs leading-relaxed text-muted text-pretty">
        Anteil der eigenen verfügbaren Zeit, nicht Minuten. Mia ist dreizehn und hat weniger Zeit —
        deshalb steht sie bei ähnlichem Anteil mit der halben Last da.
      </p>

      <ul className="mt-3 grid gap-1 border-t border-line pt-3 text-xs text-subtle">
        {zeilen.map((z) => (
          <li key={z.name} className="flex justify-between gap-2">
            <span>{z.name}</span>
            <span className="tabular-nums">{z.minuten}</span>
          </li>
        ))}
      </ul>
    </div>
  );
}

function AbspracheVorschau() {
  const raster = [
    { tag: "Mo", wer: "Anna", farbe: "person-2" },
    { tag: "Di", wer: "Ben", farbe: "person-4" },
    { tag: "Mi", wer: "Anna", farbe: "person-2" },
    { tag: "Do", wer: "Ben", farbe: "person-4" },
    { tag: "Fr", wer: "Anna", farbe: "person-2" },
    { tag: "Sa", wer: "", farbe: "" },
    { tag: "So", wer: "", farbe: "" },
  ];

  return (
    <div aria-hidden="true" className="karte-tief rounded-2xl border border-line bg-surface p-6">
      <div className="flex items-center gap-2">
        <Zurueck className="size-4 text-subtle" />
        <p className="text-sm font-bold">Zur Kita bringen</p>
      </div>
      <p className="mt-1.5 text-xs leading-relaxed text-muted text-pretty">
        Das sprecht ihr ab. Wer das kann, hängt an euren Arbeitszeiten — die kennt die App nicht,
        und raten wäre hier schlimmer als fragen.
      </p>

      <ul className="mt-5 space-y-1.5">
        {raster.map((r) => (
          <li key={r.tag} className={`${r.farbe} flex items-center gap-3`}>
            <span className="w-8 shrink-0 text-sm font-semibold text-muted">{r.tag}</span>
            <span
              className={`flex min-h-9 flex-1 items-center rounded-lg px-3 text-sm ${
                r.wer === ""
                  ? "border border-dashed border-line-strong text-subtle"
                  : "bg-[var(--person-soft)] font-semibold text-[var(--person)]"
              }`}
            >
              {r.wer === "" ? "— niemand —" : r.wer}
            </span>
          </li>
        ))}
      </ul>
    </div>
  );
}
