import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { render, screen, waitFor } from '@testing-library/react';
import React from 'react';
import { MemoryRouter } from 'react-router-dom';
import { useDatasetViewModel } from '../datasetViewModel';
import { ServiceContext, ServiceBag } from '../../context/ServiceContext';

describe('useDatasetViewModel', () => {
  const datasetService = {
    getDataset: vi.fn().mockResolvedValue({ id: 1, name: 'VM Dataset', is_public: true }),
    listVersions: vi.fn().mockResolvedValue({ items: [], meta: { total: 0, limit: 5, offset: 0 } }),
    listReviews: vi.fn().mockResolvedValue({ items: [], meta: { total: 0, limit: 10, offset: 0 } })
  } as any;

  const subscriptionService = {
    subscribe: vi.fn().mockResolvedValue({ id: 1 })
  } as any;

  const serviceBag: ServiceBag = {
    apiClient: {} as any,
    authService: {} as any,
    datasetService,
    notificationService: {} as any,
    subscriptionService
  };

  const Wrapper = ({ children }: { children: React.ReactNode }) => (
    <MemoryRouter initialEntries={[{ pathname: '/datasets/1' }]}> 
      <QueryClientProvider client={new QueryClient({ defaultOptions: { queries: { retry: false } } })}>
        <ServiceContext.Provider value={serviceBag}>{children}</ServiceContext.Provider>
      </QueryClientProvider>
    </MemoryRouter>
  );

  it('fetches dataset data', async () => {
    const Probe: React.FC = () => {
      const vm = useDatasetViewModel(1);
      return <div>{vm.datasetQuery.data?.name || 'loading'}</div>;
    };

    render(<Probe />, { wrapper: Wrapper });

    expect(screen.getByText('loading')).toBeInTheDocument();
    await waitFor(() => expect(screen.getByText('VM Dataset')).toBeInTheDocument());
  });
});
