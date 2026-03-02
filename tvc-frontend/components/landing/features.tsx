"use client";

import { Shield, Radio, Play, BarChart3, Lock } from "lucide-react";
import { FadeIn } from "@/components/ui/fade-in";

/* ── Inline visuals (data-visual motif) ── */

function TerminalVisual() {
  return (
    <div className="terminal-research" style={{ marginTop: 20 }}>
      <div className="terminal-research-header">
        <div className="dot" />
        <div className="dot" />
        <div className="dot" />
      </div>
      <div
        style={{
          padding: "12px 16px",
          fontFamily: "var(--font-mono)",
          fontSize: 11,
          lineHeight: 1.8,
        }}
      >
        <p style={{ color: "rgba(255,255,255,0.55)" }}>
          <span style={{ color: "var(--accent-teal)" }}>$</span> surge schema diff --fail-on-breaking
        </p>
        <p style={{ color: "rgba(255,255,255,0.25)", marginTop: 4 }}>
          Comparing 47 endpoints…
        </p>
        <p style={{ color: "var(--accent-orange)", marginTop: 4 }}>
          ✗ BREAKING POST /api/users
        </p>
        <p style={{ color: "rgba(255,255,255,0.25)", marginLeft: 16 }}>
          └─ Required field removed: &quot;email_verified&quot;
        </p>
        <p style={{ color: "var(--accent-yellow)" }}>
          ⚠ WARNING  GET /api/users/:id
        </p>
        <p style={{ color: "rgba(255,255,255,0.25)", marginLeft: 16 }}>
          └─ Type: &quot;age&quot; string → number
        </p>
        <p style={{ color: "var(--accent-teal)" }}>
          ✓ SAFE     45 endpoints unchanged
        </p>
      </div>
    </div>
  );
}

function WaveformMiniVisual() {
  const bars = [40, 65, 28, 72, 45, 88, 35, 60, 50, 78, 22, 68, 55, 82, 38];
  const colors = ["var(--accent-purple)", "var(--accent-blue)", "var(--accent-teal)"];
  return (
    <div
      style={{
        marginTop: 20,
        display: "flex",
        alignItems: "end",
        gap: 3,
        height: 48,
        padding: "8px 0",
      }}
    >
      {bars.map((h, i) => (
        <div
          key={i}
          style={{
            flex: 1,
            height: `${h}%`,
            backgroundColor: colors[i % colors.length],
            borderRadius: 2,
            opacity: 0.6,
          }}
        />
      ))}
    </div>
  );
}

function ReplayVisual() {
  const rows = [
    { path: "POST /checkout", status: "matched", ms: 42 },
    { path: "GET  /cart", status: "drift", ms: 18 },
    { path: "PUT  /profile", status: "matched", ms: 31 },
  ];
  return (
    <div style={{ marginTop: 20, display: "flex", flexDirection: "column", gap: 6 }}>
      {rows.map((r) => (
        <div
          key={r.path}
          className="card-flat"
          style={{
            display: "flex",
            alignItems: "center",
            justifyContent: "space-between",
            padding: "10px 14px",
            fontSize: 11,
          }}
        >
          <span
            style={{
              fontFamily: "var(--font-mono)",
              color: "var(--text-secondary)",
            }}
          >
            {r.path}
          </span>
          <div style={{ display: "flex", alignItems: "center", gap: 8 }}>
            <span style={{ color: "var(--text-faint)" }}>{r.ms}ms</span>
            <span
              style={{
                padding: "2px 8px",
                borderRadius: 999,
                fontSize: 10,
                fontWeight: 500,
                background:
                  r.status === "matched"
                    ? "rgba(91,168,154,0.12)"
                    : "rgba(212,168,67,0.12)",
                color:
                  r.status === "matched"
                    ? "var(--accent-teal)"
                    : "var(--accent-yellow)",
              }}
            >
              {r.status}
            </span>
          </div>
        </div>
      ))}
    </div>
  );
}

function DashboardVisual() {
  const bars = [64, 42, 78, 52, 85, 38, 72, 58, 80, 46, 70, 55, 90, 48, 66, 74, 50, 82, 44, 76];
  const colors = [
    "var(--accent-purple)",
    "var(--accent-blue)",
    "var(--accent-teal)",
    "var(--accent-yellow)",
    "var(--accent-orange)",
    "var(--accent-rose)",
  ];
  return (
    <div style={{ marginTop: 20 }}>
      <div
        style={{
          display: "flex",
          justifyContent: "space-between",
          fontSize: 10,
          color: "var(--text-faint)",
          marginBottom: 8,
        }}
      >
        <span>Requests / min</span>
        <span style={{ fontWeight: 500, color: "var(--text-secondary)" }}>
          1,247 total
        </span>
      </div>
      <div style={{ display: "flex", alignItems: "end", gap: 2, height: 64 }}>
        {bars.map((h, i) => (
          <div
            key={i}
            style={{
              flex: 1,
              height: `${h}%`,
              backgroundColor: colors[i % colors.length],
              borderRadius: 2,
              opacity: 0.5 + h / 250,
            }}
          />
        ))}
      </div>
      <div
        style={{
          marginTop: 8,
          display: "flex",
          justifyContent: "space-between",
          fontSize: 10,
          color: "var(--text-faint)",
        }}
      >
        <span>12:00</span>
        <span>15:00</span>
        <span>18:00</span>
      </div>
    </div>
  );
}

