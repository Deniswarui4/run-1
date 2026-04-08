'use client';

import Link from 'next/link';
import { usePathname } from 'next/navigation';
import { useAuth } from '@/lib/auth-context';
import { cn } from '@/lib/utils';
import { Button } from '@/components/ui/button';
import { Separator } from '@/components/ui/separator';
import {
  LayoutDashboard,
  Calendar,
  Plus,
  Ticket,
  DollarSign,
  Users,
  Settings,
  CheckCircle,
  BarChart3,
  CreditCard,
  UserCheck,
  Shield,
  Tag,
  LogOut
} from 'lucide-react';

interface SidebarProps {
  className?: string;
}

export function Sidebar({ className }: SidebarProps) {
  const pathname = usePathname();
  const { user, logout } = useAuth();

  if (!user) return null;

  const isActive = (path: string) => pathname === path;

  const organizerNavItems = [
    {
      title: 'Dashboard',
      href: '/organizer',
      icon: LayoutDashboard,
    },
    {
      title: 'My Events',
      href: '/organizer/events',
      icon: Calendar,
    },
    {
      title: 'Create Event',
      href: '/organizer/events/create',
      icon: Plus,
    },
    {
      title: 'Statistics',
      href: '/organizer/stats',
      icon: BarChart3,
    },
    {
      title: 'Earnings',
      href: '/organizer/earnings',
      icon: DollarSign,
    },
    {
      title: 'Withdrawals',
      href: '/organizer/withdrawals',
      icon: CreditCard,
    },
    {
      title: 'Settings',
      href: '/organizer/settings',
      icon: Settings,
    },
  ];

  const adminNavItems = [
    {
      title: 'Dashboard',
      href: '/admin',
      icon: LayoutDashboard,
    },
    {
      title: 'Events',
      href: '/admin/events',
      icon: Calendar,
    },
    {
      title: 'Categories',
      href: '/admin/categories',
      icon: Tag,
    },
    {
      title: 'Users',
      href: '/admin/users',
      icon: Users,
    },
    {
      title: 'Withdrawals',
      href: '/admin/withdrawals',
      icon: CreditCard,
    },
    {
      title: 'Platform Stats',
      href: '/admin/stats',
      icon: BarChart3,
    },
    {
      title: 'Settings',
      href: '/admin/settings',
      icon: Settings,
    },
  ];

  const moderatorNavItems = [
    {
      title: 'Dashboard',
      href: '/moderator',
      icon: LayoutDashboard,
    },
    {
      title: 'Pending Events',
      href: '/moderator/events/pending',
      icon: CheckCircle,
    },
    {
      title: 'My Reviews',
      href: '/moderator/reviews',
      icon: UserCheck,
    },
    {
      title: 'Statistics',
      href: '/moderator/stats',
      icon: BarChart3,
    },
  ];

  const attendeeNavItems = [
    {
      title: 'Dashboard',
      href: '/dashboard',
      icon: LayoutDashboard,
    },
    {
      title: 'Browse Events',
      href: '/events',
      icon: Calendar,
    },
  ];

  const getNavItems = () => {
    switch (user.role) {
      case 'organizer':
        return organizerNavItems;
      case 'admin':
        return adminNavItems;
      case 'moderator':
        return moderatorNavItems;
      case 'attendee':
        return attendeeNavItems;
      default:
        return [];
    }
  };

  const navItems = getNavItems();

  if (navItems.length === 0) return null;

  return (
    <div className={cn('w-64 h-screen flex flex-col', className)}>
      {/* Logo at top */}
      <div className="p-6 border-b">
        <Link href="/" className="flex items-center space-x-2">
          <div className="bg-primary rounded-lg p-1.5">
            <Ticket className="h-5 w-5 text-primary-foreground" />
          </div>
          <span className="text-xl font-bold tracking-tight text-primary">
            Runtown
          </span>
        </Link>
      </div>

      {/* Main navigation - scrollable if needed */}
      <div className="flex-1 overflow-y-auto">
        <div className="space-y-4 py-4">
          <div className="px-3 py-2">
            <div className="flex items-center gap-2 mb-4">
              <Shield className="h-5 w-5" />
              <h2 className="text-lg font-semibold capitalize">
                {user.role} Panel
              </h2>
            </div>
            <div className="space-y-1">
              {navItems.map((item) => (
                <Link key={item.href} href={item.href}>
                  <Button
                    variant={isActive(item.href) ? 'secondary' : 'ghost'}
                    className={cn(
                      'w-full justify-start',
                      isActive(item.href) && 'bg-secondary'
                    )}
                  >
                    <item.icon className="mr-2 h-4 w-4" />
                    {item.title}
                  </Button>
                </Link>
              ))}
            </div>
          </div>

          <Separator />

          <div className="px-3 py-2">
            <h3 className="mb-2 px-4 text-sm font-semibold text-muted-foreground">
              General
            </h3>
            <div className="space-y-1">
              <Link href="/profile">
                <Button
                  variant={isActive('/profile') ? 'secondary' : 'ghost'}
                  className="w-full justify-start"
                >
                  <Settings className="mr-2 h-4 w-4" />
                  Profile
                </Button>
              </Link>
              <Link href="/my-tickets">
                <Button
                  variant={isActive('/my-tickets') ? 'secondary' : 'ghost'}
                  className="w-full justify-start"
                >
                  <Ticket className="mr-2 h-4 w-4" />
                  My Tickets
                </Button>
              </Link>
              <Link href="/transactions">
                <Button
                  variant={isActive('/transactions') ? 'secondary' : 'ghost'}
                  className="w-full justify-start"
                >
                  <CreditCard className="mr-2 h-4 w-4" />
                  Transactions
                </Button>
              </Link>
              <Link href="/events">
                <Button
                  variant={isActive('/events') ? 'secondary' : 'ghost'}
                  className="w-full justify-start"
                >
                  <Calendar className="mr-2 h-4 w-4" />
                  Browse Events
                </Button>
              </Link>
            </div>
          </div>
        </div>
      </div>

      {/* Logout at bottom */}
      <div className="p-4 border-t mt-auto">
        <Button
          variant="ghost"
          className="w-full justify-start text-destructive hover:text-destructive hover:bg-destructive/10"
          onClick={logout}
        >
          <LogOut className="mr-2 h-4 w-4" />
          Log out
        </Button>
      </div>
    </div>
  );
}
