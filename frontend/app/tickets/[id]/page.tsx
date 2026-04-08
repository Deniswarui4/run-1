'use client';

import { useState, useEffect } from 'react';
import { useParams, useRouter } from 'next/navigation';
import Link from 'next/link';
import { apiClient } from '@/lib/api-client';
import { Ticket } from '@/lib/types';
import { useAuth } from '@/lib/auth-context';
import { Navbar } from '@/components/navbar';
import { Footer } from '@/components/footer';
import { Button } from '@/components/ui/button';
import { Badge } from '@/components/ui/badge';
import { Skeleton } from '@/components/ui/skeleton';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { Separator } from '@/components/ui/separator';
import { Calendar, MapPin, Download, QrCode, ArrowLeft, Clock, User, Ticket as TicketIcon } from 'lucide-react';
import { toast } from 'sonner';
import { useCurrency } from '@/lib/currency';

export default function TicketDetailsPage() {
    const params = useParams();
    const router = useRouter();
    const { user, loading: authLoading } = useAuth();
    const { formatAmount } = useCurrency();
    const [ticket, setTicket] = useState<Ticket | null>(null);
    const [loading, setLoading] = useState(true);

    useEffect(() => {
        if (!authLoading) {
            if (!user) {
                router.push('/login');
            } else if (params.id) {
                loadTicket(params.id as string);
            }
        }
    }, [user, authLoading, params.id, router]);

    const loadTicket = async (id: string) => {
        try {
            setLoading(true);
            const data = await apiClient.getTicket(id);
            setTicket(data);
        } catch (error) {
            toast.error('Failed to load ticket details');
            console.error(error);
        } finally {
            setLoading(false);
        }
    };

    const handleDownloadTicket = async () => {
        if (!ticket) return;
        try {
            const blob = await apiClient.downloadTicketPDF(ticket.id);
            const url = window.URL.createObjectURL(blob);
            const a = document.createElement('a');
            a.href = url;
            a.download = `ticket-${ticket.ticket_number}.pdf`;
            document.body.appendChild(a);
            a.click();
            window.URL.revokeObjectURL(url);
            document.body.removeChild(a);
            toast.success('Ticket downloaded successfully');
        } catch (error) {
            toast.error('Failed to download ticket');
            console.error(error);
        }
    };

    const formatDate = (dateString: string) => {
        return new Date(dateString).toLocaleDateString('en-US', {
            weekday: 'long',
            year: 'numeric',
            month: 'long',
            day: 'numeric',
        });
    };

    const formatTime = (dateString: string) => {
        return new Date(dateString).toLocaleTimeString('en-US', {
            hour: '2-digit',
            minute: '2-digit',
        });
    };

    const getStatusColor = (status: string) => {
        switch (status) {
            case 'confirmed': return 'default';
            case 'used': return 'secondary';
            case 'cancelled': return 'destructive';
            default: return 'outline';
        }
    };

    if (authLoading || loading) {
        return (
            <div className="min-h-screen flex flex-col bg-background">
                <Navbar />
                <main className="flex-1 container mx-auto px-4 py-20">
                    <div className="max-w-3xl mx-auto">
                        <Skeleton className="h-8 w-32 mb-6" />
                        <Skeleton className="h-[500px] w-full rounded-xl" />
                    </div>
                </main>
            </div>
        );
    }

    if (!ticket) {
        return (
            <div className="min-h-screen flex flex-col bg-background">
                <Navbar />
                <main className="flex-1 container mx-auto px-4 py-20">
                    <div className="max-w-md mx-auto text-center">
                        <h2 className="text-2xl font-bold mb-2">Ticket Not Found</h2>
                        <p className="text-muted-foreground mb-6">The ticket you are looking for does not exist or you do not have permission to view it.</p>
                        <Button onClick={() => router.push('/my-tickets')}>Back to My Tickets</Button>
                    </div>
                </main>
                <Footer />
            </div>
        );
    }

    return (
        <div className="min-h-screen flex flex-col bg-background">
            <Navbar />

            <main className="flex-1 flex flex-col relative">
                {/* Background Decor */}
                <div className="absolute inset-0 bg-primary/5 -z-10" />
                <div className="absolute top-0 left-0 w-full h-96 bg-gradient-to-b from-primary/10 to-transparent -z-10" />

                <div className="container mx-auto px-4 py-12 md:py-24 flex-1 flex flex-col">
                    <div className="mb-8">
                        <Link href="/my-tickets" className="inline-flex items-center text-sm font-medium text-muted-foreground hover:text-primary transition-colors">
                            <ArrowLeft className="h-4 w-4 mr-2" />
                            Back to My Tickets
                        </Link>
                    </div>

                    <div className="max-w-5xl mx-auto w-full grid md:grid-cols-12 gap-8 items-start">
                        {/* Ticket Card - Main Section */}
                        <div className="md:col-span-8 animate-in fade-in slide-in-from-bottom-4 duration-500">
                            <div className="bg-card rounded-3xl shadow-xl border border-border/50 overflow-hidden relative group">
                                {/* Top Gradient Bar */}
                                <div className="h-3 w-full bg-gradient-to-r from-primary via-purple-500 to-blue-600" />

                                <div className="p-8 md:p-10">
                                    <div className="flex flex-col md:flex-row justify-between items-start gap-6 mb-8">
                                        <div>
                                            <Badge variant={getStatusColor(ticket.status)} className="mb-4 capitalize px-3 py-1 text-sm shadow-sm">
                                                {ticket.status}
                                            </Badge>
                                            <h1 className="text-3xl md:text-4xl font-bold text-foreground mb-3 tracking-tight">{ticket.event?.title}</h1>
                                            <div className="flex items-center text-muted-foreground text-lg">
                                                <MapPin className="h-5 w-5 mr-2 text-primary" />
                                                <span>{ticket.event?.venue}, {ticket.event?.city}</span>
                                            </div>
                                        </div>
                                        <div className="text-right hidden md:block">
                                            <p className="text-sm text-muted-foreground mb-1">Ticket ID</p>
                                            <p className="font-mono font-medium text-lg tracking-wider">{ticket.ticket_number}</p>
                                        </div>
                                    </div>

                                    <Separator className="my-8" />

                                    <div className="grid md:grid-cols-2 gap-10">
                                        <div className="space-y-8">
                                            <div>
                                                <p className="text-sm font-medium text-muted-foreground mb-2 uppercase tracking-wider">Date & Time</p>
                                                <div className="flex items-start">
                                                    <div className="bg-primary/10 p-3 rounded-xl mr-4">
                                                        <Calendar className="h-6 w-6 text-primary" />
                                                    </div>
                                                    <div>
                                                        <p className="font-bold text-lg">{ticket.event && formatDate(ticket.event.start_date)}</p>
                                                        <p className="text-muted-foreground">{ticket.event && formatTime(ticket.event.start_date)}</p>
                                                    </div>
                                                </div>
                                            </div>

                                            <div>
                                                <p className="text-sm font-medium text-muted-foreground mb-2 uppercase tracking-wider">Ticket Type</p>
                                                <div className="flex items-start">
                                                    <div className="bg-primary/10 p-3 rounded-xl mr-4">
                                                        <TicketIcon className="h-6 w-6 text-primary" />
                                                    </div>
                                                    <div>
                                                        <p className="font-bold text-lg">{ticket.ticket_type?.name}</p>
                                                        <p className="text-primary font-bold text-xl mt-1">{formatAmount(ticket.price)}</p>
                                                    </div>
                                                </div>
                                            </div>
                                        </div>

                                        <div className="flex flex-col items-center justify-center bg-muted/30 rounded-2xl p-6 border border-border/50">
                                            {ticket.status === 'confirmed' && ticket.qr_code_url ? (
                                                <>
                                                    <div className="bg-white p-4 rounded-xl shadow-sm mb-4">
                                                        <img
                                                            src={`${process.env.NEXT_PUBLIC_API_URL?.replace('/api/v1', '') || 'http://localhost:8080'}${ticket.qr_code_url}`}
                                                            alt="Ticket QR Code"
                                                            className="w-40 h-40 object-contain"
                                                        />
                                                    </div>
                                                    <p className="font-mono text-sm font-medium tracking-wider text-center mb-2">{ticket.ticket_number}</p>
                                                    <p className="text-xs text-muted-foreground text-center flex items-center">
                                                        <QrCode className="h-3 w-3 mr-1" />
                                                        Scan at entrance
                                                    </p>
                                                </>
                                            ) : (
                                                <div className="flex flex-col items-center justify-center h-full text-muted-foreground py-10">
                                                    <QrCode className="h-16 w-16 mb-4 opacity-20" />
                                                    <p className="font-medium">QR Code unavailable</p>
                                                </div>
                                            )}
                                        </div>
                                    </div>
                                </div>

                                {/* Ticket Footer Actions */}
                                <div className="bg-muted/20 p-6 md:p-8 border-t border-border/50 flex flex-col sm:flex-row gap-4">
                                    <Button className="flex-1 h-12 text-base shadow-lg shadow-primary/20" onClick={handleDownloadTicket} disabled={ticket.status !== 'confirmed'}>
                                        <Download className="h-5 w-5 mr-2" />
                                        Download PDF Ticket
                                    </Button>
                                    <Button variant="outline" className="flex-1 h-12 text-base bg-background" onClick={() => router.push(`/events/${ticket.event_id}`)}>
                                        View Event Details
                                    </Button>
                                </div>
                            </div>
                        </div>

                        {/* Sidebar */}
                        <div className="md:col-span-4 space-y-6 animate-in fade-in slide-in-from-bottom-8 duration-700 delay-100">
                            <Card className="border-border/50 shadow-md overflow-hidden">
                                <div className="h-2 bg-secondary" />
                                <CardHeader>
                                    <CardTitle className="text-lg">Organizer</CardTitle>
                                </CardHeader>
                                <CardContent>
                                    <div className="flex items-center gap-4 mb-6">
                                        <div className="h-12 w-12 rounded-full bg-gradient-to-br from-primary to-purple-600 flex items-center justify-center text-white font-bold text-lg shadow-md">
                                            {ticket.event?.organizer?.first_name?.[0]}
                                        </div>
                                        <div>
                                            <p className="font-bold text-lg">{ticket.event?.organizer?.first_name} {ticket.event?.organizer?.last_name}</p>
                                            <p className="text-xs font-medium text-muted-foreground uppercase tracking-wide">Event Host</p>
                                        </div>
                                    </div>
                                    <Button variant="outline" className="w-full">
                                        Contact Organizer
                                    </Button>
                                </CardContent>
                            </Card>

                            <Card className="border-border/50 shadow-md">
                                <CardHeader>
                                    <CardTitle className="text-lg">Purchase Summary</CardTitle>
                                </CardHeader>
                                <CardContent className="space-y-4">
                                    <div className="flex justify-between py-2 border-b border-border/50">
                                        <span className="text-muted-foreground">Purchase Date</span>
                                        <span className="font-medium">{formatDate(ticket.created_at)}</span>
                                    </div>
                                    <div className="flex justify-between py-2 border-b border-border/50">
                                        <span className="text-muted-foreground">Payment Method</span>
                                        <span className="font-medium flex items-center">
                                            <div className="h-2 w-2 rounded-full bg-green-500 mr-2" />
                                            Online
                                        </span>
                                    </div>
                                    <div className="flex justify-between py-2 pt-4">
                                        <span className="font-bold text-lg">Total</span>
                                        <span className="font-bold text-lg text-primary">{formatAmount(ticket.price)}</span>
                                    </div>
                                </CardContent>
                            </Card>

                            <div className="bg-primary/5 rounded-2xl p-6 border border-primary/10">
                                <h3 className="font-semibold mb-2 flex items-center">
                                    <User className="h-4 w-4 mr-2 text-primary" />
                                    Need Help?
                                </h3>
                                <p className="text-sm text-muted-foreground mb-4">
                                    Having trouble with your ticket? Check our help center or contact support.
                                </p>
                                <Button variant="link" className="p-0 h-auto text-primary font-medium">
                                    Visit Help Center &rarr;
                                </Button>
                            </div>
                        </div>
                    </div>
                </div>
            </main>
            <Footer />
        </div>
    );
}
