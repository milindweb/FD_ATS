import type { ReactNode } from "react";
import { statusLabel, statusTone, type Tone } from "../../lib/statusTone";

export interface BadgeProps {
  tone?: Tone;
  children: ReactNode;
}

export function Badge({ tone = "neutral", children }: BadgeProps) {
  return <span className={`hs-badge hs-badge--${tone}`}>{children}</span>;
}

export interface StatusBadgeProps {
  /** Status or type key straight from the API — never a UI-invented colour. */
  status: string | null | undefined;
  /** Optional secondary key, e.g. closure type on a CLOSED FD. */
  detail?: string | null;
}

/** Shows status text with its mapped tone (design.md §8, §55). */
export function StatusBadge({ status, detail }: StatusBadgeProps) {
  return (
    <span className="u-row u-gap-1">
      <Badge tone={statusTone(status)}>{statusLabel(status)}</Badge>
      {detail && <Badge tone={statusTone(detail)}>{statusLabel(detail)}</Badge>}
    </span>
  );
}
