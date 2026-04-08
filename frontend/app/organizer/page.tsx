'use client';

import { useState, useEffect } from 'react';
import { useRouter } from 'next/navigation';
import Link from 'next/link';
import { apiClient } from '@/lib/api-client';
import { Event, OrganizerBalance } from '@/lib/types';
import { useAuth } from '@/lib/auth-context';
import { DashboardLayout } from '@/components/dashboard-layout';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
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
import { Calendar, DollarSign, Plus, TrendingUp, Eye } from 'lucide-react';
import { toast } from 'sonner';
import { useCurrency } from '@/lib/currency';

export default function OrganizerDashboard() {
  const router = useRouter();
  const { user, loading: authLoading } = useAuth();
  const { formatAmount } = useCurrency();
  const [events, setEvents] = useState<Event[]>([]);
  const [balance, setBalance] = useState<OrganizerBalance | null>(null);
  const [loading, setLoading] = useState(true);

  // Pagination state for each tab
  const [currentPageAll, setCurrentPageAll] = useState(1);
  const [currentPageDraft, setCurrentPageDraft] = useState(1);
  const [currentPagePending, setCurrentPagePending] = useState(1);
  const [currentPagePublished, setCurrentPagePublished] = useState(1);
  const itemsPerPage = 8;

  useEffect(() => {
    if (!authLoading) {
      if (!user || (user.role !== 'organizer' && user.role !== 'admin')) {
        router.push('/');
      } else {
        loadData();
      }
    }
  }, [user, authLoading, router]);

  const loadData = async () => {
    try {
      setLoading(true);
      const [eventsData, balanceData] = await Promise.all([
        apiClient.getMyEvents(),
        apiClient.getOrganizerBalance(),
      ]);
      setEvents(eventsData);
      setBalance(balanceData);
    } catch (error) {
      toast.error('Failed to load dashboard data');
      console.error(error);
    } finally {
      setLoading(false);
    }
  };

  const getStatusColor = (status: string) => {
    switch (status) {
      case 'draft':
        return 'secondary';
      case 'pending':
        return 'default';
      case 'approved':
        return 'outline';
      case 'published':
        return 'default';
      case 'rejected':
        return 'destructive';
      case 'cancelled':
        return 'destructive';
      default:
        return 'secondary';
    }
  };

  const filterEventsByStatus = (status: string) => {
    return events.filter(e => e.status === status);
  };

  const formatDate = (dateString: string) => {
    return new Date(dateString).toLocaleDateString('en-US', {
      month: 'short',
      day: 'numeric',
      year: 'numeric',
    });
  };

  // Pagination helper function
  const getPaginatedItems = (items: Event[], currentPage: number) => {
    const startIndex = (currentPage - 1) * itemsPerPage;
    const endIndex = startIndex + itemsPerPage;
    return items.slice(startIndex, endIndex);
  };

  const getTotalPages = (items: Event[]) => Math.ceil(items.length / itemsPerPage);

  const getPageNumbers = (totalPages: number, currentPage: number) => {
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

  const renderPagination = (items: Event[], currentPage: number, setCurrentPage: (page: number) => void) => {
    const totalPages = getTotalPages(items);
    if (totalPages <= 1) return null;

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
            {getPageNumbers(totalPages, currentPage).map((page, index) => (
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
          Showing {Math.min((currentPage - 1) * itemsPerPage + 1, items.length)} to {Math.min(currentPage * itemsPerPage, items.length)} of {items.length} events
        </p>
      </div>
    );
  };

  if (authLoading || loading) {
    return (
      <DashboardLayout>
        <div className="container mx-auto px-4 py-8">
          <Skeleton className="h-8 w-64 mb-6" />
          <div className="grid md:grid-cols-3 gap-6 mb-8">
            {[...Array(3)].map((_, i) => (
              <Skeleton key={i} className="h-32" />
            ))}
          </div>
        </div>
      </DashboardLayout>
    );
  }

  return (
    <DashboardLayout>
      <div className="container mx-auto px-4 py-8">
        <div className="flex justify-between items-center mb-8">
          <h1 className="text-4xl font-bold">Organizer Dashboard</h1>
          <Link href="/organizer/events/create">
            <Button size="lg">
              <Plus className="h-4 w-4 mr-2" />
              Create Event
            </Button>
          </Link>
        </div>

        {/* Balance Cards */}
        <div className="grid md:grid-cols-3 gap-6 mb-8">
          <Card>
            <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
              <CardTitle className="text-sm font-medium">Total Earnings</CardTitle>
              <DollarSign className="h-4 w-4 text-muted-foreground" />
            </CardHeader>
            <CardContent>
              <div className="text-2xl font-bold">
                {balance ? formatAmount(balance.total_earnings) : formatAmount(0)}
              </div>
            </CardContent>
          </Card>

          <Card>
            <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
              <CardTitle className="text-sm font-medium">Available Balance</CardTitle>
              <TrendingUp className="h-4 w-4 text-muted-foreground" />
            </CardHeader>
            <CardContent>
              <div className="text-2xl font-bold">
                {balance ? formatAmount(balance.available_balance) : formatAmount(0)}
              </div>
              <Link href="/organizer/withdrawals">
                <Button variant="link" className="px-0 mt-2">
                  Request Withdrawal
                </Button>
              </Link>
            </CardContent>
          </Card>

          <Card>
            <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
              <CardTitle className="text-sm font-medium">Total Events</CardTitle>
              <Calendar className="h-4 w-4 text-muted-foreground" />
            </CardHeader>
            <CardContent>
              <div className="text-2xl font-bold">{events.length}</div>
              <p className="text-xs text-muted-foreground mt-2">
                {filterEventsByStatus('published').length} published
              </p>
            </CardContent>
          </Card>
        </div>

        {/* Events List */}
        <Card>
          <CardHeader>
            <CardTitle>My Events</CardTitle>
            <CardDescription>Manage your events and track their status</CardDescription>
          </CardHeader>
          <CardContent>
            <Tabs defaultValue="all">
              <TabsList className="mb-4">
                <TabsTrigger value="all">All ({events.length})</TabsTrigger>
                <TabsTrigger value="draft">Draft ({filterEventsByStatus('draft').length})</TabsTrigger>
                <TabsTrigger value="pending">Pending ({filterEventsByStatus('pending').length})</TabsTrigger>
                <TabsTrigger value="published">Published ({filterEventsByStatus('published').length})</TabsTrigger>
              </TabsList>

              <TabsContent value="all" className="space-y-4">
                {events.length === 0 ? (
                  <div className="text-center py-12">
                    <p className="text-muted-foreground mb-4">You haven&apos;t created any events yet</p>
                    <Link href="/organizer/events/create">
                      <Button>Create Your First Event</Button>
                    </Link>
                  </div>
                ) : (
                  <>
                    {getPaginatedItems(events, currentPageAll).map((event) => (
                      <Card key={event.id}>
                        <CardHeader>
                          <div className="flex justify-between items-start">
                            <div className="flex-1">
                              <CardTitle className="text-xl">{event.title}</CardTitle>
                              <CardDescription className="mt-2">
                                <div className="flex items-center gap-2">
                                  <Calendar className="h-4 w-4" />
                                  {formatDate(event.start_date)}
                                </div>
                              </CardDescription>
                            </div>
                            <Badge variant={getStatusColor(event.status)} className="capitalize">
                              {event.status}
                            </Badge>
                          </div>
                        </CardHeader>
                        <CardContent>
                          <div className="flex gap-2">
                            <Link href={`/organizer/events/${event.id}`} className="flex-1">
                              <Button variant="outline" className="w-full">
                                <Eye className="h-4 w-4 mr-2" />
                                Manage
                              </Button>
                            </Link>
                            {event.status === 'published' && (
                              <Link href={`/events/${event.id}`} className="flex-1">
                                <Button variant="outline" className="w-full">
                                  View Public Page
                                </Button>
                              </Link>
                            )}
                          </div>
                        </CardContent>
                      </Card>
                    ))}
                    {renderPagination(events, currentPageAll, setCurrentPageAll)}
                  </>
                )}
              </TabsContent>

              {['draft', 'pending', 'published'].map((status) => (
                <TabsContent key={status} value={status} className="space-y-4">
                  {filterEventsByStatus(status).length === 0 ? (
                    <div className="text-center py-12">
                      <p className="text-muted-foreground">No {status} events</p>
                    </div>
                  ) : (
                    <>
                      {(() => {
                        const filteredEvents = filterEventsByStatus(status);
                        const currentPage = status === 'draft' ? currentPageDraft : status === 'pending' ? currentPagePending : currentPagePublished;
                        const setCurrentPage = status === 'draft' ? setCurrentPageDraft : status === 'pending' ? setCurrentPagePending : setCurrentPagePublished;

                        return (
                          <>
                            {getPaginatedItems(filteredEvents, currentPage).map((event) => (
                              <Card key={event.id}>
                                <CardHeader>
                                  <div className="flex justify-between items-start">
                                    <div className="flex-1">
                                      <CardTitle className="text-xl">{event.title}</CardTitle>
                                      <CardDescription className="mt-2">
                                        <div className="flex items-center gap-2">
                                          <Calendar className="h-4 w-4" />
                                          {formatDate(event.start_date)}
                                        </div>
                                      </CardDescription>
                                    </div>
                                  </div>
                                </CardHeader>
                                <CardContent>
                                  <div className="flex gap-2">
                                    <Link href={`/organizer/events/${event.id}`} className="flex-1">
                                      <Button variant="outline" className="w-full">
                                        <Eye className="h-4 w-4 mr-2" />
                                        Manage
                                      </Button>
                                    </Link>
                                    {event.status === 'published' && (
                                      <Link href={`/events/${event.id}`} className="flex-1">
                                        <Button variant="outline" className="w-full">
                                          View Public Page
                                        </Button>
                                      </Link>
                                    )}
                                  </div>
                                </CardContent>
                              </Card>
                            ))}
                            {renderPagination(filteredEvents, currentPage, setCurrentPage)}
                          </>
                        );
                      })()}
                    </>
                  )}
                </TabsContent>
              ))}
            </Tabs>
          </CardContent>
        </Card>
      </div>
    </DashboardLayout>
  );
}
