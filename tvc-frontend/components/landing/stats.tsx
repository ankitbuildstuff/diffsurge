"use client";

import { stats } from "@/lib/constants";
import { FadeIn } from "@/components/ui/fade-in";

export function Stats() {
  return (
    <section
      style={{
        background: "var(--bg-secondary)",
        paddingTop: 56,
        paddingBottom: 56,
      }}
    >
      <div className="mx-auto max-w-[1120px] px-6">
        <div
          className="grid gap-8 sm:grid-cols-2 md:grid-cols-4"
          style={{
            borderTop: "1px solid var(--border-light)",
            borderBottom: "1px solid var(--border-light)",
            paddingTop: 40,
            paddingBottom: 40,
          }}
        >
          {stats.map((s, i) => (
            <FadeIn key={s.label} delay={i * 0.08}>
              <div className="text-center md:text-left">
                <p
                  className="font-editorial"
                  style={{
                    fontSize: 32,
                    lineHeight: 1,
                    color: "var(--text-primary)",
                  }}
                >
                  {s.value}
                </p>
                <p
                  className="micro-label"
                  style={{ marginTop: 8, fontSize: 10 }}
                >
                  {s.label}
                </p>
              </div>
            </FadeIn>
          ))}
        </div>
      </div>
    </section>
  );
}
