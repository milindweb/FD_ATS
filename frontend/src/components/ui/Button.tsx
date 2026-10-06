import type { ButtonHTMLAttributes, ReactNode } from "react";
import { Icon, type IconName } from "../../lib/icons";

export type ButtonVariant = "primary" | "secondary" | "danger" | "ghost";
export type ButtonSize = "sm" | "md" | "lg";

export interface ButtonProps extends ButtonHTMLAttributes<HTMLButtonElement> {
  variant?: ButtonVariant;
  size?: ButtonSize;
  loading?: boolean;
  icon?: IconName;
  block?: boolean;
  children?: ReactNode;
}

/** The single button component — every action in the app uses it. */
export function Button({
  variant = "secondary",
  size = "md",
  loading = false,
  icon,
  block = false,
  className = "",
  children,
  disabled,
  type = "button",
  ...rest
}: ButtonProps) {
  const classes = [
    "hs-btn",
    `hs-btn--${variant}`,
    size !== "md" ? `hs-btn--${size}` : "",
    block ? "hs-btn--block" : "",
    className,
  ]
    .filter(Boolean)
    .join(" ");

  return (
    <button type={type} className={classes} disabled={disabled || loading} {...rest}>
      {loading ? <span className="hs-spinner" style={{ width: 14, height: 14 }} /> : icon ? <Icon name={icon} /> : null}
      {children}
    </button>
  );
}

export interface IconButtonProps extends ButtonHTMLAttributes<HTMLButtonElement> {
  icon: IconName;
  label: string;
  size?: "md" | "lg";
}

/** Icon-only control; `label` becomes the accessible name. */
export function IconButton({ icon, label, size = "md", className = "", type = "button", ...rest }: IconButtonProps) {
  const classes = ["hs-icon-btn", size === "lg" ? "hs-icon-btn--lg" : "", className].filter(Boolean).join(" ");
  return (
    <button type={type} className={classes} aria-label={label} title={label} {...rest}>
      <Icon name={icon} size={size === "lg" ? 20 : 16} />
    </button>
  );
}
