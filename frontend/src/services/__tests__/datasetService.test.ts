import { describe, expect, it, vi } from 'vitest';
import { ApiClient } from '../../api/client';
import { DatasetService } from '../datasetService';

describe('DatasetService', () => {
  const createService = () => {
    const api = {
      get: vi.fn().mockResolvedValue({ items: [], meta: { total: 0, limit: 12, offset: 0 } }),
      post: vi.fn()
    } as unknown as ApiClient;
    return { service: new DatasetService(api), api };
  };

  it('maps filters to query params', async () => {
    const { service, api } = createService();
    await service.listDatasets({ search: 'traffic', visibility: 'public', page: 2, limit: 20 });

    expect(api.get).toHaveBeenCalledWith(
      '/datasets',
      expect.objectContaining({
        search: 'traffic',
        is_public: true,
        offset: 20,
        limit: 20
      })
    );
  });

  it('parses and stringifies URL search params', () => {
    const { service } = createService();
    const params = new URLSearchParams('q=ml&category=3&visibility=private&page=3&limit=5');
    const filters = service.buildFiltersFromSearch(params);

    expect(filters.search).toBe('ml');
    expect(filters.categoryId).toBe(3);
    expect(filters.visibility).toBe('private');
    expect(filters.page).toBe(3);

    const next = service.stringifyFilters(filters);
    expect(next.get('q')).toBe('ml');
    expect(next.get('category')).toBe('3');
    expect(next.get('page')).toBe('3');
  });
});
