"use client";

import { Suspense } from "react";
import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { useSearchParams } from "next/navigation";
import { Server, Plus, Trash2, Settings } from "lucide-react";
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
} from "@/components/ui/dialog";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Textarea } from "@/components/ui/textarea";
import { Badge } from "@/components/ui/badge";
import { EmptyState } from "@/components/ui/empty-state";
import { LoadingPage } from "@/components/ui/loading-spinner";
import { CopyButton } from "@/components/ui/copy-button";
import { useState } from "react";
import Link from "next/link";
import { environmentsApi } from "@/lib/api/environments";
import { toast } from "sonner";

interface Environment {
  id: string;
  name: string;
  slug: string;
  base_url: string;
  description: string;
  is_default: boolean;
  created_at: string;
}

function EnvironmentsPageContent() {
  const searchParams = useSearchParams();
  const projectId = searchParams.get("project") || "";
  const queryClient = useQueryClient();
  const [createOpen, setCreateOpen] = useState(false);
  const [formData, setFormData] = useState({
    name: "",
    base_url: "",
    is_source: false,
  });

  const { data: environments, isLoading } = useQuery({
    queryKey: ["environments", projectId],
    queryFn: async () => {
      return await environmentsApi.list(projectId);
    },
    enabled: !!projectId,
  });

  const createMutation = useMutation({
    mutationFn: async (data: typeof formData) => {
      return await environmentsApi.create(projectId, data);
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["environments", projectId] });
      setCreateOpen(false);
      setFormData({ name: "", base_url: "", is_source: false });
      toast.success("Environment created successfully");
    },
    onError: (error: any) => {
      toast.error(error.message || "Failed to create environment");
    },
  });

  const deleteMutation = useMutation({
    mutationFn: async (envId: string) => {
      return await environmentsApi.delete(projectId, envId);
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["environments", projectId] });
      toast.success("Environment deleted successfully");
    },
    onError: (error: any) => {
      toast.error(error.message || "Failed to delete environment");
    },
  });

  const handleCreate = (e: React.FormEvent) => {
    e.preventDefault();
    createMutation.mutate(formData);
  };

  const handleNameChange = (name: string) => {
    setFormData({ ...formData, name });
  };

  if (isLoading) {
    return <LoadingPage />;
  }

  if (!projectId) {
    return (
      <EmptyState
        icon={<Server size={28} className="text-zinc-400" />}
        title="No project selected"
        description="Select a project to manage environments."
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
          <h1 className="text-2xl font-semibold text-zinc-900">Environments</h1>
          <p className="mt-1 text-sm text-zinc-500">
            Configure environments for traffic replay and testing
          </p>
        </div>

        <Dialog open={createOpen} onOpenChange={setCreateOpen}>
          <DialogTrigger asChild>
            <Button>
              <Plus size={16} />
              Add Environment
            </Button>
          </DialogTrigger>
          <DialogContent>
            <form onSubmit={handleCreate}>
              <DialogHeader>
                <DialogTitle>Create environment</DialogTitle>
                <DialogDescription>
                  Add a new environment for traffic replay.
                </DialogDescription>
              </DialogHeader>
              <div className="space-y-4 py-4">
                <div className="space-y-2">
                  <Label htmlFor="name" required>
                    Environment name
                  </Label>
                  <Input
                    id="name"
                    placeholder="Staging"
                    value={formData.name}
                    onChange={(e) => handleNameChange(e.target.value)}
                    required
                  />
                </div>
                <div className="space-y-2"></div>
                <div className="space-y-2">
                  <Label htmlFor="base_url" required>
                    Base URL
                  </Label>
                  <Input
                    id="base_url"
                    type="url"
                    placeholder="https://staging.api.example.com"
                    value={formData.base_url}
                    onChange={(e) =>
                      setFormData({ ...formData, base_url: e.target.value })
                    }
                    required
                  />
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
                  disabled={
                    createMutation.isPending ||
                    !formData.name ||
                    !formData.base_url
                  }
                >
                  {createMutation.isPending
                    ? "Creating..."
                    : "Create Environment"}
                </Button>
              </DialogFooter>
            </form>
          </DialogContent>
        </Dialog>
      </div>

      {!environments || environments.length === 0 ? (
        <Card>
          <CardContent className="py-12">
            <EmptyState
              icon={<Server size={28} className="text-zinc-400" />}
              title="No environments configured"
              description="Create environments to replay captured traffic against different API servers."
              action={
                <Button onClick={() => setCreateOpen(true)}>
                  <Plus size={16} />
                  Create First Environment
                </Button>
              }
            />
          </CardContent>
        </Card>
      ) : (
        <div className="grid gap-4 md:grid-cols-2">
          {environments.map((env) => (
            <Card key={env.id}>
              <CardHeader>
                <div className="flex items-start justify-between">
                  <div className="flex items-center gap-2">
                    <Server size={18} className="text-zinc-500" />
                    <CardTitle className="text-lg">{env.name}</CardTitle>
                    {env.is_default && (
                      <Badge variant="info" className="text-xs">
                        Default
                      </Badge>
                    )}
                  </div>
                  <Button
                    variant="ghost"
                    size="sm"
                    className="text-red-600"
                    onClick={() => deleteMutation.mutate(env.id)}
                    disabled={env.is_default}
                  >
                    <Trash2 size={14} />
                  </Button>
                </div>
                <CardDescription>
                  {env.description || "No description"}
                </CardDescription>
              </CardHeader>
              <CardContent className="space-y-3">
                <div>
                  <Label className="text-xs text-zinc-500">Slug</Label>
                  <div className="mt-1 flex items-center gap-2">
                    <code className="text-sm font-mono text-zinc-900">
                      {env.slug}
                    </code>
                    <CopyButton value={env.slug} />
                  </div>
                </div>
                <div>
                  <Label className="text-xs text-zinc-500">Base URL</Label>
                  <div className="mt-1 flex items-center gap-2">
                    <code className="text-sm font-mono text-zinc-600 truncate">
                      {env.base_url}
                    </code>
                    <CopyButton value={env.base_url} />
                  </div>
                </div>
                <div className="pt-2 text-xs text-zinc-400">
                  Created {new Date(env.created_at).toLocaleDateString()}
                </div>
              </CardContent>
            </Card>
          ))}
        </div>
      )}

      <Card>
        <CardHeader>
          <CardTitle>How Environments Work</CardTitle>
        </CardHeader>
        <CardContent className="space-y-2 text-sm text-zinc-600">
          <p>
            Environments define target servers where TVC replays captured
            traffic:
          </p>
          <ul className="list-disc list-inside space-y-1 ml-2">
            <li>Each replay session targets one environment</li>
            <li>Base URL is used to reconstruct request destinations</li>
            <li>
              Path and query parameters are preserved from original traffic
            </li>
            <li>
              Authentication headers may need environment-specific configuration
            </li>
          </ul>
        </CardContent>
      </Card>
    </div>
  );
}

export default function EnvironmentsPage() {
  return (
    <Suspense fallback={<LoadingPage />}>
      <EnvironmentsPageContent />
    </Suspense>
  );
}
