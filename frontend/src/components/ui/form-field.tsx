import * as React from 'react';

import { Input, type InputProps } from '@/components/ui/input';

export interface FormFieldProps extends InputProps {
  label: string;
  error?: string;
}

/** Label + input + validation message. Works with react-hook-form's `register`. */
const FormField = React.forwardRef<HTMLInputElement, FormFieldProps>(
  ({ label, error, id, name, ...props }, ref) => {
    const fieldId = id ?? name;
    return (
      <div className="space-y-1.5">
        <label htmlFor={fieldId} className="text-sm font-medium">
          {label}
        </label>
        <Input
          id={fieldId}
          name={name}
          ref={ref}
          aria-invalid={error ? true : undefined}
          aria-describedby={error ? `${fieldId}-error` : undefined}
          {...props}
        />
        {error && (
          <p id={`${fieldId}-error`} className="text-xs text-destructive">
            {error}
          </p>
        )}
      </div>
    );
  }
);
FormField.displayName = 'FormField';

export { FormField };
