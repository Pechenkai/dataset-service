import { render, screen } from '@testing-library/react';
import React from 'react';
import { MemoryRouter } from 'react-router-dom';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import CategoriesPage from '../CategoriesPage';
import { ServiceContext, ServiceBag } from '../../context/ServiceContext';
import { AuthContext, AuthContextValue } from '../../context/AuthContext';

const queryClient = new QueryClient({ defaultOptions: { queries: { retry: false } } });

const serviceBag: ServiceBag = {
  apiClient: {} as any,
  authService: {} as any,
  datasetService: {
    listCategories: () =>
      Promise.resolve({
        items: [
          { id: 1, name: 'CV', description: 'Vision' },
          { id: 2, name: 'NLP', description: 'Text' }
        ],
        meta: { total: 2, limit: 100, offset: 0 }
      })
  } as any,
  notificationService: {} as any,
  subscriptionService: {} as any
};

const authValue: AuthContextValue = {
  session: { token: 't', user: { id: 1, email: 'u', username: 'u', role: 'admin', is_blocked: false } as any },
  isAuthenticated: true,
  login: vi.fn() as any,
  registerAndLogin: vi.fn() as any,
  logout: vi.fn() as any
};

const Wrapper = ({ children }: { children: React.ReactNode }) => (
  <MemoryRouter>
    <QueryClientProvider client={queryClient}>
      <ServiceContext.Provider value={serviceBag}>
        <AuthContext.Provider value={authValue}>{children}</AuthContext.Provider>
      </ServiceContext.Provider>
    </QueryClientProvider>
  </MemoryRouter>
);

describe('CategoriesPage', () => {
  it('renders categories table', async () => {
    render(<CategoriesPage />, { wrapper: Wrapper });

    expect(await screen.findByText('CV')).toBeInTheDocument();
    expect(screen.getByText('Vision')).toBeInTheDocument();
  });
});
