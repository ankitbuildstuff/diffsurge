"use client";

import Link from "next/link";
import { useState } from "react";
import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { schemasApi } from "@/lib/api/schemas";
import { useProject } from "@/lib/providers/project-provider";
import { FileCode2, Plus, GitBranch, GitCommit, GitCompare } from "lucide-react";
import { toast } from "sonner";
import { ErrorState } from "@/components/ui/error-state";

export default function SchemasPage() {
  const { activeProject } = useProject();
  const projectId = activeProject?.id || "";
  const queryClient = useQueryClient();
  const [showUpload, setShowUpload] = useState(false);
  const [version, setVersion] = useState("");
  const [schemaType, setSchemaType] = useState("openapi");
  const [content, setContent] = useState("");

  const { data, isLoading, isError, refetch } = useQuery({
    queryKey: ["schemas", projectId],
    queryFn: () => schemasApi.list(projectId),
    enabled: !!projectId,
  });

  const uploadMutation = useMutation({
    mutationFn: () => {
      let parsed: unknown;
      try {
        parsed = JSON.parse(content);
      } catch {
        throw new Error("Invalid JSON content");
      }
      return schemasApi.upload(projectId, {
        version,
        schema_type: schemaType,
        schema_content: parsed,
      });
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["schemas", projectId] });
      setShowUpload(false);
      setVersion("");
      setContent("");
      toast.success("Schema uploaded successfully");
    },
    onError: (err) => toast.error(err instanceof Error ? err.message : "Upload failed"),
  });

  const versions = data?.data ?? [];

  if (!projectId) {
    return (
      <div className="flex flex-col items-center justify-center py-20">
        <p className="text-sm text-zinc-500">Select a project from the sidebar to manage schemas</p>
      </div>
    );
  }

  if (isError) {
    return (
      <ErrorState
        title="Failed to load schemas"
        message="Could not fetch schema data. Check your connection and try again."
        onRetry={() => refetch()}
      />
    );
  }

  return (
    <div className="space-y-4">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-xl font-semibold text-zinc-900">Schemas</h1>
          <p className="mt-0.5 text-sm text-zinc-500">
            API schema versions and diff history
          </p>
        </div>
        <div className="flex items-center gap-2">
          {versions.length >= 2 && (
            <Link
              href="/schemas/diff"
              className="flex items-center gap-1.5 rounded-full border border-zinc-200 bg-white px-4 py-2 text-sm font-medium text-zinc-700 shadow-sm transition-colors hover:bg-zinc-50"
            >
              <GitCompare size={14} />
              Compare Versions
            </Link>
          )}
          <button
            onClick={() => setShowUpload(!showUpload)}
            className="flex items-center gap-1.5 rounded-full bg-zinc-900 px-4 py-2 text-sm font-medium text-white shadow-sm transition-colors hover:bg-zinc-800"
          >
            <Plus size={14} />
            Upload Schema
          </button>
        </div>
      </div>

      {showUpload && (
        <div className="rounded-xl border border-zinc-100 bg-white p-5 shadow-sm">
          <h3 className="mb-4 text-sm font-semibold text-zinc-900">
            Upload New Schema Version
          </h3>
          <form
            onSubmit={(e) => {
              e.preventDefault();
              uploadMutation.mutate();
            }}
            className="space-y-3"
          >
            <div className="grid grid-cols-2 gap-3">
              <div>
                <label className="block text-xs font-medium text-zinc-600 mb-1">
                  Version
                </label>
                <input
                  value={version}
                  onChange={(e) => setVersion(e.target.value)}
                  placeholder="v1.2.0"
                  required
                  className="w-full rounded-lg border border-zinc-200 px-3 py-2 text-sm focus:border-zinc-400 focus:outline-none focus:ring-1 focus:ring-zinc-400"
                />
              </div>
              <div>
                <label className="block text-xs font-medium text-zinc-600 mb-1">
                  Type
                </label>
                <select
                  value={schemaType}
                  onChange={(e) => setSchemaType(e.target.value)}
                  className="w-full rounded-lg border border-zinc-200 px-3 py-2 text-sm focus:border-zinc-400 focus:outline-none focus:ring-1 focus:ring-zinc-400"
                >
                  <option value="openapi">OpenAPI</option>
                  <option value="graphql">GraphQL</option>
                </select>
              </div>
            </div>
            <div>
              <label className="block text-xs font-medium text-zinc-600 mb-1">
                Schema Content (JSON)
              </label>
              <textarea
                value={content}
                onChange={(e) => setContent(e.target.value)}
                rows={8}
                required
                placeholder='{"openapi": "3.0.0", ...}'
                className="w-full rounded-lg border border-zinc-200 px-3 py-2 font-mono text-xs focus:border-zinc-400 focus:outline-none focus:ring-1 focus:ring-zinc-400"
              />
            </div>
            <div className="flex justify-end gap-2">
              <button
                type="button"
                onClick={() => setShowUpload(false)}
                className="rounded-lg border border-zinc-200 px-4 py-2 text-sm font-medium text-zinc-600 hover:bg-zinc-50"
              >
                Cancel
              </button>
              <button
                type="submit"
                disabled={uploadMutation.isPending}
                className="rounded-lg bg-zinc-900 px-4 py-2 text-sm font-medium text-white hover:bg-zinc-800 disabled:opacity-50"
              >
                {uploadMutation.isPending ? "Uploading..." : "Upload"}
              </button>
            </div>
          </form>
        </div>
      )}

      <div className="rounded-xl border border-zinc-100 bg-white shadow-sm">
        {isLoading ? (
          <div className="flex items-center justify-center py-16">
            <div className="h-5 w-5 animate-spin rounded-full border-2 border-zinc-300 border-t-zinc-600" />
          </div>
        ) : versions.length === 0 ? (
          <div className="flex flex-col items-center py-16">
            <FileCode2 size={32} className="mb-3 text-zinc-300" />
            <p className="text-sm text-zinc-500">No schemas uploaded yet</p>
          </div>
        ) : (
          <div className="divide-y divide-zinc-50">
            {versions.map((v) => (
              <div
                key={v.id}
                className="flex items-center justify-between px-5 py-4"
              >
                <div className="flex items-center gap-3">
                  <div className="flex h-9 w-9 items-center justify-center rounded-lg bg-zinc-50">
                    <FileCode2 size={16} className="text-zinc-500" />
                  </div>
                  <div>
                    <p className="text-sm font-medium text-zinc-900">
                      {v.version}
                    </p>
                    <p className="text-xs text-zinc-500">
                      {v.schema_type} &middot;{" "}
                      {new Date(v.created_at).toLocaleDateString()}
                    </p>
                  </div>
                </div>
                <div className="flex items-center gap-3 text-xs text-zinc-400">
                  {v.git_branch && (
                    <span className="flex items-center gap-1">
                      <GitBranch size={12} />
                      {v.git_branch}
                    </span>
                  )}
                  {v.git_commit && (
                    <span className="flex items-center gap-1">
                      <GitCommit size={12} />
                      {v.git_commit.slice(0, 7)}
                    </span>
                  )}
                </div>
              </div>
            ))}
          </div>
        )}
      </div>
    </div>
  );
}
