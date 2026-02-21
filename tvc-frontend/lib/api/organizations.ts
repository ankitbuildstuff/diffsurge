import { apiRequest } from "./client";

export interface Organization {
  id: string;
  name: string;
  slug: string;
  created_at: string;
  updated_at: string;
}

export interface OrganizationMember {
  user_id: string;
  email: string;
  full_name?: string;
  role: "admin" | "member" | "viewer";
  joined_at: string;
}

export interface CreateOrganizationInput {
  name: string;
  slug?: string;
}

export interface UpdateOrganizationInput {
  name?: string;
  slug?: string;
}

export interface AddMemberInput {
  user_id?: string;
  email?: string;
  role: "admin" | "member" | "viewer";
}

export const organizationsApi = {
  list: async (): Promise<Organization[]> => {
    const response = await apiRequest<{ data: Organization[] }>(
      "/api/v1/organizations",
    );
    return response.data;
  },

  get: async (orgId: string): Promise<Organization> => {
    const response = await apiRequest<{ data: Organization }>(
      `/api/v1/organizations/${orgId}`,
    );
    return response.data;
  },

  create: async (input: CreateOrganizationInput): Promise<Organization> => {
    const response = await apiRequest<{ data: Organization }>(
      "/api/v1/organizations",
      {
        method: "POST",
        body: input,
      },
    );
    return response.data;
  },

  update: async (
    orgId: string,
    input: UpdateOrganizationInput,
  ): Promise<Organization> => {
    const response = await apiRequest<{ data: Organization }>(
      `/api/v1/organizations/${orgId}`,
      {
        method: "PUT",
        body: input,
      },
    );
    return response.data;
  },

  delete: async (orgId: string): Promise<void> => {
    await apiRequest(`/api/v1/organizations/${orgId}`, {
      method: "DELETE",
    });
  },

  // Member management
  listMembers: async (orgId: string): Promise<OrganizationMember[]> => {
    const response = await apiRequest<{ data: OrganizationMember[] }>(
      `/api/v1/organizations/${orgId}/members`,
    );
    return response.data;
  },

  addMember: async (
    orgId: string,
    input: AddMemberInput,
  ): Promise<{ message: string }> => {
    const response = await apiRequest<{ data: { message: string } }>(
      `/api/v1/organizations/${orgId}/members`,
      {
        method: "POST",
        body: input,
      },
    );
    return response.data;
  },

  removeMember: async (orgId: string, userId: string): Promise<void> => {
    await apiRequest(`/api/v1/organizations/${orgId}/members/${userId}`, {
      method: "DELETE",
    });
  },
};
