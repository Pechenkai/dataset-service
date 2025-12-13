import { ApiClient } from '../api/client';
import { PaginatedResponse, Subscription } from '../api/types';

export class SubscriptionService {
  constructor(private readonly api: ApiClient) {}

  async listByUser(userId?: number, datasetId?: number, page = 1, limit = 10) {
    const offset = (page - 1) * limit;
    return this.api.get<PaginatedResponse<Subscription>>('/subscriptions', {
      user_id: userId,
      dataset_id: datasetId,
      offset,
      limit
    });
  }

  async subscribe(datasetId: number, userId?: number) {
    return this.api.post<Subscription>('/subscriptions', { dataset_id: datasetId, user_id: userId });
  }

  async unsubscribe(subscriptionId: number) {
    return this.api.delete<void>(`/subscriptions/${subscriptionId}`);
  }
}
