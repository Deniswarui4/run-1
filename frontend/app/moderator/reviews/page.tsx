'use client';

import { useState, useEffect } from 'react';
import { useRouter } from 'next/navigation';
import Link from 'next/link';
import { apiClient } from '@/lib/api-client';
import { Event } from '@/lib/types';
import { useAuth } from '@/lib/auth-context';
import { DashboardLayout } from '@/components/dashboard-layout';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import { Badge } from '@/components/ui/badge';
import { Skeleton } from '@/components/ui/skeleton';
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs';
import { Calendar, MapPin, User, CheckCircle, XCircle, Eye } from 'lucide-react';
import { toast } from 'sonner';

export default function ModeratorReviewsPage() {
  const router = useRouter();
  const { user, loading: authLoading } = useAuth();
  const [myReviews, setMyReviews] = useState<Event[]>([]);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    if (!authLoading) {
      if (!user || (user.role !== 'moderator' && user.role !== 'admin')) {
        router.push('/');
      } else {
        loadMyReviews();
      }
    }
  }, [user, authLoading, router]);

  const loadMyReviews = async () => {
    try {
      setLoading(true);
      const data = await apiClient.getMyReviews();
      setMyReviews(data);
    } catch (error) {
      toast.error('Failed to load reviews');
      console.error(error);
    } finally {
      setLoading(false);
    }
  };

  const formatDate = (dateString: string) => {
    return new Date(dateString).toLocaleDateString('en-US', {
      weekday: 'short',
      month: 'short',
      day: 'numeric',
      year: 'numeric',
    });
  };

  const getStatusColor = (status: string) => {
    switch (status) {
      case 'approved': return 'default';
      case 'rejected': return 'destructive';
      case 'published': return 'default';
      case 'cancelled': return 'destructive';
      default: return 'outline';
    }
  };

  const getStatusIcon = (status: string) => {
    switch (status) {
      case 'approved':
      case 'published':
        return <CheckCircle className="h-4 w-4" />;
      case 'rejected':
      case 'cancelled':
        return <XCircle className="h-4 w-4" />;
      default:
        return null;
    }
  };

  const filterEventsByStatus = (status: string) => {
    return myReviews.filter(e => e.status === status);
  };

  const renderEventCard = (event: Event) => (
    <Card key={event.id}>
      <CardHeader>
        <div className="flex justify-between items-start">
          <div className="flex-1">
            <CardTitle className="text-xl">{event.title}</CardTitle>
            <CardDescription className="mt-2 space-y-1">
              <div className="flex items-center gap-2">
                <Calendar className="h-4 w-4" />
                {formatDate(event.start_date)}
              </div>
              <div className="flex items-center gap-2">
                <MapPin className="h-4 w-4" />
                {event.venue}, {event.city}
              </div>
              {event.organizer && (
                <div className="flex items-center gap-2">
                  <User className="h-4 w-4" />
                  {event.organizer.first_name} {event.organizer.last_name}
                </div>
              )}
            </CardDescription>
          </div>
          <Badge variant={getStatusColor(event.status)} className="capitalize">
            <span className="flex items-center gap-1">
              {getStatusIcon(event.status)}
              {event.status}
            </span>
          </Badge>
        </div>
      </CardHeader>
      <CardContent>
        <p className="text-muted-foreground text-sm mb-4 line-clamp-2">
          {event.description}
        </p>
        
        <div className="mb-4">
          <div className="grid grid-cols-2 gap-4 text-sm">
            <div>
              <span className="text-muted-foreground">Category:</span>
              <span className="ml-2 font-medium">{event.category}</span>
            </div>
            <div>
              <span className="text-muted-foreground">Country:</span>
              <span className="ml-2 font-medium">{event.country}</span>
            </div>
          </div>
        </div>

        <div className="flex gap-2">
          <Link href={`/moderator/events/${event.id}`}>
            <Button variant="outline" size="sm">
              <Eye className="h-4 w-4 mr-2" />
              View Event
            </Button>
          </Link>
          {event.status === 'published' && (
            <Link href={`/events/${event.id}`}>
              <Button variant="outline" size="sm">
                View Public Page
              </Button>
            </Link>
          )}
        </div>
      </CardContent>
    </Card>
  );

  if (authLoading || loading) {
    return (
      <DashboardLayout>
        <div className="container mx-auto px-4 py-8">
          <Skeleton className="h-8 w-64 mb-6" />
          <div className="space-y-4">
            {[...Array(3)].map((_, i) => (
              <Skeleton key={i} className="h-48 w-full" />
            ))}
          </div>
        </div>
      </DashboardLayout>
    );
  }

  return (
    <DashboardLayout>
      <div className="container mx-auto px-4 py-8">
        <h1 className="text-4xl font-bold mb-8">My Reviews</h1>

        <Tabs defaultValue="all" className="space-y-6">
          <TabsList>
            <TabsTrigger value="all">
              All Reviews ({myReviews.length})
            </TabsTrigger>
            <TabsTrigger value="approved">
              Approved ({filterEventsByStatus('approved').length + filterEventsByStatus('published').length})
            </TabsTrigger>
            <TabsTrigger value="rejected">
              Rejected ({filterEventsByStatus('rejected').length})
            </TabsTrigger>
          </TabsList>

          <TabsContent value="all" className="space-y-4">
            {myReviews.length === 0 ? (
              <Card>
                <CardContent className="py-12 text-center">
                  <CheckCircle className="h-16 w-16 mx-auto text-muted-foreground mb-4" />
                  <h3 className="text-lg font-semibold mb-2">No reviews yet</h3>
                  <p className="text-muted-foreground">
                    You haven&apos;t reviewed any events yet. Check the pending events to start reviewing.
                  </p>
                  <Link href="/moderator/events/pending" className="mt-4 inline-block">
                    <Button>
                      Review Pending Events
                    </Button>
                  </Link>
                </CardContent>
              </Card>
            ) : (
              myReviews.map(renderEventCard)
            )}
          </TabsContent>

          <TabsContent value="approved" className="space-y-4">
            {filterEventsByStatus('approved').length + filterEventsByStatus('published').length === 0 ? (
              <Card>
                <CardContent className="py-12 text-center">
                  <CheckCircle className="h-16 w-16 mx-auto text-green-500 mb-4" />
                  <h3 className="text-lg font-semibold mb-2">No approved events</h3>
                  <p className="text-muted-foreground">
                    You haven&apos;t approved any events yet.
                  </p>
                </CardContent>
              </Card>
            ) : (
              [...filterEventsByStatus('approved'), ...filterEventsByStatus('published')].map(renderEventCard)
            )}
          </TabsContent>

          <TabsContent value="rejected" className="space-y-4">
            {filterEventsByStatus('rejected').length === 0 ? (
              <Card>
                <CardContent className="py-12 text-center">
                  <XCircle className="h-16 w-16 mx-auto text-red-500 mb-4" />
                  <h3 className="text-lg font-semibold mb-2">No rejected events</h3>
                  <p className="text-muted-foreground">
                    You haven&apos;t rejected any events yet.
                  </p>
                </CardContent>
              </Card>
            ) : (
              filterEventsByStatus('rejected').map(renderEventCard)
            )}
          </TabsContent>
        </Tabs>

        {/* Quick Stats */}
        {myReviews.length > 0 && (
          <div className="mt-8 grid md:grid-cols-3 gap-6">
            <Card>
              <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
                <CardTitle className="text-sm font-medium">Total Reviews</CardTitle>
                <CheckCircle className="h-4 w-4 text-muted-foreground" />
              </CardHeader>
              <CardContent>
                <div className="text-2xl font-bold">{myReviews.length}</div>
                <p className="text-xs text-muted-foreground">
                  Events reviewed
                </p>
              </CardContent>
            </Card>

            <Card>
              <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
                <CardTitle className="text-sm font-medium">Approved</CardTitle>
                <CheckCircle className="h-4 w-4 text-green-500" />
              </CardHeader>
              <CardContent>
                <div className="text-2xl font-bold text-green-600">
                  {filterEventsByStatus('approved').length + filterEventsByStatus('published').length}
                </div>
                <p className="text-xs text-muted-foreground">
                  Events approved
                </p>
              </CardContent>
            </Card>

            <Card>
              <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
                <CardTitle className="text-sm font-medium">Rejected</CardTitle>
                <XCircle className="h-4 w-4 text-red-500" />
              </CardHeader>
              <CardContent>
                <div className="text-2xl font-bold text-red-600">
                  {filterEventsByStatus('rejected').length}
                </div>
                <p className="text-xs text-muted-foreground">
                  Events rejected
                </p>
              </CardContent>
            </Card>
          </div>
        )}
      </div>
    </DashboardLayout>
  );
}
