import type { Metadata } from "next";
import { Inter } from "next/font/google";
import "./globals.css";

const inter = Inter({
  variable: "--font-inter",
  subsets: ["latin"],
  display: "swap",
});

export const metadata: Metadata = {
  title: "Driftsurge — Catch breaking API changes before your users do",
  description:
    "Driftsurge captures production traffic, replays it against new deployments, and surfaces breaking changes before a single user is affected. Schema diffing, traffic replay, and drift reports in one CLI.",
  keywords: [
    "API testing",
    "API breaking changes",
    "traffic replay",
    "schema diffing",
    "API versioning",
    "OpenAPI diff",
    "GraphQL breaking changes",
    "developer tools",
    "CI/CD integration",
    "API regression testing",
  ],
  icons: {
    icon: "/logo.svg",
    apple: "/logo.svg",
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
