import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { MemoryRouter, Route, Routes } from "react-router-dom";
import { beforeEach, describe, expect, it, vi } from "vitest";

vi.mock("../../wailsjs/go/main/App", () => ({
  ListFDs: vi.fn().mockResolvedValue({
    items: [
      {
        fdNumber: "FD-1",
        customerName: "Asha",
        customerNumber: "GEN-1",
        fdFormNo: "FORM-1",
        principal: 50000,
        startDate: "2026-01-01",
        tenureDays: 365,
        interestRate: 8,
        maturityDate: "2027-01-01",
        interestAmount: 4000,
        maturityAmount: 54000,
        status: "ACTIVE",
        createdAt: "",
        updatedAt: "",
      },
      {
        fdNumber: "FD-2",
        customerName: "Bela",
        customerNumber: "GEN-2",
        fdFormNo: "",
        principal: 25000,
        startDate: "2026-02-01",
        tenureDays: 100,
        interestRate: 4,
        maturityDate: "2026-05-11",
        interestAmount: 274,
        maturityAmount: 25274,
        status: "CLOSED",
        closureDate: "2026-03-01",
        closureType: "PREMATURE",
        createdAt: "",
        updatedAt: "",
      },
    ],
    total: 2,
    page: 1,
    pageSize: 25,
  }),
}));

import { ListFDs } from "../../wailsjs/go/main/App";
import { FDMasterPage } from "./FDMasterPage";

function renderMaster(path = "/fds") {
  return render(
    <MemoryRouter initialEntries={[path]}>
      <Routes>
        <Route path="/fds" element={<FDMasterPage />} />
        <Route path="/fd/:fdNumber" element={<div>FD details</div>} />
      </Routes>
    </MemoryRouter>,
  );
}

describe("FDMasterPage", () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it("lists the FDs with the SRS member column (SRS §9.2)", async () => {
    renderMaster();

    expect(await screen.findByText("FD-1")).toBeInTheDocument();
    expect(screen.getByText("FD-2")).toBeInTheDocument();
    expect(screen.getByText("Asha")).toBeInTheDocument();
    expect(screen.getByText("Member")).toBeInTheDocument();
    expect(screen.getByText("Create New FD")).toBeInTheDocument();
    expect(ListFDs).toHaveBeenCalledWith({ search: "", filter: "ALL", page: 1, pageSize: 25 });
  });

  it("prefills the search from ?q= and passes it to the backend", async () => {
    renderMaster("/fds?q=FD-1");

    expect(await screen.findByDisplayValue("FD-1")).toBeInTheDocument();
    await vi.waitFor(() => {
      expect(ListFDs).toHaveBeenCalledWith({ search: "FD-1", filter: "ALL", page: 1, pageSize: 25 });
    });
  });

  it("refetches when a status filter is chosen", async () => {
    renderMaster();

    await screen.findByText("FD-1");
    await userEvent.click(screen.getByText("Active"));

    await vi.waitFor(() => {
      expect(ListFDs).toHaveBeenCalledWith({ search: "", filter: "ACTIVE", page: 1, pageSize: 25 });
    });
  });

  it("opens the FD details screen when a row is clicked", async () => {
    renderMaster();

    await userEvent.click(await screen.findByText("FD-1"));

    expect(await screen.findByText("FD details")).toBeInTheDocument();
  });
});
