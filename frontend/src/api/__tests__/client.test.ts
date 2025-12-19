import { ApiClient, ApiError } from '../client';

describe('ApiClient', () => {
  const originalFetch = global.fetch;

  afterEach(() => {
    global.fetch = originalFetch as any;
  });

  it('builds url with query params and adds auth header', async () => {
    const fetchMock = vi.fn().mockResolvedValue({
      ok: true,
      status: 200,
      headers: new Headers({ 'Content-Type': 'application/json' }),
      json: async () => ({ ok: true }),
      text: async () => ''
    });
    global.fetch = fetchMock as any;

    const client = new ApiClient('/api', () => 'token-123');
    await client.get('/datasets', { search: 'cat', limit: 10 });

    expect(fetchMock).toHaveBeenCalled();
    const calledUrl = (fetchMock.mock.calls[0] as any)[0] as string;
    expect(calledUrl).toContain('search=cat');
    expect(fetchMock.mock.calls[0][1]?.headers.get('Authorization')).toBe('Bearer token-123');
  });

  it('throws ApiError on non-ok responses', async () => {
    const response = new Response(JSON.stringify({ message: 'boom' }), {
      status: 500,
      headers: { 'Content-Type': 'application/json' }
    });
    const fetchMock = vi.fn().mockResolvedValue(response);
    global.fetch = fetchMock as any;

    const client = new ApiClient('/api');
    await expect(client.get('/broken')).rejects.toBeInstanceOf(ApiError);
  });
});
