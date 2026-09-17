"use client";

import Link from "next/link";
import { useRouter, useSearchParams } from "next/navigation";
import { Suspense, useState } from "react";
import { Passwortfeld } from "@/app/components/passwortfeld";
import { Passwortstaerke } from "@/app/components/passwortstaerke";
import { anmeldeFehler } from "@/lib/anmelde-fehler";
import { resetPassword } from "@/lib/auth-client";

/**
 * Passwort vergessen — Schritt zwei: das neue setzen.
 *
 * Der Token steht in der Adresse und wird nicht angezeigt. Ein Formular, das
 * ihn sichtbar macht, lädt dazu ein, ihn weiterzugeben.
 *
 * Ein Feld mit Auge statt zweier Felder: Menschen machen denselben Tippfehler
 * zweimal, und dann bestätigt das zweite Feld ihn nur.
 */
export default function Seite() {
  return (
    <Suspense fallback={<Rahmen>Einen Moment …</Rahmen>}>
      <PasswortNeu />
    </Suspense>
  );
}

function PasswortNeu() {
  const router = useRouter();
  const suche = useSearchParams();
  const token = suche.get("token") ?? "";
  const [passwort, setPasswort] = useState("");
  const [wiederholung, setWiederholung] = useState("");
  const [laeuft, setLaeuft] = useState(false);
  const [fehler, setFehler] = useState<string | null>(null);

  // Ohne Token gibt es nichts zu setzen. Das passiert echten Menschen: Der
  // Link wurde aus der Mail kopiert und dabei abgeschnitten.
  if (!token) {
    return (
      <Rahmen>
        <h1 className="mb-2 text-2xl font-extrabold tracking-tight">Der Link ist unvollständig</h1>
        <p className="mb-6 text-sm leading-relaxed text-muted text-pretty">
          In der Adresse fehlt der Teil, der dich ausweist. Am sichersten ist es, den Link in der
          Mail direkt anzutippen statt ihn zu kopieren.
        </p>
        <Link
          href="/passwort-vergessen"
          className="text-sm font-semibold text-primary underline underline-offset-2"
        >
          Neuen Link anfordern
        </Link>
      </Rahmen>
    );
  }

  async function absenden(e: React.FormEvent) {
    e.preventDefault();
    setFehler(null);
    if (passwort !== wiederholung) {
      setFehler("Die beiden Passwörter sind nicht gleich.");
      return;
    }
    setLaeuft(true);
    const antwort = await resetPassword({ newPassword: passwort, token });
    setLaeuft(false);
    if (antwort.error) {
      setFehler(
        anmeldeFehler(
          antwort.error,
          "Das hat nicht funktioniert. Vielleicht ist der Link älter als eine Stunde.",
        ),
      );
      return;
    }
    router.push("/anmelden");
    router.refresh();
  }

  return (
    <Rahmen>
      <h1 className="mb-2 text-2xl font-extrabold tracking-tight">Neues Passwort</h1>
      <p className="mb-7 text-sm leading-relaxed text-muted text-pretty">
        Danach meldest du dich damit an.
      </p>

      <form onSubmit={absenden} className="space-y-4">
        <Passwortfeld
          beschriftung="Neues Passwort"
          value={passwort}
          onChange={(e) => setPasswort(e.target.value)}
          autoComplete="new-password"
          minLength={8}
          required
          hinweis="Mindestens acht Zeichen."
        />

        <Passwortstaerke passwort={passwort} />

        <Passwortfeld
          beschriftung="Wiederholen"
          value={wiederholung}
          onChange={(e) => setWiederholung(e.target.value)}
          autoComplete="new-password"
          required
          hinweis={
            wiederholung.length > 0 && passwort !== wiederholung ? "Noch nicht gleich." : undefined
          }
        />

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
          {laeuft ? "Einen Moment …" : "Passwort setzen"}
        </button>
      </form>
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
