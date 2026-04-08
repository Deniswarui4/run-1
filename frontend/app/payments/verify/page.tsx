'use client';

import { Suspense, useEffect, useState } from 'react';
import { useSearchParams } from 'next/navigation';
import { apiClient } from '@/lib/api-client';
import { Navbar } from '@/components/navbar';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import { Skeleton } from '@/components/ui/skeleton';
import { CheckCircle, XCircle, Loader2 } from 'lucide-react';
import Link from 'next/link';

function VerifyContent() {
  const searchParams = useSearchParams();
  const [status, setStatus] = useState<'loading' | 'success' | 'failed'>('loading');
  const [message, setMessage] = useState('');

  useEffect(() => {
    const reference = searchParams.get('reference');
    if (reference) {
      verifyPayment(reference);
    } else {
      setStatus('failed');
      setMessage('No payment reference provided');
    }
  }, [searchParams]);

  const verifyPayment = async (reference: string) => {
    try {
      const response = await apiClient.verifyPayment(reference);
      setStatus('success');
      setMessage(response.message);
    } catch (error) {
      setStatus('failed');
      setMessage(error instanceof Error ? error.message : 'Payment verification failed');
    }
  };

  return (
    <main className="flex-1 container mx-auto px-4 py-8">
      <div className="max-w-2xl mx-auto">
        {status === 'loading' && (
          <Card>
            <CardHeader>
              <div className="flex items-center justify-center mb-4">
                <Loader2 className="h-16 w-16 animate-spin text-primary" />
              </div>
              <CardTitle className="text-center">Verifying Payment</CardTitle>
              <CardDescription className="text-center">
                Please wait while we confirm your payment...
              </CardDescription>
            </CardHeader>
            <CardContent>
              <Skeleton className="h-4 w-full mb-2" />
              <Skeleton className="h-4 w-3/4" />
            </CardContent>
          </Card>
        )}

        {status === 'success' && (
          <Card>
            <CardHeader>
              <div className="flex items-center justify-center mb-4">
                <CheckCircle className="h-16 w-16 text-green-500" />
              </div>
              <CardTitle className="text-center text-2xl">Payment Successful!</CardTitle>
              <CardDescription className="text-center text-lg">
                {message}
              </CardDescription>
            </CardHeader>
            <CardContent className="space-y-4">
              <p className="text-center text-muted-foreground">
                Your tickets have been confirmed and sent to your email.
              </p>
              <div className="flex flex-col gap-3">
                <Link href="/my-tickets">
                  <Button className="w-full" size="lg">
                    View My Tickets
                  </Button>
                </Link>
                <Link href="/events">
                  <Button variant="outline" className="w-full" size="lg">
                    Browse More Events
                  </Button>
                </Link>
              </div>
            </CardContent>
          </Card>
        )}

        {status === 'failed' && (
          <Card>
            <CardHeader>
              <div className="flex items-center justify-center mb-4">
                <XCircle className="h-16 w-16 text-destructive" />
              </div>
              <CardTitle className="text-center text-2xl">Payment Failed</CardTitle>
              <CardDescription className="text-center text-lg">
                {message}
              </CardDescription>
            </CardHeader>
            <CardContent className="space-y-4">
              <p className="text-center text-muted-foreground">
                Your payment could not be processed. Please try again.
              </p>
              <div className="flex flex-col gap-3">
                <Link href="/events">
                  <Button className="w-full" size="lg">
                    Back to Events
                  </Button>
                </Link>
              </div>
            </CardContent>
          </Card>
        )}
      </div>
    </main>
  );
}

export default function PaymentVerifyPage() {
  return (
    <div className="min-h-screen flex flex-col">
      <Navbar />
      <Suspense fallback={
        <main className="flex-1 container mx-auto px-4 py-8">
          <div className="max-w-2xl mx-auto">
            <Card>
              <CardHeader>
                <div className="flex items-center justify-center mb-4">
                  <Loader2 className="h-16 w-16 animate-spin text-primary" />
                </div>
                <CardTitle className="text-center">Loading...</CardTitle>
              </CardHeader>
            </Card>
          </div>
        </main>
      }>
        <VerifyContent />
      </Suspense>
    </div>
  );
}
