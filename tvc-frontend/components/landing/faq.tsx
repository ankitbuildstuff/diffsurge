"use client";

import { useState } from "react";
import { cn } from "@/lib/utils";
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
    <div className="border-b border-zinc-100">
      <button
        onClick={onClick}
        className="flex w-full items-center justify-between py-5 text-left cursor-pointer"
      >
        <span className="text-[14px] font-medium text-zinc-900 pr-4">{q}</span>
        <ChevronDown
          size={16}
          className={cn(
            "shrink-0 text-zinc-400 transition-transform duration-200",
            open && "rotate-180"
          )}
        />
      </button>
      <div
        className={cn(
          "overflow-hidden transition-all duration-300 ease-out",
          open ? "max-h-[500px] pb-5" : "max-h-0"
        )}
      >
        <p className="text-[13px] leading-[1.7] text-zinc-500">{a}</p>
      </div>
    </div>
  );
}

export function FAQ() {
  const [openIndex, setOpenIndex] = useState<number | null>(0);

  return (
    <section id="faq" className="bg-white py-20 md:py-28">
      <div className="mx-auto max-w-[1200px] px-6">
        <div className="grid gap-12 md:grid-cols-5">
          {/* Left heading */}
          <div className="md:col-span-2">
            <FadeIn>
              <p className="text-[12px] font-medium uppercase tracking-widest text-teal-600">
                FAQ
              </p>
              <h2 className="mt-3 text-[1.75rem] font-bold tracking-tight sm:text-3xl">
                Frequently asked questions
              </h2>
              <p className="mt-3 text-[14px] leading-relaxed text-zinc-500">
                Everything you need to know about Driftsurge. Can&apos;t find
                what you&apos;re looking for?{" "}
                <a
                  href="/contact"
                  className="text-teal-600 hover:text-teal-700 underline underline-offset-2"
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
