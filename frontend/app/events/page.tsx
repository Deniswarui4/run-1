'use client';

import { useState, useEffect, Suspense } from 'react';
import { useSearchParams } from 'next/navigation';
import { apiClient } from '@/lib/api-client';
import { Event, Category } from '@/lib/types';
import { Navbar } from '@/components/navbar';
import { Button } from '@/components/ui/button';
import { Skeleton } from '@/components/ui/skeleton';
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select';
import {
  Pagination,
  PaginationContent,
  PaginationEllipsis,
  PaginationItem,
  PaginationLink,
  PaginationNext,
  PaginationPrevious,
} from "@/components/ui/pagination";
import { Search, Star, X } from 'lucide-react';
import { EventCard } from '@/components/event-card';
import { FeaturedEventsBadge } from '@/components/featured-badge';
import { toast } from 'sonner';
import { useCurrency } from '@/lib/currency';

function EventsContent() {
  const { formatAmount } = useCurrency();
  const searchParams = useSearchParams();
  const [events, setEvents] = useState<Event[]>([]);
  const [featuredEvents, setFeaturedEvents] = useState<Event[]>([]);
  const [categories, setCategories] = useState<Category[]>([]);
  const [loading, setLoading] = useState(true);
  const [loadingFeatured, setLoadingFeatured] = useState(true);
  const [loadingCategories, setLoadingCategories] = useState(true);
  const [searchTerm, setSearchTerm] = useState('');
  const [categoryFilter, setCategoryFilter] = useState('');
  const [cityFilter, setCityFilter] = useState('');
  const [initialized, setInitialized] = useState(false);

  // Pagination state
  const [currentPage, setCurrentPage] = useState(1);
  const [itemsPerPage] = useState(12);

  // Initialize filters from URL parameters
  useEffect(() => {
    const categoryFromUrl = searchParams.get('category');
    if (categoryFromUrl) {
      setCategoryFilter(categoryFromUrl);
    }
    setInitialized(true);
  }, [searchParams]);

  useEffect(() => {
    if (!initialized) return;

    loadCategories();
    loadFeaturedEvents();

    // If category filter is set from URL, search with it
    if (categoryFilter) {
      handleSearch();
    } else {
      loadEvents();
    }
  }, [initialized]);

  const loadEvents = async () => {
    try {
      setLoading(true);
      const data = await apiClient.getEvents();
      setEvents(data);
      setCurrentPage(1); // Reset to first page when loading new data
    } catch (error) {
      toast.error('Failed to load events');
      console.error(error);
    } finally {
      setLoading(false);
    }
  };

  const loadFeaturedEvents = async () => {
    try {
      setLoadingFeatured(true);
      const data = await apiClient.getFeaturedEvents();
      // Sort by start_date ascending (Requirement 5.3 / 5.8)
      const sorted = Array.isArray(data)
        ? [...data].sort((a, b) => new Date(a.start_date).getTime() - new Date(b.start_date).getTime())
        : [];
      setFeaturedEvents(sorted);
    } catch (error) {
      console.error('Failed to load featured events:', error);
      setFeaturedEvents([]);
    } finally {
      setLoadingFeatured(false);
    }
  };

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

  const handleSearch = async () => {
    try {
      setLoading(true);
      const data = await apiClient.getEvents({
        search: searchTerm || undefined,
        category: categoryFilter || undefined,
        city: cityFilter || undefined,
      });
      setEvents(data);
      setCurrentPage(1); // Reset to first page on search
    } catch (error) {
      toast.error('Search failed');
      console.error(error);
    } finally {
      setLoading(false);
    }
  };

  const formatDate = (dateString: string) => {
    return new Date(dateString).toLocaleDateString('en-US', {
      weekday: 'short',
      year: 'numeric',
      month: 'short',
      day: 'numeric',
    });
  };

  const formatTime = (dateString: string) => {
    return new Date(dateString).toLocaleTimeString('en-US', {
      hour: 'numeric',
      minute: '2-digit',
    });
  };

  const getMinPrice = (event: Event) => {
    if (!event.ticket_types || event.ticket_types.length === 0) return null;
    const prices = event.ticket_types.map(t => t.price);
    return Math.min(...prices);
  };

  // Pagination calculations
  const totalPages = Math.ceil(events.length / itemsPerPage);
  const startIndex = (currentPage - 1) * itemsPerPage;
  const endIndex = startIndex + itemsPerPage;
  const currentEvents = events.slice(startIndex, endIndex);

  const goToPage = (page: number) => {
    setCurrentPage(page);
    window.scrollTo({ top: 0, behavior: 'smooth' });
  };

  // Generate page numbers to display
  const getPageNumbers = () => {
    const pages = [];
    const maxVisiblePages = 5;

    if (totalPages <= maxVisiblePages) {
      for (let i = 1; i <= totalPages; i++) {
        pages.push(i);
      }
    } else {
      if (currentPage <= 3) {
        for (let i = 1; i <= 4; i++) pages.push(i);
        pages.push('ellipsis');
        pages.push(totalPages);
      } else if (currentPage >= totalPages - 2) {
        pages.push(1);
        pages.push('ellipsis');
        for (let i = totalPages - 3; i <= totalPages; i++) pages.push(i);
      } else {
        pages.push(1);
        pages.push('ellipsis');
        for (let i = currentPage - 1; i <= currentPage + 1; i++) pages.push(i);
        pages.push('ellipsis');
        pages.push(totalPages);
      }
    }

    return pages;
  };

  return (
    <div className="min-h-screen flex flex-col bg-background">
      <Navbar />

      <main className="flex-1 container mx-auto px-4 py-8 pt-24">
        <div className="mb-12">
          <h1 className="text-4xl font-bold mb-6 text-center">Browse Events</h1>

          {/* Search and Filters - Airbnb Style */}
          <div className="max-w-4xl mx-auto bg-card border border-border/50 rounded-full shadow-lg p-2 flex flex-col md:flex-row items-center gap-2">
            {/* Search Input */}
            <div className="flex-1 w-full md:w-auto relative group px-4 py-2 hover:bg-secondary/50 rounded-full transition-colors">
              <div className="flex items-center gap-3">
                <Search className="h-5 w-5 text-primary" />
                <div className="flex flex-col flex-1">
                  <span className="text-xs font-bold text-foreground">Search</span>
                  <input
                    type="text"
                    placeholder="Search events..."
                    value={searchTerm}
                    onChange={(e) => setSearchTerm(e.target.value)}
                    onKeyDown={(e) => e.key === 'Enter' && handleSearch()}
                    className="bg-transparent border-none p-0 text-sm focus:ring-0 placeholder:text-muted-foreground/70 w-full focus:outline-none"
                  />
                </div>
              </div>
            </div>

            <div className="h-8 w-[1px] bg-border hidden md:block" />

            {/* Category Select */}
            <div className="flex-1 w-full md:w-auto relative group px-4 py-2 hover:bg-secondary/50 rounded-full transition-colors">
              <div className="flex flex-col w-full">
                <span className="text-xs font-bold text-foreground px-1">Category</span>
                {loadingCategories ? (
                  <div className="h-5 w-24 bg-muted animate-pulse rounded" />
                ) : (
                  <Select
                    value={categoryFilter}
                    onValueChange={(value) => setCategoryFilter(value === 'all' ? '' : value)}
                  >
                    <SelectTrigger className="border-none bg-transparent h-auto p-0 focus:ring-0 shadow-none text-sm text-muted-foreground font-normal">
                      <SelectValue placeholder="All Categories" />
                    </SelectTrigger>
                    <SelectContent>
                      <SelectItem value="all">All Categories</SelectItem>
                      {categories.map((cat) => (
                        <SelectItem key={cat.id} value={cat.name}>
                          <div className="flex items-center gap-2">
                            <div
                              className="w-2 h-2 rounded-full"
                              style={{ backgroundColor: cat.color || '#3B82F6' }}
                            />
                            <span>{cat.name}</span>
                          </div>
                        </SelectItem>
                      ))}
                    </SelectContent>
                  </Select>
                )}
              </div>
            </div>

            <div className="h-8 w-[1px] bg-border hidden md:block" />

            {/* City Input */}
            <div className="flex-1 w-full md:w-auto relative group px-4 py-2 hover:bg-secondary/50 rounded-full transition-colors">
              <div className="flex flex-col w-full">
                <span className="text-xs font-bold text-foreground">City</span>
                <input
                  type="text"
                  placeholder="Anywhere"
                  value={cityFilter}
                  onChange={(e) => setCityFilter(e.target.value)}
                  className="bg-transparent border-none p-0 text-sm focus:ring-0 placeholder:text-muted-foreground/70 w-full focus:outline-none"
                />
              </div>
            </div>

            {/* Search Button */}
            <Button
              onClick={handleSearch}
              size="lg"
              className="rounded-full w-full md:w-auto px-8 shadow-lg shadow-primary/25"
            >
              Search
            </Button>
          </div>

          {/* Active Filters Clear */}
          {(searchTerm || categoryFilter || cityFilter) && (
            <div className="flex justify-center mt-4">
              <Button
                variant="ghost"
                size="sm"
                onClick={() => {
                  setSearchTerm('');
                  setCategoryFilter('');
                  setCityFilter('');
                  loadEvents();
                }}
                className="text-muted-foreground hover:text-destructive"
              >
                <X className="h-4 w-4 mr-2" />
                Clear Filters
              </Button>
            </div>
          )}
        </div>

        {/* Featured Events Section */}
        {loadingFeatured ? (
          <div className="mb-12">
            <div className="flex items-center gap-2 mb-6">
              <Star className="h-5 w-5 text-yellow-500 fill-current" />
              <h2 className="text-2xl font-bold">Featured Events</h2>
            </div>
            <div className="grid md:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4 gap-6">
              {[...Array(4)].map((_, i) => (
                <Skeleton key={i} className="h-64 w-full rounded-xl" />
              ))}
            </div>
          </div>
        ) : featuredEvents.length > 0 ? (
          <div className="mb-12">
            <div className="flex items-center gap-2 mb-6">
              <Star className="h-5 w-5 text-yellow-500 fill-current" />
              <h2 className="text-2xl font-bold">Featured Events</h2>
            </div>
            <div className="grid md:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4 gap-8">
              {featuredEvents.map((event) => (
                <div key={event.id} className="relative">
                  <div className="absolute top-3 left-3 z-10">
                    <FeaturedEventsBadge />
                  </div>
                  <EventCard
                    event={event}
                    formatAmount={formatAmount}
                    formatDate={formatDate}
                    formatTime={formatTime}
                    getMinPrice={getMinPrice}
                    isUpcoming={(date) => new Date(date) > new Date()}
                  />
                </div>
              ))}
            </div>
          </div>
        ) : null}

        {/* Events Grid */}
        {loading ? (
          <div className="grid md:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4 gap-6">
            {[...Array(8)].map((_, i) => (
              <div key={i} className="space-y-3">
                <Skeleton className="h-48 w-full rounded-xl" />
                <Skeleton className="h-4 w-3/4" />
                <Skeleton className="h-4 w-1/2" />
              </div>
            ))}
          </div>
        ) : events.length === 0 ? (
          <div className="text-center py-24">
            <div className="bg-muted/30 rounded-full h-20 w-20 flex items-center justify-center mx-auto mb-4">
              <Search className="h-10 w-10 text-muted-foreground/50" />
            </div>
            <h3 className="text-xl font-semibold mb-2">No events found</h3>
            <p className="text-muted-foreground max-w-md mx-auto">
              We couldn&apos;t find any events matching your search. Try adjusting your filters or check back later.
            </p>
            <Button
              variant="outline"
              className="mt-6"
              onClick={() => {
                setSearchTerm('');
                setCategoryFilter('');
                setCityFilter('');
                loadEvents();
              }}
            >
              Clear all filters
            </Button>
          </div>
        ) : (
          <>
            <div className="grid md:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4 gap-8">
              {currentEvents.map((event) => (
                <EventCard
                  key={event.id}
                  event={event}
                  formatAmount={formatAmount}
                  formatDate={formatDate}
                  formatTime={formatTime}
                  getMinPrice={getMinPrice}
                  isUpcoming={(date) => new Date(date) > new Date()}
                />
              ))}
            </div>

            {/* Pagination */}
            {totalPages > 1 && (
              <div className="mt-12">
                <Pagination>
                  <PaginationContent>
                    <PaginationItem>
                      <PaginationPrevious
                        onClick={() => currentPage > 1 && goToPage(currentPage - 1)}
                        className={currentPage === 1 ? 'pointer-events-none opacity-50' : 'cursor-pointer'}
                      />
                    </PaginationItem>

                    {getPageNumbers().map((page, index) => (
                      <PaginationItem key={index}>
                        {page === 'ellipsis' ? (
                          <PaginationEllipsis />
                        ) : (
                          <PaginationLink
                            onClick={() => goToPage(page as number)}
                            isActive={currentPage === page}
                            className="cursor-pointer"
                          >
                            {page}
                          </PaginationLink>
                        )}
                      </PaginationItem>
                    ))}

                    <PaginationItem>
                      <PaginationNext
                        onClick={() => currentPage < totalPages && goToPage(currentPage + 1)}
                        className={currentPage === totalPages ? 'pointer-events-none opacity-50' : 'cursor-pointer'}
                      />
                    </PaginationItem>
                  </PaginationContent>
                </Pagination>

                <p className="text-center text-sm text-muted-foreground mt-4">
                  Showing {startIndex + 1} to {Math.min(endIndex, events.length)} of {events.length} events
                </p>
              </div>
            )}
          </>
        )}
      </main>
    </div>
  );
}

export default function EventsPage() {
  return (
    <Suspense fallback={
      <div className="min-h-screen flex flex-col bg-background">
        <Navbar />
        <main className="flex-1 container mx-auto px-4 py-24">
          <div className="grid md:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4 gap-6">
            {[...Array(8)].map((_, i) => (
              <div key={i} className="space-y-3">
                <Skeleton className="h-48 w-full rounded-xl" />
                <Skeleton className="h-4 w-3/4" />
                <Skeleton className="h-4 w-1/2" />
              </div>
            ))}
          </div>
        </main>
      </div>
    }>
      <EventsContent />
    </Suspense>
  );
}
