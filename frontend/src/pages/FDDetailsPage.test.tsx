import { fireEvent, render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { MemoryRouter, Route, Routes } from "react-router-dom";
import { beforeEach, describe, expect, it, vi } from "vitest";

vi.mock("../../wailsjs/go/main/App", () => ({
  GetFD: vi.fn(),
  ReopenFD: vi.fn(),
  ReverseRenewal: vi.fn(),
}));

import { GetFD, ReopenFD, ReverseRenewal } from "../../wailsjs/go/main/App";
import { FDDetailsPage } from "./FDDetailsPage";

const baseFd = {
  fdNumber: "FD-26-001",
  customerName: "Asha",
  customerNumber: "M-1",
  principal: 50000,
  startDate: "2026-01-01",
  tenureDays: 365,
  interestRate: 4,
  maturityDate: "2027-01-01",
  interestAmount: 2000,
  maturityAmount: 52000,
  status: "ACTIVE",
  createdAt: "",
  updatedAt: "",
};

const closedMaturedFd = {
  ...baseFd,
  status: "CLOSED",
  closureDate: "2027-01-01",
  closureType: "MATURED",
  closureRemark: "Paid out",
  closurePayable: 52000,
};

const renewedOldFd = {
  ...baseFd,
  status: "CLOSED",
  closureDate: "2027-01-01",
  closureType: "RENEWED",
  closureRemark: "Renewed as FD-27-002",
  renewedTo: "FD-27-002",
};

const renewalNewFd = {
  ...baseFd,
  fdNumber: "FD-27-002",
  startDate: "2027-01-01",
  maturityDate: "2028-01-01",
  renewedFrom: "FD-26-001",
};

function renderDetails(fdNumber = "FD-26-001") {
  return render(
    <MemoryRouter initialEntries={[`/fd/${fdNumber}`]}>
      <Routes>
        <Route path="/fd/:fdNumber" element={<FDDetailsPage />} />
      </Routes>
    </MemoryRouter>,
  );
}

describe("FDDetailsPage reversal actions", () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it("reopens a closed FD with a mandatory reason", async () => {
    vi.mocked(GetFD).mockResolvedValue({ fd: closedMaturedFd, history: [] } as never);
    vi.mocked(ReopenFD).mockResolvedValue({ ...closedMaturedFd, status: "ACTIVE" } as never);

    renderDetails();

    fireEvent.click(await screen.findByRole("button", { name: "Reopen FD" }));

    const confirm = await screen.findByRole("button", { name: "Confirm Reopen" });
    expect(confirm).toBeDisabled();

    await userEvent.type(screen.getByLabelText(/Reason/), "wrong closure date");
    expect(screen.getByRole("button", { name: "Confirm Reopen" })).toBeEnabled();

    fireEvent.click(screen.getByRole("button", { name: "Confirm Reopen" }));
    await vi.waitFor(() => {
      expect(ReopenFD).toHaveBeenCalledWith({ fdNumber: "FD-26-001", remark: "wrong closure date" });
    });
  });

  it("offers reversal on a renewed FD but not reopen", async () => {
    vi.mocked(GetFD).mockResolvedValue({ fd: renewedOldFd, history: [] } as never);

    renderDetails();

    expect(await screen.findByRole("button", { name: "Reverse Renewal" })).toBeInTheDocument();
    expect(screen.queryByRole("button", { name: "Reopen FD" })).not.toBeInTheDocument();
  });

  it("reverses a renewal from the successor FD page", async () => {
    vi.mocked(GetFD).mockResolvedValue({ fd: renewalNewFd, history: [] } as never);
    vi.mocked(ReverseRenewal).mockResolvedValue({ ...baseFd, status: "ACTIVE" } as never);

    renderDetails("FD-27-002");

    fireEvent.click(await screen.findByRole("button", { name: "Reverse Renewal" }));
    await userEvent.type(screen.getByLabelText(/Reason/), "renewed by mistake");
    fireEvent.click(screen.getByRole("button", { name: "Confirm Reversal" }));

    await vi.waitFor(() => {
      expect(ReverseRenewal).toHaveBeenCalledWith({ fdNumber: "FD-26-001", remark: "renewed by mistake" });
    });
  });

  it("hides reversal actions on a plain active FD", async () => {
    vi.mocked(GetFD).mockResolvedValue({ fd: baseFd, history: [] } as never);

    renderDetails();

    expect(await screen.findByRole("button", { name: "Close FD" })).toBeInTheDocument();
    expect(screen.queryByRole("button", { name: "Reopen FD" })).not.toBeInTheDocument();
    expect(screen.queryByRole("button", { name: "Reverse Renewal" })).not.toBeInTheDocument();
  });

  it("disables Renew until the FD matures", async () => {
    vi.mocked(GetFD).mockResolvedValue({ fd: baseFd, history: [] } as never);

    renderDetails();

    const renew = await screen.findByRole("button", { name: "Renew" });
    expect(renew).toBeDisabled();
    expect(renew).toHaveAttribute("title", expect.stringContaining("Available from"));
  });

  it("enables Renew on a matured active FD", async () => {
    vi.mocked(GetFD).mockResolvedValue({ fd: { ...baseFd, maturityDate: "2026-01-01" }, history: [] } as never);

    renderDetails();

    expect(await screen.findByRole("button", { name: "Renew" })).toBeEnabled();
  });

  it("offers Edit on an active FD but not on a closed one", async () => {
    vi.mocked(GetFD).mockResolvedValue({ fd: baseFd, history: [] } as never);

    renderDetails();

    expect(await screen.findByRole("button", { name: "Edit" })).toBeInTheDocument();
  });

  it("hides Edit on a closed FD", async () => {
    vi.mocked(GetFD).mockResolvedValue({ fd: closedMaturedFd, history: [] } as never);

    renderDetails();

    expect(await screen.findByText("Maturity Date")).toBeInTheDocument();
    expect(screen.queryByRole("button", { name: "Edit" })).not.toBeInTheDocument();
    expect(screen.queryByRole("button", { name: "Renew" })).not.toBeInTheDocument();
  });
});
