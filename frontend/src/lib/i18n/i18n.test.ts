import { describe, expect, it } from 'vitest';
import en from './en';
import es from './es';
import { resolveLocale, translate } from './index';

describe('resolveLocale', () => {
  it('prefers the stored choice', () => {
    expect(resolveLocale('es', ['en-US'])).toBe('es');
    expect(resolveLocale('en', ['es-MX'])).toBe('en');
  });

  it('falls back to the first supported browser language', () => {
    expect(resolveLocale(null, ['fr-FR', 'es-MX', 'en'])).toBe('es');
    expect(resolveLocale('xx', ['EN-gb'])).toBe('en');
  });

  it('defaults to English', () => {
    expect(resolveLocale(null, ['fr', 'de'])).toBe('en');
    expect(resolveLocale(null, [])).toBe('en');
  });
});

describe('translate', () => {
  it('interpolates params and leaves unknown placeholders intact', () => {
    expect(translate('en', 'chat.leaveConfirm', { room: 'Games' })).toBe('Leave Games?');
    expect(translate('es', 'chat.leaveConfirm', { room: 'Juegos' })).toBe('¿Salir de Juegos?');
    expect(translate('en', 'chat.leaveConfirm')).toBe('Leave {room}?');
  });
});

describe('dictionaries', () => {
  it('use the same placeholders in every language', () => {
    const names = (s: string) => (s.match(/\{\w+\}/g) ?? []).sort().join();
    for (const key of Object.keys(en) as (keyof typeof en)[]) {
      expect(names(es[key]), key).toBe(names(en[key]));
    }
  });
});
