import { Dataset } from '../api/types';
import { extractDatasetSize, extractTags } from '../utils/dataset';

export type DatasetStats = {
  total: number;
  publicCount: number;
  privateCount: number;
  averageRating: number;
  avgSize: number;
  topTags: { tag: string; count: number }[];
};

export const computeDatasetStats = (datasets: Array<Pick<Dataset, 'is_public' | 'rating_summary' | 'metadata' | 'latest_version'>>) => {
  let ratingSum = 0;
  let ratingCount = 0;
  let sizeSum = 0;
  let sizeCount = 0;
  let publicCount = 0;
  const tagCounter = new Map<string, number>();

  datasets.forEach((d) => {
    if (d.is_public) publicCount += 1;
    const rating = d.rating_summary;
    if (rating) {
      ratingSum += rating.average * rating.count;
      ratingCount += rating.count;
    }
    const size = extractDatasetSize(d as Dataset);
    if (size) {
      sizeSum += size;
      sizeCount += 1;
    }
    const tags = extractTags(d as Dataset);
    tags.forEach((tag) => tagCounter.set(tag, (tagCounter.get(tag) ?? 0) + 1));
  });

  const topTags = Array.from(tagCounter.entries())
    .sort((a, b) => b[1] - a[1])
    .slice(0, 5)
    .map(([tag, count]) => ({ tag, count }));

  return {
    total: datasets.length,
    publicCount,
    privateCount: datasets.length - publicCount,
    averageRating: ratingCount > 0 ? Number((ratingSum / ratingCount).toFixed(2)) : 0,
    avgSize: sizeCount > 0 ? Math.round(sizeSum / sizeCount) : 0,
    topTags
  } satisfies DatasetStats;
};

type WorkerMessage =
  | { type: 'computeStats'; datasets: Array<Pick<Dataset, 'is_public' | 'rating_summary' | 'metadata' | 'latest_version'>> };

type WorkerResponse = { type: 'stats'; stats: DatasetStats };

const ctx: DedicatedWorkerGlobalScope = self as unknown as DedicatedWorkerGlobalScope;

ctx.onmessage = (event: MessageEvent<WorkerMessage>) => {
  const message = event.data;
  if (!message || message.type !== 'computeStats') return;
  const stats = computeDatasetStats(message.datasets);
  ctx.postMessage({ type: 'stats', stats } as WorkerResponse);
};
