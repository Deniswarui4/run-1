// Feature: guest-checkout-cart-sharing-metrics
// Pure helpers extracted from app/events/[id]/page.tsx so they can be unit/property tested
// without the Next.js runtime.

export const BASE_URL = process.env.NEXT_PUBLIC_BASE_URL || 'http://localhost:3000';
export const API_BASE_URL = process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080/api/v1';

export type EventStatus = 'draft' | 'pending' | 'approved' | 'published' | 'rejected' | 'cancelled' | 'completed';

export interface EventData {
  id: string;
  title: string;
  description: string;
  image_url?: string | null;
  status: EventStatus;
}

export interface OGMetadata {
  title: string;
  description: string;
  openGraph: {
    title: string;
    description: string;
    images: { url: string }[];
    url: string;
    type: string;
  };
  twitter: {
    card: string;
    title: string;
    description: string;
    images: string[];
  };
}

/** Returns null for non-published events (no share metadata exposed). */
export function buildOGMetadata(event: EventData): OGMetadata | null {
  if (event.status !== 'published') return null;

  const canonicalUrl = buildShareLink(event.id);
  const imageUrl = event.image_url
    ? `${API_BASE_URL.replace('/api/v1', '')}${event.image_url}`
    : `${BASE_URL}/og-default.png`;

  return {
    title: event.title,
    description: event.description,
    openGraph: {
      title: event.title,
      description: event.description,
      images: [{ url: imageUrl }],
      url: canonicalUrl,
      type: 'website',
    },
    twitter: {
      card: 'summary_large_image',
      title: event.title,
      description: event.description,
      images: [imageUrl],
    },
  };
}

/** Canonical share link for a published event. */
export function buildShareLink(eventId: string): string {
  return `${BASE_URL}/events/${eventId}`;
}

/** Social share URLs for Facebook, Twitter/X, and WhatsApp. */
export function buildSocialShareURLs(event: EventData): {
  facebook: string;
  twitter: string;
  whatsapp: string;
} {
  const link = buildShareLink(event.id);
  const encodedLink = encodeURIComponent(link);
  const encodedTitle = encodeURIComponent(event.title);

  return {
    facebook: `https://www.facebook.com/sharer/sharer.php?u=${encodedLink}`,
    twitter: `https://twitter.com/intent/tweet?url=${encodedLink}&text=${encodedTitle}`,
    whatsapp: `https://wa.me/?text=${encodedTitle}%20${encodedLink}`,
  };
}
