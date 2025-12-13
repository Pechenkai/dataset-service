import React, { createContext, useContext, useEffect, useMemo, useState } from 'react';
import { AuthService, AuthSession } from '../services/authService';
import { AuthenticateRequest } from '../api/types';

interface AuthContextValue {
  session: AuthSession | null;
  isAuthenticated: boolean;
  login: (payload: AuthenticateRequest) => Promise<void>;
  logout: () => Promise<void>;
}

const AuthContext = createContext<AuthContextValue | undefined>(undefined);

export const AuthProvider: React.FC<{ authService: AuthService; children: React.ReactNode }> = ({
  authService,
  children
}) => {
  const [session, setSession] = useState<AuthSession | null>(() => authService.loadSession());

  useEffect(() => {
    authService.persistSession(session);
  }, [session, authService]);

  const value = useMemo<AuthContextValue>(
    () => ({
      session,
      isAuthenticated: Boolean(session?.token),
      login: async (payload) => {
        const result = await authService.authenticate(payload);
        setSession({ token: result.token, user: result.user, expires_at: result.expires_at });
      },
      logout: async () => {
        if (session?.token) {
          await authService.revoke();
        }
        setSession(null);
      }
    }),
    [authService, session]
  );

  return <AuthContext.Provider value={value}>{children}</AuthContext.Provider>;
};

export const useAuth = () => {
  const ctx = useContext(AuthContext);
  if (!ctx) {
    throw new Error('useAuth must be used within AuthProvider');
  }
  return ctx;
};
