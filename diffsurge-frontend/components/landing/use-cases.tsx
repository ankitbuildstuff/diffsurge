"use client";

import { FadeIn } from "@/components/ui/fade-in";
import { ArrowUpCircle, Smartphone, Link2, Boxes } from "lucide-react";

const useCases = [
  {
    icon: ArrowUpCircle,
    title: "API Version Upgrades",
    description:
      "Validate that v2 of your API is backward compatible with all existing consumers before cutting over.",
  },
  {
    icon: Smartphone,
    title: "Prevent Mobile Breakage",
    description:
      "Mobile apps can't be force-updated. Ensure API changes won't crash apps still running older versions.",
  },
  {
    icon: Link2,
    title: "Detect Integration Regressions",
    description:
      "Third-party integrations depend on exact response shapes. Catch field removals and type changes early.",
  },
  {
    icon: Boxes,
    title: "Validate Microservice Changes",
    description:
      "When one service changes its contract, replay traffic to verify all downstream consumers still work.",
  },
];

export function UseCases() {
  return (
    <section
      className="section-spacing"
      style={{ background: "var(--bg-secondary)" }}
    >
      <div className="mx-auto px-6" style={{ maxWidth: 1200 }}>
        <FadeIn>
          <div
            style={{ textAlign: "center", maxWidth: 560, margin: "0 auto" }}
          >
            <p className="micro-label" style={{ marginBottom: 16 }}>
              Use Cases
            </p>
            <h2
              style={{
                fontSize: "clamp(1.8rem, 4vw, 2.5rem)",
                fontWeight: 500,
                lineHeight: 1.1,
                letterSpacing: "-0.02em",
                color: "var(--text-primary)",
              }}
            >
              Built for real-world scenarios
            </h2>
          </div>
        </FadeIn>

        <div className="mt-16 grid gap-5 sm:grid-cols-2">
          {useCases.map((uc, i) => (
            <FadeIn key={uc.title} delay={i * 0.08}>
              <div
                className="card h-full flex gap-5"
                style={{ padding: 32, cursor: "default" }}
              >
                <div
                  style={{
                    width: 40,
                    height: 40,
                    borderRadius: 10,
                    border: "1px solid var(--border-subtle)",
                    display: "flex",
                    alignItems: "center",
                    justifyContent: "center",
                    flexShrink: 0,
                  }}
                >
                  <uc.icon
                    size={18}
                    strokeWidth={1.5}
                    style={{ color: "var(--text-muted)" }}
                  />
                </div>
                <div>
                  <h3
                    style={{
                      fontSize: 16,
                      fontWeight: 500,
                      color: "var(--text-primary)",
                      lineHeight: 1.3,
                    }}
                  >
                    {uc.title}
                  </h3>
                  <p
                    style={{
                      marginTop: 8,
                      fontSize: 14,
                      lineHeight: 1.7,
                      color: "var(--text-muted)",
                    }}
                  >
                    {uc.description}
                  </p>
                </div>
              </div>
            </FadeIn>
          ))}
        </div>
      </div>
    </section>
  );
}
