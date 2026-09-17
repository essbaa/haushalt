"use client";

import { useState } from "react";
import Link from "next/link";
import { useRouter } from "next/navigation";
import { Passwortfeld } from "@/app/components/passwortfeld";
import { Passwortstaerke } from "@/app/components/passwortstaerke";
import { Feld } from "@/app/components/ui";
import { anmeldeFehler } from "@/lib/anmelde-fehler";
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
  const [wiederholung, setWiederholung] = useState("");
  // Aus der Adresse vorbefüllt: Ein Einladungslink bringt den Code mit, und
  // wer eingeladen ist, soll nicht noch eine zweite Zeichenkette abtippen.
  const [zugang, setZugang] = useState(() => {
    if (typeof window === "undefined") return "";
    return new URLSearchParams(window.location.search).get("zugang") ?? "";
  });
  const [fehler, setFehler] = useState<string | null>(null);
  const [laeuft, setLaeuft] = useState(false);
  const [bestaetigen, setBestaetigen] = useState(false);

  async function absenden(e: React.FormEvent) {
    e.preventDefault();
    setFehler(null);

    // Vor dem Netz prüfen, was ohne Netz zu prüfen ist. Ein Rundgang zum
    // Server, um zu erfahren, dass zwei Felder nicht gleich sind, ist eine
    // Wartezeit ohne Erkenntnis.
    if (neu && passwort !== wiederholung) {
      setFehler("Die beiden Passwörter sind nicht gleich.");
      return;
    }

    setLaeuft(true);
    const antwort = neu
      ? await signUp.email(
          { name, email, password: passwort },
          // Der Code reist im Kopf der Anfrage. Geprüft wird er im Dienst
          // (lib/auth.ts) — dieses Feld ist die Eingabe, nicht die Sperre.
          { headers: { "x-zugangscode": zugang.trim() } },
        )
      : await signIn.email({ email, password: passwort });
    setLaeuft(false);
    if (antwort.error) {
      setFehler(anmeldeFehler(antwort.error, "Das hat nicht funktioniert."));
      return;
    }
    // push statt window.location: Next kennt die Route und muss die Seite
    // nicht neu laden. refresh danach, weil die Startseite auf dem Server
    // gerendert wird und die frische Sitzung noch nicht kennt.
    // Nach der Registrierung ist eine Bestätigungsmail unterwegs. Das gehört
    // gesagt — eine Mail, die unangekündigt eintrifft, wirkt wie ein
    // Versehen, und wer sie nicht erwartet, bestätigt sie auch nicht.
    if (neu) {
      setBestaetigen(true);
      return;
    }

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

  if (bestaetigen) {
    return (
      <Rahmen>
        <h1 className="mb-2 text-2xl font-extrabold tracking-tight">Konto angelegt</h1>
        <p className="mb-6 text-sm leading-relaxed text-muted text-pretty">
          Wir haben eine Mail an <strong className="text-fg">{email}</strong> geschickt. Bestätige
          die Adresse, sobald du magst — danach können wir dir ein neues Passwort schicken, falls du
          es vergisst.
        </p>
        <button
          type="button"
          onClick={() => {
            router.push(weiter());
            router.refresh();
          }}
          className="inline-flex min-h-12 w-full items-center justify-center rounded-md bg-primary px-4 text-sm font-semibold text-on-primary transition-colors hover:bg-primary-hover"
        >
          Weiter zum Wochenplan
        </button>
      </Rahmen>
    );
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
          ? "Die App ist noch geschlossen — zum Anlegen brauchst du einen Zugangscode. Danach gehört dir ein Haushalt, und du kannst weitere Personen einladen."
          : "Mit der E-Mail-Adresse, mit der du dich registriert hast."}
      </p>

      <form onSubmit={absenden} className="space-y-4">
        {/* Der Zugangscode steht oben und nicht unten: Wer keinen hat, soll es
            erfahren, bevor er ein Passwort ausdenkt. */}
        {neu && (
          <Feld
            beschriftung="Zugangscode"
            value={zugang}
            onChange={(e) => setZugang(e.target.value)}
            autoComplete="off"
            autoCapitalize="characters"
            spellCheck={false}
            required
            hinweis="Die App ist in Erprobung. Den Code bekommst du von der Person, die dich einlädt."
          />
        )}
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
        <Passwortfeld
          beschriftung="Passwort"
          value={passwort}
          onChange={(e) => setPasswort(e.target.value)}
          autoComplete={neu ? "new-password" : "current-password"}
          minLength={neu ? 8 : undefined}
          required
          hinweis={neu ? "Mindestens acht Zeichen." : undefined}
        />

        {neu && (
          <>
            <Passwortstaerke passwort={passwort} umfeld={[name, email.split("@")[0] ?? ""]} />
            <Passwortfeld
              beschriftung="Passwort wiederholen"
              value={wiederholung}
              onChange={(e) => setWiederholung(e.target.value)}
              autoComplete="new-password"
              required
              hinweis={
                wiederholung.length > 0 && passwort !== wiederholung
                  ? "Noch nicht gleich."
                  : undefined
              }
            />
          </>
        )}

        {!neu && (
          <Link
            href="/passwort-vergessen"
            className="inline-flex min-h-11 items-center text-sm text-muted underline underline-offset-2 hover:text-fg"
          >
            Passwort vergessen?
          </Link>
        )}

        {fehler && (
          <p
            role="alert"
            className="rounded-md border border-danger/40 px-3 py-2 text-sm text-danger"
          >
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
