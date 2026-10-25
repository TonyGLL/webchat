import { readdirSync, readFileSync, statSync } from 'node:fs';
import { join, relative } from 'node:path';
import { describe, expect, it } from 'vitest';

// Every user string is rendered as React text. A raw HTML sink would be the first step to XSS,
// so any new one must be added here on purpose, after review.
const ALLOWED = new Set(['app/layout.tsx']); // the constant theme init script

const SRC = __dirname;

function sources(dir: string): string[] {
  return readdirSync(dir).flatMap((name) => {
    const path = join(dir, name);
    if (statSync(path).isDirectory()) return sources(path);
    return /\.(tsx?|jsx?)$/.test(name) && !/\.test\./.test(name) ? [path] : [];
  });
}

describe('raw HTML sinks', () => {
  it('only appear in the allowlisted files', () => {
    const offenders = sources(SRC)
      .filter((file) => /dangerouslySetInnerHTML|\.innerHTML\b|\.outerHTML\b|insertAdjacentHTML|document\.write\(/.test(readFileSync(file, 'utf8')))
      .map((file) => relative(SRC, file))
      .filter((file) => !ALLOWED.has(file));
    expect(offenders).toEqual([]);
  });
});
