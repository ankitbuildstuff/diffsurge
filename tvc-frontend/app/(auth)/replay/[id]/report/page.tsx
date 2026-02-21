"use client";

import { Suspense } from "react";
import { useQuery } from "@tanstack/react-query";
import { useParams, useSearchParams } from "next/navigation";
import {
  ArrowLeft,
  CheckCircle2,
  XCircle,
  AlertTriangle,
  FileJson,
} from "lucide-react";
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";
import { Badge, SeverityBadge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { LoadingPage } from "@/components/ui/loading-spinner";
import Link from "next/link";
import { replaysApi, type ReplaySession } from "@/lib/api/replays";

interface ReplayResultDiff {
  path: string;
  type: "added" | "removed" | "changed";
  old_value?: unknown;
  new_value?: unknown;
}

interface ReplayResultDetail {
  id: string;
  request_path: string;
  request_method: string;
  original_status: number;
  replay_status: number;
  severity: "critical" | "major" | "minor" | "none";
  diff_summary: {
    status_changed: boolean;
    body_changed: boolean;
    headers_changed: boolean;
  };
  original_body: Record<string, unknown>;
  replay_body: Record<string, unknown>;
  diff: ReplayResultDiff[];
}

function DiffViewer({ diff }: { diff: ReplayResultDiff[] }) {
  if (diff.length === 0) {
    return (
      <div className="flex items-center gap-2 text-sm text-emerald-600">
        <CheckCircle2 size={16} />
        <span>No differences detected</span>
      </div>
    );
  }

  return (
    <div className="space-y-1">
      {diff.map((change, idx) => (
        <div
          key={idx}
          className="flex items-start gap-2 text-sm font-mono p-2 rounded border"
        >
          {change.type === "added" && (
            <span className="flex-1 text-emerald-600">
              + {change.path}: {JSON.stringify(change.new_value)}
            </span>
          )}
          {change.type === "removed" && (
            <span className="flex-1 text-red-600">
              - {change.path}: {JSON.stringify(change.old_value)}
            </span>
          )}
          {change.type === "changed" && (
            <span className="flex-1">
              <span className="text-red-600">
                - {change.path}: {JSON.stringify(change.old_value)}
              </span>
              <br />
              <span className="text-emerald-600">
                + {change.path}: {JSON.stringify(change.new_value)}
              </span>
            </span>
          )}
        </div>
      ))}
    </div>
  );
}

function JsonComparison({
  original,
  replay,
  title,
}: {
  original: Record<string, unknown>;
  replay: Record<string, unknown>;
  title: string;
}) {
  return (
    <div className="space-y-2">
      <h4 className="text-sm font-medium text-zinc-900">{title}</h4>
      <div className="grid grid-cols-2 gap-4">
        <div>
          <Label className="text-xs text-zinc-500 mb-1">Original</Label>
          <pre className="text-xs bg-zinc-50 border border-zinc-200 rounded p-3 overflow-auto max-h-[400px]">
            {JSON.stringify(original, null, 2)}
          </pre>
        </div>
        <div>
          <Label className="text-xs text-zinc-500 mb-1">Replay</Label>
          <pre className="text-xs bg-zinc-50 border border-zinc-200 rounded p-3 overflow-auto max-h-[400px]">
            {JSON.stringify(replay, null, 2)}
          </pre>
        </div>
      </div>
    </div>
  );
}

function ReplayReportPageContent() {
  const params = useParams();
  const searchParams = useSearchParams();
  const replayId = params.id as string;
  const projectId = searchParams.get("project") || "";

  const { data: session, isLoading: sessionLoading } = useQuery({
    queryKey: ["replay-session", replayId],
    queryFn: async () => {
      if (!projectId || !replayId) return null;
      const response = await replaysApi.get(projectId, replayId);
      return response.data;
    },
    enabled: !!replayId && !!projectId,
  });

  const { data: results, isLoading: resultsLoading } = useQuery({
    queryKey: ["replay-results", replayId],
    queryFn: async () => {
      if (!projectId || !replayId) return [] as ReplayResultDetail[];
      const response = await replaysApi.results(projectId, replayId);
      return response.data as unknown as ReplayResultDetail[];
    },
    enabled: !!replayId && !!projectId,
  });

  if (sessionLoading || resultsLoading) {
    return <LoadingPage />;
  }

  if (!session || !results) {
    return (
      <div className="text-center py-12">
        <p className="text-zinc-500">Replay session not found</p>
      </div>
    );
  }

  const criticalCount = results.filter((r) => r.severity === "critical").length;
  const majorCount = results.filter((r) => r.severity === "major").length;
  const minorCount = results.filter((r) => r.severity === "minor").length;

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <div className="flex items-center gap-3">
          <Link href={`/replay?project=${projectId}`}>
            <Button variant="ghost" size="sm">
              <ArrowLeft size={16} />
            </Button>
          </Link>
          <div>
            <h1 className="text-2xl font-semibold text-zinc-900">
              Replay Report
            </h1>
            <p className="mt-1 text-sm text-zinc-500">{session.name}</p>
          </div>
        </div>
        <Badge
          variant={
            session.status === "completed"
              ? "success"
              : session.status === "failed"
                ? "error"
                : "warning"
          }
        >
          {session.status}
        </Badge>
      </div>

      <div className="grid gap-4 md:grid-cols-4">
        <Card>
          <CardHeader className="pb-3">
            <CardTitle className="text-sm font-medium text-zinc-500">
              Total Requests
            </CardTitle>
          </CardHeader>
          <CardContent>
            <p className="text-2xl font-semibold">{session.total_requests}</p>
          </CardContent>
        </Card>
        <Card>
          <CardHeader className="pb-3">
            <CardTitle className="text-sm font-medium text-emerald-600">
              Passed
            </CardTitle>
          </CardHeader>
          <CardContent>
            <p className="text-2xl font-semibold text-emerald-600">
              {session.successful_requests}
            </p>
          </CardContent>
        </Card>
        <Card>
          <CardHeader className="pb-3">
            <CardTitle className="text-sm font-medium text-red-600">
              Failed
            </CardTitle>
          </CardHeader>
          <CardContent>
            <p className="text-2xl font-semibold text-red-600">
              {session.failed_requests}
            </p>
          </CardContent>
        </Card>
        <Card>
          <CardHeader className="pb-3">
            <CardTitle className="text-sm font-medium text-zinc-500">
              Duration
            </CardTitle>
          </CardHeader>
          <CardContent>
            <p className="text-2xl font-semibold">
              {session.started_at && session.completed_at
                ? `${Math.round(
                    (new Date(session.completed_at).getTime() -
                      new Date(session.started_at).getTime()) /
                      1000,
                  )}s`
                : "—"}
            </p>
          </CardContent>
        </Card>
      </div>

      <Card>
        <CardHeader>
          <CardTitle>Severity Breakdown</CardTitle>
        </CardHeader>
        <CardContent>
          <div className="flex items-center gap-6">
            {criticalCount > 0 && (
              <div className="flex items-center gap-2">
                <XCircle size={18} className="text-red-600" />
                <span className="text-sm">
                  <span className="font-semibold">{criticalCount}</span>{" "}
                  Critical
                </span>
              </div>
            )}
            {majorCount > 0 && (
              <div className="flex items-center gap-2">
                <AlertTriangle size={18} className="text-amber-600" />
                <span className="text-sm">
                  <span className="font-semibold">{majorCount}</span> Major
                </span>
              </div>
            )}
            {minorCount > 0 && (
              <div className="flex items-center gap-2">
                <CheckCircle2 size={18} className="text-blue-600" />
                <span className="text-sm">
                  <span className="font-semibold">{minorCount}</span> Minor
                </span>
              </div>
            )}
            {criticalCount === 0 && majorCount === 0 && minorCount === 0 && (
              <div className="flex items-center gap-2 text-emerald-600">
                <CheckCircle2 size={18} />
                <span className="text-sm font-semibold">All checks passed</span>
              </div>
            )}
          </div>
        </CardContent>
      </Card>

      <Card>
        <CardHeader>
          <CardTitle>Detailed Results</CardTitle>
          <CardDescription>
            Side-by-side comparison of original and replayed responses
          </CardDescription>
        </CardHeader>
        <CardContent>
          <div className="space-y-4">
            {results.map((result) => (
              <Card key={result.id} className="border-zinc-200">
                <CardHeader>
                  <div className="flex items-center justify-between">
                    <div className="flex items-center gap-3">
                      <Badge variant="info">{result.request_method}</Badge>
                      <code className="text-sm font-mono">
                        {result.request_path}
                      </code>
                    </div>
                    <SeverityBadge severity={result.severity} />
                  </div>
                </CardHeader>
                <CardContent>
                  <Tabs defaultValue="diff">
                    <TabsList>
                      <TabsTrigger value="diff">Differences</TabsTrigger>
                      <TabsTrigger value="bodies">Response Bodies</TabsTrigger>
                    </TabsList>
                    <TabsContent value="diff" className="space-y-3 mt-4">
                      <div className="flex items-center gap-4 text-sm">
                        <div>
                          Status:{" "}
                          <code className="font-mono">
                            {result.original_status}
                          </code>{" "}
                          →{" "}
                          <code className="font-mono">
                            {result.replay_status}
                          </code>
                        </div>
                        {result.diff_summary.status_changed && (
                          <Badge variant="warning">Status changed</Badge>
                        )}
                      </div>
                      <DiffViewer diff={result.diff} />
                    </TabsContent>
                    <TabsContent value="bodies" className="mt-4">
                      <JsonComparison
                        original={result.original_body}
                        replay={result.replay_body}
                        title="Response Bodies"
                      />
                    </TabsContent>
                  </Tabs>
                </CardContent>
              </Card>
            ))}
          </div>
        </CardContent>
      </Card>
    </div>
  );
}

function Label({
  children,
  className,
}: {
  children: React.ReactNode;
  className?: string;
}) {
  return <div className={className}>{children}</div>;
}

export default function ReplayReportPage() {
  return (
    <Suspense fallback={<LoadingPage />}>
      <ReplayReportPageContent />
    </Suspense>
  );
}
