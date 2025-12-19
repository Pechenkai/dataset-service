import { fireEvent, render, screen, waitFor } from '@testing-library/react';
import React from 'react';
import { MemoryRouter } from 'react-router-dom';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import NotificationsPage from '../NotificationsPage';
import { ServiceContext, ServiceBag } from '../../context/ServiceContext';

const markRead = vi.fn().mockResolvedValue(undefined);
const serviceBag: ServiceBag = {
  apiClient: {} as any,
  authService: {} as any,
  datasetService: {} as any,
  notificationService: {
    list: () =>
      Promise.resolve({
        items: [
          { id: 1, message: 'Hello', is_read: false, user_id: 1 } as any
        ],
        meta: { total: 1, limit: 10, offset: 0 }
      }),
    markRead
  } as any,
  subscriptionService: {} as any
};

const Wrapper = ({ children }: { children: React.ReactNode }) => (
  <MemoryRouter>
    <QueryClientProvider client={new QueryClient({ defaultOptions: { queries: { retry: false } } })}>
      <ServiceContext.Provider value={serviceBag}>{children}</ServiceContext.Provider>
    </QueryClientProvider>
  </MemoryRouter>
);

describe('NotificationsPage', () => {
  it('renders notifications and allows mark read', async () => {
    render(<NotificationsPage />, { wrapper: Wrapper });

    expect(await screen.findByText('Hello')).toBeInTheDocument();
    fireEvent.click(screen.getByRole('button', { name: /mark read/i }));

    await waitFor(() => expect(markRead).toHaveBeenCalledWith(1, true));
  });
});
