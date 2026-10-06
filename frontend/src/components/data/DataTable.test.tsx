import { render, screen, within } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { describe, expect, it, vi } from "vitest";
import { DataTable } from "./DataTable";

interface Row {
  id: string;
  name: string;
  amount: number;
}

const columns = [
  { key: "id", header: "FD Number", render: (r: Row) => r.id },
  { key: "name", header: "Customer", render: (r: Row) => r.name },
  { key: "amount", header: "Deposit", numeric: true, render: (r: Row) => r.amount },
];

const rows: Row[] = [
  { id: "FD-26-001", name: "Asha", amount: 50000 },
  { id: "FD-26-002", name: "Ravi", amount: 125000 },
];

describe("DataTable", () => {
  it("renders headers and rows", () => {
    render(<DataTable columns={columns} rows={rows} rowKey={(r) => r.id} />);
    expect(screen.getByRole("columnheader", { name: "FD Number" })).toBeInTheDocument();
    expect(screen.getByText("FD-26-001")).toBeInTheDocument();
    expect(screen.getByText("Ravi")).toBeInTheDocument();
    expect(screen.getAllByRole("row")).toHaveLength(3); // header + 2 rows
  });

  it("shows the loading state without rows", () => {
    render(<DataTable columns={columns} rows={[]} rowKey={(r) => r.id} loading />);
    expect(screen.getByRole("status")).toBeInTheDocument();
    expect(screen.queryByText("FD-26-001")).not.toBeInTheDocument();
  });

  it("shows the empty state with a hint", () => {
    render(<DataTable columns={columns} rows={[]} rowKey={(r) => r.id} />);
    expect(screen.getByText("No FDs found")).toBeInTheDocument();
    expect(screen.getByText(/create a new Fixed Deposit/i)).toBeInTheDocument();
  });

  it("notifies when a row is clicked", async () => {
    const onRowClick = vi.fn();
    render(<DataTable columns={columns} rows={rows} rowKey={(r) => r.id} onRowClick={onRowClick} />);
    await userEvent.click(screen.getByText("FD-26-002"));
    expect(onRowClick).toHaveBeenCalledWith(rows[1]);
  });

  it("labels each cell with its header for the responsive card layout", () => {
    render(<DataTable columns={columns} rows={rows} rowKey={(r) => r.id} />);
    const cell = within(screen.getByText("Asha").closest("td")!).getByText("Asha");
    expect(cell.closest("td")).toHaveAttribute("data-label", "Customer");
  });
});
