"use client";

import { useState, useEffect, useRef, Fragment } from "react";
import { FadeIn } from "@/components/ui/fade-in";
import { Github, ArrowRight, Globe, Radio, Play, CheckCircle2 } from "lucide-react";
import { siteConfig } from "@/lib/constants";
import Link from "next/link";

/* ── Pipeline ── */
function PipelineAnimation() {
  const nodes = [
    { icon: Globe, label: "Production", sublabel: "Live Traffic" },
    { icon: Radio, label: "Capture", sublabel: "Record & Sample" },
    { icon: Play, label: "Replay", sublabel: "Test Staging" },
    { icon: CheckCircle2, label: "Report", sublabel: "Detect Changes" },
  ];

  return (
    <div className="w-full" style={{ paddingTop: 48 }}>
      {/* Desktop: horizontal */}
      <div className="hidden md:flex items-center justify-center gap-0">
        {nodes.map((node, i) => (
          <Fragment key={i}>
            <div
              className="card flex flex-col items-center gap-3 px-6 py-5"
              style={{ cursor: "default", width: 180 }}
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
                }}
              >
                <node.icon size={18} strokeWidth={1.5} style={{ color: "var(--text-muted)" }} />
              </div>
              <div style={{ textAlign: "center" }}>
                <p style={{ fontSize: 14, fontWeight: 500, color: "var(--text-primary)" }}>
                  {node.label}
                </p>
                <p style={{ fontSize: 12, color: "var(--text-muted)", marginTop: 2 }}>
                  {node.sublabel}
                </p>
              </div>
            </div>
            {i < nodes.length - 1 && (
              <div
                style={{
                  width: 48,
                  height: 0,
                  borderTop: "1px dashed var(--border-light)",
                  flexShrink: 0,
                }}
              />
            )}
          </Fragment>
        ))}
      </div>

      {/* Mobile: vertical */}
      <div className="flex md:hidden flex-col items-center gap-0">
        {nodes.map((node, i) => (
          <Fragment key={i}>
            <div
              className="card flex items-center gap-3 px-5 py-4 w-full"
              style={{ cursor: "default" }}
            >
              <div
                style={{
                  width: 36,
                  height: 36,
                  borderRadius: 8,
                  border: "1px solid var(--border-subtle)",
                  display: "flex",
                  alignItems: "center",
                  justifyContent: "center",
                  flexShrink: 0,
                }}
              >
                <node.icon size={16} strokeWidth={1.5} style={{ color: "var(--text-muted)" }} />
              </div>
              <div>
                <p style={{ fontSize: 14, fontWeight: 500, color: "var(--text-primary)" }}>
                  {node.label}
                </p>
                <p style={{ fontSize: 12, color: "var(--text-muted)" }}>
                  {node.sublabel}
                </p>
              </div>
            </div>
            {i < nodes.length - 1 && (
              <div
                style={{
                  width: 0,
                  height: 24,
                  borderLeft: "1px dashed var(--border-light)",
                }}
              />
            )}
          </Fragment>
        ))}
      </div>
    </div>
  );
}

