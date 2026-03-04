import type { Metadata } from "next";
import { Inter } from "next/font/google";
import "./globals.css";

const inter = Inter({
  variable: "--font-inter",
  subsets: ["latin"],
  display: "swap",
});

export const metadata: Metadata = {
  metadataBase: new URL("https://diffsurge.com"),
  title: "Diffsurge — Catch breaking API changes before your users do",
  description:
    "Diffsurge captures production traffic, replays it against new deployments, and surfaces breaking changes before a single user is affected. Schema diffing, traffic replay, and drift reports in one CLI.",
  keywords: [
    "API testing",
    "API breaking changes",
    "traffic replay",
    "schema diffing",
    "API versioning",
    "OpenAPI diff",
    "OpenAPI diff CLI",
    "GraphQL breaking changes",
    "GraphQL schema comparison",
    "gRPC proto diff",
    "developer tools",
    "CI/CD integration",
    "API regression testing",
    "API regression testing tool",
    "API contract testing",
    "API drift detection",
    "production traffic replay",
    "API compatibility checker",
    "API schema validation",
    "breaking change detection",
    "REST API testing",
    "API monitoring",
    "API governance",
    "schema breaking change",
    "diffsurge",
  ],
  icons: {
    icon: "/logo.svg",
    apple: "/logo.svg",
  },
  openGraph: {
    title: "Diffsurge — Catch breaking API changes before your users do",
    description:
      "Schema diffing, traffic replay, and drift reports in one CLI. Diffsurge surfaces breaking API changes before a single user is affected.",
    url: "https://diffsurge.com",
    siteName: "Diffsurge",
    type: "website",
    locale: "en_US",
  },
  twitter: {
    card: "summary_large_image",
    title: "Diffsurge — Catch breaking API changes before your users do",
    description:
      "Schema diffing, traffic replay, and drift reports in one CLI. Diffsurge surfaces breaking API changes before a single user is affected.",
  },
  alternates: {
    canonical: "https://diffsurge.com",
  },
  robots: {
    index: true,
    follow: true,
    googleBot: {
      index: true,
      follow: true,
      "max-video-preview": -1,
      "max-image-preview": "large",
      "max-snippet": -1,
    },
  },
};

export default function RootLayout({
  children,
}: Readonly<{
  children: React.ReactNode;
}>) {
  return (
    <html lang="en" className="scroll-smooth" style={{ colorScheme: "light" }}>
      <head>
        <link rel="preconnect" href="https://fonts.googleapis.com" />
        <link
          rel="preconnect"
          href="https://fonts.gstatic.com"
          crossOrigin="anonymous"
        />
        <link
          href="https://fonts.googleapis.com/css2?family=Instrument+Serif:ital@0;1&display=swap"
          rel="stylesheet"
        />
      </head>
      <body
        className={`${inter.variable} antialiased`}
        style={{
          fontFamily:
            'var(--font-inter), "Inter", system-ui, -apple-system, sans-serif',
          backgroundColor: "#FAF9F6",
          color: "#2D2926",
        }}
      >
        {children}
      </body>
    </html>
  );
}
