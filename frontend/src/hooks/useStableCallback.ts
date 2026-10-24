import { useCallback, useEffect, useRef } from 'react';

/** A function with a constant identity that always calls the latest `fn`: lets memoized children skip re-renders. */
export const useStableCallback = <A extends unknown[], R>(fn: (...args: A) => R): ((...args: A) => R) => {
  const ref = useRef(fn);
  useEffect(() => {
    ref.current = fn;
  });
  return useCallback((...args: A) => ref.current(...args), []);
};
