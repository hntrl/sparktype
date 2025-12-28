/**
 * Checkout form demonstrating type-safe form handling with generated types.
 *
 * This shows how to use generated request types for form state and validation.
 */

import { useState, type FormEvent } from "react";
import type { CreateOrderRequest, Address, Order } from "../types/api";
import { createOrderRequestSchema } from "../schemas/api";
import { createOrder, ApiRequestError } from "../api/client";

// ============================================================================
// Form state types derived from generated types
// ============================================================================

// Use the Address type from the spec for address form state
type AddressFormState = Omit<Address, "country"> & {
  country: string; // Allow any string during input, validate on submit
};

function createEmptyAddress(): AddressFormState {
  return {
    line1: "",
    line2: "",
    city: "",
    state: "",
    country: "",
    postalCode: "",
  };
}

// ============================================================================
// Component
// ============================================================================

interface CheckoutFormProps {
  onSuccess: (order: Order) => void;
  onCancel: () => void;
}

export function CheckoutForm({ onSuccess, onCancel }: CheckoutFormProps) {
  const [shippingAddress, setShippingAddress] = useState<AddressFormState>(
    createEmptyAddress()
  );
  const [billingAddress, setBillingAddress] = useState<AddressFormState>(
    createEmptyAddress()
  );
  const [sameAsShipping, setSameAsShipping] = useState(true);
  const [notes, setNotes] = useState("");

  const [submitting, setSubmitting] = useState(false);
  const [errors, setErrors] = useState<Record<string, string>>({});

  const handleAddressChange = (
    setter: typeof setShippingAddress,
    field: keyof AddressFormState,
    value: string
  ) => {
    setter((prev) => ({ ...prev, [field]: value }));
  };

  const handleSubmit = async (e: FormEvent) => {
    e.preventDefault();
    setErrors({});
    setSubmitting(true);

    // Build the request object matching CreateOrderRequest type
    const request: CreateOrderRequest = {
      shippingAddress: shippingAddress as Address,
      ...(sameAsShipping ? {} : { billingAddress: billingAddress as Address }),
      ...(notes ? { notes } : {}),
    };

    // Validate the request using the generated Zod schema
    const validation = createOrderRequestSchema.safeParse(request);

    if (!validation.success) {
      // Map Zod errors to form fields
      const fieldErrors: Record<string, string> = {};
      validation.error.issues.forEach((issue) => {
        const path = issue.path.join(".");
        fieldErrors[path] = issue.message;
      });
      setErrors(fieldErrors);
      setSubmitting(false);
      return;
    }

    try {
      // The validated data matches CreateOrderRequest exactly
      const order = await createOrder(validation.data);
      onSuccess(order);
    } catch (err) {
      if (err instanceof ApiRequestError) {
        // Handle field-specific errors from the API
        if (err.error.details) {
          const fieldErrors: Record<string, string> = {};
          err.error.details.forEach((detail) => {
            fieldErrors[detail.field] = detail.message;
          });
          setErrors(fieldErrors);
        } else {
          setErrors({ _form: err.error.message });
        }
      } else {
        setErrors({ _form: "An unexpected error occurred" });
      }
    } finally {
      setSubmitting(false);
    }
  };

  return (
    <form onSubmit={handleSubmit} className="checkout-form">
      <h2>Checkout</h2>

      {errors._form && <div className="form-error">{errors._form}</div>}

      <fieldset>
        <legend>Shipping Address</legend>
        <AddressFields
          address={shippingAddress}
          onChange={(field, value) =>
            handleAddressChange(setShippingAddress, field, value)
          }
          errors={errors}
          prefix="shippingAddress"
        />
      </fieldset>

      <label className="checkbox-label">
        <input
          type="checkbox"
          checked={sameAsShipping}
          onChange={(e) => setSameAsShipping(e.target.checked)}
        />
        Billing address same as shipping
      </label>

      {!sameAsShipping && (
        <fieldset>
          <legend>Billing Address</legend>
          <AddressFields
            address={billingAddress}
            onChange={(field, value) =>
              handleAddressChange(setBillingAddress, field, value)
            }
            errors={errors}
            prefix="billingAddress"
          />
        </fieldset>
      )}

      <div className="form-field">
        <label htmlFor="notes">Order Notes (optional)</label>
        <textarea
          id="notes"
          value={notes}
          onChange={(e) => setNotes(e.target.value)}
          maxLength={500}
          placeholder="Special delivery instructions..."
        />
        {errors.notes && <span className="field-error">{errors.notes}</span>}
      </div>

      <div className="form-actions">
        <button type="button" onClick={onCancel} disabled={submitting}>
          Cancel
        </button>
        <button type="submit" disabled={submitting}>
          {submitting ? "Processing..." : "Place Order"}
        </button>
      </div>
    </form>
  );
}

// ============================================================================
// Address fields component
// ============================================================================

interface AddressFieldsProps {
  address: AddressFormState;
  onChange: (field: keyof AddressFormState, value: string) => void;
  errors: Record<string, string>;
  prefix: string;
}

function AddressFields({ address, onChange, errors, prefix }: AddressFieldsProps) {
  const getError = (field: string) => errors[`${prefix}.${field}`];

  return (
    <div className="address-fields">
      <div className="form-field">
        <label htmlFor={`${prefix}-line1`}>Address Line 1 *</label>
        <input
          id={`${prefix}-line1`}
          type="text"
          value={address.line1}
          onChange={(e) => onChange("line1", e.target.value)}
          required
        />
        {getError("line1") && (
          <span className="field-error">{getError("line1")}</span>
        )}
      </div>

      <div className="form-field">
        <label htmlFor={`${prefix}-line2`}>Address Line 2</label>
        <input
          id={`${prefix}-line2`}
          type="text"
          value={address.line2}
          onChange={(e) => onChange("line2", e.target.value)}
        />
      </div>

      <div className="form-row">
        <div className="form-field">
          <label htmlFor={`${prefix}-city`}>City *</label>
          <input
            id={`${prefix}-city`}
            type="text"
            value={address.city}
            onChange={(e) => onChange("city", e.target.value)}
            required
          />
          {getError("city") && (
            <span className="field-error">{getError("city")}</span>
          )}
        </div>

        <div className="form-field">
          <label htmlFor={`${prefix}-state`}>State</label>
          <input
            id={`${prefix}-state`}
            type="text"
            value={address.state}
            onChange={(e) => onChange("state", e.target.value)}
          />
        </div>
      </div>

      <div className="form-row">
        <div className="form-field">
          <label htmlFor={`${prefix}-country`}>Country Code *</label>
          <input
            id={`${prefix}-country`}
            type="text"
            value={address.country}
            onChange={(e) => onChange("country", e.target.value.toUpperCase())}
            placeholder="US"
            maxLength={2}
            pattern="[A-Z]{2}"
            required
          />
          {getError("country") && (
            <span className="field-error">{getError("country")}</span>
          )}
        </div>

        <div className="form-field">
          <label htmlFor={`${prefix}-postalCode`}>Postal Code *</label>
          <input
            id={`${prefix}-postalCode`}
            type="text"
            value={address.postalCode}
            onChange={(e) => onChange("postalCode", e.target.value)}
            required
          />
          {getError("postalCode") && (
            <span className="field-error">{getError("postalCode")}</span>
          )}
        </div>
      </div>
    </div>
  );
}

