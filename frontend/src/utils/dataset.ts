import { Dataset } from '../api/types';

export const extractDatasetSize = (d: Dataset): number | null => {
  const candidates = [
    d.metadata?.size,
    (d.metadata as any)?.size_bytes,
    (d.metadata as any)?.bytes,
    (d as any).size,
    (d as any).size_bytes,
    (d as any).file_size,
    d.latest_version?.metadata?.size,
    (d.latest_version?.metadata as any)?.size_bytes,
    (d.latest_version?.metadata as any)?.bytes
  ];

  for (const value of candidates) {
    const num = Number(value);
    if (Number.isFinite(num) && num > 0) return num;
  }

  return null;
};

export const extractTags = (d: Dataset): string[] => {
  const raw = (d.metadata as any)?.tags ?? (d as any).tags ?? d.latest_version?.metadata?.tags;
  if (!raw) return [];
  if (Array.isArray(raw)) return raw.filter(Boolean).map(String);
  return String(raw)
    .split(',')
    .map((tag) => tag.trim())
    .filter(Boolean);
};
