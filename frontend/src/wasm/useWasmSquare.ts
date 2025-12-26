// Minimal WebAssembly module (compiled from: (module (func (export "square") (param i32) (result i32) local.get 0 local.get 0 i32.mul)))
const wasmBase64 = 'AGFzbQEAAAABBwFgAX8BfwMCAQAHCgEGc3F1YXJlAAAKCQEHIAAgAGwL';

const decodeWasm = (b64: string) => Uint8Array.from(atob(b64), (c) => c.charCodeAt(0));

let wasmInstance: WebAssembly.Instance | null = null;

async function loadWasm() {
  if (wasmInstance) return wasmInstance;
  const bytes = decodeWasm(wasmBase64);
  const { instance } = await WebAssembly.instantiate(bytes.buffer);
  wasmInstance = instance;
  return instance;
}

export async function squareViaWasm(value: number): Promise<number> {
  const instance = await loadWasm();
  const fn = (instance.exports as any).square as (n: number) => number;
  return fn(Math.floor(value));
}
