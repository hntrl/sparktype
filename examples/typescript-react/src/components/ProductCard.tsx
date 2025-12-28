/**
 * Product card component demonstrating type-safe props with generated types.
 */

import type { Product, ProductCategory } from "../types/api";

// ============================================================================
// Type-safe props using generated types
// ============================================================================

interface ProductCardProps {
  product: Product;
  onAddToCart?: (productId: string, quantity: number) => void;
}

// ============================================================================
// Helper functions with type safety
// ============================================================================

/**
 * Format a price for display.
 * The price property is typed from the OpenAPI spec.
 */
function formatPrice(price: Product["price"]): string {
  // price.formatted is optional (readOnly in spec), so we provide a fallback
  if (price.formatted) {
    return price.formatted;
  }

  const amount = price.amount / 100; // Convert from cents
  return new Intl.NumberFormat("en-US", {
    style: "currency",
    currency: price.currency,
  }).format(amount);
}

/**
 * Get display name for a category.
 * ProductCategory is a string enum from the spec.
 */
function getCategoryLabel(category: ProductCategory): string {
  const labels: Record<ProductCategory, string> = {
    electronics: "Electronics",
    clothing: "Clothing",
    home: "Home & Garden",
    books: "Books",
    sports: "Sports & Outdoors",
    toys: "Toys & Games",
  };
  return labels[category];
}

// ============================================================================
// Component
// ============================================================================

export function ProductCard({ product, onAddToCart }: ProductCardProps) {
  // All properties are type-checked against the Product schema
  const primaryImage = product.images?.find((img) => img.isPrimary) ?? product.images?.[0];

  return (
    <div className="product-card">
      {primaryImage && (
        <img
          src={primaryImage.url}
          alt={primaryImage.alt}
          className="product-image"
        />
      )}

      <div className="product-info">
        <span className="product-category">
          {getCategoryLabel(product.category)}
        </span>

        <h3 className="product-name">{product.name}</h3>

        {product.description && (
          <p className="product-description">{product.description}</p>
        )}

        <div className="product-price">{formatPrice(product.price)}</div>

        {/* Tags are optional in the schema */}
        {product.tags && product.tags.length > 0 && (
          <div className="product-tags">
            {product.tags.map((tag) => (
              <span key={tag} className="tag">
                {tag}
              </span>
            ))}
          </div>
        )}

        <div className="product-stock">
          {product.inStock ? (
            <span className="in-stock">
              In Stock {product.stockCount && `(${product.stockCount} available)`}
            </span>
          ) : (
            <span className="out-of-stock">Out of Stock</span>
          )}
        </div>

        {onAddToCart && product.inStock && (
          <button
            className="add-to-cart-btn"
            onClick={() => onAddToCart(product.id, 1)}
          >
            Add to Cart
          </button>
        )}
      </div>
    </div>
  );
}

