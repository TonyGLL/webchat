import { afterEach, describe, expect, it, vi } from 'vitest';
import { cleanup, flush, mount } from '@/test/harness';
import { formatTimestamp } from './MessageItem';
import CreateRoomForm from './CreateRoomForm';

afterEach(async () => {
  vi.unstubAllGlobals();
  await cleanup();
});

describe('formatTimestamp', () => {
  it('shows only the time for today and adds the date for older messages', () => {
    const now = new Date();
    const old = new Date(now.getFullYear() - 1, 0, 15, 9, 30);
    const today = formatTimestamp(now, 'en');
    const older = formatTimestamp(old, 'en');
    expect(today).toBe(new Intl.DateTimeFormat('en', { hour: 'numeric', minute: 'numeric' }).format(now));
    expect(older).toContain(String(old.getFullYear()));
    expect(older).toContain('Jan');
    expect(formatTimestamp(old, 'es')).toContain('ene');
  });
});

describe('CreateRoomForm', () => {
  it('shows an error instead of swallowing a failed creation', async () => {
    vi.stubGlobal('ResizeObserver', class { observe() {} unobserve() {} disconnect() {} });
    vi.spyOn(console, 'error').mockImplementation(() => undefined);
    const createRoom = vi.fn().mockRejectedValue(new Error('conflict'));
    const { container } = await mount(<CreateRoomForm createRoom={createRoom} />);
    const open = container.querySelector('button') as HTMLButtonElement;
    open.click();
    await flush();
    const input = container.querySelector('input') as HTMLInputElement;
    const set = Object.getOwnPropertyDescriptor(HTMLInputElement.prototype, 'value')!.set!;
    set.call(input, 'general');
    input.dispatchEvent(new Event('input', { bubbles: true }));
    await flush();
    (container.querySelector('form') as HTMLFormElement).requestSubmit();
    await flush();
    expect(createRoom).toHaveBeenCalled();
    expect(container.querySelector('[role="alert"]')?.textContent).toMatch(/could not create/i);
  });
});
