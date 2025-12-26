import { renderHook, act } from '@testing-library/react';
import { describe, expect, it, beforeEach } from 'vitest';
import { usePersistentState } from '../usePersistentState';

describe('usePersistentState', () => {
  beforeEach(() => {
    localStorage.clear();
  });

  it('hydrates from storage and persists updates', async () => {
    localStorage.setItem('prefs', JSON.stringify({ foo: 'bar' }));
    const { result } = renderHook(() => usePersistentState('prefs', { foo: 'baz' }));

    expect(result.current[0]).toEqual({ foo: 'bar' });

    await act(async () => {
      result.current[1]({ foo: 'next' });
    });

    const raw = localStorage.getItem('prefs');
    expect(raw).toBeTruthy();
    expect(JSON.parse(raw || '{}')).toEqual({ foo: 'next' });
  });
});
