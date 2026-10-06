import type { ReactNode } from "react";

export interface CardProps {
  children: ReactNode;
  className?: string;
}

export function Card({ children, className = "" }: CardProps) {
  return <section className={`hs-card ${className}`}>{children}</section>;
}

export function CardHead({ title, actions }: { title: ReactNode; actions?: ReactNode }) {
  return (
    <header className="hs-card__head">
      <h2 className="hs-card__title">{title}</h2>
      {actions}
    </header>
  );
}

export function CardBody({ children, flush }: { children: ReactNode; flush?: boolean }) {
  return <div className={`hs-card__body ${flush ? "hs-card__body--flush" : ""}`}>{children}</div>;
}
