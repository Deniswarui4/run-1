'use client';

import { useState, useEffect } from 'react';
import { useParams, useRouter } from 'next/navigation';
import { apiClient } from '@/lib/api-client';
import { Event, EventStats } from '@/lib/types';
import { useAuth } from '@/lib/auth-context';
import { DashboardLayout } from '@/components/dashboard-layout';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import { Badge } from '@/components/ui/badge';
import { Skeleton } from '@/components/ui/skeleton';
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs';
import { Dialog, DialogContent, DialogDescription, DialogFooter, DialogHeader, DialogTitle } from '@/components/ui/dialog';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { Textarea } from '@/components/ui/textarea';
import { Progress } from '@/components/ui/progress';
import { Calendar, MapPin, Plus, Eye, DollarSign, TrendingUp, Ticket, Users, Upload, Send, Sparkles } from 'lucide-react';
import { toast } from 'sonner';
import { useCurrency } from '@/lib/currency';

export default function ManageEventPage() {
  const params = useParams();
  const router = useRouter();
  const { user } = useAuth();
  const { formatAmount, symbol } = useCurrency();
  const [event, setEvent] = useState<Event | null>(null);
  const [stats, setStats] = useState<EventStats | null>(null);
  const [loading, setLoading] = useState(true);
  const [ticketDialogOpen, setTicketDialogOpen] = useState(false);
  const [imageUploading, setImageUploading] = useState(false);
  const [uploadProgress, setUploadProgress] = useState(0);
  const [submitting, setSubmitting] = useState(false);
  const [publishing, setPublishing] = useState(false);

  const [ticketForm, setTicketForm] = useState({
    name: '',
    description: '',
    price: '',
    quantity: '',
    max_per_order: '10',
    sale_start: '',
    sale_end: '',
  });

  // New states for ticket sale date/time
  const [saleStartDate, setSaleStartDate] = useState('');
  const [saleStartTime, setSaleStartTime] = useState({ hour: '09', minute: '00' });
  const [saleEndDate, setSaleEndDate] = useState('');
  const [saleEndTime, setSaleEndTime] = useState({ hour: '23', minute: '00' });


  useEffect(() => {
    if (params.id && user) {
      loadEvent(params.id as string);
    }
  }, [params.id, user]);

  // Combine sale start date and time with local timezone
  useEffect(() => {
    if (saleStartDate) {
      const dateTime = new Date(`${saleStartDate}T${saleStartTime.hour}:${saleStartTime.minute}:00`);
      setTicketForm(prev => ({ ...prev, sale_start: dateTime.toISOString() }));
    }
  }, [saleStartDate, saleStartTime]);

  // Combine sale end date and time with local timezone
  useEffect(() => {
    if (saleEndDate) {
      const dateTime = new Date(`${saleEndDate}T${saleEndTime.hour}:${saleEndTime.minute}:00`);
      setTicketForm(prev => ({ ...prev, sale_end: dateTime.toISOString() }));
    }
  }, [saleEndDate, saleEndTime]);

  const loadEvent = async (id: string) => {
    try {
      setLoading(true);
      const [eventData, statsData] = await Promise.all([
        apiClient.getOrganizerEvent(id),
        apiClient.getEventStats(id).catch(() => null),
      ]);
      setEvent(eventData);
      setStats(statsData);
    } catch (error) {
      toast.error('Failed to load event');
      console.error(error);
    } finally {
      setLoading(false);
    }
  };

  const handleImageUpload = async (e: React.ChangeEvent<HTMLInputElement>) => {
    const file = e.target.files?.[0];
    if (!file || !event) return;

    try {
      setImageUploading(true);
      setUploadProgress(0);

      // Simulate progress for better UX (real progress would require xhr)
      const progressInterval = setInterval(() => {
        setUploadProgress(prev => {
          if (prev >= 90) {
            clearInterval(progressInterval);
            return 90;
          }
          return prev + 10;
        });
      }, 200);

      const response = await apiClient.uploadEventImage(event.id, file);

      clearInterval(progressInterval);
      setUploadProgress(100);

      setEvent({ ...event, image_url: response.image_url });
      toast.success('Image uploaded successfully!');

      // Reset progress after a brief delay
      setTimeout(() => setUploadProgress(0), 1000);
    } catch (error) {
      toast.error('Failed to upload image');
      console.error(error);
      setUploadProgress(0);
    } finally {
      setImageUploading(false);
    }
  };

  const handleOpenTicketDialog = () => {
    if (event) {
      if (!saleStartDate) {
        const now = new Date();
        setSaleStartDate(now.toISOString().split('T')[0]);
        setSaleStartTime({
          hour: now.getHours().toString().padStart(2, '0'),
          minute: now.getMinutes().toString().padStart(2, '0')
        });
      }

      if (!saleEndDate && event.end_date) {
        const endDate = new Date(event.end_date);
        setSaleEndDate(endDate.toISOString().split('T')[0]);
        setSaleEndTime({
          hour: '23',
          minute: '59'
        });
      }
    }
    setTicketDialogOpen(true);
  };

  const handleCreateTicketType = async () => {
    if (!event) return;

    if (!ticketForm.sale_start || !ticketForm.sale_end) {
      toast.error('Please select sale start and end dates');
      return;
    }

    try {
      const ticketData = {
        name: ticketForm.name,
        description: ticketForm.description,
        price: parseFloat(ticketForm.price),
        quantity: parseInt(ticketForm.quantity),
        max_per_order: parseInt(ticketForm.max_per_order),
        sale_start: new Date(ticketForm.sale_start).toISOString(),
        sale_end: new Date(ticketForm.sale_end).toISOString(),
      };

      await apiClient.createTicketType(event.id, ticketData);
      toast.success('Ticket type created successfully!');
      setTicketDialogOpen(false);
      loadEvent(event.id);

      setTicketForm({
        name: '',
        description: '',
        price: '',
        quantity: '',
        max_per_order: '10',
        sale_start: '',
        sale_end: '',
      });
      setSaleStartDate('');
      setSaleEndDate('');
    } catch (error) {
      toast.error('Failed to create ticket type');
      console.error(error);
    }
  };

  const handleSubmitForReview = async () => {
    if (!event) return;

    try {
      setSubmitting(true);
      await apiClient.submitEventForReview(event.id);
      toast.success('Event submitted for review!');
      loadEvent(event.id);
    } catch (error) {
      toast.error('Failed to submit event');
      console.error(error);
    } finally {
      setSubmitting(false);
    }
  };

  const handlePublish = async () => {
    if (!event) return;

    try {
      setPublishing(true);
      await apiClient.publishEvent(event.id);
      toast.success('Event published successfully!');
      loadEvent(event.id);
    } catch (error) {
      toast.error('Failed to publish event');
      console.error(error);
    } finally {
      setPublishing(false);
    }
  };

  const getStatusColor = (status: string) => {
    switch (status) {
      case 'draft': return 'secondary';
      case 'pending': return 'default';
      case 'approved': return 'outline';
      case 'published': return 'default';
      case 'rejected': return 'destructive';
      case 'cancelled': return 'destructive';
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

  if (loading) {
    return (
      <DashboardLayout>
        <div className="p-8">
          <Skeleton className="h-8 w-64 mb-6" />
          <Skeleton className="h-96 w-full" />
        </div>
      </DashboardLayout>
    );
  }

  if (!event) {
    return (
      <DashboardLayout>
        <div className="p-8">
          <div className="text-center py-12">
            <p className="text-muted-foreground text-lg">Event not found</p>
          </div>
        </div>
      </DashboardLayout>
    );
  }

  return (
    <DashboardLayout>
      <div className="p-8">
        {/* Header Section */}
        <div className="mb-8">
          <div className="flex flex-col md:flex-row justify-between items-start md:items-center gap-4">
            <div>
              <div className="flex items-center gap-3 mb-2">
                <h1 className="text-3xl font-bold tracking-tight">{event.title}</h1>
                <Badge variant={getStatusColor(event.status)} className="capitalize">
                  {event.status}
                </Badge>
              </div>
              <div className="flex items-center gap-4 text-muted-foreground text-sm">
                <div className="flex items-center">
                  <Calendar className="h-4 w-4 mr-1.5" />
                  {formatDate(event.start_date)}
                </div>
                <div className="flex items-center">
                  <MapPin className="h-4 w-4 mr-1.5" />
                  {event.venue}, {event.city}
                </div>
              </div>
            </div>

            <div className="flex gap-3">
              {(event.status === 'draft' || event.status === 'rejected') && (
                <Button variant="outline" onClick={() => router.push(`/organizer/events/${event.id}/edit`)}>
                  Edit Event
                </Button>
              )}

              {event.status === 'published' && (
                <Button variant="outline" asChild>
                  <a href={`/events/${event.id}`} target="_blank">
                    <Eye className="h-4 w-4 mr-2" />
                    View Public Page
                  </a>
                </Button>
              )}

              {event.status === 'draft' && (
                <Button onClick={handleSubmitForReview} disabled={submitting}>
                  <Send className="h-4 w-4 mr-2" />
                  {submitting ? 'Submitting...' : 'Submit for Review'}
                </Button>
              )}

              {event.status === 'approved' && (
                <Button onClick={handlePublish} disabled={publishing}>
                  <Sparkles className="h-4 w-4 mr-2" />
                  {publishing ? 'Publishing...' : 'Publish Event'}
                </Button>
              )}
            </div>
          </div>
        </div>

        {/* Tabs */}
        <Tabs defaultValue="overview" className="space-y-6">
          <TabsList className="bg-muted/50">
            <TabsTrigger value="overview">Overview</TabsTrigger>
            <TabsTrigger value="tickets">Tickets</TabsTrigger>
            <TabsTrigger value="stats">Statistics</TabsTrigger>
          </TabsList>

          <TabsContent value="overview" className="space-y-6">
            <div className="grid lg:grid-cols-3 gap-6">
              {/* Event Details */}
              <div className="lg:col-span-2 space-y-6">
                <Card>
                  <CardHeader>
                    <CardTitle>Event Details</CardTitle>
                  </CardHeader>
                  <CardContent className="space-y-6">
                    <div>
                      <h3 className="text-sm font-medium text-muted-foreground mb-2">Description</h3>
                      <p className="text-foreground whitespace-pre-wrap leading-relaxed">{event.description}</p>
                    </div>

                    <div className="grid md:grid-cols-2 gap-6 pt-4 border-t">
                      <div>
                        <h3 className="text-sm font-medium text-muted-foreground mb-3">Location</h3>
                        <div className="flex items-start gap-3">
                          <div className="w-10 h-10 rounded-lg bg-primary/10 flex items-center justify-center flex-shrink-0">
                            <MapPin className="h-5 w-5 text-primary" />
                          </div>
                          <div>
                            <p className="font-medium">{event.venue}</p>
                            <p className="text-sm text-muted-foreground">{event.address}</p>
                            <p className="text-sm text-muted-foreground">{event.city}, {event.country}</p>
                          </div>
                        </div>
                      </div>

                      <div>
                        <h3 className="text-sm font-medium text-muted-foreground mb-3">Date & Time</h3>
                        <div className="flex items-start gap-3">
                          <div className="w-10 h-10 rounded-lg bg-primary/10 flex items-center justify-center flex-shrink-0">
                            <Calendar className="h-5 w-5 text-primary" />
                          </div>
                          <div>
                            <p className="font-medium">{formatDate(event.start_date)}</p>
                            <p className="text-sm text-muted-foreground">to {formatDate(event.end_date)}</p>
                          </div>
                        </div>
                      </div>
                    </div>
                  </CardContent>
                </Card>
              </div>

              {/* Image Upload */}
              <div className="lg:col-span-1">
                <Card>
                  <CardHeader>
                    <CardTitle>Event Image</CardTitle>
                    <CardDescription>Cover image for your event</CardDescription>
                  </CardHeader>
                  <CardContent>
                    <div className="space-y-4">
                      {event.image_url ? (
                        <div className="relative aspect-video w-full overflow-hidden rounded-lg border">
                          <img
                            src={`${process.env.NEXT_PUBLIC_API_URL?.replace('/api/v1', '') || 'http://localhost:8080'}${event.image_url}`}
                            alt={event.title}
                            className="object-cover w-full h-full"
                          />
                        </div>
                      ) : (
                        <div className="aspect-video bg-muted border-2 border-dashed rounded-lg flex flex-col items-center justify-center">
                          <Upload className="h-8 w-8 text-muted-foreground/40 mb-2" />
                          <p className="text-sm text-muted-foreground">No image uploaded</p>
                        </div>
                      )}

                      <div className="space-y-3">
                        <Input
                          type="file"
                          accept="image/*"
                          onChange={handleImageUpload}
                          disabled={imageUploading}
                          className="hidden"
                          id="image-upload"
                        />
                        <Label htmlFor="image-upload" className="cursor-pointer w-full block">
                          <Button variant="outline" className="w-full" disabled={imageUploading} asChild>
                            <span>
                              <Upload className="h-4 w-4 mr-2" />
                              {imageUploading ? 'Uploading...' : 'Upload Image'}
                            </span>
                          </Button>
                        </Label>

                        {/* Progress Bar */}
                        {imageUploading && (
                          <div className="space-y-2">
                            <Progress value={uploadProgress} className="h-2" />
                            <p className="text-xs text-center text-muted-foreground">
                              {uploadProgress}% uploaded
                            </p>
                          </div>
                        )}
                      </div>
                    </div>
                  </CardContent>
                </Card>
              </div>
            </div>
          </TabsContent>

          <TabsContent value="tickets" className="space-y-6">
            <Card>
              <CardHeader className="flex flex-row items-center justify-between">
                <div>
                  <CardTitle>Ticket Types</CardTitle>
                  <CardDescription>Manage ticket inventory and pricing</CardDescription>
                </div>
                <Button onClick={handleOpenTicketDialog}>
                  <Plus className="h-4 w-4 mr-2" />
                  Add Ticket Type
                </Button>
              </CardHeader>
              <CardContent>
                {!event.ticket_types || event.ticket_types.length === 0 ? (
                  <div className="text-center py-16 border-2 border-dashed rounded-xl bg-muted/10">
                    <div className="w-16 h-16 rounded-full bg-muted flex items-center justify-center mx-auto mb-4">
                      <Ticket className="h-8 w-8 text-muted-foreground" />
                    </div>
                    <h3 className="text-lg font-semibold mb-2">No tickets created</h3>
                    <p className="text-muted-foreground mb-6 max-w-sm mx-auto">
                      Start selling by creating your first ticket type for this event.
                    </p>
                    <Button onClick={handleOpenTicketDialog}>
                      <Plus className="h-4 w-4 mr-2" />
                      Create Ticket Type
                    </Button>
                  </div>
                ) : (
                  <div className="grid md:grid-cols-2 lg:grid-cols-3 gap-6">
                    {event.ticket_types.map((ticket) => (
                      <Card key={ticket.id} className="overflow-hidden border-2 hover:border-primary/50 transition-colors">
                        <div className="h-2 bg-gradient-to-r from-primary to-primary/60" />
                        <CardHeader className="pb-3">
                          <div className="flex justify-between items-start">
                            <div>
                              <CardTitle className="text-lg">{ticket.name}</CardTitle>
                              <CardDescription className="line-clamp-1">{ticket.description}</CardDescription>
                            </div>
                            <Badge variant="secondary" className="font-bold text-primary">
                              {formatAmount(ticket.price)}
                            </Badge>
                          </div>
                        </CardHeader>
                        <CardContent>
                          <div className="space-y-3">
                            <div className="flex justify-between text-sm">
                              <span className="text-muted-foreground">Sold / Total</span>
                              <span className="font-medium">{ticket.sold} / {ticket.quantity}</span>
                            </div>
                            <div className="w-full bg-muted h-2 rounded-full overflow-hidden">
                              <div
                                className="bg-primary h-full transition-all duration-500"
                                style={{ width: `${(ticket.sold / ticket.quantity) * 100}%` }}
                              />
                            </div>
                            <div className="flex justify-between text-xs text-muted-foreground">
                              <span>Max {ticket.max_per_order} per order</span>
                              <span>{ticket.quantity - ticket.sold} remaining</span>
                            </div>
                          </div>
                        </CardContent>
                      </Card>
                    ))}
                  </div>
                )}
              </CardContent>
            </Card>
          </TabsContent>

          <TabsContent value="stats" className="space-y-6">
            {stats ? (
              <div className="grid md:grid-cols-2 lg:grid-cols-4 gap-6">
                <Card>
                  <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
                    <CardTitle className="text-sm font-medium text-muted-foreground">Tickets Sold</CardTitle>
                    <Ticket className="h-4 w-4 text-primary" />
                  </CardHeader>
                  <CardContent>
                    <div className="text-3xl font-bold">{stats.total_tickets_sold}</div>
                    <p className="text-xs text-muted-foreground mt-1">Total tickets distributed</p>
                  </CardContent>
                </Card>

                <Card>
                  <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
                    <CardTitle className="text-sm font-medium text-muted-foreground">Total Revenue</CardTitle>
                    <DollarSign className="h-4 w-4 text-primary" />
                  </CardHeader>
                  <CardContent>
                    <div className="text-3xl font-bold">{formatAmount(stats.total_revenue)}</div>
                    <p className="text-xs text-muted-foreground mt-1">Gross revenue from sales</p>
                  </CardContent>
                </Card>

                <Card>
                  <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
                    <CardTitle className="text-sm font-medium text-muted-foreground">Net Revenue</CardTitle>
                    <TrendingUp className="h-4 w-4 text-green-500" />
                  </CardHeader>
                  <CardContent>
                    <div className="text-3xl font-bold">{formatAmount(stats.net_revenue)}</div>
                    <p className="text-xs text-muted-foreground mt-1">After platform fees</p>
                  </CardContent>
                </Card>

                <Card>
                  <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
                    <CardTitle className="text-sm font-medium text-muted-foreground">Check-ins</CardTitle>
                    <Users className="h-4 w-4 text-blue-500" />
                  </CardHeader>
                  <CardContent>
                    <div className="text-3xl font-bold">{stats.checked_in_tickets}</div>
                    <p className="text-xs text-muted-foreground mt-1">Attendees checked in</p>
                  </CardContent>
                </Card>
              </div>
            ) : (
              <Card>
                <CardContent className="py-16 text-center">
                  <div className="w-16 h-16 rounded-full bg-muted flex items-center justify-center mx-auto mb-4">
                    <TrendingUp className="h-8 w-8 text-muted-foreground" />
                  </div>
                  <h3 className="text-lg font-semibold mb-2">No statistics yet</h3>
                  <p className="text-muted-foreground">Statistics will appear here once you start selling tickets.</p>
                </CardContent>
              </Card>
            )}
          </TabsContent>
        </Tabs>
      </div>

      {/* Create Ticket Type Dialog */}
      <Dialog open={ticketDialogOpen} onOpenChange={setTicketDialogOpen}>
        <DialogContent className="max-w-2xl">
          <DialogHeader>
            <DialogTitle>Create Ticket Type</DialogTitle>
            <DialogDescription>Add a new ticket type for your event</DialogDescription>
          </DialogHeader>

          <div className="space-y-4 py-4">
            <div className="grid grid-cols-2 gap-4">
              <div className="space-y-2">
                <Label htmlFor="ticket-name">Name *</Label>
                <Input
                  id="ticket-name"
                  placeholder="e.g., General Admission"
                  value={ticketForm.name}
                  onChange={(e) => setTicketForm({ ...ticketForm, name: e.target.value })}
                />
              </div>

              <div className="space-y-2">
                <Label htmlFor="ticket-price">Price ({symbol}) *</Label>
                <Input
                  id="ticket-price"
                  type="number"
                  placeholder="5000"
                  value={ticketForm.price}
                  onChange={(e) => setTicketForm({ ...ticketForm, price: e.target.value })}
                />
              </div>
            </div>

            <div className="space-y-2">
              <Label htmlFor="ticket-description">Description</Label>
              <Textarea
                id="ticket-description"
                placeholder="Describe this ticket type..."
                value={ticketForm.description}
                onChange={(e) => setTicketForm({ ...ticketForm, description: e.target.value })}
              />
            </div>

            <div className="grid grid-cols-2 gap-4">
              <div className="space-y-2">
                <Label htmlFor="ticket-quantity">Quantity *</Label>
                <Input
                  id="ticket-quantity"
                  type="number"
                  placeholder="100"
                  value={ticketForm.quantity}
                  onChange={(e) => setTicketForm({ ...ticketForm, quantity: e.target.value })}
                />
              </div>

              <div className="space-y-2">
                <Label htmlFor="ticket-max">Max per order</Label>
                <Input
                  id="ticket-max"
                  type="number"
                  placeholder="10"
                  value={ticketForm.max_per_order}
                  onChange={(e) => setTicketForm({ ...ticketForm, max_per_order: e.target.value })}
                />
              </div>
            </div>

            <div className="grid grid-cols-2 gap-4">
              <div className="space-y-2">
                <Label>Sale starts *</Label>
                <Input
                  type="date"
                  value={saleStartDate}
                  onChange={(e) => setSaleStartDate(e.target.value)}
                  required
                />
              </div>

              <div className="space-y-2">
                <Label>Sale ends *</Label>
                <Input
                  type="date"
                  value={saleEndDate}
                  onChange={(e) => setSaleEndDate(e.target.value)}
                  required
                />
              </div>
            </div>
          </div>

          <DialogFooter>
            <Button variant="outline" onClick={() => setTicketDialogOpen(false)}>
              Cancel
            </Button>
            <Button onClick={handleCreateTicketType}>
              Create Ticket Type
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </DashboardLayout>
  );
}
