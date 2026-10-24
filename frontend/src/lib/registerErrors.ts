import type { MessageKey } from '@/lib/i18n/en';

/**
 * Maps a failed registration to a translated message key. Server text is never shown raw:
 * only the known password-policy reasons (400) are recognised, the rest is generic.
 */
export function registerErrorKey(error: unknown): MessageKey {
  const e = error as { response?: { status?: number; data?: { error?: string } } } | null;
  const status = e?.response?.status;
  if (status === 409) return 'register.conflict';
  if (status === 429) return 'register.rateLimited';
  if (status === 400) {
    const message = String(e?.response?.data?.error ?? '').toLowerCase();
    if (message.includes('too common')) return 'reset.common';
    if (message.includes('username or email')) return 'reset.personal';
    if (message.includes('at least')) return 'validation.passwordMin8';
    if (message.includes('at most')) return 'validation.passwordMax72';
  }
  return 'register.failed';
}
