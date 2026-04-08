'use client';

import { useState, useEffect } from 'react';
import { useParams, useRouter } from 'next/navigation';
import { Scanner } from '@yudiel/react-qr-scanner';
import { apiClient } from '@/lib/api-client';
import { Ticket, Event } from '@/lib/types';
import { DashboardLayout } from '@/components/dashboard-layout';
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { toast } from 'sonner';
import { CheckCircle2, XCircle, AlertTriangle, QrCode, ArrowLeft, Search } from 'lucide-react';

export default function ScanTicketsPage() {
    const params = useParams();
    const router = useRouter();
    const eventId = params.id as string;

    const [event, setEvent] = useState<Event | null>(null);
    const [loading, setLoading] = useState(true);
    const [scanning, setScanning] = useState(true);
    const [manualInput, setManualInput] = useState('');
    const [lastScannedResult, setLastScannedResult] = useState<{
        status: 'success' | 'error' | 'warning';
        message: string;
        ticket?: Ticket;
    } | null>(null);

    useEffect(() => {
        loadEvent();
    }, [eventId]);

    const loadEvent = async () => {
        try {
            setLoading(true);
            const data = await apiClient.getOrganizerEvent(eventId);
            setEvent(data);
        } catch (error) {
            toast.error('Failed to load event details');
            console.error(error);
        } finally {
            setLoading(false);
        }
    };

    const handleScan = async (result: string) => {
        if (!result) return;

        // If the result is a URL, try to extract the ticket number or ID if possible
        // For now, let's assume the QR code contains the ticket number directly or a URL ending with it
        // Adjust this logic based on what your QR codes actually contain
        const ticketNumber = result;

        // Simple check if it's a URL
        if (result.startsWith('http')) {
            try {
                new URL(result);
                // Example: https://example.com/tickets/TKT-12345678
                // We might need to parse this. For now, let's assume the QR code IS the ticket number
                // or we need to extract it.
                // If your QR code is just the ticket number, this block is skipped.
            } catch {
                // Not a valid URL, treat as ticket number
            }
        }

        verifyTicket(ticketNumber);
    };

    const handleManualSubmit = (e: React.FormEvent) => {
        e.preventDefault();
        if (!manualInput.trim()) return;
        verifyTicket(manualInput.trim());
    };

    const verifyTicket = async (ticketNumber: string) => {
        // Prevent double scanning the same ticket immediately if it was just scanned
        if (lastScannedResult?.ticket?.ticket_number === ticketNumber && lastScannedResult.status === 'success') {
            return;
        }

        try {
            // Pause scanning while processing
            setScanning(false);

            const response = await apiClient.verifyTicket(eventId, ticketNumber);

            setLastScannedResult({
                status: 'success',
                message: 'Ticket Verified Successfully',
                ticket: response.ticket,
            });
            toast.success('Ticket verified!');
            setManualInput('');
        } catch (error) {
            console.error(error);

            // Handle specific error cases
            const errorMessage = error instanceof Error ? error.message : 'Invalid Ticket';

            if (errorMessage.includes('already used')) {
                setLastScannedResult({
                    status: 'warning',
                    message: 'Ticket Already Used',
                    // We might want to fetch the ticket details to show who used it, 
                    // but the error message might not contain the full ticket object unless we parse it from the API error response
                    // For now, just show the warning.
                });
                toast.warning('Ticket already used');
            } else {
                setLastScannedResult({
                    status: 'error',
                    message: errorMessage,
                });
                toast.error(errorMessage);
            }
        } finally {
            // Resume scanning after a short delay to allow viewing the result
            setTimeout(() => {
                setScanning(true);
            }, 2000);
        }
    };

    const resetScanner = () => {
        setLastScannedResult(null);
        setScanning(true);
        setManualInput('');
    };

    if (loading) {
        return (
            <DashboardLayout>
                <div className="container mx-auto px-4 py-8 flex justify-center">
                    <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-primary"></div>
                </div>
            </DashboardLayout>
        );
    }

    return (
        <DashboardLayout>
            <div className="container mx-auto px-4 py-8 max-w-4xl">
                <div className="flex items-center mb-8">
                    <Button variant="ghost" className="mr-4" onClick={() => router.back()}>
                        <ArrowLeft className="h-4 w-4 mr-2" />
                        Back
                    </Button>
                    <div>
                        <h1 className="text-3xl font-bold tracking-tight">Scan Tickets</h1>
                        <p className="text-muted-foreground">
                            {event?.title} • {event?.venue}
                        </p>
                    </div>
                </div>

                <div className="grid md:grid-cols-2 gap-8">
                    {/* Scanner Section */}
                    <Card className="overflow-hidden border-2 border-primary/20 shadow-lg">
                        <CardHeader className="bg-muted/30 pb-4">
                            <CardTitle className="flex items-center">
                                <QrCode className="h-5 w-5 mr-2 text-primary" />
                                QR Scanner
                            </CardTitle>
                            <CardDescription>Point camera at ticket QR code</CardDescription>
                        </CardHeader>
                        <CardContent className="p-0 relative aspect-square bg-black">
                            {scanning ? (
                                <Scanner
                                    onScan={(result) => {
                                        if (result && result.length > 0) {
                                            handleScan(result[0].rawValue);
                                        }
                                    }}
                                    onError={(error) => console.error(error)}
                                    components={{
                                        onOff: false,
                                        torch: false,
                                        zoom: false,
                                        finder: true,
                                    }}
                                    styles={{
                                        container: { width: '100%', height: '100%' },
                                        video: { width: '100%', height: '100%', objectFit: 'cover' }
                                    }}
                                />
                            ) : (
                                <div className="absolute inset-0 flex items-center justify-center bg-black/80 text-white p-6 text-center">
                                    <div>
                                        <p className="mb-4 text-lg">Processing...</p>
                                        <Button variant="secondary" onClick={() => setScanning(true)}>
                                            Resume Scanning
                                        </Button>
                                    </div>
                                </div>
                            )}
                        </CardContent>
                    </Card>

                    {/* Manual Entry & Results Section */}
                    <div className="space-y-6">
                        <Card>
                            <CardHeader>
                                <CardTitle>Manual Entry</CardTitle>
                                <CardDescription>Enter ticket number manually</CardDescription>
                            </CardHeader>
                            <CardContent>
                                <form onSubmit={handleManualSubmit} className="flex gap-2">
                                    <Input
                                        placeholder="TKT-12345678"
                                        value={manualInput}
                                        onChange={(e) => setManualInput(e.target.value)}
                                        className="font-mono uppercase"
                                    />
                                    <Button type="submit">
                                        <Search className="h-4 w-4 mr-2" />
                                        Verify
                                    </Button>
                                </form>
                            </CardContent>
                        </Card>

                        {/* Result Display */}
                        {lastScannedResult && (
                            <Card className={`border-2 ${lastScannedResult.status === 'success' ? 'border-green-500 bg-green-50/50 dark:bg-green-900/10' :
                                lastScannedResult.status === 'warning' ? 'border-yellow-500 bg-yellow-50/50 dark:bg-yellow-900/10' :
                                    'border-red-500 bg-red-50/50 dark:bg-red-900/10'
                                } animate-in fade-in slide-in-from-bottom-4 duration-300`}>
                                <CardContent className="p-6 text-center">
                                    <div className="mb-4 flex justify-center">
                                        {lastScannedResult.status === 'success' && <CheckCircle2 className="h-16 w-16 text-green-500" />}
                                        {lastScannedResult.status === 'warning' && <AlertTriangle className="h-16 w-16 text-yellow-500" />}
                                        {lastScannedResult.status === 'error' && <XCircle className="h-16 w-16 text-red-500" />}
                                    </div>

                                    <h3 className={`text-2xl font-bold mb-2 ${lastScannedResult.status === 'success' ? 'text-green-600' :
                                        lastScannedResult.status === 'warning' ? 'text-yellow-600' :
                                            'text-red-600'
                                        }`}>
                                        {lastScannedResult.message}
                                    </h3>

                                    {lastScannedResult.ticket && (
                                        <div className="mt-6 space-y-3 text-left bg-background/50 p-4 rounded-xl border border-border/50">
                                            <div className="flex justify-between">
                                                <span className="text-muted-foreground">Attendee</span>
                                                <span className="font-semibold">{lastScannedResult.ticket.attendee?.first_name} {lastScannedResult.ticket.attendee?.last_name}</span>
                                            </div>
                                            <div className="flex justify-between">
                                                <span className="text-muted-foreground">Ticket Type</span>
                                                <span className="font-semibold">{lastScannedResult.ticket.ticket_type?.name}</span>
                                            </div>
                                            <div className="flex justify-between">
                                                <span className="text-muted-foreground">Ticket Number</span>
                                                <span className="font-mono text-xs">{lastScannedResult.ticket.ticket_number}</span>
                                            </div>
                                            {lastScannedResult.ticket.checked_in_at && (
                                                <div className="flex justify-between text-yellow-600">
                                                    <span className="text-muted-foreground">Checked In</span>
                                                    <span className="font-mono text-xs">
                                                        {new Date(lastScannedResult.ticket.checked_in_at).toLocaleTimeString()}
                                                    </span>
                                                </div>
                                            )}
                                        </div>
                                    )}

                                    <Button
                                        className="mt-6 w-full"
                                        variant={lastScannedResult.status === 'success' ? 'default' : 'secondary'}
                                        onClick={resetScanner}
                                    >
                                        Scan Next Ticket
                                    </Button>
                                </CardContent>
                            </Card>
                        )}

                        {!lastScannedResult && (
                            <div className="flex flex-col items-center justify-center h-64 text-muted-foreground border-2 border-dashed border-border rounded-xl bg-muted/10">
                                <QrCode className="h-12 w-12 mb-4 opacity-20" />
                                <p>Ready to scan tickets</p>
                            </div>
                        )}
                    </div>
                </div>
            </div>
        </DashboardLayout>
    );
}
