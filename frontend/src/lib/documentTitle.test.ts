import { describe, expect, it } from 'vitest';
import { composeTitle, setBaseTitle, setUnreadCount, titleKeyForPath } from './documentTitle';

describe('titleKeyForPath', () => {
  it('maps each route to its title key', () => {
    expect(titleKeyForPath('/login')).toBe('title.login');
    expect(titleKeyForPath('/reset-password/abc')).toBe('title.reset');
    expect(titleKeyForPath('/verify/abc')).toBe('title.verify');
    expect(titleKeyForPath('/dashboard')).toBe('title.dashboard');
  });

  it('has none for unknown routes', () => {
    expect(titleKeyForPath('/')).toBeNull();
    expect(titleKeyForPath('/nope')).toBeNull();
  });
});

describe('composeTitle and the unread prefix', () => {
  it('joins page and app name', () => {
    expect(composeTitle('Chats', 'WebChat')).toBe('Chats · WebChat');
    expect(composeTitle(null, 'WebChat')).toBe('WebChat');
  });

  it('keeps the page title when the unread count changes', () => {
    setBaseTitle('Chats · WebChat');
    setUnreadCount(3);
    expect(document.title).toBe('(3) Chats · WebChat');
    setBaseTitle('Chats · WebChat');
    expect(document.title).toBe('(3) Chats · WebChat');
    setUnreadCount(0);
    expect(document.title).toBe('Chats · WebChat');
  });
});