/* ── Hero Terminal ── */
function HeroTerminal() {
  const [lines, setLines] = useState<{ text: string; color?: string }[]>([]);
  const [currentTyping, setCurrentTyping] = useState("");
  const [showCursor, setShowCursor] = useState(true);
  const terminalRef = useRef<HTMLDivElement>(null);

  const commands = [
    { text: "$ npm install -g diffsurge", delay: 40, pause: 800 },
    { text: "\u2713 installed diffsurge@1.2.0", delay: 0, pause: 400, output: true, color: "var(--accent-green)" },
    { text: "", delay: 0, pause: 300, output: true },
    { text: "$ surge capture --proxy :8080", delay: 40, pause: 800 },
    { text: "\u25B8 Proxy listening on :8080", delay: 0, pause: 300, output: true, color: "var(--accent-cyan)" },
    { text: "\u25B8 Capturing traffic...", delay: 0, pause: 600, output: true, color: "#888" },
    { text: "", delay: 0, pause: 300, output: true },
    { text: "$ surge replay --target http://staging:3000", delay: 35, pause: 800 },
    { text: "Replaying 1,247 requests...", delay: 0, pause: 500, output: true, color: "#888" },
    { text: "\u2713 1,241 matched (99.5%)", delay: 0, pause: 200, output: true, color: "var(--accent-green)" },
    { text: "\u2717 2 breaking changes detected", delay: 0, pause: 200, output: true, color: "var(--accent-red)" },
    { text: "", delay: 0, pause: 300, output: true },
    { text: "$ surge diff", delay: 40, pause: 600 },
    { text: "\u2717 BREAKING POST /api/users", delay: 0, pause: 200, output: true, color: "#f97316" },
    { text: "  \u2514\u2500 Field removed: \"email_verified\"", delay: 0, pause: 200, output: true, color: "#888" },
    { text: "\u26A0 WARNING  GET /api/users/:id", delay: 0, pause: 200, output: true, color: "var(--accent-amber)" },
    { text: "  \u2514\u2500 Type: \"age\" string \u2192 number", delay: 0, pause: 200, output: true, color: "#888" },
    { text: "\u2713 SAFE     45 endpoints unchanged", delay: 0, pause: 3000, output: true, color: "var(--accent-green)" },
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
      while (!cancelled) {
        setLines([]);
        setCurrentTyping("");
        for (const cmd of commands) {
          if (cancelled) return;
          if (cmd.output) {
            setCurrentTyping("");
            setLines((prev) => [...prev, { text: cmd.text, color: cmd.color }]);
            await sleep(cmd.pause);
          } else {
            await typeChar(cmd.text, cmd.delay);
            await sleep(cmd.pause);
            setLines((prev) => [...prev, { text: cmd.text }]);
            setCurrentTyping("");
          }
        }
        await sleep(3000);
      }
    }

    run();
    const cursorInterval = setInterval(() => setShowCursor((v) => !v), 530);
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
    <div className="terminal">
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
      <div
        ref={terminalRef}
        className="scrollbar-hide"
        style={{
          padding: "16px 20px",
          fontFamily: "var(--font-mono)",
          fontSize: 13,
          lineHeight: 1.8,
          height: 320,
          overflowY: "auto",
        }}
      >
        {lines.map((line, i) => (
          <p key={i} style={{ color: line.color || "#ccc", margin: 0 }}>
            {line.text || "\u00A0"}
          </p>
        ))}
        {currentTyping !== undefined && (
          <p style={{ color: "#ccc", margin: 0 }}>
            {currentTyping}
            <span
              style={{
                opacity: showCursor ? 1 : 0,
                color: "var(--accent-cyan)",
                transition: "opacity 0.1s",
              }}
            >
              ▊
            </span>
          </p>
        )}
      </div>
    </div>
  );
}

export function Hero() {
  return (
    <section style={{ paddingTop: 64 }}>
      <div
        className="mx-auto px-6 text-center"
        style={{ maxWidth: 1200, paddingTop: 96, paddingBottom: 64 }}
      >
        {/* Badge */}
        <FadeIn>
          <div
            style={{
              display: "inline-flex",
              alignItems: "center",
              gap: 8,
              padding: "6px 14px",
              borderRadius: 999,
              border: "1px solid var(--border-subtle)",
              marginBottom: 32,
            }}
          >
            <div
              style={{
                width: 6,
                height: 6,
                borderRadius: "50%",
                background: "var(--accent-green)",
              }}
            />
            <span style={{ fontSize: 12, fontWeight: 500, color: "var(--text-muted)" }}>
              Open Source API Regression Testing
            </span>
          </div>
        </FadeIn>

        {/* Headline */}
        <FadeIn delay={0.1}>
          <h1
            style={{
              fontSize: "clamp(2.5rem, 5.5vw, 4.5rem)",
              fontWeight: 500,
              lineHeight: 1.08,
              letterSpacing: "-0.02em",
              maxWidth: 800,
              margin: "0 auto",
              color: "var(--text-primary)",
            }}
          >
            Test your new API version with real production traffic
          </h1>
        </FadeIn>

        {/* Subheadline */}
        <FadeIn delay={0.2}>
          <p
            style={{
              marginTop: 24,
              maxWidth: 560,
              marginLeft: "auto",
              marginRight: "auto",
              fontSize: 17,
              lineHeight: 1.7,
              color: "var(--text-secondary)",
            }}
          >
            DiffSurge captures real production requests, replays them against
            your staging API, and detects breaking changes before deployment.
          </p>
        </FadeIn>

        {/* CTAs */}
        <FadeIn delay={0.3}>
          <div
            style={{
              marginTop: 36,
              display: "flex",
              flexWrap: "wrap",
              justifyContent: "center",
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
              Star on GitHub
            </a>
            <Link href="/docs" className="btn-secondary">
              View Docs
              <ArrowRight size={14} />
            </Link>
          </div>
        </FadeIn>

        {/* Terminal */}
        <FadeIn delay={0.4}>
          <div style={{ marginTop: 64, maxWidth: 720, marginLeft: "auto", marginRight: "auto", textAlign: "left" }}>
            <HeroTerminal />
          </div>
        </FadeIn>

        {/* Pipeline */}
        <FadeIn delay={0.5}>
          <PipelineAnimation />
        </FadeIn>
      </div>
    </section>
  );
}
