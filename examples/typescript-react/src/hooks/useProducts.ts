/**
 * React hooks for fetching and managing products.
 *
 * This demonstrates how generated types flow through your application,
 * providing type safety from API response to component props.
 */

import { useState, useEffect, useCallback } from "react";
import type {
  Product,
  ProductCategory,
  ProductListResponse,
} from "../types/api";
import { listProducts, getProduct, ApiRequestError } from "../api/client";

// ============================================================================
// useProducts - Fetch a list of products
// ============================================================================

interface UseProductsOptions {
  category?: ProductCategory;
  limit?: number;
  initialFetch?: boolean;
}

interface UseProductsResult {
  products: Product[];
  total: number;
  hasMore: boolean;
  loading: boolean;
  error: string | null;
  loadMore: () => Promise<void>;
  refetch: () => Promise<void>;
}

export function useProducts(options: UseProductsOptions = {}): UseProductsResult {
  const { category, limit = 20, initialFetch = true } = options;

  const [data, setData] = useState<ProductListResponse | null>(null);
  const [loading, setLoading] = useState(initialFetch);
  const [error, setError] = useState<string | null>(null);
  const [offset, setOffset] = useState(0);

  const fetchProducts = useCallback(
    async (currentOffset: number, append = false) => {
      setLoading(true);
      setError(null);

      try {
        const response = await listProducts({
          category,
          limit,
          offset: currentOffset,
        });

        setData((prev) => {
          if (append && prev) {
            return {
              ...response,
              items: [...prev.items, ...response.items],
            };
          }
          return response;
        });
      } catch (err) {
        if (err instanceof ApiRequestError) {
          setError(err.error.message);
        } else {
          setError("An unexpected error occurred");
        }
      } finally {
        setLoading(false);
      }
    },
    [category, limit]
  );

  // Initial fetch
  useEffect(() => {
    if (initialFetch) {
      fetchProducts(0);
    }
  }, [fetchProducts, initialFetch]);

  const loadMore = useCallback(async () => {
    if (!data?.hasMore || loading) return;

    const newOffset = offset + limit;
    setOffset(newOffset);
    await fetchProducts(newOffset, true);
  }, [data, loading, offset, limit, fetchProducts]);

  const refetch = useCallback(async () => {
    setOffset(0);
    await fetchProducts(0);
  }, [fetchProducts]);

  return {
    products: data?.items ?? [],
    total: data?.total ?? 0,
    hasMore: data?.hasMore ?? false,
    loading,
    error,
    loadMore,
    refetch,
  };
}

// ============================================================================
// useProduct - Fetch a single product by ID
// ============================================================================

interface UseProductResult {
  product: Product | null;
  loading: boolean;
  error: string | null;
  refetch: () => Promise<void>;
}

export function useProduct(id: string | null): UseProductResult {
  const [product, setProduct] = useState<Product | null>(null);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const fetchProduct = useCallback(async () => {
    if (!id) return;

    setLoading(true);
    setError(null);

    try {
      const data = await getProduct(id);
      setProduct(data);
    } catch (err) {
      if (err instanceof ApiRequestError) {
        setError(err.error.message);
      } else {
        setError("An unexpected error occurred");
      }
      setProduct(null);
    } finally {
      setLoading(false);
    }
  }, [id]);

  useEffect(() => {
    fetchProduct();
  }, [fetchProduct]);

  return {
    product,
    loading,
    error,
    refetch: fetchProduct,
  };
}

