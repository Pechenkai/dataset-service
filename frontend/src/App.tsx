import React, { useMemo } from 'react';
import { createBrowserRouter, Navigate, Outlet, RouterProvider } from 'react-router-dom';
import { QueryClientProvider } from '@tanstack/react-query';
import { createApiClient } from './api/client';
import { queryClient } from './api/queryClient';
import AuthPage from './routes/AuthPage';
import CatalogPage from './routes/CatalogPage';
import DatasetPage from './routes/DatasetPage';
import NotificationsPage from './routes/NotificationsPage';
import NotFoundPage from './routes/NotFoundPage';
import UploadPage from './routes/UploadPage';
import ProfilePage from './routes/ProfilePage';
import MyDatasetsPage from './routes/MyDatasetsPage';
import AdminUsersPage from './routes/AdminUsersPage';
import { AuthProvider } from './context/AuthContext';
import { ServiceProvider } from './context/ServiceContext';
import { AUTH_STORAGE_KEY } from './services/authService';
import { RequireAuth } from './routes/RequireAuth';
import CategoriesPage from './routes/CategoriesPage';
import { AppHeader } from './components/layout/AppHeader';
import { AuthLayout } from './components/layout/AuthLayout';

const RootLayout = () => (
  <>
    <AppHeader />
    <div className="page-container">
      <Outlet />
    </div>
  </>
);

const router = createBrowserRouter(
  [
    {
      path: '/',
      element: <RootLayout />,
      children: [
        { index: true, element: <Navigate to="/catalog" replace /> },
        { path: 'catalog', element: <CatalogPage /> },
        { path: 'categories', element: <CategoriesPage /> },
        { path: 'datasets/:datasetId', element: <DatasetPage /> },
        {
          path: 'upload',
          element: (
            <RequireAuth>
              <UploadPage />
            </RequireAuth>
          )
        },
        {
          path: 'notifications',
          element: (
            <RequireAuth>
              <NotificationsPage />
            </RequireAuth>
          )
        },
        {
          path: 'my',
          element: (
            <RequireAuth>
              <MyDatasetsPage />
            </RequireAuth>
          )
        },
        {
          path: 'profile',
          element: (
            <RequireAuth>
              <ProfilePage />
            </RequireAuth>
          )
        },
        {
          path: 'admin/users',
          element: (
            <RequireAuth>
              <AdminUsersPage />
            </RequireAuth>
          )
        },
        { path: '*', element: <NotFoundPage /> }
      ]
    },
    {
      path: '/auth',
      element: (
        <AuthLayout>
          <AuthPage />
        </AuthLayout>
      )
    }
  ],
  { basename: '/app' }
);

const readStoredToken = () => {
  try {
    const raw = localStorage.getItem(AUTH_STORAGE_KEY);
    if (!raw) return null;
    const parsed = JSON.parse(raw);
    return parsed.token as string;
  } catch (e) {
    console.warn('Failed to read auth token', e);
    return null;
  }
};

const App = () => {
  const apiClient = useMemo(() => createApiClient(readStoredToken), []);

  return (
    <QueryClientProvider client={queryClient}>
      <ServiceProvider apiClient={apiClient}>
        <AuthProvider>
          <RouterProvider router={router} />
        </AuthProvider>
      </ServiceProvider>
    </QueryClientProvider>
  );
};

export default App;
