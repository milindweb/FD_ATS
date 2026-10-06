import type { InputHTMLAttributes, ReactNode, SelectHTMLAttributes, TextareaHTMLAttributes } from "react";

export interface FormFieldProps {
  label: string;
  required?: boolean;
  error?: string | null;
  hint?: string;
  htmlFor?: string;
  children: ReactNode;
}

/** Label + control + hint/error wrapper. Every input in the app sits in one. */
export function FormField({ label, required, error, hint, htmlFor, children }: FormFieldProps) {
  return (
    <div className="hs-field">
      <label className="hs-field__label" htmlFor={htmlFor}>
        {label}
        {required && (
          <span className="hs-field__required" aria-hidden="true">
            *
          </span>
        )}
      </label>
      {children}
      {error ? (
        <span className="hs-field__error" role="alert">
          {error}
        </span>
      ) : hint ? (
        <span className="hs-field__hint">{hint}</span>
      ) : null}
    </div>
  );
}

export interface InputProps extends InputHTMLAttributes<HTMLInputElement> {
  invalid?: boolean;
  large?: boolean;
}

export function Input({ invalid, large, className = "", ...rest }: InputProps) {
  const classes = ["hs-input", large ? "hs-input--lg" : "", invalid ? "is-invalid" : "", className]
    .filter(Boolean)
    .join(" ");
  return <input className={classes} aria-invalid={invalid || undefined} {...rest} />;
}

export interface SelectProps extends SelectHTMLAttributes<HTMLSelectElement> {
  invalid?: boolean;
}

export function Select({ invalid, className = "", children, ...rest }: SelectProps) {
  const classes = ["hs-select", invalid ? "is-invalid" : "", className].filter(Boolean).join(" ");
  return (
    <select className={classes} aria-invalid={invalid || undefined} {...rest}>
      {children}
    </select>
  );
}

export interface TextareaProps extends TextareaHTMLAttributes<HTMLTextAreaElement> {
  invalid?: boolean;
}

export function Textarea({ invalid, className = "", ...rest }: TextareaProps) {
  const classes = ["hs-textarea", invalid ? "is-invalid" : "", className].filter(Boolean).join(" ");
  return <textarea className={classes} aria-invalid={invalid || undefined} {...rest} />;
}
