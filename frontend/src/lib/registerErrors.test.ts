import { describe, expect, it } from 'vitest';
import { registerErrorKey } from './registerErrors';

const res = (status: number, error?: string) => ({ response: { status, data: { error } } });

describe('registerErrorKey', () => {
  it('maps status codes', () => {
    expect(registerErrorKey(res(409, 'User with this email or username already exists'))).toBe('register.conflict');
    expect(registerErrorKey(res(429))).toBe('register.rateLimited');
  });

  it('maps the known password reasons on 400', () => {
    expect(registerErrorKey(res(400, 'password is too common, choose a less predictable one'))).toBe('reset.common');
    expect(registerErrorKey(res(400, 'password must not contain your username or email'))).toBe('reset.personal');
    expect(registerErrorKey(res(400, 'password must be at least 8 characters'))).toBe('validation.passwordMin8');
    expect(registerErrorKey(res(400, 'password must be at most 72 characters'))).toBe('validation.passwordMax72');
  });

  it('never surfaces unknown server text', () => {
    expect(registerErrorKey(res(400, 'Key: RegisterRequest.Email failed on the email tag'))).toBe('register.failed');
    expect(registerErrorKey(res(500, 'password is too common'))).toBe('register.failed');
    expect(registerErrorKey(new Error('Network Error'))).toBe('register.failed');
    expect(registerErrorKey(null)).toBe('register.failed');
  });
});
