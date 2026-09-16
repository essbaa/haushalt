import { Bereiche } from "@/app/components/bereiche";
import { KastenSkelett, KopfSkelett } from "@/app/components/skelett";

export default function Laden() {
  return (
    <main className="mx-auto w-full max-w-lg px-5 py-10">
      <Bereiche planend aktiv="einstellungen" skelett />
      <KopfSkelett />
      <div className="space-y-6">
        <KastenSkelett hoch="h-40" />
        <KastenSkelett hoch="h-56" />
        <KastenSkelett hoch="h-32" />
      </div>
    </main>
  );
}
