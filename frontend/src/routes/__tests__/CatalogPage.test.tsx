import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { render, screen, waitFor } from '@testing-library/react';
import React from 'react';
import { MemoryRouter } from 'react-router-dom';
import { Dataset } from '../../api/types';
import { ServiceContext } from '../../context/ServiceContext';
import CatalogPage from '../CatalogPage';

const createWrapper = (datasets: Dataset[] = []) => {
  const queryClient = new QueryClient({ defaultOptions: { queries: { retry: false } } });
  const serviceBag = {
    datasetService: {
      listCategories: () =>
        Promise.resolve({ items: [{ id: 1, name: 'CV', description: '' }], meta: { total: 1, limit: 100, offset: 0 } }),
      listDatasets: () =>
        Promise.resolve({ items: datasets, meta: { total: datasets.length, limit: 12, offset: 0 } }),
      buildFiltersFromSearch: () => ({}),
      stringifyFilters: () => new URLSearchParams()
    },
    notificationService: { list: () => Promise.resolve({ items: [], meta: { total: 0, offset: 0, limit: 10 } }) },
    subscriptionService: {
      listByUser: () => Promise.resolve({ items: [], meta: { total: 0, offset: 0, limit: 10 } }),
      subscribe: () => Promise.resolve(),
      unsubscribe: () => Promise.resolve()
    }
  } as any;

  return ({ children }: { children: React.ReactNode }) => (
    <MemoryRouter>
      <QueryClientProvider client={queryClient}>
        <ServiceContext.Provider value={serviceBag}>{children}</ServiceContext.Provider>
      </QueryClientProvider>
    </MemoryRouter>
  );
};

describe('CatalogPage', () => {
  it('renders datasets from view model', async () => {
    const dataset: Dataset = {
      id: 1,
      name: 'Demo dataset',
      description: 'Synthetic data',
      category_id: 1,
      owner_id: 1,
      is_public: true
    };

    render(<CatalogPage />, { wrapper: createWrapper([dataset]) });

    await waitFor(() => screen.getByText('Demo dataset'));
    expect(screen.getByText('Demo dataset')).toBeInTheDocument();
    expect(screen.getByText(/Synthetic/)).toBeInTheDocument();
  });
});
