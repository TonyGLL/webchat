'use client';

import { useEffect, useId, useRef, useState } from 'react';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { useFocusTrap } from '@/hooks/useFocusTrap';
import { useT } from '@/lib/i18n/useT';

export interface ConfirmField {
  name: string;
  label: string;
  /** Help text linked to the input. */
  hint?: string;
  /** Returns an error message, or null when the value is fine. */
  validate?: (value: string) => string | null;
}

export interface ConfirmSpec {
  title: string;
  description?: string;
  confirmLabel: string;
  destructive?: boolean;
  fields?: ConfirmField[];
  /** Runs with the field values once the user confirms and every field validates. */
  run: (values: Record<string, string>) => void | Promise<void>;
}

/** Accessible replacement for window.confirm / window.prompt: alertdialog with focus trap, Escape and focus restore. */
export default function ConfirmDialog({ spec, onClose }: { spec: ConfirmSpec; onClose: () => void }) {
  const { t } = useT();
  const id = useId();
  const dialog = useRef<HTMLFormElement>(null);
  const [values, setValues] = useState<Record<string, string>>({});
  const [errors, setErrors] = useState<Record<string, string>>({});
  const hasFields = (spec.fields?.length ?? 0) > 0;

  // Declared before the autofocus effect so it records the opener before focus moves.
  useFocusTrap(dialog);

  useEffect(() => {
    // First field when there is one; otherwise the safe choice (Cancel).
    const target = dialog.current?.querySelector<HTMLElement>(hasFields ? 'input' : '[data-cancel]');
    target?.focus();
  }, [hasFields]);

  useEffect(() => {
    const onKey = (e: KeyboardEvent) => e.key === 'Escape' && onClose();
    window.addEventListener('keydown', onKey);
    return () => window.removeEventListener('keydown', onKey);
  }, [onClose]);

  const submit = (e: React.FormEvent) => {
    e.preventDefault();
    const found: Record<string, string> = {};
    for (const f of spec.fields ?? []) {
      const message = f.validate?.(values[f.name] ?? '');
      if (message) found[f.name] = message;
    }
    setErrors(found);
    if (Object.keys(found).length > 0) {
      dialog.current?.querySelector<HTMLElement>('[aria-invalid=true]')?.focus();
      return;
    }
    onClose();
    void spec.run(values);
  };

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/40 p-4">
      <form
        ref={dialog}
        role="alertdialog"
        aria-modal="true"
        aria-labelledby={`${id}-title`}
        aria-describedby={spec.description ? `${id}-desc` : undefined}
        onSubmit={submit}
        noValidate
        className="w-full max-w-sm rounded-xl border bg-card p-5 shadow-xl"
      >
        <h2 id={`${id}-title`} className="text-base font-semibold tracking-tight">
          {spec.title}
        </h2>
        {spec.description && (
          <p id={`${id}-desc`} className="mt-2 text-sm text-muted-foreground">
            {spec.description}
          </p>
        )}
        {spec.fields?.map((f) => (
          <div key={f.name} className="mt-4 space-y-1">
            <label htmlFor={`${id}-${f.name}`} className="text-sm font-medium">
              {f.label}
            </label>
            <Input
              id={`${id}-${f.name}`}
              value={values[f.name] ?? ''}
              onChange={(e) => setValues((v) => ({ ...v, [f.name]: e.target.value }))}
              aria-invalid={errors[f.name] ? true : undefined}
              aria-describedby={[f.hint ? `${id}-${f.name}-hint` : '', errors[f.name] ? `${id}-${f.name}-err` : ''].filter(Boolean).join(' ') || undefined}
            />
            {f.hint && (
              <p id={`${id}-${f.name}-hint`} className="text-xs text-muted-foreground">
                {f.hint}
              </p>
            )}
            {errors[f.name] && (
              <p id={`${id}-${f.name}-err`} className="text-xs text-destructive" role="alert">
                {errors[f.name]}
              </p>
            )}
          </div>
        ))}
        <div className="mt-5 flex justify-end gap-2">
          <Button type="button" variant="ghost" data-cancel onClick={onClose}>
            {t('common.cancel')}
          </Button>
          <Button type="submit" variant={spec.destructive ? 'destructive' : 'default'}>
            {spec.confirmLabel}
          </Button>
        </div>
      </form>
    </div>
  );
}
