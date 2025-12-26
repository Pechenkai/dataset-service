import { describe, expect, it } from 'vitest';
import { computeDatasetStats } from '../datasetWorker';

describe('datasetWorker.computeDatasetStats', () => {
  it('aggregates stats and tags', () => {
    const stats = computeDatasetStats([
      {
        is_public: true,
        rating_summary: { average: 4, count: 10 },
        metadata: { tags: ['cv', 'nlp'], size: 1024 },
        latest_version: undefined
      } as any,
      {
        is_public: false,
        rating_summary: { average: 5, count: 2 },
        metadata: { tags: ['cv'], size: 2048 },
        latest_version: undefined
      } as any
    ]);

    expect(stats.total).toBe(2);
    expect(stats.publicCount).toBe(1);
    expect(stats.privateCount).toBe(1);
    expect(stats.averageRating).toBeGreaterThan(0);
    expect(stats.topTags[0].tag).toBe('cv');
  });
});
