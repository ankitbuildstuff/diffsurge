"use client";

import { FadeIn } from "@/components/ui/fade-in";
import {
  AlertTriangle,
  ArrowRightLeft,
  Trash2,
  Plus,
  Layers,
  Timer,
  Hash,
  ShieldAlert,
  Terminal,
  Container,
  Server,
  Cpu,
} from "lucide-react";

const catches = [
  {
    icon: <Trash2 size={15} strokeWidth={1.5} />,
    title: "Removed fields",
    description: "A required field is removed from a response body",
    severity: "breaking" as const,
  },
  {
    icon: <ArrowRightLeft size={15} strokeWidth={1.5} />,
    title: "Type changes",
    description: "A field type changes — string becomes number",
    severity: "breaking" as const,
  },
  {
    icon: <AlertTriangle size={15} strokeWidth={1.5} />,
    title: "Deleted endpoints",
    description: "An endpoint is renamed or removed entirely",
    severity: "breaking" as const,
  },
  {
    icon: <Plus size={15} strokeWidth={1.5} />,
    title: "New required params",
    description: "A new required query parameter is added",
    severity: "breaking" as const,
  },
  {
    icon: <Layers size={15} strokeWidth={1.5} />,
    title: "Shape changes",
    description: "A nested object structure changes its shape",
    severity: "warning" as const,
  },
  {
    icon: <Timer size={15} strokeWidth={1.5} />,
    title: "Latency regressions",
    description: "Response time regresses beyond a threshold",
    severity: "warning" as const,
  },
  {
    icon: <Hash size={15} strokeWidth={1.5} />,
    title: "Status code drift",
    description: "A status code changes for identical requests",
    severity: "warning" as const,
  },
  {
    icon: <ShieldAlert size={15} strokeWidth={1.5} />,
    title: "PII leaks",
    description: "PII appears in a field that was previously clean",
    severity: "info" as const,
  },
];

const platforms = [
  { icon: <Terminal size={15} strokeWidth={1.5} />, name: "macOS" },
  { icon: <Terminal size={15} strokeWidth={1.5} />, name: "Linux" },
  { icon: <Terminal size={15} strokeWidth={1.5} />, name: "Windows" },
  { icon: <Container size={15} strokeWidth={1.5} />, name: "Docker" },
  { icon: <Server size={15} strokeWidth={1.5} />, name: "CI/CD" },
  { icon: <Cpu size={15} strokeWidth={1.5} />, name: "Kubernetes" },
];

const severityAccent = {
  breaking: "var(--accent-orange)",
  warning: "var(--accent-yellow)",
  info: "var(--accent-teal)",
};

