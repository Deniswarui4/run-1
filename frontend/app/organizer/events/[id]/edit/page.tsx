'use client';

import { useState, useEffect } from 'react';
import { useParams, useRouter } from 'next/navigation';
import { apiClient } from '@/lib/api-client';
import { Event, Category } from '@/lib/types';

import { DashboardLayout } from '@/components/dashboard-layout';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { Textarea } from '@/components/ui/textarea';
import { Alert, AlertDescription } from '@/components/ui/alert';
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select';
import { Calendar, MapPin, AlignLeft, Sparkles, Video } from 'lucide-react';
import { toast } from 'sonner';

export default function EditEventPage() {
  const params = useParams();
  const router = useRouter();

  const [loading, setLoading] = useState(true);
  const [saving, setSaving] = useState(false);
  const [error, setError] = useState('');
  const [event, setEvent] = useState<Event | null>(null);
  const [categories, setCategories] = useState<Category[]>([]);

  // State for form data
  const [formData, setFormData] = useState({
    title: '',
    description: '',
    category: '',
    venue: '',
    address: '',
    city: '',
    country: '',
    start_date: '',
    end_date: '',
    video_url: '',
  });

  // States for date/time components
  const [startDate, setStartDate] = useState('');
  const [startTime, setStartTime] = useState({ hour: '09', minute: '00' });
  const [endDate, setEndDate] = useState('');
  const [endTime, setEndTime] = useState({ hour: '17', minute: '00' });

  const hours = Array.from({ length: 24 }, (_, i) => i.toString().padStart(2, '0'));
  const minutes = Array.from({ length: 60 }, (_, i) => i.toString().padStart(2, '0'));

  useEffect(() => {
    if (params.id) {
      loadEvent(params.id as string);
      loadCategories();
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [params.id]);

  // Combine date and time into proper ISO format
  useEffect(() => {
    if (startDate) {
      const dateTime = new Date(`${startDate}T${startTime.hour}:${startTime.minute}:00`);
      setFormData(prev => ({ ...prev, start_date: dateTime.toISOString() }));
    }
  }, [startDate, startTime]);

  useEffect(() => {
    if (endDate) {
      const dateTime = new Date(`${endDate}T${endTime.hour}:${endTime.minute}:00`);
      setFormData(prev => ({ ...prev, end_date: dateTime.toISOString() }));
    }
  }, [endDate, endTime]);

  const loadEvent = async (id: string) => {
    try {
      setLoading(true);
      const data = await apiClient.getOrganizerEvent(id);
      
      if (data.status !== 'draft' && data.status !== 'rejected') {
        toast.error('Only draft or rejected events can be edited');
        router.push(`/organizer/events/${id}`);
        return;
      }

      setEvent(data);
      
      // Populate form
      setFormData({
        title: data.title,
        description: data.description,
        category: data.category,
        venue: data.venue,
        address: data.address || '',
        city: data.city,
        country: data.country || '',
        start_date: data.start_date,
        end_date: data.end_date,
        video_url: data.video_url || '',
      });

      // Parse dates
      const start = new Date(data.start_date);
      setStartDate(start.toISOString().split('T')[0]);
      setStartTime({
        hour: start.getHours().toString().padStart(2, '0'),
        minute: start.getMinutes().toString().padStart(2, '0'),
      });

      const end = new Date(data.end_date);
      setEndDate(end.toISOString().split('T')[0]);
      setEndTime({
        hour: end.getHours().toString().padStart(2, '0'),
        minute: end.getMinutes().toString().padStart(2, '0'),
      });
    } catch (error) {
      toast.error('Failed to load event');
      console.error(error);
    } finally {
      setLoading(false);
    }
  };

  const loadCategories = async () => {
    try {
      const data = await apiClient.getCategories();
      setCategories(data.filter(c => c.is_active));
    } catch (error) {
      console.error('Failed to load categories:', error);
    }
  };

  const getTodayDate = () => {
    const today = new Date();
    return today.toISOString().split('T')[0];
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setError('');

    if (!event) return;

    // Validate dates
    if (!formData.start_date || !formData.end_date) {
      setError('Please select both start and end dates and times.');
      return;
    }

    const startDateObj = new Date(formData.start_date);
    const endDateObj = new Date(formData.end_date);

    if (startDateObj >= endDateObj) {
      setError('End date must be after start date');
      return;
    }

    setSaving(true);

    try {
      await apiClient.updateEvent(event.id, formData);
      toast.success('Event updated successfully!');
      router.push(`/organizer/events/${event.id}`);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to update event');
      toast.error('Failed to update event');
    } finally {
      setSaving(false);
    }
  };

  if (loading) {
    return (
      <DashboardLayout>
        <div className="p-8 max-w-5xl mx-auto">
          <div className="animate-pulse space-y-4">
            <div className="h-8 bg-muted rounded w-1/3"></div>
            <div className="h-64 bg-muted rounded"></div>
          </div>
        </div>
      </DashboardLayout>
    );
  }

  if (!event) {
    return (
      <DashboardLayout>
        <div className="p-8 max-w-5xl mx-auto text-center">
          <h2 className="text-2xl font-bold mb-4">Event Not Found</h2>
          <Button onClick={() => router.push('/organizer')}>Back to Dashboard</Button>
        </div>
      </DashboardLayout>
    );
  }

  return (
    <DashboardLayout>
      <div className="p-8 max-w-5xl mx-auto">
        {/* Header */}
        <div className="mb-8">
          <div className="flex items-center gap-3 mb-2">
            <div className="w-12 h-12 rounded-xl bg-primary/10 flex items-center justify-center">
              <Sparkles className="h-6 w-6 text-primary" />
            </div>
            <div>
              <h1 className="text-3xl font-bold">Edit Event</h1>
              <p className="text-muted-foreground">
                Update your event details
              </p>
            </div>
          </div>
        </div>

        <form onSubmit={handleSubmit} className="space-y-6">
          {error && (
            <Alert variant="destructive">
              <AlertDescription>{error}</AlertDescription>
            </Alert>
          )}

          {/* Basic Information */}
          <Card>
            <CardHeader>
              <div className="flex items-center gap-2">
                <AlignLeft className="h-5 w-5 text-primary" />
                <CardTitle>Basic Information</CardTitle>
              </div>
              <CardDescription>Update your event details</CardDescription>
            </CardHeader>
            <CardContent className="space-y-4">
              <div className="space-y-2">
                <Label htmlFor="title">Event Title *</Label>
                <Input
                  id="title"
                  value={formData.title}
                  onChange={(e) => setFormData({ ...formData, title: e.target.value })}
                  required
                />
              </div>

              <div className="space-y-2">
                <Label htmlFor="description">Description *</Label>
                <Textarea
                  id="description"
                  rows={6}
                  value={formData.description}
                  onChange={(e) => setFormData({ ...formData, description: e.target.value })}
                  required
                />
              </div>

              <div className="space-y-2">
                <Label htmlFor="category">Category *</Label>
                {categories.length > 0 ? (
                  <Select
                    value={formData.category}
                    onValueChange={(value) => setFormData({ ...formData, category: value })}
                    required
                  >
                    <SelectTrigger>
                      <SelectValue />
                    </SelectTrigger>
                    <SelectContent>
                      {categories.map((cat) => (
                        <SelectItem key={cat.id} value={cat.name}>
                          {cat.name}
                        </SelectItem>
                      ))}
                    </SelectContent>
                  </Select>
                ) : (
                  <Input
                    id="category"
                    value={formData.category}
                    onChange={(e) => setFormData({ ...formData, category: e.target.value })}
                    required
                  />
                )}
              </div>

              <div className="space-y-2">
                <Label htmlFor="video_url">
                  <div className="flex items-center gap-2">
                    <Video className="h-4 w-4" />
                    Promotional Video (Optional)
                  </div>
                </Label>
                <Input
                  id="video_url"
                  type="url"
                  placeholder="Paste video URL (YouTube, Shorts, Vimeo, TikTok, Instagram, Facebook)"
                  value={formData.video_url}
                  onChange={(e) => setFormData({ ...formData, video_url: e.target.value })}
                />
                <p className="text-xs text-muted-foreground">
                  Supported: YouTube (including Shorts), Vimeo, TikTok, Instagram, Facebook
                </p>
              </div>
            </CardContent>
          </Card>

          {/* Date & Time */}
          <Card>
            <CardHeader>
              <div className="flex items-center gap-2">
                <Calendar className="h-5 w-5 text-primary" />
                <CardTitle>Date & Time</CardTitle>
              </div>
            </CardHeader>
            <CardContent className="space-y-6">
              <div className="grid md:grid-cols-2 gap-6">
                {/* Start Date & Time */}
                <div className="space-y-4">
                  <Label className="text-base font-semibold">Start Date & Time *</Label>
                  <div className="space-y-3">
                    <Input
                      type="date"
                      value={startDate}
                      onChange={(e) => setStartDate(e.target.value)}
                      min={getTodayDate()}
                      required
                    />
                    <div className="grid grid-cols-2 gap-2">
                      <Select
                        value={startTime.hour}
                        onValueChange={(value) => setStartTime({ ...startTime, hour: value })}
                      >
                        <SelectTrigger>
                          <SelectValue />
                        </SelectTrigger>
                        <SelectContent>
                          {hours.map((h) => (
                            <SelectItem key={h} value={h}>{h}</SelectItem>
                          ))}
                        </SelectContent>
                      </Select>
                      <Select
                        value={startTime.minute}
                        onValueChange={(value) => setStartTime({ ...startTime, minute: value })}
                      >
                        <SelectTrigger>
                          <SelectValue />
                        </SelectTrigger>
                        <SelectContent>
                          {minutes.filter((m) => parseInt(m) % 5 === 0).map((m) => (
                            <SelectItem key={m} value={m}>{m}</SelectItem>
                          ))}
                        </SelectContent>
                      </Select>
                    </div>
                  </div>
                </div>

                {/* End Date & Time */}
                <div className="space-y-4">
                  <Label className="text-base font-semibold">End Date & Time *</Label>
                  <div className="space-y-3">
                    <Input
                      type="date"
                      value={endDate}
                      onChange={(e) => setEndDate(e.target.value)}
                      min={startDate || getTodayDate()}
                      required
                    />
                    <div className="grid grid-cols-2 gap-2">
                      <Select
                        value={endTime.hour}
                        onValueChange={(value) => setEndTime({ ...endTime, hour: value })}
                      >
                        <SelectTrigger>
                          <SelectValue />
                        </SelectTrigger>
                        <SelectContent>
                          {hours.map((h) => (
                            <SelectItem key={h} value={h}>{h}</SelectItem>
                          ))}
                        </SelectContent>
                      </Select>
                      <Select
                        value={endTime.minute}
                        onValueChange={(value) => setEndTime({ ...endTime, minute: value })}
                      >
                        <SelectTrigger>
                          <SelectValue />
                        </SelectTrigger>
                        <SelectContent>
                          {minutes.filter((m) => parseInt(m) % 5 === 0).map((m) => (
                            <SelectItem key={m} value={m}>{m}</SelectItem>
                          ))}
                        </SelectContent>
                      </Select>
                    </div>
                  </div>
                </div>
              </div>
            </CardContent>
          </Card>

          {/* Location */}
          <Card>
            <CardHeader>
              <div className="flex items-center gap-2">
                <MapPin className="h-5 w-5 text-primary" />
                <CardTitle>Location</CardTitle>
              </div>
            </CardHeader>
            <CardContent className="space-y-4">
              <div className="space-y-2">
                <Label htmlFor="venue">Venue Name *</Label>
                <Input
                  id="venue"
                  value={formData.venue}
                  onChange={(e) => setFormData({ ...formData, venue: e.target.value })}
                  required
                />
              </div>

              <div className="space-y-2">
                <Label htmlFor="address">Street Address *</Label>
                <Input
                  id="address"
                  value={formData.address}
                  onChange={(e) => setFormData({ ...formData, address: e.target.value })}
                  required
                />
              </div>

              <div className="grid md:grid-cols-2 gap-4">
                <div className="space-y-2">
                  <Label htmlFor="city">City *</Label>
                  <Input
                    id="city"
                    value={formData.city}
                    onChange={(e) => setFormData({ ...formData, city: e.target.value })}
                    required
                  />
                </div>

                <div className="space-y-2">
                  <Label htmlFor="country">Country *</Label>
                  <Input
                    id="country"
                    value={formData.country}
                    onChange={(e) => setFormData({ ...formData, country: e.target.value })}
                    required
                  />
                </div>
              </div>
            </CardContent>
          </Card>

          {/* Action Buttons */}
          <div className="flex gap-4 pt-4">
            <Button
              type="button"
              variant="outline"
              onClick={() => router.push(`/organizer/events/${event.id}`)}
              className="flex-1"
              disabled={saving}
            >
              Cancel
            </Button>
            <Button type="submit" disabled={saving} className="flex-1">
              {saving ? 'Saving...' : 'Save Changes'}
            </Button>
          </div>
        </form>
      </div>
    </DashboardLayout>
  );
}
