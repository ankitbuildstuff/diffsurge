"use client";

import { useState } from "react";
import { faqs } from "@/lib/constants";
import { FadeIn } from "@/components/ui/fade-in";
import { ChevronDown } from "lucide-react";

function AccordionItem({
  q,
  a,
  open,
  onClick,
}: {
  q: string;
  a: string;
  open: boolean;
  onClick: () => void;
}) {
  return (
    <div style={{ borderBottom: "1px solid var(--border-subtle)" }}>
      <button
        onClick={onClick}
        className="flex w-full items-center justify-between py-5 text-left cursor-pointer"
      >
        <span
          style={{
            fontSize: 14,
            fontWeight: 500,
            color: "var(--text-primary)",
            paddingRight: 16,
          }}
        >
          {q}
        </span>
        <ChevronDown
          size={15}
          strokeWidth={1.5}
          style={{
            flexShrink: 0,
            color: "var(--text-faint)",
            transition: "transform 0.2s ease",
            transform: open ? "rotate(180deg)" : "rotate(0deg)",
          }}
        />
      </button>
      <div
        style={{
          overflow: "hidden",
          transition: "all 0.3s ease-out",
          maxHeight: open ? 500 : 0,
          paddingBottom: open ? 20 : 0,
        }}
      >
        <p
          style={{
            fontSize: 13,
            lineHeight: 1.7,
            color: "var(--text-muted)",
          }}
        >
          {a}
        </p>
      </div>
    </div>
  );
}

export function FAQ() {
  const [openIndex, setOpenIndex] = useState<number | null>(0);

  return (
    <section
      id="faq"
      className="section-spacing"
      style={{ background: "var(--bg-secondary)" }}
    >
      <div className="mx-auto max-w-[1120px] px-6">
        <div className="grid gap-12 md:grid-cols-5">
          {/* Left heading — editorial */}
          <div className="md:col-span-2">
            <FadeIn>
              <p className="micro-label" style={{ marginBottom: 16 }}>
                FAQ
              </p>
              <h2
                className="font-editorial"
                style={{
                  fontSize: "clamp(1.8rem, 4vw, 2.4rem)",
                  lineHeight: 1.1,
                  color: "var(--text-primary)",
                }}
              >
                Frequently asked{" "}
                <span className="font-editorial-italic">questions</span>
              </h2>
              <p
                style={{
                  marginTop: 14,
                  fontSize: 14,
                  lineHeight: 1.7,
                  color: "var(--text-muted)",
                }}
              >
                Everything you need to know about Driftsurge. Can&apos;t find
                what you&apos;re looking for?{" "}
                <a
                  href="/contact"
                  style={{
                    color: "var(--text-secondary)",
                    textDecoration: "underline",
                    textUnderlineOffset: 3,
                  }}
                >
                  Reach out
                </a>
                .
              </p>
            </FadeIn>
          </div>

          {/* Right accordion */}
          <div className="md:col-span-3">
            <FadeIn delay={0.1}>
              {faqs.map((faq, i) => (
                <AccordionItem
                  key={i}
                  q={faq.q}
                  a={faq.a}
                  open={openIndex === i}
                  onClick={() => setOpenIndex(openIndex === i ? null : i)}
                />
              ))}
            </FadeIn>
          </div>
        </div>
      </div>
    </section>
  );
}
