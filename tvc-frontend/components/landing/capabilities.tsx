"use client";

import { FadeIn } from "@/components/ui/fade-in";
import {
  GitBranch,
  Webhook,
  FileJson,
  ShieldCheck,
  Zap,
  Layers,
} from "lucide-react";

const scenarios = [
  "A required field is removed from a response body",
  "A field type changes — string becomes number",
  "An endpoint is renamed or deleted",
  "A new required query parameter is added",
  "A nested object structure changes shape",
  "Response time regresses beyond a threshold",
  "A status code changes for identical requests",
  "PII appears in a field that was previously clean",
];

const integrations = [
  { name: "GitHub Actions", icon: <GitBranch size={16} /> },
  { name: "GitLab CI", icon: <GitBranch size={16} /> },
  { name: "Jenkins", icon: <Layers size={16} /> },
  { name: "CircleCI", icon: <Zap size={16} /> },
  { name: "OpenAPI 3.x", icon: <FileJson size={16} /> },
  { name: "GraphQL SDL", icon: <Webhook size={16} /> },
  { name: "gRPC / Proto", icon: <Layers size={16} /> },
  { name: "JSON Schema", icon: <FileJson size={16} /> },
];

export function Capabilities() {
  return (
    <section className="bg-white py-20 md:py-28">
      <div className="mx-auto max-w-[1200px] px-6">
        <div className="grid gap-16 md:grid-cols-2">
          {/* Left — what it catches */}
          <FadeIn>
            <div>
              <p className="text-[12px] font-medium uppercase tracking-widest text-teal-600">
                What Driftsurge catches
              </p>
              <h2 className="mt-3 text-[1.75rem] font-bold tracking-tight sm:text-3xl">
                The breaking changes that slip through unit tests
              </h2>
              <p className="mt-4 text-[14px] leading-[1.7] text-zinc-500">
                Unit tests verify logic. Integration tests verify contracts.
                But neither tests against the thousands of real-world payload
                shapes your API handles every day. Driftsurge replays actual
                production traffic — with real edge cases, nested structures,
                and parameter combinations — against your new build. When the
                diff engine finds a mismatch, you see it before your users do.
              </p>

              <div className="mt-8 grid grid-cols-1 gap-2">
                {scenarios.map((s, i) => (
                  <div
                    key={i}
                    className="flex items-center gap-2.5 rounded-lg border border-zinc-100 bg-zinc-50/50 px-3.5 py-2.5"
                  >
                    <ShieldCheck
                      size={14}
                      className="shrink-0 text-teal-500"
                    />
                    <span className="text-[13px] text-zinc-600">{s}</span>
                  </div>
                ))}
              </div>
            </div>
          </FadeIn>

          {/* Right — integrations + SEO text */}
          <FadeIn delay={0.15}>
            <div>
              <p className="text-[12px] font-medium uppercase tracking-widest text-teal-600">
                Integrations & formats
              </p>
              <h2 className="mt-3 text-[1.75rem] font-bold tracking-tight sm:text-3xl">
                Works with your stack, not against it
              </h2>
              <p className="mt-4 text-[14px] leading-[1.7] text-zinc-500">
                Driftsurge is a single binary that runs anywhere — macOS,
                Linux, Windows, Docker. Drop it into your existing CI/CD
                pipeline with one line. The CLI returns standard exit codes
                (0 = clean, 1 = breaking, 2 = error) so your workflow blocks
                automatically when a breaking change is detected. No SDK, no
                agent, no code changes.
              </p>
              <p className="mt-3 text-[14px] leading-[1.7] text-zinc-500">
                On the infrastructure side, the traffic proxy deploys as a
                sidecar or standalone container. It supports path-based routing,
                per-route sampling rates, and hot-reload configuration — so you
                can go from zero to capturing production traffic in under five
                minutes.
              </p>

              <div className="mt-8 grid grid-cols-2 gap-3">
                {integrations.map((int) => (
                  <div
                    key={int.name}
                    className="flex items-center gap-2.5 rounded-lg border border-zinc-100 bg-zinc-50/50 px-3.5 py-2.5"
                  >
                    <span className="text-zinc-400">{int.icon}</span>
                    <span className="text-[13px] font-medium text-zinc-600">
                      {int.name}
                    </span>
                  </div>
                ))}
              </div>

              <p className="mt-6 text-[14px] leading-[1.7] text-zinc-500">
                Whether you&apos;re running a monolith behind Nginx or a
                microservices mesh on Kubernetes, Driftsurge slots into your
                architecture. The diff engine normalises paths, types, and
                field ordering across OpenAPI, GraphQL, gRPC, and raw JSON —
                giving you consistent breaking-change detection regardless of
                your API technology.
              </p>
            </div>
          </FadeIn>
        </div>
      </div>
    </section>
  );
}
