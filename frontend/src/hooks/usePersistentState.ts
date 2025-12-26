import { useEffect, useState } from 'react';

const canUseStorage = () => typeof window !== 'undefined' && !!window.localStorage;

type Options<T> = {
  serializer?: (value: T) => string;
  deserializer?: (raw: string) => T;
};

export function usePersistentState<T>(
  key: string,
  initial: T,
  options: Options<T> = {}
): [T, React.Dispatch<React.SetStateAction<T>>, { reset: () => void }] {
  const { serializer = JSON.stringify, deserializer = (raw: string) => JSON.parse(raw) as T } = options;

  const [value, setValue] = useState<T>(() => {
    if (!canUseStorage()) return initial;
    try {
      const raw = window.localStorage.getItem(key);
      return raw ? deserializer(raw) : initial;
    } catch {
      return initial;
    }
  });

  useEffect(() => {
    if (!canUseStorage()) return;
    try {
      window.localStorage.setItem(key, serializer(value));
    } catch {
      // ignore quota errors
    }
  }, [key, value]);

  const reset = () => setValue(initial);

  return [value, setValue, { reset }];
}
