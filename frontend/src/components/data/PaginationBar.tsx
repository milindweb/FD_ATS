import { Button } from "../ui/Button";

export interface PaginationBarProps {
  page: number;
  pageSize: number;
  total: number;
  onPageChange: (page: number) => void;
}

/** Page navigation + result count below list tables. */
export function PaginationBar({ page, pageSize, total, onPageChange }: PaginationBarProps) {
  const lastPage = Math.max(1, Math.ceil(total / pageSize));
  const from = total === 0 ? 0 : (page - 1) * pageSize + 1;
  const to = Math.min(total, page * pageSize);

  return (
    <div className="hs-pagination">
      <span>
        Showing {from}–{to} of {total} FD{total === 1 ? "" : "s"}
      </span>
      <div className="hs-pagination__controls">
        <Button size="sm" icon="chevron-left" disabled={page <= 1} onClick={() => onPageChange(page - 1)}>
          Previous
        </Button>
        <Button size="sm" disabled={page >= lastPage} onClick={() => onPageChange(page + 1)}>
          Next
        </Button>
      </div>
    </div>
  );
}
