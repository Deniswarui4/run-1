'use client';

import { useState, useEffect } from 'react';
import { useRouter } from 'next/navigation';
import Link from 'next/link';
import { apiClient } from '@/lib/api-client';
import { Event, ModeratorStats } from '@/lib/types';
import { useAuth } from '@/lib/auth-context';
import { DashboardLayout } from '@/components/dashboard-layout';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import { Badge } from '@/components/ui/badge';
import { Skeleton } from '@/components/ui/skeleton';
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs';
import { Dialog, DialogContent, DialogDescription, DialogFooter, DialogHeader, DialogTitle } from '@/components/ui/dialog';
import { Textarea } from '@/components/ui/textarea';
import { Label } from '@/components/ui/label';
import { Calendar, MapPin, User, CheckCircle, XCircle, Clock, TrendingUp } from 'lucide-react';
import { toast } from 'sonner';

export default function ModeratorDashboard() {
  const router = useRouter();
  const { user, loading: authLoading } = useAuth();
  const [pendingEvents, setPendingEvents] = useState<Event[]>([]);
  const [myReviews, setMyReviews] = useState<Event[]>([]);
  const [stats, setStats] = useState<ModeratorStats | null>(null);
  const [loading, setLoading] = useState(true);
  const [reviewDialogOpen, setReviewDialogOpen] = useState(false);
  const [selectedEvent, setSelectedEvent] = useState<Event | null>(null);
  const [reviewComment, setReviewComment] = useState('');
  const [reviewAction, setReviewAction] = useState<'approve' | 'reject'>('approve');
  const [submitting, setSubmitting] = useState(false);

  useEffect(() => {
    if (!authLoading) {
      if (!user || (user.role !== 'moderator' && user.role !== 'admin')) {
        router.push('/');
      } else {
        loadData();
      }
    }
  }, [user, authLoading, router]);

  const loadData = async () => {
    try {
      setLoading(true);
      const [pendingData, reviewsData, statsData] = await Promise.all([
        apiClient.getPendingEvents(),
        apiClient.getMyReviews(),
        apiClient.getModeratorStats(),
      ]);
      setPendingEvents(pendingData);
      setMyReviews(reviewsData);
      setStats(statsData);
    } catch (error) {
      toast.error('Failed to load moderator data');
      console.error(error);
    } finally {
      setLoading(false);
    }
  };

  const handleReviewEvent = (event: Event, action: 'approve' | 'reject') => {
    setSelectedEvent(event);
    setReviewAction(action);
    setReviewComment('');
    setReviewDialogOpen(true);
  };

  const submitReview = async () => {
    if (!selectedEvent) return;

    try {
      setSubmitting(true);
      await apiClient.reviewEvent(selectedEvent.id, {
        action: reviewAction,
        comment: reviewComment,
      });
      toast.success(`Event ${reviewAction}d successfully`);
      setReviewDialogOpen(false);
      loadData(); // Reload data
    } catch (error) {
      toast.error(`Failed to ${reviewAction} event`);
      console.error(error);
    } finally {
      setSubmitting(false);
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
      case 'pending': return 'secondary';
      default: return 'outline';
    }
  };

  const renderEventCard = (event: Event, showActions = false) => (
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
            {event.status}
          </Badge>
        </div>
      </CardHeader>
      <CardContent>
        <p className="text-muted-foreground text-sm mb-4 line-clamp-2">
          {event.description}
        </p>
        
        {showActions && (
          <div className="flex gap-2">
            <Link href={`/events/${event.id}`} className="flex-1">
              <Button variant="outline" className="w-full">
                View Details
              </Button>
            </Link>
            <Button
              className="flex-1"
              onClick={() => handleReviewEvent(event, 'approve')}
            >
              <CheckCircle className="h-4 w-4 mr-2" />
              Approve
            </Button>
            <Button
              variant="destructive"
              className="flex-1"
              onClick={() => handleReviewEvent(event, 'reject')}
            >
              <XCircle className="h-4 w-4 mr-2" />
              Reject
            </Button>
          </div>
        )}
      </CardContent>
    </Card>
  );

  if (authLoading || loading) {
    return (
      <DashboardLayout>
        <div className="container mx-auto px-4 py-8">
          <Skeleton className="h-8 w-64 mb-6" />
          <div className="grid md:grid-cols-4 gap-6 mb-8">
            {[...Array(4)].map((_, i) => (
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
        <h1 className="text-4xl font-bold mb-8">Moderator Dashboard</h1>

        {/* Stats Cards */}
        {stats && (
          <div className="grid md:grid-cols-2 lg:grid-cols-4 gap-6 mb-8">
            <Card>
              <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
                <CardTitle className="text-sm font-medium">Pending Events</CardTitle>
                <Clock className="h-4 w-4 text-muted-foreground" />
              </CardHeader>
              <CardContent>
                <div className="text-2xl font-bold">{stats.pending_events}</div>
              </CardContent>
            </Card>

            <Card>
              <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
                <CardTitle className="text-sm font-medium">Approved Events</CardTitle>
                <CheckCircle className="h-4 w-4 text-muted-foreground" />
              </CardHeader>
              <CardContent>
                <div className="text-2xl font-bold">{stats.approved_events}</div>
              </CardContent>
            </Card>

            <Card>
              <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
                <CardTitle className="text-sm font-medium">Rejected Events</CardTitle>
                <XCircle className="h-4 w-4 text-muted-foreground" />
              </CardHeader>
              <CardContent>
                <div className="text-2xl font-bold">{stats.rejected_events}</div>
              </CardContent>
            </Card>

            <Card>
              <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
                <CardTitle className="text-sm font-medium">My Reviews</CardTitle>
                <TrendingUp className="h-4 w-4 text-muted-foreground" />
              </CardHeader>
              <CardContent>
                <div className="text-2xl font-bold">{stats.my_reviews}</div>
              </CardContent>
            </Card>
          </div>
        )}

        <Tabs defaultValue="pending" className="space-y-6">
          <TabsList>
            <TabsTrigger value="pending">
              Pending Review ({pendingEvents.length})
            </TabsTrigger>
            <TabsTrigger value="reviewed">
              My Reviews ({myReviews.length})
            </TabsTrigger>
          </TabsList>

          <TabsContent value="pending" className="space-y-4">
            {pendingEvents.length === 0 ? (
              <Card>
                <CardContent className="py-12 text-center">
                  <p className="text-muted-foreground">No events pending review</p>
                </CardContent>
              </Card>
            ) : (
              pendingEvents.map((event) => renderEventCard(event, true))
            )}
          </TabsContent>

          <TabsContent value="reviewed" className="space-y-4">
            {myReviews.length === 0 ? (
              <Card>
                <CardContent className="py-12 text-center">
                  <p className="text-muted-foreground">You haven&apos;t reviewed any events yet</p>
                </CardContent>
              </Card>
            ) : (
              myReviews.map((event) => renderEventCard(event, false))
            )}
          </TabsContent>
        </Tabs>

        {/* Review Dialog */}
      <Dialog open={reviewDialogOpen} onOpenChange={setReviewDialogOpen}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>
              {reviewAction === 'approve' ? 'Approve' : 'Reject'} Event
            </DialogTitle>
            <DialogDescription>
              {selectedEvent && (
                <>
                  {reviewAction === 'approve' ? 'Approve' : 'Reject'} &quot;{selectedEvent.title}&quot;
                </>
              )}
            </DialogDescription>
          </DialogHeader>
          
          <div className="space-y-4 py-4">
            {selectedEvent && (
              <div className="space-y-2">
                <h4 className="font-medium">Event Details</h4>
                <div className="text-sm text-muted-foreground space-y-1">
                  <p><strong>Organizer:</strong> {selectedEvent.organizer?.first_name} {selectedEvent.organizer?.last_name}</p>
                  <p><strong>Date:</strong> {formatDate(selectedEvent.start_date)}</p>
                  <p><strong>Location:</strong> {selectedEvent.venue}, {selectedEvent.city}</p>
                  <p><strong>Category:</strong> {selectedEvent.category}</p>
                </div>
              </div>
            )}
            
            <div className="space-y-2">
              <Label htmlFor="comment">Review Comment *</Label>
              <Textarea
                id="comment"
                placeholder={
                  reviewAction === 'approve' 
                    ? "Event meets all guidelines and requirements..."
                    : "Please specify the reason for rejection..."
                }
                value={reviewComment}
                onChange={(e) => setReviewComment(e.target.value)}
                rows={4}
                required
              />
            </div>
          </div>

          <DialogFooter>
            <Button variant="outline" onClick={() => setReviewDialogOpen(false)}>
              Cancel
            </Button>
            <Button
              variant={reviewAction === 'approve' ? 'default' : 'destructive'}
              onClick={submitReview}
              disabled={submitting || !reviewComment.trim()}
            >
              {submitting 
                ? (reviewAction === 'approve' ? 'Approving...' : 'Rejecting...')
                : (reviewAction === 'approve' ? 'Approve Event' : 'Reject Event')
              }
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
      </div>
    </DashboardLayout>
  );
}
