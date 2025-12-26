export const formatSize = (raw?: number | string | null) => {
  if (raw === undefined || raw === null) return null;
  const size = typeof raw === 'string' ? Number(raw) : raw;
  if (!Number.isFinite(size) || size <= 0) return null;
  if (size > 1024 * 1024) return `${(size / (1024 * 1024)).toFixed(1)} MB`;
  if (size > 1024) return `${(size / 1024).toFixed(1)} KB`;
  return `${size} B`;
};
