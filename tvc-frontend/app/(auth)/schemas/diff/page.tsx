"use client";

import { Suspense, useState } from "react";
import { useQuery } from "@tanstack/react-query";
import { useSearchParams } from "next/navigation";
import {
  ArrowLeft,
  AlertTriangle,
  CheckCircle2,
  GitCompare,
  FileJson,
} from "lucide-react";
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { LoadingPage } from "@/components/ui/loading-spinner";
import Link from "next/link";
import { schemasApi } from "@/lib/api/schemas";

interface SchemaVersion {
  id: string;
  version: string;
  uploaded_at: string;
  schema: Record<string, unknown>;
}

interface SchemaDiff {
  breaking_changes: Array<{
    path: string;
    type: "field_removed" | "type_changed" | "required_added";
    description: string;
  }>;
  non_breaking_changes: Array<{
    path: string;
    type: "field_added" | "optional_added";
    description: string;
  }>;
}

function DiffSection({
  title,
  changes,
  severity,
}: {
  title: string;
  changes: Array<{ path: string; type: string; description: string }>;
  severity: "breaking" | "non-breaking";
}) {
  if (changes.length === 0) return null;

  const icon =
    severity === "breaking" ? (
      <AlertTriangle size={18} className="text-red-600" />
    ) : (
      <CheckCircle2 size={18} className="text-blue-600" />
    );

  const badgeVariant = severity === "breaking" ? "error" : "info";

  return (
    <Card>
      <CardHeader>
        <div className="flex items-center gap-2">
          {icon}
          <CardTitle className="text-lg">{title}</CardTitle>
          <Badge variant={badgeVariant}>{changes.length}</Badge>
        </div>
      </CardHeader>
      <CardContent>
        <div className="space-y-2">
          {changes.map((change, idx) => (
            <div
              key={idx}
              className="flex items-start gap-3 border border-zinc-200 rounded-lg p-3"
            >
              <div className="flex-shrink-0 mt-0.5">
                <Badge variant={badgeVariant} className="text-xs">
                  {change.type.replace("_", " ")}
                </Badge>
              </div>
              <div className="flex-1 min-w-0">
                <code className="text-sm font-mono text-zinc-900">
                  {change.path}
                </code>
                <p className="mt-1 text-sm text-zinc-600">
                  {change.description}
                </p>
              </div>
            </div>
          ))}
        </div>
      </CardContent>
    </Card>
  );
}

function SchemaViewer({
  schema,
  title,
}: {
  schema: Record<string, unknown>;
  title: string;
}) {
  return (
    <div className="space-y-2">
      <h4 className="text-sm font-medium text-zinc-700">{title}</h4>
      <pre className="text-xs bg-zinc-900 text-zinc-100 rounded-lg p-4 overflow-auto max-h-[600px]">
        {JSON.stringify(schema, null, 2)}
      </pre>
    </div>
  );
}

