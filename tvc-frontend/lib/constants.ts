export const siteConfig = {
  name: "Driftsurge",
  tagline: "Catch breaking API changes before your users do",
  description:
    "Driftsurge captures production traffic, replays it against new deployments, and surfaces breaking changes — before a single user is affected. Schema diffing, traffic replay, and drift reports in one CLI.",
  url: "https://driftsurge.dev",
  github: "https://github.com/driftsurge/driftsurge",
  docs: "https://docs.driftsurge.dev",
} as const;

export const navLinks = [
  { label: "Features", href: "#features" },
  { label: "How it Works", href: "#how-it-works" },
  { label: "Pricing", href: "#pricing" },
  { label: "Docs", href: "/docs" },
  { label: "GitHub", href: "https://github.com/driftsurge/driftsurge" },
] as const;

export const stats = [
  { value: "< 5ms", label: "Proxy latency overhead" },
  { value: "1,000+", label: "Requests per second replay" },
  { value: "99.9%", label: "Uptime SLA for proxy" },
  { value: "5 min", label: "Setup to first diff" },
] as const;

export const plans = [
  {
    name: "Free",
    price: 0,
    description: "Schema governance for every team. No limits, no expiry.",
    features: [
      "CLI tool & schema diffing",
      "OpenAPI 3.x, GraphQL, gRPC support",
      "CI/CD integration (GitHub Actions, GitLab CI)",
      "Breaking change detection with exit codes",
      "JSON & YAML diff reports",
      "Community support via GitHub",
    ],
    cta: "Get Started Free",
    highlighted: false,
  },
  {
    name: "Pro",
    price: 99,
    description: "Traffic capture, replay, and a live dashboard for production.",
    features: [
      "Everything in Free",
      "100K traffic logs / month",
      "50 replay sessions / month",
      "Automatic PII redaction (email, phone, CC)",
      "Live traffic dashboard with filtering",
      "Drift reports with severity scoring",
      "Email support with 24h response",
    ],
    cta: "Start Pro Trial",
    highlighted: true,
  },
  {
    name: "Enterprise",
    price: 499,
    description: "Unlimited traffic, dedicated support, and compliance tooling.",
    features: [
      "Everything in Pro",
      "Unlimited traffic capture & replays",
      "Custom PII detection rules",
      "SSO / SAML authentication",
      "Audit log export (PDF / CSV)",
      "Priority support with SLA",
      "Dedicated onboarding",
    ],
    cta: "Contact Sales",
    highlighted: false,
  },
] as const;

export const faqs = [
  {
    q: "What is Driftsurge and how does it work?",
    a: "Driftsurge is a developer infrastructure tool that acts as a safety net between your staging and production environments. It works in three stages: First, the CLI detects breaking changes in your API schemas (OpenAPI, GraphQL, gRPC) during CI. Second, a lightweight reverse proxy captures a sample of your production traffic. Third, the replay engine fires that traffic at your staging build and compares every response, surfacing any drift before you deploy.",
  },
  {
    q: "What types of breaking changes can Driftsurge detect?",
    a: "Driftsurge catches static schema issues — removed required fields, type changes (string → number), deleted endpoints, new required parameters — as well as runtime differences like changed response shapes, missing fields in JSON bodies, status code mismatches, and latency regressions. The diff engine uses semantic comparison, so it ignores cosmetic differences like whitespace or field ordering.",
  },
  {
    q: "How does the traffic proxy affect my API latency?",
    a: "The proxy adds less than 5 ms of latency at p95. It uses an async capture pipeline — request/response pairs are written to a buffered channel and processed by a background worker pool, so the forwarding path is never blocked by storage I/O. You can also configure the sampling rate (e.g., capture 10% of traffic) to reduce overhead further.",
  },
  {
    q: "Does Driftsurge handle sensitive data?",
    a: "Yes. The proxy includes automatic PII redaction that runs before any data is stored. It detects and masks email addresses, phone numbers, credit card numbers, and SSNs using configurable regex patterns. On the Enterprise plan, you can define custom detection rules for domain-specific sensitive fields.",
  },
  {
    q: "Can I integrate Driftsurge into my existing CI/CD pipeline?",
    a: "Absolutely. The CLI is a single binary that runs on macOS, Linux, and Windows. Add `driftsurge schema diff --fail-on-breaking` to your GitHub Actions, GitLab CI, Jenkins, or CircleCI workflow. It returns exit code 0 for no changes, 1 for breaking changes, and 2 for errors — so your pipeline blocks automatically when a breaking change is detected.",
  },
  {
    q: "What API schema formats are supported?",
    a: "Driftsurge supports OpenAPI 3.0 and 3.1 (Swagger), GraphQL SDL, gRPC Protocol Buffers, and raw JSON schemas. The diff engine normalises paths and types across formats, so you get consistent breaking-change detection regardless of your API technology.",
  },
  {
    q: "How is the replay engine different from load testing?",
    a: "Load testing checks if your system handles volume. Replay testing checks if your system returns correct responses. Driftsurge replays real production traffic — with real data shapes, edge cases, and parameter combinations — against your staging build and semantically compares every response. It finds the bugs that unit tests and synthetic test suites miss.",
  },
  {
    q: "Is Driftsurge open source?",
    a: "The CLI and schema diffing engine are open source under the MIT licence. The traffic proxy, replay engine, and dashboard are available on the Pro and Enterprise plans. You can self-host the entire stack or use our managed service.",
  },
] as const;
