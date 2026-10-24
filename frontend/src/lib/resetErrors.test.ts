import { describe, expect, it } from 'vitest';
import { resetErrorKey } from './resetErrors';

const fromServer = (error: string) => ({ response: { data: { error } } });

describe('resetErrorKey', () => {
  it('recognises each reason the server gives', () => {
    expect(resetErrorKey(fromServer('This link is invalid or has expired'))).toBe('reset.invalidLink');
    expect(resetErrorKey(fromServer('password is too common, choose a less predictable one'))).toBe('reset.common');
    expect(resetErrorKey(fromServer('password must not contain your username or email'))).toBe('reset.personal');
    expect(resetErrorKey(fromServer('password must be at least 8 characters'))).toBe('validation.passwordMin8');
    expect(resetErrorKey(fromServer('password must be at most 72 characters'))).toBe('validation.passwordMax72');
  });

  it('falls back to a generic failure for anything else', () => {
    expect(resetErrorKey(fromServer('something new'))).toBe('reset.failed');
    expect(resetErrorKey(new Error('Network Error'))).toBe('reset.failed');
    expect(resetErrorKey(null)).toBe('reset.failed');
  });
});
