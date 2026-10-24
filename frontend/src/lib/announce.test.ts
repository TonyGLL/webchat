import { describe, expect, it } from 'vitest';
import { makeTranslate } from '@/lib/i18n';
import { announcementText, incomingSince } from './announce';

const msgs = (...ids: string[]) => ids.map((id) => ({ id }));

describe('incomingSince', () => {
  it('returns the messages in front of the previous newest', () => {
    expect(incomingSince(msgs('c', 'b', 'a'), 'b')).toEqual({ known: true, fresh: msgs('c') });
    expect(incomingSince(msgs('b', 'a'), 'b')).toEqual({ known: true, fresh: [] });
  });

  it('is unknown without a previous id or when it is gone (room switch, reload)', () => {
    expect(incomingSince(msgs('b', 'a'), null).known).toBe(false);
    expect(incomingSince(msgs('z', 'y'), 'b')).toEqual({ known: false, fresh: [] });
  });
});

describe('announcementText', () => {
  it('names the author of a single message and counts several', () => {
    const t = makeTranslate('en');
    expect(announcementText([], t)).toBe('');
    expect(announcementText(['Ana'], t)).toBe('New message from Ana');
    expect(announcementText(['Ana', 'Luis', 'Ana'], t)).toBe('3 new messages');
  });

  it('speaks Spanish too', () => {
    expect(announcementText(['Ana'], makeTranslate('es'))).toContain('Ana');
  });
});
