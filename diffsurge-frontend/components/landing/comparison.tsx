"use client";

import { FadeIn } from "@/components/ui/fade-in";
import { Check, X, Minus } from "lucide-react";

const rows = [
  {
    feature: "Real production traffic",
    unitTests: false,
    contractTests: false,
    diffSurge: true,
  },
  {
    feature: "Edge case coverage",
    unitTests: "limited" as const,
    contractTests: "limited" as const,
    diffSurge: true,
  },
  {
    feature: "Schema validation",
    unitTests: false,
    contractTests: true,
    diffSurge: true,
  },
  {
    feature: "Runtime behavior testing",
    unitTests: "limited" as const,
    contractTests: false,
    diffSurge: true,
  },
  {
    feature: "Status code drift detection",
    unitTests: false,
    contractTests: false,
    diffSurge: true,
  },
  {
    feature: "Latency regression detection",
    unitTests: false,
    contractTests: false,
    diffSurge: true,
  },
  {
    feature: "PII redaction",
    unitTests: false,
    contractTests: false,
    diffSurge: true,
  },
  {
    feature: "CI/CD integration",
    unitTests: true,
    contractTests: true,
    diffSurge: true,
  },
];

function CellIcon({ value }: { value: boolean | string }) {
  if (value === true)
    return <Check size={16} style={{ color: "var(--accent-green)" }} />;
  if (value === false)
    return <X size={16} style={{ color: "var(--text-faint)" }} />;
  return <Minus size={16} style={{ color: "var(--accent-amber)" }} />;
}

export function Comparison() {
  return (
    <section className="section-spacing">
      <div className="mx-auto px-6" style={{ maxWidth: 1200 }}>
        <FadeIn>
          <div
            style={{ textAlign: "center", maxWidth: 560, margin: "0 auto" }}
          >
            <p className="micro-label" style={{ marginBottom: 16 }}>
              Comparison
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
              Beyond traditional API testing
            </h2>
            <p
              style={{
                marginTop: 16,
                fontSize: 15,
                lineHeight: 1.7,
                color: "var(--text-muted)",
              }}
            >
              See how DiffSurge compares to existing testing approaches.
            </p>
          </div>
        </FadeIn>

        <FadeIn delay={0.15}>
          <div
            className="card mt-16 overflow-hidden"
            style={{ cursor: "default", padding: 0 }}
          >
            <div className="overflow-x-auto">
              <table style={{ width: "100%", borderCollapse: "collapse" }}>
                <thead>
                  <tr
                    style={{
                      borderBottom: "1px solid var(--border-subtle)",
                    }}
                  >
                    <th
                      style={{
                        padding: "16px 24px",
                        textAlign: "left",
                        fontSize: 13,
                        fontWeight: 500,
                        color: "var(--text-muted)",
                      }}
                    >
                      Feature
                    </th>
                    <th
                      style={{
                        padding: "16px 24px",
                        textAlign: "center",
                        fontSize: 13,
                        fontWeight: 500,
                        color: "var(--text-muted)",
                      }}
                    >
                      Unit Tests
                    </th>
                    <th
                      style={{
                        padding: "16px 24px",
                        textAlign: "center",
                        fontSize: 13,
                        fontWeight: 500,
                        color: "var(--text-muted)",
                      }}
                    >
                      Contract Testing
                    </th>
                    <th
                      style={{
                        padding: "16px 24px",
                        textAlign: "center",
                        fontSize: 13,
                        fontWeight: 500,
                        color: "var(--text-primary)",
                        background: "var(--bg-tertiary)",
                      }}
                    >
                      DiffSurge
                    </th>
                  </tr>
                </thead>
                <tbody>
                  {rows.map((row, i) => (
                    <tr
                      key={i}
                      style={{
                        borderBottom:
                          i < rows.length - 1
                            ? "1px solid var(--border-subtle)"
                            : "none",
                      }}
                    >
                      <td
                        style={{
                          padding: "14px 24px",
                          fontSize: 14,
                          color: "var(--text-secondary)",
                        }}
                      >
                        {row.feature}
                      </td>
                      <td
                        style={{ padding: "14px 24px", textAlign: "center" }}
                      >
                        <div
                          style={{
                            display: "flex",
                            justifyContent: "center",
                          }}
                        >
                          <CellIcon value={row.unitTests} />
                        </div>
                      </td>
                      <td
                        style={{ padding: "14px 24px", textAlign: "center" }}
                      >
                        <div
                          style={{
                            display: "flex",
                            justifyContent: "center",
                          }}
                        >
                          <CellIcon value={row.contractTests} />
                        </div>
                      </td>
                      <td
                        style={{
                          padding: "14px 24px",
                          textAlign: "center",
                          background: "var(--bg-tertiary)",
                        }}
                      >
                        <div
                          style={{
                            display: "flex",
                            justifyContent: "center",
                          }}
                        >
                          <CellIcon value={row.diffSurge} />
                        </div>
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          </div>
        </FadeIn>
      </div>
    </section>
  );
}
