'use client';

import { useState, useEffect } from 'react';
import { useRouter } from 'next/navigation';
import { apiClient } from '@/lib/api-client';
import { Withdrawal } from '@/lib/types';
import { useAuth } from '@/lib/auth-context';
import { DashboardLayout } from '@/components/dashboard-layout';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import { Badge } from '@/components/ui/badge';
import { Skeleton } from '@/components/ui/skeleton';
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs';
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@/components/ui/table';
import { Dialog, DialogContent, DialogDescription, DialogFooter, DialogHeader, DialogTitle } from '@/components/ui/dialog';
import { Textarea } from '@/components/ui/textarea';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { CheckCircle, XCircle, Clock, DollarSign, CreditCard } from 'lucide-react';
import { toast } from 'sonner';
import { useCurrency } from '@/lib/currency';

export default function AdminWithdrawalsPage() {
  const router = useRouter();
  const { user, loading: authLoading } = useAuth();
  const { formatAmount } = useCurrency();
  const [withdrawals, setWithdrawals] = useState<Withdrawal[]>([]);
  const [loading, setLoading] = useState(true);
  const [reviewDialogOpen, setReviewDialogOpen] = useState(false);
  const [processDialogOpen, setProcessDialogOpen] = useState(false);
  const [selectedWithdrawal, setSelectedWithdrawal] = useState<Withdrawal | null>(null);
  const [reviewComment, setReviewComment] = useState('');
  const [reviewAction, setReviewAction] = useState<'approve' | 'reject'>('approve');
  const [transactionRef, setTransactionRef] = useState('');
  const [submitting, setSubmitting] = useState(false);

  useEffect(() => {
    if (!authLoading) {
      if (!user || user.role !== 'admin') {
        router.push('/');
      } else {
        loadWithdrawals();
      }
    }
  }, [user, authLoading, router]);

  const loadWithdrawals = async () => {
    try {
      setLoading(true);
      const data = await apiClient.getWithdrawalRequests();
      setWithdrawals(Array.isArray(data) ? data : []);
    } catch (error) {
      toast.error('Failed to load withdrawals');
      console.error(error);
      setWithdrawals([]);
    } finally {
      setLoading(false);
    }
  };

  const handleReviewWithdrawal = (withdrawal: Withdrawal, action: 'approve' | 'reject') => {
    setSelectedWithdrawal(withdrawal);
    setReviewAction(action);
    setReviewComment('');
    setReviewDialogOpen(true);
  };

  const handleProcessWithdrawal = (withdrawal: Withdrawal) => {
    setSelectedWithdrawal(withdrawal);
    setTransactionRef('');
    setProcessDialogOpen(true);
  };

  const submitReview = async () => {
    if (!selectedWithdrawal) return;

    try {
      setSubmitting(true);
      await apiClient.reviewWithdrawal(selectedWithdrawal.id, {
        action: reviewAction,
        comment: reviewComment,
      });
      toast.success(`Withdrawal ${reviewAction}d successfully`);
      setReviewDialogOpen(false);
      loadWithdrawals();
    } catch (error) {
      toast.error(`Failed to ${reviewAction} withdrawal`);
      console.error(error);
    } finally {
      setSubmitting(false);
    }
  };

  const submitProcess = async () => {
    if (!selectedWithdrawal || !transactionRef.trim()) return;

    try {
      setSubmitting(true);
      await apiClient.processWithdrawal(selectedWithdrawal.id, transactionRef);
      toast.success('Withdrawal processed successfully');
      setProcessDialogOpen(false);
      loadWithdrawals();
    } catch (error) {
      toast.error('Failed to process withdrawal');
      console.error(error);
    } finally {
      setSubmitting(false);
    }
  };

  const getStatusColor = (status: string) => {
    switch (status) {
      case 'pending':
        return 'secondary';
      case 'approved':
        return 'default';
      case 'processed':
        return 'default';
      case 'rejected':
        return 'destructive';
      default:
        return 'outline';
    }
  };

  const getStatusIcon = (status: string) => {
    switch (status) {
      case 'pending':
        return <Clock className="h-4 w-4" />;
      case 'approved':
        return <CheckCircle className="h-4 w-4" />;
      case 'processed':
        return <CheckCircle className="h-4 w-4" />;
      case 'rejected':
        return <XCircle className="h-4 w-4" />;
      default:
        return <Clock className="h-4 w-4" />;
    }
  };

  const formatDate = (dateString: string) => {
    return new Date(dateString).toLocaleDateString('en-US', {
      year: 'numeric',
      month: 'short',
      day: 'numeric',
      hour: '2-digit',
      minute: '2-digit',
    });
  };

  // Remove the old formatAmount function since we're using the hook

  const filterWithdrawalsByStatus = (status: string) => {
    return withdrawals.filter(w => w.status === status);
  };

  const getWithdrawalStats = () => {
    return {
      total: withdrawals.length,
      pending: filterWithdrawalsByStatus('pending').length,
      approved: filterWithdrawalsByStatus('approved').length,
      processed: filterWithdrawalsByStatus('processed').length,
      rejected: filterWithdrawalsByStatus('rejected').length,
      totalAmount: withdrawals.reduce((sum, w) => sum + w.amount, 0),
      processedAmount: filterWithdrawalsByStatus('processed').reduce((sum, w) => sum + w.net_amount, 0),
    };
  };

  const stats = getWithdrawalStats();

  const renderWithdrawalRow = (withdrawal: Withdrawal) => (
    <TableRow key={withdrawal.id}>
      <TableCell>
        {withdrawal.organizer && (
          <div>
            <p className="font-medium">
              {withdrawal.organizer.first_name} {withdrawal.organizer.last_name}
            </p>
            <p className="text-sm text-muted-foreground">
              {withdrawal.organizer.email}
            </p>
          </div>
        )}
      </TableCell>
      <TableCell>
        <div className="text-sm">
          <p className="font-medium">{withdrawal.bank_name}</p>
          <p className="text-muted-foreground">
            {withdrawal.account_number}
          </p>
          <p className="text-muted-foreground">
            {withdrawal.account_name}
          </p>
        </div>
      </TableCell>
      <TableCell className="font-medium">
        {formatAmount(withdrawal.amount)}
      </TableCell>
      <TableCell className="font-medium text-green-600">
        {formatAmount(withdrawal.net_amount)}
      </TableCell>
      <TableCell>
        <Badge variant={getStatusColor(withdrawal.status)} className="capitalize">
          <span className="flex items-center gap-1">
            {getStatusIcon(withdrawal.status)}
            {withdrawal.status}
          </span>
        </Badge>
      </TableCell>
      <TableCell>
        {formatDate(withdrawal.created_at)}
      </TableCell>
      <TableCell>
        <div className="flex gap-2">
          {withdrawal.status === 'pending' && (
            <>
              <Button
                variant="outline"
                size="sm"
                onClick={() => handleReviewWithdrawal(withdrawal, 'approve')}
              >
                <CheckCircle className="h-4 w-4 mr-1" />
                Approve
              </Button>
              <Button
                variant="destructive"
                size="sm"
                onClick={() => handleReviewWithdrawal(withdrawal, 'reject')}
              >
                <XCircle className="h-4 w-4 mr-1" />
                Reject
              </Button>
            </>
          )}
          {withdrawal.status === 'approved' && (
            <Button
              size="sm"
              onClick={() => handleProcessWithdrawal(withdrawal)}
            >
              <CreditCard className="h-4 w-4 mr-1" />
              Process
            </Button>
          )}
        </div>
      </TableCell>
    </TableRow>
  );

  if (authLoading || loading) {
    return (
      <DashboardLayout>
        <div className="container mx-auto px-4 py-8">
          <Skeleton className="h-8 w-64 mb-6" />
          <div className="grid md:grid-cols-4 gap-6 mb-8">
            {[...Array(4)].map((_, i) => (
              <Skeleton key={i} className="h-32" />
            ))}
          </div>
          <Skeleton className="h-96 w-full" />
        </div>
      </DashboardLayout>
    );
  }

  return (
    <DashboardLayout>
      <div className="container mx-auto px-4 py-8">
        <h1 className="text-4xl font-bold mb-8">Withdrawal Management</h1>

        {/* Stats Cards */}
        <div className="grid md:grid-cols-2 lg:grid-cols-4 gap-6 mb-8">
          <Card>
            <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
              <CardTitle className="text-sm font-medium">Total Requests</CardTitle>
              <CreditCard className="h-4 w-4 text-muted-foreground" />
            </CardHeader>
            <CardContent>
              <div className="text-2xl font-bold">{stats.total}</div>
              <p className="text-xs text-muted-foreground">
                {stats.pending} pending review
              </p>
            </CardContent>
          </Card>

          <Card>
            <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
              <CardTitle className="text-sm font-medium">Total Amount</CardTitle>
              <DollarSign className="h-4 w-4 text-muted-foreground" />
            </CardHeader>
            <CardContent>
              <div className="text-2xl font-bold">
                {formatAmount(stats.totalAmount)}
              </div>
              <p className="text-xs text-muted-foreground">
                All requests
              </p>
            </CardContent>
          </Card>

          <Card>
            <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
              <CardTitle className="text-sm font-medium">Processed</CardTitle>
              <CheckCircle className="h-4 w-4 text-muted-foreground" />
            </CardHeader>
            <CardContent>
              <div className="text-2xl font-bold text-green-600">
                {stats.processed}
              </div>
              <p className="text-xs text-muted-foreground">
                {formatAmount(stats.processedAmount)} paid
              </p>
            </CardContent>
          </Card>

          <Card>
            <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
              <CardTitle className="text-sm font-medium">Rejected</CardTitle>
              <XCircle className="h-4 w-4 text-muted-foreground" />
            </CardHeader>
            <CardContent>
              <div className="text-2xl font-bold text-red-600">
                {stats.rejected}
              </div>
              <p className="text-xs text-muted-foreground">
                Declined requests
              </p>
            </CardContent>
          </Card>
        </div>

        {/* Withdrawals Table */}
        <Tabs defaultValue="pending" className="space-y-6">
          <TabsList>
            <TabsTrigger value="pending">
              Pending ({stats.pending})
            </TabsTrigger>
            <TabsTrigger value="approved">
              Approved ({stats.approved})
            </TabsTrigger>
            <TabsTrigger value="processed">
              Processed ({stats.processed})
            </TabsTrigger>
            <TabsTrigger value="rejected">
              Rejected ({stats.rejected})
            </TabsTrigger>
          </TabsList>

          <TabsContent value="pending">
            <Card>
              <CardHeader>
                <CardTitle>Pending Withdrawals</CardTitle>
                <CardDescription>
                  Withdrawal requests awaiting review
                </CardDescription>
              </CardHeader>
              <CardContent>
                <Table>
                  <TableHeader>
                    <TableRow>
                      <TableHead>Organizer</TableHead>
                      <TableHead>Bank Details</TableHead>
                      <TableHead>Amount</TableHead>
                      <TableHead>Net Amount</TableHead>
                      <TableHead>Status</TableHead>
                      <TableHead>Date</TableHead>
                      <TableHead>Actions</TableHead>
                    </TableRow>
                  </TableHeader>
                  <TableBody>
                    {filterWithdrawalsByStatus('pending').map(renderWithdrawalRow)}
                  </TableBody>
                </Table>
              </CardContent>
            </Card>
          </TabsContent>

          <TabsContent value="approved">
            <Card>
              <CardHeader>
                <CardTitle>Approved Withdrawals</CardTitle>
                <CardDescription>
                  Approved withdrawals ready for processing
                </CardDescription>
              </CardHeader>
              <CardContent>
                <Table>
                  <TableHeader>
                    <TableRow>
                      <TableHead>Organizer</TableHead>
                      <TableHead>Bank Details</TableHead>
                      <TableHead>Amount</TableHead>
                      <TableHead>Net Amount</TableHead>
                      <TableHead>Status</TableHead>
                      <TableHead>Date</TableHead>
                      <TableHead>Actions</TableHead>
                    </TableRow>
                  </TableHeader>
                  <TableBody>
                    {filterWithdrawalsByStatus('approved').map(renderWithdrawalRow)}
                  </TableBody>
                </Table>
              </CardContent>
            </Card>
          </TabsContent>

          <TabsContent value="processed">
            <Card>
              <CardHeader>
                <CardTitle>Processed Withdrawals</CardTitle>
                <CardDescription>
                  Successfully processed withdrawals
                </CardDescription>
              </CardHeader>
              <CardContent>
                <Table>
                  <TableHeader>
                    <TableRow>
                      <TableHead>Organizer</TableHead>
                      <TableHead>Bank Details</TableHead>
                      <TableHead>Amount</TableHead>
                      <TableHead>Net Amount</TableHead>
                      <TableHead>Status</TableHead>
                      <TableHead>Date</TableHead>
                      <TableHead>Transaction Ref</TableHead>
                    </TableRow>
                  </TableHeader>
                  <TableBody>
                    {filterWithdrawalsByStatus('processed').map((withdrawal) => (
                      <TableRow key={withdrawal.id}>
                        <TableCell>
                          {withdrawal.organizer && (
                            <div>
                              <p className="font-medium">
                                {withdrawal.organizer.first_name} {withdrawal.organizer.last_name}
                              </p>
                              <p className="text-sm text-muted-foreground">
                                {withdrawal.organizer.email}
                              </p>
                            </div>
                          )}
                        </TableCell>
                        <TableCell>
                          <div className="text-sm">
                            <p className="font-medium">{withdrawal.bank_name}</p>
                            <p className="text-muted-foreground">
                              {withdrawal.account_number}
                            </p>
                            <p className="text-muted-foreground">
                              {withdrawal.account_name}
                            </p>
                          </div>
                        </TableCell>
                        <TableCell className="font-medium">
                          {formatAmount(withdrawal.amount)}
                        </TableCell>
                        <TableCell className="font-medium text-green-600">
                          {formatAmount(withdrawal.net_amount)}
                        </TableCell>
                        <TableCell>
                          <Badge variant={getStatusColor(withdrawal.status)} className="capitalize">
                            <span className="flex items-center gap-1">
                              {getStatusIcon(withdrawal.status)}
                              {withdrawal.status}
                            </span>
                          </Badge>
                        </TableCell>
                        <TableCell>
                          {formatDate(withdrawal.created_at)}
                        </TableCell>
                        <TableCell className="font-mono text-sm">
                          {withdrawal.transaction_ref || 'N/A'}
                        </TableCell>
                      </TableRow>
                    ))}
                  </TableBody>
                </Table>
              </CardContent>
            </Card>
          </TabsContent>

          <TabsContent value="rejected">
            <Card>
              <CardHeader>
                <CardTitle>Rejected Withdrawals</CardTitle>
                <CardDescription>
                  Declined withdrawal requests
                </CardDescription>
              </CardHeader>
              <CardContent>
                <Table>
                  <TableHeader>
                    <TableRow>
                      <TableHead>Organizer</TableHead>
                      <TableHead>Bank Details</TableHead>
                      <TableHead>Amount</TableHead>
                      <TableHead>Status</TableHead>
                      <TableHead>Date</TableHead>
                      <TableHead>Comment</TableHead>
                    </TableRow>
                  </TableHeader>
                  <TableBody>
                    {filterWithdrawalsByStatus('rejected').map((withdrawal) => (
                      <TableRow key={withdrawal.id}>
                        <TableCell>
                          {withdrawal.organizer && (
                            <div>
                              <p className="font-medium">
                                {withdrawal.organizer.first_name} {withdrawal.organizer.last_name}
                              </p>
                              <p className="text-sm text-muted-foreground">
                                {withdrawal.organizer.email}
                              </p>
                            </div>
                          )}
                        </TableCell>
                        <TableCell>
                          <div className="text-sm">
                            <p className="font-medium">{withdrawal.bank_name}</p>
                            <p className="text-muted-foreground">
                              {withdrawal.account_number}
                            </p>
                          </div>
                        </TableCell>
                        <TableCell className="font-medium">
                          {formatAmount(withdrawal.amount)}
                        </TableCell>
                        <TableCell>
                          <Badge variant={getStatusColor(withdrawal.status)} className="capitalize">
                            <span className="flex items-center gap-1">
                              {getStatusIcon(withdrawal.status)}
                              {withdrawal.status}
                            </span>
                          </Badge>
                        </TableCell>
                        <TableCell>
                          {formatDate(withdrawal.created_at)}
                        </TableCell>
                        <TableCell className="text-sm">
                          {withdrawal.comment || 'No comment'}
                        </TableCell>
                      </TableRow>
                    ))}
                  </TableBody>
                </Table>
              </CardContent>
            </Card>
          </TabsContent>
        </Tabs>

        {/* Review Dialog */}
        <Dialog open={reviewDialogOpen} onOpenChange={setReviewDialogOpen}>
          <DialogContent>
            <DialogHeader>
              <DialogTitle>
                {reviewAction === 'approve' ? 'Approve' : 'Reject'} Withdrawal
              </DialogTitle>
              <DialogDescription>
                {selectedWithdrawal && (
                  <>
                    {reviewAction === 'approve' ? 'Approve' : 'Reject'} withdrawal request for{' '}
                    {formatAmount(selectedWithdrawal.amount)}
                  </>
                )}
              </DialogDescription>
            </DialogHeader>
            
            <div className="space-y-4 py-4">
              {selectedWithdrawal && (
                <div className="space-y-2">
                  <h4 className="font-medium">Withdrawal Details</h4>
                  <div className="text-sm text-muted-foreground space-y-1">
                    <p><strong>Organizer:</strong> {selectedWithdrawal.organizer?.first_name} {selectedWithdrawal.organizer?.last_name}</p>
                    <p><strong>Amount:</strong> {formatAmount(selectedWithdrawal.amount)}</p>
                    <p><strong>Net Amount:</strong> {formatAmount(selectedWithdrawal.net_amount)}</p>
                    <p><strong>Bank:</strong> {selectedWithdrawal.bank_name}</p>
                    <p><strong>Account:</strong> {selectedWithdrawal.account_number} - {selectedWithdrawal.account_name}</p>
                  </div>
                </div>
              )}
              
              <div className="space-y-2">
                <Label htmlFor="comment">Review Comment *</Label>
                <Textarea
                  id="comment"
                  placeholder={
                    reviewAction === 'approve' 
                      ? "Withdrawal approved for processing..."
                      : "Please specify the reason for rejection..."
                  }
                  value={reviewComment}
                  onChange={(e) => setReviewComment(e.target.value)}
                  rows={4}
                  required
                />
              </div>
            </div>

            <DialogFooter>
              <Button variant="outline" onClick={() => setReviewDialogOpen(false)}>
                Cancel
              </Button>
              <Button
                variant={reviewAction === 'approve' ? 'default' : 'destructive'}
                onClick={submitReview}
                disabled={submitting || !reviewComment.trim()}
              >
                {submitting 
                  ? (reviewAction === 'approve' ? 'Approving...' : 'Rejecting...')
                  : (reviewAction === 'approve' ? 'Approve Withdrawal' : 'Reject Withdrawal')
                }
              </Button>
            </DialogFooter>
          </DialogContent>
        </Dialog>

        {/* Process Dialog */}
        <Dialog open={processDialogOpen} onOpenChange={setProcessDialogOpen}>
          <DialogContent>
            <DialogHeader>
              <DialogTitle>Process Withdrawal</DialogTitle>
              <DialogDescription>
                {selectedWithdrawal && (
                  <>Mark withdrawal of {formatAmount(selectedWithdrawal.amount)} as processed</>
                )}
              </DialogDescription>
            </DialogHeader>
            
            <div className="space-y-4 py-4">
              <div className="space-y-2">
                <Label htmlFor="transactionRef">Transaction Reference *</Label>
                <Input
                  id="transactionRef"
                  placeholder="Enter bank transaction reference"
                  value={transactionRef}
                  onChange={(e) => setTransactionRef(e.target.value)}
                  required
                />
                <p className="text-sm text-muted-foreground">
                  Enter the bank transaction reference number for this withdrawal
                </p>
              </div>
            </div>

            <DialogFooter>
              <Button variant="outline" onClick={() => setProcessDialogOpen(false)}>
                Cancel
              </Button>
              <Button
                onClick={submitProcess}
                disabled={submitting || !transactionRef.trim()}
              >
                {submitting ? 'Processing...' : 'Mark as Processed'}
              </Button>
            </DialogFooter>
          </DialogContent>
        </Dialog>
      </div>
    </DashboardLayout>
  );
}
