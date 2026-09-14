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
          className="inline-flex min-h-11 items-center rounded-md bg-primary px-4 text-sm font-semibold text-on-primary transition-colors hover:bg-primary-hover"
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
        <span className="block text-sm font-semibold">Code</span>
        <input
          value={code}
          onChange={(e) => setCode(e.target.value)}
          required
          autoCapitalize="characters"
          autoComplete="off"
          spellCheck={false}
          placeholder="K7MQ2XPD"
          className="block min-h-14 w-full rounded-md border border-line-strong bg-surface px-4 text-center text-2xl font-bold tracking-[0.3em] uppercase placeholder:font-normal placeholder:text-subtle focus:border-primary"
        />
      </label>

      {fehler && (
        <p role="alert" className="rounded-md border border-danger/40 px-3 py-2 text-sm text-danger">
          {fehler}
        </p>
      )}

      <button
        type="submit"
        disabled={laeuft}
        className="inline-flex min-h-12 w-full items-center justify-center rounded-md bg-primary px-4 text-sm font-semibold text-on-primary transition-colors hover:bg-primary-hover disabled:opacity-45"
      >
        {laeuft ? "Einen Moment …" : "Beitreten"}
      </button>
    </form>
  );
}
