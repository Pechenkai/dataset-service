import { ApiClient } from '../api/client';
import { AuthenticateRequest, AuthenticateResponse, User } from '../api/types';

export interface AuthSession {
  token: string;
  user: User;
  expires_at?: string;
}

export const AUTH_STORAGE_KEY = 'dataset-platform.auth';

export class AuthService {
  constructor(private readonly api: ApiClient) {}

  async authenticate(payload: AuthenticateRequest) {
    return this.api.post<AuthenticateResponse>('/auth/tokens', payload, undefined, true);
  }

  async revoke(tokenId?: string) {
    return this.api.post<void>('/auth/tokens/revoke', tokenId ? { token_id: tokenId } : undefined);
  }

  loadSession(): AuthSession | null {
    try {
      const raw = localStorage.getItem(AUTH_STORAGE_KEY);
      if (!raw) return null;
      const parsed = JSON.parse(raw) as AuthSession;
      if (!parsed.token || !parsed.user) return null;
      return parsed;
    } catch (e) {
      console.warn('Failed to parse auth session', e);
      return null;
    }
  }

  persistSession(session: AuthSession | null) {
    if (!session) {
      localStorage.removeItem(AUTH_STORAGE_KEY);
      return;
    }
    localStorage.setItem(AUTH_STORAGE_KEY, JSON.stringify(session));
  }
}

export const createEmptySession = (): AuthSession | null => null;
