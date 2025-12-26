import { Dataset } from '../api/types';
import { extractDatasetSize } from '../utils/dataset';

type MixExports = {
  mix?: (a: number, b: number) => number;
};

const wasmBinary = new Uint8Array([
  0x00, 0x61, 0x73, 0x6d, 0x01, 0x00, 0x00, 0x00, 0x01, 0x07, 0x01, 0x60, 0x02, 0x7f, 0x7f, 0x01, 0x7f, 0x03,
  0x02, 0x01, 0x00, 0x07, 0x07, 0x01, 0x03, 0x6d, 0x69, 0x78, 0x00, 0x00, 0x0a, 0x0f, 0x01, 0x0d, 0x00, 0x20, 0x00,
  0x41, 0x07, 0x6c, 0x20, 0x01, 0x41, 0x03, 0x6c, 0x6a, 0x0b
]);

const fallbackMix = (a: number, b: number) => a * 7 + b * 3;

export class WasmQualityEngine {
  private instance: MixExports | null = null;
  private initPromise: Promise<void> | null = null;
  private version = 0;

  async init() {
    if (this.instance) return;
    if (this.initPromise) return this.initPromise;
    if (typeof WebAssembly === 'undefined') {
      this.instance = null;
      return;
    }

    this.initPromise = WebAssembly.instantiate(wasmBinary.buffer, {}).then(({ instance }) => {
      this.instance = instance.exports as MixExports;
      this.version += 1;
    }).catch(() => {
      this.instance = null;
    }).finally(() => {
      this.initPromise = null;
    });

    return this.initPromise;
  }

  isReady() {
    return Boolean(this.instance);
  }

  getVersion() {
    return this.version;
  }

  score(dataset: Dataset) {
    const rating = dataset.rating_summary?.average ?? 0;
    const votes = dataset.rating_summary?.count ?? 0;
    const size = extractDatasetSize(dataset) ?? 0;
    const penalty = size > 0 ? Math.max(0, Math.round(Math.log10(size + 1))) : 0;
    const base = this.instance?.mix ? this.instance.mix(Math.round(rating * 100), votes) : fallbackMix(Math.round(rating * 100), votes);
    return Math.max(0, base - penalty);
  }
}

export const qualityEngine = new WasmQualityEngine();
