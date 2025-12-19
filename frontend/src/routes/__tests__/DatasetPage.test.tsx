import { fireEvent, render, screen } from '@testing-library/react';
import React from 'react';
import { MemoryRouter, Route, Routes } from 'react-router-dom';
import DatasetPage from '../DatasetPage';

vi.mock('../../viewmodels/datasetViewModel', () => ({
  useDatasetViewModel: () => ({
    tab: 'versions',
    vpage: 1,
    rpage: 1,
    datasetQuery: { isLoading: false, data: { id: 1, name: 'Dataset X', description: 'Desc', is_public: true } },
    versionsQuery: { data: { items: [], meta: { total: 0, limit: 5, offset: 0 } } },
    reviewsQuery: { data: { items: [], meta: { total: 0, limit: 10, offset: 0 } } },
    createReview: { mutateAsync: vi.fn(), isLoading: false },
    subscribe: { mutate: vi.fn() },
    actions: { setTab: vi.fn(), setVersionsPage: vi.fn(), setReviewsPage: vi.fn() }
  })
}));

vi.mock('../../context/ServiceContext', () => ({
  useServices: () => ({ datasetService: { listCategories: () => Promise.resolve({ items: [] }) } })
}));

describe('DatasetPage', () => {
  it('renders dataset title and subscribe button', async () => {
    render(
      <MemoryRouter initialEntries={[{ pathname: '/datasets/1' }]}> 
        <Routes>
          <Route path="/datasets/:datasetId" element={<DatasetPage />} />
        </Routes>
      </MemoryRouter>
    );

    expect(screen.getByText('Dataset X')).toBeInTheDocument();
    expect(screen.getByText(/Subscribe/i)).toBeInTheDocument();
  });
});
