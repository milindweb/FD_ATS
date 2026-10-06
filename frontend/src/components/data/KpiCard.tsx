import { Icon, type IconName } from "../../lib/icons";

export type KpiTone = "danger" | "success" | "warning" | "info" | "violet";

export interface KpiCardProps {
  label: string;
  value: string;
  meta?: string;
  tone?: KpiTone;
  icon?: IconName;
}

/** One dashboard metric (SRS §9.1). */
export function KpiCard({ label, value, meta, tone = "info", icon }: KpiCardProps) {
  return (
    <div className={`hs-kpi hs-kpi--${tone}`}>
      <div className="hs-kpi__top">
        <span className="hs-kpi__label">{label}</span>
        {icon && <Icon name={icon} size={16} className="hs-kpi__icon" />}
      </div>
      <div className="hs-kpi__value">{value}</div>
      {meta && <div className="hs-kpi__meta">{meta}</div>}
    </div>
  );
}
