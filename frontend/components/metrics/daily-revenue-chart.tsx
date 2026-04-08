'use client';

import { DailyRevenue } from '@/lib/types';
import {
  LineChart,
  Line,
  XAxis,
  YAxis,
  CartesianGrid,
  Tooltip,
  ResponsiveContainer,
} from 'recharts';
import { useCurrency } from '@/lib/currency';

interface DailyRevenueChartProps {
  data: (DailyRevenue | { date: string; [key: string]: number | string })[];
  /** The key in each data object to plot. Defaults to "revenue". */
  dataKey?: string;
  /** Whether the Y-axis values are currency amounts. Defaults to true. */
  isCurrency?: boolean;
  /** Label shown in the tooltip. Defaults to "Revenue". */
  valueLabel?: string;
}

function formatShortDate(dateStr: string): string {
  const [, month, day] = dateStr.split('-');
  return `${month}/${day}`;
}

export function DailyRevenueChart({
  data,
  dataKey = 'revenue',
  isCurrency = true,
  valueLabel = 'Revenue',
}: DailyRevenueChartProps) {
  const { formatAmount } = useCurrency();

  if (!data || data.length === 0) {
    return (
      <p className="text-sm text-muted-foreground text-center py-6">
        No data available.
      </p>
    );
  }

  const chartData = data.map((d) => ({
    date: formatShortDate(d.date),
    // eslint-disable-next-line @typescript-eslint/no-explicit-any
    [dataKey]: (d as any)[dataKey],
  }));

  const formatY = (v: number) => (isCurrency ? formatAmount(v) : v.toLocaleString());
  // eslint-disable-next-line @typescript-eslint/no-explicit-any
  const formatTooltip = (value: any) =>
    [isCurrency ? formatAmount(value) : value.toLocaleString(), valueLabel] as [string, string];

  return (
    <ResponsiveContainer width="100%" height={260}>
      <LineChart data={chartData} margin={{ top: 4, right: 16, left: 0, bottom: 0 }}>
        <CartesianGrid strokeDasharray="3 3" className="stroke-muted" />
        <XAxis
          dataKey="date"
          tick={{ fontSize: 11 }}
          tickLine={false}
          axisLine={false}
          interval="preserveStartEnd"
        />
        <YAxis
          tick={{ fontSize: 11 }}
          tickLine={false}
          axisLine={false}
          tickFormatter={formatY}
          width={80}
        />
        <Tooltip
          formatter={formatTooltip}
          labelFormatter={(label) => `Date: ${label}`}
          contentStyle={{ fontSize: 12 }}
        />
        <Line
          type="monotone"
          dataKey={dataKey}
          stroke="hsl(var(--primary))"
          strokeWidth={2}
          dot={false}
          activeDot={{ r: 4 }}
        />
      </LineChart>
    </ResponsiveContainer>
  );
}
