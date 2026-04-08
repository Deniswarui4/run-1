'use client';

import { useState, useEffect } from 'react';
import { useRouter } from 'next/navigation';
import Link from 'next/link';
import { apiClient } from '@/lib/api-client';
import { OrganizerBalance, Event, EventStats } from '@/lib/types';
import { useAuth } from '@/lib/auth-context';
import { DashboardLayout } from '@/components/dashboard-layout';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import { Skeleton } from '@/components/ui/skeleton';
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@/components/ui/table';
import { DollarSign, TrendingUp, Clock, CreditCard, Calendar, Ticket } from 'lucide-react';
import { toast } from 'sonner';
import { useCurrency } from '@/lib/currency';

export default function OrganizerEarningsPage() {
  const router = useRouter();
  const { user, loading: authLoading } = useAuth();
  const { formatAmount } = useCurrency();
  const [balance, setBalance] = useState<OrganizerBalance | null>(null);
  const [events, setEvents] = useState<Event[]>([]);
  const [eventStats, setEventStats] = useState<Record<string, EventStats>>({});
  const [loading, setLoading] = useState(true);

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
      const [balanceData, eventsData] = await Promise.all([
        apiClient.getOrganizerBalance(),
        apiClient.getMyEvents('published'),
      ]);
      
      setBalance(balanceData);
      setEvents(eventsData);

      // Load stats for each event
      const statsPromises = eventsData.map(event => 
        apiClient.getEventStats(event.id).catch(() => null)
      );
      const statsResults = await Promise.all(statsPromises);
      
      const statsMap: Record<string, EventStats> = {};
      eventsData.forEach((event, index) => {
        if (statsResults[index]) {
          statsMap[event.id] = statsResults[index];
        }
      });
      setEventStats(statsMap);
    } catch (error) {
      toast.error('Failed to load earnings data');
      console.error(error);
    } finally {
      setLoading(false);
    }
  };

  // Remove the old formatAmount function since we're using the hook

  const formatDate = (dateString: string) => {
    return new Date(dateString).toLocaleDateString('en-US', {
      month: 'short',
      day: 'numeric',
      year: 'numeric',
    });
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
          <Skeleton className="h-96 w-full" />
        </div>
      </DashboardLayout>
    );
  }

  return (
    <DashboardLayout>
      <div className="container mx-auto px-4 py-8">
        <div className="flex justify-between items-center mb-8">
          <h1 className="text-4xl font-bold">Earnings</h1>
          <Link href="/organizer/withdrawals">
            <Button>
              <CreditCard className="h-4 w-4 mr-2" />
              Request Withdrawal
            </Button>
          </Link>
        </div>

        {/* Balance Cards */}
        {balance && (
          <div className="grid md:grid-cols-3 gap-6 mb-8">
            <Card>
              <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
                <CardTitle className="text-sm font-medium">Total Earnings</CardTitle>
                <DollarSign className="h-4 w-4 text-muted-foreground" />
              </CardHeader>
              <CardContent>
                <div className="text-2xl font-bold">
                  {formatAmount(balance.total_earnings)}
                </div>
                <p className="text-xs text-muted-foreground">
                  All time earnings
                </p>
              </CardContent>
            </Card>

            <Card>
              <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
                <CardTitle className="text-sm font-medium">Available Balance</CardTitle>
                <TrendingUp className="h-4 w-4 text-muted-foreground" />
              </CardHeader>
              <CardContent>
                <div className="text-2xl font-bold text-green-600">
                  {formatAmount(balance.available_balance)}
                </div>
                <p className="text-xs text-muted-foreground">
                  Ready for withdrawal
                </p>
              </CardContent>
            </Card>

            <Card>
              <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
                <CardTitle className="text-sm font-medium">Pending Balance</CardTitle>
                <Clock className="h-4 w-4 text-muted-foreground" />
              </CardHeader>
              <CardContent>
                <div className="text-2xl font-bold text-yellow-600">
                  {formatAmount(balance.pending_balance)}
                </div>
                <p className="text-xs text-muted-foreground">
                  Processing
                </p>
              </CardContent>
            </Card>
          </div>
        )}

        {/* Event Performance */}
        <Card>
          <CardHeader>
            <CardTitle>Event Performance</CardTitle>
            <CardDescription>
              Revenue breakdown by event
            </CardDescription>
          </CardHeader>
          <CardContent>
            {events.length === 0 ? (
              <div className="text-center py-8">
                <p className="text-muted-foreground">No published events yet</p>
              </div>
            ) : (
              <Table>
                <TableHeader>
                  <TableRow>
                    <TableHead>Event</TableHead>
                    <TableHead>Date</TableHead>
                    <TableHead>Tickets Sold</TableHead>
                    <TableHead>Total Revenue</TableHead>
                    <TableHead>Net Revenue</TableHead>
                    <TableHead>Check-ins</TableHead>
                  </TableRow>
                </TableHeader>
                <TableBody>
                  {events.map((event) => {
                    const stats = eventStats[event.id];
                    return (
                      <TableRow key={event.id}>
                        <TableCell>
                          <div>
                            <p className="font-medium">{event.title}</p>
                            <p className="text-sm text-muted-foreground">
                              {event.venue}, {event.city}
                            </p>
                          </div>
                        </TableCell>
                        <TableCell>
                          <div className="flex items-center">
                            <Calendar className="h-4 w-4 mr-2 text-muted-foreground" />
                            {formatDate(event.start_date)}
                          </div>
                        </TableCell>
                        <TableCell>
                          <div className="flex items-center">
                            <Ticket className="h-4 w-4 mr-2 text-muted-foreground" />
                            {stats?.total_tickets_sold || 0}
                          </div>
                        </TableCell>
                        <TableCell className="font-medium">
                          {stats ? formatAmount(stats.total_revenue) : formatAmount(0)}
                        </TableCell>
                        <TableCell className="font-medium text-green-600">
                          {stats ? formatAmount(stats.net_revenue) : formatAmount(0)}
                        </TableCell>
                        <TableCell>
                          {stats?.checked_in_tickets || 0}
                        </TableCell>
                      </TableRow>
                    );
                  })}
                </TableBody>
              </Table>
            )}
          </CardContent>
        </Card>

        {/* Quick Actions */}
        <div className="mt-8 grid md:grid-cols-2 gap-6">
          <Card>
            <CardHeader>
              <CardTitle>Quick Actions</CardTitle>
            </CardHeader>
            <CardContent className="space-y-3">
              <Link href="/organizer/withdrawals">
                <Button className="w-full" variant="outline">
                  <CreditCard className="h-4 w-4 mr-2" />
                  View Withdrawal History
                </Button>
              </Link>
              <Link href="/organizer/events">
                <Button className="w-full" variant="outline">
                  <Calendar className="h-4 w-4 mr-2" />
                  Manage Events
                </Button>
              </Link>
            </CardContent>
          </Card>

          <Card>
            <CardHeader>
              <CardTitle>Earnings Summary</CardTitle>
            </CardHeader>
            <CardContent>
              {balance && (
                <div className="space-y-2">
                  <div className="flex justify-between">
                    <span className="text-muted-foreground">Total Withdrawn:</span>
                    <span className="font-medium">{formatAmount(balance.withdrawn_amount)}</span>
                  </div>
                  <div className="flex justify-between">
                    <span className="text-muted-foreground">Pending:</span>
                    <span className="font-medium">{formatAmount(balance.pending_balance)}</span>
                  </div>
                  <div className="flex justify-between">
                    <span className="text-muted-foreground">Available:</span>
                    <span className="font-medium text-green-600">{formatAmount(balance.available_balance)}</span>
                  </div>
                </div>
              )}
            </CardContent>
          </Card>
        </div>
      </div>
    </DashboardLayout>
  );
}
