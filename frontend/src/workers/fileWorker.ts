// Simple worker that reads the first bytes of a file and computes basic stats

export type FileWorkerRequest =
  | { type: 'inspect'; file: File }
  | { type: 'hash'; file: File };

export type FileWorkerResponse =
  | { type: 'inspect'; size: number; preview: string }
  | { type: 'hash'; hash: string }
  | { type: 'error'; message: string };

const toHex = (buffer: ArrayBuffer) =>
  [...new Uint8Array(buffer)].map((b) => b.toString(16).padStart(2, '0')).join('');

self.onmessage = async (evt: MessageEvent<FileWorkerRequest>) => {
  const data = evt.data;
  try {
    if (data.type === 'inspect') {
      const size = data.file.size;
      const slice = data.file.slice(0, 1024);
      const text = await slice.text();
      self.postMessage({ type: 'inspect', size, preview: text } satisfies FileWorkerResponse);
      return;
    }
    if (data.type === 'hash') {
      const buf = await data.file.arrayBuffer();
      const digest = await crypto.subtle.digest('SHA-256', buf);
      self.postMessage({ type: 'hash', hash: toHex(digest) } satisfies FileWorkerResponse);
      return;
    }
    self.postMessage({ type: 'error', message: 'unknown message' } satisfies FileWorkerResponse);
  } catch (e: any) {
    self.postMessage({ type: 'error', message: e?.message || 'worker failed' } satisfies FileWorkerResponse);
  }
};
