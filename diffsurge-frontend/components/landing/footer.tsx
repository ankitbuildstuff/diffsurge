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
    { label: "GitHub", href: "https://github.com/ankit12301/diffsurge" },
    { label: "Docker Hub", href: "https://hub.docker.com/u/equixankit" },
    { label: "npm", href: "https://www.npmjs.com/package/diffsurge" },
  ],
  Company: [{ label: "Contact", href: "/contact" }],
};

export function Footer() {
  return (
    <footer style={{ background: "var(--bg-dark)" }}>
      {/* Subtle separator */}
      <div style={{ height: 1, background: "var(--border-dark)" }} />

      <div
        className="mx-auto px-6"
        style={{ maxWidth: 1200, paddingTop: 64, paddingBottom: 64 }}
      >
        <div className="grid gap-12 sm:grid-cols-2 md:grid-cols-5">
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
                <rect width="28" height="28" rx="6" fill="#222" />
                <path d="M7 10l7-4 7 4-7 4-7-4z" fill="#888" />
                <path d="M7 14l7 4 7-4" stroke="#fff" strokeWidth="1.5" />
                <path d="M7 18l7 4 7-4" stroke="#888" strokeWidth="1.5" />
              </svg>
              <span
                style={{
                  fontSize: 16,
                  fontWeight: 500,
                  color: "var(--text-on-dark)",
                }}
              >
                {siteConfig.name}
              </span>
            </a>
            <p
              style={{
                marginTop: 12,
                fontSize: 13,
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
                  fontSize: 11,
                }}
              >
                {cat}
              </h4>
              <ul
                style={{
                  marginTop: 16,
                  display: "flex",
                  flexDirection: "column",
                  gap: 10,
                }}
              >
                {links.map((l) => (
                  <li key={l.label}>
                    <a
                      href={l.href}
                      style={{
                        fontSize: 13,
                        color: "rgba(255, 255, 255, 0.4)",
                        transition: "color 0.2s ease",
                        textDecoration: "none",
                      }}
                      onMouseEnter={(e) =>
                        (e.currentTarget.style.color =
                          "var(--text-on-dark)")
                      }
                      onMouseLeave={(e) =>
                        (e.currentTarget.style.color =
                          "rgba(255, 255, 255, 0.4)")
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
              fontSize: 12,
              color: "var(--text-on-dark-muted)",
            }}
          >
            &copy; {new Date().getFullYear()} Diffsurge. All rights reserved.
          </p>
          <a
            href="https://hub.docker.com/u/equixankit"
            style={{
              fontSize: 12,
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
