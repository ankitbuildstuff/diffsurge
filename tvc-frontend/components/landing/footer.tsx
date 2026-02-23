import { siteConfig } from "@/lib/constants";

const footerLinks = {
  Product: [
    { label: "Features", href: "#features" },
    { label: "Pricing", href: "#pricing" },
    { label: "Changelog", href: "/changelog" },
    { label: "Roadmap", href: "/roadmap" },
  ],
  Developers: [
    { label: "Documentation", href: "/docs" },
    { label: "API Reference", href: "/docs/api" },
    { label: "CLI Reference", href: "/docs/cli" },
    { label: "GitHub", href: siteConfig.github },
  ],
  Company: [
    { label: "About", href: "/about" },
    { label: "Blog", href: "/blog" },
    { label: "Contact", href: "/contact" },
  ],
  Legal: [
    { label: "Privacy", href: "/privacy" },
    { label: "Terms", href: "/terms" },
    { label: "Security", href: "/security" },
  ],
};

export function Footer() {
  return (
    <footer className="border-t border-zinc-800 bg-zinc-950">
      <div className="mx-auto max-w-[1200px] px-6 py-14">
        <div className="grid gap-10 sm:grid-cols-2 md:grid-cols-5">
          <div className="md:col-span-1">
            <a href="/" className="flex items-center gap-2">
              <svg width="22" height="22" viewBox="0 0 28 28" fill="none">
                <rect width="28" height="28" rx="6" fill="#18181B" />
                <path d="M7 10l7-4 7 4-7 4-7-4z" fill="#A1A1AA" />
                <path d="M7 14l7 4 7-4" stroke="#fff" strokeWidth="1.5" />
                <path d="M7 18l7 4 7-4" stroke="#71717A" strokeWidth="1.5" />
              </svg>
              <span className="text-[14px] font-semibold text-zinc-100">
                {siteConfig.name}
              </span>
            </a>
            <p className="mt-3 text-[12px] leading-relaxed text-zinc-500">
              Catch breaking API changes
              <br />
              before your users do.
            </p>
          </div>

          {Object.entries(footerLinks).map(([cat, links]) => (
            <div key={cat}>
              <h4 className="text-[11px] font-semibold uppercase tracking-wider text-zinc-500">
                {cat}
              </h4>
              <ul className="mt-4 space-y-2">
                {links.map((l) => (
                  <li key={l.label}>
                    <a
                      href={l.href}
                      className="text-[13px] text-zinc-400 transition-colors hover:text-zinc-100"
                    >
                      {l.label}
                    </a>
                  </li>
                ))}
              </ul>
            </div>
          ))}
        </div>

        <div className="mt-12 flex flex-col items-center justify-between gap-3 border-t border-zinc-800 pt-7 sm:flex-row">
          <p className="text-[11px] text-zinc-600">
            &copy; {new Date().getFullYear()} Driftsurge. All rights reserved.
          </p>
          <div className="flex gap-5">
            <a
              href={siteConfig.github}
              className="text-[11px] text-zinc-600 hover:text-zinc-300 transition-colors"
            >
              GitHub
            </a>
            <a
              href="https://x.com/driftsurge"
              className="text-[11px] text-zinc-600 hover:text-zinc-300 transition-colors"
            >
              X / Twitter
            </a>
          </div>
        </div>
      </div>
    </footer>
  );
}
