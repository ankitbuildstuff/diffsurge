"use client";

import { FadeIn } from "@/components/ui/fade-in";
import { Github, Star, GitFork, Users, Scale } from "lucide-react";
import { siteConfig } from "@/lib/constants";

const highlights = [
  {
    icon: Scale,
    label: "MIT Licensed",
    description: "Free to use, modify, and distribute",
  },
  {
    icon: GitFork,
    label: "Fork & Contribute",
    description: "Open architecture, PRs welcome",
  },
  {
    icon: Users,
    label: "Community Driven",
    description: "Built by developers, for developers",
  },
];

function hashIntensity(week: number, day: number): number {
  const n = ((week * 7 + day) * 2654435761) >>> 0;
  return (n % 100) / 100;
}

export function OpenSource() {
  return (
    <section
      className="section-spacing"
      style={{ background: "var(--bg-secondary)" }}
    >
      <div className="mx-auto px-6" style={{ maxWidth: 1200 }}>
        <div className="grid gap-16 lg:grid-cols-2 items-center">
          <FadeIn>
            <div>
              <p className="micro-label" style={{ marginBottom: 16 }}>
                Open Source
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
                Built in the open. Transparent by default.
              </h2>
              <p
                style={{
                  marginTop: 20,
                  fontSize: 16,
                  lineHeight: 1.7,
                  color: "var(--text-secondary)",
                }}
              >
                DiffSurge is open source and developer-driven. The CLI and
                schema diffing engine are MIT licensed — free forever with no
                limits.
              </p>
              <ul
                style={{
                  marginTop: 24,
                  display: "flex",
                  flexDirection: "column",
                  gap: 12,
                }}
              >
                {[
                  "MIT licensed CLI and core engine",
                  "Transparent architecture and roadmap",
                  "Community contributions welcome",
                  "Self-host or use managed service",
                ].map((item) => (
                  <li
                    key={item}
                    style={{
                      display: "flex",
                      alignItems: "center",
                      gap: 10,
                      fontSize: 14,
                      color: "var(--text-muted)",
                    }}
                  >
                    <div
                      style={{
                        width: 5,
                        height: 5,
                        borderRadius: "50%",
                        background: "var(--text-faint)",
                        flexShrink: 0,
                      }}
                    />
                    {item}
                  </li>
                ))}
              </ul>
              <div
                style={{
                  marginTop: 32,
                  display: "flex",
                  flexWrap: "wrap",
                  gap: 12,
                }}
              >
                <a
                  href={siteConfig.github}
                  target="_blank"
                  rel="noopener noreferrer"
                  className="btn-primary"
                >
                  <Github size={16} />
                  <Star size={14} />
                  Star on GitHub
                </a>
                <a
                  href={siteConfig.github}
                  target="_blank"
                  rel="noopener noreferrer"
                  className="btn-secondary"
                >
                  View Source
                </a>
              </div>
            </div>
          </FadeIn>

          <FadeIn delay={0.15}>
            <div className="grid gap-4">
              {highlights.map((h, i) => (
                <div
                  key={i}
                  className="card flex items-start gap-4"
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
                      flexShrink: 0,
                    }}
                  >
                    <h.icon
                      size={18}
                      strokeWidth={1.5}
                      style={{ color: "var(--text-muted)" }}
                    />
                  </div>
                  <div>
                    <h3
                      style={{
                        fontSize: 15,
                        fontWeight: 500,
                        color: "var(--text-primary)",
                      }}
                    >
                      {h.label}
                    </h3>
                    <p
                      style={{
                        fontSize: 14,
                        color: "var(--text-muted)",
                        marginTop: 4,
                      }}
                    >
                      {h.description}
                    </p>
                  </div>
                </div>
              ))}

              {/* GitHub activity visual */}
              <div
                className="card"
                style={{ padding: 24, cursor: "default" }}
              >
                <p
                  style={{
                    fontSize: 11,
                    fontWeight: 500,
                    color: "var(--text-faint)",
                    marginBottom: 12,
                    textTransform: "uppercase",
                    letterSpacing: "0.05em",
                  }}
                >
                  Contribution Activity
                </p>
                <div
                  style={{
                    display: "flex",
                    gap: 3,
                    overflow: "hidden",
                  }}
                >
                  {Array.from({ length: 40 }, (_, week) => (
                    <div
                      key={week}
                      style={{
                        display: "flex",
                        flexDirection: "column",
                        gap: 3,
                      }}
                    >
                      {Array.from({ length: 7 }, (_, day) => {
                        const intensity = hashIntensity(week, day);
                        return (
                          <div
                            key={day}
                            style={{
                              width: 10,
                              height: 10,
                              borderRadius: 2,
                              background:
                                intensity > 0.7
                                  ? "#22c55e"
                                  : intensity > 0.4
                                    ? "rgba(34,197,94,0.4)"
                                    : intensity > 0.15
                                      ? "rgba(34,197,94,0.15)"
                                      : "var(--bg-tertiary)",
                            }}
                          />
                        );
                      })}
                    </div>
                  ))}
                </div>
              </div>
            </div>
          </FadeIn>
        </div>
      </div>
    </section>
  );
}
