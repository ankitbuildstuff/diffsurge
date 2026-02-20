"use client";

import { Check } from "lucide-react";
import { cn } from "@/lib/utils";
import { plans } from "@/lib/constants";
import { Button } from "@/components/ui/button";
import { FadeIn } from "@/components/ui/fade-in";
import { SpotlightCard } from "@/components/ui/spotlight-card";

export function Pricing() {
  return (
    <section id="pricing" className="bg-white py-20 md:py-28">
      <div className="mx-auto max-w-[1200px] px-6">
        <FadeIn>
          <p className="text-[12px] font-medium uppercase tracking-widest text-teal-600">
            Pricing
          </p>
          <h2 className="mt-3 text-[1.75rem] font-bold tracking-tight sm:text-3xl">
            Start free, scale when you&apos;re ready
          </h2>
          <p className="mt-3 max-w-lg text-[14px] leading-relaxed text-zinc-500">
            The CLI is free forever — no limits, no expiry. Add traffic capture
            and replay when your team needs production-grade testing.
          </p>
        </FadeIn>

        <div className="mt-12 grid gap-5 md:grid-cols-3">
          {plans.map((plan, i) => (
            <FadeIn key={plan.name} delay={i * 0.08}>
              <SpotlightCard
                className={cn(
                  "h-full",
                  plan.highlighted &&
                    "border-teal-200 shadow-[0_4px_24px_rgba(13,148,136,0.08)]"
                )}
                spotlightColor={
                  plan.highlighted
                    ? "rgba(13,148,136,0.08)"
                    : "rgba(0,0,0,0.03)"
                }
              >
                <div className="relative flex h-full flex-col p-6">
                  {plan.highlighted && (
                    <span className="absolute -top-2.5 left-5 rounded-full bg-gradient-to-r from-teal-500 to-cyan-500 px-3 py-0.5 text-[11px] font-medium text-white">
                      Most popular
                    </span>
                  )}

                  <h3 className="text-[14px] font-semibold">{plan.name}</h3>

                  <div className="mt-3 flex items-baseline gap-1">
                    {plan.price === 0 ? (
                      <span className="text-3xl font-bold tracking-tight">
                        Free
                      </span>
                    ) : (
                      <>
                        <span className="text-3xl font-bold tracking-tight">
                          ${plan.price}
                        </span>
                        <span className="text-[13px] text-zinc-400">/mo</span>
                      </>
                    )}
                  </div>

                  <p className="mt-1.5 text-[13px] text-zinc-500">
                    {plan.description}
                  </p>

                  <ul className="mt-6 flex-1 space-y-2.5">
                    {plan.features.map((f) => (
                      <li key={f} className="flex items-start gap-2">
                        <Check
                          size={14}
                          className={cn(
                            "mt-0.5 shrink-0",
                            plan.highlighted
                              ? "text-teal-500"
                              : "text-zinc-300"
                          )}
                        />
                        <span className="text-[13px] text-zinc-600">{f}</span>
                      </li>
                    ))}
                  </ul>

                  <Button
                    variant={plan.highlighted ? "primary" : "secondary"}
                    className={cn(
                      "mt-7 w-full",
                      plan.highlighted && "btn-gradient border-0 text-white"
                    )}
                  >
                    {plan.cta}
                  </Button>
                </div>
              </SpotlightCard>
            </FadeIn>
          ))}
        </div>
      </div>
    </section>
  );
}
