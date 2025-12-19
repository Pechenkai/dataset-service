import { fireEvent, render, screen } from '@testing-library/react';
import React from 'react';
import { MemoryRouter } from 'react-router-dom';
import AuthPage from '../AuthPage';
import { AuthContext, AuthContextValue } from '../../context/AuthContext';

const createWrapper = (ctx: Partial<AuthContextValue> = {}) => {
  const value: AuthContextValue = {
    session: null,
    isAuthenticated: false,
    login: vi.fn().mockResolvedValue(undefined),
    logout: vi.fn().mockResolvedValue(undefined),
    ...ctx
  };
  return ({ children }: { children: React.ReactNode }) => (
    <AuthContext.Provider value={value}>{children}</AuthContext.Provider>
  );
};

describe('AuthPage', () => {
  it('submits login form', async () => {
    const login = vi.fn().mockResolvedValue(undefined);
    const Wrapper = createWrapper({ login });

    render(
      <MemoryRouter>
        <AuthPage />
      </MemoryRouter>,
      { wrapper: Wrapper as any }
    );

    const email = screen.getByRole('textbox');
    const password = document.querySelector('input[type=\"password\"]') as HTMLInputElement;
    fireEvent.change(email, { target: { value: 'user@example.com' } });
    fireEvent.change(password, { target: { value: 'secret' } });
    fireEvent.click(screen.getByRole('button', { name: /Log in/i }));

    expect(login).toHaveBeenCalledWith({ email: 'user@example.com', password: 'secret' });
  });
});
