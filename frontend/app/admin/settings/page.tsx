'use client';

import { useState, useEffect } from 'react';
import { useRouter } from 'next/navigation';
import { apiClient } from '@/lib/api-client';
import { PlatformSettings } from '@/lib/types';
import { useAuth } from '@/lib/auth-context';
import { DashboardLayout } from '@/components/dashboard-layout';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { Skeleton } from '@/components/ui/skeleton';
import { Alert, AlertDescription } from '@/components/ui/alert';
import { Separator } from '@/components/ui/separator';
import { Settings, DollarSign, Percent, AlertCircle, Save } from 'lucide-react';
import { toast } from 'sonner';
import { clearCurrencyCache, useCurrency } from '@/lib/currency';

export default function AdminSettingsPage() {
  const router = useRouter();
  const { user, loading: authLoading } = useAuth();
  const { formatAmount } = useCurrency();
  const [settings, setSettings] = useState<PlatformSettings | null>(null);
  const [loading, setLoading] = useState(true);
  const [saving, setSaving] = useState(false);
  const [error, setError] = useState('');
  const [formData, setFormData] = useState({
    platform_fee_percentage: 0,
    withdrawal_fee_percentage: 0,
    min_withdrawal_amount: 0,
    currency: 'NGN',
  });

  useEffect(() => {
    if (!authLoading) {
      if (!user || user.role !== 'admin') {
        router.push('/');
      } else {
        loadSettings();
      }
    }
  }, [user, authLoading, router]);

  const loadSettings = async () => {
    try {
      setLoading(true);
      const data = await apiClient.getPlatformSettings();
      setSettings(data);
      setFormData({
        platform_fee_percentage: data.platform_fee_percentage,
        withdrawal_fee_percentage: data.withdrawal_fee_percentage,
        min_withdrawal_amount: data.min_withdrawal_amount,
        currency: data.currency,
      });
    } catch (error) {
      toast.error('Failed to load platform settings');
      console.error(error);
    } finally {
      setLoading(false);
    }
  };

  const handleSaveSettings = async () => {
    setError('');

    // Validation
    if (formData.platform_fee_percentage < 0 || formData.platform_fee_percentage > 100) {
      setError('Platform fee percentage must be between 0 and 100');
      return;
    }

    if (formData.withdrawal_fee_percentage < 0 || formData.withdrawal_fee_percentage > 100) {
      setError('Withdrawal fee percentage must be between 0 and 100');
      return;
    }

    if (formData.min_withdrawal_amount < 0) {
      setError('Minimum withdrawal amount must be positive');
      return;
    }

    try {
      setSaving(true);
      const updatedSettings = await apiClient.updatePlatformSettings(formData);
      setSettings(updatedSettings);
      
      // Clear currency cache to force reload of new currency settings
      clearCurrencyCache();
      
      toast.success('Platform settings updated successfully');
    } catch (error) {
      setError(error instanceof Error ? error.message : 'Failed to update settings');
      toast.error('Failed to update settings');
    } finally {
      setSaving(false);
    }
  };

  const formatDate = (dateString: string) => {
    return new Date(dateString).toLocaleDateString('en-US', {
      year: 'numeric',
      month: 'long',
      day: 'numeric',
      hour: '2-digit',
      minute: '2-digit',
    });
  };

  if (authLoading || loading) {
    return (
      <DashboardLayout>
        <div className="container mx-auto px-4 py-8">
          <Skeleton className="h-8 w-64 mb-6" />
          <div className="space-y-6">
            <Skeleton className="h-96 w-full" />
            <Skeleton className="h-48 w-full" />
          </div>
        </div>
      </DashboardLayout>
    );
  }

  return (
    <DashboardLayout>
      <div className="container mx-auto px-4 py-8">
        <div className="flex items-center gap-2 mb-8">
          <Settings className="h-8 w-8" />
          <h1 className="text-4xl font-bold">Platform Settings</h1>
        </div>

        <div className="max-w-2xl space-y-6">
          {/* Fee Configuration */}
          <Card>
            <CardHeader>
              <CardTitle className="flex items-center gap-2">
                <Percent className="h-5 w-5" />
                Fee Configuration
              </CardTitle>
              <CardDescription>
                Configure platform and withdrawal fees
              </CardDescription>
            </CardHeader>
            <CardContent className="space-y-4">
              {error && (
                <Alert variant="destructive">
                  <AlertCircle className="h-4 w-4" />
                  <AlertDescription>{error}</AlertDescription>
                </Alert>
              )}

              <div className="space-y-2">
                <Label htmlFor="platform_fee">Platform Fee Percentage (%)</Label>
                <Input
                  id="platform_fee"
                  type="number"
                  min="0"
                  max="100"
                  step="0.01"
                  value={formData.platform_fee_percentage}
                  onChange={(e) => setFormData({ 
                    ...formData, 
                    platform_fee_percentage: parseFloat(e.target.value) || 0 
                  })}
                />
                <p className="text-sm text-muted-foreground">
                  Percentage fee charged on each ticket sale
                </p>
              </div>

              <div className="space-y-2">
                <Label htmlFor="withdrawal_fee">Withdrawal Fee Percentage (%)</Label>
                <Input
                  id="withdrawal_fee"
                  type="number"
                  min="0"
                  max="100"
                  step="0.01"
                  value={formData.withdrawal_fee_percentage}
                  onChange={(e) => setFormData({ 
                    ...formData, 
                    withdrawal_fee_percentage: parseFloat(e.target.value) || 0 
                  })}
                />
                <p className="text-sm text-muted-foreground">
                  Percentage fee charged on withdrawal requests
                </p>
              </div>

              <Separator />

              <div className="space-y-2">
                <Label htmlFor="min_withdrawal">Minimum Withdrawal Amount</Label>
                <div className="relative">
                  <DollarSign className="absolute left-3 top-3 h-4 w-4 text-muted-foreground" />
                  <Input
                    id="min_withdrawal"
                    type="number"
                    min="0"
                    step="100"
                    value={formData.min_withdrawal_amount}
                    onChange={(e) => setFormData({ 
                      ...formData, 
                      min_withdrawal_amount: parseFloat(e.target.value) || 0 
                    })}
                    className="pl-10"
                  />
                </div>
                <p className="text-sm text-muted-foreground">
                  Minimum amount organizers can withdraw
                </p>
              </div>

              <div className="space-y-2">
                <Label htmlFor="currency">Currency</Label>
                <Input
                  id="currency"
                  value={formData.currency}
                  onChange={(e) => setFormData({ ...formData, currency: e.target.value })}
                  placeholder="NGN"
                />
                <p className="text-sm text-muted-foreground">
                  Platform currency code (e.g., NGN, USD)
                </p>
              </div>

              <Button 
                onClick={handleSaveSettings} 
                disabled={saving}
                className="w-full"
              >
                <Save className="h-4 w-4 mr-2" />
                {saving ? 'Saving...' : 'Save Settings'}
              </Button>
            </CardContent>
          </Card>

          {/* Current Settings Info */}
          {settings && (
            <Card>
              <CardHeader>
                <CardTitle>Current Settings</CardTitle>
                <CardDescription>
                  Overview of current platform configuration
                </CardDescription>
              </CardHeader>
              <CardContent>
                <div className="grid grid-cols-2 gap-4 text-sm">
                  <div>
                    <span className="text-muted-foreground">Platform Fee:</span>
                    <span className="ml-2 font-medium">{settings.platform_fee_percentage}%</span>
                  </div>
                  <div>
                    <span className="text-muted-foreground">Withdrawal Fee:</span>
                    <span className="ml-2 font-medium">{settings.withdrawal_fee_percentage}%</span>
                  </div>
                  <div>
                    <span className="text-muted-foreground">Min Withdrawal:</span>
                    <span className="ml-2 font-medium">{formatAmount(settings.min_withdrawal_amount)}</span>
                  </div>
                  <div>
                    <span className="text-muted-foreground">Currency:</span>
                    <span className="ml-2 font-medium">{settings.currency}</span>
                  </div>
                  <div className="col-span-2">
                    <span className="text-muted-foreground">Last Updated:</span>
                    <span className="ml-2 font-medium">{formatDate(settings.updated_at)}</span>
                  </div>
                </div>
              </CardContent>
            </Card>
          )}

          {/* Fee Calculator */}
          <Card>
            <CardHeader>
              <CardTitle>Fee Calculator</CardTitle>
              <CardDescription>
                Preview how fees will be calculated
              </CardDescription>
            </CardHeader>
            <CardContent>
              <div className="space-y-4">
                <div className="p-4 bg-muted rounded-lg">
                  <h4 className="font-medium mb-2">Example: {formatAmount(10000)} ticket sale</h4>
                  <div className="space-y-1 text-sm">
                    <div className="flex justify-between">
                      <span>Ticket Price:</span>
                      <span>{formatAmount(10000)}</span>
                    </div>
                    <div className="flex justify-between">
                      <span>Platform Fee ({formData.platform_fee_percentage}%):</span>
                      <span>{formatAmount(10000 * formData.platform_fee_percentage / 100)}</span>
                    </div>
                    <Separator />
                    <div className="flex justify-between font-medium">
                      <span>Organizer Receives:</span>
                      <span>{formatAmount(10000 - (10000 * formData.platform_fee_percentage / 100))}</span>
                    </div>
                  </div>
                </div>

                <div className="p-4 bg-muted rounded-lg">
                  <h4 className="font-medium mb-2">Example: {formatAmount(50000)} withdrawal</h4>
                  <div className="space-y-1 text-sm">
                    <div className="flex justify-between">
                      <span>Withdrawal Amount:</span>
                      <span>{formatAmount(50000)}</span>
                    </div>
                    <div className="flex justify-between">
                      <span>Withdrawal Fee ({formData.withdrawal_fee_percentage}%):</span>
                      <span>{formatAmount(50000 * formData.withdrawal_fee_percentage / 100)}</span>
                    </div>
                    <Separator />
                    <div className="flex justify-between font-medium">
                      <span>Net Amount:</span>
                      <span>{formatAmount(50000 - (50000 * formData.withdrawal_fee_percentage / 100))}</span>
                    </div>
                  </div>
                </div>
              </div>
            </CardContent>
          </Card>

          {/* Important Notes */}
          <Alert>
            <AlertCircle className="h-4 w-4" />
            <AlertDescription>
              <strong>Important:</strong> Changes to fee settings will apply to all new transactions. 
              Existing pending withdrawals will use the fee rates that were active when they were created.
            </AlertDescription>
          </Alert>
        </div>
      </div>
    </DashboardLayout>
  );
}
