'use client';

import { useState, useEffect } from 'react';
import { useRouter } from 'next/navigation';
import { DashboardLayout } from '@/components/dashboard-layout';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { Skeleton } from '@/components/ui/skeleton';
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select';
import { apiClient } from '@/lib/api-client';
import { useAuth } from '@/lib/auth-context';
import { BarChart, Bar, LineChart, Line, PieChart, Pie, Cell, XAxis, YAxis, CartesianGrid, Tooltip, Legend, ResponsiveContainer } from 'recharts';
import { TrendingUp, TrendingDown, DollarSign, Ticket, Users, Calendar } from 'lucide-react';

interface EventStats {
  total_tickets_sold: number;
  total_revenue: number;
  net_revenue: number;
  checked_in_tickets: number;
  ticket_type_breakdown: Array<{
    ticket_type_id: string;
    name: string;
    sold: number;
    remaining: number;
    gross_revenue: number;
    capacity_pct: number;
  }>;
  daily_revenue: Array<{
    date: string;
    revenue: number;
  }>;
  check_in_stats: {
    total_confirmed: number;
    checked_in: number;
    check_in_rate: number;
  };
}

interface Event {
  id: string;
  title: string;
}

const COLORS = ['#3b82f6', '#10b981', '#f59e0b', '#ef4444', '#8b5cf6', '#ec4899'];

