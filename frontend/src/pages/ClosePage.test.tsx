import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { MemoryRouter, Route, Routes } from "react-router-dom";
import { beforeEach, describe, expect, it, vi } from "vitest";

vi.mock("../../wailsjs/go/main/App", () => ({
  GetFD: vi.fn(),
  PreviewClosure: vi.fn(),
  CloseFD: vi.fn(),
}));

import { GetFD, PreviewClosure } from "../../wailsjs/go/main/App";
import { ClosePage } from "./ClosePage";

const fd = {
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

const prematurePreview = {
  fdNumber: fd.fdNumber,
  closureDate: "2026-10-07",
  isPremature: true,
  closureType: "PREMATURE",
  daysHeld: 1,
  ratePercent: 4,
  interest: 5,
  payable: 50005,
  maturityDate: fd.maturityDate,
};

function renderClose() {
  return render(
    <MemoryRouter initialEntries={["/fd/FD-26-001/close"]}>
      <Routes>
        <Route path="/fd/:fdNumber/close" element={<ClosePage />} />
      </Routes>
    </MemoryRouter>,
  );
}

describe("ClosePage remark flow", () => {
  beforeEach(() => {
    vi.clearAllMocks();
    vi.mocked(GetFD).mockResolvedValue({ fd, history: [] } as never);
  });

  it("re-runs the payable preview with the remark and enables the close button", async () => {
    vi.mocked(PreviewClosure)
      .mockRejectedValueOnce(new Error("Please enter a closure remark."))
      .mockResolvedValue(prematurePreview as never);

    renderClose();

    expect(
      await screen.findByText("Please enter a closure remark.", {}, { timeout: 3000 }),
    ).toBeInTheDocument();
    expect(screen.queryByText("Close FD")).not.toBeInTheDocument();

    const remark = screen.getByLabelText(/Remark/);
    await userEvent.type(remark, "urgent need");

    await vi.waitFor(
      () => {
        expect(PreviewClosure).toHaveBeenLastCalledWith(
          expect.objectContaining({ remark: "urgent need" }),
        );
      },
      { timeout: 3000 },
    );
    expect(await screen.findByText("Close FD", {}, { timeout: 3000 })).toBeInTheDocument();
    expect(screen.queryByText("Please enter a closure remark.")).not.toBeInTheDocument();
  });

  it("keeps the close button disabled while a premature remark is missing", async () => {
    vi.mocked(PreviewClosure).mockResolvedValue(prematurePreview as never);

    renderClose();

    const closeButton = await screen.findByRole("button", { name: "Close FD" }, { timeout: 3000 });
    expect(closeButton).toBeDisabled();

    await userEvent.type(screen.getByLabelText(/Remark/), "reason");
    await vi.waitFor(
      () => {
        expect(screen.getByRole("button", { name: "Close FD" })).toBeEnabled();
      },
      { timeout: 3000 },
    );
  });
});
