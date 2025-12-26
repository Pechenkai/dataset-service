import { useEffect, useMemo, useRef, useState } from 'react';
import { Dataset } from '../api/types';
import { DatasetStats, computeDatasetStats } from '../workers/datasetWorker';

export const useDatasetWorker = (datasets: Dataset[]) => {
  const [stats, setStats] = useState<DatasetStats | null>(null);
  const workerRef = useRef<Worker | null>(null);

  useEffect(() => {
    if (typeof Worker === 'undefined') return;
    try {
      workerRef.current = new Worker(new URL('../workers/datasetWorker.ts', import.meta.url), { type: 'module' });
    } catch {
      workerRef.current = null;
    }
    return () => {
      workerRef.current?.terminate();
      workerRef.current = null;
    };
  }, []);

  const payload = useMemo(
    () =>
      datasets.map((d) => ({
        is_public: d.is_public,
        rating_summary: d.rating_summary,
        metadata: d.metadata,
        latest_version: d.latest_version
      })),
    [datasets]
  );

  useEffect(() => {
    if (!datasets.length) {
      setStats(null);
      return;
    }

    if (!workerRef.current) {
      setStats(computeDatasetStats(payload as any));
      return;
    }

    const handleMessage = (event: MessageEvent) => {
      if (event.data?.type === 'stats') {
        setStats(event.data.stats as DatasetStats);
      }
    };

    workerRef.current.addEventListener('message', handleMessage);
    workerRef.current.postMessage({ type: 'computeStats', datasets: payload });
    return () => workerRef.current?.removeEventListener('message', handleMessage);
  }, [datasets, payload]);

  return stats;
};
