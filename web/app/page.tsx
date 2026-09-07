import { ApiStatus } from "@/app/components/api-status";

/**
 * Startseite. Server Component — sie rendert nur den Rahmen.
 *
 * Die Statuszeile ist bewusst eine Client Component: Sie soll den Go-Dienst
 * aus dem Browser heraus fragen, weil erst das die Sprachgrenze samt CORS
 * wirklich durchmisst.
 */
export default function Page() {
  return (
    <main className="mx-auto flex w-full max-w-xl flex-1 flex-col justify-center gap-10 px-6 py-16">
      <header className="space-y-3">
        <p className="font-mono text-xs uppercase tracking-widest text-muted">
          Tag 0 · Durchstich
        </p>
        <h1 className="text-4xl font-semibold tracking-tight text-balance">
          Haushalt
        </h1>
        <p className="text-lg leading-relaxed text-muted text-pretty">
          Wochenplanung für berufstätige Eltern. Die App verteilt nicht nur
          Zeit, sondern auch Kopflast — die Arbeit, an die jemand denken muss.
        </p>
      </header>

      <section className="space-y-3">
        <h2 className="font-mono text-xs uppercase tracking-widest text-muted">
          Verbindung
        </h2>
        <ApiStatus />
      </section>

      <footer className="border-t border-line pt-6 text-sm leading-relaxed text-muted">
        Diese Seite ruft den Go-Dienst wirklich auf. Steht oben &bdquo;API
        erreichbar&ldquo;, sind zwei Deployments, zwei Sprachen, CORS und die
        Datenbank auf einmal bewiesen.
      </footer>
    </main>
  );
}
