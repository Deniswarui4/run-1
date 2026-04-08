import { Metadata } from 'next';
import EventDetailsClient from './EventDetailsClient';

const API_BASE_URL = process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080/api/v1';
const BASE_URL = process.env.NEXT_PUBLIC_BASE_URL || 'http://localhost:3000';

interface PageProps {
  params: Promise<{ id: string }>;
}

async function fetchEvent(id: string) {
  try {
    const res = await fetch(`${API_BASE_URL}/events/${id}`, { next: { revalidate: 60 } });
    if (!res.ok) return null;
    return res.json();
  } catch {
    return null;
  }
}

export async function generateMetadata({ params }: PageProps): Promise<Metadata> {
  const { id } = await params;
  const event = await fetchEvent(id);

  // Guard: non-published events must not expose share metadata
  if (!event || event.status !== 'published') {
    return {};
  }

  const canonicalUrl = `${BASE_URL}/events/${event.id}`;
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

export default function EventDetailsPage() {
  return <EventDetailsClient />;
}
