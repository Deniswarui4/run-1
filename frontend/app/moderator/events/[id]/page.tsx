'use client';

import { useState, useEffect } from 'react';
import { useParams, useRouter } from 'next/navigation';
import { apiClient } from '@/lib/api-client';
import { Event } from '@/lib/types';
import { useAuth } from '@/lib/auth-context';
import { DashboardLayout } from '@/components/dashboard-layout';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import { Badge } from '@/components/ui/badge';
import { Skeleton } from '@/components/ui/skeleton';
import { Dialog, DialogContent, DialogDescription, DialogFooter, DialogHeader, DialogTitle } from '@/components/ui/dialog';
import { Textarea } from '@/components/ui/textarea';
import { Label } from '@/components/ui/label';
import { Calendar, MapPin, User, CheckCircle, XCircle } from 'lucide-react';
import { toast } from 'sonner';

export default function ModeratorEventReviewPage() {
  const params = useParams();
  const router = useRouter();
  const { user, loading: authLoading } = useAuth();
  const [event, setEvent] = useState<Event | null>(null);
  const [loading, setLoading] = useState(true);
  const [reviewDialogOpen, setReviewDialogOpen] = useState(false);
  const [reviewComment, setReviewComment] = useState('');
  const [reviewAction, setReviewAction] = useState<'approve' | 'reject'>('approve');
  const [submitting, setSubmitting] = useState(false);

  useEffect(() => {
    if (!authLoading) {
      if (!user || (user.role !== 'moderator' && user.role !== 'admin')) {
        router.push('/');
      } else if (params.id) {
        loadEventForReview(params.id as string);
      }
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [user, authLoading, router, params.id]);

  const loadEventForReview = async (id: string) => {
    try {
      setLoading(true);
      const eventData = await apiClient.getEventForReview(id);
      setEvent(eventData);
    } catch (error) {
      toast.error('Failed to load event for review');
      console.error(error);
      router.push('/moderator/events/pending'); // Redirect if event not found or other error
    } finally {
      setLoading(false);
    }
  };

  const handleReviewEvent = (action: 'approve' | 'reject') => {
    setReviewAction(action);
    setReviewComment('');
    setReviewDialogOpen(true);
  };

  const submitReview = async () => {
    if (!event) return;

    try {
      setSubmitting(true);
      await apiClient.reviewEvent(event.id, {
        action: reviewAction,
        comment: reviewComment,
      });
      toast.success(`Event ${reviewAction}d successfully`);
      setReviewDialogOpen(false);
      router.push('/moderator/events/pending'); // Go back to the list after review
    } catch (error) {
      toast.error(`Failed to ${reviewAction} event`);
      console.error(error);
    } finally {
      setSubmitting(false);
    }
  };

  const getStatusColor = (status: string) => {
    switch (status) {
      case 'pending': return 'default';
      case 'approved': return 'outline';
      case 'rejected': return 'destructive';
      default: return 'secondary';
    }
  };

  const formatDate = (dateString: string) => {
    return new Date(dateString).toLocaleDateString('en-US', {
      weekday: 'long',
      year: 'numeric',
      month: 'long',
      day: 'numeric',
      hour: '2-digit',
      minute: '2-digit',
    });
  };

  if (authLoading || loading) {
    return (
      <DashboardLayout>
        <main className="flex-1 container mx-auto px-4 py-8">
          <Skeleton className="h-8 w-64 mb-6" />
          <Skeleton className="h-96 w-full" />
        </main>
      </DashboardLayout>
    );
  }

  if (!event) {
    return (
      <DashboardLayout>
        <main className="flex-1 container mx-auto px-4 py-8">
          <div className="text-center py-12">
            <p className="text-muted-foreground text-lg">Event not found or you do not have permission to view it.</p>
          </div>
        </main>
      </DashboardLayout>
    );
  }

  return (
    <DashboardLayout>
      <main className="flex-1 container mx-auto px-4 py-8">
        <div className="flex justify-between items-start mb-8">
          <div>
            <h1 className="text-4xl font-bold mb-2">Review Event</h1>
            <div className="flex items-center gap-2">
              <Badge variant={getStatusColor(event.status)} className="capitalize">
                {event.status}
              </Badge>
              <span className="text-muted-foreground">•</span>
              <span className="text-muted-foreground">{event.title}</span>
            </div>
          </div>
          <div className="flex gap-2">
            <Button onClick={() => handleReviewEvent('approve')}>
              <CheckCircle className="h-4 w-4 mr-2" />
              Approve
            </Button>
            <Button variant="destructive" onClick={() => handleReviewEvent('reject')}>
              <XCircle className="h-4 w-4 mr-2" />
              Reject
            </Button>
          </div>
        </div>

        <div className="grid grid-cols-1 lg:grid-cols-3 gap-8">
          <div className="lg:col-span-2 space-y-6">
            {/* Event Image */}
            <Card>
              <CardHeader>
                <CardTitle>Event Image</CardTitle>
              </CardHeader>
              <CardContent>
                {event.image_url ? (
                  <div className="relative h-80 w-full overflow-hidden rounded-lg">
                    <img
                      src={`${process.env.NEXT_PUBLIC_API_URL?.replace('/api/v1', '') || 'http://localhost:8080'}${event.image_url}`}
                      alt={event.title}
                      className="object-cover w-full h-full"
                    />
                  </div>
                ) : (
                  <div className="h-80 bg-muted flex items-center justify-center rounded-lg">
                    <p className="text-muted-foreground">No image uploaded</p>
                  </div>
                )}
              </CardContent>
            </Card>

            {/* Event Description */}
            <Card>
              <CardHeader>
                <CardTitle>Description</CardTitle>
              </CardHeader>
              <CardContent>
                <p className="text-muted-foreground whitespace-pre-wrap">{event.description}</p>
              </CardContent>
            </Card>
          </div>

          <div className="space-y-6">
            {/* Event Details */}
            <Card>
              <CardHeader>
                <CardTitle>Event Details</CardTitle>
              </CardHeader>
              <CardContent className="space-y-4">
                <div className="flex items-start">
                  <Calendar className="h-4 w-4 mr-3 mt-1 text-muted-foreground" />
                  <div>
                    <h3 className="font-semibold">Date & Time</h3>
                    <p>{formatDate(event.start_date)}</p>
                    <p className="text-sm text-muted-foreground">to {formatDate(event.end_date)}</p>
                  </div>
                </div>
                <div className="flex items-start">
                  <MapPin className="h-4 w-4 mr-3 mt-1 text-muted-foreground" />
                  <div>
                    <h3 className="font-semibold">Location</h3>
                    <p>{event.venue}</p>
                    <p className="text-sm text-muted-foreground">{event.address}</p>
                    <p className="text-sm text-muted-foreground">{event.city}, {event.country}</p>
                  </div>
                </div>
                <div className="flex items-start">
                  <User className="h-4 w-4 mr-3 mt-1 text-muted-foreground" />
                  <div>
                    <h3 className="font-semibold">Organizer</h3>
                    <p>{event.organizer?.first_name} {event.organizer?.last_name}</p>
                    <p className="text-sm text-muted-foreground">{event.organizer?.email}</p>
                  </div>
                </div>
              </CardContent>
            </Card>

            {/* Ticket Details */}
            <Card>
              <CardHeader>
                <CardTitle>Ticket Types</CardTitle>
              </CardHeader>
              <CardContent>
                {event.ticket_types && event.ticket_types.length > 0 ? (
                  <ul className="space-y-2 text-sm">
                    {event.ticket_types.map(ticket => (
                      <li key={ticket.id} className="flex justify-between">
                        <span>{ticket.name}</span>
                        <span className="font-semibold">{ticket.price === 0 ? 'Free' : `$${ticket.price.toFixed(2)}`}</span>
                      </li>
                    ))}
                  </ul>
                ) : (
                  <p className="text-sm text-muted-foreground">No ticket types defined.</p>
                )}
              </CardContent>
            </Card>
          </div>
        </div>

        {/* Review Dialog */}
        <Dialog open={reviewDialogOpen} onOpenChange={setReviewDialogOpen}>
          <DialogContent>
            <DialogHeader>
              <DialogTitle>
                {reviewAction === 'approve' ? 'Approve' : 'Reject'} Event
              </DialogTitle>
              <DialogDescription>
                You are about to {reviewAction} the event &quot;{event.title}&quot;.
              </DialogDescription>
            </DialogHeader>
            <div className="space-y-2 py-4">
              <Label htmlFor="comment">Review Comment *</Label>
              <Textarea
                id="comment"
                placeholder={
                  reviewAction === 'approve' 
                    ? "Optional: Add a comment for the organizer..."
                    : "Please specify the reason for rejection..."
                }
                value={reviewComment}
                onChange={(e) => setReviewComment(e.target.value)}
                rows={4}
                required={reviewAction === 'reject'}
              />
            </div>
            <DialogFooter>
              <Button variant="outline" onClick={() => setReviewDialogOpen(false)}>
                Cancel
              </Button>
              <Button
                variant={reviewAction === 'approve' ? 'default' : 'destructive'}
                onClick={submitReview}
                disabled={submitting || (reviewAction === 'reject' && !reviewComment.trim())}
              >
                {submitting 
                  ? (reviewAction === 'approve' ? 'Approving...' : 'Rejecting...')
                  : `Confirm ${reviewAction === 'approve' ? 'Approval' : 'Rejection'}`
                }
              </Button>
            </DialogFooter>
          </DialogContent>
        </Dialog>
      </main>
    </DashboardLayout>
  );
}
