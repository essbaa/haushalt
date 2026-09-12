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
    <main className="mx-auto w-full max-w-md px-6 py-16">
      <Link href="/" className="text-sm text-muted underline">
        Zurück zum Wochenplan
      </Link>

      <h1 className="mt-6 mb-1 text-2xl font-semibold tracking-tight">
        Einem Haushalt beitreten
      </h1>
      <p className="mb-8 text-sm leading-relaxed text-muted">
        Du brauchst einen Code von jemandem, der im Haushalt plant. Wenn du noch
        kein Konto hast, leg zuerst eines an — der Code wartet.
      </p>

      <BeitretenFormular vorgabe={code ?? ""} />
    </main>
  );
}
