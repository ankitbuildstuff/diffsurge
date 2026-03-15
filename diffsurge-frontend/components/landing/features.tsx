"use client";

import { FadeIn } from "@/components/ui/fade-in";
import {
  Radio,
  GitCompare,
  Zap,
  ShieldCheck,
  GitBranch,
  Terminal,
} from "lucide-react";

const features = [
  {
    icon: Radio,
    title: "Real Traffic Replay",
    description:
      "Replay real production requests against your staging API. Catch the edge cases synthetic tests miss.",
  },
  {
    icon: GitCompare,
    title: "Schema Diff Detection",
    description:
      "Detect breaking changes in OpenAPI, GraphQL, and gRPC schemas automatically.",
  },
  {
    icon: Zap,
    title: "High Performance Engine",
    description:
      "Replay thousands of requests per second with configurable concurrency and faithful timing.",
  },
  {
    icon: ShieldCheck,
    title: "Privacy Safe",
    description:
      "Automatic PII masking for emails, phones, credit cards, and SSNs before any data is stored.",
  },
  {
    icon: GitBranch,
    title: "CI Integration",
    description:
      "Block deployments when breaking changes are detected. Standard exit codes for any CI/CD pipeline.",
  },
  {
    icon: Terminal,
    title: "Open Source CLI",
    description:
      "Fully open source core and CLI. MIT licensed. Run anywhere \u2014 macOS, Linux, Windows, Docker.",
  },
];

export function Features() {
  return (
    <section
      id="features"
      className="section-spacing"
      style={{ background: "var(--bg-secondary)" }}
    >
      <div className="mx-auto px-6" style={{ maxWidth: 1200 }}>
        <FadeIn>
          <div
            style={{ textAlign: "center", maxWidth: 560, margin: "0 auto" }}
          >
            <p className="micro-label" style={{ marginBottom: 16 }}>
              Features
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
              Everything you need to ship without fear
            </h2>
            <p
              style={{
                marginTop: 16,
                fontSize: 15,
                lineHeight: 1.7,
                color: "var(--text-muted)",
              }}
            >
              From static schema checks in CI to live traffic replay — a
              complete safety net for every API change.
            </p>
          </div>
        </FadeIn>

        <div className="mt-16 grid gap-5 sm:grid-cols-2 lg:grid-cols-3">
          {features.map((feature, i) => (
            <FadeIn key={feature.title} delay={i * 0.06}>
              <div
                className="card h-full"
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
                    marginBottom: 16,
                  }}
                >
                  <feature.icon
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
                    lineHeight: 1.3,
                  }}
                >
                  {feature.title}
                </h3>
                <p
                  style={{
                    marginTop: 8,
                    fontSize: 14,
                    lineHeight: 1.7,
                    color: "var(--text-muted)",
                  }}
                >
                  {feature.description}
                </p>
              </div>
            </FadeIn>
          ))}
        </div>
      </div>
    </section>
  );
}
