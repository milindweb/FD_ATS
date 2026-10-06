import type { ReactNode } from "react";
import { EmptyState, Loading } from "../ui/States";

export interface Column<T> {
  key: string;
  header: string;
  render: (row: T) => ReactNode;
  numeric?: boolean;
  mono?: boolean;
}

export interface DataTableProps<T> {
  columns: Array<Column<T>>;
  rows: T[];
  rowKey: (row: T) => string;
  loading?: boolean;
  emptyTitle?: string;
  emptyDescription?: ReactNode;
  onRowClick?: (row: T) => void;
  caption?: string;
}

/**
 * Table with loading/empty states and a small-screen card transform.
 * Each cell carries its header as data-label so the mobile layout can label
 * cards without duplicating markup (SRS §7.2).
 */
export function DataTable<T>({
  columns,
  rows,
  rowKey,
  loading,
  emptyTitle = "No FDs found",
  emptyDescription = "Try a different search or create a new Fixed Deposit.",
  onRowClick,
  caption,
}: DataTableProps<T>) {
  if (loading) return <Loading />;
  if (rows.length === 0) return <EmptyState title={emptyTitle} description={emptyDescription} />;

  return (
    <div className="hs-table-wrap is-responsive">
      <table className="hs-table">
        {caption && <caption className="u-hide-sm">{caption}</caption>}
        <thead>
          <tr>
            {columns.map((c) => (
              <th key={c.key} scope="col" className={c.numeric ? "hs-num" : ""}>
                {c.header}
              </th>
            ))}
          </tr>
        </thead>
        <tbody>
          {rows.map((row) => (
            <tr
              key={rowKey(row)}
              className={onRowClick ? "is-clickable" : ""}
              tabIndex={onRowClick ? 0 : undefined}
              onClick={onRowClick ? () => onRowClick(row) : undefined}
              onKeyDown={
                onRowClick
                  ? (e) => {
                      if (e.key === "Enter" || e.key === " ") {
                        e.preventDefault();
                        onRowClick(row);
                      }
                    }
                  : undefined
              }
            >
              {columns.map((c) => (
                <td
                  key={c.key}
                  data-label={c.header}
                  className={[c.numeric ? "hs-num" : "", c.mono ? "hs-mono" : ""].filter(Boolean).join(" ")}
                >
                  {c.render(row)}
                </td>
              ))}
            </tr>
          ))}
        </tbody>
      </table>
    </div>
  );
}
