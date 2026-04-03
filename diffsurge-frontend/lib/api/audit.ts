import { apiRequest } from "./client";

export type AuditAction =
  | "create"
  | "update"
  | "delete"
  | "invite"
  | "remove"
  | "login"
  | "logout"
  | "access";

export interface AuditLog {
  id: string;
  organization_id: string;
  user_id?: string;
  action: AuditAction;
  resource_type: string;
  resource_id?: string;
  details?: Record<string, any>;
  ip_address?: string;
  user_agent?: string;
  created_at: string;
}

export interface AuditLogFilter {
  user_id?: string;
  action?: AuditAction;
  resource_type?: string;
  start_time?: string; // RFC3339 format
  end_time?: string; // RFC3339 format
  limit?: number;
  offset?: number;
}

export const auditApi = {
  list: async (orgId: string, filter?: AuditLogFilter): Promise<AuditLog[]> => {
    const params = new URLSearchParams();

    if (filter?.user_id) params.append("user_id", filter.user_id);
    if (filter?.action) params.append("action", filter.action);
    if (filter?.resource_type)
      params.append("resource_type", filter.resource_type);
    if (filter?.start_time) params.append("start_time", filter.start_time);
    if (filter?.end_time) params.append("end_time", filter.end_time);
    if (filter?.limit) params.append("limit", filter.limit.toString());
    if (filter?.offset) params.append("offset", filter.offset.toString());

    const queryString = params.toString();
    const url = `/api/v1/organizations/${orgId}/audit-logs${queryString ? `?${queryString}` : ""}`;

    const response = await apiRequest<{ data: AuditLog[] } | AuditLog[]>(url);
    return Array.isArray(response) ? response : response.data ?? [];
  },
};
