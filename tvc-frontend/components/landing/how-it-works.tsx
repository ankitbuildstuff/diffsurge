"use client";

import { FadeIn } from "@/components/ui/fade-in";

const steps = [
  {
    number: "01",
    title: "Install the CLI",
    description:
      "One command, zero dependencies. Install the Surge CLI globally via npm or run it directly with Docker. Works on macOS, Linux, and Windows.",
    code: `$ npm install -g diffsurge

  added 1 package in 2.1s

$ surge --help

  surge — Catch breaking API changes
  before your users do

  Commands:
    diff       Compare two schema files
    schema     Schema management
    replay     Replay traffic
    version    Print version`,
  },
  {
    number: "02",
    title: "Diff your API schemas",
    description:
      "Surge compares OpenAPI, GraphQL, and gRPC schemas — flagging every breaking change with severity and a JSON path.",
    code: `$ surge schema diff \\
    --old api-v1.yaml \\
    --new api-v2.yaml

  Comparing 47 endpoints…

  ✗ BREAKING  POST /api/users
    └─ Required field removed: "email_verified"
  ⚠ WARNING   GET /api/users/:id
    └─ Type changed: "age" string → number
  ✓ SAFE      45 endpoints unchanged

  1 breaking · 1 warning — exit code 1`,
  },
  {
    number: "03",
    title: "Capture production traffic",
    description:
      "Deploy the proxy as a sidecar or standalone container. It samples real traffic, strips PII, and buffers asynchronously.",
    code: `$ docker run -d \\
    -e TVC_STORAGE_POSTGRES_URL=... \\
    -e TVC_STORAGE_REDIS_URL=... \\
    -p 8081:8080 \\
    equixankit/diffsurge-proxy

  ▸ Proxy listening on :8080
  ▸ Sampling 10% of traffic
  ▸ PII redaction: enabled
  ▸ Buffer: 10,000 slots / 20 workers`,
  },
  {
    number: "04",
    title: "Replay and compare",
    description:
      "Point the replay engine at your staging build. It fires captured requests, compares every response, and produces a drift report.",
    code: `$ surge replay \\
    --source traffic.json \\
    --target http://staging:8080

  Replaying 1,247 requests...

  ✓ 1,241 responses matched (99.5%)
  ⚠ 4 warnings  (type coercion)
  ✗ 2 breaking   (missing fields)
  ▸ Report saved to drift-report.json`,
  },
];

export function HowItWorks() {
  return (
    <section
      id="how-it-works"
      className="section-spacing"
      style={{ background: "var(--bg-secondary)" }}
    >
      <div className="mx-auto max-w-[1120px] px-6">
        <FadeIn>
          <div style={{ maxWidth: 480 }}>
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
              <span>How it works</span>
            </div>
            <h2
              className="font-editorial"
              style={{
                fontSize: "clamp(1.8rem, 4vw, 2.6rem)",
                lineHeight: 1.1,
                color: "var(--text-primary)",
              }}
            >
              Four steps to{" "}
              <span className="font-editorial-italic">safer deployments</span>
            </h2>
            <p
              style={{
                marginTop: 16,
                fontSize: 14,
                lineHeight: 1.7,
                color: "var(--text-muted)",
              }}
            >
              Go from zero to production-grade API regression testing in
              minutes.
            </p>
          </div>
        </FadeIn>

        <div style={{ marginTop: 56, display: "flex", flexDirection: "column", gap: 20 }}>
          {steps.map((step, i) => (
            <FadeIn key={step.number} delay={i * 0.08}>
              <div
                className="card-flat grid gap-6 p-6 md:p-7 md:grid-cols-5"
                style={{ overflow: "hidden" }}
              >
                {/* Text */}
                <div className="md:col-span-2">
                  <div
                    style={{
                      display: "flex",
                      alignItems: "center",
                      gap: 12,
                    }}
                  >
                    <span
                      style={{
                        display: "inline-flex",
                        alignItems: "center",
                        justifyContent: "center",
                        width: 32,
                        height: 32,
                        borderRadius: "50%",
                        border: "1px solid var(--border-light)",
                        fontSize: 12,
                        fontWeight: 600,
                        color: "var(--text-faint)",
                        fontFamily: "var(--font-mono)",
                      }}
                    >
                      {step.number}
                    </span>
                    <h3
                      style={{
                        fontSize: 16,
                        fontWeight: 600,
                        color: "var(--text-primary)",
                      }}
                    >
                      {step.title}
                    </h3>
                  </div>
                  <p
                    style={{
                      marginTop: 12,
                      fontSize: 13,
                      lineHeight: 1.7,
                      color: "var(--text-muted)",
                    }}
                  >
                    {step.description}
                  </p>
                </div>

                {/* Code */}
                <div className="md:col-span-3" style={{ minWidth: 0 }}>
                  <div className="terminal-research">
                    <div className="terminal-research-header">
                      <div className="dot" />
                      <div className="dot" />
                      <div className="dot" />
                    </div>
                    <pre
                      className="hiw-code-pre"
                      style={{
                        padding: "14px 18px",
                        fontFamily: "var(--font-mono)",
                        fontSize: 11.5,
                        lineHeight: 1.8,
                        color: "rgba(255,255,255,0.5)",
                        overflowX: "auto",
                        margin: 0,
                        maxWidth: "100%",
                      }}
                    >
                      {step.code}
                    </pre>
                  </div>
                </div>
              </div>
            </FadeIn>
          ))}
        </div>
      </div>
    </section>
  );
}
