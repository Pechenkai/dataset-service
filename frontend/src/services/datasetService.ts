import { ApiClient } from '../api/client';
import {
  Category,
  Dataset,
  DatasetMetadata,
  DatasetVersion,
  PaginatedResponse,
  Review
} from '../api/types';

export interface DatasetFilters {
  search?: string;
  categoryId?: number;
  visibility?: 'public' | 'private' | 'all';
  tags?: string;
  ownerId?: number;
  page?: number;
  limit?: number;
}

export interface CreateDatasetPayload {
  name: string;
  description?: string;
  categoryId: number;
  isPublic?: boolean;
  file: File | Blob;
  metadata?: DatasetMetadata;
}

const DEFAULT_PAGE_SIZE = 12;

export class DatasetService {
  private readonly api: ApiClient;

  constructor(api: ApiClient) {
    this.api = api;
  }

  async listCategories(search?: string) {
    return this.api.get<PaginatedResponse<Category>>('/categories', {
      limit: 100,
      offset: 0,
      search
    });
  }

  async createCategory(payload: { name: string; description?: string }) {
    return this.api.post<Category>('/categories', {
      name: payload.name,
      description: payload.description
    });
  }

  async listDatasets(filters: DatasetFilters = {}) {
    const limit = filters.limit ?? DEFAULT_PAGE_SIZE;
    const page = filters.page && filters.page > 0 ? filters.page : 1;
    const offset = (page - 1) * limit;

    return this.api.get<PaginatedResponse<Dataset>>('/datasets', {
      limit,
      offset,
      search: filters.search,
      category_id: filters.categoryId,
      owner_id: filters.ownerId,
      is_public:
        filters.visibility === 'public'
          ? true
          : filters.visibility === 'private'
          ? false
          : undefined,
      tags: filters.tags
    });
  }

  async getDataset(datasetId: number) {
    return this.api.get<Dataset>(`/datasets/${datasetId}`);
  }

  async listVersions(datasetId: number, page = 1, limit = 5) {
    const offset = (page - 1) * limit;
    return this.api.get<PaginatedResponse<DatasetVersion>>(
      `/datasets/${datasetId}/versions`,
      {
        limit,
        offset
      }
    );
  }

  async createDataset(payload: CreateDatasetPayload) {
    const formData = new FormData();
    formData.append('name', payload.name);
    formData.append('category_id', String(payload.categoryId));
    formData.append('file', payload.file);

    if (payload.description) {
      formData.append('description', payload.description);
    }
    if (payload.metadata) {
      formData.append('metadata', JSON.stringify(payload.metadata));
    }
    if (payload.isPublic !== undefined) {
      formData.append('is_public', String(payload.isPublic));
    }

    return this.api.post<Dataset>('/datasets', formData);
  }

  async createVersion(datasetId: number, payload: { file: File | Blob; changeLog?: string; metadata?: DatasetMetadata }) {
    const formData = new FormData();
    formData.append('file', payload.file);
    if (payload.changeLog) {
      formData.append('change_log', payload.changeLog);
    }
    if (payload.metadata) {
      formData.append('metadata', JSON.stringify(payload.metadata));
    }

    return this.api.post<DatasetVersion>(`/datasets/${datasetId}/versions`, formData);
  }

  async listReviews(datasetId: number, page = 1, limit = 10) {
    const offset = (page - 1) * limit;
    return this.api.get<PaginatedResponse<Review>>(`/datasets/${datasetId}/reviews`, {
      offset,
      limit
    });
  }

  async createReview(datasetId: number, rating: number, text?: string) {
    // v2 API принимает POST /reviews с dataset_id
    return this.api.post<Review>(
      '/reviews',
      { dataset_id: datasetId, rating, text },
      undefined,
      false
    );
  }

  async getLatestVersion(datasetId: number) {
    const res = await this.listVersions(datasetId, 1, 1);
    return res.items?.[0] ?? null;
  }


  async updateDataset(datasetId: number, payload: Partial<Pick<Dataset, 'is_public' | 'name' | 'description'>>) {
    return this.api.patch<Dataset>(`/datasets/${datasetId}`, payload);
  }

  buildFiltersFromSearch(params: URLSearchParams): DatasetFilters {
    return {
      search: params.get('q') || undefined,
      categoryId: params.get('category') ? Number(params.get('category')) : undefined,
      visibility: (params.get('visibility') as DatasetFilters['visibility']) || 'all',
      tags: params.get('tags') || undefined,
      ownerId: params.get('owner') ? Number(params.get('owner')) : undefined,
      page: params.get('page') ? Number(params.get('page')) : 1,
      limit: params.get('limit') ? Number(params.get('limit')) : DEFAULT_PAGE_SIZE
    };
  }

  stringifyFilters(filters: DatasetFilters) {
    const params = new URLSearchParams();
    if (filters.search) params.set('q', filters.search);
    if (filters.categoryId) params.set('category', String(filters.categoryId));
    if (filters.visibility && filters.visibility !== 'all') params.set('visibility', filters.visibility);
    if (filters.tags) params.set('tags', filters.tags);
    if (filters.ownerId) params.set('owner', String(filters.ownerId));
    if (filters.page && filters.page > 1) params.set('page', String(filters.page));
    if (filters.limit && filters.limit !== DEFAULT_PAGE_SIZE) params.set('limit', String(filters.limit));
    return params;
  }
}
