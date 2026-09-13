"use client";

import { useState } from "react";
import Link from "next/link";
import { useRouter } from "next/navigation";
import { Feld } from "@/app/components/ui";
import { signIn, signUp, useSession } from "@/lib/auth-client";

/**
 * Anmelden und registrieren auf einer Seite.
 *
 * Client Component, weil hier Formularzustand lebt — und weil die Anmeldung
 * dieselbe Adresse hat wie die Seite, ist sie ein gewöhnlicher Aufruf ohne
 * CORS und ohne Token.
 */
export default function Anmelden() {
  const router = useRouter();
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
    // push statt window.location: Next kennt die Route und muss die Seite
    // nicht neu laden. refresh danach, weil die Startseite auf dem Server
    // gerendert wird und die frische Sitzung noch nicht kennt.
    router.push(weiter());
    router.refresh();
  }

  /**
   * Wohin nach dem Anmelden.
   *
   * Aus der Adresse gelesen, damit ein Einladungslink nicht verloren geht:
   * /anmelden?weiter=/beitreten?code=… führt hinterher dorthin zurück.
   *
   * Nur Pfade, und keine, die mit zwei Schrägstrichen beginnen. Sonst wäre
   * ?weiter=//fremde.example eine offene Weiterleitung — der klassische Weg,
   * eine vertrauenswürdige Anmeldeseite als Sprungbrett zu missbrauchen.
   */
  function weiter(): string {
    const ziel = new URLSearchParams(window.location.search).get("weiter");
    if (!ziel || !ziel.startsWith("/") || ziel.startsWith("//")) return "/";
    return ziel;
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
        <Link href="/" className="text-primary underline">
          Zum Wochenplan
        </Link>
      </Rahmen>
    );
  }

  return (
    <Rahmen>
      <h1 className="mb-2 text-2xl font-extrabold tracking-tight">
        {neu ? "Konto anlegen" : "Anmelden"}
      </h1>
      <p className="mb-7 text-sm leading-relaxed text-muted">
        {neu
          ? "Danach gehört dir ein Haushalt, und du kannst weitere Personen einladen."
          : "Mit der E-Mail-Adresse, mit der du dich registriert hast."}
      </p>

      <form onSubmit={absenden} className="space-y-4">
        {neu && (
          <Feld
            beschriftung="Name"
            value={name}
            onChange={(e) => setName(e.target.value)}
            autoComplete="name"
            required
          />
        )}
        <Feld
          beschriftung="E-Mail"
          type="email"
          inputMode="email"
          value={email}
          onChange={(e) => setEmail(e.target.value)}
          autoComplete="email"
          required
        />
        <Feld
          beschriftung="Passwort"
          type="password"
          value={passwort}
          onChange={(e) => setPasswort(e.target.value)}
          autoComplete={neu ? "new-password" : "current-password"}
          required
          hinweis={neu ? "Mindestens acht Zeichen." : undefined}
        />

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
          {laeuft ? "Einen Moment …" : neu ? "Konto anlegen" : "Anmelden"}
        </button>
      </form>

      <button
        type="button"
        onClick={() => {
          setNeu(!neu);
          setFehler(null);
        }}
        className="mt-6 inline-flex min-h-11 items-center text-sm font-semibold text-primary underline underline-offset-2"
      >
        {neu ? "Ich habe schon ein Konto" : "Ich brauche ein Konto"}
      </button>
    </Rahmen>
  );
}

function Rahmen({ children }: { children: React.ReactNode }) {
  return (
    <main className="mx-auto flex w-full max-w-sm flex-1 flex-col justify-center px-5 py-12">
      <div className="rounded-lg border border-line bg-surface p-6 sm:p-7">{children}</div>
    </main>
  );
}
