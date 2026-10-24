'use client';

import { useCallback, useEffect, useSyncExternalStore } from 'react';
import { DEFAULT_LOCALE, Locale, LOCALE_KEY, makeTranslate, resolveLocale, Translate } from './index';

const listeners = new Set<() => void>();

function subscribe(onChange: () => void) {
  listeners.add(onChange);
  window.addEventListener('storage', onChange); // another tab changed the language
  return () => {
    listeners.delete(onChange);
    window.removeEventListener('storage', onChange);
  };
}

function getLocale(): Locale {
  let stored: string | null = null;
  try {
    stored = localStorage.getItem(LOCALE_KEY);
  } catch {
    // storage blocked: fall back to the browser language
  }
  return resolveLocale(stored, navigator.languages ?? [navigator.language]);
}

const getServerLocale = (): Locale => DEFAULT_LOCALE;

export function setLocale(locale: Locale) {
  try {
    localStorage.setItem(LOCALE_KEY, locale);
  } catch {
    // storage blocked: the choice just won't persist
  }
  listeners.forEach((l) => l());
}

export function useLocale() {
  const locale = useSyncExternalStore(subscribe, getLocale, getServerLocale);
  return { locale, setLocale };
}

/** Translation function bound to the active language; re-renders when it changes. */
export function useT(): { t: Translate; locale: Locale } {
  const { locale } = useLocale();
  const t = useCallback<Translate>((key, params) => makeTranslate(locale)(key, params), [locale]);
  return { t, locale };
}

/** Keeps <html lang> in step with the active language (screen readers, hyphenation). */
export function useHtmlLang() {
  const { locale } = useLocale();
  useEffect(() => {
    document.documentElement.lang = locale;
  }, [locale]);
}
