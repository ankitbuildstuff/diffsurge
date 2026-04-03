"use client";

import { useState } from "react";
import { useInfiniteQuery } from "@tanstack/react-query";
import { trafficApi, type TrafficFilters } from "@/lib/api/traffic";
import { useProject } from "@/lib/providers/project-provider";
import Link from "next/link";
import { cn } from "@/lib/utils";
import { Filter, RefreshCw } from "lucide-react";
import { ErrorState } from "@/components/ui/error-state";

const HTTP_METHODS = ["GET", "POST", "PUT", "DELETE", "PATCH"];

function StatusBadge({ code }: { code: number }) {
  const color =
    code < 300
      ? "bg-emerald-50 text-emerald-700"
      : code < 400
        ? "bg-blue-50 text-blue-700"
        : code < 500
          ? "bg-amber-50 text-amber-700"
          : "bg-red-50 text-red-700";

  return (
    <span className={cn("rounded-md px-1.5 py-0.5 text-xs font-medium", color)}>
      {code}
    </span>
  );
}

function MethodBadge({ method }: { method: string }) {
  const colors: Record<string, string> = {
    GET: "text-emerald-600",
    POST: "text-blue-600",
    PUT: "text-amber-600",
    DELETE: "text-red-600",
    PATCH: "text-purple-600",
  };

  return (
    <span className={cn("text-xs font-semibold", colors[method] ?? "text-zinc-600")}>
      {method}
    </span>
  );
}

export default function TrafficPage() {
  const { activeProject } = useProject();
  const projectId = activeProject?.id || "";

  const [filters, setFilters] = useState<TrafficFilters>({});
  const [methodFilter, setMethodFilter] = useState<string>("");
  const [showFilters, setShowFilters] = useState(false);

  const { data, isLoading, isError, refetch, fetchNextPage, hasNextPage, isFetchingNextPage } = useInfiniteQuery({
    queryKey: ["traffic", projectId, filters],
    queryFn: ({ pageParam }) => trafficApi.list(projectId, filters, pageParam),
    initialPageParam: undefined as string | undefined,
    getNextPageParam: (lastPage) =>
      lastPage.pagination.has_more ? lastPage.pagination.next_cursor : undefined,
    enabled: !!projectId,
    staleTime: 15_000,
  });

  const logs = data?.pages.flatMap((page) => page.data) ?? [];

  if (!projectId) {
    return (
      <div className="flex flex-col items-center justify-center py-20">
        <p className="text-sm text-zinc-500">
          Select a project from the sidebar to view traffic
        </p>
      </div>
    );
  }

  if (isError) {
    return (
      <ErrorState
        title="Failed to load traffic"
        message="Could not fetch traffic data. Check your connection and try again."
        onRetry={() => refetch()}
      />
    );
  }

  return (
    <div className="space-y-4">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-xl font-semibold text-zinc-900">Traffic</h1>
          <p className="mt-0.5 text-sm text-zinc-500">
            Captured API traffic for this project
          </p>
        </div>
        <div className="flex items-center gap-2">
          <button
            onClick={() => setShowFilters(!showFilters)}
            className={cn(
              "flex items-center gap-1.5 rounded-lg border px-3 py-2 text-xs font-medium transition-colors",
              showFilters
                ? "border-zinc-300 bg-zinc-100 text-zinc-900"
                : "border-zinc-200 bg-white text-zinc-600 hover:bg-zinc-50"
            )}
          >
            <Filter size={14} />
            Filters
          </button>
          <button
            onClick={() => refetch()}
            className="flex items-center gap-1.5 rounded-lg border border-zinc-200 bg-white px-3 py-2 text-xs font-medium text-zinc-600 transition-colors hover:bg-zinc-50"
          >
            <RefreshCw size={14} />
            Refresh
          </button>
        </div>
      </div>

      {showFilters && (
        <div className="flex flex-wrap items-center gap-2 rounded-xl border border-zinc-100 bg-white p-4">
          <span className="text-xs font-medium text-zinc-500">Method:</span>
          {HTTP_METHODS.map((m) => (
            <button
              key={m}
              onClick={() => {
                const newMethod = methodFilter === m ? "" : m;
                setMethodFilter(newMethod);
                setFilters({
                  ...filters,
                  methods: newMethod ? [newMethod] : undefined,
                });
              }}
              className={cn(
                "rounded-md px-2 py-1 text-xs font-medium transition-colors",
                methodFilter === m
                  ? "bg-zinc-900 text-white"
                  : "bg-zinc-100 text-zinc-600 hover:bg-zinc-200"
              )}
            >
              {m}
            </button>
          ))}
        </div>
      )}

      <div className="rounded-xl border border-zinc-100 bg-white shadow-sm">
        {isLoading ? (
          <div className="flex items-center justify-center py-16">
            <div className="h-5 w-5 animate-spin rounded-full border-2 border-zinc-300 border-t-zinc-600" />
          </div>
        ) : logs.length === 0 ? (
          <div className="py-16 text-center text-sm text-zinc-500">
            No traffic captured yet
          </div>
        ) : (
          <div className="overflow-x-auto">
            <table className="w-full">
              <thead>
                <tr className="border-b border-zinc-100 text-left text-xs font-medium text-zinc-500">
                  <th className="px-4 py-3">Method</th>
                  <th className="px-4 py-3">Path</th>
                  <th className="px-4 py-3">Status</th>
                  <th className="px-4 py-3">Latency</th>
                  <th className="px-4 py-3">Time</th>
                  <th className="px-4 py-3">PII</th>
                </tr>
              </thead>
              <tbody className="divide-y divide-zinc-50">
                {logs.map((log) => (
                  <tr key={log.id} className="transition-colors hover:bg-zinc-50">
                    <td className="px-4 py-2.5">
                      <MethodBadge method={log.method} />
                    </td>
                    <td className="max-w-xs truncate px-4 py-2.5 text-sm font-mono text-zinc-700">
                      <Link
                        href={`/traffic/${log.id}`}
                        className="hover:text-zinc-900 hover:underline"
                      >
                        {log.path}
                      </Link>
                    </td>
                    <td className="px-4 py-2.5">
                      <StatusBadge code={log.status_code} />
                    </td>
                    <td className="px-4 py-2.5 text-xs text-zinc-500">
                      {log.latency_ms}ms
                    </td>
                    <td className="px-4 py-2.5 text-xs text-zinc-500">
                      {new Date(log.timestamp).toLocaleTimeString()}
                    </td>
                    <td className="px-4 py-2.5">
                      {log.pii_redacted && (
                        <span className="rounded-full bg-amber-50 px-2 py-0.5 text-[10px] font-medium text-amber-700">
                          Redacted
                        </span>
                      )}
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        )}

        {hasNextPage && (
          <div className="border-t border-zinc-100 px-4 py-3 text-center">
            <button
              onClick={() => fetchNextPage()}
              disabled={isFetchingNextPage}
              className="text-xs font-medium text-zinc-600 hover:text-zinc-900 disabled:opacity-50"
            >
              {isFetchingNextPage ? "Loading..." : "Load more"}
            </button>
          </div>
        )}
      </div>
    </div>
  );
}
