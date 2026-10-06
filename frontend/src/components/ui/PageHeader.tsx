import type { ReactNode } from "react";

export interface PageHeaderProps {
  title: ReactNode;
  context?: ReactNode;
  actions?: ReactNode;
}

/** Standard page heading: title, context line and primary actions. */
export function PageHeader({ title, context, actions }: PageHeaderProps) {
  return (
    <div className="hs-page-head">
      <div className="hs-page-head__titles">
        <h1 className="hs-page-head__title">{title}</h1>
        {context && <div className="hs-page-head__context">{context}</div>}
      </div>
      {actions && <div className="hs-page-head__actions">{actions}</div>}
    </div>
  );
}
