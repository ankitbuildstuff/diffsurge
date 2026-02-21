import { Sidebar } from "@/components/dashboard/sidebar";
import { DashboardHeader } from "@/components/dashboard/header";
import { QueryProvider } from "@/lib/providers/query-provider";
import { OrganizationProvider } from "@/lib/providers/organization-provider";
import { Toaster } from "sonner";

export default function DashboardLayout({
  children,
}: {
  children: React.ReactNode;
}) {
  return (
    <QueryProvider>
      <OrganizationProvider>
        <div className="flex h-screen overflow-hidden">
          <Sidebar />
          <div className="flex flex-1 flex-col overflow-hidden">
            <DashboardHeader />
            <main className="flex-1 overflow-y-auto bg-zinc-50 p-6">
              {children}
            </main>
          </div>
        </div>
        <Toaster position="bottom-right" richColors />
      </OrganizationProvider>
    </QueryProvider>
  );
}
