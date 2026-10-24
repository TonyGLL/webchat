import { describe, expect, it } from 'vitest';
import { clampStep, isOnboardingDone, markOnboardingDone, ONBOARDING_STEPS } from './onboarding';

const memoryStorage = () => {
  const data = new Map<string, string>();
  return { getItem: (k: string) => data.get(k) ?? null, setItem: (k: string, v: string) => void data.set(k, v) };
};

describe('onboarding', () => {
  it('starts as not done and remembers completion per user', () => {
    const storage = memoryStorage();
    expect(isOnboardingDone(storage, 1)).toBe(false);
    markOnboardingDone(storage, 1);
    expect(isOnboardingDone(storage, 1)).toBe(true);
    expect(isOnboardingDone(storage, 2)).toBe(false);
  });

  it('survives a storage that throws', () => {
    const broken = {
      getItem: () => {
        throw new Error('blocked');
      },
      setItem: () => {
        throw new Error('blocked');
      },
    };
    expect(isOnboardingDone(broken, 1)).toBe(false);
    expect(() => markOnboardingDone(broken, 1)).not.toThrow();
  });

  it('clamps step indexes into the tour', () => {
    expect(clampStep(-3)).toBe(0);
    expect(clampStep(1)).toBe(1);
    expect(clampStep(99)).toBe(ONBOARDING_STEPS.length - 1);
  });
});
