import { describe, expect, it } from 'vitest';
import { resolveTheme, themeInitScript } from './theme';

describe('resolveTheme', () => {
  it('prefers the stored choice over the system', () => {
    expect(resolveTheme('light', true)).toBe('light');
    expect(resolveTheme('dark', false)).toBe('dark');
  });

  it('follows the system when nothing valid is stored', () => {
    expect(resolveTheme(null, true)).toBe('dark');
    expect(resolveTheme(null, false)).toBe('light');
    expect(resolveTheme('purple', true)).toBe('dark');
  });
});

describe('themeInitScript locale', () => {
  const run = (stored: string | null, languages: string[]) => {
    localStorage.clear();
    if (stored) localStorage.setItem('locale', stored);
    Object.defineProperty(navigator, 'languages', { value: languages, configurable: true });
    document.documentElement.lang = '';
    new Function(themeInitScript)();
    return document.documentElement.lang;
  };

  it('sets <html lang> like resolveLocale does', () => {
    expect(run('es', ['en-US'])).toBe('es');
    expect(run(null, ['fr-FR', 'es-MX'])).toBe('es');
    expect(run('xx', ['de'])).toBe('en');
    expect(run(null, [])).toBe('en');
  });
});
