'use client';

import { Suspense, useState, useEffect } from 'react';
import { useSearchParams } from 'next/navigation';
import Link from 'next/link';
import { Button } from '@/components/ui/button';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { Alert, AlertDescription } from '@/components/ui/alert';
import { Navbar } from '@/components/navbar';
import { Mail, CheckCircle, Clock } from 'lucide-react';
import { apiClient } from '@/lib/api-client';
import { toast } from 'sonner';

function VerifyEmailContent() {
  const searchParams = useSearchParams();
  const email = searchParams.get('email');
  const [timeLeft, setTimeLeft] = useState(60);
  const [canResend, setCanResend] = useState(false);

  useEffect(() => {
    if (timeLeft > 0) {
      const timer = setTimeout(() => setTimeLeft(timeLeft - 1), 1000);
      return () => clearTimeout(timer);
    } else {
      setCanResend(true);
    }
  }, [timeLeft]);

  const handleResendEmail = async () => {
    if (!email) {
      toast.error('Email address is required');
      return;
    }

    try {
      await apiClient.resendVerification(email);
      toast.success('Verification email sent! Check your inbox.');
      setTimeLeft(60);
      setCanResend(false);
    } catch (error) {
      toast.error('Failed to resend verification email');
      console.error('Resend error:', error);
    }
  };

  return (
    <div className="flex-1 flex items-center justify-center p-4">
      <Card className="w-full max-w-md">
        <CardHeader className="text-center">
          <div className="mx-auto w-12 h-12 bg-primary/10 rounded-full flex items-center justify-center mb-4">
            <Mail className="h-6 w-6 text-primary" />
          </div>
          <CardTitle>Check Your Email</CardTitle>
          <CardDescription>
            We&#39;ve sent a verification link to your email address
          </CardDescription>
        </CardHeader>
        
        <CardContent className="space-y-4">
          {email && (
            <Alert>
              <CheckCircle className="h-4 w-4" />
              <AlertDescription>
                Verification email sent to: <strong>{email}</strong>
              </AlertDescription>
            </Alert>
          )}

          <div className="text-center space-y-4">
            <p className="text-sm text-muted-foreground">
              Click the verification link in your email to activate your account.
              You can then log in with your credentials.
            </p>

            <div className="space-y-2">
              <p className="text-sm font-medium">Didn&#39;t receive the email?</p>
              <div className="space-y-2">
                <p className="text-xs text-muted-foreground">
                  • Check your spam/junk folder
                  • Make sure the email address is correct
                  • Wait a few minutes for delivery
                </p>
                
                {canResend ? (
                  <Button 
                    variant="outline" 
                    onClick={handleResendEmail}
                    className="w-full"
                  >
                    Resend Verification Email
                  </Button>
                ) : (
                  <Button variant="outline" disabled className="w-full">
                    <Clock className="h-4 w-4 mr-2" />
                    Resend in {timeLeft}s
                  </Button>
                )}
              </div>
            </div>

            <div className="pt-4 border-t">
              <p className="text-sm text-muted-foreground mb-2">
                Already verified your email?
              </p>
              <Link href="/login">
                <Button className="w-full">
                  Go to Login
                </Button>
              </Link>
            </div>
          </div>
        </CardContent>
      </Card>
    </div>
  );
}

export default function VerifyEmailPage() {
  return (
    <div className="min-h-screen flex flex-col">
      <Navbar />
      <Suspense fallback={
        <div className="flex-1 flex items-center justify-center p-4">
          <Card className="w-full max-w-md">
            <CardHeader className="text-center">
              <CardTitle>Loading...</CardTitle>
            </CardHeader>
          </Card>
        </div>
      }>
        <VerifyEmailContent />
      </Suspense>
    </div>
  );
}
