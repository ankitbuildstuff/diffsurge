"use client";

import { Shield, Radio, Play, BarChart3, Lock } from "lucide-react";
import { SpotlightCard } from "@/components/ui/spotlight-card";
import { FadeIn } from "@/components/ui/fade-in";

/* ── Inline visuals ── */

function TerminalVisual() {
  return (
    <div className="mt-5 overflow-hidden rounded-lg border border-zinc-200 text-[11px] font-mono leading-relaxed">
      <div className="border-b border-zinc-100 bg-zinc-50 px-3 py-1.5 text-zinc-400">
        schema-diff.sh
      </div>
      <div className="bg-[#0a0a0f] px-3.5 py-3 text-zinc-500 space-y-0.5">
        <p><span className="text-teal-400">$</span> surge schema diff --fail-on-breaking</p>
        <p className="text-zinc-600 mt-1">Comparing 47 endpoints…</p>
        <p className="text-red-400 mt-1">✗ BREAKING  POST /api/users</p>
        <p className="text-zinc-600 ml-4">└─ Required field removed: &quot;email_verified&quot;</p>
        <p className="text-amber-400">⚠ WARNING   GET /api/users/:id</p>
        <p className="text-zinc-600 ml-4">└─ Type changed: &quot;age&quot; string → number</p>
        <p className="text-emerald-400">✓ SAFE      45 endpoints unchanged</p>
        <p className="text-zinc-600 mt-1 border-t border-zinc-800 pt-1.5">
          Exit code: 1 · <span className="text-red-400">1 breaking</span> · <span className="text-amber-400">1 warning</span>
        </p>
      </div>
    </div>
  );
}

function ProxyVisual() {
  return (
    <div className="mt-5 space-y-2 text-[11px]">
      {[
        { label: "Avg latency added", value: "3.2 ms", color: "text-emerald-600 bg-emerald-50" },
        { label: "Traffic sampled", value: "10%", color: "text-teal-600 bg-teal-50" },
        { label: "Requests buffered", value: "4,219", color: "text-zinc-600 bg-zinc-100" },
      ].map((m) => (
        <div key={m.label} className="flex items-center justify-between rounded-lg border border-zinc-100 bg-zinc-50/50 px-3 py-2">
          <span className="text-zinc-500">{m.label}</span>
          <span className={`rounded-full px-2 py-0.5 font-semibold ${m.color}`}>{m.value}</span>
        </div>
      ))}
    </div>
  );
}

function ReplayVisual() {
  const rows = [
    { path: "POST /checkout", status: "matched", ms: 42 },
    { path: "GET  /cart", status: "drift", ms: 18 },
    { path: "PUT  /profile", status: "matched", ms: 31 },
    { path: "GET  /products", status: "matched", ms: 12 },
  ];
  return (
    <div className="mt-5 space-y-1.5">
      {rows.map((r) => (
        <div key={r.path} className="flex items-center justify-between rounded-lg border border-zinc-100 bg-zinc-50/50 px-3 py-2 text-[11px]">
          <span className="font-mono text-zinc-600">{r.path}</span>
          <div className="flex items-center gap-2">
            <span className="text-zinc-400">{r.ms}ms</span>
            <span className={r.status === "matched"
              ? "rounded-full bg-emerald-50 px-2 py-0.5 font-medium text-emerald-600"
              : "rounded-full bg-amber-50 px-2 py-0.5 font-medium text-amber-600"
            }>{r.status}</span>
          </div>
        </div>
      ))}
    </div>
  );
}

function DashboardVisual() {
  const bars = [64, 42, 78, 52, 85, 38, 72, 58, 80, 46, 70, 55, 90, 48, 66, 74, 50, 82, 44, 76];
  return (
    <div className="mt-5 rounded-lg border border-zinc-100 bg-zinc-50/50 p-4">
      <div className="flex items-center justify-between text-[10px] text-zinc-400 mb-2">
        <span>Requests / min</span>
        <span className="font-medium text-zinc-600">1,247 total</span>
      </div>
      <div className="flex items-end gap-[3px] h-20">
        {bars.map((h, i) => (
          <div
            key={i}
            className="flex-1 rounded-sm bg-linear-to-t from-teal-500 to-teal-300"
            style={{ height: `${h}%`, opacity: 0.6 + (h / 300) }}
          />
        ))}
      </div>
      <div className="mt-2 flex justify-between text-[10px] text-zinc-400">
        <span>12:00</span>
        <span>15:00</span>
        <span>18:00</span>
      </div>
    </div>
  );
}

