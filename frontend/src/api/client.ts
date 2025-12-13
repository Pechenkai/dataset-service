import { ErrorResponse } from './types';

type HttpMethod = 'GET' | 'POST' | 'PATCH' | 'DELETE';

export interface RequestOptions {
  method?: HttpMethod;
  query?: Record<string, string | number | boolean | undefined | null>;
  body?: unknown;
  headers?: Record<string, string>;
  skipAuth?: boolean;
}

export class ApiError extends Error {
  status: number;
  body?: unknown;

  constructor(message: string, status: number, body?: unknown) {
    super(message);
    this.status = status;
    this.body = body;
  }
}

const ensureOrigin = () => {
  if (typeof window !== 'undefined' && window.location) {
    return window.location.origin;
  }
  return 'http://localhost';
};

export class ApiClient {
  private readonly baseUrl: string;
  private readonly tokenProvider?: () => string | null | undefined;

  constructor(basePath = '/api/v2', tokenProvider?: () => string | null | undefined) {
    this.baseUrl = basePath.startsWith('http')
      ? basePath.replace(/\/$/, '')
      : `${ensureOrigin()}${basePath.replace(/\/$/, '')}`;
    this.tokenProvider = tokenProvider;
  }

  private buildUrl(path: string, query?: RequestOptions['query']): string {
    const url = path.startsWith('http') ? new URL(path) : new URL(`${this.baseUrl}${path}`);
    if (query) {
      Object.entries(query)
        .filter(([, value]) => value !== undefined && value !== null && value !== '')
        .forEach(([key, value]) => url.searchParams.set(key, String(value)));
    }
    return url.toString();
  }

  private buildHeaders(body: unknown, skipAuth?: boolean, extra?: Record<string, string>) {
    const headers = new Headers();
    const token = this.tokenProvider?.();
    if (!skipAuth && token) {
      headers.set('Authorization', `Bearer ${token}`);
    }
    const isFormData = typeof FormData !== 'undefined' && body instanceof FormData;
    if (!isFormData && body !== undefined && body !== null) {
      headers.set('Content-Type', 'application/json');
    }
    Object.entries(extra ?? {}).forEach(([key, value]) => headers.set(key, value));
    return headers;
  }

  async request<T>(path: string, options: RequestOptions = {}): Promise<T> {
    const { method = 'GET', body, query, headers, skipAuth } = options;
    const url = this.buildUrl(path, query);
    const isFormData = typeof FormData !== 'undefined' && body instanceof FormData;
    const payload = !isFormData && body !== undefined && body !== null ? JSON.stringify(body) : (body as BodyInit | null);

    const response = await fetch(url, {
      method,
      headers: this.buildHeaders(body, skipAuth, headers),
      body: method === 'GET' || method === 'DELETE' ? undefined : payload
    });

    const contentType = response.headers.get('Content-Type') || '';
    const isJSON = contentType.includes('application/json');
    const parsed = isJSON ? await response.json() : await response.text();

    if (!response.ok) {
      const errorPayload = isJSON ? (parsed as ErrorResponse) : { message: parsed };
      throw new ApiError(errorPayload.message || 'Request failed', response.status, errorPayload);
    }

    return parsed as T;
  }

  get<T>(path: string, query?: RequestOptions['query']) {
    return this.request<T>(path, { method: 'GET', query });
  }

  post<T>(path: string, body?: unknown, query?: RequestOptions['query'], skipAuth?: boolean) {
    return this.request<T>(path, { method: 'POST', body, query, skipAuth });
  }

  patch<T>(path: string, body?: unknown, query?: RequestOptions['query']) {
    return this.request<T>(path, { method: 'PATCH', body, query });
  }

  delete<T>(path: string, query?: RequestOptions['query']) {
    return this.request<T>(path, { method: 'DELETE', query });
  }
}

export const createApiClient = (tokenProvider?: () => string | null | undefined) =>
  new ApiClient('/api/v2', tokenProvider);
