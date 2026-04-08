'use client';

import { useState } from 'react';
import { apiClient } from '@/lib/api-client';
import { CartItem } from '@/lib/types';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';

// Simplified RFC 5322-compatible client-side email regex
const EMAIL_REGEX = /^[^\s@]+@[^\s@]+\.[^\s@]+$/;

interface GuestCheckoutFormProps {
  eventId: string;
  cart: CartItem[];
  onSuccess?: () => void;
}

export function GuestCheckoutForm({ eventId, cart, onSuccess }: GuestCheckoutFormProps) {
  const [email, setEmail] = useState('');
  const [emailError, setEmailError] = useState('');
  const [loading, setLoading] = useState(false);

  const validateEmail = (value: string): boolean => {
    if (!value) {
      setEmailError('Email is required');
      return false;
    }
    if (!EMAIL_REGEX.test(value)) {
      setEmailError('Please enter a valid email address');
      return false;
    }
    setEmailError('');
    return true;
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!validateEmail(email)) return;
    if (cart.length === 0) return;

    setLoading(true);
    try {
      const response = await apiClient.initiateGuestCheckout({
        event_id: eventId,
        email,
        items: cart.map(item => ({
          ticket_type_id: item.ticket_type_id,
          quantity: item.quantity,
        })),
      });
      onSuccess?.();
      window.location.href = response.authorization_url;
    } catch (err) {
      setEmailError(err instanceof Error ? err.message : 'Checkout failed. Please try again.');
    } finally {
      setLoading(false);
    }
  };

  return (
    <form onSubmit={handleSubmit} className="space-y-3">
      <div className="space-y-1">
        <Label htmlFor="guest-email" className="text-sm font-medium">
          Email address
        </Label>
        <Input
          id="guest-email"
          type="email"
          placeholder="you@example.com"
          value={email}
          onChange={e => {
            setEmail(e.target.value);
            if (emailError) validateEmail(e.target.value);
          }}
          onBlur={() => validateEmail(email)}
          disabled={loading}
          aria-describedby={emailError ? 'guest-email-error' : undefined}
          aria-invalid={!!emailError}
        />
        {emailError && (
          <p id="guest-email-error" className="text-xs text-destructive">
            {emailError}
          </p>
        )}
      </div>
      <Button
        type="submit"
        className="w-full h-12 text-base font-semibold"
        disabled={loading || cart.length === 0}
      >
        {loading ? 'Processing...' : 'Continue as Guest'}
      </Button>
    </form>
  );
}
