import { ApiClient } from '../api/client';
import { PaginatedResponse, User } from '../api/types';

export class UserService {
  constructor(private readonly api: ApiClient) {}

  async getUser(id: number) {
    return this.api.get<User>(`/users/${id}`);
  }

  async listUsers() {
    return this.api.get<PaginatedResponse<User>>('/users');
  }

  async updateUser(id: number, payload: Partial<Pick<User, 'is_blocked' | 'role' | 'country' | 'username' | 'email'>>) {
    return this.api.patch<User>(`/users/${id}`, payload);
  }

  async deleteUser(id: number) {
    return this.api.delete<void>(`/users/${id}`);
  }
}

export const createUserService = (api: ApiClient) => new UserService(api);
