import { describe, expect, it, vi, beforeEach } from 'vitest';
import { ApiClient } from '../../api/client';
import { AUTH_STORAGE_KEY, AuthService } from '../authService';

const createService = () => {
  const api = {
    post: vi.fn()
  } as unknown as ApiClient;
  return new AuthService(api);
};

describe('AuthService', () => {
  beforeEach(() => {
    localStorage.clear();
  });

  it('persists and restores session', () => {
    const service = createService();
    service.persistSession({ token: '123', user: { id: 1, username: 'demo', email: 'd@x.y', role: 'user', is_blocked: false }, expires_at: '2024-10-10' });

    const raw = localStorage.getItem(AUTH_STORAGE_KEY);
    expect(raw).toBeTruthy();

    const restored = service.loadSession();
    expect(restored?.token).toBe('123');
    expect(restored?.user.username).toBe('demo');
  });

  it('calls auth endpoint', async () => {
    const api = {
      post: vi.fn().mockResolvedValue({ token: 'jwt', user: { id: 1, username: 'test', email: 't@t', role: 'user', is_blocked: false } })
    } as unknown as ApiClient;
    const service = new AuthService(api);

    const response = await service.authenticate({ email: 't@t', password: 'pwd' });
    expect(api.post).toHaveBeenCalledWith('/auth/tokens', { email: 't@t', password: 'pwd' }, undefined, true);
    expect(response.token).toBe('jwt');
  });
});
