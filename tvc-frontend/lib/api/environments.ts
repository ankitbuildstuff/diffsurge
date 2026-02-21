import { apiRequest } from "./client";

export interface Environment {
  id: string;
  project_id: string;
  name: string;
  base_url: string;
  is_source: boolean;
  created_at: string;
}

export interface CreateEnvironmentInput {
  name: string;
  base_url: string;
  is_source: boolean;
}

export interface UpdateEnvironmentInput {
  name?: string;
  base_url?: string;
  is_source?: boolean;
}

export const environmentsApi = {
  list: async (projectId: string): Promise<Environment[]> => {
    const response = await apiRequest<{ data: Environment[] }>(
      `/api/v1/projects/${projectId}/environments`,
    );
    return response.data;
  },

  get: async (projectId: string, envId: string): Promise<Environment> => {
    const response = await apiRequest<{ data: Environment }>(
      `/api/v1/projects/${projectId}/environments/${envId}`,
    );
    return response.data;
  },

  create: async (
    projectId: string,
    input: CreateEnvironmentInput,
  ): Promise<Environment> => {
    const response = await apiRequest<{ data: Environment }>(
      `/api/v1/projects/${projectId}/environments`,
      {
        method: "POST",
        body: input,
      },
    );
    return response.data;
  },

  update: async (
    projectId: string,
    envId: string,
    input: UpdateEnvironmentInput,
  ): Promise<Environment> => {
    const response = await apiRequest<{ data: Environment }>(
      `/api/v1/projects/${projectId}/environments/${envId}`,
      {
        method: "PUT",
        body: input,
      },
    );
    return response.data;
  },

  delete: async (projectId: string, envId: string): Promise<void> => {
    await apiRequest(`/api/v1/projects/${projectId}/environments/${envId}`, {
      method: "DELETE",
    });
  },
};
