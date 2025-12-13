import React, { useMemo } from 'react';
import { createBrowserRouter, Outlet, RouterProvider } from 'react-router-dom';
import { QueryClientProvider } from '@tanstack/react-query';
import { createApiClient } from './api/client';
import { queryClient } from './api/queryClient';
import { AppShell } from './components/layout/AppShell';
import AuthPage from './routes/AuthPage';
import CatalogPage from './routes/CatalogPage';
import DatasetPage from './routes/DatasetPage';
import NotificationsPage from './routes/NotificationsPage';
import NotFoundPage from './routes/NotFoundPage';
import UploadPage from './routes/UploadPage';
import { AuthProvider } from './context/AuthContext';
import { ServiceProvider } from './context/ServiceContext';
import { AUTH_STORAGE_KEY, AuthService } from './services/authService';

const RootLayout = () => (
  <AppShell>
    <Outlet />
  </AppShell>
);

const router = createBrowserRouter(
  [
    {
      path: '/',
      element: <RootLayout />,
      children: [
        { index: true, element: <CatalogPage /> },
        { path: 'datasets/:datasetId', element: <DatasetPage /> },
        { path: 'upload', element: <UploadPage /> },
        { path: 'auth', element: <AuthPage /> },
        { path: 'notifications', element: <NotificationsPage /> },
        { path: '*', element: <NotFoundPage /> }
      ]
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
  const authService = useMemo(() => new AuthService(apiClient), [apiClient]);

  return (
    <QueryClientProvider client={queryClient}>
      <AuthProvider authService={authService}>
        <ServiceProvider apiClient={apiClient}>
          <RouterProvider router={router} />
        </ServiceProvider>
      </AuthProvider>
    </QueryClientProvider>
  );
};

export default App;
