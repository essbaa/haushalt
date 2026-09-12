import Link from "next/link";
import { EinladenFormular } from "@/app/components/einladen-formular";

/**
 * Jemanden in den Haushalt einladen.
 *
 * Server Component nur als Rahmen: Sie liest den Haushalt aus der Adresse und
 * reicht ihn weiter. Damit braucht die Client Component kein useSearchParams
 * und keine Suspense-Grenze.
 */
export default async function Einladen({
  searchParams,
}: {
  searchParams: Promise<{ haushalt?: string }>;
}) {
  const { haushalt } = await searchParams;

  return (
    <main className="mx-auto w-full max-w-md px-6 py-16">
      <Link href="/" className="text-sm text-muted underline">
        Zurück zum Wochenplan
      </Link>

      <h1 className="mt-6 mb-1 text-2xl font-semibold tracking-tight">
        Jemanden einladen
      </h1>
      <p className="mb-8 text-sm leading-relaxed text-muted">
        Du bekommst einen Code, der einmal gilt und nach einer Woche abläuft.
        Die Rolle legst du jetzt fest — wer dazukommt, entscheidet nicht selbst,
        was er im Haushalt darf.
      </p>

      {haushalt ? (
        <EinladenFormular haushaltId={haushalt} />
      ) : (
        <p className="text-sm text-clay">
          Es fehlt der Haushalt in der Adresse. Geh über den Wochenplan hierher.
        </p>
      )}
    </main>
  );
}
