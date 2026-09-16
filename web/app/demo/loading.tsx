import { Balken, TagSkelett } from "@/app/components/skelett";

export default function Laden() {
  return (
    <main className="mx-auto w-full max-w-2xl px-5 py-8">
      <div className="mb-8 h-20 animate-pulse rounded-lg border border-clay/40 bg-clay-soft" />
      <div className="mb-8 space-y-3">
        <Balken breit="w-56" hoch="h-9" />
        <Balken breit="w-40" hoch="h-4" />
      </div>
      <TagSkelett erster zeilen={2} />
      <TagSkelett zeilen={3} />
    </main>
  );
}
