import { apiClient } from './api-client';

// Currency symbols mapping
const CURRENCY_SYMBOLS: Record<string, string> = {
  NGN: '₦',
  USD: '$',
  EUR: '€',
  GBP: '£',
  JPY: '¥',
  CAD: 'C$',
  AUD: 'A$',
  CHF: 'CHF',
  CNY: '¥',
  INR: '₹',
  KES: 'KSh',
  ZAR: 'R',
  GHS: '₵',
  // Add more currencies as needed
};

// Cache for platform settings
let cachedSettings: { currency: string; symbol: string } | null = null;
let settingsPromise: Promise<{ currency: string; symbol: string }> | null = null;

// Get currency settings from platform settings
export async function getCurrencySettings(): Promise<{ currency: string; symbol: string }> {
  // Return cached settings if available
  if (cachedSettings) {
    return cachedSettings;
  }

  // Return existing promise if already loading
  if (settingsPromise) {
    return settingsPromise;
  }

  // Create new promise to load settings
  settingsPromise = (async () => {
    try {
      const settings = await apiClient.getPlatformSettings();
      const currency = settings.currency || 'NGN';
      const symbol = CURRENCY_SYMBOLS[currency] || currency;
      
      cachedSettings = { currency, symbol };
      return cachedSettings;
    } catch (error) {
      console.warn('Failed to load platform settings, using default currency:', error);
      // Fallback to NGN if settings can't be loaded
      cachedSettings = { currency: 'NGN', symbol: '₦' };
      return cachedSettings;
    } finally {
      settingsPromise = null;
    }
  })();

  return settingsPromise;
}

// Format currency amount with proper symbol
export async function formatCurrency(amount: number): Promise<string> {
  const { symbol } = await getCurrencySettings();
  return `${symbol}${amount.toLocaleString()}`;
}

// Format currency amount synchronously (for components that can't use async)
// This will use cached settings or fallback to NGN
export function formatCurrencySync(amount: number, fallbackCurrency = 'NGN'): string {
  const symbol = cachedSettings?.symbol || CURRENCY_SYMBOLS[fallbackCurrency] || fallbackCurrency;
  return `${symbol}${amount.toLocaleString()}`;
}

// Get currency symbol synchronously
export function getCurrencySymbolSync(fallbackCurrency = 'NGN'): string {
  return cachedSettings?.symbol || CURRENCY_SYMBOLS[fallbackCurrency] || fallbackCurrency;
}

// Clear cache (useful when settings are updated)
export function clearCurrencyCache(): void {
  cachedSettings = null;
  settingsPromise = null;
}

// Initialize currency settings (call this early in the app)
export async function initializeCurrency(): Promise<void> {
  try {
    await getCurrencySettings();
  } catch (error) {
    console.warn('Failed to initialize currency settings:', error);
  }
}

// Hook for React components to use currency
export function useCurrency() {
  const [currency, setCurrency] = React.useState<{ currency: string; symbol: string } | null>(cachedSettings);

  React.useEffect(() => {
    if (!currency) {
      getCurrencySettings().then(setCurrency);
    }
  }, [currency]);

  const formatAmount = React.useCallback((amount: number) => {
    if (currency) {
      return `${currency.symbol}${amount.toLocaleString()}`;
    }
    return formatCurrencySync(amount);
  }, [currency]);

  return {
    currency: currency?.currency || 'NGN',
    symbol: currency?.symbol || '₦',
    formatAmount,
  };
}

// Add React import for the hook
import React from 'react';
