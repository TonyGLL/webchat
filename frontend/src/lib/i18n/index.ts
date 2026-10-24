import en, { MessageKey } from './en';
import es from './es';

export type { MessageKey };
export type Locale = 'en' | 'es';

export const LOCALES: Locale[] = ['en', 'es'];
export const LOCALE_KEY = 'locale';
export const DEFAULT_LOCALE: Locale = 'en';

const dictionaries: Record<Locale, Record<MessageKey, string>> = { en, es };

export type Params = Record<string, string | number>;
export type Translate = (key: MessageKey, params?: Params) => string;

const isLocale = (value: string | null | undefined): value is Locale => value === 'en' || value === 'es';

/** Stored choice wins; with none, follow the browser language, falling back to English. */
export function resolveLocale(stored: string | null, browserLanguages: readonly string[]): Locale {
  if (isLocale(stored)) return stored;
  for (const lang of browserLanguages) {
    const base = lang.toLowerCase().split('-')[0];
    if (isLocale(base)) return base;
  }
  return DEFAULT_LOCALE;
}

export function translate(locale: Locale, key: MessageKey, params?: Params): string {
  const text = dictionaries[locale][key] ?? en[key];
  if (!params) return text;
  return text.replace(/\{(\w+)\}/g, (match, name) => (name in params ? String(params[name]) : match));
}

export const makeTranslate = (locale: Locale): Translate => (key, params) => translate(locale, key, params);

/** BCP 47 tag for Intl / toLocale* formatting. */
export const intlLocale = (locale: Locale) => (locale === 'es' ? 'es-ES' : 'en-US');
