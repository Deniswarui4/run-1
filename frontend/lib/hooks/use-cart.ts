'use client';

import { useState, useEffect, useRef, useCallback } from 'react';
import { toast } from 'sonner';
import { apiClient } from '@/lib/api-client';
import { CartItem, TicketType } from '@/lib/types';

interface UseCartReturn {
  cart: CartItem[];
  addToCart: (ticketType: TicketType) => void;
  removeFromCart: (ticketTypeId: string) => void;
  updateQuantity: (ticketTypeId: string, quantity: number) => void;
  clearCart: () => void;
  syncing: boolean;
}

export function useCart(eventId: string | null): UseCartReturn {
  const [cart, setCart] = useState<CartItem[]>([]);
  const [syncing, setSyncing] = useState(false);
  // Track whether the current cart state was loaded from server (skip re-sync on initial load)
  const isInitialLoad = useRef(true);
  const debounceTimer = useRef<ReturnType<typeof setTimeout> | null>(null);

  // Restore cart on mount
  useEffect(() => {
    if (!eventId) return;

    const loadCart = async () => {
      try {
        const draftCart = await apiClient.getCart(eventId);
        if (draftCart && draftCart.items && draftCart.items.length > 0) {
          const restored: CartItem[] = draftCart.items
            .filter(item => item.ticket_type)
            .map(item => ({
              ticket_type_id: item.ticket_type_id,
              quantity: item.quantity,
              ticket_type: item.ticket_type!,
            }));
          isInitialLoad.current = true;
          setCart(restored);
        }
      } catch {
        // Silently ignore — no cart to restore
      }
    };

    loadCart();
  }, [eventId]);

  // Sync cart to server (debounced)
  useEffect(() => {
    // Skip sync on initial load from server
    if (isInitialLoad.current) {
      isInitialLoad.current = false;
      return;
    }

    if (!eventId) return;

    if (debounceTimer.current) {
      clearTimeout(debounceTimer.current);
    }

    debounceTimer.current = setTimeout(async () => {
      setSyncing(true);
      const items = cart.map(item => ({
        ticket_type_id: item.ticket_type_id,
        quantity: item.quantity,
      }));

      try {
        await apiClient.upsertCart(eventId, items);
      } catch {
        // Retry once after 1s
        await new Promise(resolve => setTimeout(resolve, 1000));
        try {
          await apiClient.upsertCart(eventId, items);
        } catch {
          toast.error(
            "Could not save your cart. Your selections are still visible but may not persist."
          );
        }
      } finally {
        setSyncing(false);
      }
    }, 500);

    return () => {
      if (debounceTimer.current) {
        clearTimeout(debounceTimer.current);
      }
    };
  }, [cart, eventId]);

  const addToCart = useCallback((ticketType: TicketType) => {
    setCart(prev => {
      const existing = prev.find(item => item.ticket_type_id === ticketType.id);
      if (existing) {
        return prev.map(item =>
          item.ticket_type_id === ticketType.id
            ? { ...item, quantity: item.quantity + 1 }
            : item
        );
      }
      return [...prev, { ticket_type_id: ticketType.id, quantity: 1, ticket_type: ticketType }];
    });
  }, []);

  const removeFromCart = useCallback((ticketTypeId: string) => {
    setCart(prev => prev.filter(item => item.ticket_type_id !== ticketTypeId));
    toast.success('Removed from cart');
  }, []);

  const updateQuantity = useCallback((ticketTypeId: string, quantity: number) => {
    if (quantity <= 0) {
      setCart(prev => prev.filter(item => item.ticket_type_id !== ticketTypeId));
      toast.success('Removed from cart');
    } else {
      setCart(prev =>
        prev.map(item =>
          item.ticket_type_id === ticketTypeId ? { ...item, quantity } : item
        )
      );
    }
  }, []);

  const clearCart = useCallback(async () => {
    setCart([]);
    if (eventId) {
      try {
        await apiClient.deleteCart(eventId);
      } catch {
        // Silently ignore — local state is already cleared
      }
    }
  }, [eventId]);

  return { cart, addToCart, removeFromCart, updateQuantity, clearCart, syncing };
}