function PiiVisual() {
  const items = [
    { field: '"email"', from: "john@acme.co", to: "***@***.co" },
    { field: '"phone"', from: "+1-555-0199", to: "+1-***-****" },
    { field: '"card"', from: "4242…4242", to: "****…****" },
  ];
  return (
    <div style={{ marginTop: 20, display: "flex", flexDirection: "column", gap: 6 }}>
      {items.map((item) => (
        <div
          key={item.field}
          className="card-flat"
          style={{
            display: "flex",
            alignItems: "center",
            justifyContent: "space-between",
            padding: "10px 14px",
            fontFamily: "var(--font-mono)",
            fontSize: 11,
          }}
        >
          <span style={{ color: "var(--text-muted)" }}>{item.field}</span>
          <div style={{ display: "flex", alignItems: "center", gap: 6 }}>
            <span
              style={{
                color: "var(--text-faint)",
                textDecoration: "line-through",
              }}
            >
              {item.from}
            </span>
            <span style={{ color: "var(--accent-orange)" }}>→</span>
            <span
              style={{
                color: "var(--accent-teal)",
                fontWeight: 600,
              }}
            >
              {item.to}
            </span>
          </div>
        </div>
      ))}
    </div>
  );
}

/* ── Feature card ── */

function Card({
  icon,
  label,
  title,
  description,
  visual,
}: {
  icon: React.ReactNode;
  label: string;
  title: string;
  description: string;
  visual?: React.ReactNode;
}) {
  return (
    <div className="card-flat" style={{ height: "100%", padding: 28 }}>
      <div
        style={{
          display: "flex",
          alignItems: "center",
          gap: 10,
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
          {icon}
        </div>
        <span className="micro-label" style={{ color: "var(--text-muted)" }}>
          {label}
        </span>
      </div>
      <h3
        style={{
          fontSize: 16,
          fontWeight: 600,
          lineHeight: 1.3,
          color: "var(--text-primary)",
        }}
      >
        {title}
      </h3>
      <p
        style={{
          marginTop: 8,
          fontSize: 13,
          lineHeight: 1.65,
          color: "var(--text-muted)",
        }}
      >
        {description}
      </p>
      {visual}
    </div>
  );
}

export function Features() {
  return (
    <section
      id="features"
      className="section-spacing"
      style={{ background: "var(--bg-primary)" }}
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
              <span>Features</span>
            </div>
            <h2
              className="font-editorial"
              style={{
                fontSize: "clamp(1.8rem, 4vw, 2.6rem)",
                lineHeight: 1.1,
                color: "var(--text-primary)",
              }}
            >
              Everything you need to ship{" "}
              <span className="font-editorial-italic">without fear</span>
            </h2>
            <p
              style={{
                marginTop: 16,
                fontSize: 14,
                lineHeight: 1.7,
                color: "var(--text-muted)",
              }}
            >
              From static schema checks in CI to live traffic replay in
              production — a complete safety net for every API change.
            </p>
          </div>
        </FadeIn>

        {/* Row 1: Schema Guardian (wide) + Traffic Proxy */}
        <div className="mt-12 grid gap-5 lg:grid-cols-3">
          <FadeIn delay={0} className="lg:col-span-2">
            <Card
              icon={<Shield size={16} strokeWidth={1.5} />}
              label="Schema Guardian"
              title="Detect breaking changes at the schema level"
              description="Parses OpenAPI 3.x, GraphQL SDL, and gRPC proto files. Compares endpoints, request params, and response shapes — flagging every breaking change with severity."
              visual={<TerminalVisual />}
            />
          </FadeIn>
          <FadeIn delay={0.06} className="lg:col-span-1">
            <Card
              icon={<Radio size={16} strokeWidth={1.5} />}
              label="Traffic Proxy"
              title="Capture production traffic without code changes"
              description="A lightweight reverse proxy that samples real request/response pairs. Async buffering keeps latency under 5 ms."
              visual={<WaveformMiniVisual />}
            />
          </FadeIn>
        </div>

        {/* Row 2: PII Redaction + Replay Engine (wide) */}
        <div className="mt-5 grid gap-5 lg:grid-cols-3">
          <FadeIn delay={0.12} className="lg:col-span-1">
            <Card
              icon={<Lock size={16} strokeWidth={1.5} />}
              label="PII Redaction"
              title="Automatic sensitive data masking"
              description="Emails, phones, credit cards, and SSNs are redacted before storage."
              visual={<PiiVisual />}
            />
          </FadeIn>
          <FadeIn delay={0.18} className="lg:col-span-2">
            <Card
              icon={<Play size={16} strokeWidth={1.5} />}
              label="Replay Engine"
              title="Replay 1,000+ requests per second against staging"
              description="Fires captured traffic at your new build with configurable concurrency. Semantic comparison ignores field order — only real differences surface."
              visual={<ReplayVisual />}
            />
          </FadeIn>
        </div>

        {/* Row 3: Dashboard (full width) */}
        <div className="mt-5">
          <FadeIn delay={0.24}>
            <Card
              icon={<BarChart3 size={16} strokeWidth={1.5} />}
              label="Live Dashboard"
              title="Real-time visibility into every request"
              description="Filter traffic by path, method, and status. Run replays from the UI. Browse drift reports sorted by severity."
              visual={<DashboardVisual />}
            />
          </FadeIn>
        </div>
      </div>
    </section>
  );
}
