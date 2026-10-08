import { useMemo, useState } from "react";
import { useNavigate } from "react-router-dom";
import {
  CommitMemberImport,
  ListMembers,
  PickMemberImportPath,
  PreviewMemberImport,
  errorMessage,
  type Member,
  type MemberImportPreview,
  type MemberImportResult,
} from "../lib/api";
import { formatMoney } from "../lib/format";
import { useAsync, useDebounced } from "../lib/hooks";
import { DataTable, type Column } from "../components/data/DataTable";
import { PaginationBar } from "../components/data/PaginationBar";
import { Alert } from "../components/ui/Alert";
import { Badge } from "../components/ui/Badge";
import { Button } from "../components/ui/Button";
import { Input } from "../components/ui/FormField";
import { Modal } from "../components/ui/Modal";
import { PageHeader } from "../components/ui/PageHeader";
import { ErrorState } from "../components/ui/States";

const PAGE_SIZE = 25;

// Row statuses returned by PreviewMemberImport (internal/api/types.go).
const importStatusTone: Record<string, "success" | "info" | "warning" | "danger"> = {
  new: "success",
  existing: "info",
  duplicate: "warning",
  invalid: "danger",
};

type ImportStep = "pick" | "preview" | "result";

export function MemberListPage() {
  const navigate = useNavigate();
  const [search, setSearch] = useState("");
  const [page, setPage] = useState(1);
  const debouncedSearch = useDebounced(search, 300);

  const list = useAsync(
    () => ListMembers({ search: debouncedSearch, page, pageSize: PAGE_SIZE }),
    [debouncedSearch, page],
  );

  const [importOpen, setImportOpen] = useState(false);
  const [importStep, setImportStep] = useState<ImportStep>("pick");
  const [importPath, setImportPath] = useState("");
  const [importPreview, setImportPreview] = useState<MemberImportPreview | null>(null);
  const [importResult, setImportResult] = useState<MemberImportResult | null>(null);
  const [importBusy, setImportBusy] = useState(false);
  const [importError, setImportError] = useState<string | null>(null);

  const columns = useMemo<Array<Column<Member>>>(
    () => [
      { key: "genNo", header: "GEN No.", mono: true, render: (r) => r.genNo },
      { key: "name", header: "Name", render: (r) => r.name },
      { key: "designation", header: "Designation", render: (r) => r.designation || "—" },
      { key: "tokenNo", header: "Token No.", render: (r) => r.tokenNo || "—" },
      { key: "mobile", header: "Mobile", render: (r) => r.mobile || "—" },
      { key: "activeFdCount", header: "Active FDs", numeric: true, render: (r) => String(r.activeFdCount) },
      {
        key: "activeFdAmount",
        header: "Total Active FD Amount",
        numeric: true,
        render: (r) => formatMoney(r.activeFdAmount),
      },
      {
        key: "actions",
        header: "Actions",
        render: (r) => (
          <span className="u-row u-gap-2" onClick={(e) => e.stopPropagation()}>
            <Button size="sm" onClick={() => navigate(`/member/${r.id}`)}>
              View
            </Button>
            <Button size="sm" onClick={() => navigate(`/member/${r.id}/edit`)}>
              Edit
            </Button>
          </span>
        ),
      },
    ],
    [navigate],
  );

  const startImport = () => {
    setImportOpen(true);
    setImportStep("pick");
    setImportPath("");
    setImportPreview(null);
    setImportResult(null);
    setImportError(null);
  };

  const chooseImportFile = async () => {
    setImportError(null);
    setImportBusy(true);
    try {
      const path = await PickMemberImportPath();
      if (!path) {
        setImportOpen(false); // dialog cancelled
        return;
      }
      setImportPath(path);
      setImportPreview(await PreviewMemberImport(path));
      setImportStep("preview");
    } catch (err: unknown) {
      setImportError(errorMessage(err));
    } finally {
      setImportBusy(false);
    }
  };

  const commitImport = async () => {
    if (!importPath) return;
    setImportError(null);
    setImportBusy(true);
    try {
      setImportResult(await CommitMemberImport(importPath));
      setImportStep("result");
      list.reload();
    } catch (err: unknown) {
      setImportError(errorMessage(err));
    } finally {
      setImportBusy(false);
    }
  };

  const restartImport = () => {
    setImportStep("pick");
    setImportPath("");
    setImportPreview(null);
    setImportResult(null);
    setImportError(null);
  };

  return (
    <>
      <PageHeader
        title="Member"
        context="Member master — search, view and manage members and their FDs"
        actions={
          <>
            <Button icon="sheet" onClick={startImport}>
              Bulk Excel Upload
            </Button>
            <Button variant="primary" icon="plus" onClick={() => navigate("/member/new")}>
              Add Member
            </Button>
          </>
        }
      />

      <div className="hs-filter-bar">
        <div className="hs-filter-bar__search">
          <Input
            type="search"
            value={search}
            placeholder="Search GEN No., Name, Token No., Designation, PAN, Aadhaar or Mobile…"
            aria-label="Search members"
            onChange={(e) => {
              setSearch(e.target.value);
              setPage(1);
            }}
          />
        </div>
      </div>

      <DataTable
        columns={columns}
        rows={list.data?.items ?? []}
        rowKey={(r) => String(r.id)}
        loading={list.loading}
        onRowClick={(r) => navigate(`/member/${r.id}`)}
        emptyTitle={search ? "No members match your search" : "No members yet"}
        emptyDescription={
          search
            ? "Try a different search term or check the GEN No."
            : "Add your first member or upload an Excel file to get started."
        }
      />
      {!list.loading && !list.error && (list.data?.total ?? 0) > 0 && (
        <PaginationBar
          page={list.data?.page ?? 1}
          pageSize={PAGE_SIZE}
          total={list.data?.total ?? 0}
          onPageChange={setPage}
        />
      )}
      {list.error && <ErrorState message={list.error} onRetry={list.reload} />}

      <Modal
        open={importOpen}
        title="Bulk Excel Upload"
        wide
        onClose={() => setImportOpen(false)}
        footer={
          <>
            <Button onClick={() => setImportOpen(false)}>Close</Button>
            {importStep === "preview" && (
              <>
                <Button onClick={restartImport} disabled={importBusy}>
                  Choose another file
                </Button>
                <Button
                  variant="primary"
                  onClick={commitImport}
                  loading={importBusy}
                  disabled={!importPreview || importPreview.valid === 0}
                >
                  Import
                </Button>
              </>
            )}
            {importStep === "result" && (
              <Button onClick={restartImport} disabled={importBusy}>
                Choose another file
              </Button>
            )}
          </>
        }
      >
        {importError && (
          <Alert tone="danger" title="Import failed">
            {importError}
          </Alert>
        )}

        {importStep === "pick" && (
          <>
            <p>
              Upload an Excel file with member records. GEN No. and Name are mandatory columns; every
              record is validated before anything is saved.
            </p>
            <div className="u-mt-4">
              <Button variant="primary" icon="sheet" onClick={chooseImportFile} loading={importBusy}>
                Choose Excel file…
              </Button>
            </div>
          </>
        )}

        {importStep === "preview" && importPreview && (
          <>
            <h3 className="u-text-lg u-semibold">Import Summary</h3>
            <dl className="hs-info-grid u-mt-2">
              <div className="hs-info">
                <dt className="hs-info__label">Total Records</dt>
                <dd className="hs-info__value">{importPreview.total}</dd>
              </div>
              <div className="hs-info">
                <dt className="hs-info__label">Valid Records</dt>
                <dd className="hs-info__value">{importPreview.valid}</dd>
              </div>
              <div className="hs-info">
                <dt className="hs-info__label">Duplicate GEN No.</dt>
                <dd className="hs-info__value">{importPreview.duplicateGen}</dd>
              </div>
              <div className="hs-info">
                <dt className="hs-info__label">Missing Mandatory</dt>
                <dd className="hs-info__value">{importPreview.missingMandatory}</dd>
              </div>
              <div className="hs-info">
                <dt className="hs-info__label">Invalid</dt>
                <dd className="hs-info__value">{importPreview.invalid}</dd>
              </div>
              <div className="hs-info">
                <dt className="hs-info__label">Existing in System</dt>
                <dd className="hs-info__value">{importPreview.existing}</dd>
              </div>
            </dl>
            {importPreview.valid === 0 && (
              <div className="u-mt-3">
                <Alert tone="warning" title="Nothing to import">
                  No valid records in this file. Correct the Excel file and choose it again.
                </Alert>
              </div>
            )}
            <div className="u-mt-4">
              <table className="hs-table">
                <thead>
                  <tr>
                    <th scope="col" className="hs-num">Row</th>
                    <th scope="col">GEN No.</th>
                    <th scope="col">Name</th>
                    <th scope="col">Status</th>
                    <th scope="col">Error</th>
                  </tr>
                </thead>
                <tbody>
                  {importPreview.rows.map((r) => (
                    <tr key={`${r.row}-${r.genNo}`}>
                      <td data-label="Row" className="hs-num">{r.row}</td>
                      <td data-label="GEN No." className="hs-mono">{r.genNo || "—"}</td>
                      <td data-label="Name">{r.name || "—"}</td>
                      <td data-label="Status">
                        <Badge tone={importStatusTone[r.status] ?? "neutral"}>{r.status.toUpperCase()}</Badge>
                      </td>
                      <td data-label="Error">{r.error || "—"}</td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
            <p className="u-mt-2 u-text-xs u-muted">
              Existing GEN Nos. are skipped — member information in the system is never overwritten.
            </p>
          </>
        )}

        {importStep === "result" && importResult && (
          <>
            <Alert tone="success" title="Import complete">
              {importResult.imported} new member{importResult.imported === 1 ? "" : "s"} imported,{" "}
              {importResult.skipped} skipped out of {importResult.total} records.
            </Alert>
            <dl className="hs-info-grid u-mt-3">
              <div className="hs-info">
                <dt className="hs-info__label">Imported</dt>
                <dd className="hs-info__value">{importResult.imported}</dd>
              </div>
              <div className="hs-info">
                <dt className="hs-info__label">Skipped</dt>
                <dd className="hs-info__value">{importResult.skipped}</dd>
              </div>
              <div className="hs-info">
                <dt className="hs-info__label">Total</dt>
                <dd className="hs-info__value">{importResult.total}</dd>
              </div>
            </dl>
          </>
        )}
      </Modal>
    </>
  );
}
