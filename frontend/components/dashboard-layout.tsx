'use client';

import { Sidebar } from '@/components/sidebar';

interface DashboardLayoutProps {
  children: React.ReactNode;
}

export function DashboardLayout({ children }: DashboardLayoutProps) {
  return (
    <div className="h-screen flex overflow-hidden">
      <aside className="hidden md:flex border-r bg-muted/10 fixed h-full">
        <Sidebar />
      </aside>
      <main className="flex-1 overflow-y-auto md:ml-64">
        {children}
      </main>
    </div>
  );
}
