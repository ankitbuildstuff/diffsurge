"use client";

import { useRouter } from "next/navigation";
import { useState } from "react";
import { LogOut, User, Building2 } from "lucide-react";
import { createClient } from "@/lib/supabase/client";
import { useOrganization } from "@/lib/providers/organization-provider";

export function DashboardHeader() {
  const { user, activeOrg } = useOrganization();
  const [showMenu, setShowMenu] = useState(false);
  const router = useRouter();

  async function handleLogout() {
    const supabase = createClient();
    await supabase.auth.signOut();
    router.push("/login");
    router.refresh();
  }

  return (
    <header className="flex h-14 items-center justify-between border-b border-zinc-100 bg-white px-6">
      {activeOrg ? (
        <div className="flex items-center gap-2 text-[13px] text-zinc-500">
          <Building2 size={14} className="text-zinc-400" />
          <span className="font-medium text-zinc-700">{activeOrg.name}</span>
        </div>
      ) : (
        <div />
      )}

      <div className="relative">
        <button
          onClick={() => setShowMenu(!showMenu)}
          className="flex items-center gap-2 rounded-lg px-2 py-1.5 text-sm text-zinc-600 transition-colors hover:bg-zinc-50"
        >
          <div className="flex h-7 w-7 items-center justify-center rounded-full bg-zinc-100">
            <User size={14} className="text-zinc-500" />
          </div>
          <span className="hidden text-[13px] sm:inline">
            {user?.email ?? ""}
          </span>
        </button>

        {showMenu && (
          <>
            <div
              className="fixed inset-0 z-40"
              onClick={() => setShowMenu(false)}
            />
            <div className="absolute right-0 top-full z-50 mt-1 w-56 rounded-lg border border-zinc-100 bg-white py-1 shadow-lg">
              {activeOrg && (
                <div className="border-b border-zinc-100 px-3 py-2">
                  <p className="text-[11px] font-medium uppercase tracking-wider text-zinc-400">
                    Organization
                  </p>
                  <p className="mt-0.5 text-[13px] font-medium text-zinc-700">
                    {activeOrg.name}
                  </p>
                </div>
              )}
              <button
                onClick={handleLogout}
                className="flex w-full items-center gap-2 px-3 py-2 text-left text-[13px] text-zinc-600 transition-colors hover:bg-zinc-50"
              >
                <LogOut size={14} />
                Sign out
              </button>
            </div>
          </>
        )}
      </div>
    </header>
  );
}
