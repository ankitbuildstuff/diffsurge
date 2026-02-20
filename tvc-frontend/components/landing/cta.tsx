"use client";

import { Button } from "@/components/ui/button";
import { FadeIn } from "@/components/ui/fade-in";
import { ArrowRight } from "lucide-react";

export function CTA() {
  return (
    <section className="relative overflow-hidden bg-zinc-950 py-24 md:py-32">
      {/* Animated gradient orbs */}
      <div className="absolute top-0 left-1/4 h-72 w-72 rounded-full bg-teal-500/10 blur-[100px] animate-pulse-soft pointer-events-none" />
      <div className="absolute bottom-0 right-1/4 h-64 w-64 rounded-full bg-indigo-500/10 blur-[100px] animate-pulse-soft pointer-events-none" />

      {/* Grid pattern */}
      <div
        className="pointer-events-none absolute inset-0 opacity-[0.04]"
        style={{
          backgroundImage:
            "linear-gradient(rgba(255,255,255,1) 1px, transparent 1px), linear-gradient(90deg, rgba(255,255,255,1) 1px, transparent 1px)",
          backgroundSize: "48px 48px",
        }}
      />

      <div className="relative mx-auto max-w-[1200px] px-6 text-center">
        <FadeIn>
          <h2 className="text-3xl font-bold tracking-tight text-white sm:text-4xl">
            Ready to stop shipping breaking changes?
          </h2>
          <p className="mx-auto mt-4 max-w-md text-[14px] leading-relaxed text-zinc-400">
            Join hundreds of engineering teams that test against real traffic
            before every deploy. Free forever for schema diffing.
          </p>
          <div className="mt-8 flex flex-wrap items-center justify-center gap-3">
            <Button
              size="lg"
              className="btn-gradient border-0 text-white px-7"
            >
              Start for free
              <ArrowRight size={15} />
            </Button>
            <Button
              variant="ghost"
              size="lg"
              className="text-zinc-400 hover:text-white hover:bg-white/5"
            >
              View Documentation
            </Button>
          </div>
        </FadeIn>
      </div>
    </section>
  );
}
