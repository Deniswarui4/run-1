'use client';

import { useState, useEffect } from 'react';
import { useRouter } from 'next/navigation';
import { apiClient } from '@/lib/api-client';
import { Category } from '@/lib/types';
import { useAuth } from '@/lib/auth-context';
import { DashboardLayout } from '@/components/dashboard-layout';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { Textarea } from '@/components/ui/textarea';
import { Alert, AlertDescription } from '@/components/ui/alert';
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select';
import { Calendar, MapPin, AlignLeft, Clock, Sparkles } from 'lucide-react';
import { toast } from 'sonner';

export default function CreateEventPage() {
  const router = useRouter();
  const { user } = useAuth();
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState('');
  const [categories, setCategories] = useState<Category[]>([]);
  const [loadingCategories, setLoadingCategories] = useState(true);

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

  // New states for date/time components - Initialize with current time
  const now = new Date();
  const [startDate, setStartDate] = useState('');
  const [startTime, setStartTime] = useState({
    hour: now.getHours().toString().padStart(2, '0'),
    minute: now.getMinutes().toString().padStart(2, '0')
  });
  const [endDate, setEndDate] = useState('');
  const [endTime, setEndTime] = useState({
    hour: (now.getHours() + 3).toString().padStart(2, '0'), // Default to 3 hours later
    minute: now.getMinutes().toString().padStart(2, '0')
  });

  const hours = Array.from({ length: 24 }, (_, i) => i.toString().padStart(2, '0'));
  const minutes = Array.from({ length: 60 }, (_, i) => i.toString().padStart(2, '0'));

  // Get today's date in YYYY-MM-DD format for min date
  const getTodayDate = () => {
    const today = new Date();
    return today.toISOString().split('T')[0];
  };

  // Filter available hours for end time based on start date/time
  const getAvailableEndHours = () => {
    if (!startDate || !endDate) return hours;

    // If same day, only show hours after start hour
    if (startDate === endDate) {
      const startHour = parseInt(startTime.hour);
      return hours.filter(h => parseInt(h) > startHour);
    }

    return hours;
  };

  // Filter available minutes for end time based on start date/time
  const getAvailableEndMinutes = () => {
    if (!startDate || !endDate) return minutes.filter((m) => parseInt(m) % 5 === 0);

    // If same day and same hour, only show minutes after start minute
    if (startDate === endDate && startTime.hour === endTime.hour) {
      const startMinute = parseInt(startTime.minute);
      return minutes.filter((m) => parseInt(m) % 5 === 0 && parseInt(m) > startMinute);
    }

    return minutes.filter((m) => parseInt(m) % 5 === 0);
  };

  useEffect(() => {
    loadCategories();
  }, []);

  // Combine date and time into proper ISO format with local timezone
  useEffect(() => {
    if (startDate) {
      // Create a date object in local timezone
      const dateTime = new Date(`${startDate}T${startTime.hour}:${startTime.minute}:00`);
      // Convert to ISO string (includes timezone offset)
      setFormData(prev => ({ ...prev, start_date: dateTime.toISOString() }));
    }
  }, [startDate, startTime]);

  useEffect(() => {
    if (endDate) {
      // Create a date object in local timezone
      const dateTime = new Date(`${endDate}T${endTime.hour}:${endTime.minute}:00`);
      // Convert to ISO string (includes timezone offset)
      setFormData(prev => ({ ...prev, end_date: dateTime.toISOString() }));
    }
  }, [endDate, endTime]);

  const loadCategories = async () => {
    try {
      const data = await apiClient.getCategories();
      setCategories(data.filter(c => c.is_active));
    } catch (error) {
      console.error('Failed to load categories:', error);
      toast.error('Failed to load categories');
    } finally {
      setLoadingCategories(false);
    }
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setError('');

    if (!user || (user.role !== 'organizer' && user.role !== 'admin')) {
      setError('You must be an organizer to create events');
      return;
    }

    // Validate dates
    if (!formData.start_date || !formData.end_date) {
      setError('Please select both start and end dates and times.');
      return;
    }

    const startDateObj = new Date(formData.start_date);
    const endDateObj = new Date(formData.end_date);
    const now = new Date();

    // Check for invalid dates
    if (isNaN(startDateObj.getTime()) || isNaN(endDateObj.getTime())) {
      setError('Invalid date format. Please use the date picker to select dates.');
      return;
    }

    if (startDateObj >= endDateObj) {
      setError('End date must be after start date');
      return;
    }

    // Allow events that start today or in the future (date-only comparison)
    const todayStart = new Date(now.getFullYear(), now.getMonth(), now.getDate());
    const eventStart = new Date(startDateObj.getFullYear(), startDateObj.getMonth(), startDateObj.getDate());

    if (eventStart < todayStart) {
      setError(`Event cannot start in the past. Start date is ${startDateObj.toLocaleDateString()}, today is ${now.toLocaleDateString()}`);
      return;
    }

    setLoading(true);

    try {
      const event = await apiClient.createEvent(formData);
      toast.success('Event created successfully!');
      router.push(`/organizer/events/${event.id}`);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to create event');
      toast.error('Failed to create event');
    } finally {
      setLoading(false);
    }
  };

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
              <h1 className="text-3xl font-bold">Create New Event</h1>
              <p className="text-muted-foreground">
                Fill in the details to create your amazing event
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
              <CardDescription>Tell us about your event</CardDescription>
            </CardHeader>
            <CardContent className="space-y-4">
              <div className="space-y-2">
                <Label htmlFor="title">Event Title *</Label>
                <Input
                  id="title"
                  placeholder="e.g., Summer Music Festival 2025"
                  value={formData.title}
                  onChange={(e) => setFormData({ ...formData, title: e.target.value })}
                  required
                  className="text-lg"
                />
              </div>

              <div className="space-y-2">
                <Label htmlFor="description">Description *</Label>
                <Textarea
                  id="description"
                  placeholder="Describe your event in detail... What makes it special? What can attendees expect?"
                  rows={6}
                  value={formData.description}
                  onChange={(e) => setFormData({ ...formData, description: e.target.value })}
                  required
                />
                <p className="text-xs text-muted-foreground">
                  Provide a compelling description that helps attendees understand your event
                </p>
              </div>

              <div className="space-y-2">
                <Label htmlFor="category">Category *</Label>
                {loadingCategories ? (
                  <div className="h-10 w-full rounded-md border border-input bg-muted animate-pulse" />
                ) : categories.length > 0 ? (
                  <Select
                    value={formData.category}
                    onValueChange={(value) => setFormData({ ...formData, category: value })}
                    required
                  >
                    <SelectTrigger>
                      <SelectValue placeholder="Select a category" />
                    </SelectTrigger>
                    <SelectContent>
                      {categories.map((cat) => (
                        <SelectItem key={cat.id} value={cat.name}>
                          <div className="flex items-center gap-2">
                            <div
                              className="w-3 h-3 rounded-full"
                              style={{ backgroundColor: cat.color || '#3B82F6' }}
                            />
                            <span>{cat.name}</span>
                          </div>
                        </SelectItem>
                      ))}
                    </SelectContent>
                  </Select>
                ) : (
                  <Input
                    id="category"
                    placeholder="e.g., Music, Conference, Sports"
                    value={formData.category}
                    onChange={(e) => setFormData({ ...formData, category: e.target.value })}
                    required
                  />
                )}
              </div>

              <div className="space-y-2">
                <Label htmlFor="video_url">Promotional Video (Optional)</Label>
                <Input
                  id="video_url"
                  type="url"
                  placeholder="Paste video URL (YouTube, Shorts, Vimeo, TikTok, Instagram, Facebook)"
                  value={formData.video_url}
                  onChange={(e) => setFormData({ ...formData, video_url: e.target.value })}
                />
                <p className="text-xs text-muted-foreground">
                  Add a promotional video to showcase your event. Supported: YouTube (including Shorts), Vimeo, TikTok, Instagram, Facebook
                </p>
                {formData.video_url && (
                  <div className="mt-2 p-2 bg-muted/50 rounded-md">
                    <p className="text-xs text-muted-foreground">Preview will be available after creating the event</p>
                  </div>
                )}
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
              <CardDescription>When will your event take place?</CardDescription>
            </CardHeader>
            <CardContent className="space-y-6">
              <div className="grid md:grid-cols-2 gap-6">
                {/* Start Date & Time */}
                <div className="space-y-4">
                  <div className="flex items-center gap-2">
                    <Clock className="h-4 w-4 text-muted-foreground" />
                    <Label className="text-base font-semibold">Start Date & Time *</Label>
                  </div>
                  <div className="space-y-3">
                    <div className="space-y-2">
                      <Label htmlFor="start-date" className="text-sm text-muted-foreground">Date</Label>
                      <Input
                        id="start-date"
                        type="date"
                        value={startDate}
                        onChange={(e) => setStartDate(e.target.value)}
                        min={getTodayDate()}
                        required
                      />
                    </div>
                    <div className="grid grid-cols-2 gap-2">
                      <div className="space-y-2">
                        <Label htmlFor="start-hour" className="text-sm text-muted-foreground">Hour</Label>
                        <Select
                          value={startTime.hour}
                          onValueChange={(value) => setStartTime({ ...startTime, hour: value })}
                        >
                          <SelectTrigger id="start-hour">
                            <SelectValue />
                          </SelectTrigger>
                          <SelectContent>
                            {hours.map((h) => (
                              <SelectItem key={h} value={h}>{h}</SelectItem>
                            ))}
                          </SelectContent>
                        </Select>
                      </div>
                      <div className="space-y-2">
                        <Label htmlFor="start-minute" className="text-sm text-muted-foreground">Minute</Label>
                        <Select
                          value={startTime.minute}
                          onValueChange={(value) => setStartTime({ ...startTime, minute: value })}
                        >
                          <SelectTrigger id="start-minute">
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

                {/* End Date & Time */}
                <div className="space-y-4">
                  <div className="flex items-center gap-2">
                    <Clock className="h-4 w-4 text-muted-foreground" />
                    <Label className="text-base font-semibold">End Date & Time *</Label>
                  </div>
                  <div className="space-y-3">
                    <div className="space-y-2">
                      <Label htmlFor="end-date" className="text-sm text-muted-foreground">Date</Label>
                      <Input
                        id="end-date"
                        type="date"
                        value={endDate}
                        onChange={(e) => setEndDate(e.target.value)}
                        min={startDate || getTodayDate()}
                        required
                      />
                    </div>
                    <div className="grid grid-cols-2 gap-2">
                      <div className="space-y-2">
                        <Label htmlFor="end-hour" className="text-sm text-muted-foreground">Hour</Label>
                        <Select
                          value={endTime.hour}
                          onValueChange={(value) => setEndTime({ ...endTime, hour: value })}
                        >
                          <SelectTrigger id="end-hour">
                            <SelectValue />
                          </SelectTrigger>
                          <SelectContent>
                            {getAvailableEndHours().map((h) => (
                              <SelectItem key={h} value={h}>{h}</SelectItem>
                            ))}
                          </SelectContent>
                        </Select>
                      </div>
                      <div className="space-y-2">
                        <Label htmlFor="end-minute" className="text-sm text-muted-foreground">Minute</Label>
                        <Select
                          value={endTime.minute}
                          onValueChange={(value) => setEndTime({ ...endTime, minute: value })}
                        >
                          <SelectTrigger id="end-minute">
                            <SelectValue />
                          </SelectTrigger>
                          <SelectContent>
                            {getAvailableEndMinutes().map((m) => (
                              <SelectItem key={m} value={m}>{m}</SelectItem>
                            ))}
                          </SelectContent>
                        </Select>
                      </div>
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
              <CardDescription>Where will your event be held?</CardDescription>
            </CardHeader>
            <CardContent className="space-y-4">
              <div className="space-y-2">
                <Label htmlFor="venue">Venue Name *</Label>
                <Input
                  id="venue"
                  placeholder="e.g., Madison Square Garden"
                  value={formData.venue}
                  onChange={(e) => setFormData({ ...formData, venue: e.target.value })}
                  required
                />
              </div>

              <div className="space-y-2">
                <Label htmlFor="address">Street Address *</Label>
                <Input
                  id="address"
                  placeholder="e.g., 4 Pennsylvania Plaza"
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
                    placeholder="e.g., Lagos"
                    value={formData.city}
                    onChange={(e) => setFormData({ ...formData, city: e.target.value })}
                    required
                  />
                </div>

                <div className="space-y-2">
                  <Label htmlFor="country">Country *</Label>
                  <Input
                    id="country"
                    placeholder="e.g., Nigeria"
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
              onClick={() => router.push('/organizer')}
              className="flex-1"
              disabled={loading}
            >
              Cancel
            </Button>
            <Button type="submit" disabled={loading} className="flex-1">
              {loading ? 'Creating Event...' : 'Create Event'}
            </Button>
          </div>

          <p className="text-sm text-muted-foreground text-center">
            After creating your event, you can add ticket types and upload images
          </p>
        </form>
      </div>
    </DashboardLayout>
  );
}
