import { useMemo, useState } from "react";
import { useLocation, useNavigate } from "react-router-dom";
import { ListFDs, fdFilters, type FD } from "../lib/api";
import { formatDate, formatMoney, formatPercent } from "../lib/format";
import { useAsync, useDebounced } from "../lib/hooks";
import { DataTable, type Column } from "../components/data/DataTable";
import { FilterBar } from "../components/data/FilterBar";
import { PaginationBar } from "../components/data/PaginationBar";
import { Button } from "../components/ui/Button";
import { ErrorState } from "../components/ui/States";
import { PageHeader } from "../components/ui/PageHeader";
import { StatusBadge } from "../components/ui/Badge";

const PAGE_SIZE = 25;

/**
 * FD Master — the full FD list with search and filters (SRS §9.2).
 * Arrives pre-filled from the Dashboard quick search via `?q=`.
 */
export function FDMasterPage() {
  const navigate = useNavigate();
  const location = useLocation();
  const [search, setSearch] = useState(
    () => new URLSearchParams(location.search).get("q") ?? "",
  );
  const [filter, setFilter] = useState<(typeof fdFilters)[number]["value"]>("ALL");
  const [page, setPage] = useState(1);
  const debouncedSearch = useDebounced(search, 300);

  const list = useAsync(
    () => ListFDs({ search: debouncedSearch, filter, page, pageSize: PAGE_SIZE }),
    [debouncedSearch, filter, page],
  );

  const columns = useMemo<Array<Column<FD>>>(
    () => [
      { key: "fdNumber", header: "FD Number", mono: true, render: (r) => r.fdNumber },
      { key: "member", header: "Member", render: (r) => r.customerName },
      { key: "principal", header: "Deposit Amount", numeric: true, render: (r) => formatMoney(r.principal) },
      { key: "rate", header: "Interest Rate", numeric: true, render: (r) => formatPercent(r.interestRate) },
      { key: "start", header: "Start Date", render: (r) => formatDate(r.startDate) },
      { key: "maturity", header: "Maturity Date", render: (r) => formatDate(r.maturityDate) },
      { key: "maturityAmount", header: "Maturity Amount", numeric: true, render: (r) => formatMoney(r.maturityAmount) },
      { key: "status", header: "Status", render: (r) => <StatusBadge status={r.status} detail={r.closureType} /> },
    ],
    [],
  );

  return (
    <>
      <PageHeader
        title="FD Master"
        context="All Fixed Deposits — search, filter and manage"
        actions={
          <Button variant="primary" icon="plus" onClick={() => navigate("/new")}>
            Create New FD
          </Button>
        }
      />

      <FilterBar
        search={search}
        onSearchChange={(v) => {
          setSearch(v);
          setPage(1);
        }}
        searchPlaceholder="Search FD Number, FD Form No., Member Name or GEN No…"
        filters={fdFilters}
        activeFilter={filter}
        onFilterChange={(v) => {
          setFilter(v);
          setPage(1);
        }}
      />

      <DataTable
        columns={columns}
        rows={list.data?.items ?? []}
        rowKey={(r) => r.fdNumber}
        loading={list.loading}
        onRowClick={(r) => navigate(`/fd/${r.fdNumber}`)}
        emptyTitle={search || filter !== "ALL" ? "No FDs match your search" : "No Fixed Deposits yet"}
        emptyDescription={
          search || filter !== "ALL"
            ? "Try a different search term or filter."
            : "Create your first Fixed Deposit to get started."
        }
      />
      {!list.loading && !list.error && (list.data?.total ?? 0) > 0 && (
        <PaginationBar page={list.data?.page ?? 1} pageSize={PAGE_SIZE} total={list.data?.total ?? 0} onPageChange={setPage} />
      )}
      {list.error && <ErrorState message={list.error} onRetry={list.reload} />}
    </>
  );
}
