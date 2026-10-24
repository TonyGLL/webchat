/** Steps of the welcome tour, in order. Texts live in the i18n dictionaries as `onboarding.<id>.title|body`. */
export const ONBOARDING_STEPS = ['people', 'rooms', 'talk'] as const;
export type OnboardingStep = (typeof ONBOARDING_STEPS)[number];

const keyFor = (userId: number) => `onboarding:${userId}`;

type ReadableStorage = Pick<Storage, 'getItem'>;
type WritableStorage = Pick<Storage, 'setItem'>;

/** The tour is shown once per person on this browser; a blocked storage counts as "not done". */
export function isOnboardingDone(storage: ReadableStorage, userId: number): boolean {
  try {
    return storage.getItem(keyFor(userId)) === 'done';
  } catch {
    return false;
  }
}

export function markOnboardingDone(storage: WritableStorage, userId: number): void {
  try {
    storage.setItem(keyFor(userId), 'done');
  } catch {
    // storage blocked: the tour may show again next time, which is harmless
  }
}

/** Clamps a step index into the tour, so stray values never break the dialog. */
export function clampStep(index: number): number {
  return Math.min(Math.max(index, 0), ONBOARDING_STEPS.length - 1);
}
