import { fireEvent, render, screen } from "@testing-library/react";
import { MemoryRouter, Route, Routes } from "react-router-dom";
import { beforeEach, describe, expect, it, vi } from "vitest";

vi.mock("../../wailsjs/go/main/App", () => ({
  GetFD: vi.fn(),
  PreviewFD: vi.fn(),
  RenewFD: vi.fn(),
}));

import { GetFD, PreviewFD, RenewFD } from "../../wailsjs/go/main/App";
import { todayISO } from "../lib/format";
import { RenewPage } from "./RenewPage";

const fd = {
  fdNumber: "FD-26-001",
  customerName: "Asha",
  customerNumber: "M-1",
  principal: 50000,
  startDate: "2026-01-01",
  tenureDays: 365,
  interestRate: 4,
  maturityDate: "2027-06-30",
  interestAmount: 2000,
  maturityAmount: 52000,
  status: "ACTIVE",
  createdAt: "",
  updatedAt: "",
};

const calculation = {
  principal: 52000,
  startDate: "2027-06-30",
  tenureDays: 1095,
  ratePercent: 8,
  interest: 13261,
  maturityAmount: 65261,
  maturityDate: "2030-06-30",
};

function renderRenew() {
  return render(
    <MemoryRouter initialEntries={["/fd/FD-26-001/renew"]}>
      <Routes>
        <Route path="/fd/:fdNumber/renew" element={<RenewPage />} />
      </Routes>
    </MemoryRouter>,
  );
}

describe("RenewPage maturity guard", () => {
  beforeEach(() => {
    vi.clearAllMocks();
    vi.mocked(GetFD).mockResolvedValue({ fd, history: [] } as never);
    vi.mocked(PreviewFD).mockResolvedValue(calculation as never);
  });

  it("defaults the start date to the maturity date", async () => {
    renderRenew();

    await screen.findByLabelText(/New Start Date/);
    await vi.waitFor(() => {
      expect((screen.getByLabelText(/New Start Date/) as HTMLInputElement).value).toBe("2027-06-30");
    });
  });

  it("blocks confirmation when the start date is before maturity", async () => {
    renderRenew();

    const start = await screen.findByLabelText(/New Start Date/);
    await vi.waitFor(() => {
      expect((screen.getByLabelText(/New Start Date/) as HTMLInputElement).value).toBe("2027-06-30");
    });
    fireEvent.change(start, { target: { value: "2027-06-29" } });

    expect(await screen.findByText("Renewal cannot start before the FD maturity date.")).toBeInTheDocument();
    expect(screen.getByRole("button", { name: "Confirm Renewal" })).toBeDisabled();
  });

  it("keeps confirmation disabled while the preview fails", async () => {
    vi.mocked(PreviewFD).mockRejectedValue(new Error("No interest rate is configured for this tenure."));

    renderRenew();

    expect(await screen.findByText("No interest rate is configured for this tenure.")).toBeInTheDocument();
    expect(screen.getByRole("button", { name: "Confirm Renewal" })).toBeDisabled();
  });

  it("keeps confirmation disabled before the FD matures", async () => {
    renderRenew();

    expect(await screen.findByText("Renewal not yet available")).toBeInTheDocument();
    expect(screen.getByText(/This FD matures on/)).toBeInTheDocument();
    expect(screen.getByRole("button", { name: "Confirm Renewal" })).toBeDisabled();
  });

  it("renews through the backend once the terms are valid", async () => {
    vi.mocked(GetFD).mockResolvedValue({ fd: { ...fd, maturityDate: "2026-06-30" }, history: [] } as never);
    vi.mocked(RenewFD).mockResolvedValue({ previousFd: fd, newFd: { ...fd, fdNumber: "FD-27-001" } } as never);

    renderRenew();

    await screen.findByRole("button", { name: "Confirm Renewal" });
    await vi.waitFor(() => {
      expect(screen.getByRole("button", { name: "Confirm Renewal" })).toBeEnabled();
    });

    fireEvent.click(screen.getByRole("button", { name: "Confirm Renewal" }));

    await vi.waitFor(() => {
      expect(RenewFD).toHaveBeenCalledWith(
        expect.objectContaining({ fdNumber: "FD-26-001", startDate: todayISO() }),
      );
    });
  });
});
