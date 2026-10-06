import type { ReactNode } from "react";
import { Icon } from "../../lib/icons";
import { Input } from "../ui/FormField";
import { SegmentedControl, type SegmentedOption } from "../ui/SegmentedControl";

export interface FilterBarProps<T extends string> {
  search: string;
  onSearchChange: (value: string) => void;
  searchPlaceholder?: string;
  filters?: ReadonlyArray<SegmentedOption<T>>;
  activeFilter?: T;
  onFilterChange?: (value: T) => void;
  actions?: ReactNode;
}

/** Search + status filter row used on the dashboard (SRS §9.2). */
export function FilterBar<T extends string>({
  search,
  onSearchChange,
  searchPlaceholder = "Search FD Number, Customer or Member No…",
  filters,
  activeFilter,
  onFilterChange,
  actions,
}: FilterBarProps<T>) {
  return (
    <div className="hs-filter-bar">
      <div className="hs-filter-bar__search">
        <div style={{ position: "relative" }}>
          <span
            aria-hidden="true"
            style={{
              position: "absolute",
              left: 10,
              top: "50%",
              transform: "translateY(-50%)",
              color: "var(--color-text-muted)",
              display: "flex",
            }}
          >
            <Icon name="search" size={15} />
          </span>
          <Input
            type="search"
            value={search}
            placeholder={searchPlaceholder}
            aria-label="Search FDs"
            onChange={(e) => onSearchChange(e.target.value)}
            style={{ paddingInlineStart: 32 }}
          />
        </div>
      </div>

      {filters && activeFilter !== undefined && onFilterChange && (
        <SegmentedControl
          label="Filter FDs"
          options={filters}
          value={activeFilter}
          onChange={onFilterChange}
        />
      )}

      {actions}
    </div>
  );
}
