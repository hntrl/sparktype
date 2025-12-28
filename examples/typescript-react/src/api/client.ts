/**
 * Type-safe API client using generated types from sparktype.
 *
 * This demonstrates how to create a fetch wrapper that:
 * - Uses generated TypeScript types for request/response bodies
 * - Validates API responses at runtime using Zod schemas
 * - Provides full type inference for consumers
 */

import type {
  Product,
  ProductCategory,
  ProductListResponse,
  Cart,
  Order,
  CreateProductRequest,
  AddToCartRequest,
  CreateOrderRequest,
  ApiError,
} from "../types/api";

import {
  productSchema,
  productListResponseSchema,
  cartSchema,
  orderSchema,
  apiErrorSchema,
} from "../schemas/api";

// Base API configuration
const API_BASE = import.meta.env.VITE_API_URL || "http://localhost:3000/api";

// Custom error class for API errors
export class ApiRequestError extends Error {
  constructor(
    public statusCode: number,
    public error: ApiError
  ) {
    super(error.message);
    this.name = "ApiRequestError";
  }
}

// Generic fetch wrapper with response validation
async function request<T>(
  path: string,
  options: RequestInit = {},
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
    // Validate error response matches our expected format
    const error = apiErrorSchema.parse(data);
    throw new ApiRequestError(response.status, error);
  }

  // Validate successful response matches expected schema
  // This catches API contract violations at runtime
  return schema.parse(data);
}

// ============================================================================
// Products API
// ============================================================================

export interface ListProductsParams {
  category?: ProductCategory;
  limit?: number;
  offset?: number;
}

export async function listProducts(
  params: ListProductsParams = {}
): Promise<ProductListResponse> {
  const searchParams = new URLSearchParams();

  if (params.category) searchParams.set("category", params.category);
  if (params.limit) searchParams.set("limit", String(params.limit));
  if (params.offset) searchParams.set("offset", String(params.offset));

  const query = searchParams.toString();
  const path = `/products${query ? `?${query}` : ""}`;

  return request(path, { method: "GET" }, productListResponseSchema);
}

export async function getProduct(id: string): Promise<Product> {
  return request(`/products/${id}`, { method: "GET" }, productSchema);
}

export async function createProduct(
  data: CreateProductRequest
): Promise<Product> {
  return request(
    "/products",
    {
      method: "POST",
      body: JSON.stringify(data),
    },
    productSchema
  );
}

// ============================================================================
// Cart API
// ============================================================================

export async function getCart(): Promise<Cart> {
  return request("/cart", { method: "GET" }, cartSchema);
}

export async function addToCart(data: AddToCartRequest): Promise<Cart> {
  return request(
    "/cart",
    {
      method: "POST",
      body: JSON.stringify(data),
    },
    cartSchema
  );
}

// ============================================================================
// Orders API
// ============================================================================

export async function createOrder(data: CreateOrderRequest): Promise<Order> {
  return request(
    "/orders",
    {
      method: "POST",
      body: JSON.stringify(data),
    },
    orderSchema
  );
}

