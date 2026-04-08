'use client';

import { useState } from 'react';
import { Dialog, DialogContent, DialogHeader, DialogTitle } from '@/components/ui/dialog';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Copy, Check, Facebook, Twitter, MessageCircle } from 'lucide-react';
import { toast } from 'sonner';

interface ShareModalProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  eventId: string;
  eventTitle: string;
}

const BASE_URL = process.env.NEXT_PUBLIC_BASE_URL || 'http://localhost:3000';

export function ShareModal({ open, onOpenChange, eventId, eventTitle }: ShareModalProps) {
  const [copied, setCopied] = useState(false);
  const [clipboardFailed, setClipboardFailed] = useState(false);

  const shareUrl = `${BASE_URL}/events/${eventId}`;
  const encodedUrl = encodeURIComponent(shareUrl);
  const encodedTitle = encodeURIComponent(eventTitle);

  const handleCopyLink = async () => {
    try {
      await navigator.clipboard.writeText(shareUrl);
      setCopied(true);
      toast.success('Link copied to clipboard!');
      setTimeout(() => setCopied(false), 2000);
    } catch {
      setClipboardFailed(true);
    }
  };

  const handleSocialShare = (url: string) => {
    window.open(url, '_blank', 'noopener,noreferrer');
  };

  const facebookUrl = `https://www.facebook.com/sharer/sharer.php?u=${encodedUrl}`;
  const twitterUrl = `https://twitter.com/intent/tweet?text=${encodedTitle}&url=${encodedUrl}`;
  const whatsappUrl = `https://wa.me/?text=${encodedTitle}%20${encodedUrl}`;

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="sm:max-w-md">
        <DialogHeader>
          <DialogTitle>Share this event</DialogTitle>
        </DialogHeader>

        <div className="space-y-4 pt-2">
          {/* Copy Link */}
          {clipboardFailed ? (
            <div className="space-y-2">
              <p className="text-sm text-muted-foreground">Copy the link below:</p>
              <div className="flex gap-2">
                <Input value={shareUrl} readOnly className="text-sm" />
              </div>
            </div>
          ) : (
            <Button
              variant="outline"
              className="w-full justify-start gap-3 h-12"
              onClick={handleCopyLink}
            >
              {copied ? (
                <Check className="h-5 w-5 text-green-500" />
              ) : (
                <Copy className="h-5 w-5" />
              )}
              <span>{copied ? 'Copied!' : 'Copy Link'}</span>
            </Button>
          )}

          <div className="relative">
            <div className="absolute inset-0 flex items-center">
              <span className="w-full border-t" />
            </div>
            <div className="relative flex justify-center text-xs uppercase">
              <span className="bg-background px-2 text-muted-foreground">or share via</span>
            </div>
          </div>

          {/* Social Buttons */}
          <div className="grid grid-cols-3 gap-3">
            <Button
              variant="outline"
              className="flex flex-col gap-1 h-16"
              onClick={() => handleSocialShare(facebookUrl)}
            >
              <Facebook className="h-5 w-5 text-blue-600" />
              <span className="text-xs">Facebook</span>
            </Button>

            <Button
              variant="outline"
              className="flex flex-col gap-1 h-16"
              onClick={() => handleSocialShare(twitterUrl)}
            >
              <Twitter className="h-5 w-5 text-sky-500" />
              <span className="text-xs">Twitter/X</span>
            </Button>

            <Button
              variant="outline"
              className="flex flex-col gap-1 h-16"
              onClick={() => handleSocialShare(whatsappUrl)}
            >
              <MessageCircle className="h-5 w-5 text-green-500" />
              <span className="text-xs">WhatsApp</span>
            </Button>
          </div>
        </div>
      </DialogContent>
    </Dialog>
  );
}
