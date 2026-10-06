import type { ReactNode } from "react";
import { Button } from "./Button";

export function Loading({ label = "Loading…" }: { label?: string }) {
  return (
    <div className="hs-loading" role="status">
      <span className="hs-spinner hs-spinner--lg" />
      <span>{label}</span>
    </div>
  );
}

export interface EmptyStateProps {
  title: string;
  description?: ReactNode;
  action?: ReactNode;
}

export function EmptyState({ title, description, action }: EmptyStateProps) {
  return (
    <div className="hs-empty">
      <div className="hs-empty__title">{title}</div>
      {description && <div>{description}</div>}
      {action}
    </div>
  );
}

export interface ErrorStateProps {
  message: string;
  onRetry?: () => void;
}

/** Error state with an explanation and a safe retry (design.md §6). */
export function ErrorState({ message, onRetry }: ErrorStateProps) {
  return (
    <div className="hs-error-state" role="alert">
      <p>{message}</p>
      {onRetry && (
        <Button variant="primary" icon="refresh" onClick={onRetry}>
          Try Again
        </Button>
      )}
    </div>
  );
}
