import { useEffect } from 'react';

/**
 * For inline panels opened by a button: remembers the element that had focus when the panel
 * mounted (the trigger) and gives focus back to it when the panel unmounts. Unlike
 * useFocusTrap it does not keep Tab inside the panel.
 */
export function useRestoreFocus() {
  useEffect(() => {
    const trigger = document.activeElement instanceof HTMLElement ? document.activeElement : null;
    return () => {
      if (trigger?.isConnected) trigger.focus();
    };
  }, []);
}
