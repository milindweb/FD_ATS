import { useEffect, useRef, useState } from "react";
import { ListMembers, errorMessage, type Member } from "../../lib/api";
import { useDebounced, useEscapeKey } from "../../lib/hooks";
import { IconButton } from "../ui/Button";
import { FormField, Input } from "../ui/FormField";

export interface MemberPickerProps {
  selected: { id: number; genNo: string; name: string } | null;
  onSelect: (m: { id: number; genNo: string; name: string }) => void;
}

/**
 * Debounced member search dropdown for FD forms (SRS §10, §53).
 * Clearing calls onSelect with id: 0 (genNo/name empty) — the parent treats
 * id === 0 as "no member selected".
 */
export function MemberPicker({ selected, onSelect }: MemberPickerProps) {
  const [search, setSearch] = useState("");
  const debouncedSearch = useDebounced(search, 300);
  const [results, setResults] = useState<Member[]>([]);
  const [open, setOpen] = useState(false);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const rootRef = useRef<HTMLDivElement>(null);

  useEffect(() => {
    const term = debouncedSearch.trim();
    if (term.length < 1) {
      setResults([]);
      setOpen(false);
      setLoading(false);
      return;
    }
    let cancelled = false;
    setLoading(true);
    ListMembers({ search: term, page: 1, pageSize: 8 })
      .then((res) => {
        if (!cancelled) {
          setResults(res.items);
          setOpen(true);
          setLoading(false);
          setError(null);
        }
      })
      .catch((err: unknown) => {
        if (!cancelled) {
          setResults([]);
          setOpen(true);
          setLoading(false);
          setError(errorMessage(err));
        }
      });
    return () => {
      cancelled = true;
    };
  }, [debouncedSearch]);

  useEffect(() => {
    if (!open) return;
    const onDown = (e: MouseEvent) => {
      if (rootRef.current && !rootRef.current.contains(e.target as Node)) setOpen(false);
    };
    document.addEventListener("mousedown", onDown);
    return () => document.removeEventListener("mousedown", onDown);
  }, [open]);

  useEscapeKey(open, () => setOpen(false));

  const chosen = selected && selected.id > 0 ? selected : null;

  const pick = (m: Member) => {
    onSelect({ id: m.id, genNo: m.genNo, name: m.name });
    setSearch("");
    setResults([]);
    setOpen(false);
    setError(null);
  };

  const clear = () => {
    // id === 0 signals "cleared" to the parent.
    onSelect({ id: 0, genNo: "", name: "" });
    setSearch("");
    setResults([]);
    setOpen(false);
  };

  return (
    <div ref={rootRef}>
      {chosen && (
        <div className="u-row u-row--between u-gap-2 u-mb-3" style={{ padding: "6px 10px", border: "1px solid var(--color-border)", borderRadius: "var(--radius-md)", background: "var(--color-surface-hover)" }}>
          <span className="u-row u-gap-2" style={{ minWidth: 0 }}>
            <span className="hs-mono u-text-sm">{chosen.genNo}</span>
            <span className="u-semibold" style={{ overflow: "hidden", textOverflow: "ellipsis", whiteSpace: "nowrap" }}>
              {chosen.name}
            </span>
          </span>
          <IconButton icon="close" label="Clear member selection" onClick={clear} />
        </div>
      )}

      <div style={{ position: "relative" }}>
        <FormField label="Member" htmlFor="member-picker-search">
          <Input
            id="member-picker-search"
            type="search"
            value={search}
            placeholder="Search member by GEN No., Name, Token, Designation, PAN, Aadhaar or Mobile"
            aria-label="Search member by GEN No., Name, Token, Designation, PAN, Aadhaar or Mobile"
            autoComplete="off"
            onChange={(e) => setSearch(e.target.value)}
            onFocus={() => {
              if (results.length > 0) setOpen(true);
            }}
          />
        </FormField>

        {open && (
          <div
            style={{
              position: "absolute",
              top: "100%",
              left: 0,
              right: 0,
              marginTop: 4,
              zIndex: 50,
              maxHeight: 320,
              overflowY: "auto",
              background: "var(--color-surface-raised)",
              border: "1px solid var(--color-border)",
              borderRadius: "var(--radius-md)",
              boxShadow: "0 4px 12px rgba(0, 0, 0, 0.08)",
              padding: 6,
            }}
          >
            {loading ? (
              <p className="u-text-sm u-muted" style={{ padding: "8px 10px" }}>
                Searching members…
              </p>
            ) : error ? (
              <p className="u-text-sm u-danger" role="alert" style={{ padding: "8px 10px" }}>
                {error}
              </p>
            ) : results.length === 0 ? (
              <p className="u-text-sm u-muted" style={{ padding: "8px 10px" }}>
                No members match your search.
              </p>
            ) : (
              <div className="hs-report-list" role="listbox" aria-label="Member results">
                {results.map((m) => (
                  <button
                    type="button"
                    key={m.id}
                    role="option"
                    aria-selected={chosen?.id === m.id}
                    className="hs-report-option"
                    onClick={() => pick(m)}
                  >
                    <span className="hs-report-option__title">
                      <span className="hs-mono">{m.genNo}</span> — {m.name}
                    </span>
                    <span className="hs-report-option__file">
                      {m.designation || "—"}
                      {m.mobile ? ` · ${m.mobile}` : ""}
                    </span>
                  </button>
                ))}
              </div>
            )}
          </div>
        )}
      </div>
    </div>
  );
}
