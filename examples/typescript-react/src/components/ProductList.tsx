/**
 * Product list component with filtering by category.
 *
 * This demonstrates using generated enum types for type-safe filtering.
 */

import { useState } from "react";
import { ProductCategory } from "../types/api";
import { useProducts } from "../hooks/useProducts";
import { useCart } from "../hooks/useCart";
import { ProductCard } from "./ProductCard";

// ============================================================================
// Category filter using generated enum
// ============================================================================

const ALL_CATEGORIES: ProductCategory[] = [
  ProductCategory.Electronics,
  ProductCategory.Clothing,
  ProductCategory.Home,
  ProductCategory.Books,
  ProductCategory.Sports,
  ProductCategory.Toys,
];

interface CategoryFilterProps {
  selected: ProductCategory | undefined;
  onChange: (category: ProductCategory | undefined) => void;
}

function CategoryFilter({ selected, onChange }: CategoryFilterProps) {
  return (
    <div className="category-filter">
      <button
        className={selected === undefined ? "active" : ""}
        onClick={() => onChange(undefined)}
      >
        All
      </button>
      {ALL_CATEGORIES.map((category) => (
        <button
          key={category}
          className={selected === category ? "active" : ""}
          onClick={() => onChange(category)}
        >
          {category.charAt(0).toUpperCase() + category.slice(1)}
        </button>
      ))}
    </div>
  );
}

// ============================================================================
// Product list component
// ============================================================================

export function ProductList() {
  // Category state is typed as ProductCategory | undefined
  const [category, setCategory] = useState<ProductCategory | undefined>();

  // Hook returns typed Product[] from the generated types
  const { products, loading, error, hasMore, loadMore } = useProducts({
    category,
    limit: 12,
  });

  const { addToCart } = useCart();

  const handleAddToCart = async (productId: string, quantity: number) => {
    try {
      // Request body is type-checked against AddToCartRequest
      await addToCart({ productId, quantity });
    } catch {
      // Error handling - error is typed as ApiError
      console.error("Failed to add to cart");
    }
  };

  if (error) {
    return <div className="error">Error: {error}</div>;
  }

  return (
    <div className="product-list-container">
      <CategoryFilter selected={category} onChange={setCategory} />

      {loading && products.length === 0 ? (
        <div className="loading">Loading products...</div>
      ) : (
        <>
          <div className="product-grid">
            {products.map((product) => (
              <ProductCard
                key={product.id}
                product={product}
                onAddToCart={handleAddToCart}
              />
            ))}
          </div>

          {hasMore && (
            <button
              className="load-more"
              onClick={loadMore}
              disabled={loading}
            >
              {loading ? "Loading..." : "Load More"}
            </button>
          )}
        </>
      )}
    </div>
  );
}