export default function OrganizerStatsPage() {
  const router = useRouter();
  const { user, loading: authLoading } = useAuth();
  const [events, setEvents] = useState<Event[]>([]);
  const [selectedEventId, setSelectedEventId] = useState<string>('');
  const [stats, setStats] = useState<EventStats | null>(null);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    if (!authLoading && user?.role !== 'organizer') {
      router.push('/');
    }
  }, [user, authLoading, router]);

  useEffect(() => {
    loadEvents();
  }, []);

  useEffect(() => {
    if (selectedEventId) {
      loadStats(selectedEventId);
    }
  }, [selectedEventId]);

  const loadEvents = async () => {
    try {
      const response = await apiClient.getOrganizerEvents();
      setEvents(response.events || []);
      if (response.events && response.events.length > 0) {
        setSelectedEventId(response.events[0].id);
      }
    } catch (error) {
      console.error('Failed to load events:', error);
    } finally {
      setLoading(false);
    }
  };

  const loadStats = async (eventId: string) => {
    setLoading(true);
    try {
      const data = await apiClient.getEventStats(eventId);
      setStats(data);
    } catch (error) {
      console.error('Failed to load stats:', error);
    } finally {
      setLoading(false);
    }
  };

  if (authLoading || loading) {
    return (
      <DashboardLayout>
        <div className="space-y-6">
          <Skeleton className="h-12 w-64" />
          <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-4">
            {[...Array(4)].map((_, i) => (
              <Skeleton key={i} className="h-32" />
            ))}
          </div>
        </div>
      </DashboardLayout>
    );
  }

  if (!events.length) {
    return (
      <DashboardLayout>
        <div className="text-center py-12">
          <Calendar className="h-12 w-12 mx-auto text-muted-foreground mb-4" />
          <h2 className="text-2xl font-bold mb-2">No Events Yet</h2>
          <p className="text-muted-foreground mb-6">Create your first event to see statistics</p>
        </div>
      </DashboardLayout>
    );
  }

  const checkInRate = stats?.check_in_stats.check_in_rate || 0;
  const revenueGrowth = stats?.daily_revenue && stats.daily_revenue.length > 1
    ? ((stats.daily_revenue[stats.daily_revenue.length - 1].revenue - stats.daily_revenue[0].revenue) / stats.daily_revenue[0].revenue) * 100
    : 0;

  return (
    <DashboardLayout>
      <div className="space-y-6">
        <div className="flex items-center justify-between">
          <div>
            <h1 className="text-3xl font-bold">Event Statistics</h1>
            <p className="text-muted-foreground">Detailed analytics for your events</p>
          </div>
          <Select value={selectedEventId} onValueChange={setSelectedEventId}>
            <SelectTrigger className="w-[300px]">
              <SelectValue placeholder="Select an event" />
            </SelectTrigger>
            <SelectContent>
              {events.map((event) => (
                <SelectItem key={event.id} value={event.id}>
                  {event.title}
                </SelectItem>
              ))}
            </SelectContent>
          </Select>
        </div>

        {stats && (
          <>
            {/* Key Metrics */}
            <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-4">
              <Card>
                <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
                  <CardTitle className="text-sm font-medium">Total Revenue</CardTitle>
                  <DollarSign className="h-4 w-4 text-muted-foreground" />
                </CardHeader>
                <CardContent>
                  <div className="text-2xl font-bold">KES {stats.total_revenue.toLocaleString()}</div>
                  <p className="text-xs text-muted-foreground">
                    Net: KES {stats.net_revenue.toLocaleString()}
                  </p>
                </CardContent>
              </Card>

              <Card>
                <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
                  <CardTitle className="text-sm font-medium">Tickets Sold</CardTitle>
                  <Ticket className="h-4 w-4 text-muted-foreground" />
                </CardHeader>
                <CardContent>
                  <div className="text-2xl font-bold">{stats.total_tickets_sold}</div>
                  <p className="text-xs text-muted-foreground">
                    {stats.ticket_type_breakdown.length} ticket types
                  </p>
                </CardContent>
              </Card>

              <Card>
                <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
                  <CardTitle className="text-sm font-medium">Check-in Rate</CardTitle>
                  <Users className="h-4 w-4 text-muted-foreground" />
                </CardHeader>
                <CardContent>
                  <div className="text-2xl font-bold">{(checkInRate * 100).toFixed(1)}%</div>
                  <p className="text-xs text-muted-foreground">
                    {stats.checked_in_tickets} / {stats.check_in_stats.total_confirmed} checked in
                  </p>
                </CardContent>
              </Card>

              <Card>
                <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
                  <CardTitle className="text-sm font-medium">Revenue Growth</CardTitle>
                  {revenueGrowth >= 0 ? (
                    <TrendingUp className="h-4 w-4 text-green-500" />
                  ) : (
                    <TrendingDown className="h-4 w-4 text-red-500" />
                  )}
                </CardHeader>
                <CardContent>
                  <div className={`text-2xl font-bold ${revenueGrowth >= 0 ? 'text-green-500' : 'text-red-500'}`}>
                    {revenueGrowth >= 0 ? '+' : ''}{revenueGrowth.toFixed(1)}%
                  </div>
                  <p className="text-xs text-muted-foreground">Since first sale</p>
                </CardContent>
              </Card>
            </div>

            {/* Charts Row 1 */}
            <div className="grid gap-4 md:grid-cols-2">
              {/* Ticket Type Breakdown - Bar Chart */}
              <Card>
                <CardHeader>
                  <CardTitle>Ticket Type Performance</CardTitle>
                  <CardDescription>Sales by ticket type</CardDescription>
                </CardHeader>
                <CardContent>
                  <ResponsiveContainer width="100%" height={300}>
                    <BarChart data={stats.ticket_type_breakdown}>
                      <CartesianGrid strokeDasharray="3 3" />
                      <XAxis dataKey="name" />
                      <YAxis />
                      <Tooltip />
                      <Legend />
                      <Bar dataKey="sold" fill="#3b82f6" name="Sold" />
                      <Bar dataKey="remaining" fill="#94a3b8" name="Remaining" />
                    </BarChart>
                  </ResponsiveContainer>
                </CardContent>
              </Card>

              {/* Revenue by Ticket Type - Pie Chart */}
              <Card>
                <CardHeader>
                  <CardTitle>Revenue Distribution</CardTitle>
                  <CardDescription>Revenue by ticket type</CardDescription>
                </CardHeader>
                <CardContent>
                  <ResponsiveContainer width="100%" height={300}>
                    <PieChart>
                      <Pie
                        data={stats.ticket_type_breakdown}
                        dataKey="gross_revenue"
                        nameKey="name"
                        cx="50%"
                        cy="50%"
                        outerRadius={100}
                        label={(props) => {
                          const entry = stats.ticket_type_breakdown[props.index];
                          return `${entry.name}: KES ${entry.gross_revenue.toLocaleString()}`;
                        }}
                      >
                        {stats.ticket_type_breakdown.map((_, index) => (
                          <Cell key={`cell-${index}`} fill={COLORS[index % COLORS.length]} />
                        ))}
                      </Pie>
                      <Tooltip />
                    </PieChart>
                  </ResponsiveContainer>
                </CardContent>
              </Card>
            </div>

            {/* Daily Revenue Chart */}
            <Card>
              <CardHeader>
                <CardTitle>Daily Revenue Trend</CardTitle>
                <CardDescription>Revenue over time</CardDescription>
              </CardHeader>
              <CardContent>
                <ResponsiveContainer width="100%" height={300}>
                  <LineChart data={stats.daily_revenue}>
                    <CartesianGrid strokeDasharray="3 3" />
                    <XAxis dataKey="date" />
                    <YAxis />
                    <Tooltip />
                    <Legend />
                    <Line type="monotone" dataKey="revenue" stroke="#3b82f6" strokeWidth={2} name="Revenue (KES)" />
                  </LineChart>
                </ResponsiveContainer>
              </CardContent>
            </Card>

            {/* Ticket Type Details Table */}
            <Card>
              <CardHeader>
                <CardTitle>Detailed Breakdown</CardTitle>
                <CardDescription>Complete ticket type statistics</CardDescription>
              </CardHeader>
              <CardContent>
                <div className="overflow-x-auto">
                  <table className="w-full">
                    <thead>
                      <tr className="border-b">
                        <th className="text-left p-2">Ticket Type</th>
                        <th className="text-right p-2">Sold</th>
                        <th className="text-right p-2">Remaining</th>
                        <th className="text-right p-2">Capacity</th>
                        <th className="text-right p-2">Revenue</th>
                      </tr>
                    </thead>
                    <tbody>
                      {stats.ticket_type_breakdown.map((type) => (
                        <tr key={type.ticket_type_id} className="border-b">
                          <td className="p-2 font-medium">{type.name}</td>
                          <td className="text-right p-2">{type.sold}</td>
                          <td className="text-right p-2">{type.remaining}</td>
                          <td className="text-right p-2">
                            <div className="flex items-center justify-end gap-2">
                              <div className="w-20 h-2 bg-gray-200 rounded-full overflow-hidden">
                                <div
                                  className="h-full bg-blue-500"
                                  style={{ width: `${type.capacity_pct}%` }}
                                />
                              </div>
                              <span className="text-sm">{type.capacity_pct.toFixed(0)}%</span>
                            </div>
                          </td>
                          <td className="text-right p-2">KES {type.gross_revenue.toLocaleString()}</td>
                        </tr>
                      ))}
                    </tbody>
                  </table>
                </div>
              </CardContent>
            </Card>
          </>
        )}
      </div>
    </DashboardLayout>
  );
}
