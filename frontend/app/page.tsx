'use client';

import { useState, useEffect } from 'react';
import Link from 'next/link';
import { Button } from '@/components/ui/button';
import { Navbar } from '@/components/navbar';
import { Footer } from '@/components/footer';
import { FeaturedEvents } from '@/components/featured-events';
import { HeroBanner } from '@/components/hero-banner';
import { Calendar, Ticket, Users, Shield, ArrowRight, Music, Palette, Utensils, Gamepad2, Dumbbell, Briefcase } from 'lucide-react';
import {
  Carousel,
  CarouselContent,
  CarouselItem,
  CarouselNext,
  CarouselPrevious,
} from "@/components/ui/carousel";
import { apiClient } from '@/lib/api-client';
import { Category } from '@/lib/types';
import { Skeleton } from '@/components/ui/skeleton';

export default function Home() {
  const [categories, setCategories] = useState<Category[]>([]);
  const [loadingCategories, setLoadingCategories] = useState(true);

  useEffect(() => {
    loadCategories();
  }, []);

  const loadCategories = async () => {
    try {
      const data = await apiClient.getCategories();
      setCategories(data.filter(c => c.is_active));
    } catch (error) {
      console.error('Failed to load categories:', error);
    } finally {
      setLoadingCategories(false);
    }
  };

  // Icon mapping for categories
  const getCategoryIcon = (categoryName: string) => {
    const name = categoryName.toLowerCase();
    if (name.includes('music') || name.includes('concert')) return Music;
    if (name.includes('art') || name.includes('exhibition')) return Palette;
    if (name.includes('food') || name.includes('drink')) return Utensils;
    if (name.includes('game') || name.includes('gaming')) return Gamepad2;
    if (name.includes('sport') || name.includes('fitness')) return Dumbbell;
    if (name.includes('business') || name.includes('conference')) return Briefcase;
    return Calendar;
  };
  return (
    <div className="min-h-screen flex flex-col bg-background selection:bg-primary/20 selection:text-primary">
      <Navbar />

      <main className="flex-1">
        {/* Hero Section */}
        <HeroBanner />

        {/* Categories Section */}
        <section className="py-24 bg-gradient-to-b from-background to-secondary/20 relative mt-12 md:mt-20">
          <div className="container mx-auto px-4 z-10 relative">
            <div className="text-center mb-16">
              <h2 className="text-4xl font-extrabold tracking-tight mb-4 text-foreground">Browse by Category</h2>
              <p className="text-muted-foreground text-xl">Find experiences uniquely curated for you</p>
            </div>

            {loadingCategories ? (
              <div className="flex gap-4 overflow-hidden">
                {[...Array(6)].map((_, i) => (
                  <Skeleton key={i} className="h-32 w-40 flex-shrink-0 rounded-2xl" />
                ))}
              </div>
            ) : (
              <div className="relative px-12">
                <Carousel
                  opts={{
                    align: "start",
                    loop: false,
                  }}
                  className="w-full"
                >
                  <CarouselContent className="-ml-4">
                    {categories.map((category) => {
                      const IconComponent = getCategoryIcon(category.name);
                      return (
                        <CarouselItem key={category.id} className="pl-4 basis-1/2 md:basis-1/4 lg:basis-1/6">
                          <Link
                            href={`/events?category=${encodeURIComponent(category.name)}`}
                            className="group block"
                          >
                            <div
                              className="relative h-32 rounded-2xl p-4 flex flex-col items-center justify-center text-center border border-border/50 hover:border-primary/50 transition-all duration-300 overflow-hidden hover:shadow-lg hover:-translate-y-1 cursor-pointer"
                              style={{
                                background: `linear-gradient(135deg, ${category.color || '#3B82F6'}15 0%, ${category.color || '#3B82F6'}05 100%)`
                              }}
                            >
                              {/* Decorative circle */}
                              <div
                                className="absolute top-0 right-0 w-16 h-16 rounded-full -mr-6 -mt-6 opacity-20 group-hover:opacity-30 transition-opacity"
                                style={{ backgroundColor: category.color || '#3B82F6' }}
                              />

                              {/* Icon */}
                              <div
                                className="w-12 h-12 rounded-full flex items-center justify-center mb-2 group-hover:scale-110 transition-transform"
                                style={{
                                  backgroundColor: `${category.color || '#3B82F6'}20`,
                                  color: category.color || '#3B82F6'
                                }}
                              >
                                <IconComponent className="h-6 w-6" />
                              </div>

                              {/* Category name */}
                              <h3
                                className="font-semibold text-sm group-hover:text-primary transition-colors line-clamp-2"
                              >
                                {category.name}
                              </h3>
                            </div>
                          </Link>
                        </CarouselItem>
                      );
                    })}
                  </CarouselContent>
                  <CarouselPrevious className="hidden md:flex -left-4" />
                  <CarouselNext className="hidden md:flex -right-4" />
                </Carousel>
              </div>
            )}
          </div>
        </section>

        {/* Featured Events Section */}
        <section className="py-10 relative">
          <div className="container mx-auto px-4">
            <div className="flex items-center justify-between mb-10">
              <div>
                <h2 className="text-3xl font-bold mb-2">Trending Events</h2>
                <p className="text-muted-foreground">Don&apos;t miss out on these popular experiences</p>
              </div>
              <Link href="/events">
                <Button variant="ghost" className="text-primary hover:bg-primary/10 group">
                  View All
                  <ArrowRight className="ml-2 h-4 w-4 transition-transform group-hover:translate-x-1" />
                </Button>
              </Link>
            </div>
            <FeaturedEvents />
          </div>
        </section>

        {/* Features Section */}
        <section className="py-24 bg-secondary/30 relative overflow-hidden">
          <div className="container mx-auto px-4">
            <div className="text-center max-w-2xl mx-auto mb-16">
              <h2 className="text-3xl md:text-4xl font-bold mb-4">
                Why Choose Runtown?
              </h2>
              <p className="text-muted-foreground text-lg">
                We provide the best tools and experience for both organizers and attendees.
              </p>
            </div>

            <div className="grid md:grid-cols-2 lg:grid-cols-4 gap-8">
              {[
                {
                  icon: Calendar,
                  title: "Wide Selection",
                  desc: "Browse thousands of events across multiple categories and locations."
                },
                {
                  icon: Ticket,
                  title: "Easy Booking",
                  desc: "Quick and secure ticket purchasing with instant digital delivery."
                },
                {
                  icon: Users,
                  title: "For Organizers",
                  desc: "Powerful tools to create, manage, and promote your events effectively."
                },
                {
                  icon: Shield,
                  title: "Secure Platform",
                  desc: "Bank-grade security for payments and verified events for peace of mind."
                }
              ].map((feature, index) => (
                <div key={index} className="group p-8 rounded-2xl bg-background border border-border/50 hover:border-primary/20 hover:shadow-lg hover:shadow-primary/5 transition-all duration-300 relative overflow-hidden">
                  <div className="absolute top-0 right-0 w-24 h-24 bg-primary/5 rounded-bl-full -mr-4 -mt-4 transition-transform group-hover:scale-110" />

                  <div className="inline-flex items-center justify-center w-16 h-16 rounded-2xl bg-primary/10 mb-6 group-hover:bg-gradient-to-br group-hover:from-primary group-hover:to-purple-600 group-hover:text-white transition-all duration-300 shadow-sm">
                    <feature.icon className="h-8 w-8 text-primary group-hover:text-white transition-colors" />
                  </div>

                  <h3 className="text-2xl font-bold mb-3 group-hover:text-primary transition-colors">{feature.title}</h3>
                  <p className="text-muted-foreground leading-relaxed">
                    {feature.desc}
                  </p>
                </div>
              ))}
            </div>
          </div>
        </section>

        {/* CTA Section */}
        <section className="py-24 relative overflow-hidden">
          <div className="absolute inset-0 bg-primary/5" />
          <div className="container mx-auto px-4 relative">
            <div className="bg-gradient-to-r from-primary to-blue-600 rounded-3xl p-12 md:p-20 text-center text-white shadow-2xl shadow-primary/25 relative overflow-hidden">
              <div className="absolute top-0 left-0 w-full h-full bg-[url('/noise.png')] opacity-10 mix-blend-overlay" />
              <div className="absolute -top-24 -left-24 w-64 h-64 bg-white/10 rounded-full blur-3xl" />
              <div className="absolute -bottom-24 -right-24 w-64 h-64 bg-black/10 rounded-full blur-3xl" />

              <div className="relative z-10 max-w-3xl mx-auto">
                <h2 className="text-3xl md:text-5xl font-bold mb-6">
                  Ready to Host Your Own Event?
                </h2>
                <p className="text-xl mb-10 text-blue-50 leading-relaxed">
                  Join thousands of successful organizers. Create, manage, and sell tickets for your events with ease.
                </p>
                <div className="flex flex-col sm:flex-row gap-4 justify-center">
                  <Link href="/register">
                    <Button size="lg" variant="secondary" className="h-14 px-8 text-lg rounded-full shadow-lg hover:shadow-xl transition-all font-semibold text-primary">
                      Become an Organizer
                    </Button>
                  </Link>
                  <Link href="/contact">
                    <Button size="lg" variant="outline" className="h-14 px-8 text-lg rounded-full border-white/30 text-white hover:bg-white/10 hover:text-white hover:border-white/50 transition-all bg-transparent backdrop-blur-sm">
                      Contact Sales
                    </Button>
                  </Link>
                </div>
              </div>
            </div>
          </div>
        </section>
      </main>

      <Footer />
    </div>
  );
}
