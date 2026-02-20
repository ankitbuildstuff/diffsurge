"use client";

import { useState } from "react";
import { cn } from "@/lib/utils";
import { navLinks, siteConfig } from "@/lib/constants";
import { Button } from "@/components/ui/button";
import { Menu, X } from "lucide-react";

export function Header() {
  const [open, setOpen] = useState(false);

  return (
    <header className="fixed top-0 left-0 right-0 z-50 border-b border-zinc-100 bg-white/80 backdrop-blur-xl">
      <div className="mx-auto flex h-14 max-w-[1200px] items-center justify-between px-6">
        <a href="/" className="flex items-center gap-2">
          <svg width="26" height="26" viewBox="0 0 28 28" fill="none">
            <rect width="28" height="28" rx="7" fill="#09090b" />
            <path
              d="M8 10L14 7L20 10V18L14 21L8 18V10Z"
              stroke="white"
              strokeWidth="1.5"
              strokeLinejoin="round"
            />
            <path d="M14 14V21" stroke="white" strokeWidth="1.5" />
            <path d="M8 10L14 14L20 10" stroke="white" strokeWidth="1.5" />
          </svg>
          <span className="text-[15px] font-semibold tracking-tight">
            {siteConfig.name}
          </span>
        </a>

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
          <a
            href="/login"
            className="text-[13px] text-zinc-500 hover:text-zinc-900 transition-colors"
          >
            Log in
          </a>
          <Button size="sm" className="btn-gradient border-0 text-white h-8 px-4 text-[13px]">
            Get Started
          </Button>
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
          <Button size="sm" className="btn-gradient border-0 text-white">
            Get Started
          </Button>
        </div>
      </div>
    </header>
  );
}
