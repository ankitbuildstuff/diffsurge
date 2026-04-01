"use client";

import Link from "next/link";
import { siteConfig } from "@/lib/constants";
import { FadeIn } from "@/components/ui/fade-in";
import {
  Terminal,
  Download,
  FileJson,
  Play,
  Server,
  Shield,
  ArrowLeft,
  ChevronRight,
} from "lucide-react";

/* ─── Section component ─── */
function Section({
  icon,
  title,
  children,
  id,
  index = 0,
}: {
  icon: React.ReactNode;
  title: string;
  children: React.ReactNode;
  id: string;
  index?: number;
}) {
  return (
    <FadeIn delay={0.05 * index}>
      <section
        id={id}
        className="scroll-mt-24"
        style={{
          paddingBottom: 48,
          borderBottom: "1px solid var(--border-subtle)",
        }}
      >
        <div
          style={{
            display: "flex",
            alignItems: "center",
            gap: 12,
            marginBottom: 20,
          }}
        >
          <div
            style={{
              width: 36,
              height: 36,
              borderRadius: 10,
              background: "var(--bg-secondary)",
              border: "1px solid var(--border-subtle)",
              display: "flex",
              alignItems: "center",
              justifyContent: "center",
              color: "var(--text-muted)",
            }}
          >
            {icon}
          </div>
          <h2
            style={{
              fontSize: 22,
              fontWeight: 500,
              color: "var(--text-primary)",
            }}
          >
            {title}
          </h2>
        </div>
        <div
          style={{
            color: "var(--text-secondary)",
            fontSize: 14,
            lineHeight: 1.7,
            display: "flex",
            flexDirection: "column",
            gap: 16,
          }}
        >
          {children}
        </div>
      </section>
    </FadeIn>
  );
}

/* ─── Code block (research terminal style) ─── */
function Code({ children }: { children: string }) {
  return (
    <div className="terminal" style={{ marginTop: 4 }}>
      <div className="terminal-header">
        <div className="terminal-dot" style={{ background: "#ff5f57" }} />
        <div className="terminal-dot" style={{ background: "#febc2e" }} />
        <div className="terminal-dot" style={{ background: "#28c840" }} />
        <span
          style={{
            marginLeft: 10,
            fontFamily: "var(--font-mono)",
            fontSize: 11,
            color: "#666",
          }}
        >
          terminal
        </span>
      </div>
      <pre
        style={{
          padding: "16px 18px",
          fontFamily: "var(--font-mono)",
          fontSize: 12,
          lineHeight: 1.9,
          color: "#ccc",
          margin: 0,
          overflowX: "auto",
        }}
      >
        {children}
      </pre>
    </div>
  );
}

/* ─── Inline code ─── */
function InlineCode({ children }: { children: string }) {
  return (
    <code
      style={{
        borderRadius: 5,
        background: "var(--bg-secondary)",
        padding: "2px 7px",
        fontSize: 13,
        fontFamily: "var(--font-mono)",
        color: "var(--text-primary)",
        border: "1px solid var(--border-subtle)",
      }}
    >
      {children}
    </code>
  );
}

/* ─── Flag item ─── */
function Flag({
  name,
  description,
}: {
  name: string;
  description: string;
}) {
  return (
    <div
      style={{
        display: "flex",
        gap: 8,
        fontSize: 13,
        lineHeight: 1.6,
      }}
    >
      <InlineCode>{name}</InlineCode>
      <span style={{ color: "var(--text-muted)" }}>— {description}</span>
    </div>
  );
}

/* ─── TOC ─── */
const tocItems = [
  { label: "Installation", href: "#installation", icon: <Download size={12} /> },
  { label: "Quick Start", href: "#quick-start", icon: <Terminal size={12} /> },
  { label: "Schema Diff", href: "#schema-diff", icon: <Shield size={12} /> },
  { label: "JSON Diff", href: "#json-diff", icon: <FileJson size={12} /> },
  { label: "Traffic Proxy", href: "#traffic-proxy", icon: <Server size={12} /> },
  { label: "Replay Engine", href: "#replay-engine", icon: <Play size={12} /> },
  { label: "CI/CD Integration", href: "#cicd", icon: <Terminal size={12} /> },
];