function PiiVisual() {
  return (
    <div className="mt-5 space-y-1.5 text-[11px] font-mono">
      <div className="rounded-lg border border-zinc-100 bg-zinc-50/50 px-3 py-2 flex items-center justify-between">
        <span className="text-zinc-500">&quot;email&quot;</span>
        <div className="flex items-center gap-1.5">
          <span className="text-zinc-400 line-through">john@acme.co</span>
          <span className="text-red-500">→</span>
          <span className="text-teal-600 font-semibold">***@***.co</span>
        </div>
      </div>
      <div className="rounded-lg border border-zinc-100 bg-zinc-50/50 px-3 py-2 flex items-center justify-between">
        <span className="text-zinc-500">&quot;phone&quot;</span>
        <div className="flex items-center gap-1.5">
          <span className="text-zinc-400 line-through">+1-555-0199</span>
          <span className="text-red-500">→</span>
          <span className="text-teal-600 font-semibold">+1-***-****</span>
        </div>
      </div>
      <div className="rounded-lg border border-zinc-100 bg-zinc-50/50 px-3 py-2 flex items-center justify-between">
        <span className="text-zinc-500">&quot;card&quot;</span>
        <div className="flex items-center gap-1.5">
          <span className="text-zinc-400 line-through">4242…4242</span>
          <span className="text-red-500">→</span>
          <span className="text-teal-600 font-semibold">****…****</span>
        </div>
      </div>
    </div>
  );
}

/* ── Feature card wrapper ── */

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
    <SpotlightCard className="h-full">
      <div className="flex h-full flex-col p-6">
        <div className="mb-3 inline-flex h-8 w-8 items-center justify-center rounded-lg bg-zinc-100 text-zinc-600">
          {icon}
        </div>
        <p className="text-[11px] font-medium uppercase tracking-wider text-teal-600">
          {label}
        </p>
        <h3 className="mt-1 text-[15px] font-semibold leading-snug text-zinc-900">
          {title}
        </h3>
        <p className="mt-2 text-[13px] leading-relaxed text-zinc-500">
          {description}
        </p>
        {visual}
      </div>
    </SpotlightCard>
  );
}

/* ── Bento grid ──
 *
 *  Desktop (lg) layout:
 *  ┌───────────────────┬──────────┐
 *  │  Schema Guardian  │ Traffic  │
 *  │  (2 cols)         │ Proxy    │
 *  ├──────────┬────────┴──────────┤
 *  │   PII    │  Replay Engine    │
 *  │ Redact   │  (2 cols)        │
 *  ├──────────┴───────────────────┤
 *  │  Live Dashboard (full width) │
 *  └─────────────────────────────-┘
 */

export function Features() {
  return (
    <section id="features" className="bg-zinc-50/60 py-20 md:py-28">
      <div className="mx-auto max-w-[1200px] px-6">
        <FadeIn>
          <p className="text-[12px] font-medium uppercase tracking-widest text-teal-600">
            Features
          </p>
          <h2 className="mt-3 text-[1.75rem] font-bold tracking-tight sm:text-3xl">
            Everything you need to ship without fear
          </h2>
          <p className="mt-3 max-w-lg text-[14px] leading-relaxed text-zinc-500">
            From static schema checks in CI to live traffic replay in production
            — Driftsurge gives your team a complete safety net for every API
            change.
          </p>
        </FadeIn>

        {/* Row 1: Schema Guardian (wide) + Traffic Proxy */}
        <div className="mt-12 grid gap-4 lg:grid-cols-3">
          <FadeIn delay={0} className="lg:col-span-2">
            <Card
              icon={<Shield size={18} />}
              label="Schema Guardian"
              title="Detect breaking changes at the schema level"
              description="Parses OpenAPI 3.x, GraphQL SDL, and gRPC proto files. Compares endpoints, request params, response shapes, and required fields — flagging every breaking change with severity and a JSON path."
              visual={<TerminalVisual />}
            />
          </FadeIn>
          <FadeIn delay={0.06} className="lg:col-span-1">
            <Card
              icon={<Radio size={18} />}
              label="Traffic Proxy"
              title="Capture production traffic without code changes"
              description="A lightweight reverse proxy that samples real request/response pairs. Async buffering keeps latency under 5 ms."
              visual={<ProxyVisual />}
            />
          </FadeIn>
        </div>

        {/* Row 2: PII Redaction + Replay Engine (wide) */}
        <div className="mt-4 grid gap-4 lg:grid-cols-3">
          <FadeIn delay={0.12} className="lg:col-span-1">
            <Card
              icon={<Lock size={18} />}
              label="PII Redaction"
              title="Automatic sensitive data masking"
              description="Emails, phones, credit cards, and SSNs are redacted before storage. Add custom patterns on Enterprise."
              visual={<PiiVisual />}
            />
          </FadeIn>
          <FadeIn delay={0.18} className="lg:col-span-2">
            <Card
              icon={<Play size={18} />}
              label="Replay Engine"
              title="Replay 1,000+ requests per second against staging"
              description="Fires captured traffic at your new build with configurable concurrency. Semantic comparison ignores field order and whitespace — only real differences surface."
              visual={<ReplayVisual />}
            />
          </FadeIn>
        </div>

        {/* Row 3: Dashboard (full width) */}
        <div className="mt-4">
          <FadeIn delay={0.24}>
            <Card
              icon={<BarChart3 size={18} />}
              label="Live Dashboard"
              title="Real-time visibility into every request"
              description="Filter traffic by path, method, and status. Run replays from the UI. Browse drift reports sorted by severity. Export audit logs as PDF or CSV."
              visual={<DashboardVisual />}
            />
          </FadeIn>
        </div>
      </div>
    </section>
  );
}
