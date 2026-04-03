"use client";

import { useQuery } from "@tanstack/react-query";
import { trafficApi } from "@/lib/api/traffic";
import { replaysApi } from "@/lib/api/replays";
import { useProject } from "@/lib/providers/project-provider";
import {
  Radio,
  RefreshCw,
  AlertTriangle,
  Clock,
  ArrowUpRight,
  ArrowDownRight,
  FolderOpen,
} from "lucide-react";
import Link from "next/link";
import { ErrorState } from "@/components/ui/error-state";

function StatCard({
  label,
  value,
  icon: Icon,
  trend,
  href,
}: {
  label: string;
  value: string | number;
  icon: React.ElementType;
  trend?: { value: string; positive: boolean };
  href?: string;
}) {
  const content = (
    <div className="rounded-xl border border-zinc-100 bg-white p-5 shadow-sm transition-shadow hover:shadow-md">
      <div className="flex items-center justify-between">
        <div className="flex h-9 w-9 items-center justify-center rounded-lg bg-zinc-50">
          <Icon size={18} className="text-zinc-500" />
        </div>
        {trend && (
          <span
            className={`flex items-center gap-0.5 text-xs font-medium ${
              trend.positive ? "text-emerald-600" : "text-red-500"
            }`}
          >
            {trend.positive ? (
              <ArrowUpRight size={12} />
            ) : (
              <ArrowDownRight size={12} />
            )}
            {trend.value}
          </span>
        )}
      </div>
      <p className="mt-3 text-2xl font-semibold text-zinc-900">{value}</p>
      <p className="mt-0.5 text-[13px] text-zinc-500">{label}</p>
    </div>
  );

  if (href) {
    return <Link href={href}>{content}</Link>;
  }
  return content;
}

function EmptyDashboard() {
  return (
    <div className="flex flex-col items-center justify-center py-20">
      <div className="mb-4 flex h-16 w-16 items-center justify-center rounded-2xl bg-zinc-100">
        <FolderOpen size={28} className="text-zinc-400" />
      </div>
      <h2 className="text-lg font-semibold text-zinc-900">
        No project selected
      </h2>
      <p className="mt-1 max-w-sm text-center text-sm text-zinc-500">
        Create a project in Settings to start capturing traffic and
        detecting API drift.
      </p>
      <Link
        href="/settings"
        className="mt-6 inline-flex items-center gap-1.5 rounded-full bg-zinc-900 px-5 py-2.5 text-sm font-medium text-white shadow-sm transition-colors hover:bg-zinc-800"
      >
        Create a project
      </Link>
    </div>
  );
}

export default function DashboardPage() {
  const { activeProject } = useProject();
  const projectId = activeProject?.id || "";

  const { data: stats, isError: statsError, refetch: refetchStats } = useQuery({
    queryKey: ["traffic-stats", projectId],
    queryFn: () => trafficApi.stats(projectId, "all"),
    enabled: !!projectId,
    staleTime: 30_000,
    refetchInterval: 60_000,
  });

  const { data: replays, isError: replaysError, refetch: refetchReplays } = useQuery({
    queryKey: ["replays", projectId],
    queryFn: () => replaysApi.list(projectId),
    enabled: !!projectId,
    staleTime: 30_000,
  });

  if (!projectId) {
    return <EmptyDashboard />;
  }

  if (statsError && replaysError) {
    return (
      <ErrorState
        title="Failed to load dashboard"
        message="Could not connect to the API. Check your connection and try again."
        onRetry={() => {
          refetchStats();
          refetchReplays();
        }}
      />
    );
  }

  const trafficStats = stats?.data;
  const replayList = replays?.data ?? [];
  const replayCount = replayList.length;

  return (
    <div className="space-y-6">
      <div>
        <h1 className="text-xl font-semibold text-zinc-900">Dashboard</h1>
        <p className="mt-0.5 text-sm text-zinc-500">
          Overview of your API monitoring
        </p>
      </div>

      <div className="grid grid-cols-1 gap-4 sm:grid-cols-2 lg:grid-cols-4">
        <StatCard
          label="Total requests"
          value={trafficStats?.total_requests ?? 0}
          icon={Radio}
          href="/traffic"
        />
        <StatCard
          label="Error rate"
          value={
            trafficStats
              ? `${(trafficStats.error_rate * 100).toFixed(1)}%`
              : "0%"
          }
          icon={AlertTriangle}
          href="/traffic"
        />
        <StatCard
          label="Avg latency"
          value={
            trafficStats
              ? `${trafficStats.avg_latency_ms.toFixed(0)}ms`
              : "—"
          }
          icon={Clock}
        />
        <StatCard
          label="Replay sessions"
          value={replayCount}
          icon={RefreshCw}
          href="/replay"
        />
      </div>

      {replayList.length > 0 && (
        <div className="rounded-xl border border-zinc-100 bg-white shadow-sm">
          <div className="flex items-center justify-between border-b border-zinc-100 px-5 py-3.5">
            <h2 className="text-sm font-semibold text-zinc-900">
              Recent Replays
            </h2>
            <Link
              href="/replay"
              className="text-xs text-zinc-500 hover:text-zinc-700"
            >
              View all
            </Link>
          </div>
          <div className="divide-y divide-zinc-50">
            {replayList.slice(0, 5).map((replay) => (
              <Link
                key={replay.id}
                href={`/replay/${replay.id}`}
                className="flex items-center justify-between px-5 py-3 transition-colors hover:bg-zinc-50"
              >
                <div>
                  <p className="text-sm font-medium text-zinc-900">
                    {replay.name}
                  </p>
                  <p className="text-xs text-zinc-500">
                    {new Date(replay.created_at).toLocaleDateString()}
                  </p>
                </div>
                <span
                  className={`rounded-full px-2.5 py-0.5 text-xs font-medium ${
                    replay.status === "completed"
                      ? "bg-emerald-50 text-emerald-700"
                      : replay.status === "running"
                        ? "bg-blue-50 text-blue-700"
                        : replay.status === "failed"
                          ? "bg-red-50 text-red-700"
                          : "bg-zinc-50 text-zinc-600"
                  }`}
                >
                  {replay.status}
                </span>
              </Link>
            ))}
          </div>
        </div>
      )}
    </div>
  );
}
