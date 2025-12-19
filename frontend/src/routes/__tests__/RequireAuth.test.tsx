import { render, screen } from '@testing-library/react';
import React from 'react';
import { MemoryRouter, Route, Routes } from 'react-router-dom';
import { AuthContext, AuthContextValue } from '../../context/AuthContext';
import { RequireAuth } from '../RequireAuth';

const renderWithAuth = (ctx: Partial<AuthContextValue>) => {
  const value: AuthContextValue = {
    session: null,
    isAuthenticated: false,
    login: vi.fn().mockResolvedValue(undefined),
    logout: vi.fn().mockResolvedValue(undefined),
    ...ctx
  };

  render(
    <AuthContext.Provider value={value}>
      <MemoryRouter initialEntries={[{ pathname: '/private', search: '?q=1' }] }>
        <Routes>
          <Route
            path="/private"
            element={
              <RequireAuth>
                <div>Secret</div>
              </RequireAuth>
            }
          />
          <Route path="/auth" element={<div>Auth page</div>} />
        </Routes>
      </MemoryRouter>
    </AuthContext.Provider>
  );
};

describe('RequireAuth', () => {
  it('redirects unauthenticated users', () => {
    renderWithAuth({ isAuthenticated: false });
    expect(screen.getByText('Auth page')).toBeInTheDocument();
  });

  it('renders children when authenticated', () => {
    renderWithAuth({ isAuthenticated: true });
    expect(screen.getByText('Secret')).toBeInTheDocument();
  });
});
