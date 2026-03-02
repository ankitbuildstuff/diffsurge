"use client";

import { FadeIn } from "@/components/ui/fade-in";
import { ArrowRight } from "lucide-react";
import Link from "next/link";

export function CTA() {
  return (
    <section
      style={{
        background: "var(--bg-dark)",
        position: "relative",
        overflow: "hidden",
        paddingTop: 96,
        paddingBottom: 96,
      }}
    >
      {/* Subtle research grid on dark */}
      <div
        className="absolute inset-0 pointer-events-none"
        style={{
          backgroundImage:
            "linear-gradient(to right, rgba(255,255,255,0.02) 1px, transparent 1px), linear-gradient(to bottom, rgba(255,255,255,0.02) 1px, transparent 1px)",
          backgroundSize: "64px 64px",
        }}
      />

      {/* Data stripe accent */}
      <div
        className="absolute top-0 left-0 right-0 data-stripe-wide animate-stripe"
        style={{ height: 3 }}
      />

      <div className="relative mx-auto max-w-[1120px] px-6 text-center">
        <FadeIn>
          <h2
            className="font-editorial"
            style={{
              fontSize: "clamp(2rem, 5vw, 3rem)",
              lineHeight: 1.1,
              color: "var(--text-on-dark)",
            }}
          >
            Ready to stop shipping{" "}
            <span className="font-editorial-italic">breaking changes?</span>
          </h2>
          <p
            style={{
              marginTop: 16,
              maxWidth: 420,
              marginLeft: "auto",
              marginRight: "auto",
              fontSize: 14,
              lineHeight: 1.7,
              color: "var(--text-on-dark-muted)",
            }}
          >
            Start catching API drift today. Free forever for schema diffing.
          </p>
          <div
            style={{
              marginTop: 32,
              display: "flex",
              flexWrap: "wrap",
              justifyContent: "center",
              gap: 12,
            }}
          >
            <Link
              href="/signup"
              className="btn-research"
              style={{
                background: "var(--bg-primary)",
                color: "var(--text-primary)",
              }}
            >
              Start for free
              <ArrowRight size={14} />
            </Link>
            <Link
              href="/docs"
              className="btn-research-outline"
              style={{
                borderColor: "var(--border-dark)",
                color: "var(--text-on-dark-muted)",
              }}
            >
              Read the docs
            </Link>
          </div>
        </FadeIn>
      </div>
    </section>
  );
}
