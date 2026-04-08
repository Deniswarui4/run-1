'use client';

import { useState, useEffect } from 'react';
import { useParams, useRouter } from 'next/navigation';
import { apiClient } from '@/lib/api-client';
import { ExtendedEventStats } from '@/lib/types';
import { useAuth } from '@/lib/auth-context';
import { DashboardLayout } from '@/components/dashboard-layout';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { Skeleton } from '@/components/ui/skeleton';
import { Button } from '@/components/ui/button';
import { TicketTypeBreakdownTable } from '@/components/metrics/ticket-type-breakdown';
import { DailyRevenueChart } from '@/components/metrics/daily-revenue-chart';
import { ArrowLeft, Ticket, DollarSign, TrendingUp, Users, UserCheck } from 'lucide-react';
import { toast } from 'sonner';
import { useCurrency } from '@/lib/currency';

export default function EventStatsPage() {
  const params = useParams();
  const router = useRouter();
  const { user } = useAuth();
  const { formatAmount } = useCurrency();
  const [stats, setStats] = useState<ExtendedEventStats | null>(null);
  const [loading, setLoading] = useState(true);

  const eventId = params.id as string;

  useEffect(() => {
    if (eventId && user) {
      loadStats(eventId);
    }
  }, [eventId, user]);

  const loadStats = async (id: string) => {
    try {
      setLoading(true);
      const data = await apiClient.getEventStats(id);
      setStats(data);
    } catch {
      toast.error('Failed to load event statistics');
    } finally {
      setLoading(false);
    }
  };

  if (loading) {
    return (
      <DashboardLayout>
        <div className="p-8 space-y-6">
          <Skeleton className="h-8 w-48" />
          <div className="grid md:grid-cols-2 lg:grid-cols-4 gap-4">
            {Array.from({ length: 4 }).map((_, i) => (
              <Skeleton key={i} className="h-28" />
            ))}
          </div>
          <Skeleton className="h-72 w-full" />
          <Skeleton className="h-48 w-full" />
        </div>
      </DashboardLayout>
    );
  }

  if (!stats) {
    return (
      <DashboardLayout>
        <div className="p-8 text-center py-16">
          <p className="text-muted-foreground">No statistics available for this event.</p>
          <Button variant="outline" className="mt-4" onClick={() => router.back()}>
            <ArrowLeft className="h-4 w-4 mr-2" />
            Go Back
          </Button>
        </div>
      </DashboardLayout>
    );
  }

  const checkInRate = stats.check_in_stats
    ? (stats.check_in_stats.check_in_rate * 100).toFixed(1)
    : '0.0';

  return (
    <DashboardLayout>
      <div className="p-8 space-y-8">
        {/* Header */}
        <div className="flex items-center gap-4">
          <Button variant="ghost" size="icon" onClick={() => router.back()}>
            <ArrowLeft className="h-5 w-5" />
          </Button>
          <h1 className="text-2xl font-bold tracking-tight">Event Statistics</h1>
        </div>

        {/* Summary Cards */}
        <div className="grid md:grid-cols-2 lg:grid-cols-4 gap-4">
          <Card>
            <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
              <CardTitle className="text-sm font-medium text-muted-foreground">
                Tickets Sold
              </CardTitle>
              <Ticket className="h-4 w-4 text-primary" />
            </CardHeader>
            <CardContent>
              <div className="text-3xl font-bold">{stats.total_tickets_sold}</div>
              <p className="text-xs text-muted-foreground mt-1">Total confirmed tickets</p>
            </CardContent>
          </Card>

          <Card>
            <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
              <CardTitle className="text-sm font-medium text-muted-foreground">
                Gross Revenue
              </CardTitle>
              <DollarSign className="h-4 w-4 text-primary" />
            </CardHeader>
            <CardContent>
              <div className="text-3xl font-bold">{formatAmount(stats.total_revenue)}</div>
              <p className="text-xs text-muted-foreground mt-1">Before platform fees</p>
            </CardContent>
          </Card>

          <Card>
            <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
              <CardTitle className="text-sm font-medium text-muted-foreground">
                Net Revenue
              </CardTitle>
              <TrendingUp className="h-4 w-4 text-green-500" />
            </CardHeader>
            <CardContent>
              <div className="text-3xl font-bold">{formatAmount(stats.net_revenue)}</div>
              <p className="text-xs text-muted-foreground mt-1">After platform fees</p>
            </CardContent>
          </Card>

          <Card>
            <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
              <CardTitle className="text-sm font-medium text-muted-foreground">
                Check-in Rate
              </CardTitle>
              <UserCheck className="h-4 w-4 text-blue-500" />
            </CardHeader>
            <CardContent>
              <div className="text-3xl font-bold">{checkInRate}%</div>
              <p className="text-xs text-muted-foreground mt-1">
                {stats.check_in_stats?.checked_in ?? 0} /{' '}
                {stats.check_in_stats?.total_confirmed ?? 0} attendees
              </p>
            </CardContent>
          </Card>
        </div>

        {/* Check-in Rate Card */}
        {stats.check_in_stats && (
          <Card>
            <CardHeader>
              <CardTitle className="flex items-center gap-2">
                <Users className="h-5 w-5 text-blue-500" />
                Check-in Summary
              </CardTitle>
            </CardHeader>
            <CardContent className="flex flex-wrap gap-8">
              <div>
                <p className="text-sm text-muted-foreground">Total Confirmed</p>
                <p className="text-2xl font-bold">{stats.check_in_stats.total_confirmed}</p>
              </div>
              <div>
                <p className="text-sm text-muted-foreground">Checked In</p>
                <p className="text-2xl font-bold">{stats.check_in_stats.checked_in}</p>
              </div>
              <div>
                <p className="text-sm text-muted-foreground">Check-in Rate</p>
                <p className="text-2xl font-bold">{checkInRate}%</p>
              </div>
            </CardContent>
          </Card>
        )}

        {/* Daily Revenue Chart */}
        <Card>
          <CardHeader>
            <CardTitle>Daily Revenue</CardTitle>
          </CardHeader>
          <CardContent>
            <DailyRevenueChart data={stats.daily_revenue ?? []} />
          </CardContent>
        </Card>

        {/* Ticket Type Breakdown */}
        <Card>
          <CardHeader>
            <CardTitle>Ticket Type Breakdown</CardTitle>
          </CardHeader>
          <CardContent>
            <TicketTypeBreakdownTable data={stats.ticket_type_breakdown ?? []} />
          </CardContent>
        </Card>
      </div>
    </DashboardLayout>
  );
}
