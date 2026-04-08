import { Badge } from '@/components/ui/badge';
import { Star } from 'lucide-react';
import { FeaturedType } from '@/lib/types';

interface FeaturedEventsBadgeProps {
  /** When provided (admin context), shows "Manual" or "Auto" label */
  featuredType?: FeaturedType;
  className?: string;
}

/**
 * Renders a "Featured" badge.
 * In public context (no featuredType): shows "Featured".
 * In admin context (featuredType provided): shows "Manual" or "Auto".
 */
export function FeaturedEventsBadge({ featuredType, className }: FeaturedEventsBadgeProps) {
  if (featuredType === 'none' || featuredType === undefined) {
    // Public context — just show "Featured"
    return (
      <Badge
        className={`bg-yellow-500/90 text-white hover:bg-yellow-500 flex items-center gap-1 ${className ?? ''}`}
      >
        <Star className="h-3 w-3 fill-current" />
        Featured
      </Badge>
    );
  }

  if (featuredType === 'manual') {
    return (
      <Badge
        className={`bg-blue-600/90 text-white hover:bg-blue-600 flex items-center gap-1 ${className ?? ''}`}
      >
        <Star className="h-3 w-3 fill-current" />
        Manual
      </Badge>
    );
  }

  // auto
  return (
    <Badge
      className={`bg-green-600/90 text-white hover:bg-green-600 flex items-center gap-1 ${className ?? ''}`}
    >
      <Star className="h-3 w-3 fill-current" />
      Auto
    </Badge>
  );
}
