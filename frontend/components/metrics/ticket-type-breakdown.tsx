'use client';

import { TicketTypeBreakdown } from '@/lib/types';
import { useCurrency } from '@/lib/currency';
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table';
import { Progress } from '@/components/ui/progress';

interface TicketTypeBreakdownTableProps {
  data: TicketTypeBreakdown[];
}

export function TicketTypeBreakdownTable({ data }: TicketTypeBreakdownTableProps) {
  const { formatAmount } = useCurrency();

  if (!data || data.length === 0) {
    return (
      <p className="text-sm text-muted-foreground text-center py-6">
        No ticket type data available.
      </p>
    );
  }

  return (
    <Table>
      <TableHeader>
        <TableRow>
          <TableHead>Ticket Type</TableHead>
          <TableHead className="text-right">Sold</TableHead>
          <TableHead className="text-right">Remaining</TableHead>
          <TableHead className="text-right">Gross Revenue</TableHead>
          <TableHead className="w-40">Capacity</TableHead>
        </TableRow>
      </TableHeader>
      <TableBody>
        {data.map((row) => (
          <TableRow key={row.ticket_type_id}>
            <TableCell className="font-medium">{row.name}</TableCell>
            <TableCell className="text-right">{row.sold}</TableCell>
            <TableCell className="text-right">{row.remaining}</TableCell>
            <TableCell className="text-right">{formatAmount(row.gross_revenue)}</TableCell>
            <TableCell>
              <div className="flex items-center gap-2">
                <Progress value={row.capacity_pct} className="h-2 flex-1" />
                <span className="text-xs text-muted-foreground w-10 text-right">
                  {row.capacity_pct.toFixed(1)}%
                </span>
              </div>
            </TableCell>
          </TableRow>
        ))}
      </TableBody>
    </Table>
  );
}
