"use client";

import { FadeIn } from "@/components/ui/fade-in";

const steps = [
  {
    number: "01",
    title: "Install the CLI",
    description:
      "One command to install. One command to integrate. Driftsurge outputs machine-readable JSON and human-readable reports — perfect for CI/CD gates and local development alike.",
    code: `$ npm install -g @driftsurge/cli
$ driftsurge init
  ✓ Config created: .driftsurge.yaml
  ✓ GitHub Actions template generated

$ driftsurge schema diff \\
    --old api-v1.yaml \\
    --new api-v2.yaml \\
    --fail-on-breaking

  ✓ 45 endpoints compared
  ✗ 1 breaking change detected
  ▸ Report saved to ./drift-report.json`,
  },
  {
    number: "02",
    title: "Capture production traffic",
    description:
      "Deploy the proxy as a sidecar or standalone container. It samples real traffic, strips PII, and buffers asynchronously — adding less than 5 ms of latency to the request path.",
    code: `$ driftsurge proxy start \\
    --target api.example.com \\
    --port 8080 \\
    --sample-rate 0.1

  ▸ Proxy listening on :8080
  ▸ Forwarding to api.example.com
  ▸ Sampling 10% of traffic
  ▸ PII redaction: enabled
  ▸ Buffer: 10,000 slots / 20 workers
  ▸ Press Ctrl+C to stop`,
  },
  {
    number: "03",
    title: "Replay and compare",
    description:
      "Point the replay engine at your staging build. It fires captured requests at configurable concurrency, semantically compares every response, and produces a drift report — sorted by severity.",
    code: `$ driftsurge replay \\
    --source prod \\
    --target staging \\
    --concurrency 200

  Replaying 1,247 requests...
  ████████████████████████████ 100%

  ✓ 1,241 responses matched (99.5%)
  ⚠ 4 warnings  (type coercion)
  ✗ 2 breaking   (missing fields)
  ▸ Full report → https://app.driftsurge.dev/r/3f7a2c`,
  },
];

export function HowItWorks() {
  return (
    <section id="how-it-works" className="bg-zinc-50/60 py-20 md:py-28">
      <div className="mx-auto max-w-[1200px] px-6">
        <FadeIn>
          <p className="text-[12px] font-medium uppercase tracking-widest text-teal-600">
            How it works
          </p>
          <h2 className="mt-3 text-[1.75rem] font-bold tracking-tight sm:text-3xl">
            Three steps to safer deployments
          </h2>
          <p className="mt-3 max-w-lg text-[14px] leading-relaxed text-zinc-500">
            Go from zero to production-grade API regression testing in minutes.
          </p>
        </FadeIn>

        <div className="mt-14 space-y-6">
          {steps.map((step, i) => (
            <FadeIn key={step.number} delay={i * 0.08}>
              <div className="grid items-start gap-8 rounded-2xl border border-zinc-200 bg-white p-6 shadow-[0_1px_3px_rgba(0,0,0,0.04)] md:grid-cols-5 md:p-8">
                {/* Text */}
                <div className="md:col-span-2">
                  <div className="flex items-center gap-3">
                    <span className="inline-flex h-8 w-8 items-center justify-center rounded-full bg-zinc-100 text-[13px] font-bold text-zinc-400">
                      {step.number}
                    </span>
                    <h3 className="text-[15px] font-semibold text-zinc-900">
                      {step.title}
                    </h3>
                  </div>
                  <p className="mt-3 text-[13px] leading-[1.7] text-zinc-500">
                    {step.description}
                  </p>
                </div>

                {/* Code */}
                <div className="md:col-span-3">
                  <div className="overflow-hidden rounded-xl border border-zinc-200">
                    <div className="flex items-center gap-1.5 border-b border-zinc-100 bg-zinc-50 px-3 py-2">
                      <div className="h-2 w-2 rounded-full bg-zinc-300" />
                      <div className="h-2 w-2 rounded-full bg-zinc-300" />
                      <div className="h-2 w-2 rounded-full bg-zinc-300" />
                    </div>
                    <pre className="overflow-x-auto bg-[#0a0a0f] p-4 font-mono text-[11.5px] leading-[1.8] text-zinc-400">
                      {step.code}
                    </pre>
                  </div>
                </div>
              </div>
            </FadeIn>
          ))}
        </div>
      </div>
    </section>
  );
}
