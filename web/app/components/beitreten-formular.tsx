"use client";

import { useState } from "react";
import Link from "next/link";
import { useRouter } from "next/navigation";
import { postMitToken } from "@/lib/browser-token";
import { useSession } from "@/lib/auth-client";

type Haushalt = { id: string; name: string };

export function BeitretenFormular({ vorgabe }: { vorgabe: string }) {
  const router = useRouter();
  const { data: sitzung, isPending } = useSession();
  const [code, setCode] = useState(vorgabe);
  const [fehler, setFehler] = useState<string | null>(null);
  const [laeuft, setLaeuft] = useState(false);

  if (isPending) return <p className="text-muted">Einen Moment …</p>;

  if (!sitzung) {
    return (
      <div className="rounded-lg border border-line bg-surface p-6">
        <p className="mb-4 text-sm leading-relaxed">
          Dafür musst du angemeldet sein. Der Code bleibt gültig.
        </p>
        <Link
          href={`/anmelden?weiter=${encodeURIComponent(`/beitreten?code=${code}`)}`}
          className="text-accent underline"
        >
          Anmelden oder Konto anlegen
        </Link>
      </div>
    );
  }

  async function absenden(e: React.FormEvent) {
    e.preventDefault();
    setFehler(null);
    setLaeuft(true);
    try {
      const haushalt = await postMitToken<Haushalt>(
        `/api/einladungen/${encodeURIComponent(code.trim().toUpperCase())}/annehmen`,
      );
      router.push(`/?haushalt=${haushalt.id}`);
      router.refresh();
    } catch (e) {
      setFehler(e instanceof Error ? e.message : "Das hat nicht funktioniert.");
      setLaeuft(false);
    }
  }

  return (
    <form onSubmit={absenden} className="space-y-4">
      <label className="block space-y-1.5">
        <span className="text-sm font-medium">Code</span>
        <input
          value={code}
          onChange={(e) => setCode(e.target.value)}
          required
          autoCapitalize="characters"
          spellCheck={false}
          placeholder="K7MQ2XPD"
          className="w-full rounded-md border border-line bg-background px-3 py-2 font-mono text-lg tracking-[0.2em] uppercase outline-none focus-visible:border-accent"
        />
      </label>

      {fehler && (
        <p role="alert" className="text-sm text-clay">
          {fehler}
        </p>
      )}

      <button
        type="submit"
        disabled={laeuft}
        className="w-full rounded-md bg-accent px-4 py-2.5 font-medium text-background disabled:opacity-60"
      >
        {laeuft ? "…" : "Beitreten"}
      </button>
    </form>
  );
}
