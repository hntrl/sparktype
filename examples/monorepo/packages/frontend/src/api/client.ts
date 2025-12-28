/**
 * API client using shared types from the monorepo.
 *
 * This demonstrates how generated types are shared across packages:
 * - TypeScript interfaces from ./types/api.ts (generated)
 * - Zod schemas from ./schemas/api.ts (generated)
 * - Domain-organized types from @example/shared (generated)
 */

import type {
  User,
  Workspace,
  Document,
  CreateUserRequest,
  CreateDocumentRequest,
  ApiError,
} from "../types/api";

import {
  userSchema,
  workspaceSchema,
  documentSchema,
  apiErrorSchema,
} from "../schemas/api";

// You can also import from the shared package for domain-organized types:
// import { Users, Documents, Workspaces } from "@example/shared";

const API_BASE = import.meta.env.VITE_API_URL || "/api";

export class ApiClient {
  private async request<T>(
    path: string,
    options: RequestInit,
    schema: { parse: (data: unknown) => T }
  ): Promise<T> {
    const response = await fetch(`${API_BASE}${path}`, {
      ...options,
      headers: {
        "Content-Type": "application/json",
        ...options.headers,
      },
    });

    const data = await response.json();

    if (!response.ok) {
      const error = apiErrorSchema.parse(data);
      throw new ApiRequestError(response.status, error);
    }

    return schema.parse(data);
  }

  // User endpoints - using generated types
  async getUser(id: string): Promise<User> {
    return this.request(`/users/${id}`, { method: "GET" }, userSchema);
  }

  async createUser(data: CreateUserRequest): Promise<User> {
    return this.request(
      "/users",
      { method: "POST", body: JSON.stringify(data) },
      userSchema
    );
  }

  // Workspace endpoints
  async getWorkspace(slug: string): Promise<Workspace> {
    return this.request(
      `/workspaces/${slug}`,
      { method: "GET" },
      workspaceSchema
    );
  }

  // Document endpoints
  async getDocument(id: string): Promise<Document> {
    return this.request(`/documents/${id}`, { method: "GET" }, documentSchema);
  }

  async createDocument(data: CreateDocumentRequest): Promise<Document> {
    return this.request(
      "/documents",
      { method: "POST", body: JSON.stringify(data) },
      documentSchema
    );
  }
}

export class ApiRequestError extends Error {
  constructor(
    public statusCode: number,
    public error: ApiError
  ) {
    super(error.message);
    this.name = "ApiRequestError";
  }
}

export const apiClient = new ApiClient();

