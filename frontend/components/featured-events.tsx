'use client';

import { useState, useEffect } from 'react';
import Link from 'next/link';
import { apiClient } from '@/lib/api-client';
import { Event } from '@/lib/types';
import { useCurrency } from '@/lib/currency';
import { Button } from '@/components/ui/button';
import { Skeleton } from '@/components/ui/skeleton';
import {
  Carousel,
  CarouselContent,
  CarouselItem,
  CarouselNext,
  CarouselPrevious,
} from '@/components/ui/carousel';
import { Star, Users } from 'lucide-react';

import { EventCard } from '@/components/event-card';

export function FeaturedEvents() {
  const { formatAmount } = useCurrency();
  const [events, setEvents] = useState<Event[]>([]);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    loadFeaturedEvents();
  }, []);

  const loadFeaturedEvents = async () => {
    try {
      setLoading(true);
      const data = await apiClient.getFeaturedEvents();
      setEvents(Array.isArray(data) ? data : []);
    } catch (error) {
      console.error('Failed to load featured events:', error);
      setEvents([]);
    } finally {
      setLoading(false);
    }
  };

  const formatDate = (dateString: string) => {
    return new Date(dateString).toLocaleDateString('en-US', {
      month: 'short',
      day: 'numeric',
      year: 'numeric',
    });
  };

  const formatTime = (dateString: string) => {
    return new Date(dateString).toLocaleTimeString('en-US', {
      hour: '2-digit',
      minute: '2-digit',
    });
  };

  const getMinPrice = (event: Event) => {
    if (!event.ticket_types || event.ticket_types.length === 0) return null;
    return Math.min(...event.ticket_types.map(t => t.price));
  };

  const isUpcoming = (dateString: string) => {
    // An event is "upcoming" (not past) if it hasn't ended yet
    // We should check the end_date, not start_date
    return new Date(dateString) > new Date();
  };

  if (loading) {
    return (
      <section className="py-16 bg-gradient-to-br from-primary/5 via-background to-secondary/5">
        <div className="container mx-auto px-4">
          <div className="text-center mb-12">
            <Skeleton className="h-8 w-64 mx-auto mb-4" />
            <Skeleton className="h-4 w-96 mx-auto" />
          </div>
          <div className="grid md:grid-cols-2 lg:grid-cols-3 gap-6">
            {[...Array(3)].map((_, i) => (
              <Skeleton key={i} className="h-80 w-full" />
            ))}
          </div>
        </div>
      </section>
    );
  }

  if (events.length === 0) {
    return null; // Don't show section if no featured events
  }

  return (
    <section className="py-16 bg-gradient-to-br from-primary/5 via-background to-secondary/5">
      <div className="container mx-auto px-4">
        {/* Section Header */}
        <div className="text-center mb-12">
          <div className="flex items-center justify-center gap-2 mb-4">
            <Star className="h-6 w-6 text-yellow-500 fill-current" />
            <h2 className="text-3xl font-bold">Featured Events</h2>
            <Star className="h-6 w-6 text-yellow-500 fill-current" />
          </div>
          <p className="text-muted-foreground text-lg max-w-2xl mx-auto">
            Don&#39;t miss out on these handpicked amazing events happening near you
          </p>
        </div>

        {/* Carousel for larger screens, Grid for smaller */}
        <div className="hidden lg:block">
          <Carousel
            opts={{
              align: "start",
              loop: true,
              skipSnaps: false,
              dragFree: true,
            }}
            className="w-full max-w-7xl mx-auto"
          >
            <CarouselContent>
              {events.map((event) => (
                <CarouselItem key={event.id} className="md:basis-1/2 lg:basis-1/3">
                  <div className="p-1">
                    <EventCard
                      event={event}
                      formatAmount={formatAmount}
                      formatDate={formatDate}
                      formatTime={formatTime}
                      getMinPrice={getMinPrice}
                      isUpcoming={isUpcoming}
                    />
                  </div>
                </CarouselItem>
              ))}
            </CarouselContent>
            <CarouselPrevious />
            <CarouselNext />
          </Carousel>
        </div>

        {/* Grid for smaller screens */}
        <div className="lg:hidden grid md:grid-cols-2 gap-6">
          {events.slice(0, 4).map((event) => (
            <EventCard
              key={event.id}
              event={event}
              formatAmount={formatAmount}
              formatDate={formatDate}
              formatTime={formatTime}
              getMinPrice={getMinPrice}
              isUpcoming={isUpcoming}
            />
          ))}
        </div>

        {/* View All Button */}
        <div className="text-center mt-12">
          <Link href="/events">
            <Button size="lg" variant="outline" className="group">
              View All Events
              <Users className="ml-2 h-4 w-4 group-hover:translate-x-1 transition-transform" />
            </Button>
          </Link>
        </div>
      </div>
    </section>
  );
}
