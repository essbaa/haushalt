"use client";

import Link from "next/link";
import { signOut, useSession } from "@/lib/auth-client";

/**
 * Wer ist angemeldet — eine Zeile, bewusst unauffällig.
 *
 * Client Component: Die Sitzung steckt in einem Cookie, und ob eines da ist,
 * weiß erst der Browser. Server-seitig ginge es auch, würde aber die
 * Zwischenspeicherung der Seite verhindern — und die Seite ist derzeit für
 * alle dieselbe.
 */
export function Sitzung() {
  const { data: sitzung, isPending } = useSession();

  if (isPending) return null;

  if (!sitzung) {
    return (
      <Link href="/anmelden" className="text-sm text-accent underline">
        Anmelden
      </Link>
    );
  }

  return (
    <span className="flex flex-wrap items-baseline gap-3 text-sm text-muted">
      Angemeldet als {sitzung.user.name || sitzung.user.email}
      <button
        type="button"
        onClick={() => signOut().then(() => window.location.reload())}
        className="underline"
      >
        Abmelden
      </button>
    </span>
  );
}
