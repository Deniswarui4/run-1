'use client';

import { useState, useEffect } from 'react';
import { useRouter } from 'next/navigation';

import { apiClient } from '@/lib/api-client';
import { Event } from '@/lib/types';
import { useAuth } from '@/lib/auth-context';
import { DashboardLayout } from '@/components/dashboard-layout';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import { Badge } from '@/components/ui/badge';
import { Skeleton } from '@/components/ui/skeleton';
import { Dialog, DialogContent, DialogDescription, DialogFooter, DialogHeader, DialogTitle } from '@/components/ui/dialog';
import { Textarea } from '@/components/ui/textarea';
import { Label } from '@/components/ui/label';
import { Calendar, MapPin, User, CheckCircle, XCircle } from 'lucide-react';
import { toast } from 'sonner';

export default function ModeratorPendingEventsPage() {
  const router = useRouter();
  const { user, loading: authLoading } = useAuth();
  const [pendingEvents, setPendingEvents] = useState<Event[]>([]);
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
        loadPendingEvents();
      }
    }
  }, [user, authLoading, router]);

  const loadPendingEvents = async () => {
    try {
      setLoading(true);
      const data = await apiClient.getPendingEvents();
      setPendingEvents(data);
    } catch (error) {
      toast.error('Failed to load pending events');
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
      loadPendingEvents(); // Reload data
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
          <Badge variant="secondary" className="capitalize">
            {event.status}
          </Badge>
        </div>
      </CardHeader>
      <CardContent>
        <p className="text-muted-foreground text-sm mb-4 line-clamp-3">
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
            {event.address && (
              <div className="col-span-2">
                <span className="text-muted-foreground">Address:</span>
                <span className="ml-2 font-medium">{event.address}</span>
              </div>
            )}
          </div>
        </div>

        <div className="flex gap-2">
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
              <Skeleton key={i} className="h-64 w-full" />
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
          <h1 className="text-4xl font-bold">Pending Events</h1>
          <div className="text-muted-foreground">
            {pendingEvents.length} event{pendingEvents.length !== 1 ? 's' : ''} awaiting review
          </div>
        </div>

        {pendingEvents.length === 0 ? (
          <Card>
            <CardContent className="py-12 text-center">
              <CheckCircle className="h-16 w-16 mx-auto text-muted-foreground mb-4" />
              <h3 className="text-lg font-semibold mb-2">All caught up!</h3>
              <p className="text-muted-foreground">
                There are no events pending review at the moment.
              </p>
            </CardContent>
          </Card>
        ) : (
          <div className="space-y-6">
            {pendingEvents.map(renderEventCard)}
          </div>
        )}

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
                    {selectedEvent.address && (
                      <p><strong>Address:</strong> {selectedEvent.address}</p>
                    )}
                  </div>
                  
                  <div className="mt-3">
                    <h5 className="font-medium text-sm">Description:</h5>
                    <p className="text-sm text-muted-foreground mt-1">
                      {selectedEvent.description}
                    </p>
                  </div>
                </div>
              )}
              
              <div className="space-y-2">
                <Label htmlFor="comment">Review Comment *</Label>
                <Textarea
                  id="comment"
                  placeholder={
                    reviewAction === 'approve' 
                      ? "Event meets all guidelines and requirements. Approved for publication."
                      : "Please specify the reason for rejection and any required changes..."
                  }
                  value={reviewComment}
                  onChange={(e) => setReviewComment(e.target.value)}
                  rows={4}
                  required
                />
                <p className="text-xs text-muted-foreground">
                  {reviewAction === 'approve' 
                    ? "Provide feedback for the organizer about the approval."
                    : "Be specific about what needs to be changed for resubmission."
                  }
                </p>
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