export function Capabilities() {
  return (
    <section
      className="section-spacing"
      style={{ background: "var(--bg-primary)" }}
    >
      <div className="mx-auto max-w-[1120px] px-6">
        <FadeIn>
          <div style={{ textAlign: "center", maxWidth: 560, margin: "0 auto" }}>
            <div
              className="micro-label"
              style={{
                display: "inline-flex",
                alignItems: "center",
                gap: 8,
                marginBottom: 16,
              }}
            >
              <span
                className="data-stripe"
                style={{
                  width: 10,
                  height: 10,
                  borderRadius: 3,
                  display: "inline-block",
                }}
              />
              <span>What Driftsurge catches</span>
            </div>
            <h2
              className="font-editorial"
              style={{
                fontSize: "clamp(1.8rem, 4vw, 2.6rem)",
                lineHeight: 1.1,
                color: "var(--text-primary)",
              }}
            >
              The breaking changes that slip through{" "}
              <span className="font-editorial-italic">unit tests</span>
            </h2>
            <p
              style={{
                marginTop: 16,
                fontSize: 14,
                lineHeight: 1.7,
                color: "var(--text-muted)",
              }}
            >
              Unit tests verify logic. Integration tests verify contracts.
              Driftsurge replays actual production traffic — catching real-world
              mismatches before your users do.
            </p>
          </div>
        </FadeIn>

        {/* Grid of catches */}
        <div
          className="mt-14 grid grid-cols-1 gap-4 sm:grid-cols-2 lg:grid-cols-4"
        >
          {catches.map((item, i) => (
            <FadeIn key={item.title} delay={i * 0.04}>
              <div
                className="card-flat"
                style={{
                  padding: 24,
                  transition: "transform 0.2s ease, border-color 0.2s ease",
                }}
                onMouseEnter={(e) => {
                  e.currentTarget.style.transform = "translateY(-2px)";
                  e.currentTarget.style.borderColor = "var(--border-light)";
                }}
                onMouseLeave={(e) => {
                  e.currentTarget.style.transform = "translateY(0)";
                  e.currentTarget.style.borderColor = "var(--border-subtle)";
                }}
              >
                <div
                  style={{
                    display: "flex",
                    alignItems: "center",
                    justifyContent: "space-between",
                    marginBottom: 16,
                  }}
                >
                  <div
                    style={{
                      width: 32,
                      height: 32,
                      display: "flex",
                      alignItems: "center",
                      justifyContent: "center",
                      borderRadius: 8,
                      border: "1px solid var(--border-subtle)",
                      color: "var(--text-muted)",
                    }}
                  >
                    {item.icon}
                  </div>
                  <span
                    className="micro-label"
                    style={{
                      fontSize: 9,
                      color: severityAccent[item.severity],
                    }}
                  >
                    {item.severity}
                  </span>
                </div>
                <h3
                  style={{
                    fontSize: 14,
                    fontWeight: 600,
                    color: "var(--text-primary)",
                  }}
                >
                  {item.title}
                </h3>
                <p
                  style={{
                    marginTop: 6,
                    fontSize: 12,
                    lineHeight: 1.6,
                    color: "var(--text-muted)",
                  }}
                >
                  {item.description}
                </p>
              </div>
            </FadeIn>
          ))}
        </div>

        {/* Runs everywhere section */}
        <FadeIn delay={0.3}>
          <div
            style={{
              marginTop: 64,
              padding: "40px 36px",
              background: "var(--bg-secondary)",
              borderRadius: 10,
              border: "1px solid var(--border-subtle)",
            }}
          >
            <div className="grid gap-10 md:grid-cols-2 items-center">
              <div>
                <p className="micro-label" style={{ marginBottom: 12 }}>
                  Runs everywhere
                </p>
                <h3
                  className="font-editorial"
                  style={{
                    fontSize: "clamp(1.3rem, 3vw, 1.8rem)",
                    lineHeight: 1.15,
                    color: "var(--text-primary)",
                  }}
                >
                  Works with your stack,
                  <br />
                  <span className="font-editorial-italic">not against it</span>
                </h3>
                <p
                  style={{
                    marginTop: 14,
                    fontSize: 14,
                    lineHeight: 1.7,
                    color: "var(--text-muted)",
                  }}
                >
                  A single binary with zero dependencies. Drop it into your
                  CI/CD pipeline — standard exit codes mean your workflow blocks
                  automatically on breaking changes.
                </p>
              </div>

              <div className="grid grid-cols-3 gap-3">
                {platforms.map((p) => (
                  <div
                    key={p.name}
                    className="card-flat"
                    style={{
                      display: "flex",
                      flexDirection: "column",
                      alignItems: "center",
                      gap: 8,
                      padding: 16,
                      textAlign: "center",
                    }}
                  >
                    <div
                      style={{
                        width: 36,
                        height: 36,
                        display: "flex",
                        alignItems: "center",
                        justifyContent: "center",
                        borderRadius: 8,
                        border: "1px solid var(--border-subtle)",
                        color: "var(--text-muted)",
                      }}
                    >
                      {p.icon}
                    </div>
                    <span
                      style={{
                        fontSize: 12,
                        fontWeight: 500,
                        color: "var(--text-secondary)",
                      }}
                    >
                      {p.name}
                    </span>
                  </div>
                ))}
              </div>
            </div>
          </div>
        </FadeIn>
      </div>
    </section>
  );
}
