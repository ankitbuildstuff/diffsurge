"use client";

import { FadeIn } from "@/components/ui/fade-in";
import { Smartphone, Blocks, Monitor, Server } from "lucide-react";

const impacts = [
  {
    icon: Smartphone,
    title: "Mobile apps",
    desc: "Crash from missing response fields",
  },
  {
    icon: Blocks,
    title: "Third-party integrations",
    desc: "Break from unexpected type changes",
  },
  {
    icon: Monitor,
    title: "Frontend clients",
    desc: "Fail from schema mismatches",
  },
  {
    icon: Server,
    title: "Internal services",
    desc: "Error from endpoint changes",
  },
];

export function Problem() {
  return (
    <section
      className="section-spacing"
      style={{ background: "var(--bg-secondary)" }}
    >
      <div className="mx-auto px-6" style={{ maxWidth: 1200 }}>
        {/* Top row: text left, terminal right */}
        <div className="grid gap-16 lg:grid-cols-2 items-center">
          <div>
            <FadeIn>
              <p className="micro-label" style={{ marginBottom: 16 }}>
                The Problem
              </p>
              <h2
                style={{
                  fontSize: "clamp(1.8rem, 4vw, 2.5rem)",
                  fontWeight: 500,
                  lineHeight: 1.1,
                  letterSpacing: "-0.02em",
                }}
              >
                APIs break more often than you think.
              </h2>
              <p
                style={{
                  marginTop: 20,
                  fontSize: 16,
                  lineHeight: 1.7,
                  color: "var(--text-secondary)",
                }}
              >
                Most API tests rely on synthetic test cases. But production
                traffic contains thousands of edge cases your tests never cover.
              </p>
              <p
                style={{
                  marginTop: 16,
                  fontSize: 15,
                  lineHeight: 1.7,
                  color: "var(--text-muted)",
                }}
              >
                A small API change can break everything downstream. DiffSurge
                catches these issues before they reach production.
              </p>
            </FadeIn>
          </div>

          {/* Terminal on the right */}
          <FadeIn delay={0.15}>
            <div className="terminal">
              <div className="terminal-header">
                <div
                  className="terminal-dot"
                  style={{ background: "#ff5f57" }}
                />
                <div
                  className="terminal-dot"
                  style={{ background: "#febc2e" }}
                />
                <div
                  className="terminal-dot"
                  style={{ background: "#28c840" }}
                />
                <span
                  style={{
                    marginLeft: 10,
                    fontFamily: "var(--font-mono)",
                    fontSize: 11,
                    color: "#666",
                  }}
                >
                  deployment log
                </span>
              </div>
              <div
                style={{
                  padding: "16px 20px",
                  fontFamily: "var(--font-mono)",
                  fontSize: 12,
                  lineHeight: 2,
                }}
              >
                <p style={{ color: "var(--accent-green)" }}>
                  ✓ API v2.1 deployed to production
                </p>
                <p style={{ color: "#888" }}>
                  &nbsp; Response field &quot;email_verified&quot; removed
                </p>
                <p style={{ color: "#888" }}>&nbsp; ...</p>
                <p style={{ color: "var(--accent-red)" }}>
                  ✗ Mobile app crash rate: 0.1% → 12.4%
                </p>
                <p style={{ color: "var(--accent-red)" }}>
                  ✗ Partner webhook failures: 847 errors
                </p>
                <p style={{ color: "var(--accent-amber)" }}>
                  ⚠ Rollback initiated — 23 min of downtime
                </p>
              </div>
            </div>
          </FadeIn>
        </div>

        {/* Impact cards below as full-width row */}
        <FadeIn delay={0.25}>
          <div className="grid grid-cols-2 lg:grid-cols-4 gap-4" style={{ marginTop: 48 }}>
            {impacts.map((item, i) => (
              <div
                key={i}
                className="card"
                style={{ padding: 24, cursor: "default" }}
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
                  <item.icon
                    size={18}
                    strokeWidth={1.5}
                    style={{ color: "var(--text-muted)" }}
                  />
                </div>
                <h3
                  style={{
                    fontSize: 14,
                    fontWeight: 500,
                    color: "var(--text-primary)",
                  }}
                >
                  {item.title}
                </h3>
                <p
                  style={{
                    fontSize: 13,
                    color: "var(--text-muted)",
                    marginTop: 6,
                    lineHeight: 1.6,
                  }}
                >
                  {item.desc}
                </p>
              </div>
            ))}
          </div>
        </FadeIn>
      </div>
    </section>
  );
}
