'use client';

import Link from 'next/link';
import { useAuth } from '@/lib/auth-context';
import { Button } from '@/components/ui/button';
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuLabel,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from '@/components/ui/dropdown-menu';
import { Avatar, AvatarFallback } from '@/components/ui/avatar';
import { Ticket, User, LogOut, LayoutDashboard, Settings, CreditCard, Menu, Globe, ChevronDown, Music, Palette, Utensils, Gamepad2, Dumbbell, Briefcase, Calendar } from 'lucide-react';
import { useState, useEffect } from 'react';
import { Sheet, SheetContent, SheetTrigger } from '@/components/ui/sheet';
import { apiClient } from '@/lib/api-client';
import { Category } from '@/lib/types';

export function Navbar() {
  const { user, logout } = useAuth();
  const [categories, setCategories] = useState<Category[]>([]);
  const [loadingCategories, setLoadingCategories] = useState(true);

  useEffect(() => {
    loadCategories();
  }, []);

  const loadCategories = async () => {
    try {
      const data = await apiClient.getCategories();
      setCategories(data.filter(c => c.is_active));
    } catch (error) {
      console.error('Failed to load categories:', error);
    } finally {
      setLoadingCategories(false);
    }
  };

  const getInitials = (firstName: string, lastName: string) => {
    return `${firstName.charAt(0)}${lastName.charAt(0)}`.toUpperCase();
  };

  const getDashboardLink = () => {
    switch (user?.role) {
      case 'admin': return '/admin';
      case 'moderator': return '/moderator';
      case 'organizer': return '/organizer';
      case 'attendee': return '/dashboard';
      default: return '/';
    }
  };

  const getCategoryIcon = (categoryName: string) => {
    const name = categoryName.toLowerCase();
    if (name.includes('music') || name.includes('concert')) return Music;
    if (name.includes('art') || name.includes('exhibition')) return Palette;
    if (name.includes('food') || name.includes('drink')) return Utensils;
    if (name.includes('game') || name.includes('gaming')) return Gamepad2;
    if (name.includes('sport') || name.includes('fitness')) return Dumbbell;
    if (name.includes('business') || name.includes('conference')) return Briefcase;
    return Calendar;
  };

  return (
    <header className={`fixed top-0 left-0 right-0 z-50 transition-all duration-200 bg-background/80 backdrop-blur-md border-b border-border/40`} suppressHydrationWarning>
      <div className="container mx-auto px-4 h-16 flex items-center justify-between" suppressHydrationWarning>
        {/* Logo & Categories */}
        <div className="flex items-center space-x-4">
          <Link href="/" className="flex items-center space-x-2">
            <div className="bg-primary rounded-lg p-1.5">
              <Ticket className="h-5 w-5 text-primary-foreground" />
            </div>
            <span className="text-xl font-bold tracking-tight text-primary">
              Runtown
            </span>
          </Link>

          {/* Categories Dropdown */}
          <DropdownMenu>
            <DropdownMenuTrigger asChild suppressHydrationWarning>
              <Button variant="ghost" size="sm" className="hidden md:flex items-center gap-1 font-medium" suppressHydrationWarning>
                Categories
                <ChevronDown className="h-4 w-4" />
              </Button>
            </DropdownMenuTrigger>
            <DropdownMenuContent align="start" className="w-64 max-h-[400px] overflow-y-auto">
              <DropdownMenuLabel>Browse by Category</DropdownMenuLabel>
              <DropdownMenuSeparator />
              {loadingCategories ? (
                <div className="px-2 py-3 text-sm text-muted-foreground">Loading...</div>
              ) : categories.length > 0 ? (
                categories.map((category) => {
                  const IconComponent = getCategoryIcon(category.name);
                  return (
                    <DropdownMenuItem key={category.id} asChild>
                      <Link
                        href={`/events?category=${encodeURIComponent(category.name)}`}
                        className="cursor-pointer flex items-center gap-3 py-2"
                      >
                        <div
                          className="w-8 h-8 rounded-lg flex items-center justify-center flex-shrink-0"
                          style={{
                            backgroundColor: `${category.color || '#3B82F6'}20`,
                            color: category.color || '#3B82F6'
                          }}
                        >
                          <IconComponent className="h-4 w-4" />
                        </div>
                        <span className="flex-1">{category.name}</span>
                      </Link>
                    </DropdownMenuItem>
                  );
                })
              ) : (
                <div className="px-2 py-3 text-sm text-muted-foreground">No categories available</div>
              )}
              <DropdownMenuSeparator />
              <DropdownMenuItem asChild>
                <Link href="/events" className="cursor-pointer font-medium text-primary">
                  View All Events
                </Link>
              </DropdownMenuItem>
            </DropdownMenuContent>
          </DropdownMenu>
        </div>

        {/* Desktop Navigation */}
        <div className="hidden md:flex items-center space-x-6">
          <Link href="/events" className="text-sm font-medium text-muted-foreground hover:text-primary transition-colors">
            Browse Events
          </Link>
          <Link href="/help" className="text-sm font-medium text-muted-foreground hover:text-primary transition-colors">
            Help Center
          </Link>

          <div className="h-6 w-px bg-border/60 mx-2" />

          <Button variant="ghost" size="sm" className="text-muted-foreground hover:text-primary">
            <Globe className="h-4 w-4 mr-2" />
            EN
          </Button>

          {user ? (
            <div className="flex items-center space-x-4">
              <Link href={getDashboardLink()}>
                <Button variant="ghost" size="sm" className="font-medium">
                  Dashboard
                </Button>
              </Link>

              <DropdownMenu>
                <DropdownMenuTrigger asChild suppressHydrationWarning>
                  <Button variant="ghost" className="relative h-9 w-9 rounded-full ring-1 ring-border hover:ring-primary/20 transition-all p-0 overflow-hidden" suppressHydrationWarning>
                    <Avatar className="h-full w-full">
                      <AvatarFallback className="bg-primary/5 text-primary font-medium text-xs">
                        {getInitials(user.first_name, user.last_name)}
                      </AvatarFallback>
                    </Avatar>
                  </Button>
                </DropdownMenuTrigger>
                <DropdownMenuContent align="end" className="w-56 p-2">
                  <DropdownMenuLabel className="font-normal">
                    <div className="flex flex-col space-y-1">
                      <p className="text-sm font-medium leading-none">
                        {user.first_name} {user.last_name}
                      </p>
                      <p className="text-xs leading-none text-muted-foreground">
                        {user.email}
                      </p>
                    </div>
                  </DropdownMenuLabel>
                  <DropdownMenuSeparator />
                  <DropdownMenuItem asChild>
                    <Link href="/profile" className="cursor-pointer">
                      <User className="mr-2 h-4 w-4" />
                      Profile
                    </Link>
                  </DropdownMenuItem>
                  <DropdownMenuItem asChild>
                    <Link href="/my-tickets" className="cursor-pointer">
                      <Ticket className="mr-2 h-4 w-4" />
                      My Tickets
                    </Link>
                  </DropdownMenuItem>
                  <DropdownMenuItem asChild>
                    <Link href="/transactions" className="cursor-pointer">
                      <CreditCard className="mr-2 h-4 w-4" />
                      Transactions
                    </Link>
                  </DropdownMenuItem>
                  {user.role === 'organizer' && (
                    <DropdownMenuItem asChild>
                      <Link href="/organizer/settings" className="cursor-pointer">
                        <Settings className="mr-2 h-4 w-4" />
                        Settings
                      </Link>
                    </DropdownMenuItem>
                  )}
                  <DropdownMenuSeparator />
                  <DropdownMenuItem onClick={logout} className="text-destructive focus:text-destructive cursor-pointer">
                    <LogOut className="mr-2 h-4 w-4" />
                    Log out
                  </DropdownMenuItem>
                </DropdownMenuContent>
              </DropdownMenu>
            </div>
          ) : (
            <div className="flex items-center space-x-2">
              <Link href="/login">
                <Button variant="ghost" size="sm" className="font-medium">Log in</Button>
              </Link>
              <Link href="/register">
                <Button size="sm" className="font-medium shadow-none">
                  Sign up
                </Button>
              </Link>
            </div>
          )}
        </div>

        {/* Mobile Navigation */}
        <div className="md:hidden">
          <Sheet>
            <SheetTrigger asChild suppressHydrationWarning>
              <Button variant="ghost" size="icon" suppressHydrationWarning>
                <Menu className="h-5 w-5" />
              </Button>
            </SheetTrigger>
            <SheetContent side="right" className="w-[300px] sm:w-[400px]">
              <div className="flex flex-col space-y-6 mt-6">
                <div className="flex items-center space-x-2">
                  <div className="bg-primary rounded-lg p-1.5">
                    <Ticket className="h-5 w-5 text-primary-foreground" />
                  </div>
                  <span className="text-lg font-bold text-primary">Runtown</span>
                </div>

                <div className="flex flex-col space-y-3">
                  <Link href="/events">
                    <Button variant="ghost" className="w-full justify-start text-base">
                      Browse Events
                    </Button>
                  </Link>
                  <Link href="/help">
                    <Button variant="ghost" className="w-full justify-start text-base">
                      Help Center
                    </Button>
                  </Link>
                </div>

                <div className="h-px bg-border/50" />

                {user ? (
                  <div className="flex flex-col space-y-3">
                    <div className="flex items-center space-x-3 px-2 py-2">
                      <Avatar className="h-8 w-8">
                        <AvatarFallback className="bg-primary/10 text-primary text-xs">
                          {getInitials(user.first_name, user.last_name)}
                        </AvatarFallback>
                      </Avatar>
                      <div>
                        <p className="font-medium text-sm">{user.first_name} {user.last_name}</p>
                        <p className="text-xs text-muted-foreground">{user.email}</p>
                      </div>
                    </div>
                    <Link href={getDashboardLink()}>
                      <Button variant="ghost" className="w-full justify-start">
                        <LayoutDashboard className="h-4 w-4 mr-2" />
                        Dashboard
                      </Button>
                    </Link>
                    <Link href="/my-tickets">
                      <Button variant="ghost" className="w-full justify-start">
                        <Ticket className="h-4 w-4 mr-2" />
                        My Tickets
                      </Button>
                    </Link>
                    <Link href="/profile">
                      <Button variant="ghost" className="w-full justify-start">
                        <User className="h-4 w-4 mr-2" />
                        Profile
                      </Button>
                    </Link>
                    <Button variant="ghost" className="w-full justify-start text-destructive hover:text-destructive" onClick={logout}>
                      <LogOut className="h-4 w-4 mr-2" />
                      Log out
                    </Button>
                  </div>
                ) : (
                  <div className="flex flex-col space-y-2">
                    <Link href="/login">
                      <Button variant="outline" className="w-full">Log in</Button>
                    </Link>
                    <Link href="/register">
                      <Button className="w-full">Sign up</Button>
                    </Link>
                  </div>
                )}
              </div>
            </SheetContent>
          </Sheet>
        </div>
      </div>
    </header>
  );
}
