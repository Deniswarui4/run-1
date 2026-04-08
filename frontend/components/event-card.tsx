'use client';

import { useState } from 'react';
import Link from 'next/link';
import { Badge } from '@/components/ui/badge';
import { Calendar, Star, Share2 } from 'lucide-react';
import { Event } from '@/lib/types';
import { ShareModal } from '@/components/share-modal';

interface EventCardProps {
    event: Event;
    formatAmount: (amount: number) => string;
    formatDate: (dateString: string) => string;
    formatTime: (dateString: string) => string;
    getMinPrice: (event: Event) => number | null;
    isUpcoming: (dateString: string) => boolean;
}

export function EventCard({
    event,
    formatAmount,
    formatDate,
    formatTime,
    getMinPrice,
    isUpcoming
}: EventCardProps) {
    const minPrice = getMinPrice(event);
    const upcoming = isUpcoming(event.end_date); // Check end_date, not start_date
    const [shareModalOpen, setShareModalOpen] = useState(false);

    return (
        <Link href={`/events/${event.id}`} className="block group">
            <div className="flex flex-col gap-3">
                {/* Event Image */}
                <div className="relative aspect-[4/3] overflow-hidden rounded-xl bg-muted">
                    {event.image_url ? (
                        <img
                            src={`${process.env.NEXT_PUBLIC_API_URL?.replace('/api/v1', '') || 'http://localhost:8080'}${event.image_url}`}
                            alt={event.title}
                            className="w-full h-full object-cover group-hover:scale-105 transition-transform duration-500"
                        />
                    ) : (
                        <div className="w-full h-full flex items-center justify-center bg-secondary">
                            <Calendar className="h-12 w-12 text-muted-foreground/50" />
                        </div>
                    )}

                    {/* Top Right Actions/Badges */}
                    <div className="absolute top-3 right-3 flex gap-2" onClick={(e) => e.preventDefault()}>
                        <div 
                            className="bg-white/90 backdrop-blur-sm rounded-full p-2 shadow-sm hover:scale-110 transition-transform cursor-pointer"
                            onClick={(e) => {
                                e.preventDefault();
                                e.stopPropagation();
                                setShareModalOpen(true);
                            }}
                        >
                            <Share2 className="h-4 w-4 text-foreground" />
                        </div>
                        <div className="bg-white/90 backdrop-blur-sm rounded-full p-2 shadow-sm hover:scale-110 transition-transform cursor-pointer">
                            <Star className="h-4 w-4 text-foreground" />
                        </div>
                    </div>

                    {/* Status Badge (if not upcoming) */}
                    {!upcoming && (
                        <div className="absolute bottom-3 left-3">
                            <Badge variant="destructive" className="rounded-md shadow-sm">
                                Past Event
                            </Badge>
                        </div>
                    )}
                </div>

                {/* Content */}
                <div className="flex flex-col gap-1">
                    <div className="flex justify-between items-start gap-2">
                        <h3 className="font-semibold text-lg leading-tight text-foreground group-hover:text-primary transition-colors line-clamp-2">
                            {event.title}
                        </h3>
                        <div className="flex items-center gap-1 text-sm font-medium">
                            <Star className="h-3.5 w-3.5 fill-primary text-primary" />
                            <span>4.9</span>
                        </div>
                    </div>

                    <p className="text-muted-foreground text-sm line-clamp-1">
                        {event.venue}, {event.city}
                    </p>

                    <p className="text-muted-foreground text-sm">
                        {formatDate(event.start_date)} • {formatTime(event.start_date)}
                    </p>

                    <div className="mt-1 flex items-baseline gap-1">
                        {minPrice !== null ? (
                            <>
                                <span className="font-bold text-lg">{formatAmount(minPrice)}</span>
                                <span className="text-muted-foreground text-sm">/ person</span>
                            </>
                        ) : (
                            <span className="text-sm text-muted-foreground">Price TBA</span>
                        )}
                    </div>
                </div>
            </div>

            {/* Share Modal */}
            <ShareModal
                open={shareModalOpen}
                onOpenChange={setShareModalOpen}
                eventId={event.id}
                eventTitle={event.title}
            />
        </Link>
    );
}
