"use client";

import { Suspense } from "react";
import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { useSearchParams } from "next/navigation";
import { Users, UserPlus, Mail, Trash2, Crown } from "lucide-react";
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
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Badge } from "@/components/ui/badge";
import { EmptyState } from "@/components/ui/empty-state";
import { LoadingPage } from "@/components/ui/loading-spinner";
import { useState } from "react";
import Link from "next/link";
import { organizationsApi } from "@/lib/api/organizations";
import { toast } from "sonner";

interface TeamMember {
  id: string;
  email: string;
  name: string;
  role: "owner" | "admin" | "member";
  status: "active" | "pending";
  joined_at: string;
}

function TeamPageContent() {
  const searchParams = useSearchParams();
  const orgId = searchParams.get("org") || "";
  const queryClient = useQueryClient();
  const [inviteOpen, setInviteOpen] = useState(false);
  const [inviteEmail, setInviteEmail] = useState("");
  const [inviteRole, setInviteRole] = useState<"admin" | "member" | "viewer">(
    "member",
  );

  const { data: members, isLoading } = useQuery({
    queryKey: ["org-members", orgId],
    queryFn: async () => {
      return await organizationsApi.listMembers(orgId);
    },
    enabled: !!orgId,
  });

  const inviteMutation = useMutation({
    mutationFn: async (data: { email: string; role: string }) => {
      return await organizationsApi.addMember(orgId, data);
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["org-members", orgId] });
      setInviteOpen(false);
      setInviteEmail("");
      toast.success("Member invited successfully");
    },
    onError: (error: any) => {
      toast.error(error.message || "Failed to invite member");
    },
  });

  const removeMutation = useMutation({
    mutationFn: async (memberId: string) => {
      return await organizationsApi.removeMember(orgId, memberId);
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["org-members", orgId] });
      toast.success("Member removed successfully");
    },
    onError: (error: any) => {
      toast.error(error.message || "Failed to remove member");
    },
  });

  const handleInvite = (e: React.FormEvent) => {
    e.preventDefault();
    inviteMutation.mutate({ email: inviteEmail, role: inviteRole });
  };

  if (isLoading) {
    return <LoadingPage />;
  }

  if (!orgId) {
    return (
      <EmptyState
        icon={<Users size={28} className="text-zinc-400" />}
        title="No organization selected"
        description="Select an organization to view and manage team members."
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
          <h1 className="text-2xl font-semibold text-zinc-900">Team Members</h1>
          <p className="mt-1 text-sm text-zinc-500">
            Manage who has access to this project
          </p>
        </div>

        <Dialog open={inviteOpen} onOpenChange={setInviteOpen}>
          <DialogTrigger asChild>
            <Button>
              <UserPlus size={16} />
              Invite Member
            </Button>
          </DialogTrigger>
          <DialogContent>
            <form onSubmit={handleInvite}>
              <DialogHeader>
                <DialogTitle>Invite team member</DialogTitle>
                <DialogDescription>
                  Send an invitation to join this project.
                </DialogDescription>
              </DialogHeader>
              <div className="space-y-4 py-4">
                <div className="space-y-2">
                  <Label htmlFor="email" required>
                    Email address
                  </Label>
                  <Input
                    id="email"
                    type="email"
                    placeholder="colleague@example.com"
                    value={inviteEmail}
                    onChange={(e) => setInviteEmail(e.target.value)}
                    required
                  />
                </div>
                <div className="space-y-2">
                  <Label htmlFor="role" required>
                    Role
                  </Label>
                  <select
                    id="role"
                    value={inviteRole}
                    onChange={(e) =>
                      setInviteRole(e.target.value as "admin" | "member")
                    }
                    className="flex h-10 w-full rounded-lg border border-zinc-200 bg-white px-3 py-2 text-sm"
                  >
                    <option value="member">Member — Can view and create</option>
                    <option value="admin">Admin — Can manage settings</option>
                  </select>
                </div>
              </div>
              <DialogFooter>
                <Button
                  type="button"
                  variant="secondary"
                  onClick={() => setInviteOpen(false)}
                >
                  Cancel
                </Button>
                <Button
                  type="submit"
                  disabled={inviteMutation.isPending || !inviteEmail}
                >
                  {inviteMutation.isPending ? "Sending..." : "Send Invitation"}
                </Button>
              </DialogFooter>
            </form>
          </DialogContent>
        </Dialog>
      </div>

      {!members || members.length === 0 ? (
        <Card>
          <CardContent className="py-12">
            <EmptyState
              icon={<Users size={28} className="text-zinc-400" />}
              title="No team members yet"
              description="Invite your team members to collaborate on this project."
              action={
                <Button onClick={() => setInviteOpen(true)}>
                  <UserPlus size={16} />
                  Invite First Member
                </Button>
              }
            />
          </CardContent>
        </Card>
      ) : (
        <Card>
          <CardHeader>
            <CardTitle>Members</CardTitle>
            <CardDescription>
              {members.length} {members.length === 1 ? "member" : "members"}
            </CardDescription>
          </CardHeader>
          <CardContent>
            <Table>
              <TableHeader>
                <TableRow>
                  <TableHead>Name</TableHead>
                  <TableHead>Email</TableHead>
                  <TableHead>Role</TableHead>
                  <TableHead>Status</TableHead>
                  <TableHead>Joined</TableHead>
                  <TableHead className="text-right">Actions</TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {members.map((member) => (
                  <TableRow key={member.id}>
                    <TableCell className="font-medium">{member.name}</TableCell>
                    <TableCell className="text-zinc-500">
                      {member.email}
                    </TableCell>
                    <TableCell>
                      <div className="flex items-center gap-1.5">
                        {member.role === "owner" && (
                          <Crown size={14} className="text-amber-500" />
                        )}
                        <span className="capitalize">{member.role}</span>
                      </div>
                    </TableCell>
                    <TableCell>
                      <Badge
                        variant={
                          member.status === "active" ? "success" : "warning"
                        }
                      >
                        {member.status}
                      </Badge>
                    </TableCell>
                    <TableCell className="text-zinc-500">
                      {new Date(member.joined_at).toLocaleDateString()}
                    </TableCell>
                    <TableCell className="text-right">
                      <DropdownMenu>
                        <DropdownMenuTrigger asChild>
                          <Button variant="ghost" size="sm">
                            •••
                          </Button>
                        </DropdownMenuTrigger>
                        <DropdownMenuContent align="end">
                          <DropdownMenuItem>
                            <Mail size={14} />
                            Resend invitation
                          </DropdownMenuItem>
                          {member.role !== "owner" && (
                            <>
                              <DropdownMenuSeparator />
                              <DropdownMenuItem
                                className="text-red-600"
                                onClick={() => removeMutation.mutate(member.id)}
                              >
                                <Trash2 size={14} />
                                Remove member
                              </DropdownMenuItem>
                            </>
                          )}
                        </DropdownMenuContent>
                      </DropdownMenu>
                    </TableCell>
                  </TableRow>
                ))}
              </TableBody>
            </Table>
          </CardContent>
        </Card>
      )}
    </div>
  );
}

export default function TeamPage() {
  return (
    <Suspense fallback={<LoadingPage />}>
      <TeamPageContent />
    </Suspense>
  );
}
