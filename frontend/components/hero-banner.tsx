'use client';

import { useState, useEffect } from 'react';
import Link from 'next/link';
import { Button } from '@/components/ui/button';
import { Calendar, MapPin, Search } from 'lucide-react';
import { apiClient } from '@/lib/api-client';
import { Event } from '@/lib/types';
import {
  Carousel,
  CarouselContent,
  CarouselItem,
  CarouselNext,
  CarouselPrevious,
} from "@/components/ui/carousel";
import Autoplay from "embla-carousel-autoplay";

export function HeroBanner() {
  const [featuredEvents, setFeaturedEvents] = useState<Event[]>([]);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    loadFeaturedEvents();
  }, []);

  const loadFeaturedEvents = async () => {
    try {
      const data = await apiClient.getFeaturedEvents();
      setFeaturedEvents(data || []);
    } catch (error) {
      console.error('Failed to load featured events:', error);
    } finally {
      setLoading(false);
    }
  };

  const SearchBar = () => (
    <div className="max-w-5xl mx-auto bg-card/95 backdrop-blur border border-border/50 rounded-full shadow-2xl p-2 flex flex-col md:flex-row items-center gap-2 mt-8 z-20 relative">
      <div className="flex-1 w-full md:w-auto relative group">
        <div className="absolute left-4 top-1/2 -translate-y-1/2 text-muted-foreground group-focus-within:text-primary transition-colors">
          <MapPin className="h-5 w-5" />
        </div>
        <div className="flex flex-col px-12 py-2 w-full hover:bg-secondary/50 rounded-full transition-colors cursor-pointer">
          <span className="text-xs font-bold text-foreground">Location</span>
          <input
            type="text"
            placeholder="Where to?"
            className="bg-transparent border-none p-0 text-sm focus:ring-0 placeholder:text-muted-foreground/70"
          />
        </div>
      </div>

      <div className="w-px h-8 bg-border hidden md:block" />

      <div className="flex-1 w-full md:w-auto relative group">
        <div className="absolute left-4 top-1/2 -translate-y-1/2 text-muted-foreground group-focus-within:text-primary transition-colors">
          <Calendar className="h-5 w-5" />
        </div>
        <div className="flex flex-col px-12 py-2 w-full hover:bg-secondary/50 rounded-full transition-colors cursor-pointer">
          <span className="text-xs font-bold text-foreground">Date</span>
          <input
            type="text"
            placeholder="Pick dates"
            className="bg-transparent border-none p-0 text-sm focus:ring-0 placeholder:text-muted-foreground/70"
          />
        </div>
      </div>

      <Button size="lg" className="rounded-full h-14 w-full md:w-auto px-8 bg-primary hover:bg-primary/90 text-primary-foreground font-bold shadow-md lg:ml-4">
        <Search className="h-5 w-5 mr-2" />
        Search
      </Button>
    </div>
  );

  if (loading || featuredEvents.length === 0) {
    return (
      <section className="relative py-24 lg:py-32 bg-gradient-to-b from-primary/10 via-background to-background overflow-hidden">
        <div className="absolute inset-0 bg-[url('/noise.png')] opacity-20 mix-blend-overlay pointer-events-none" />
        <div className="container mx-auto px-4 relative z-10">
          <div className="flex flex-col items-center text-center mb-8">
             <div className="inline-block px-4 py-1.5 rounded-full bg-primary/10 text-primary font-medium text-sm mb-6 border border-primary/20 backdrop-blur-sm">
                🎉 Discover amazing events near you
             </div>
            <h1 className="text-5xl md:text-7xl font-extrabold tracking-tight text-foreground mb-6 drop-shadow-sm">
              Find your next <span className="text-transparent bg-clip-text bg-gradient-to-r from-primary to-purple-600">experience</span>
            </h1>
            <p className="text-xl text-muted-foreground max-w-2xl font-medium">
              Discover concerts, festivals, and events happening around you.
            </p>
          </div>
          <SearchBar />
        </div>
      </section>
    );
  }

  return (
    <section className="relative w-full">
      <Carousel
        plugins={[
          Autoplay({
            delay: 5000,
          }),
        ]}
        opts={{
          loop: true,
          align: 'start',
        }}
        className="w-full"
      >
        <CarouselContent>
          {featuredEvents.map((event) => (
            <CarouselItem key={event.id} className="relative w-full h-[600px] md:h-[70vh]">
              <div className="absolute inset-0 bg-black/60 z-10" />
              {event.image_url ? (
                <img
                  src={`${process.env.NEXT_PUBLIC_API_URL?.replace('/api/v1', '') || 'http://localhost:8080'}${event.image_url}`}
                  alt={event.title}
                  className="absolute inset-0 w-full h-full object-cover"
                />
              ) : (
                <div className="absolute inset-0 w-full h-full bg-gradient-to-r from-primary/80 to-purple-800" />
              )}
              
              <div className="absolute inset-0 z-20 flex flex-col justify-center items-center text-center container mx-auto px-4">
                <span className="px-4 py-1.5 rounded-full bg-white/20 text-white backdrop-blur-md font-semibold text-sm mb-6 tracking-wider uppercase border border-white/30 shadow-lg">
                  Featured Event
                </span>
                <h1 className="text-5xl md:text-7xl font-extrabold text-white mb-6 drop-shadow-xl max-w-4xl tracking-tight">
                  {event.title}
                </h1>
                <p className="text-xl md:text-2xl text-white/90 mb-10 max-w-2xl font-medium drop-shadow-md line-clamp-2">
                  {event.venue}, {event.city}
                </p>
                <Link href={`/events/${event.id}`}>
                  <Button size="lg" className="rounded-full h-14 px-10 bg-white text-primary hover:bg-white/90 font-bold text-lg shadow-2xl transition-transform hover:scale-105">
                    Get Tickets Now
                  </Button>
                </Link>
              </div>
            </CarouselItem>
          ))}
        </CarouselContent>
        <div className="absolute bottom-8 right-8 z-30 flex gap-4 hidden md:flex">
          <CarouselPrevious className="relative inset-0 translate-y-0 h-12 w-12 bg-white/10 text-white border-white/20 hover:bg-white/30 backdrop-blur" />
          <CarouselNext className="relative inset-0 translate-y-0 h-12 w-12 bg-white/10 text-white border-white/20 hover:bg-white/30 backdrop-blur" />
        </div>
      </Carousel>
      
      {/* Search Bar positioned overlaying the bottom of the hero */}
      <div className="absolute bottom-0 left-0 right-0 transform translate-y-1/2 z-40 px-4">
         <SearchBar />
      </div>
      
      {/* Add padding below the hero section to account for the overlapping search bar */}
      <div className="h-16 md:h-12 bg-transparent" />
    </section>
  );
}
