import { act, createElement, type ReactElement, type ReactNode } from 'react';
import { createRoot } from 'react-dom/client';

(globalThis as { IS_REACT_ACT_ENVIRONMENT?: boolean }).IS_REACT_ACT_ENVIRONMENT = true;

type Wrapper = (props: { children: ReactNode }) => ReactElement;

const mounted = new Set<() => Promise<void>>();

/** Unmounts everything mounted so far (call in afterEach, or listeners of earlier tests keep running). */
export const cleanup = async () => {
  for (const unmount of [...mounted]) await unmount();
};

/** Mounts `ui` (inside `wrapper`) and returns helpers; no @testing-library needed. */
export async function mount(ui: ReactElement, wrapper?: Wrapper) {
  const container = document.createElement('div');
  document.body.appendChild(container);
  const root = createRoot(container);
  const wrap = (el: ReactElement) => (wrapper ? createElement(wrapper, null, el) : el);
  await act(async () => root.render(wrap(ui)));
  const unmount = async () => {
    mounted.delete(unmount);
    await act(async () => root.unmount());
    container.remove();
  };
  mounted.add(unmount);
  return {
    container,
    rerender: (next: ReactElement) => act(async () => root.render(wrap(next))),
    unmount,
  };
}

/** Renders a hook and counts how often its host component renders. */
export async function renderHookCounting<T>(hook: () => T, wrapper?: Wrapper) {
  const state = { renders: 0, current: undefined as T };
  function Host() {
    state.renders++;
    state.current = hook();
    return null;
  }
  const view = await mount(createElement(Host), wrapper);
  /** Re-renders the wrapper and host (e.g. after a variable the wrapper reads changed). */
  const refresh = () => view.rerender(createElement(Host));
  return { state, ...view, refresh };
}

/** Lets pending promises and zero-delay timers settle inside act. */
export const flush = () => act(async () => new Promise<void>((resolve) => setTimeout(resolve, 0)));
