/**
 * React hook for managing the shopping cart.
 *
 * This demonstrates type-safe state management with generated types.
 */

import { useState, useCallback } from "react";
import type { Cart, AddToCartRequest, ApiError } from "../types/api";
import { getCart, addToCart as addToCartApi, ApiRequestError } from "../api/client";

interface UseCartResult {
  cart: Cart | null;
  loading: boolean;
  error: ApiError | null;
  fetchCart: () => Promise<void>;
  addToCart: (request: AddToCartRequest) => Promise<void>;
  itemCount: number;
}

export function useCart(): UseCartResult {
  const [cart, setCart] = useState<Cart | null>(null);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<ApiError | null>(null);

  const fetchCart = useCallback(async () => {
    setLoading(true);
    setError(null);

    try {
      const data = await getCart();
      setCart(data);
    } catch (err) {
      if (err instanceof ApiRequestError) {
        setError(err.error);
      }
    } finally {
      setLoading(false);
    }
  }, []);

  const addToCart = useCallback(async (request: AddToCartRequest) => {
    setLoading(true);
    setError(null);

    try {
      // The request body is type-checked against AddToCartRequest
      const updatedCart = await addToCartApi(request);
      setCart(updatedCart);
    } catch (err) {
      if (err instanceof ApiRequestError) {
        setError(err.error);
      }
      throw err; // Re-throw so the caller can handle it
    } finally {
      setLoading(false);
    }
  }, []);

  return {
    cart,
    loading,
    error,
    fetchCart,
    addToCart,
    // Type-safe computed property - itemCount is optional in Cart
    itemCount: cart?.itemCount ?? 0,
  };
}

