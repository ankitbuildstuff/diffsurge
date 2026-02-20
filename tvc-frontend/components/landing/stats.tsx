"use client";

import { stats } from "@/lib/constants";
import { FadeIn } from "@/components/ui/fade-in";

export function Stats() {
  return (
    <section className="border-y border-zinc-100 bg-white py-14">
      <div className="mx-auto grid max-w-[1200px] gap-8 px-6 sm:grid-cols-2 md:grid-cols-4">
        {stats.map((s, i) => (
          <FadeIn key={s.label} delay={i * 0.08}>
            <div className="text-center md:text-left">
              <p className="text-3xl font-bold tracking-tight text-zinc-900">
                {s.value}
              </p>
              <p className="mt-1 text-[13px] text-zinc-500">{s.label}</p>
            </div>
          </FadeIn>
        ))}
      </div>
    </section>
  );
}
