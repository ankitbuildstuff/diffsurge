"use client";

import {
  createContext,
  useContext,
  useEffect,
  useRef,
  useState,
  type ReactNode,
} from "react";
import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { createClient } from "@/lib/supabase/client";
import { organizationsApi, type Organization } from "@/lib/api/organizations";
import type { User as SupabaseUser } from "@supabase/supabase-js";

interface OrganizationContextValue {
  user: SupabaseUser | null;
  organizations: Organization[];
  activeOrg: Organization | null;
  isLoading: boolean;
}

const OrganizationContext = createContext<OrganizationContextValue>({
  user: null,
  organizations: [],
  activeOrg: null,
  isLoading: true,
});

export function useOrganization() {
  return useContext(OrganizationContext);
}

export function OrganizationProvider({ children }: { children: ReactNode }) {
  const [user, setUser] = useState<SupabaseUser | null>(null);
  const [userLoading, setUserLoading] = useState(true);
  const queryClient = useQueryClient();
  const autoCreateAttempted = useRef(false);

  useEffect(() => {
    const supabase = createClient();
    supabase.auth.getUser().then(({ data }) => {
      setUser(data.user);
      setUserLoading(false);
    });
  }, []);

  const {
    data: organizations = [],
    isLoading: orgsLoading,
    isError: orgsError,
  } = useQuery({
    queryKey: ["organizations"],
    queryFn: () => organizationsApi.list(),
    enabled: !!user,
    staleTime: 60_000,
    retry: 1,
  });

  const createOrgMutation = useMutation({
    mutationFn: (name: string) => organizationsApi.create({ name }),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["organizations"] });
    },
  });

  useEffect(() => {
    if (
      user &&
      !orgsLoading &&
      !orgsError &&
      organizations.length === 0 &&
      !autoCreateAttempted.current
    ) {
      autoCreateAttempted.current = true;
      const orgName =
        user.user_metadata?.full_name ||
        user.email?.split("@")[0] ||
        "My Organization";
      createOrgMutation.mutate(`${orgName}'s Org`);
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [user, orgsLoading, orgsError, organizations.length]);

  const activeOrg = organizations[0] ?? null;
  const isLoading = userLoading || orgsLoading || createOrgMutation.isPending;

  return (
    <OrganizationContext.Provider
      value={{ user, organizations, activeOrg, isLoading }}
    >
      {children}
    </OrganizationContext.Provider>
  );
}
