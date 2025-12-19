import { render, screen } from '@testing-library/react';
import React from 'react';
import { MemoryRouter } from 'react-router-dom';
import { AppHeader } from '../AppHeader';
import { AuthContext, AuthContextValue } from '../../../context/AuthContext';

const renderWithAuth = (ctx: Partial<AuthContextValue>) => {
  const value: AuthContextValue = {
    session: null,
    isAuthenticated: false,
    login: vi.fn().mockResolvedValue(undefined),
    logout: vi.fn().mockResolvedValue(undefined),
    ...ctx
  };

  render(
    <MemoryRouter>
      <AuthContext.Provider value={value}>
        <AppHeader />
      </AuthContext.Provider>
    </MemoryRouter>
  );
};

describe('AppHeader', () => {
  it('shows login button when guest', () => {
    renderWithAuth({ isAuthenticated: false });
    expect(screen.getByText(/Login/i)).toBeInTheDocument();
  });

  it('shows profile button when authenticated', () => {
    renderWithAuth({ isAuthenticated: true });
    expect(screen.getByText(/Categories/)).toBeInTheDocument();
    expect(screen.getByLabelText(/Profile/)).toBeInTheDocument();
  });
});
