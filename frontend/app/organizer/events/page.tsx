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
import { Calendar, MapPin, Plus, Eye, Edit, Send, Globe } from 'lucide-react';
import { toast } from 'sonner';

export default function OrganizerEventsPage() {
  const router = useRouter();
  const { user, loading: authLoading } = useAuth();
  const [events, setEvents] = useState<Event[]>([]);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    if (!authLoading) {
      if (!user || (user.role !== 'organizer' && user.role !== 'admin')) {
        router.push('/');
      } else {
        loadEvents();
      }
    }
  }, [user, authLoading, router]);

  const loadEvents = async () => {
    try {
      setLoading(true);
      const data = await apiClient.getMyEvents();
      setEvents(data);
    } catch (error) {
      toast.error('Failed to load events');
      console.error(error);
    } finally {
      setLoading(false);
    }
  };

  const handleSubmitForReview = async (eventId: string) => {
    try {
      await apiClient.submitEventForReview(eventId);
      toast.success('Event submitted for review');
      loadEvents(); // Reload events
    } catch (error) {
      toast.error('Failed to submit event for review');
      console.error(error);
    }
  };

  const handlePublishEvent = async (eventId: string) => {
    try {
      await apiClient.publishEvent(eventId);
      toast.success('Event published successfully');
      loadEvents(); // Reload events
    } catch (error) {
      toast.error('Failed to publish event');
      console.error(error);
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
      weekday: 'short',
      month: 'short',
      day: 'numeric',
      year: 'numeric',
    });
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
            </CardDescription>
          </div>
          <Badge variant={getStatusColor(event.status)} className="capitalize">
            {event.status}
          </Badge>
        </div>
      </CardHeader>
      <CardContent>
        <p className="text-muted-foreground text-sm mb-4 line-clamp-2">
          {event.description}
        </p>
        
        <div className="flex gap-2 flex-wrap">
          <Link href={`/organizer/events/${event.id}`}>
            <Button variant="outline" size="sm">
              <Eye className="h-4 w-4 mr-2" />
              View Details
            </Button>
          </Link>
          
          {event.status === 'draft' && (
            <>
              <Link href={`/organizer/events/${event.id}`}>
                <Button variant="outline" size="sm">
                  <Edit className="h-4 w-4 mr-2" />
                  Edit
                </Button>
              </Link>
              <Button 
                size="sm"
                onClick={() => handleSubmitForReview(event.id)}
              >
                <Send className="h-4 w-4 mr-2" />
                Submit for Review
              </Button>
            </>
          )}
          
          {event.status === 'approved' && (
            <Button 
              size="sm"
              onClick={() => handlePublishEvent(event.id)}
            >
              <Globe className="h-4 w-4 mr-2" />
              Publish
            </Button>
          )}
          
          {event.status === 'published' && (
            <Link href={`/events/${event.id}`}>
              <Button variant="outline" size="sm">
                <Globe className="h-4 w-4 mr-2" />
                View Public
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
        <div className="flex justify-between items-center mb-8">
          <h1 className="text-4xl font-bold">My Events</h1>
          <Link href="/organizer/events/create">
            <Button size="lg">
              <Plus className="h-4 w-4 mr-2" />
              Create Event
            </Button>
          </Link>
        </div>

        <Tabs defaultValue="all" className="space-y-6">
          <TabsList>
            <TabsTrigger value="all">
              All Events ({events.length})
            </TabsTrigger>
            <TabsTrigger value="draft">
              Draft ({filterEventsByStatus('draft').length})
            </TabsTrigger>
            <TabsTrigger value="pending">
              Pending ({filterEventsByStatus('pending').length})
            </TabsTrigger>
            <TabsTrigger value="published">
              Published ({filterEventsByStatus('published').length})
            </TabsTrigger>
          </TabsList>

          <TabsContent value="all" className="space-y-4">
            {events.length === 0 ? (
              <Card>
                <CardContent className="py-12 text-center">
                  <p className="text-muted-foreground mb-4">You haven&#39;t created any events yet</p>
                  <Link href="/organizer/events/create">
                    <Button>
                      <Plus className="h-4 w-4 mr-2" />
                      Create Your First Event
                    </Button>
                  </Link>
                </CardContent>
              </Card>
            ) : (
              events.map(renderEventCard)
            )}
          </TabsContent>

          <TabsContent value="draft" className="space-y-4">
            {filterEventsByStatus('draft').length === 0 ? (
              <Card>
                <CardContent className="py-12 text-center">
                  <p className="text-muted-foreground">No draft events</p>
                </CardContent>
              </Card>
            ) : (
              filterEventsByStatus('draft').map(renderEventCard)
            )}
          </TabsContent>

          <TabsContent value="pending" className="space-y-4">
            {filterEventsByStatus('pending').length === 0 ? (
              <Card>
                <CardContent className="py-12 text-center">
                  <p className="text-muted-foreground">No events pending review</p>
                </CardContent>
              </Card>
            ) : (
              filterEventsByStatus('pending').map(renderEventCard)
            )}
          </TabsContent>

          <TabsContent value="published" className="space-y-4">
            {filterEventsByStatus('published').length === 0 ? (
              <Card>
                <CardContent className="py-12 text-center">
                  <p className="text-muted-foreground">No published events</p>
                </CardContent>
              </Card>
            ) : (
              filterEventsByStatus('published').map(renderEventCard)
            )}
          </TabsContent>
        </Tabs>
      </div>
    </DashboardLayout>
  );
}
