"use client";

import { useState, useEffect } from "react";
import { cn } from "@/lib/utils";
import { navLinks, siteConfig } from "@/lib/constants";
import { Button } from "@/components/ui/button";
import { Menu, X } from "lucide-react";
import { createClient } from "@/lib/supabase/client";
import Link from "next/link";

export function Header() {
  const [open, setOpen] = useState(false);
  const [isLoggedIn, setIsLoggedIn] = useState(false);

  useEffect(() => {
    const supabase = createClient();
    supabase.auth.getUser().then(({ data: { user } }) => {
      setIsLoggedIn(!!user);
    });
  }, []);

  return (
    <header className="fixed top-0 left-0 right-0 z-50 border-b border-zinc-100 bg-white/80 backdrop-blur-xl">
      <div className="mx-auto flex h-14 max-w-[1200px] items-center justify-between px-6">
        <Link href="/" className="flex items-center gap-2">
          <svg width="26" height="26" viewBox="0 0 28 28" fill="none">
            <rect width="28" height="28" rx="6" fill="#18181B" />
            <path d="M7 10l7-4 7 4-7 4-7-4z" fill="#A1A1AA" />
            <path d="M7 14l7 4 7-4" stroke="#fff" strokeWidth="1.5" />
            <path d="M7 18l7 4 7-4" stroke="#71717A" strokeWidth="1.5" />
          </svg>
          <span className="text-[15px] font-semibold tracking-tight">
            {siteConfig.name}
          </span>
        </Link>

        <nav className="hidden items-center gap-0.5 md:flex">
          {navLinks.map((l) => (
            <a
              key={l.href}
              href={l.href}
              className="rounded-md px-3 py-1.5 text-[13px] text-zinc-500 transition-colors hover:text-zinc-900"
            >
              {l.label}
            </a>
          ))}
        </nav>

        <div className="hidden items-center gap-3 md:flex">
          {isLoggedIn ? (
            <Link href="/dashboard">
              <Button size="sm" className="btn-gradient border-0 text-white h-8 px-4 text-[13px]">
                Dashboard
              </Button>
            </Link>
          ) : (
            <>
              <Link
                href="/login"
                className="text-[13px] text-zinc-500 hover:text-zinc-900 transition-colors"
              >
                Log in
              </Link>
              <Link href="/signup">
                <Button size="sm" className="btn-gradient border-0 text-white h-8 px-4 text-[13px]">
                  Get Started
                </Button>
              </Link>
            </>
          )}
        </div>

        <button
          onClick={() => setOpen(!open)}
          className="flex h-8 w-8 items-center justify-center rounded-md text-zinc-500 hover:text-zinc-900 md:hidden"
          aria-label="Toggle menu"
        >
          {open ? <X size={18} /> : <Menu size={18} />}
        </button>
      </div>

      <div
        className={cn(
          "overflow-hidden border-t border-zinc-100 bg-white transition-all duration-200 md:hidden",
          open ? "max-h-96" : "max-h-0"
        )}
      >
        <div className="flex flex-col gap-0.5 px-6 py-3">
          {navLinks.map((l) => (
            <a
              key={l.href}
              href={l.href}
              onClick={() => setOpen(false)}
              className="rounded-md px-3 py-2.5 text-sm text-zinc-600 hover:bg-zinc-50 hover:text-zinc-900"
            >
              {l.label}
            </a>
          ))}
          <hr className="my-2 border-zinc-100" />
          {isLoggedIn ? (
            <Link href="/dashboard">
              <Button size="sm" className="btn-gradient border-0 text-white w-full">
                Dashboard
              </Button>
            </Link>
          ) : (
            <Link href="/signup">
              <Button size="sm" className="btn-gradient border-0 text-white w-full">
                Get Started
              </Button>
            </Link>
          )}
        </div>
      </div>
    </header>
  );
}
