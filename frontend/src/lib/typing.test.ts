import { describe, expect, it } from 'vitest';
import { makeTranslate } from '@/lib/i18n';
import { typingLabel } from './typing';

describe('typingLabel', () => {
  it('describes zero, one, two and many typists', () => {
    const t = makeTranslate('en');
    expect(typingLabel([], t)).toBe('');
    expect(typingLabel(['Ana'], t)).toBe('Ana is typing…');
    expect(typingLabel(['Ana', 'Luis'], t)).toBe('Ana and Luis are typing…');
    expect(typingLabel(['Ana', 'Luis', 'Eva'], t)).toBe('Several people are typing…');
  });

  it('speaks Spanish too', () => {
    const t = makeTranslate('es');
    expect(typingLabel(['Ana'], t)).toBe('Ana está escribiendo…');
    expect(typingLabel(['Ana', 'Luis', 'Eva'], t)).toBe('Varias personas están escribiendo…');
  });
});
