import * as React from 'react';

import { Input, type InputProps } from '@/components/ui/input';

export interface FormFieldProps extends InputProps {
  label: string;
  error?: string;
}

/** Label + input + validation message. Works with react-hook-form's `register`. */
const FormField = React.forwardRef<HTMLInputElement, FormFieldProps>(
  ({ label, error, id, name, required, ...props }, ref) => {
    const generatedId = React.useId();
    const fieldId = id ?? generatedId;
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
          aria-describedby={error ? `${fieldId}-error` : undefined}
          {...props}
        />
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
