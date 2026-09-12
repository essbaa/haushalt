"use client";

import { useState } from "react";
import Link from "next/link";
import { signIn, signUp, useSession } from "@/lib/auth-client";

/**
 * Anmelden und registrieren auf einer Seite.
 *
 * Client Component, weil hier Formularzustand lebt — und weil die Anmeldung
 * dieselbe Adresse hat wie die Seite, ist sie ein gewöhnlicher Aufruf ohne
 * CORS und ohne Token.
 */
export default function Anmelden() {
  const { data: sitzung, isPending } = useSession();
  const [neu, setNeu] = useState(false);
  const [name, setName] = useState("");
  const [email, setEmail] = useState("");
  const [passwort, setPasswort] = useState("");
  const [fehler, setFehler] = useState<string | null>(null);
  const [laeuft, setLaeuft] = useState(false);

  async function absenden(e: React.FormEvent) {
    e.preventDefault();
    setFehler(null);
    setLaeuft(true);
    const antwort = neu
      ? await signUp.email({ name, email, password: passwort })
      : await signIn.email({ email, password: passwort });
    setLaeuft(false);
    if (antwort.error) {
      // Die Meldung kommt vom Anmeldedienst und ist englisch. Bis es eine
      // eigene Übersetzung gibt, ist die echte Meldung besser als eine
      // erfundene.
      setFehler(antwort.error.message ?? "Das hat nicht funktioniert.");
      return;
    }
    window.location.href = "/";
  }

  if (isPending) {
    return <Rahmen>Einen Moment …</Rahmen>;
  }

  if (sitzung) {
    return (
      <Rahmen>
        <p className="mb-4">
          Angemeldet als <strong>{sitzung.user.email}</strong>.
        </p>
        <Link href="/" className="text-accent underline">
          Zum Wochenplan
        </Link>
      </Rahmen>
    );
  }

  return (
    <Rahmen>
      <h1 className="mb-1 text-2xl font-semibold tracking-tight">
        {neu ? "Konto anlegen" : "Anmelden"}
      </h1>
      <p className="mb-6 text-sm leading-relaxed text-muted">
        {neu
          ? "Danach gehört dir ein Haushalt, und du kannst weitere Personen einladen."
          : "Mit der E-Mail-Adresse, mit der du dich registriert hast."}
      </p>

      <form onSubmit={absenden} className="space-y-4">
        {neu && (
          <Feld label="Name" wert={name} setzen={setName} typ="text" autoComplete="name" />
        )}
        <Feld
          label="E-Mail"
          wert={email}
          setzen={setEmail}
          typ="email"
          autoComplete="email"
        />
        <Feld
          label="Passwort"
          wert={passwort}
          setzen={setPasswort}
          typ="password"
          autoComplete={neu ? "new-password" : "current-password"}
        />

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
          {laeuft ? "…" : neu ? "Konto anlegen" : "Anmelden"}
        </button>
      </form>

      <button
        type="button"
        onClick={() => {
          setNeu(!neu);
          setFehler(null);
        }}
        className="mt-6 text-sm text-muted underline"
      >
        {neu ? "Ich habe schon ein Konto" : "Ich brauche ein Konto"}
      </button>
    </Rahmen>
  );
}

function Rahmen({ children }: { children: React.ReactNode }) {
  return (
    <main className="mx-auto w-full max-w-sm px-6 py-20">
      <div className="rounded-lg border border-line bg-surface p-7">{children}</div>
    </main>
  );
}

function Feld({
  label,
  wert,
  setzen,
  typ,
  autoComplete,
}: {
  label: string;
  wert: string;
  setzen: (v: string) => void;
  typ: string;
  autoComplete: string;
}) {
  return (
    <label className="block space-y-1.5">
      <span className="text-sm font-medium">{label}</span>
      <input
        type={typ}
        value={wert}
        onChange={(e) => setzen(e.target.value)}
        autoComplete={autoComplete}
        required
        className="w-full rounded-md border border-line bg-background px-3 py-2 outline-none focus-visible:border-accent"
      />
    </label>
  );
}
