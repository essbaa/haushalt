import Link from "next/link";
import { BeitretenFormular } from "@/app/components/beitreten-formular";

/** Einer Einladung folgen. Der Code kommt aus dem Link oder wird getippt. */
export default async function Beitreten({
  searchParams,
}: {
  searchParams: Promise<{ code?: string }>;
}) {
  const { code } = await searchParams;

  return (
    <main className="mx-auto w-full max-w-md px-5 py-10">
      <Link href="/" className="inline-flex min-h-11 items-center gap-1.5 text-sm text-muted transition-colors hover:text-fg">
        <span aria-hidden="true">←</span>
        Zurück zum Wochenplan
      </Link>

      <h1 className="mt-5 mb-2 text-2xl font-extrabold tracking-tight text-balance">
        Einem Haushalt beitreten
      </h1>
      <p className="mb-8 max-w-prose text-sm leading-relaxed text-muted text-pretty">
        Du brauchst einen Code von jemandem, der im Haushalt plant. Wenn du noch
        kein Konto hast, leg zuerst eines an — der Code wartet.
      </p>

      <BeitretenFormular vorgabe={code ?? ""} />
    </main>
  );
}
