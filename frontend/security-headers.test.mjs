import { describe, expect, it } from 'vitest';
import { apiConnectSources, assertProductionApiUrl, buildCsp, securityHeaders } from './security-headers.mjs';

describe('apiConnectSources', () => {
  it('derives the API origin and its ws counterpart', () => {
    expect(apiConnectSources('http://localhost:8081/api/v1')).toEqual(['http://localhost:8081', 'ws://localhost:8081']);
    expect(apiConnectSources('https://api.example.com/api/v1')).toEqual(['https://api.example.com', 'wss://api.example.com']);
  });

  it('falls back to the local default on a missing or invalid value', () => {
    const fallback = ['http://localhost:8081', 'ws://localhost:8081'];
    expect(apiConnectSources(undefined)).toEqual(fallback);
    expect(apiConnectSources('not a url')).toEqual(fallback);
    expect(apiConnectSources('javascript:alert(1)')).toEqual(fallback);
  });
});

describe('buildCsp', () => {
  it('only allows unsafe-eval outside production', () => {
    expect(buildCsp({ isProduction: true })).not.toContain('unsafe-eval');
    expect(buildCsp({ isProduction: false })).toContain("'unsafe-eval'");
  });

  it('locks framing, base and form targets', () => {
    const csp = buildCsp({ isProduction: true });
    expect(csp).toContain("frame-ancestors 'none'");
    expect(csp).toContain("object-src 'none'");
    expect(csp).toContain("base-uri 'self'");
    expect(csp).toContain("form-action 'self'");
    expect(csp).toContain("connect-src 'self' http://localhost:8081 ws://localhost:8081");
  });
});

describe('securityHeaders', () => {
  const names = (isProduction) => securityHeaders({ isProduction }).map((h) => h.key);

  it('sends HSTS only in production', () => {
    expect(names(true)).toContain('Strict-Transport-Security');
    expect(names(false)).not.toContain('Strict-Transport-Security');
  });

  it('includes the baseline headers', () => {
    expect(names(false)).toEqual(
      expect.arrayContaining(['Content-Security-Policy', 'X-Frame-Options', 'X-Content-Type-Options', 'Referrer-Policy', 'Permissions-Policy', 'Cross-Origin-Opener-Policy'])
    );
  });
});

describe('assertProductionApiUrl', () => {
  it('accepts https, localhost over http and an unset value', () => {
    expect(() => assertProductionApiUrl(undefined)).not.toThrow();
    expect(() => assertProductionApiUrl('https://api.example.com/api/v1')).not.toThrow();
    expect(() => assertProductionApiUrl('http://localhost:8081/api/v1')).not.toThrow();
    expect(() => assertProductionApiUrl('http://127.0.0.1:8081/api/v1')).not.toThrow();
  });

  it('rejects clear-text remote hosts, relative and invalid values', () => {
    expect(() => assertProductionApiUrl('http://api.example.com/api/v1')).toThrow(/https/);
    expect(() => assertProductionApiUrl('/api/v1')).toThrow(/valid/);
    expect(() => assertProductionApiUrl('ftp://localhost/x')).toThrow(/https/);
  });
});
