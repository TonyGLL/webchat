import { describe, expect, it } from 'vitest';
import { setRegisteredFlash, takeRegisteredFlash } from './flash';

const memory = () => {
  const data = new Map<string, string>();
  return {
    getItem: (k: string) => data.get(k) ?? null,
    setItem: (k: string, v: string) => void data.set(k, v),
    removeItem: (k: string) => void data.delete(k),
  };
};

describe('registered flash', () => {
  it('is delivered once and then cleared', () => {
    const s = memory();
    expect(takeRegisteredFlash(s)).toBe(false);
    setRegisteredFlash(s);
    expect(takeRegisteredFlash(s)).toBe(true);
    expect(takeRegisteredFlash(s)).toBe(false);
  });

  it('survives blocked storage', () => {
    const fail = () => {
      throw new Error('blocked');
    };
    const blocked = { getItem: fail, setItem: fail, removeItem: fail };
    expect(() => setRegisteredFlash(blocked)).not.toThrow();
    expect(takeRegisteredFlash(blocked)).toBe(false);
  });
});
