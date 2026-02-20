import type { Metadata } from "next";
import { Geist, Geist_Mono } from "next/font/google";
import "./globals.css";

const geistSans = Geist({
  variable: "--font-geist-sans",
  subsets: ["latin"],
});

const geistMono = Geist_Mono({
  variable: "--font-geist-mono",
  subsets: ["latin"],
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
};

export default function RootLayout({
  children,
}: Readonly<{
  children: React.ReactNode;
}>) {
  return (
    <html lang="en" className="scroll-smooth" style={{ colorScheme: "light" }}>
      <body
        className={`${geistSans.variable} ${geistMono.variable} font-sans antialiased bg-white text-zinc-950`}
      >
        {children}
      </body>
    </html>
  );
}
