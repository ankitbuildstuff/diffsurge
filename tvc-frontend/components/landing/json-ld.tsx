import { faqs } from "@/lib/constants";

export function JsonLd() {
  const softwareApp = {
    "@context": "https://schema.org",
    "@type": "SoftwareApplication",
    name: "Diffsurge",
    description:
      "Diffsurge captures production traffic, replays it against new deployments, and surfaces breaking API changes before a single user is affected. Schema diffing, traffic replay, and drift reports in one open-source CLI.",
    url: "https://diffsurge.com",
    applicationCategory: "DeveloperApplication",
    operatingSystem: "macOS, Linux, Windows, Docker",
    offers: [
      {
        "@type": "Offer",
        price: "0",
        priceCurrency: "USD",
        name: "Free",
        description:
          "CLI tool & schema diffing — OpenAPI, GraphQL, gRPC. Free forever, no limits.",
      },
      {
        "@type": "Offer",
        price: "99",
        priceCurrency: "USD",
        name: "Pro",
        description:
          "Traffic capture, replay engine, live dashboard, and automatic PII redaction.",
      },
    ],
    featureList: [
      "OpenAPI 3.x schema diffing",
      "GraphQL SDL breaking change detection",
      "gRPC Protocol Buffers diff",
      "Production traffic capture via reverse proxy",
      "Traffic replay against staging builds",
      "Automatic PII redaction",
      "CI/CD integration with exit codes",
      "Live traffic dashboard",
      "Drift reports with severity scoring",
    ],
  };

  const faqPage = {
    "@context": "https://schema.org",
    "@type": "FAQPage",
    mainEntity: faqs.map((faq) => ({
      "@type": "Question",
      name: faq.q,
      acceptedAnswer: {
        "@type": "Answer",
        text: faq.a,
      },
    })),
  };

  const organization = {
    "@context": "https://schema.org",
    "@type": "Organization",
    name: "Diffsurge",
    url: "https://diffsurge.com",
    logo: "https://diffsurge.com/logo.svg",
    description:
      "Diffsurge is the open-source API regression testing platform. Catch breaking API changes before your users do with schema diffing, traffic replay, and drift reports.",
    sameAs: ["https://github.com/ankit12301/tvc"],
  };

  const webSite = {
    "@context": "https://schema.org",
    "@type": "WebSite",
    name: "Diffsurge",
    url: "https://diffsurge.com",
    description:
      "Catch breaking API changes before your users do. Schema diffing, traffic replay, and drift reports in one CLI.",
    potentialAction: {
      "@type": "SearchAction",
      target: "https://diffsurge.com/docs?q={search_term_string}",
      "query-input": "required name=search_term_string",
    },
  };

  return (
    <>
      <script
        type="application/ld+json"
        dangerouslySetInnerHTML={{ __html: JSON.stringify(softwareApp) }}
      />
      <script
        type="application/ld+json"
        dangerouslySetInnerHTML={{ __html: JSON.stringify(faqPage) }}
      />
      <script
        type="application/ld+json"
        dangerouslySetInnerHTML={{ __html: JSON.stringify(organization) }}
      />
      <script
        type="application/ld+json"
        dangerouslySetInnerHTML={{ __html: JSON.stringify(webSite) }}
      />
    </>
  );
}
