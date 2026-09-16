import { Bereiche } from "@/app/components/bereiche";
import { KastenSkelett, KopfSkelett } from "@/app/components/skelett";

export default function Laden() {
  return (
    <main className="mx-auto w-full max-w-lg px-5 py-10">
      <Bereiche planend aktiv="termine" skelett />
      <KopfSkelett />
      <div className="space-y-3">
        {Array.from({ length: 4 }, (_, i) => (
          <KastenSkelett key={i} hoch="h-16" />
        ))}
      </div>
    </main>
  );
}