function SchemaDiffPageContent() {
  const searchParams = useSearchParams();
  const projectId = searchParams.get("project") || "";
  const [baseVersion, setBaseVersion] = useState<string>("");
  const [compareVersion, setCompareVersion] = useState<string>("");

  const { data: versions, isLoading: versionsLoading } = useQuery({
    queryKey: ["schema-versions", projectId],
    queryFn: async () => {
      const response = await schemasApi.list(projectId);
      return response.data;
    },
    enabled: !!projectId,
  });

  const { data: diff, isLoading: diffLoading } = useQuery({
    queryKey: ["schema-diff", baseVersion, compareVersion],
    queryFn: async () => {
      if (!projectId || !baseVersion || !compareVersion) return null;
      const response = await schemasApi.diff(
        projectId,
        baseVersion,
        compareVersion,
      );
      return response.data.diff_report as SchemaDiff;
    },
    enabled: !!baseVersion && !!compareVersion,
  });

  const baseSchema = versions?.find((v) => v.id === baseVersion);
  const compareSchema = versions?.find((v) => v.id === compareVersion);

  if (versionsLoading) {
    return <LoadingPage />;
  }

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <div className="flex items-center gap-3">
          <Link href={`/schemas?project=${projectId}`}>
            <Button variant="ghost" size="sm">
              <ArrowLeft size={16} />
            </Button>
          </Link>
          <div>
            <h1 className="text-2xl font-semibold text-zinc-900">
              Schema Comparison
            </h1>
            <p className="mt-1 text-sm text-zinc-500">
              Compare schema versions to detect breaking changes
            </p>
          </div>
        </div>
      </div>

      <Card>
        <CardHeader>
          <CardTitle>Select Versions</CardTitle>
          <CardDescription>
            Choose two schema versions to compare
          </CardDescription>
        </CardHeader>
        <CardContent>
          <div className="grid grid-cols-1 md:grid-cols-3 gap-4 items-center">
            <div className="space-y-2">
              <label className="text-sm font-medium text-zinc-700">
                Base Version
              </label>
              <Select value={baseVersion} onValueChange={setBaseVersion}>
                <SelectTrigger>
                  <SelectValue placeholder="Select base version" />
                </SelectTrigger>
                <SelectContent>
                  {versions?.map((version) => (
                    <SelectItem key={version.id} value={version.id}>
                      v{version.version} •{" "}
                      {new Date(version.uploaded_at).toLocaleDateString()}
                    </SelectItem>
                  ))}
                </SelectContent>
              </Select>
            </div>

            <div className="flex justify-center">
              <GitCompare size={24} className="text-zinc-400" />
            </div>

            <div className="space-y-2">
              <label className="text-sm font-medium text-zinc-700">
                Compare Version
              </label>
              <Select value={compareVersion} onValueChange={setCompareVersion}>
                <SelectTrigger>
                  <SelectValue placeholder="Select compare version" />
                </SelectTrigger>
                <SelectContent>
                  {versions
                    ?.filter((v) => v.id !== baseVersion)
                    .map((version) => (
                      <SelectItem key={version.id} value={version.id}>
                        v{version.version} •{" "}
                        {new Date(version.uploaded_at).toLocaleDateString()}
                      </SelectItem>
                    ))}
                </SelectContent>
              </Select>
            </div>
          </div>
        </CardContent>
      </Card>

      {diffLoading && (
        <div className="flex items-center justify-center py-12">
          <LoadingPage />
        </div>
      )}

      {diff && (
        <>
          <div className="grid gap-4 md:grid-cols-2">
            <Card>
              <CardHeader className="pb-3">
                <CardTitle className="text-sm font-medium text-red-600">
                  Breaking Changes
                </CardTitle>
              </CardHeader>
              <CardContent>
                <p className="text-3xl font-semibold text-red-600">
                  {diff.breaking_changes.length}
                </p>
                <p className="mt-1 text-xs text-zinc-500">
                  May break existing API consumers
                </p>
              </CardContent>
            </Card>

            <Card>
              <CardHeader className="pb-3">
                <CardTitle className="text-sm font-medium text-blue-600">
                  Non-Breaking Changes
                </CardTitle>
              </CardHeader>
              <CardContent>
                <p className="text-3xl font-semibold text-blue-600">
                  {diff.non_breaking_changes.length}
                </p>
                <p className="mt-1 text-xs text-zinc-500">
                  Safe, backward-compatible changes
                </p>
              </CardContent>
            </Card>
          </div>

          <div className="space-y-4">
            <DiffSection
              title="Breaking Changes"
              changes={diff.breaking_changes}
              severity="breaking"
            />

            <DiffSection
              title="Non-Breaking Changes"
              changes={diff.non_breaking_changes}
              severity="non-breaking"
            />
          </div>

          {diff.breaking_changes.length === 0 &&
            diff.non_breaking_changes.length === 0 && (
              <Card>
                <CardContent className="py-12">
                  <div className="flex flex-col items-center gap-3 text-center">
                    <CheckCircle2 size={48} className="text-emerald-500" />
                    <div>
                      <h3 className="text-lg font-semibold text-zinc-900">
                        No differences found
                      </h3>
                      <p className="mt-1 text-sm text-zinc-500">
                        These schema versions are identical.
                      </p>
                    </div>
                  </div>
                </CardContent>
              </Card>
            )}
        </>
      )}

      {baseSchema && compareSchema && (
        <Card>
          <CardHeader>
            <CardTitle>Full Schema Comparison</CardTitle>
            <CardDescription>
              Side-by-side view of both schema versions
            </CardDescription>
          </CardHeader>
          <CardContent>
            <div className="grid grid-cols-1 lg:grid-cols-2 gap-4">
              <SchemaViewer
                schema={baseSchema.schema}
                title={`Base (v${baseSchema.version})`}
              />
              <SchemaViewer
                schema={compareSchema.schema}
                title={`Compare (v${compareSchema.version})`}
              />
            </div>
          </CardContent>
        </Card>
      )}

      {!baseVersion || !compareVersion ? (
        <Card>
          <CardContent className="py-12">
            <div className="flex flex-col items-center gap-3 text-center">
              <FileJson size={48} className="text-zinc-300" />
              <div>
                <h3 className="text-lg font-semibold text-zinc-900">
                  Select versions to compare
                </h3>
                <p className="mt-1 text-sm text-zinc-500">
                  Choose two schema versions above to see their differences.
                </p>
              </div>
            </div>
          </CardContent>
        </Card>
      ) : null}
    </div>
  );
}

export default function SchemaDiffPage() {
  return (
    <Suspense fallback={<LoadingPage />}>
      <SchemaDiffPageContent />
    </Suspense>
  );
}
