import { useCallback, useState, useSyncExternalStore } from 'react';
import { isOnboardingDone, markOnboardingDone } from '@/lib/onboarding';

const listeners = new Set<() => void>();
const subscribe = (onChange: () => void) => {
  listeners.add(onChange);
  return () => void listeners.delete(onChange);
};

/**
 * Whether to show the welcome tour for this user. It opens by itself the first time, can be
 * reopened from the help button, and records completion in this browser (never on the server).
 */
export const useOnboarding = (userId: number | undefined) => {
  const done = useSyncExternalStore(
    subscribe,
    () => (userId === undefined ? true : isOnboardingDone(localStorage, userId)),
    () => true // no tour during server rendering
  );
  const [reopened, setReopened] = useState(false);

  const finish = useCallback(() => {
    if (userId !== undefined) markOnboardingDone(localStorage, userId);
    setReopened(false);
    listeners.forEach((l) => l());
  }, [userId]);

  const reopen = useCallback(() => setReopened(true), []);

  return { open: userId !== undefined && (!done || reopened), finish, reopen };
};
