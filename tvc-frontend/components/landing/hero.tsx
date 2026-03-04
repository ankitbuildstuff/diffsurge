"use client";

import { useState, useEffect, useRef } from "react";
import { FadeIn } from "@/components/ui/fade-in";
import { ArrowRight } from "lucide-react";
import { createClient } from "@/lib/supabase/client";
import Link from "next/link";

/* ── Waveform Visual (data-visual motif) ── */
function WaveformVisual() {
  const bars = [
    35, 55, 25, 70, 40, 85, 30, 60, 45, 75, 20, 65, 50, 80, 35, 55, 90, 28,
    72, 42, 68, 38, 82, 22, 58, 48, 78, 32, 62, 52, 88, 26, 70, 44, 76, 34,
    64, 54, 86, 30,
  ];

  const colors = [
    "var(--accent-purple)",
    "var(--accent-blue)",
    "var(--accent-teal)",
    "var(--accent-yellow)",
    "var(--accent-orange)",
    "var(--accent-rose)",
  ];

  return (
    <div className="flex items-end gap-[2px]" style={{ height: 120 }}>
      {bars.map((h, i) => (
        <div
          key={i}
          style={{
            width: 3,
            height: `${h}%`,
            backgroundColor: colors[i % colors.length],
            borderRadius: 1,
            opacity: 0.7 + (h / 400),
          }}
        />
      ))}
    </div>
  );
}

/* ── Spectrum Stripe Block ── */
function SpectrumBlock() {
  return (
    <div
      style={{
        width: "100%",
        display: "flex",
        flexDirection: "column",
        gap: 16,
      }}
    >
      {/* Waveform */}
      <WaveformVisual />

      {/* Data stripe bar */}
      <div
        className="data-stripe-wide animate-stripe"
        style={{ height: 6, borderRadius: 3, width: "100%" }}
      />

      {/* Signal metrics */}
      <div
        style={{
          display: "grid",
          gridTemplateColumns: "1fr 1fr 1fr",
          gap: 12,
          marginTop: 4,
        }}
      >
        {[
          { label: "ENDPOINTS SCANNED", value: "2,847" },
          { label: "BREAKING CHANGES", value: "12" },
          { label: "CONFIDENCE", value: "99.5%" },
        ].map((m) => (
          <div key={m.label}>
            <p
              className="micro-label"
              style={{ fontSize: 9, marginBottom: 4 }}
            >
              {m.label}
            </p>
            <p
              className="font-editorial"
              style={{ fontSize: 22, lineHeight: 1, color: "var(--text-primary)" }}
            >
              {m.value}
            </p>
          </div>
        ))}
      </div>
    </div>
  );
}

