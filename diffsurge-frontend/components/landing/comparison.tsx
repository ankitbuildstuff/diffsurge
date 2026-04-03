"use client";

import { FadeIn } from "@/components/ui/fade-in";
import {
  Check,
  X,
  Minus,
  LockKeyhole,
  FlaskConical,
  Activity,
} from "lucide-react";

const rows = [
  {
    feature: "Open-source core",
    closedPlatforms: false,
    contractTools: "limited" as const,
    observability: false,
    diffSurge: true,
  },
  {
    feature: "Self-host or managed",
    closedPlatforms: false,
    contractTools: "limited" as const,
    observability: "limited" as const,
    diffSurge: true,
  },
  {
    feature: "Real traffic replay before deploy",
    closedPlatforms: false,
    contractTools: false,
    observability: false,
    diffSurge: true,
  },
  {
    feature: "Schema diffing in CI",
    closedPlatforms: "limited" as const,
    contractTools: true,
    observability: false,
    diffSurge: true,
  },
  {
    feature: "Runtime drift detection",
    closedPlatforms: "limited" as const,
    contractTools: false,
    observability: "limited" as const,
    diffSurge: true,
  },
  {
    feature: "Deployment blocking with CLI exit codes",
    closedPlatforms: false,
    contractTools: "limited" as const,
    observability: false,
    diffSurge: true,
  },
  {
    feature: "PII redaction",
    closedPlatforms: "limited" as const,
    contractTools: false,
    observability: "limited" as const,
    diffSurge: true,
  },
  {
    feature: "Transparent workflows developers can inspect",
    closedPlatforms: false,
    contractTools: "limited" as const,
    observability: false,
    diffSurge: true,
  },
];

const platformCards = [
  {
    icon: LockKeyhole,
    title: "Closed API platforms",
    description:
      "Great for hosted workflows, but often tied to seats, opaque internals, and limited control over how replay or governance actually runs.",
  },
  {
    icon: FlaskConical,
    title: "Contract-only tools",
    description:
      "Useful for schema agreements, but they still depend on synthetic examples and usually miss production-only request shapes.",
  },
  {
    icon: Activity,
    title: "Observability suites",
    description:
      "Excellent after something breaks in prod. Much weaker at blocking risky API changes before the deploy happens.",
  },
];

const advantages = [
  "Open-source CLI and core workflow",
  "Real request replay instead of demo traffic",
  "Self-host when compliance or cost matters",
  "Schema diff plus runtime drift in one product",
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
              Why teams choose DiffSurge over closed platforms
            </h2>
            <p
              style={{
                marginTop: 16,
                fontSize: 15,
                lineHeight: 1.7,
                color: "var(--text-muted)",
              }}
            >
              Most API tools solve one slice of the problem. DiffSurge combines
              schema governance, real traffic replay, and open-source control in
              one workflow.
            </p>
          </div>
        </FadeIn>

        <FadeIn delay={0.15}>
          <div className="mt-16 grid gap-5 lg:grid-cols-[1.15fr_0.85fr]">
            <div className="grid gap-5 sm:grid-cols-3">
              {platformCards.map((card, index) => (
                <div
                  key={card.title}
                  className="card h-full"
                  style={{
                    padding: 24,
                    cursor: "default",
                    background:
                      index === 0
                        ? "linear-gradient(180deg, #ffffff 0%, #fafafa 100%)"
                        : "var(--bg-primary)",
                  }}
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
                      marginBottom: 16,
                    }}
                  >
                    <card.icon
                      size={18}
                      strokeWidth={1.5}
                      style={{ color: "var(--text-muted)" }}
                    />
                  </div>
                  <h3
                    style={{
                      fontSize: 16,
                      fontWeight: 500,
                      color: "var(--text-primary)",
                    }}
                  >
                    {card.title}
                  </h3>
                  <p
                    style={{
                      marginTop: 8,
                      fontSize: 14,
                      lineHeight: 1.7,
                      color: "var(--text-muted)",
                    }}
                  >
                    {card.description}
                  </p>
                </div>
              ))}
            </div>

            <div
              className="card"
              style={{
                padding: 28,
                cursor: "default",
                background:
                  "linear-gradient(135deg, rgba(34,197,94,0.08) 0%, rgba(37,99,235,0.04) 100%)",
              }}
            >
              <p className="micro-label" style={{ marginBottom: 14 }}>
                DiffSurge Edge
              </p>
              <h3
                style={{
                  fontSize: 24,
                  fontWeight: 500,
                  lineHeight: 1.2,
                  color: "var(--text-primary)",
                }}
              >
                Built for teams that want leverage, not lock-in.
              </h3>
              <p
                style={{
                  marginTop: 14,
                  fontSize: 15,
                  lineHeight: 1.8,
                  color: "var(--text-secondary)",
                }}
              >
                You keep the transparent CLI, keep the option to self-host, and
                still get a managed dashboard when you need it. That makes
                DiffSurge easier to adopt for startups and more defensible for
                infra teams.
              </p>
              <div
                style={{
                  marginTop: 22,
                  display: "grid",
                  gap: 12,
                }}
              >
                {advantages.map((item) => (
                  <div
                    key={item}
                    style={{
                      display: "flex",
                      alignItems: "center",
                      gap: 10,
                      fontSize: 14,
                      color: "var(--text-secondary)",
                    }}
                  >
                    <Check
                      size={16}
                      strokeWidth={2}
                      style={{ color: "var(--accent-green)", flexShrink: 0 }}
                    />
                    <span>{item}</span>
                  </div>
                ))}
              </div>
            </div>
          </div>
        </FadeIn>

        <FadeIn delay={0.24}>
          <div
            className="card mt-8 overflow-hidden"
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
                      Closed Platforms
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
                      Contract Tools
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
                      Observability
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
                          <CellIcon value={row.closedPlatforms} />
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
                          <CellIcon value={row.contractTools} />
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
                          <CellIcon value={row.observability} />
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
