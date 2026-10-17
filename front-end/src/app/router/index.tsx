import { lazy } from 'react'
import { createBrowserRouter, Navigate } from 'react-router-dom'

import { AuthLayout } from '@/components/layouts/AuthLayout'
import { MainLayout } from '@/components/layouts/MainLayout'
import { PublicLayout } from '@/components/layouts/PublicLayout'

import { GuestRoute } from './GuestRoute'
import { ProtectedRoute } from './ProtectedRoute'

const LoginPage = lazy(() => import('@/features/auth/pages/LoginPage'))
const AccountPage = lazy(() => import('@/features/auth/pages/AccountPage'))
const StudentDashboardPage = lazy(() => import('@/features/dashboard/pages/StudentDashboardPage'))
const PracticePage = lazy(() => import('@/features/practice/pages/PracticePage'))
const SubmissionsPage = lazy(() => import('@/features/submissions/pages/SubmissionsPage'))
const ShareLinksPage = lazy(() => import('@/features/share-links/pages/ShareLinksPage'))
const ShareLinkDetailPage = lazy(() => import('@/features/share-links/pages/ShareLinkDetailPage'))
const PublicSharePage = lazy(() => import('@/features/public-share/pages/PublicSharePage'))
const AuthDebugPage = lazy(() => import('@/features/auth-debug/pages/AuthDebugPage'))

// The debug panel is only available in development.
const devRoutes = import.meta.env.DEV
  ? [{ path: '/auth/debug', element: <AuthDebugPage /> }]
  : []

export const router = createBrowserRouter([
  {
    element: <GuestRoute />,
    children: [
      {
        path: '/login',
        element: <AuthLayout />,
        children: [{ index: true, element: <LoginPage /> }],
      },
    ],
  },
  {
    element: <ProtectedRoute />,
    children: [
      {
        element: <MainLayout />,
        children: [
          { index: true, element: <Navigate to="/dashboard" replace /> },
          { path: '/dashboard', element: <StudentDashboardPage /> },
          { path: '/practice/:id', element: <PracticePage /> },
          { path: '/submissions', element: <SubmissionsPage /> },
          { path: '/share-links', element: <ShareLinksPage /> },
          { path: '/share-links/:code', element: <ShareLinkDetailPage /> },
          { path: '/account', element: <AccountPage /> },
          ...devRoutes,
        ],
      },
    ],
  },
  {
    element: <PublicLayout />,
    children: [{ path: '/shared/:code', element: <PublicSharePage /> }],
  },
  { path: '*', element: <Navigate to="/" replace /> },
])
