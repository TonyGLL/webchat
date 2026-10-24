import type { MessageKey } from '@/lib/i18n/en';

/**
 * Maps the server's (English) reason for refusing a reset to a translated message key.
 * Anything unrecognised becomes the generic failure, so no server text is ever shown raw.
 */
export function resetErrorKey(error: unknown): MessageKey {
  const message = String((error as { response?: { data?: { error?: string } } } | null)?.response?.data?.error ?? '').toLowerCase();
  if (message.includes('invalid or has expired')) return 'reset.invalidLink';
  if (message.includes('too common')) return 'reset.common';
  if (message.includes('username or email')) return 'reset.personal';
  if (message.includes('at least')) return 'validation.passwordMin8';
  if (message.includes('at most')) return 'validation.passwordMax72';
  return 'reset.failed';
}
