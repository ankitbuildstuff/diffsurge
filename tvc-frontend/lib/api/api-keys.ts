import { apiRequest } from "./client";

export interface APIKey {
  id: string;
  organization_id: string;
  project_id?: string;
  name: string;
  key_prefix: string;
  last_used_at?: string;
  expires_at?: string;
  created_at: string;
  created_by: string;
}

export interface CreateAPIKeyInput {
  name: string;
  project_id?: string;
  expires_at?: string;
}

export interface CreateAPIKeyResponse {
  key: string;
  api_key: APIKey;
}

export const apiKeysApi = {
  list: async (orgId: string): Promise<APIKey[]> => {
    const response = await apiRequest<{ data: APIKey[] }>(
      `/api/v1/organizations/${orgId}/api-keys`,
    );
    return response.data;
  },

  create: async (
    orgId: string,
    input: CreateAPIKeyInput,
  ): Promise<CreateAPIKeyResponse> => {
    return await apiRequest<CreateAPIKeyResponse>(
      `/api/v1/organizations/${orgId}/api-keys`,
      {
        method: "POST",
        body: input,
      },
    );
  },

  delete: async (orgId: string, keyId: string): Promise<void> => {
    await apiRequest(`/api/v1/organizations/${orgId}/api-keys/${keyId}`, {
      method: "DELETE",
    });
  },
};
