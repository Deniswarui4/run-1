'use client';

import { useState } from 'react';
import { apiClient } from '@/lib/api-client';
import { useAuth } from '@/lib/auth-context';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { Alert, AlertDescription } from '@/components/ui/alert';
import { Badge } from '@/components/ui/badge';
import { toast } from 'sonner';
import { Shield, QrCode, Smartphone, AlertCircle, CheckCircle, XCircle } from 'lucide-react';
import Image from 'next/image';

interface TwoFASetupData {
  secret: string;
  qr_code: string;
  manual_entry_key: string;
}

export function TwoFactorAuth() {
  const { user, refreshUser } = useAuth();
  const [setupLoading, setSetupLoading] = useState(false);
  const [enableLoading, setEnableLoading] = useState(false);
  const [disableLoading, setDisableLoading] = useState(false);
  const [showSetup, setShowSetup] = useState(false);
  const [verificationCode, setVerificationCode] = useState('');
  const [setupData, setSetupData] = useState<TwoFASetupData | null>(null);
  const [error, setError] = useState('');

  const handleSetup2FA = async () => {
    setError('');
    setSetupLoading(true);

    try {
      const response = await apiClient.setup2FA();
      setSetupData(response);
      setShowSetup(true);
      toast.success('2FA setup initiated. Please scan the QR code with your authenticator app.');
    } catch (err: unknown) {
      const error = err as { response?: { data?: { error?: string } } };
      setError(error.response?.data?.error || 'Failed to setup 2FA');
      toast.error('Failed to setup 2FA');
    } finally {
      setSetupLoading(false);
    }
  };

  const handleEnable2FA = async () => {
    if (!verificationCode.trim()) {
      setError('Please enter the verification code');
      return;
    }

    setError('');
    setEnableLoading(true);

    try {
      await apiClient.enable2FA({ code: verificationCode.trim() });
      await refreshUser();
      setShowSetup(false);
      setVerificationCode('');
      setSetupData(null);
      toast.success('2FA has been enabled successfully!');
    } catch (err: unknown) {
      const error = err as { response?: { data?: { error?: string } } };
      setError(error.response?.data?.error || 'Failed to enable 2FA');
      toast.error('Failed to enable 2FA');
    } finally {
      setEnableLoading(false);
    }
  };

  const handleDisable2FA = async () => {
    if (!verificationCode.trim()) {
      setError('Please enter the verification code');
      return;
    }

    setError('');
    setDisableLoading(true);

    try {
      await apiClient.disable2FA({ code: verificationCode.trim() });
      await refreshUser();
      setVerificationCode('');
      toast.success('2FA has been disabled successfully!');
    } catch (err: unknown) {
      const error = err as { response?: { data?: { error?: string } } };
      setError(error.response?.data?.error || 'Failed to disable 2FA');
      toast.error('Failed to disable 2FA');
    } finally {
      setDisableLoading(false);
    }
  };

  const handleCancelSetup = () => {
    setShowSetup(false);
    setVerificationCode('');
    setSetupData(null);
    setError('');
  };

  if (user?.two_factor_enabled) {
    return (
      <Card>
        <CardHeader>
          <div className="flex items-center gap-2">
            <Shield className="h-5 w-5 text-green-600" />
            <CardTitle>Two-Factor Authentication</CardTitle>
            <Badge variant="secondary" className="bg-green-100 text-green-800">
              <CheckCircle className="h-3 w-3 mr-1" />
              Enabled
            </Badge>
          </div>
          <CardDescription>
            Your account is protected with two-factor authentication using an authenticator app.
          </CardDescription>
        </CardHeader>
        <CardContent className="space-y-4">
          <div className="flex items-center gap-2 p-3 bg-green-50 rounded-lg border border-green-200">
            <CheckCircle className="h-5 w-5 text-green-600" />
            <div>
              <p className="font-medium text-green-800">2FA is active</p>
              <p className="text-sm text-green-600">
                You&apos;ll need to enter a code from your authenticator app when signing in.
              </p>
            </div>
          </div>

          <div className="space-y-3">
            <Label htmlFor="disable-code">Enter verification code to disable 2FA</Label>
            <div className="flex gap-2">
              <Input
                id="disable-code"
                type="text"
                placeholder="Enter 6-digit code"
                value={verificationCode}
                onChange={(e) => setVerificationCode(e.target.value)}
                maxLength={6}
                className="flex-1"
              />
              <Button
                onClick={handleDisable2FA}
                disabled={disableLoading || !verificationCode.trim()}
                variant="destructive"
              >
                {disableLoading ? 'Disabling...' : 'Disable 2FA'}
              </Button>
            </div>
          </div>

          {error && (
            <Alert variant="destructive">
              <AlertCircle className="h-4 w-4" />
              <AlertDescription>{error}</AlertDescription>
            </Alert>
          )}
        </CardContent>
      </Card>
    );
  }

  return (
    <Card>
      <CardHeader>
        <div className="flex items-center gap-2">
          <Shield className="h-5 w-5 text-orange-600" />
          <CardTitle>Two-Factor Authentication</CardTitle>
          <Badge variant="outline">
            <XCircle className="h-3 w-3 mr-1" />
            Disabled
          </Badge>
        </div>
        <CardDescription>
          Add an extra layer of security to your account by requiring a verification code from your authenticator app.
        </CardDescription>
      </CardHeader>
      <CardContent className="space-y-4">
        <div className="flex items-center gap-2 p-3 bg-orange-50 rounded-lg border border-orange-200">
          <AlertCircle className="h-5 w-5 text-orange-600" />
          <div>
            <p className="font-medium text-orange-800">2FA is not enabled</p>
            <p className="text-sm text-orange-600">
              Enable two-factor authentication for enhanced account security.
            </p>
          </div>
        </div>

        {!showSetup ? (
          <Button onClick={handleSetup2FA} disabled={setupLoading} className="w-full">
            <QrCode className="h-4 w-4 mr-2" />
            {setupLoading ? 'Setting up...' : 'Set up 2FA'}
          </Button>
        ) : (
          <div className="space-y-4">
            {setupData && (
              <>
                <div className="space-y-4">
                  <div className="text-center">
                    <h4 className="font-medium mb-2">Scan QR Code</h4>
                    <p className="text-sm text-muted-foreground mb-4">
                      Scan this QR code with your authenticator app (Google Authenticator, Authy, etc.)
                    </p>

                    <div className="flex justify-center mb-4">
                      <Image
                        src={setupData.qr_code}
                        alt="2FA QR Code"
                        className="border rounded-lg p-2 bg-white"
                        width={200}
                        height={200}
                        style={{ maxWidth: '200px', height: 'auto' }}
                      />
                    </div>

                    <div className="bg-muted p-3 rounded-lg mb-4">
                      <p className="text-sm font-mono text-center break-all">
                        {setupData.manual_entry_key}
                      </p>
                    </div>

                    <p className="text-sm text-muted-foreground">
                      Or enter this code manually in your authenticator app
                    </p>
                  </div>

                  <div className="space-y-3">
                    <Label htmlFor="verification-code">Enter the 6-digit code from your app</Label>
                    <div className="flex gap-2">
                      <Input
                        id="verification-code"
                        type="text"
                        placeholder="Enter 6-digit code"
                        value={verificationCode}
                        onChange={(e) => setVerificationCode(e.target.value)}
                        maxLength={6}
                        className="flex-1"
                      />
                      <Button
                        onClick={handleEnable2FA}
                        disabled={enableLoading || !verificationCode.trim()}
                      >
                        {enableLoading ? 'Enabling...' : 'Enable 2FA'}
                      </Button>
                    </div>
                  </div>
                </div>

                <div className="flex gap-2">
                  <Button variant="outline" onClick={handleCancelSetup} className="flex-1">
                    Cancel
                  </Button>
                </div>
              </>
            )}

            {error && (
              <Alert variant="destructive">
                <AlertCircle className="h-4 w-4" />
                <AlertDescription>{error}</AlertDescription>
              </Alert>
            )}
          </div>
        )}

        <div className="text-sm text-muted-foreground">
          <p className="font-medium mb-2">Supported apps:</p>
          <div className="flex flex-wrap gap-2">
            <div className="flex items-center gap-1">
              <Smartphone className="h-4 w-4" />
              <span>Google Authenticator</span>
            </div>
            <div className="flex items-center gap-1">
              <Smartphone className="h-4 w-4" />
              <span>Microsoft Authenticator</span>
            </div>
            <div className="flex items-center gap-1">
              <Smartphone className="h-4 w-4" />
              <span>Authy</span>
            </div>
          </div>
        </div>
      </CardContent>
    </Card>
  );
}
