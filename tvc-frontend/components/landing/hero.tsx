"use client";

import { Button } from "@/components/ui/button";
import { FadeIn } from "@/components/ui/fade-in";
import { ArrowRight, Sparkles } from "lucide-react";

function TerminalMockup() {
  return (
    <div className="relative animate-float">
      {/* Glow orbs behind */}
      <div className="absolute -top-12 -right-12 h-48 w-48 rounded-full bg-teal-400/10 blur-3xl animate-pulse-soft" />
      <div className="absolute -bottom-8 -left-8 h-36 w-36 rounded-full bg-indigo-400/10 blur-3xl animate-pulse-soft" />

      <div className="relative overflow-hidden rounded-xl border border-zinc-200 bg-white shadow-[0_8px_40px_rgba(0,0,0,0.08),0_0_0_1px_rgba(0,0,0,0.02)]">
        {/* macOS title bar */}
        <div className="flex items-center gap-2 border-b border-zinc-100 bg-zinc-50 px-4 py-2.5">
          <div className="flex gap-1.5">
            <div className="h-[10px] w-[10px] rounded-full bg-[#ff5f57]" />
            <div className="h-[10px] w-[10px] rounded-full bg-[#febc2e]" />
            <div className="h-[10px] w-[10px] rounded-full bg-[#28c840]" />
          </div>
          <span className="ml-2 font-mono text-[11px] text-zinc-400">
            ~/project
          </span>
        </div>

        {/* Terminal body */}
        <div className="bg-[#0a0a0f] p-5 font-mono text-[12px] leading-[1.8]">
          <p className="text-zinc-500">
            <span className="text-teal-400">$</span> driftsurge schema diff --old
            api-v1.yaml --new api-v2.yaml
          </p>
          <p className="mt-2.5 text-zinc-500">Comparing 47 endpoints…</p>
          <div className="mt-2.5 space-y-2">
            <div>
              <span className="font-semibold text-red-400">✗ BREAKING</span>
              <span className="text-zinc-300">{"  "}POST /api/users</span>
              <p className="ml-4 text-zinc-600">
                └─ Required field removed:{" "}
                <span className="text-red-300">&quot;email_verified&quot;</span>
              </p>
            </div>
            <div>
              <span className="font-semibold text-amber-400">⚠ WARNING</span>
              <span className="text-zinc-300">{"  "}GET /api/users/:id</span>
              <p className="ml-4 text-zinc-600">
                └─ Type changed:{" "}
                <span className="text-amber-200">&quot;age&quot;</span> string →
                number
              </p>
            </div>
            <div>
              <span className="font-semibold text-emerald-400">✓ SAFE</span>
              <span className="text-zinc-300">{"    "}GET /api/products</span>
              <p className="ml-4 text-zinc-600">
                └─ Optional field added:{" "}
                <span className="text-emerald-300">&quot;metadata&quot;</span>
              </p>
            </div>
          </div>
          <div className="mt-2.5 border-t border-zinc-800 pt-2.5 text-zinc-500">
            <span className="text-red-400">1 breaking</span> ·{" "}
            <span className="text-amber-400">1 warning</span> ·{" "}
            <span className="text-emerald-400">1 safe</span> — exit code 1
          </div>
        </div>
      </div>
    </div>
  );
}

export function Hero() {
  return (
    <section className="relative overflow-hidden bg-white pt-14">
      {/* Background mesh + grid */}
      <div className="hero-mesh absolute inset-0 pointer-events-none" />
      <div className="bg-grid-pattern absolute inset-0 pointer-events-none" />

      <div className="relative mx-auto grid max-w-[1200px] items-center gap-10 px-6 pt-24 pb-20 md:grid-cols-2 md:gap-16 md:pt-32 md:pb-28">
        {/* Left — copy */}
        <div>
          <FadeIn delay={0}>
            <div className="mb-5 inline-flex items-center gap-2 rounded-full border border-zinc-200 bg-white px-3 py-1 shadow-[0_1px_4px_rgba(0,0,0,0.04)]">
              <Sparkles size={12} className="text-teal-500" />
              <span className="text-[12px] font-medium text-zinc-500">
                Introducing Driftsurge
              </span>
            </div>
          </FadeIn>

          <FadeIn delay={0.1}>
            <h1 className="text-4xl font-bold leading-[1.08] tracking-tight sm:text-5xl lg:text-[3.4rem]">
              Catch{" "}
              <span className="text-gradient-animated">
                breaking API changes
              </span>{" "}
              before your users do
            </h1>
          </FadeIn>

          <FadeIn delay={0.2}>
            <p className="mt-5 max-w-lg text-[15px] leading-[1.7] text-zinc-500">
              Driftsurge captures production traffic, replays it against your
              staging builds, and surfaces every breaking change, type mismatch,
              and missing field — so you ship with confidence instead of
              crossing your fingers.
            </p>
          </FadeIn>

          <FadeIn delay={0.3}>
            <div className="mt-8 flex flex-wrap items-center gap-3">
              <Button
                size="lg"
                className="btn-gradient border-0 text-white px-7"
              >
                Start for free
                <ArrowRight size={15} />
              </Button>
              <Button variant="secondary" size="lg">
                Watch demo
              </Button>
            </div>
            <p className="mt-4 text-[12px] text-zinc-400">
              Free forever for schema diffing · No credit card required
            </p>
          </FadeIn>
        </div>

        {/* Right — terminal */}
        <FadeIn delay={0.2} direction="right">
          <TerminalMockup />
        </FadeIn>
      </div>

      {/* Fade-out bottom edge */}
      <div className="absolute bottom-0 left-0 right-0 h-24 bg-gradient-to-t from-white to-transparent pointer-events-none" />
    </section>
  );
}
