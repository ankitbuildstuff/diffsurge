"use client";

import { useState, useEffect } from "react";
import { navLinks, siteConfig } from "@/lib/constants";
import { Menu, X, ArrowRight } from "lucide-react";
import { createClient } from "@/lib/supabase/client";
import Link from "next/link";

export function Header() {
  const [open, setOpen] = useState(false);
  const [isLoggedIn, setIsLoggedIn] = useState(false);
  const [scrolled, setScrolled] = useState(false);

  useEffect(() => {
    const supabase = createClient();
    supabase.auth.getUser().then(({ data: { user } }) => {
      setIsLoggedIn(!!user);
    });
  }, []);

  useEffect(() => {
    const handler = () => setScrolled(window.scrollY > 10);
    window.addEventListener("scroll", handler, { passive: true });
    return () => window.removeEventListener("scroll", handler);
  }, []);

  return (
    <header
      className="fixed top-0 left-0 right-0 z-50 transition-all duration-300"
      style={{
        background: scrolled
          ? "rgba(255, 255, 255, 0.85)"
          : "transparent",
        backdropFilter: scrolled ? "blur(12px)" : "none",
        borderBottom: scrolled
          ? "1px solid var(--border-subtle)"
          : "1px solid transparent",
      }}
    >
      <div className="mx-auto flex items-center justify-between px-6" style={{ maxWidth: 1200, height: 64 }}>
        {/* Logo */}
        <Link href="/" className="flex items-center gap-2.5">
          <svg width="24" height="24" viewBox="0 0 28 28" fill="none">
            <rect width="28" height="28" rx="6" fill="#111" />
            <path d="M7 10l7-4 7 4-7 4-7-4z" fill="#888" />
            <path d="M7 14l7 4 7-4" stroke="#fff" strokeWidth="1.5" />
            <path d="M7 18l7 4 7-4" stroke="#888" strokeWidth="1.5" />
          </svg>
          <span
            style={{
              fontSize: 17,
              fontWeight: 500,
              letterSpacing: "-0.02em",
              color: "var(--text-primary)",
            }}
          >
            {siteConfig.name}
          </span>
        </Link>

        {/* Nav links */}
        <nav className="hidden items-center gap-1 md:flex">
          {navLinks.map((l) => (
            <a
              key={l.href}
              href={l.href}
              className="rounded-md px-3 py-1.5 text-[14px] transition-colors"
              style={{ color: "var(--text-muted)" }}
              onMouseEnter={(e) =>
                (e.currentTarget.style.color = "var(--text-primary)")
              }
              onMouseLeave={(e) =>
                (e.currentTarget.style.color = "var(--text-muted)")
              }
            >
              {l.label}
            </a>
          ))}
        </nav>

        {/* Actions */}
        <div className="hidden items-center gap-3 md:flex">
          {isLoggedIn ? (
            <Link
              href="/dashboard"
              className="btn-primary"
              style={{ padding: "8px 16px", fontSize: 13 }}
            >
              Dashboard
              <ArrowRight size={14} />
            </Link>
          ) : (
            <>
              <Link
                href="/login"
                className="text-[14px] transition-colors"
                style={{ color: "var(--text-muted)" }}
                onMouseEnter={(e) =>
                  (e.currentTarget.style.color = "var(--text-primary)")
                }
                onMouseLeave={(e) =>
                  (e.currentTarget.style.color = "var(--text-muted)")
                }
              >
                Log in
              </Link>
              <Link
                href="/signup"
                className="btn-primary"
                style={{ padding: "8px 16px", fontSize: 13 }}
              >
                Get Started
                <ArrowRight size={14} />
              </Link>
            </>
          )}
        </div>

        {/* Mobile toggle */}
        <button
          onClick={() => setOpen(!open)}
          className="flex h-8 w-8 items-center justify-center rounded-md md:hidden"
          style={{ color: "var(--text-muted)" }}
          aria-label="Toggle menu"
        >
          {open ? <X size={18} /> : <Menu size={18} />}
        </button>
      </div>

      {/* Mobile menu */}
      <div
        className="overflow-hidden transition-all duration-200 md:hidden"
        style={{
          maxHeight: open ? "400px" : "0",
          background: "var(--bg-primary)",
          borderTop: open ? "1px solid var(--border-subtle)" : "none",
        }}
      >
        <div className="flex flex-col gap-1 px-6 py-4">
          {navLinks.map((l) => (
            <a
              key={l.href}
              href={l.href}
              onClick={() => setOpen(false)}
              className="rounded-md px-3 py-2.5 text-[14px] transition-colors"
              style={{ color: "var(--text-secondary)" }}
            >
              {l.label}
            </a>
          ))}
          <div
            style={{
              height: 1,
              background: "var(--border-subtle)",
              margin: "12px 0",
            }}
          />
          {isLoggedIn ? (
            <Link
              href="/dashboard"
              className="btn-primary text-center"
              style={{ fontSize: 13 }}
            >
              Dashboard
            </Link>
          ) : (
            <>
              <Link
                href="/login"
                onClick={() => setOpen(false)}
                className="rounded-md px-3 py-2.5 text-[14px]"
                style={{ color: "var(--text-secondary)" }}
              >
                Log in
              </Link>
              <Link
                href="/signup"
                onClick={() => setOpen(false)}
                className="btn-primary text-center mt-1"
                style={{ fontSize: 13 }}
              >
                Get Started
              </Link>
            </>
          )}
        </div>
      </div>
    </header>
  );
}
