'use client';

import { useState, useEffect } from 'react';
import { useRouter } from 'next/navigation';
import { apiClient } from '@/lib/api-client';
import { ModeratorStats, Event } from '@/lib/types';
import { useAuth } from '@/lib/auth-context';
import { DashboardLayout } from '@/components/dashboard-layout';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { Skeleton } from '@/components/ui/skeleton';
import { Badge } from '@/components/ui/badge';
import { CheckCircle, XCircle, Clock, TrendingUp, BarChart3, Calendar, User } from 'lucide-react';
import { toast } from 'sonner';

export default function ModeratorStatsPage() {
  const router = useRouter();
  const { user, loading: authLoading } = useAuth();
  const [stats, setStats] = useState<ModeratorStats | null>(null);
  const [recentReviews, setRecentReviews] = useState<Event[]>([]);
  const [loading, setLoading] = useState(true);

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
      const [statsData, reviewsData] = await Promise.all([
        apiClient.getModeratorStats(),
        apiClient.getMyReviews(),
      ]);
      setStats(statsData);
      setRecentReviews(reviewsData.slice(0, 5)); // Get last 5 reviews
    } catch (error) {
      toast.error('Failed to load moderator statistics');
      console.error(error);
    } finally {
      setLoading(false);
    }
  };

  const formatDate = (dateString: string) => {
    return new Date(dateString).toLocaleDateString('en-US', {
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
      default: return 'outline';
    }
  };

  const getStatusIcon = (status: string) => {
    switch (status) {
      case 'approved':
      case 'published':
        return <CheckCircle className="h-4 w-4" />;
      case 'rejected':
        return <XCircle className="h-4 w-4" />;
      default:
        return <Clock className="h-4 w-4" />;
    }
  };

  const calculateApprovalRate = () => {
    if (!stats) return 0;
    const total = stats.approved_events + stats.rejected_events;
    if (total === 0) return 0;
    return ((stats.approved_events / total) * 100).toFixed(1);
  };



  if (authLoading || loading) {
    return (
      <DashboardLayout>
        <div className="container mx-auto px-4 py-8">
          <Skeleton className="h-8 w-64 mb-6" />
          <div className="grid md:grid-cols-2 lg:grid-cols-4 gap-6 mb-8">
            {[...Array(4)].map((_, i) => (
              <Skeleton key={i} className="h-32" />
            ))}
          </div>
          <div className="grid md:grid-cols-2 gap-6">
            <Skeleton className="h-96" />
            <Skeleton className="h-96" />
          </div>
        </div>
      </DashboardLayout>
    );
  }

  return (
    <DashboardLayout>
      <div className="container mx-auto px-4 py-8">
        <div className="flex items-center gap-2 mb-8">
          <BarChart3 className="h-8 w-8" />
          <h1 className="text-4xl font-bold">Moderation Statistics</h1>
        </div>

        {stats && (
          <>
            {/* Main Stats Cards */}
            <div className="grid md:grid-cols-2 lg:grid-cols-4 gap-6 mb-8">
              <Card>
                <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
                  <CardTitle className="text-sm font-medium">Pending Events</CardTitle>
                  <Clock className="h-4 w-4 text-muted-foreground" />
                </CardHeader>
                <CardContent>
                  <div className="text-2xl font-bold text-yellow-600">
                    {stats.pending_events}
                  </div>
                  <p className="text-xs text-muted-foreground">
                    Awaiting review
                  </p>
                </CardContent>
              </Card>

              <Card>
                <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
                  <CardTitle className="text-sm font-medium">Approved Events</CardTitle>
                  <CheckCircle className="h-4 w-4 text-muted-foreground" />
                </CardHeader>
                <CardContent>
                  <div className="text-2xl font-bold text-green-600">
                    {stats.approved_events}
                  </div>
                  <p className="text-xs text-muted-foreground">
                    Successfully approved
                  </p>
                </CardContent>
              </Card>

              <Card>
                <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
                  <CardTitle className="text-sm font-medium">Rejected Events</CardTitle>
                  <XCircle className="h-4 w-4 text-muted-foreground" />
                </CardHeader>
                <CardContent>
                  <div className="text-2xl font-bold text-red-600">
                    {stats.rejected_events}
                  </div>
                  <p className="text-xs text-muted-foreground">
                    Did not meet criteria
                  </p>
                </CardContent>
              </Card>

              <Card>
                <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
                  <CardTitle className="text-sm font-medium">Total Reviews</CardTitle>
                  <TrendingUp className="h-4 w-4 text-muted-foreground" />
                </CardHeader>
                <CardContent>
                  <div className="text-2xl font-bold">
                    {stats.my_reviews}
                  </div>
                  <p className="text-xs text-muted-foreground">
                    Events reviewed
                  </p>
                </CardContent>
              </Card>
            </div>

            {/* Performance Metrics */}
            <div className="grid md:grid-cols-3 gap-6 mb-8">
              <Card>
                <CardHeader>
                  <CardTitle>Approval Rate</CardTitle>
                  <CardDescription>
                    Percentage of events approved
                  </CardDescription>
                </CardHeader>
                <CardContent>
                  <div className="text-3xl font-bold text-green-600">
                    {calculateApprovalRate()}%
                  </div>
                  <p className="text-sm text-muted-foreground mt-2">
                    {stats.approved_events} approved out of {stats.approved_events + stats.rejected_events} reviewed
                  </p>
                </CardContent>
              </Card>

              <Card>
                <CardHeader>
                  <CardTitle>Review Workload</CardTitle>
                  <CardDescription>
                    Current pending vs completed
                  </CardDescription>
                </CardHeader>
                <CardContent>
                  <div className="space-y-2">
                    <div className="flex justify-between items-center">
                      <span className="text-sm">Pending:</span>
                      <span className="font-medium text-yellow-600">{stats.pending_events}</span>
                    </div>
                    <div className="flex justify-between items-center">
                      <span className="text-sm">Completed:</span>
                      <span className="font-medium text-green-600">{stats.my_reviews}</span>
                    </div>
                    <div className="w-full bg-gray-200 rounded-full h-2 mt-3">
                      <div 
                        className="bg-green-600 h-2 rounded-full" 
                        style={{ 
                          width: `${stats.my_reviews > 0 ? (stats.my_reviews / (stats.my_reviews + stats.pending_events)) * 100 : 0}%` 
                        }}
                      ></div>
                    </div>
                  </div>
                </CardContent>
              </Card>

              <Card>
                <CardHeader>
                  <CardTitle>Review Distribution</CardTitle>
                  <CardDescription>
                    Breakdown of review decisions
                  </CardDescription>
                </CardHeader>
                <CardContent>
                  <div className="space-y-3">
                    <div className="flex items-center justify-between">
                      <div className="flex items-center gap-2">
                        <div className="w-3 h-3 bg-green-500 rounded-full"></div>
                        <span className="text-sm">Approved</span>
                      </div>
                      <span className="font-medium">{stats.approved_events}</span>
                    </div>
                    <div className="flex items-center justify-between">
                      <div className="flex items-center gap-2">
                        <div className="w-3 h-3 bg-red-500 rounded-full"></div>
                        <span className="text-sm">Rejected</span>
                      </div>
                      <span className="font-medium">{stats.rejected_events}</span>
                    </div>
                    <div className="flex items-center justify-between">
                      <div className="flex items-center gap-2">
                        <div className="w-3 h-3 bg-yellow-500 rounded-full"></div>
                        <span className="text-sm">Pending</span>
                      </div>
                      <span className="font-medium">{stats.pending_events}</span>
                    </div>
                  </div>
                </CardContent>
              </Card>
            </div>

            {/* Recent Reviews */}
            <div className="grid md:grid-cols-2 gap-6">
              <Card>
                <CardHeader>
                  <CardTitle>Recent Reviews</CardTitle>
                  <CardDescription>
                    Your latest event reviews
                  </CardDescription>
                </CardHeader>
                <CardContent>
                  {recentReviews.length === 0 ? (
                    <div className="text-center py-8">
                      <CheckCircle className="h-12 w-12 mx-auto text-muted-foreground mb-2" />
                      <p className="text-muted-foreground">No reviews yet</p>
                    </div>
                  ) : (
                    <div className="space-y-4">
                      {recentReviews.map((event) => (
                        <div key={event.id} className="flex items-start justify-between p-3 border rounded-lg">
                          <div className="flex-1">
                            <h4 className="font-medium text-sm line-clamp-1">{event.title}</h4>
                            <div className="flex items-center gap-2 mt-1 text-xs text-muted-foreground">
                              <Calendar className="h-3 w-3" />
                              {formatDate(event.start_date)}
                              {event.organizer && (
                                <>
                                  <User className="h-3 w-3 ml-2" />
                                  {event.organizer.first_name} {event.organizer.last_name}
                                </>
                              )}
                            </div>
                          </div>
                          <Badge variant={getStatusColor(event.status)} className="capitalize ml-2">
                            <span className="flex items-center gap-1">
                              {getStatusIcon(event.status)}
                              {event.status}
                            </span>
                          </Badge>
                        </div>
                      ))}
                    </div>
                  )}
                </CardContent>
              </Card>

              <Card>
                <CardHeader>
                  <CardTitle>Moderation Guidelines</CardTitle>
                  <CardDescription>
                    Key criteria for event approval
                  </CardDescription>
                </CardHeader>
                <CardContent>
                  <div className="space-y-4">
                    <div className="p-3 bg-green-50 border border-green-200 rounded-lg">
                      <h4 className="font-medium text-green-900 mb-1">✓ Approve When</h4>
                      <ul className="text-sm text-green-700 space-y-1">
                        <li>• Event details are complete and accurate</li>
                        <li>• Venue and date information is valid</li>
                        <li>• Content follows community guidelines</li>
                        <li>• Ticket pricing is reasonable</li>
                      </ul>
                    </div>
                    
                    <div className="p-3 bg-red-50 border border-red-200 rounded-lg">
                      <h4 className="font-medium text-red-900 mb-1">✗ Reject When</h4>
                      <ul className="text-sm text-red-700 space-y-1">
                        <li>• Missing critical information</li>
                        <li>• Inappropriate or offensive content</li>
                        <li>• Suspicious or fraudulent activity</li>
                        <li>• Violates platform terms of service</li>
                      </ul>
                    </div>
                  </div>
                </CardContent>
              </Card>
            </div>

            {/* Quick Actions */}
            <div className="mt-8">
              <Card>
                <CardHeader>
                  <CardTitle>Quick Actions</CardTitle>
                  <CardDescription>
                    Common moderation tasks
                  </CardDescription>
                </CardHeader>
                <CardContent>
                  <div className="grid md:grid-cols-3 gap-4">
                    <div className="p-4 border rounded-lg text-center">
                      <Clock className="h-8 w-8 mx-auto text-yellow-500 mb-2" />
                      <h4 className="font-medium mb-1">Review Pending</h4>
                      <p className="text-sm text-muted-foreground mb-3">
                        {stats.pending_events} events waiting
                      </p>
                      <button 
                        onClick={() => router.push('/moderator/events/pending')}
                        className="text-sm text-blue-600 hover:underline"
                      >
                        Start Reviewing →
                      </button>
                    </div>
                    
                    <div className="p-4 border rounded-lg text-center">
                      <CheckCircle className="h-8 w-8 mx-auto text-green-500 mb-2" />
                      <h4 className="font-medium mb-1">View Reviews</h4>
                      <p className="text-sm text-muted-foreground mb-3">
                        {stats.my_reviews} completed reviews
                      </p>
                      <button 
                        onClick={() => router.push('/moderator/reviews')}
                        className="text-sm text-blue-600 hover:underline"
                      >
                        View History →
                      </button>
                    </div>
                    
                    <div className="p-4 border rounded-lg text-center">
                      <BarChart3 className="h-8 w-8 mx-auto text-blue-500 mb-2" />
                      <h4 className="font-medium mb-1">Dashboard</h4>
                      <p className="text-sm text-muted-foreground mb-3">
                        Overview and summary
                      </p>
                      <button 
                        onClick={() => router.push('/moderator')}
                        className="text-sm text-blue-600 hover:underline"
                      >
                        Go to Dashboard →
                      </button>
                    </div>
                  </div>
                </CardContent>
              </Card>
            </div>
          </>
        )}
      </div>
    </DashboardLayout>
  );
}
