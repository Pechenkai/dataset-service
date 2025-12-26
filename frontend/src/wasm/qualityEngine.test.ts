import { describe, expect, it } from 'vitest';
import { WasmQualityEngine } from './qualityEngine';

describe('WasmQualityEngine', () => {
  it('computes stable score and initializes wasm', async () => {
    const engine = new WasmQualityEngine();
    const dataset: any = {
      id: 1,
      is_public: true,
      rating_summary: { average: 4.5, count: 20 },
      metadata: { size: 1024 }
    };

    const beforeInit = engine.score(dataset);
    await engine.init();
    const afterInit = engine.score(dataset);

    expect(beforeInit).toBeGreaterThan(0);
    expect(afterInit).toBeGreaterThan(0);
  });
});
