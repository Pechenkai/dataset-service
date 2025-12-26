import { useEffect, useMemo, useRef, useState } from 'react';
import type { FileWorkerRequest, FileWorkerResponse } from '../workers/fileWorker';

export type FileAnalysis = {
  size?: number;
  preview?: string;
  hash?: string;
  error?: string;
  loading: boolean;
};

export function useFileWorker(file: File | null) {
  const workerRef = useRef<Worker>();
  const [result, setResult] = useState<FileAnalysis>({ loading: false });

  useEffect(() => {
    workerRef.current = new Worker(new URL('../workers/fileWorker.ts', import.meta.url), { type: 'module' });
    return () => {
      workerRef.current?.terminate();
    };
  }, []);

  useEffect(() => {
    if (!workerRef.current || !file) return;

    setResult({ loading: true });
    const worker = workerRef.current;
    const onMessage = (e: MessageEvent<FileWorkerResponse>) => {
      if (e.data.type === 'inspect') {
        setResult((prev) => ({ ...prev, loading: false, size: e.data.size, preview: e.data.preview }));
      }
      if (e.data.type === 'hash') {
        setResult((prev) => ({ ...prev, hash: e.data.hash }));
      }
      if (e.data.type === 'error') {
        setResult({ loading: false, error: e.data.message });
      }
    };
    worker.addEventListener('message', onMessage);
    worker.postMessage({ type: 'inspect', file } as FileWorkerRequest);
    worker.postMessage({ type: 'hash', file } as FileWorkerRequest);

    return () => {
      worker.removeEventListener('message', onMessage);
    };
  }, [file]);

  return result;
}