export default function DocsPage() {
  return (
    <div
      style={{
        minHeight: "100vh",
        background: "var(--bg-primary)",
      }}
    >
      {/* ─── Header ─── */}
      <header
        style={{
          position: "sticky",
          top: 0,
          zIndex: 40,
          background: "rgba(255, 255, 255, 0.92)",
          backdropFilter: "blur(16px)",
          borderBottom: "1px solid var(--border-subtle)",
        }}
      >
        <div
          className="mx-auto"
          style={{
            maxWidth: 1120,
            height: 56,
            display: "flex",
            alignItems: "center",
            gap: 16,
            padding: "0 24px",
          }}
        >
          <Link
            href="/"
            style={{
              display: "flex",
              alignItems: "center",
              gap: 8,
              fontSize: 13,
              color: "var(--text-muted)",
              textDecoration: "none",
              transition: "color 0.2s",
            }}
          >
            <ArrowLeft size={14} />
            Back
          </Link>
          <div
            style={{
              width: 1,
              height: 16,
              background: "var(--border-light)",
            }}
          />
          <Link
            href="/"
            style={{
              display: "flex",
              alignItems: "center",
              gap: 8,
              textDecoration: "none",
            }}
          >
            <svg width="22" height="22" viewBox="0 0 28 28" fill="none">
              <rect width="28" height="28" rx="6" fill="#1A1714" />
              <path d="M7 10l7-4 7 4-7 4-7-4z" fill="#A1A1AA" />
              <path d="M7 14l7 4 7-4" stroke="#fff" strokeWidth="1.5" />
              <path d="M7 18l7 4 7-4" stroke="#71717A" strokeWidth="1.5" />
            </svg>
            <span
              className="font-medium"
              style={{ fontSize: 16, letterSpacing: "-0.02em" }}
            >
              {siteConfig.name}
            </span>
            <span
              style={{
                fontSize: 13,
                color: "var(--text-muted)",
                fontWeight: 400,
              }}
            >
              Docs
            </span>
          </Link>

          <div style={{ flex: 1 }} />
        </div>
      </header>

      {/* ─── Main Content ─── */}
      <div
        className="mx-auto"
        style={{
          position: "relative",
          zIndex: 1,
          maxWidth: 1120,
          padding: "40px 24px 80px",
        }}
      >
        <div
          style={{
            display: "grid",
            gap: 48,
            gridTemplateColumns: "1fr",
          }}
          className="md:grid-cols-[240px_1fr]!"
        >
          {/* ─── Sidebar TOC ─── */}
          <aside className="hidden md:block">
            <nav
              style={{
                position: "sticky",
                top: 80,
              }}
            >
              <p className="micro-label" style={{ marginBottom: 16 }}>
                On this page
              </p>
              <div
                style={{
                  display: "flex",
                  flexDirection: "column",
                  gap: 2,
                }}
              >
                {tocItems.map((item) => (
                  <a
                    key={item.href}
                    href={item.href}
                    style={{
                      display: "flex",
                      alignItems: "center",
                      gap: 8,
                      padding: "8px 12px",
                      borderRadius: 8,
                      fontSize: 13,
                      color: "var(--text-muted)",
                      textDecoration: "none",
                      transition: "all 0.2s ease",
                    }}
                    onMouseEnter={(e) => {
                      e.currentTarget.style.background = "var(--bg-secondary)";
                      e.currentTarget.style.color = "var(--text-primary)";
                    }}
                    onMouseLeave={(e) => {
                      e.currentTarget.style.background = "transparent";
                      e.currentTarget.style.color = "var(--text-muted)";
                    }}
                  >
                    <span style={{ opacity: 0.5 }}>{item.icon}</span>
                    {item.label}
                  </a>
                ))}
              </div>

              {/* Quick links card */}
              <div
                className="card"
                style={{
                  marginTop: 24,
                  padding: 16,
                }}
              >
                <p
                  className="micro-label"
                  style={{
                    fontSize: 9,
                    marginBottom: 12,
                  }}
                >
                  Resources
                </p>
                <div
                  style={{
                    display: "flex",
                    flexDirection: "column",
                    gap: 8,
                  }}
                >
                  <a
                    href="https://www.npmjs.com/package/diffsurge"
                    target="_blank"
                    rel="noopener noreferrer"
                    style={{
                      fontSize: 12,
                      color: "var(--text-secondary)",
                      textDecoration: "none",
                      display: "flex",
                      alignItems: "center",
                      gap: 6,
                      transition: "color 0.2s",
                    }}
                    onMouseEnter={(e) =>
                      (e.currentTarget.style.color = "var(--text-primary)")
                    }
                    onMouseLeave={(e) =>
                      (e.currentTarget.style.color = "var(--text-secondary)")
                    }
                  >
                    <ChevronRight size={10} />
                    npm Package
                  </a>
                  <a
                    href="https://hub.docker.com/u/equixankit"
                    target="_blank"
                    rel="noopener noreferrer"
                    style={{
                      fontSize: 12,
                      color: "var(--text-secondary)",
                      textDecoration: "none",
                      display: "flex",
                      alignItems: "center",
                      gap: 6,
                      transition: "color 0.2s",
                    }}
                    onMouseEnter={(e) =>
                      (e.currentTarget.style.color = "var(--text-primary)")
                    }
                    onMouseLeave={(e) =>
                      (e.currentTarget.style.color = "var(--text-secondary)")
                    }
                  >
                    <ChevronRight size={10} />
                    Docker Hub
                  </a>
                </div>
              </div>
            </nav>
          </aside>

          {/* ─── Content ─── */}
          <main
            style={{
              minWidth: 0,
              display: "flex",
              flexDirection: "column",
              gap: 48,
            }}
          >
            {/* Hero section */}
            <FadeIn>
              <div>
                <p className="micro-label" style={{ marginBottom: 16 }}>
                  CLI Reference
                </p>
                <h1
                  style={{
                    fontSize: "clamp(1.8rem, 4vw, 2.8rem)",
                    fontWeight: 500,
                    lineHeight: 1.1,
                    letterSpacing: "-0.02em",
                    color: "var(--text-primary)",
                  }}
                >
                  Documentation
                </h1>
                <p
                  style={{
                    marginTop: 16,
                    fontSize: 15,
                    lineHeight: 1.7,
                    color: "var(--text-secondary)",
                    maxWidth: 560,
                  }}
                >
                  Diffsurge is a CLI tool and infrastructure for catching breaking
                  API changes. Compare schemas and replay captured requests
                  against staging builds.
                </p>

              </div>
            </FadeIn>

            {/* ─── Installation ─── */}
            <Section
              id="installation"
              icon={<Download size={16} />}
              title="Installation"
              index={1}
            >
              <p>Install the Surge CLI globally via npm:</p>
              <Code>{`npm install -g diffsurge`}</Code>
              <p>Or use Docker (no install required):</p>
              <Code>{`docker run equixankit/diffsurge-cli --help`}</Code>
              <p>Verify the installation:</p>
              <Code>{`surge --help`}</Code>
              <p>
                Available platforms: <InlineCode>macOS (Intel & Apple Silicon)</InlineCode>,{" "}
                <InlineCode>Linux (x64, ARM64)</InlineCode>,{" "}
                <InlineCode>Windows (x64)</InlineCode>.
              </p>
            </Section>

            {/* ─── Quick Start ─── */}
            <Section
              id="quick-start"
              icon={<Terminal size={16} />}
              title="Quick Start"
              index={2}
            >
              <p>
                Compare two API schemas and detect breaking changes in under 30 seconds:
              </p>
              <Code>{`# Compare two OpenAPI schemas
surge schema diff \\
  --old api-v1.yaml \\
  --new api-v2.yaml \\
  --fail-on-breaking

# Compare two JSON response files
surge diff --old response-v1.json --new response-v2.json

# Replay captured traffic against staging
surge replay \\
  --source traffic.json \\
  --target http://staging.example.com`}</Code>
            </Section>

            {/* ─── Schema Diff ─── */}
            <Section
              id="schema-diff"
              icon={<Shield size={16} />}
              title="Schema Diff"
              index={3}
            >
              <p>
                Compare two OpenAPI 3.x schema files and detect breaking changes
                like removed endpoints, type changes, and new required parameters.
              </p>
              <Code>{`surge schema diff \\
  --old api-v1.yaml \\
  --new api-v2.yaml \\
  --fail-on-breaking \\
  --format text`}</Code>
              <p
                style={{
                  fontWeight: 500,
                  color: "var(--text-primary)",
                  fontSize: 13,
                  letterSpacing: "0.02em",
                  textTransform: "uppercase",
                  marginTop: 8,
                }}
              >
                Flags
              </p>
              <div style={{ display: "flex", flexDirection: "column", gap: 8 }}>
                <Flag name="--old" description="Path to the old/original schema file (required)" />
                <Flag name="--new" description="Path to the new/modified schema file (required)" />
                <Flag name="--fail-on-breaking" description="Exit with code 1 if breaking changes found" />
                <Flag name="--breaking-only" description="Show only breaking changes" />
                <Flag name="--format" description="Output format: text or json" />
                <Flag name="--output" description="Write output to a file" />
              </div>
              <p
                style={{
                  fontWeight: 500,
                  color: "var(--text-primary)",
                  fontSize: 13,
                  letterSpacing: "0.02em",
                  textTransform: "uppercase",
                  marginTop: 8,
                }}
              >
                Exit codes
              </p>
              <div style={{ display: "flex", flexDirection: "column", gap: 6 }}>
                <div style={{ fontSize: 13 }}>
                  <InlineCode>0</InlineCode>{" "}
                  <span style={{ color: "var(--text-muted)" }}>— No breaking changes</span>
                </div>
                <div style={{ fontSize: 13 }}>
                  <InlineCode>1</InlineCode>{" "}
                  <span style={{ color: "var(--text-muted)" }}>
                    — Breaking changes detected (with <InlineCode>--fail-on-breaking</InlineCode>)
                  </span>
                </div>
                <div style={{ fontSize: 13 }}>
                  <InlineCode>2</InlineCode>{" "}
                  <span style={{ color: "var(--text-muted)" }}>— Error occurred</span>
                </div>
              </div>
            </Section>

            {/* ─── JSON Diff ─── */}
            <Section
              id="json-diff"
              icon={<FileJson size={16} />}
              title="JSON Diff"
              index={4}
            >
              <p>
                Compare two JSON files and produce a detailed diff report with
                type changes, additions, and removals.
              </p>
              <Code>{`surge diff \\
  --old response-v1.json \\
  --new response-v2.json \\
  --array-as-set \\
  --format json`}</Code>
              <p
                style={{
                  fontWeight: 500,
                  color: "var(--text-primary)",
                  fontSize: 13,
                  letterSpacing: "0.02em",
                  textTransform: "uppercase",
                  marginTop: 8,
                }}
              >
                Flags
              </p>
              <div style={{ display: "flex", flexDirection: "column", gap: 8 }}>
                <Flag name="--old" description="Path to the old JSON file (required)" />
                <Flag name="--new" description="Path to the new JSON file (required)" />
                <Flag name="--array-as-set" description="Compare arrays as sets (ignore order)" />
                <Flag name="--ignore" description="JSON paths to ignore (comma-separated)" />
                <Flag name="--format" description="Output format: text or json" />
              </div>
            </Section>

            {/* ─── Traffic Proxy ─── */}
            <Section
              id="traffic-proxy"
              icon={<Server size={16} />}
              title="Traffic Proxy"
              index={5}
            >
              <p>
                Deploy the Diffsurge proxy to capture production traffic. It
                runs as a reverse proxy, sampling request/response pairs with
                automatic PII redaction.
              </p>
              <Code>{`docker run -d \\
  -e DIFFSURGE_STORAGE_POSTGRES_URL=postgresql://... \\
  -e DIFFSURGE_STORAGE_REDIS_URL=rediss://... \\
  -p 8081:8080 \\
  equixankit/diffsurge-proxy`}</Code>
              <p>
                The proxy adds less than 5ms of latency at p95. It uses async
                buffering so the forwarding path is never blocked by storage I/O.
                Configure sampling rates to capture 1–100% of traffic.
              </p>
            </Section>

            {/* ─── Replay Engine ─── */}
            <Section
              id="replay-engine"
              icon={<Play size={16} />}
              title="Replay Engine"
              index={6}
            >
              <p>
                Replay captured traffic against a target server and compare
                every response semantically.
              </p>
              <Code>{`surge replay \\
  --source traffic.json \\
  --target http://staging.example.com \\
  --workers 20 \\
  --rate-limit 500 \\
  --format json \\
  --output drift-report.json`}</Code>
              <p
                style={{
                  fontWeight: 500,
                  color: "var(--text-primary)",
                  fontSize: 13,
                  letterSpacing: "0.02em",
                  textTransform: "uppercase",
                  marginTop: 8,
                }}
              >
                Flags
              </p>
              <div style={{ display: "flex", flexDirection: "column", gap: 8 }}>
                <Flag name="--source" description="Traffic JSON file (required)" />
                <Flag name="--target" description="Target server URL (required)" />
                <Flag name="--workers" description="Concurrent workers (default: 10)" />
                <Flag name="--rate-limit" description="Max requests per second (0 = unlimited)" />
                <Flag name="--timeout" description="Per-request timeout (default: 30s)" />
                <Flag name="--max-retries" description="Max retries per request (default: 2)" />
              </div>
            </Section>

            {/* ─── CI/CD ─── */}
            <Section
              id="cicd"
              icon={<Terminal size={16} />}
              title="CI/CD Integration"
              index={7}
            >
              <p>
                Add Diffsurge to your CI/CD pipeline to automatically block
                deploys with breaking changes.
              </p>
              <p
                style={{
                  fontWeight: 500,
                  color: "var(--text-primary)",
                  fontSize: 13,
                  letterSpacing: "0.02em",
                  textTransform: "uppercase",
                }}
              >
                GitHub Actions
              </p>
              <Code>{`# .github/workflows/api-check.yml
name: API Schema Check
on: [pull_request]

jobs:
  schema-diff:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4

      - name: Install Surge CLI
        run: npm install -g diffsurge

      - name: Check for breaking changes
        run: |
          surge schema diff \\
            --old api/schema-main.yaml \\
            --new api/schema.yaml \\
            --fail-on-breaking`}</Code>
              <p>
                The CLI returns standard exit codes so your pipeline blocks
                automatically: <InlineCode>0</InlineCode> = clean,{" "}
                <InlineCode>1</InlineCode> = breaking changes,{" "}
                <InlineCode>2</InlineCode> = error.
              </p>
            </Section>

            {/* ─── Footer CTA ─── */}
            <FadeIn delay={0.3}>
              <div
                className="card"
                style={{
                  padding: 32,
                  textAlign: "center",
                }}
              >
                <p className="micro-label" style={{ marginBottom: 12 }}>
                  Ready to get started?
                </p>
                <p
                  style={{
                    fontSize: 20,
                    fontWeight: 500,
                    color: "var(--text-primary)",
                    marginBottom: 20,
                  }}
                >
                  Start catching breaking changes today
                </p>
                <div
                  style={{
                    display: "flex",
                    justifyContent: "center",
                    gap: 12,
                    flexWrap: "wrap",
                  }}
                >
                  <Link href="/signup" className="btn-primary">
                    Get Started Free
                  </Link>
                </div>
              </div>
            </FadeIn>
          </main>
        </div>
      </div>
    </div>
  );
}
