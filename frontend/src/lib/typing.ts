import type { Translate } from '@/lib/i18n';

/** Text for the "who is typing" line; empty when nobody is. */
export function typingLabel(names: string[], t: Translate): string {
  if (names.length === 0) return '';
  if (names.length === 1) return t('typing.one', { a: names[0] });
  if (names.length === 2) return t('typing.two', { a: names[0], b: names[1] });
  return t('typing.many');
}
