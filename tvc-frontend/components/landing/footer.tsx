"use client";

import { siteConfig } from "@/lib/constants";

const footerLinks = {
  Product: [
    { label: "Features", href: "#features" },
    { label: "How it Works", href: "#how-it-works" },
    { label: "Pricing", href: "#pricing" },
    { label: "Documentation", href: "/docs" },
  ],
  Developers: [
    { label: "CLI Reference", href: "/docs" },
    { label: "GitHub", href: "https://github.com/ankit12301/tvc" },
    { label: "Docker Hub", href: "https://hub.docker.com/u/equixankit" },
    { label: "npm", href: "https://www.npmjs.com/package/diffsurge" },
  ],
  Company: [{ label: "Contact", href: "/contact" }],
};

export function Footer() {
  return (
    <footer style={{ background: "var(--bg-dark)" }}>
      {/* Thin data stripe separator */}
      <div
        className="data-stripe-wide"
        style={{ height: 2, opacity: 0.5 }}
      />

      <div className="mx-auto max-w-[1120px] px-6" style={{ paddingTop: 56, paddingBottom: 56 }}>
        <div className="grid gap-10 sm:grid-cols-2 md:grid-cols-5">
          {/* Brand */}
          <div className="md:col-span-1">
            <a
              href="/"
              style={{
                display: "flex",
                alignItems: "center",
                gap: 8,
                textDecoration: "none",
              }}
            >
              <svg width="20" height="20" viewBox="0 0 28 28" fill="none">
                <rect width="28" height="28" rx="6" fill="#1A1714" />
                <path d="M7 10l7-4 7 4-7 4-7-4z" fill="#A1A1AA" />
                <path d="M7 14l7 4 7-4" stroke="#fff" strokeWidth="1.5" />
                <path d="M7 18l7 4 7-4" stroke="#71717A" strokeWidth="1.5" />
              </svg>
              <span
                className="font-editorial"
                style={{
                  fontSize: 16,
                  color: "var(--text-on-dark)",
                }}
              >
                {siteConfig.name}
              </span>
            </a>
            <p
              style={{
                marginTop: 12,
                fontSize: 12,
                lineHeight: 1.6,
                color: "var(--text-on-dark-muted)",
              }}
            >
              Catch breaking API changes
              <br />
              before your users do.
            </p>
          </div>

          {/* Link columns */}
          {Object.entries(footerLinks).map(([cat, links]) => (
            <div key={cat}>
              <h4
                className="micro-label"
                style={{
                  color: "var(--text-on-dark-muted)",
                  fontSize: 10,
                }}
              >
                {cat}
              </h4>
              <ul style={{ marginTop: 16, display: "flex", flexDirection: "column", gap: 10 }}>
                {links.map((l) => (
                  <li key={l.label}>
                    <a
                      href={l.href}
                      style={{
                        fontSize: 13,
                        color: "rgba(232, 228, 223, 0.5)",
                        transition: "color 0.2s ease",
                        textDecoration: "none",
                      }}
                      onMouseEnter={(e) =>
                        (e.currentTarget.style.color = "var(--text-on-dark)")
                      }
                      onMouseLeave={(e) =>
                        (e.currentTarget.style.color =
                          "rgba(232, 228, 223, 0.5)")
                      }
                    >
                      {l.label}
                    </a>
                  </li>
                ))}
              </ul>
            </div>
          ))}
        </div>

        <div
          style={{
            marginTop: 48,
            paddingTop: 28,
            borderTop: "1px solid var(--border-dark)",
            display: "flex",
            flexWrap: "wrap",
            alignItems: "center",
            justifyContent: "space-between",
            gap: 12,
          }}
        >
          <p
            style={{
              fontSize: 11,
              color: "var(--text-on-dark-muted)",
            }}
          >
            &copy; {new Date().getFullYear()} Diffsurge. All rights reserved.
          </p>
          <a
            href="https://hub.docker.com/u/equixankit"
            style={{
              fontSize: 11,
              color: "var(--text-on-dark-muted)",
              transition: "color 0.2s ease",
              textDecoration: "none",
            }}
            onMouseEnter={(e) =>
              (e.currentTarget.style.color = "var(--text-on-dark)")
            }
            onMouseLeave={(e) =>
              (e.currentTarget.style.color = "var(--text-on-dark-muted)")
            }
          >
            Docker Hub
          </a>
        </div>
      </div>
    </footer>
  );
}
