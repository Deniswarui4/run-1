'use client';

import { useState } from 'react';
import { apiClient } from '@/lib/api-client';
import { GuestTicket } from '@/lib/types';
import { Navbar } from '@/components/navbar';
import { Footer } from '@/components/footer';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { Badge } from '@/components/ui/badge';
import { Ticket, Search, QrCode } from 'lucide-react';

export default function GuestTicketLookupPage() {
  const [email, setEmail] = useState('');
  const [reference, setReference] = useState('');
  const [tickets, setTickets] = useState<GuestTicket[] | null>(null);
  const [error, setError] = useState<string | null>(null);
  const [loading, setLoading] = useState(false);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setError(null);
    setTickets(null);

    if (!email.trim() || !reference.trim()) {
      setError('Please enter both your email address and order reference.');
      return;
    }

    try {
      setLoading(true);
      const result = await apiClient.lookupGuestTickets(email.trim(), reference.trim());
      if (!result.tickets || result.tickets.length === 0) {
        setError('No tickets found for the provided email and order reference. Please check your details and try again.');
      } else {
        setTickets(result.tickets);
      }
    } catch (err) {
      const message = err instanceof Error ? err.message : 'An unexpected error occurred.';
      if (message.toLowerCase().includes('not found') || message.includes('404')) {
        setError('No tickets found for the provided email and order reference. Please check your details and try again.');
      } else {
        setError(`Could not retrieve your tickets: ${message}`);
      }
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="min-h-screen flex flex-col bg-background">
      <Navbar />
      <main className="flex-1 container max-w-2xl mx-auto px-4 py-12">
        <div className="mb-8 text-center">
          <div className="flex justify-center mb-3">
            <Ticket className="h-10 w-10 text-primary" />
          </div>
          <h1 className="text-3xl font-bold">Find Your Tickets</h1>
          <p className="text-muted-foreground mt-2">
            Enter the email address you used at checkout and your order reference to retrieve your tickets.
          </p>
        </div>

        <Card>
          <CardHeader>
            <CardTitle className="text-lg">Ticket Lookup</CardTitle>
          </CardHeader>
          <CardContent>
            <form onSubmit={handleSubmit} className="space-y-4">
              <div className="space-y-2">
                <Label htmlFor="email">Email Address</Label>
                <Input
                  id="email"
                  type="email"
                  placeholder="you@example.com"
                  value={email}
                  onChange={(e) => setEmail(e.target.value)}
                  required
                  autoComplete="email"
                />
              </div>
              <div className="space-y-2">
                <Label htmlFor="reference">Order Reference</Label>
                <Input
                  id="reference"
                  type="text"
                  placeholder="e.g. TXN-ABC123"
                  value={reference}
                  onChange={(e) => setReference(e.target.value)}
                  required
                  autoComplete="off"
                />
              </div>
              {error && (
                <p className="text-sm text-destructive bg-destructive/10 rounded-md px-3 py-2">
                  {error}
                </p>
              )}
              <Button type="submit" className="w-full" disabled={loading}>
                <Search className="h-4 w-4 mr-2" />
                {loading ? 'Searching...' : 'Find Tickets'}
              </Button>
            </form>
          </CardContent>
        </Card>

        {tickets && tickets.length > 0 && (
          <div className="mt-8 space-y-4">
            <h2 className="text-xl font-semibold">Your Tickets ({tickets.length})</h2>
            {tickets.map((ticket, index) => (
              <Card key={ticket.ticket_number ?? index}>
                <CardContent className="pt-6">
                  <div className="flex items-start justify-between gap-4">
                    <div className="space-y-1 flex-1 min-w-0">
                      <p className="font-semibold text-base truncate">{ticket.event_name}</p>
                      <div className="flex items-center gap-2 flex-wrap">
                        <Badge variant="secondary">{ticket.ticket_type}</Badge>
                        <span className="text-sm text-muted-foreground font-mono">
                          #{ticket.ticket_number}
                        </span>
                      </div>
                    </div>
                    {ticket.qr_code_url && (
                      <a
                        href={ticket.qr_code_url}
                        target="_blank"
                        rel="noopener noreferrer"
                        className="flex-shrink-0 flex flex-col items-center gap-1 text-xs text-muted-foreground hover:text-primary transition-colors"
                        aria-label={`View QR code for ticket ${ticket.ticket_number}`}
                      >
                        <QrCode className="h-8 w-8" />
                        <span>QR Code</span>
                      </a>
                    )}
                  </div>
                </CardContent>
              </Card>
            ))}
          </div>
        )}
      </main>
      <Footer />
    </div>
  );
}
