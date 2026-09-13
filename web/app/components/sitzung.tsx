"use client";

import Link from "next/link";
import { useRouter } from "next/navigation";
import { signOut, useSession } from "@/lib/auth-client";

/**
 * Wer angemeldet ist — eine Zeile in der Kopfleiste.
 *
 * Client Component: Die Sitzung steckt in einem Cookie, und ob eines da ist,
 * weiß erst der Browser. Solange das unklar ist, bleibt der Platz leer statt
 * zu flackern.
 */
export function Sitzung() {
  const router = useRouter();
  const { data: sitzung, isPending } = useSession();

  if (isPending) return <span className="h-11" aria-hidden="true" />;

  if (!sitzung) {
    return (
      <Link
        href="/anmelden"
        className="inline-flex min-h-11 items-center rounded-md bg-primary px-4 text-sm font-semibold text-on-primary transition-colors hover:bg-primary-hover"
      >
        Anmelden
      </Link>
    );
  }

  const name = sitzung.user.name || sitzung.user.email;

  return (
    <div className="flex items-center gap-2">
      <span className="hidden text-sm text-muted sm:inline">{name}</span>
      <button
        type="button"
        onClick={() =>
          signOut().then(() => {
            router.push("/");
            router.refresh();
          })
        }
        className="inline-flex min-h-11 items-center rounded-md px-3 text-sm font-semibold text-muted transition-colors hover:bg-surface-2 hover:text-fg"
      >
        Abmelden
      </button>
    </div>
  );
}
