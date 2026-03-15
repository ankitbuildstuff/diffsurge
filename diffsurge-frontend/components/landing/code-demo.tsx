"use client";

import { useState, useEffect, useRef } from "react";
import { FadeIn } from "@/components/ui/fade-in";

const demoLines: {
  text: string;
  color?: string;
  delay: number;
  pause: number;
  typed?: boolean;
}[] = [
  {
    text: "# install",
    color: "#666",
    delay: 0,
    pause: 400,
  },
  {
    text: "$ npm install -g diffsurge",
    delay: 40,
    pause: 800,
    typed: true,
  },
  {
    text: "✓ diffsurge@1.2.0 installed",
    color: "var(--accent-green)",
    delay: 0,
    pause: 600,
  },
  { text: "", delay: 0, pause: 300 },
  {
    text: "# capture traffic",
    color: "#666",
    delay: 0,
    pause: 400,
  },
  {
    text: "$ surge capture --proxy :8080",
    delay: 35,
    pause: 800,
    typed: true,
  },
  {
    text: "▸ Proxy listening on :8080",
    color: "var(--accent-cyan)",
    delay: 0,
    pause: 300,
  },
  {
    text: "▸ Sampling 10% of traffic",
    color: "#888",
    delay: 0,
    pause: 300,
  },
  {
    text: "▸ PII redaction: enabled",
    color: "#888",
    delay: 0,
    pause: 600,
  },
  { text: "", delay: 0, pause: 300 },
  {
    text: "# replay traffic against staging",
    color: "#666",
    delay: 0,
    pause: 400,
  },
  {
    text: "$ surge replay --target http://staging:3000",
    delay: 30,
    pause: 1000,
    typed: true,
  },
  {
    text: "Replaying 1,247 requests...",
    color: "#888",
    delay: 0,
    pause: 500,
  },
  {
    text: "✓ 1,241 responses matched (99.5%)",
    color: "var(--accent-green)",
    delay: 0,
    pause: 200,
  },
  {
    text: "⚠ 4 warnings (type coercion)",
    color: "var(--accent-amber)",
    delay: 0,
    pause: 200,
  },
  {
    text: "✗ 2 breaking (missing fields)",
    color: "var(--accent-red)",
    delay: 0,
    pause: 600,
  },
  { text: "", delay: 0, pause: 300 },
  {
    text: "# detect breaking responses",
    color: "#666",
    delay: 0,
    pause: 400,
  },
  { text: "$ surge diff", delay: 40, pause: 600, typed: true },
  {
    text: "✗ BREAKING  POST /api/users",
    color: "#f97316",
    delay: 0,
    pause: 200,
  },
  {
    text: '  └─ Required field removed: "email_verified"',
    color: "#888",
    delay: 0,
    pause: 200,
  },
  {
    text: "⚠ WARNING   GET /api/users/:id",
    color: "var(--accent-amber)",
    delay: 0,
    pause: 200,
  },
  {
    text: '  └─ Type changed: "age" string → number',
    color: "#888",
    delay: 0,
    pause: 200,
  },
  {
    text: "✓ SAFE      45 endpoints unchanged",
    color: "var(--accent-green)",
    delay: 0,
    pause: 200,
  },
  { text: "", delay: 0, pause: 200 },
  {
    text: "1 breaking · 1 warning — exit code 1",
    color: "#666",
    delay: 0,
    pause: 3000,
  },
];

export function CodeDemo() {
  const [lines, setLines] = useState<{ text: string; color?: string }[]>([]);
  const [currentTyping, setCurrentTyping] = useState("");
  const [showCursor, setShowCursor] = useState(true);
  const terminalRef = useRef<HTMLDivElement>(null);

  useEffect(() => {
    let cancelled = false;
    const sleep = (ms: number) => new Promise((r) => setTimeout(r, ms));

    async function typeText(text: string, delay: number) {
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
        for (const line of demoLines) {
          if (cancelled) return;
          if (line.typed) {
            await typeText(line.text, line.delay);
            await sleep(line.pause);
            setLines((prev) => [
              ...prev,
              { text: line.text, color: "#ccc" },
            ]);
            setCurrentTyping("");
          } else {
            setCurrentTyping("");
            setLines((prev) => [
              ...prev,
              { text: line.text, color: line.color },
            ]);
            await sleep(line.pause);
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
  }, []);

  useEffect(() => {
    if (terminalRef.current)
      terminalRef.current.scrollTop = terminalRef.current.scrollHeight;
  }, [lines, currentTyping]);

  return (
    <section className="section-spacing">
      <div className="mx-auto px-6" style={{ maxWidth: 1200 }}>
        <FadeIn>
          <div
            style={{ textAlign: "center", maxWidth: 560, margin: "0 auto" }}
          >
            <p className="micro-label" style={{ marginBottom: 16 }}>
              Developer Workflow
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
              CLI-first. Developer-friendly.
            </h2>
            <p
              style={{
                marginTop: 16,
                fontSize: 15,
                lineHeight: 1.7,
                color: "var(--text-muted)",
              }}
            >
              Four commands. Zero configuration. From install to breaking change
              detection in under a minute.
            </p>
          </div>
        </FadeIn>

        <FadeIn delay={0.15}>
          <div
            className="terminal"
            style={{ maxWidth: 720, margin: "48px auto 0" }}
          >
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
                ~ diffsurge
              </span>
            </div>
            <div
              ref={terminalRef}
              className="scrollbar-hide"
              style={{
                padding: "20px 24px",
                fontFamily: "var(--font-mono)",
                fontSize: 13,
                lineHeight: 1.9,
                minHeight: 420,
                maxHeight: 500,
                overflowY: "auto",
              }}
            >
              {lines.map((line, i) => (
                <p
                  key={i}
                  style={{
                    color: line.color || "#ccc",
                    margin: 0,
                  }}
                >
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
        </FadeIn>
      </div>
    </section>
  );
}
