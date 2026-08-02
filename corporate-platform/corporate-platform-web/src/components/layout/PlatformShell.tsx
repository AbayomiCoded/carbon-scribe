'use client'

import { ReactNode } from 'react'
import { usePathname } from 'next/navigation'
import { useAuth } from '@/contexts/AuthContext'
import { ConnectivityProvider } from '@/contexts/ConnectivityContext'
import RouteGuard from '@/components/auth/RouteGuard'
import CorporateNavbar from '@/components/layout/CorporateNavbar'
import CorporateSidebar from '@/components/layout/CorporateSidebar'
import AuthNavbar from '@/components/layout/AuthNavbar'
import ConnectionStatus from '@/components/layout/ConnectionStatus'
import SessionExpiryBanner from '@/components/layout/SessionExpiryBanner'
import { ClientOnly } from '@/components/common/ClientOnly'

const PUBLIC_ROUTES = ['/login', '/register', '/forgot-password', '/reset-password']

function isPublicRoute(path: string): boolean {
  return PUBLIC_ROUTES.some((route) => path.startsWith(route))
}

interface PlatformShellProps {
  children: ReactNode
}

export default function PlatformShell({ children }: PlatformShellProps) {
  const pathname = usePathname() || '/'
  const { isLoading, isAuthenticated } = useAuth()
  const publicRoute = isPublicRoute(pathname)

  // Public routes - render with AuthNavbar
  if (publicRoute) {
    return (
      <ConnectivityProvider>
        <div className="flex min-h-screen flex-col">
          <AuthNavbar />
          <main className="flex-1" role="main">
            {children}
          </main>
        </div>
      </ConnectivityProvider>
    )
  }

  // Loading state - show skeleton
  if (isLoading) {
    return (
      <ConnectivityProvider>
        <div
          className="flex min-h-screen items-center justify-center text-gray-500 dark:text-gray-400"
          role="status"
          aria-live="polite"
        >
          <div className="flex flex-col items-center gap-3">
            <div
              className="h-8 w-8 animate-spin rounded-full border-4 border-corporate-blue border-t-transparent"
              aria-hidden="true"
            />
            <span>Loading session...</span>
          </div>
        </div>
      </ConnectivityProvider>
    )
  }

  // Not authenticated - render nothing (redirect handled by AuthContext)
  if (!isAuthenticated) {
    return null
  }

  // Authenticated - render full shell with hydration-safe content
  return (
    <ConnectivityProvider>
      <div className="flex min-h-screen">
        <CorporateSidebar />
        <div className="flex flex-1 flex-col">
          <ClientOnly fallback={<div className="h-12" />}>
            <SessionExpiryBanner />
          </ClientOnly>
          <CorporateNavbar />
          <ConnectionStatus />
          <main
            className="flex-1 overflow-auto p-4 md:p-6 lg:p-8"
            role="main"
            aria-label="Main content"
          >
            <div className="mx-auto w-full max-w-7xl">
              <RouteGuard>{children}</RouteGuard>
            </div>
          </main>
        </div>
      </div>
    </ConnectivityProvider>
  )
}