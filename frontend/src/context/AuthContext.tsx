import React, { createContext, useContext, useEffect, useMemo, useState } from 'react';
import { useServices } from './ServiceContext';
import { AuthSession } from '../services/authService';
import { AuthenticateRequest } from '../api/types';

export interface AuthContextValue {
  session: AuthSession | null;
  isAuthenticated: boolean;
  login: (payload: AuthenticateRequest) => Promise<void>;
  registerAndLogin: (payload: { username: string; email: string; password: string; country: string }) => Promise<void>;
  logout: () => Promise<void>;
}

export const AuthContext = createContext<AuthContextValue | undefined>(undefined);

export const AuthProvider: React.FC<{ children: React.ReactNode }> = ({ children }) => {
  const { authService } = useServices();
  const [session, setSession] = useState<AuthSession | null>(() => authService.loadSession());

  useEffect(() => {
    authService.persistSession(session);
  }, [session, authService]);

  const value = useMemo<AuthContextValue>(() => {
    return {
      session,
      isAuthenticated: Boolean(session?.token),
      login: async (payload) => {
        const result = await authService.authenticate(payload);
        const next = { token: result.token, user: result.user, expires_at: result.expires_at };
        authService.persistSession(next);
        setSession(next);
      },
      registerAndLogin: async ({ username, email, password, country }) => {
        await authService.register({ username, email, password, country });
        const result = await authService.authenticate({ email, password });
        const next = { token: result.token, user: result.user, expires_at: result.expires_at };
        authService.persistSession(next);
        setSession(next);
      },
      logout: async () => {
        authService.persistSession(null);
        setSession(null);
      }
    };
  }, [authService, session]);

  return <AuthContext.Provider value={value}>{children}</AuthContext.Provider>;
};

export const useAuth = () => {
  const ctx = useContext(AuthContext);
  if (!ctx) throw new Error('useAuth must be used within AuthProvider');
  return ctx;
};
