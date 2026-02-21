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
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { EmptyState } from "@/components/ui/empty-state";
import { LoadingPage } from "@/components/ui/loading-spinner";
import { CopyButton } from "@/components/ui/copy-button";
import { useState } from "react";
import Link from "next/link";
import { apiKeysApi } from "@/lib/api/api-keys";
import { toast } from "sonner";

function ApiKeysPageContent() {
  const searchParams = useSearchParams();
  const orgId = searchParams.get("org") || "";
  const queryClient = useQueryClient();
  const [createOpen, setCreateOpen] = useState(false);
  const [keyName, setKeyName] = useState("");
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

  const createMutation = useMutation({
    mutationFn: async (name: string) => {
      if (!orgId) throw new Error("Organization ID required");
      return await apiKeysApi.create(orgId, { name });
    },
    onSuccess: (data) => {
      queryClient.invalidateQueries({ queryKey: ["api-keys", orgId] });
      setNewKey(data.key);
      setKeyName("");
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
    createMutation.mutate(keyName);
  };

  const maskKey = (prefix: string) => `${prefix}${"•".repeat(32)}`;

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
            Manage API keys for programmatic access
          </p>
        </div>

        <Dialog
          open={createOpen}
          onOpenChange={(open) => {
            setCreateOpen(open);
            if (!open) setNewKey(null);
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
                  <div className="flex items-center gap-2 rounded-lg border border-zinc-200 bg-zinc-50 p-3">
                    <code className="flex-1 text-sm font-mono">{newKey}</code>
                    <CopyButton value={newKey} />
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
                    Generate a new API key for programmatic access.
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
                    disabled={createMutation.isPending || !keyName}
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
            API keys allow programmatic access to the TVC API
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
                      <div className="flex items-center gap-2">
                        <code className="text-sm font-mono text-zinc-600">
                          {maskKey(key.key_prefix)}
                        </code>
                        <CopyButton value={key.key_prefix} />
                      </div>
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
            <code>{`curl -H "Authorization: Bearer YOUR_API_KEY" \\
  https://api.driftsurge.dev/v1/projects/{projectId}/traffic`}</code>
          </pre>
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
