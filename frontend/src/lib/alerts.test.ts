import { describe, expect, it } from 'vitest';
import { decideAlert, DEFAULT_ALERT_PREFS, formatTitle, parseAlertPrefs, previewText, type AlertContext } from './alerts';

const base: AlertContext = {
  authorId: 2,
  myId: 1,
  roomId: 'room-a',
  activeRoomId: 'room-b',
  pageActive: true,
  prefs: { sound: true, desktop: true },
  permission: 'granted',
};

describe('decideAlert', () => {
  it('plays a sound for a message in another room while you are looking at the app', () => {
    expect(decideAlert(base)).toEqual({ sound: true, desktop: false });
  });

  it('shows a desktop notification only when the page is in the background', () => {
    expect(decideAlert({ ...base, pageActive: false })).toEqual({ sound: true, desktop: true });
  });

  it('stays quiet for your own messages', () => {
    expect(decideAlert({ ...base, authorId: 1 })).toEqual({ sound: false, desktop: false });
  });

  it('stays quiet for the conversation you are reading, but not when the window is in the background', () => {
    expect(decideAlert({ ...base, activeRoomId: 'room-a' })).toEqual({ sound: false, desktop: false });
    expect(decideAlert({ ...base, activeRoomId: 'room-a', pageActive: false })).toEqual({ sound: true, desktop: true });
  });

  it('respects the preferences and the browser permission', () => {
    expect(decideAlert({ ...base, prefs: { sound: false, desktop: true }, pageActive: false })).toEqual({ sound: false, desktop: true });
    expect(decideAlert({ ...base, permission: 'denied', pageActive: false })).toEqual({ sound: true, desktop: false });
    expect(decideAlert({ ...base, permission: 'unsupported', pageActive: false }).desktop).toBe(false);
  });

  it('does nothing before the user is known', () => {
    expect(decideAlert({ ...base, myId: undefined })).toEqual({ sound: false, desktop: false });
  });
});

describe('parseAlertPrefs', () => {
  it('uses the defaults for missing or broken data', () => {
    expect(parseAlertPrefs(null)).toEqual(DEFAULT_ALERT_PREFS);
    expect(parseAlertPrefs('not json')).toEqual(DEFAULT_ALERT_PREFS);
    expect(parseAlertPrefs('{"sound":"yes"}')).toEqual(DEFAULT_ALERT_PREFS);
  });

  it('reads valid values and fills the gaps', () => {
    expect(parseAlertPrefs('{"sound":false,"desktop":true}')).toEqual({ sound: false, desktop: true });
    expect(parseAlertPrefs('{"desktop":true}')).toEqual({ sound: true, desktop: true });
  });
});

describe('formatTitle and previewText', () => {
  it('adds an unread counter to the title', () => {
    expect(formatTitle('WebChat', 0)).toBe('WebChat');
    expect(formatTitle('WebChat', 3)).toBe('(3) WebChat');
    expect(formatTitle('WebChat', 250)).toBe('(99+) WebChat');
  });

  it('flattens and truncates message previews', () => {
    expect(previewText('  hello\n\n  world  ')).toBe('hello world');
    expect(previewText('x'.repeat(200), 10)).toBe('xxxxxxxxx…');
  });
});
