"use client";

import { useState } from "react";
import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { projectsApi, type Project } from "@/lib/api/projects";
import { useOrganization } from "@/lib/providers/organization-provider";
import { toast } from "sonner";
import { useRouter } from "next/navigation";
import {
  Settings,
  Plus,
  FolderOpen,
  Trash2,
  Building2,
  Loader2,
} from "lucide-react";
import Link from "next/link";

export default function SettingsPage() {
  const { activeOrg, isLoading: orgLoading } = useOrganization();
  const [showCreate, setShowCreate] = useState(false);
  const [projectName, setProjectName] = useState("");
  const [description, setDescription] = useState("");
  const queryClient = useQueryClient();
  const router = useRouter();

  const { data: projectsResponse, isLoading: projectsLoading } = useQuery({
    queryKey: ["projects", activeOrg?.id],
    queryFn: () => projectsApi.list(activeOrg!.id),
    enabled: !!activeOrg,
  });

  const projects = projectsResponse?.data ?? [];

  const createMutation = useMutation({
    mutationFn: () =>
      projectsApi.create({
        name: projectName,
        description: description || undefined,
        organization_id: activeOrg!.id,
      }),
    onSuccess: (project) => {
      toast.success("Project created");
      queryClient.invalidateQueries({ queryKey: ["projects", activeOrg?.id] });
      setShowCreate(false);
      setProjectName("");
      setDescription("");
      router.push(`/dashboard?project=${project.id}`);
    },
    onError: () => toast.error("Failed to create project"),
  });

  const deleteMutation = useMutation({
    mutationFn: (id: string) => projectsApi.delete(id),
    onSuccess: () => {
      toast.success("Project deleted");
      queryClient.invalidateQueries({ queryKey: ["projects", activeOrg?.id] });
    },
    onError: () => toast.error("Failed to delete project"),
  });

  if (orgLoading) {
    return (
      <div className="flex items-center justify-center py-20">
        <Loader2 size={24} className="animate-spin text-zinc-400" />
      </div>
    );
  }

  return (
    <div className="space-y-6">
      <div>
        <h1 className="text-xl font-semibold text-zinc-900">Settings</h1>
        <p className="mt-0.5 text-sm text-zinc-500">
          Manage your projects and organization
        </p>
      </div>

      {activeOrg && (
        <div className="rounded-xl border border-zinc-100 bg-white px-5 py-4 shadow-sm">
          <div className="flex items-center gap-3">
            <div className="flex h-9 w-9 items-center justify-center rounded-lg bg-zinc-100">
              <Building2 size={18} className="text-zinc-500" />
            </div>
            <div>
              <p className="text-sm font-semibold text-zinc-900">
                {activeOrg.name}
              </p>
              <p className="text-xs text-zinc-400">
                Org ID: {activeOrg.id}
              </p>
            </div>
          </div>
        </div>
      )}

      <div className="rounded-xl border border-zinc-100 bg-white shadow-sm">
        <div className="flex items-center justify-between border-b border-zinc-100 px-5 py-4">
          <div className="flex items-center gap-2">
            <FolderOpen size={16} className="text-zinc-500" />
            <h2 className="text-sm font-semibold text-zinc-900">Projects</h2>
          </div>
          <button
            onClick={() => setShowCreate(!showCreate)}
            className="flex items-center gap-1.5 rounded-lg border border-zinc-200 px-3 py-1.5 text-xs font-medium text-zinc-600 transition-colors hover:bg-zinc-50"
          >
            <Plus size={12} />
            New Project
          </button>
        </div>

        {showCreate && (
          <div className="border-b border-zinc-100 p-5">
            <form
              onSubmit={(e) => {
                e.preventDefault();
                createMutation.mutate();
              }}
              className="space-y-3"
            >
              <div>
                <label className="block text-xs font-medium text-zinc-600 mb-1">
                  Project Name
                </label>
                <input
                  value={projectName}
                  onChange={(e) => setProjectName(e.target.value)}
                  required
                  placeholder="My API"
                  className="w-full rounded-lg border border-zinc-200 px-3 py-2 text-sm focus:border-zinc-400 focus:outline-none focus:ring-1 focus:ring-zinc-400"
                />
              </div>
              <div>
                <label className="block text-xs font-medium text-zinc-600 mb-1">
                  Description
                </label>
                <input
                  value={description}
                  onChange={(e) => setDescription(e.target.value)}
                  placeholder="Optional description"
                  className="w-full rounded-lg border border-zinc-200 px-3 py-2 text-sm focus:border-zinc-400 focus:outline-none focus:ring-1 focus:ring-zinc-400"
                />
              </div>
              <div className="flex gap-2">
                <button
                  type="submit"
                  disabled={createMutation.isPending || !projectName}
                  className="rounded-lg bg-zinc-900 px-4 py-2 text-sm font-medium text-white hover:bg-zinc-800 disabled:opacity-50"
                >
                  {createMutation.isPending ? "Creating..." : "Create Project"}
                </button>
                <button
                  type="button"
                  onClick={() => setShowCreate(false)}
                  className="rounded-lg border border-zinc-200 px-4 py-2 text-sm font-medium text-zinc-600 hover:bg-zinc-50"
                >
                  Cancel
                </button>
              </div>
            </form>
          </div>
        )}

        <div className="p-5">
          {projectsLoading ? (
            <div className="flex justify-center py-8">
              <Loader2 size={20} className="animate-spin text-zinc-400" />
            </div>
          ) : projects.length === 0 ? (
            <div className="flex flex-col items-center py-8">
              <Settings size={28} className="mb-3 text-zinc-300" />
              <p className="text-sm text-zinc-500">
                Create a project to get started
              </p>
            </div>
          ) : (
            <div className="space-y-2">
              {projects.map((project: Project) => (
                <div
                  key={project.id}
                  className="flex items-center justify-between rounded-lg border border-zinc-100 px-4 py-3 transition-colors hover:bg-zinc-50"
                >
                  <Link
                    href={`/dashboard?project=${project.id}`}
                    className="flex-1"
                  >
                    <p className="text-sm font-medium text-zinc-900">
                      {project.name}
                    </p>
                    {project.description && (
                      <p className="mt-0.5 text-xs text-zinc-400">
                        {project.description}
                      </p>
                    )}
                  </Link>
                  <button
                    onClick={() => deleteMutation.mutate(project.id)}
                    disabled={deleteMutation.isPending}
                    className="rounded-md p-1.5 text-zinc-400 transition-colors hover:bg-red-50 hover:text-red-500"
                  >
                    <Trash2 size={14} />
                  </button>
                </div>
              ))}
            </div>
          )}
        </div>
      </div>

      <div className="rounded-xl border border-zinc-100 bg-white shadow-sm">
        <div className="border-b border-zinc-100 px-5 py-4">
          <h2 className="text-sm font-semibold text-zinc-900">
            Organization Settings
          </h2>
        </div>
        <div className="px-5 py-4 space-y-3">
          <Link
            href={`/settings/team?org=${activeOrg?.id || ""}`}
            className="flex items-center justify-between rounded-lg border border-zinc-100 px-4 py-3 text-sm text-zinc-600 transition-colors hover:bg-zinc-50"
          >
            Team Members
            <span className="text-zinc-400">&rarr;</span>
          </Link>
          <Link
            href={`/settings/api-keys?org=${activeOrg?.id || ""}`}
            className="flex items-center justify-between rounded-lg border border-zinc-100 px-4 py-3 text-sm text-zinc-600 transition-colors hover:bg-zinc-50"
          >
            API Keys
            <span className="text-zinc-400">&rarr;</span>
          </Link>
        </div>
      </div>
    </div>
  );
}
