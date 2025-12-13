import { ApiClient } from '../api/client';
import { NotificationItem, PaginatedResponse } from '../api/types';

export class NotificationService {
  constructor(private readonly api: ApiClient) {}

  async list(page = 1, limit = 10) {
    const offset = (page - 1) * limit;
    return this.api.get<PaginatedResponse<NotificationItem>>('/notifications', { limit, offset });
  }

  async markRead(id: number, isRead: boolean) {
    return this.api.patch<NotificationItem>(`/notifications/${id}`, { is_read: isRead });
  }
}
