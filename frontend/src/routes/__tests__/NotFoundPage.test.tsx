import { render, screen } from '@testing-library/react';
import React from 'react';
import NotFoundPage from '../NotFoundPage';
import { MemoryRouter } from 'react-router-dom';

describe('NotFoundPage', () => {
  it('renders fallback text and link', () => {
    render(
      <MemoryRouter>
        <NotFoundPage />
      </MemoryRouter>
    );

    expect(screen.getByText(/Страница не найдена/)).toBeInTheDocument();
    expect(screen.getByText(/К каталогу/)).toBeInTheDocument();
  });
});
