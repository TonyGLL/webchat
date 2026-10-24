import * as React from 'react';

import { Input, type InputProps } from '@/components/ui/input';

export interface FormFieldProps extends InputProps {
  label: string;
  error?: string;
  /** Persistent help text (e.g. password rules), linked to the input through aria-describedby. */
  hint?: string;
}

/** Label + input + validation message. Works with react-hook-form's `register`. */
const FormField = React.forwardRef<HTMLInputElement, FormFieldProps>(
  ({ label, error, hint, id, name, required, ...props }, ref) => {
    const generatedId = React.useId();
    const fieldId = id ?? generatedId;
    const describedBy = [hint ? `${fieldId}-hint` : null, error ? `${fieldId}-error` : null].filter(Boolean).join(' ') || undefined;
    return (
      <div className="space-y-1.5">
        <label htmlFor={fieldId} className="text-sm font-medium">
          {label}
          {required && (
            <span aria-hidden="true" className="ml-0.5 text-destructive">
              *
            </span>
          )}
        </label>
        <Input
          id={fieldId}
          name={name}
          ref={ref}
          aria-required={required ? true : undefined}
          aria-invalid={error ? true : undefined}
          aria-describedby={describedBy}
          {...props}
        />
        {hint && (
          <p id={`${fieldId}-hint`} className="text-xs text-muted-foreground">
            {hint}
          </p>
        )}
        {error && (
          <p id={`${fieldId}-error`} role="alert" className="text-xs text-destructive">
            {error}
          </p>
        )}
      </div>
    );
  }
);
FormField.displayName = 'FormField';

export { FormField };
