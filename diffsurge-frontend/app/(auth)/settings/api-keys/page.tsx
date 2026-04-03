"use client";

import { Suspense } from "react";
import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { useSearchParams } from "next/navigation";
import { Key, Plus, Trash2 } from "lucide-react";
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
} from "@/components/ui/dialog";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { EmptyState } from "@/components/ui/empty-state";
import { LoadingPage } from "@/components/ui/loading-spinner";
import { CopyButton } from "@/components/ui/copy-button";
import { useState } from "react";
import Link from "next/link";
import { apiKeysApi } from "@/lib/api/api-keys";
import { projectsApi, type Project } from "@/lib/api/projects";
import { toast } from "sonner";
import { useOrganization } from "@/lib/providers/organization-provider";

function ApiKeysPageContent() {
  const searchParams = useSearchParams();
  const { activeOrg } = useOrganization();
  const orgId = searchParams.get("org") || activeOrg?.id || "";
  const queryClient = useQueryClient();
  const [createOpen, setCreateOpen] = useState(false);
  const [keyName, setKeyName] = useState("");
  const [selectedProjectId, setSelectedProjectId] = useState<string>("");
  const [newKey, setNewKey] = useState<string | null>(null);

  const { data: keys, isLoading } = useQuery({
    queryKey: ["api-keys", orgId],
    queryFn: async () => {
      if (!orgId) return [];
      return await apiKeysApi.list(orgId);
    },
    enabled: !!orgId,
    staleTime: 30_000,
  });

  const { data: projectsData } = useQuery({
    queryKey: ["projects", orgId],
    queryFn: async () => {
      if (!orgId) return { data: [] };
      return await projectsApi.list(orgId);
    },
    enabled: !!orgId,
    staleTime: 30_000,
  });

  const projects: Project[] = projectsData?.data ?? [];

  const createMutation = useMutation({
    mutationFn: async ({ name, projectId }: { name: string; projectId: string }) => {
      if (!orgId) throw new Error("Organization ID required");
      return await apiKeysApi.create(orgId, { name, project_id: projectId });
    },
    onSuccess: (data) => {
      queryClient.invalidateQueries({ queryKey: ["api-keys", orgId] });
      setNewKey(data.key);
      setKeyName("");
      setSelectedProjectId("");
      toast.success("API key created successfully");
    },
    onError: (error: Error) => {
      toast.error(error.message || "Failed to create API key");
    },
  });

  const revokeMutation = useMutation({
    mutationFn: async (keyId: string) => {
      if (!orgId) throw new Error("Organization ID required");
      await apiKeysApi.delete(orgId, keyId);
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["api-keys", orgId] });
      toast.success("API key revoked");
    },
    onError: (error: Error) => {
      toast.error(error.message || "Failed to revoke API key");
    },
  });

  const handleCreate = (e: React.FormEvent) => {
    e.preventDefault();
    if (!selectedProjectId) {
      toast.error("Please select a project for this API key");
      return;
    }
    createMutation.mutate({ name: keyName, projectId: selectedProjectId });
  };

  const maskKey = (prefix: string) => `${prefix}${"•".repeat(32)}`;

  // Helper to get project name from ID
  const getProjectName = (projectId?: string) => {
    if (!projectId) return "—";
    const project = projects.find((p) => p.id === projectId);
    return project?.name ?? projectId.slice(0, 8) + "…";
  };

  if (isLoading) {
    return <LoadingPage />;
  }

  if (!orgId) {
    return (
      <EmptyState
        icon={<Key size={28} className="text-zinc-400" />}
        title="No organization selected"
        description="Select an organization to manage API keys."
        action={
          <Link href="/settings">
            <Button>Go to Settings</Button>
          </Link>
        }
      />
    );
  }

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-semibold text-zinc-900">API Keys</h1>
          <p className="mt-1 text-sm text-zinc-500">
            Manage API keys for programmatic access. Each key is linked to a project.
          </p>
        </div>

        <Dialog
          open={createOpen}
          onOpenChange={(open) => {
            setCreateOpen(open);
            if (!open) {
              setNewKey(null);
              setSelectedProjectId("");
              setKeyName("");
            }
          }}
        >
          <DialogTrigger asChild>
            <Button>
              <Plus size={16} />
              Create API Key
            </Button>
          </DialogTrigger>
          <DialogContent>
            {newKey ? (
              <>
                <DialogHeader>
                  <DialogTitle>API Key Created</DialogTitle>
                  <DialogDescription>
                    Save this key securely. You won&apos;t be able to see it
                    again.
                  </DialogDescription>
                </DialogHeader>
                <div className="space-y-4 py-4">
                  <div className="flex items-center gap-2 rounded-lg border border-zinc-200 bg-zinc-50 p-3 min-w-0">
                    <code className="flex-1 min-w-0 break-all text-sm font-mono">{newKey}</code>
                    <CopyButton value={newKey} />
                  </div>
                  <div className="rounded-lg border border-amber-200 bg-amber-50 p-3">
                    <p className="text-xs text-amber-800">
                      <strong>Usage:</strong> Pass this key via the{" "}
                      <code className="rounded bg-amber-100 px-1 py-0.5 text-amber-900">X-API-Key</code>
                      {" "}header when sending traffic through the proxy.
                    </p>
                  </div>
                </div>
                <DialogFooter>
                  <Button onClick={() => setCreateOpen(false)}>Done</Button>
                </DialogFooter>
              </>
            ) : (
              <form onSubmit={handleCreate}>
                <DialogHeader>
                  <DialogTitle>Create API key</DialogTitle>
                  <DialogDescription>
                    Generate a new API key linked to a project. Traffic sent with
                    this key will be associated with the selected project.
                  </DialogDescription>
                </DialogHeader>
                <div className="space-y-4 py-4">
                  <div className="space-y-2">
                    <Label htmlFor="name" required>
                      Key name
                    </Label>
                    <Input
                      id="name"
                      placeholder="production-server"
                      value={keyName}
                      onChange={(e) => setKeyName(e.target.value)}
                      required
                    />
                    <p className="text-xs text-zinc-500">
                      A descriptive name to identify this key
                    </p>
                  </div>
                  <div className="space-y-2">
                    <Label required>Project</Label>
                    {projects.length === 0 ? (
                      <div className="rounded-lg border border-zinc-200 bg-zinc-50 p-3">
                        <p className="text-sm text-zinc-600">
                          No projects found.{" "}
                          <Link href="/settings" className="text-teal-600 hover:underline">
                            Create a project
                          </Link>{" "}
                          first.
                        </p>
                      </div>
                    ) : (
                      <Select
                        value={selectedProjectId}
                        onValueChange={setSelectedProjectId}
                      >
                        <SelectTrigger>
                          <SelectValue placeholder="Select a project" />
                        </SelectTrigger>
                        <SelectContent>
                          {projects.map((project) => (
                            <SelectItem key={project.id} value={project.id}>
                              {project.name}
                            </SelectItem>
                          ))}
                        </SelectContent>
                      </Select>
                    )}
                    <p className="text-xs text-zinc-500">
                      Traffic captured with this key will appear under this project
                    </p>
                  </div>
                </div>
                <DialogFooter>
                  <Button
                    type="button"
                    variant="secondary"
                    onClick={() => setCreateOpen(false)}
                  >
                    Cancel
                  </Button>
                  <Button
                    type="submit"
                    disabled={createMutation.isPending || !keyName || !selectedProjectId}
                  >
                    {createMutation.isPending ? "Creating..." : "Create Key"}
                  </Button>
                </DialogFooter>
              </form>
            )}
          </DialogContent>
        </Dialog>
      </div>

      <Card>
        <CardHeader>
          <CardTitle>Active Keys</CardTitle>
          <CardDescription>
            API keys allow programmatic access to the Diffsurge API. Each key is linked to a project.
          </CardDescription>
        </CardHeader>
        <CardContent>
          {!keys || keys.length === 0 ? (
            <EmptyState
              icon={<Key size={28} className="text-zinc-400" />}
              title="No API keys yet"
              description="Create an API key to access this project programmatically."
              action={
                <Button onClick={() => setCreateOpen(true)}>
                  <Plus size={16} />
                  Create First Key
                </Button>
              }
            />
          ) : (
            <Table>
              <TableHeader>
                <TableRow>
                  <TableHead>Name</TableHead>
                  <TableHead>Project</TableHead>
                  <TableHead>Key</TableHead>
                  <TableHead>Created</TableHead>
                  <TableHead>Last Used</TableHead>
                  <TableHead className="text-right">Actions</TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {keys.map((key) => (
                  <TableRow key={key.id}>
                    <TableCell className="font-medium">{key.name}</TableCell>
                    <TableCell>
                      <span className="inline-flex items-center rounded-full bg-zinc-100 px-2 py-0.5 text-xs font-medium text-zinc-700">
                        {getProjectName(key.project_id)}
                      </span>
                    </TableCell>
                    <TableCell>
                        <code className="text-sm font-mono text-zinc-600">
                          {maskKey(key.key_prefix)}
                        </code>
                    </TableCell>
                    <TableCell className="text-zinc-500">
                      {new Date(key.created_at).toLocaleDateString()}
                    </TableCell>
                    <TableCell className="text-zinc-500">
                      {key.last_used_at
                        ? new Date(key.last_used_at).toLocaleDateString()
                        : "Never"}
                    </TableCell>
                    <TableCell className="text-right">
                      <Button
                        variant="ghost"
                        size="sm"
                        className="text-red-600 hover:text-red-700"
                        onClick={() => revokeMutation.mutate(key.id)}
                        disabled={revokeMutation.isPending}
                      >
                        <Trash2 size={14} />
                        Revoke
                      </Button>
                    </TableCell>
                  </TableRow>
                ))}
              </TableBody>
            </Table>
          )}
        </CardContent>
      </Card>

      <Card>
        <CardHeader>
          <CardTitle>Usage Example</CardTitle>
        </CardHeader>
        <CardContent>
          <pre className="rounded-lg bg-zinc-900 p-4 text-sm text-zinc-100 overflow-x-auto">
            <code>{`# Send traffic through the proxy with your API key
curl -H "X-API-Key: YOUR_API_KEY" \\
  https://proxy.yourdomain.com/api/endpoint

# Or use the CLI with your API key and project ID
surge check --api-key YOUR_API_KEY --project-id YOUR_PROJECT_ID`}</code>
          </pre>
          <p className="mt-3 text-xs text-zinc-500">
            Pass your API key via the <code className="rounded bg-zinc-100 px-1.5 py-0.5 text-zinc-700">X-API-Key</code> header or as a <code className="rounded bg-zinc-100 px-1.5 py-0.5 text-zinc-700">Bearer</code> token in the <code className="rounded bg-zinc-100 px-1.5 py-0.5 text-zinc-700">Authorization</code> header.
            Traffic sent with the key will automatically be associated with the linked project.
          </p>
        </CardContent>
      </Card>
    </div>
  );
}

export default function ApiKeysPage() {
  return (
    <Suspense fallback={<LoadingPage />}>
      <ApiKeysPageContent />
    </Suspense>
  );
}
