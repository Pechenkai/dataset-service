const MEMORY_FALLBACK = new Map<string, string>();

const hasLocalStorage = () => {
  try {
    const key = '__tg_storage_probe__';
    localStorage.setItem(key, '1');
    localStorage.removeItem(key);
    return true;
  } catch {
    return false;
  }
};

const openIndexedDb = (): Promise<IDBDatabase | null> => {
  return new Promise((resolve) => {
    if (typeof indexedDB === 'undefined') return resolve(null);
    const request = indexedDB.open('dataset-hub-cache', 1);
    request.onupgradeneeded = () => {
      request.result.createObjectStore('kv');
    };
    request.onsuccess = () => resolve(request.result);
    request.onerror = () => resolve(null);
  });
};

const idbGet = async (key: string) => {
  const db = await openIndexedDb();
  if (!db) return null;
  return new Promise<string | null>((resolve) => {
    const tx = db.transaction('kv', 'readonly');
    const store = tx.objectStore('kv');
    const req = store.get(key);
    req.onsuccess = () => resolve((req.result as string) ?? null);
    req.onerror = () => resolve(null);
  });
};

const idbSet = async (key: string, value: string | null) => {
  const db = await openIndexedDb();
  if (!db) return;
  return new Promise<void>((resolve) => {
    const tx = db.transaction('kv', 'readwrite');
    const store = tx.objectStore('kv');
    const req = value === null ? store.delete(key) : store.put(value, key);
    req.onsuccess = () => resolve();
    req.onerror = () => resolve();
  });
};

export const persistentStore = {
  getSync(key: string): string | null {
    if (!hasLocalStorage()) {
      return MEMORY_FALLBACK.get(key) ?? null;
    }
    try {
      return localStorage.getItem(key);
    } catch {
      return MEMORY_FALLBACK.get(key) ?? null;
    }
  },
  async get(key: string): Promise<string | null> {
    const local = this.getSync(key);
    if (local !== null) return local;
    try {
      return await idbGet(key);
    } catch {
      return null;
    }
  },
  async set(key: string, value: string | null) {
    if (!hasLocalStorage()) {
      if (value === null) MEMORY_FALLBACK.delete(key);
      else MEMORY_FALLBACK.set(key, value);
      await idbSet(key, value);
      return;
    }

    try {
      if (value === null) localStorage.removeItem(key);
      else localStorage.setItem(key, value);
    } catch {
      if (value === null) MEMORY_FALLBACK.delete(key);
      else MEMORY_FALLBACK.set(key, value);
    }

    await idbSet(key, value);
  }
};
