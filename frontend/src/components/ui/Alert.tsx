import type { ReactNode } from "react";
import { Icon } from "../../lib/icons";
import type { Tone } from "../../lib/statusTone";

export interface AlertProps {
  tone?: Extract<Tone, "info" | "success" | "warning" | "danger">;
  title?: string;
  children: ReactNode;
}

const iconForTone = { info: "info", success: "check", warning: "alert", danger: "alert" } as const;

/** Inline message used for validation summaries and confirmations. */
export function Alert({ tone = "info", title, children }: AlertProps) {
  return (
    <div className={`hs-alert hs-alert--${tone}`} role={tone === "danger" ? "alert" : "status"}>
      <Icon name={iconForTone[tone]} size={18} />
      <div className="hs-alert__body">
        {title && <div className="hs-alert__title">{title}</div>}
        <div>{children}</div>
      </div>
    </div>
  );
}
