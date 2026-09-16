import { redirect } from "next/navigation";
import { Bereiche } from "@/app/components/bereiche";
import { Kopfzeile } from "@/app/components/kopfzeile";
import { Melden } from "@/app/components/melden";
import { VorlagenListe } from "@/app/components/vorlagen-liste";
import { ladeHaushalte, ladeVorlagen } from "@/lib/api";
import { serverToken } from "@/lib/auth-token";
import { waehleHaushalt } from "@/lib/haushalt-wahl";

/**
 * Alles, was die App kennt — und was davon bei euch gilt.
 *
 * Die Antwort auf einen berechtigten Einwand: Ein Plan, dem man nicht
 * widersprechen kann, ist eine Behauptung. Hier steht die ganze Bibliothek,
 * jede Zeile mit dem Grund, warum sie gilt oder nicht, und jede lässt sich
 * umdrehen.
 *
 * Bewusst keine Hürde vor dem ersten Plan: Wer 62 Häkchen setzen muss, bevor
 * etwas passiert, hat einen Fragebogen bekommen. Diese Seite ist ein Ort zum
 * Nachsehen, kein Eingangstor.
 */
export default async function Vorlagen({
  searchParams,
}: {
  searchParams: Promise<{ haushalt?: string }>;
}) {
  const token = await serverToken();
  if (!token) redirect("/anmelden?weiter=/vorlagen");

  const haushalte = await ladeHaushalte(token);
  const eigene = haushalte.filter((h) => h.meine_rolle);
  if (eigene.length === 0) redirect("/einrichten");

  const { haushalt } = await searchParams;
  const gewaehlt = (await waehleHaushalt(eigene, haushalt))!;
  const vorlagen = await ladeVorlagen(gewaehlt.id, token);

  return (
    <>
      <Kopfzeile breit />
      <main className="mx-auto w-full max-w-2xl px-5 py-10">
        <Bereiche
          haushaltId={gewaehlt.id}
          planend={gewaehlt.meine_rolle === "planend"}
          aktiv="vorlagen"
        />

        <h1 className="mt-5 mb-2 text-2xl font-extrabold tracking-tight text-balance">Aufgaben</h1>
        <p className="mb-8 max-w-prose text-sm leading-relaxed text-muted text-pretty">
          Alles, was die App kennt, und was davon bei euch gilt. Was ihr nicht braucht, schaltet ihr
          hier dauerhaft ab — es kommt dann in keinem Plan mehr vor.
        </p>

        <VorlagenListe
          haushaltId={gewaehlt.id}
          vorlagen={vorlagen}
          mitglieder={gewaehlt.mitglieder}
          planend={gewaehlt.meine_rolle === "planend"}
        />

        <Melden haushaltId={gewaehlt.id} />
      </main>
    </>
  );
}
