'use client';

import { useState, useEffect } from 'react';
import { useRouter } from 'next/navigation';
import Link from 'next/link';
import { apiClient } from '@/lib/api-client';
import { Ticket } from '@/lib/types';
import { useAuth } from '@/lib/auth-context';
import { DashboardLayout } from '@/components/dashboard-layout';
import { Button } from '@/components/ui/button';
import { Badge } from '@/components/ui/badge';
import { Skeleton } from '@/components/ui/skeleton';
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs';
import {
  Pagination,
  PaginationContent,
  PaginationEllipsis,
  PaginationItem,
  PaginationLink,
  PaginationNext,
  PaginationPrevious,
} from '@/components/ui/pagination';
import { Calendar, MapPin, QrCode, Ticket as TicketIcon } from 'lucide-react';
import { toast } from 'sonner';
import { useCurrency } from '@/lib/currency';

export default function MyTicketsPage() {
  const router = useRouter();
  const { user, loading: authLoading } = useAuth();
  const { formatAmount } = useCurrency();
  const [tickets, setTickets] = useState<Ticket[]>([]);
  const [loading, setLoading] = useState(true);

  // Pagination state
  const [currentPageUpcoming, setCurrentPageUpcoming] = useState(1);
  const [currentPagePast, setCurrentPagePast] = useState(1);
  const itemsPerPage = 10;

  useEffect(() => {
    if (!authLoading) {
      if (!user) {
        router.push('/login');
      } else {
        loadTickets();
      }
    }
  }, [user, authLoading, router]);

  const loadTickets = async () => {
    try {
      setLoading(true);
      const data = await apiClient.getMyTickets();
      setTickets(data);
    } catch (error) {
      toast.error('Failed to load tickets');
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
      hour: '2-digit',
      minute: '2-digit',
    });
  };

  const getStatusColor = (status: string) => {
    switch (status) {
      case 'confirmed':
        return 'default';
      case 'used':
        return 'secondary';
      case 'cancelled':
        return 'destructive';
      default:
        return 'outline';
    }
  };

  const upcomingTickets = tickets.filter(
    t => t.status === 'confirmed' && t.event && new Date(t.event.start_date) > new Date()
  );

  const pastTickets = tickets.filter(
    t => t.event && new Date(t.event.start_date) <= new Date()
  );

  // Pagination helpers
  const getPaginatedItems = (items: Ticket[], currentPage: number) => {
    const startIndex = (currentPage - 1) * itemsPerPage;
    return items.slice(startIndex, startIndex + itemsPerPage);
  };

  const getTotalPages = (items: Ticket[]) => Math.ceil(items.length / itemsPerPage);

  const renderPagination = (items: Ticket[], currentPage: number, setCurrentPage: (page: number) => void) => {
    const totalPages = getTotalPages(items);
    if (totalPages <= 1) return null;

    const getPageNumbers = () => {
      const pages: (number | string)[] = [];
      if (totalPages <= 5) {
        for (let i = 1; i <= totalPages; i++) pages.push(i);
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
      <div className="mt-6">
        <Pagination>
          <PaginationContent>
            <PaginationItem>
              <PaginationPrevious
                onClick={() => {
                  if (currentPage > 1) {
                    setCurrentPage(currentPage - 1);
                    window.scrollTo({ top: 0, behavior: 'smooth' });
                  }
                }}
                className={currentPage === 1 ? 'pointer-events-none opacity-50' : 'cursor-pointer'}
              />
            </PaginationItem>
            {getPageNumbers().map((page, index) => (
              <PaginationItem key={index}>
                {page === 'ellipsis' ? (
                  <PaginationEllipsis />
                ) : (
                  <PaginationLink
                    onClick={() => {
                      setCurrentPage(page as number);
                      window.scrollTo({ top: 0, behavior: 'smooth' });
                    }}
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
                onClick={() => {
                  if (currentPage < totalPages) {
                    setCurrentPage(currentPage + 1);
                    window.scrollTo({ top: 0, behavior: 'smooth' });
                  }
                }}
                className={currentPage === totalPages ? 'pointer-events-none opacity-50' : 'cursor-pointer'}
              />
            </PaginationItem>
          </PaginationContent>
        </Pagination>
        <p className="text-center text-sm text-muted-foreground mt-4">
          Showing {Math.min((currentPage - 1) * itemsPerPage + 1, items.length)} to {Math.min(currentPage * itemsPerPage, items.length)} of {items.length} tickets
        </p>
      </div>
    );
  };

  const renderTicketCard = (ticket: Ticket) => (
    <div key={ticket.id} className="group relative bg-card rounded-3xl border border-border/50 overflow-hidden hover:shadow-xl hover:shadow-primary/5 transition-all duration-300 hover:-translate-y-1">
      {/* Decorative top bar */}
      <div className={`h-2 w-full ${ticket.status === 'confirmed' ? 'bg-gradient-to-r from-primary to-purple-500' : 'bg-muted'}`} />

      <div className="p-6">
        <div className="flex justify-between items-start mb-4">
          <Badge variant={getStatusColor(ticket.status)} className="capitalize shadow-sm">
            {ticket.status}
          </Badge>
          <span className="font-mono text-xs text-muted-foreground tracking-wider">#{ticket.ticket_number.slice(-6)}</span>
        </div>

        <h3 className="text-xl font-bold mb-2 line-clamp-1 group-hover:text-primary transition-colors">{ticket.event?.title}</h3>

        <div className="space-y-2 text-sm text-muted-foreground mb-6">
          <div className="flex items-center gap-2">
            <Calendar className="h-4 w-4 text-primary/70" />
            {ticket.event && formatDate(ticket.event.start_date)}
          </div>
          <div className="flex items-center gap-2">
            <MapPin className="h-4 w-4 text-primary/70" />
            <span className="line-clamp-1">{ticket.event?.venue}, {ticket.event?.city}</span>
          </div>
        </div>

        <div className="flex items-center justify-between pt-4 border-t border-dashed border-border">
          <div className="flex flex-col">
            <span className="text-xs text-muted-foreground uppercase tracking-wider">Type</span>
            <span className="font-medium">{ticket.ticket_type?.name}</span>
          </div>
          <div className="flex flex-col items-end">
            <span className="text-xs text-muted-foreground uppercase tracking-wider">Price</span>
            <span className="font-bold text-primary">{formatAmount(ticket.price)}</span>
          </div>
        </div>
      </div>

      <div className="bg-muted/30 p-4 flex gap-3">
        {ticket.status === 'confirmed' ? (
          <>
            {ticket.qr_code_url && (
              <Button variant="outline" size="sm" className="flex-1 bg-background hover:bg-background/80" asChild>
                <a
                  href={`${process.env.NEXT_PUBLIC_API_URL?.replace('/api/v1', '') || 'http://localhost:8080'}${ticket.qr_code_url}`}
                  target="_blank"
                  rel="noopener noreferrer"
                >
                  <QrCode className="h-4 w-4 mr-2" />
                  QR Code
                </a>
              </Button>
            )}
            <Link href={`/tickets/${ticket.id}`} className="flex-1">
              <Button size="sm" className="w-full shadow-md shadow-primary/20">
                View Ticket
              </Button>
            </Link>
          </>
        ) : (
          <Link href={`/tickets/${ticket.id}`} className="w-full">
            <Button variant="outline" size="sm" className="w-full">
              View Details
            </Button>
          </Link>
        )}
      </div>
    </div>
  );

  if (authLoading || loading) {
    return (
      <DashboardLayout>
        <div className="container mx-auto px-4 py-8">
          <div className="flex items-center justify-between mb-8">
            <Skeleton className="h-10 w-48" />
            <Skeleton className="h-10 w-32" />
          </div>
          <div className="grid md:grid-cols-2 lg:grid-cols-3 gap-6">
            {[...Array(6)].map((_, i) => (
              <Skeleton key={i} className="h-64 w-full rounded-3xl" />
            ))}
          </div>
        </div>
      </DashboardLayout>
    );
  }

  return (
    <DashboardLayout>
      <div className="container mx-auto px-4 py-8">
        <div className="flex flex-col md:flex-row md:items-center justify-between gap-4 mb-10">
          <div>
            <h1 className="text-3xl md:text-4xl font-bold tracking-tight">My Tickets</h1>
            <p className="text-muted-foreground mt-1">Manage your upcoming events and purchase history</p>
          </div>
          <Link href="/events">
            <Button className="shadow-lg shadow-primary/20 rounded-full px-6">
              Browse Events
            </Button>
          </Link>
        </div>

        <Tabs defaultValue="upcoming" className="w-full">
          <TabsList className="mb-8 p-1 bg-muted/50 rounded-full inline-flex">
            <TabsTrigger value="upcoming" className="rounded-full px-6 py-2 data-[state=active]:bg-background data-[state=active]:shadow-sm transition-all">
              Upcoming ({upcomingTickets.length})
            </TabsTrigger>
            <TabsTrigger value="past" className="rounded-full px-6 py-2 data-[state=active]:bg-background data-[state=active]:shadow-sm transition-all">
              Past Events ({pastTickets.length})
            </TabsTrigger>
          </TabsList>

          <TabsContent value="upcoming" className="space-y-6 animate-in fade-in slide-in-from-bottom-4 duration-500">
            {upcomingTickets.length === 0 ? (
              <div className="text-center py-20 bg-muted/10 rounded-3xl border border-dashed border-border">
                <div className="bg-primary/10 w-20 h-20 rounded-full flex items-center justify-center mx-auto mb-6">
                  <TicketIcon className="h-10 w-10 text-primary" />
                </div>
                <h3 className="text-xl font-semibold mb-2">No upcoming tickets</h3>
                <p className="text-muted-foreground mb-6 max-w-md mx-auto">You don&apos;t have any upcoming events scheduled. Explore our events to find your next experience!</p>
                <Button onClick={() => router.push('/events')} size="lg" className="rounded-full px-8">
                  Browse Events
                </Button>
              </div>
            ) : (
              <>
                <div className="grid md:grid-cols-2 lg:grid-cols-3 gap-6">
                  {getPaginatedItems(upcomingTickets, currentPageUpcoming).map(renderTicketCard)}
                </div>
                {renderPagination(upcomingTickets, currentPageUpcoming, setCurrentPageUpcoming)}
              </>
            )}
          </TabsContent>

          <TabsContent value="past" className="space-y-6 animate-in fade-in slide-in-from-bottom-4 duration-500">
            {pastTickets.length === 0 ? (
              <div className="text-center py-20 bg-muted/10 rounded-3xl border border-dashed border-border">
                <p className="text-muted-foreground">No past tickets found</p>
              </div>
            ) : (
              <>
                <div className="grid md:grid-cols-2 lg:grid-cols-3 gap-6">
                  {getPaginatedItems(pastTickets, currentPagePast).map(renderTicketCard)}
                </div>
                {renderPagination(pastTickets, currentPagePast, setCurrentPagePast)}
              </>
            )}
          </TabsContent>
        </Tabs>
      </div>
    </DashboardLayout>
  );
}
