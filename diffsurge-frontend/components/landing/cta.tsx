"use client";

import { FadeIn } from "@/components/ui/fade-in";
import { ArrowRight } from "lucide-react";
import Link from "next/link";

export function CTA() {
  return (
    <section
      style={{
        background: "var(--bg-dark)",
        paddingTop: 128,
        paddingBottom: 128,
      }}
    >
      <div
        className="mx-auto px-6 text-center"
        style={{ maxWidth: 1200 }}
      >
        <FadeIn>
          <h2
            style={{
              fontSize: "clamp(2rem, 5vw, 3rem)",
              fontWeight: 500,
              lineHeight: 1.1,
              letterSpacing: "-0.02em",
              color: "var(--text-on-dark)",
            }}
          >
            Ready to stop shipping breaking changes?
          </h2>
          <p
            style={{
              marginTop: 16,
              maxWidth: 420,
              marginLeft: "auto",
              marginRight: "auto",
              fontSize: 15,
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
              className="btn-primary"
              style={{
                background: "#ffffff",
                color: "#111111",
              }}
            >
              Start for free
              <ArrowRight size={14} />
            </Link>
            <Link
              href="/docs"
              className="btn-secondary"
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