/* ── Animated Terminal (restrained) ── */
function ResearchTerminal() {
  const [lines, setLines] = useState<
    { text: string; color?: string; dim?: boolean }[]
  >([]);
  const [currentTyping, setCurrentTyping] = useState("");
  const [showCursor, setShowCursor] = useState(true);
  const terminalRef = useRef<HTMLDivElement>(null);
  const animating = useRef(true);

  const demoLines = [
    { text: "$ surge schema diff --old v1.yaml --new v2.yaml", delay: 35, pause: 700 },
    { text: "", delay: 0, pause: 100, output: true },
    { text: "Comparing 47 endpoints…", delay: 0, pause: 500, output: true, dim: true },
    { text: "", delay: 0, pause: 200, output: true },
    { text: "✗ BREAKING  POST /api/users", delay: 0, pause: 200, output: true, color: "var(--accent-orange)" },
    { text: "  └─ Required field removed: \"email_verified\"", delay: 0, pause: 200, output: true, dim: true },
    { text: "⚠ WARNING   GET /api/users/:id", delay: 0, pause: 200, output: true, color: "var(--accent-yellow)" },
    { text: "  └─ Type changed: \"age\" string → number", delay: 0, pause: 200, output: true, dim: true },
    { text: "✓ SAFE      45 endpoints unchanged", delay: 0, pause: 200, output: true, color: "var(--accent-teal)" },
    { text: "", delay: 0, pause: 200, output: true },
    { text: "1 breaking · 1 warning — exit code 1", delay: 0, pause: 2000, output: true, dim: true },
  ];

  useEffect(() => {
    let cancelled = false;
    const sleep = (ms: number) => new Promise((r) => setTimeout(r, ms));

    async function typeChar(text: string, delay: number) {
      for (let i = 0; i <= text.length; i++) {
        if (cancelled) return;
        setCurrentTyping(text.slice(0, i));
        await sleep(delay);
      }
    }

    async function run() {
      while (animating.current && !cancelled) {
        setLines([]);
        setCurrentTyping("");
        for (const line of demoLines) {
          if (cancelled) return;
          if (line.output) {
            setCurrentTyping("");
            setLines((prev) => [
              ...prev,
              { text: line.text, color: line.color, dim: line.dim },
            ]);
            await sleep(line.pause);
          } else {
            await typeChar(line.text, line.delay);
            await sleep(line.pause);
            setLines((prev) => [...prev, { text: line.text }]);
            setCurrentTyping("");
          }
        }
        await sleep(3000);
      }
    }

    run();
    const cursorInterval = setInterval(
      () => setShowCursor((v) => !v),
      530
    );
    return () => {
      cancelled = true;
      clearInterval(cursorInterval);
    };
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, []);

  useEffect(() => {
    if (terminalRef.current) {
      terminalRef.current.scrollTop = terminalRef.current.scrollHeight;
    }
  }, [lines, currentTyping]);

  return (
    <div className="terminal-research">
      <div className="terminal-research-header">
        <div className="dot" />
        <div className="dot" />
        <div className="dot" />
        <span
          style={{
            marginLeft: 8,
            fontFamily: "var(--font-mono)",
            fontSize: 11,
            color: "rgba(255,255,255,0.25)",
          }}
        >
          terminal
        </span>
      </div>
      <div
        ref={terminalRef}
        className="scrollbar-hide"
        style={{
          padding: "16px 18px",
          fontFamily: "var(--font-mono)",
          fontSize: 12,
          lineHeight: 1.9,
          height: 260,
          overflowY: "auto",
        }}
      >
        {lines.map((line, i) => (
          <p
            key={i}
            style={{
              color: line.color || (line.dim ? "rgba(255,255,255,0.3)" : "rgba(255,255,255,0.55)"),
              margin: 0,
            }}
          >
            {line.text || "\u00A0"}
          </p>
        ))}
        {currentTyping !== undefined && (
          <p style={{ color: "rgba(255,255,255,0.55)", margin: 0 }}>
            {currentTyping}
            <span
              style={{
                opacity: showCursor ? 1 : 0,
                color: "var(--accent-teal)",
                transition: "opacity 0.1s",
              }}
            >
              █
            </span>
          </p>
        )}
      </div>
    </div>
  );
}

export function Hero() {
  const [isLoggedIn, setIsLoggedIn] = useState(false);

  useEffect(() => {
    const supabase = createClient();
    supabase.auth.getUser().then(({ data: { user } }) => {
      setIsLoggedIn(!!user);
    });
  }, []);

  return (
    <section
      className="relative overflow-hidden"
      style={{ paddingTop: 80, background: "var(--bg-primary)" }}
    >
      {/* Subtle research grid */}
      <div
        className="bg-research-grid absolute inset-0 pointer-events-none"
        style={{ opacity: 0.5 }}
      />

      <div className="relative mx-auto max-w-[1120px] px-6 pt-16 pb-24 md:pt-24 md:pb-32">
        {/* Micro label */}
        <FadeIn delay={0}>
          <div
            className="micro-label"
            style={{
              display: "inline-flex",
              alignItems: "center",
              gap: 8,
              marginBottom: 24,
            }}
          >
            <span
              className="data-stripe"
              style={{
                width: 12,
                height: 12,
                borderRadius: 3,
                display: "inline-block",
              }}
            />
            <span>API Drift Detection</span>
          </div>
        </FadeIn>

        {/* Main headline — serif, large, deliberate */}
        <FadeIn delay={0.1}>
          <h1
            className="font-editorial"
            style={{
              fontSize: "clamp(2.4rem, 6vw, 4.2rem)",
              lineHeight: 1.08,
              maxWidth: 720,
              color: "var(--text-primary)",
            }}
          >
            Catch breaking{" "}
            <span className="font-editorial-italic">API changes</span>{" "}
            before your users do
          </h1>
        </FadeIn>

        {/* Supporting paragraph — sans, light */}
        <FadeIn delay={0.2}>
          <p
            style={{
              marginTop: 24,
              maxWidth: 520,
              fontSize: 15,
              lineHeight: 1.7,
              color: "var(--text-secondary)",
              fontWeight: 400,
            }}
          >
            Diffsurge captures production traffic, replays it against your
            staging builds, and surfaces every breaking change — so you ship
            with confidence instead of crossing your fingers.
          </p>
          <p
            style={{
              marginTop: 12,
              maxWidth: 520,
              fontSize: 13,
              lineHeight: 1.7,
              color: "var(--text-muted)",
              fontWeight: 400,
            }}
          >
            An open-source API regression testing tool with an OpenAPI diff
            CLI, GraphQL schema comparison, gRPC proto diffing, and
            production traffic replay — all in one developer-friendly
            platform.
          </p>
        </FadeIn>

        {/* CTA — compact, research-grade */}
        <FadeIn delay={0.3}>
          <div
            style={{
              marginTop: 36,
              display: "flex",
              flexWrap: "wrap",
              alignItems: "center",
              gap: 12,
            }}
          >
            <Link
              href={isLoggedIn ? "/dashboard" : "/signup"}
              className="btn-research"
            >
              Start for free
              <ArrowRight size={14} />
            </Link>
            <Link href="/docs" className="btn-research-outline">
              Read the docs
            </Link>
          </div>
          <p
            style={{
              marginTop: 16,
              fontSize: 12,
              color: "var(--text-faint)",
            }}
          >
            Free forever for schema diffing · No credit card required
          </p>
        </FadeIn>

        {/* Data visual area */}
        <FadeIn delay={0.3}>
          <div
            className="grid grid-cols-1 md:grid-cols-2 gap-6"
            style={{ marginTop: 64 }}
          >
            {/* Terminal */}
            <div>
              <ResearchTerminal />
            </div>

            {/* Spectrum / Data visual */}
            <div
              className="card-flat"
              style={{ padding: 28 }}
            >
              <p className="micro-label" style={{ marginBottom: 20 }}>
                Signal Analysis
              </p>
              <SpectrumBlock />
            </div>
          </div>
        </FadeIn>
      </div>
    </section>
  );
}
