import { describe, expect, it } from 'vitest';
import { isValidBanDuration } from './banDuration';

describe('isValidBanDuration', () => {
  it('accepts empty (permanent) and Go durations', () => {
    for (const v of ['', '  ', '24h', '168h', '90m', '1h30m', '1.5h']) expect(isValidBanDuration(v), v).toBe(true);
  });

  it('rejects other text, missing units and zero', () => {
    for (const v of ['24', 'forever', '1d', 'h', '-5h', '0h', '0s']) expect(isValidBanDuration(v), v).toBe(false);
  });
});
