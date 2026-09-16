import { Bereiche } from "@/app/components/bereiche";
import { KastenSkelett, KopfSkelett } from "@/app/components/skelett";

export default function Laden() {
  return (
    <main className="mx-auto w-full max-w-2xl px-5 py-10">
      <Bereiche planend aktiv="vorlagen" skelett />
      <KopfSkelett />
      <div className="space-y-5">
        <KastenSkelett hoch="h-5" />
        <div className="flex flex-wrap gap-2">
          {Array.from({ length: 4 }, (_, i) => (
            <KastenSkelett key={i} hoch="h-10 w-32" />
          ))}
        </div>
        {Array.from({ length: 6 }, (_, i) => (
          <KastenSkelett key={i} />
        ))}
      </div>
    </main>
  );
}
