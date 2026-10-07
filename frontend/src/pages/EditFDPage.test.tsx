import { fireEvent, render, screen } from "@testing-library/react";
import { MemoryRouter, Route, Routes } from "react-router-dom";
import { beforeEach, describe, expect, it, vi } from "vitest";

vi.mock("../../wailsjs/go/main/App", () => ({
  GetFD: vi.fn(),
  PreviewFD: vi.fn(),
  EditFD: vi.fn(),
}));

import { EditFD, GetFD, PreviewFD } from "../../wailsjs/go/main/App";
import { EditFDPage } from "./EditFDPage";

const activeFd = {
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

const calculation = {
  principal: 75000,
  startDate: "2026-01-01",
  tenureDays: 365,
  ratePercent: 8,
  interest: 6000,
  maturityAmount: 81000,
  maturityDate: "2027-01-01",
};

function renderEdit(fdNumber = "FD-26-001") {
  return render(
    <MemoryRouter initialEntries={[`/fd/${fdNumber}/edit`]}>
      <Routes>
        <Route path="/fd/:fdNumber/edit" element={<EditFDPage />} />
      </Routes>
    </MemoryRouter>,
  );
}

describe("EditFDPage", () => {
  beforeEach(() => {
    vi.clearAllMocks();
    vi.mocked(GetFD).mockResolvedValue({ fd: activeFd, history: [] } as never);
    vi.mocked(PreviewFD).mockResolvedValue(calculation as never);
    vi.mocked(EditFD).mockResolvedValue(activeFd as never);
  });

  it("preloads current values, previews and saves the edit", async () => {
    renderEdit();

    const name = await screen.findByLabelText(/Customer \/ Member Name/);
    expect(name).toHaveValue("Asha");
    expect(screen.getByLabelText(/Deposit Amount/)).toHaveValue(50000);
    expect(screen.getByLabelText(/^Tenure \(days\)/)).toHaveValue(365);

    fireEvent.change(screen.getByLabelText(/Deposit Amount/), { target: { value: "75000" } });

    const save = await screen.findByRole("button", { name: "Save Changes" });
    await vi.waitFor(() => {
      expect(save).toBeEnabled();
    });

    fireEvent.click(save);

    await vi.waitFor(() => {
      expect(EditFD).toHaveBeenCalledWith({
        fdNumber: "FD-26-001",
        customerName: "Asha",
        customerNumber: "M-1",
        principal: 75000,
        startDate: "2026-01-01",
        tenureDays: 365,
      });
    });
  });

  it("disables save when required details are cleared", async () => {
    renderEdit();

    const save = await screen.findByRole("button", { name: "Save Changes" });
    await vi.waitFor(() => {
      expect(save).toBeEnabled();
    });

    fireEvent.change(screen.getByLabelText(/Customer \/ Member Name/), { target: { value: "" } });
    expect(save).toBeDisabled();
  });

  it("refuses to edit a closed FD", async () => {
    vi.mocked(GetFD).mockResolvedValue({ fd: { ...activeFd, status: "CLOSED" }, history: [] } as never);

    renderEdit();

    expect(await screen.findByText("This FD cannot be edited")).toBeInTheDocument();
    expect(screen.queryByRole("button", { name: "Save Changes" })).not.toBeInTheDocument();
    expect(EditFD).not.toHaveBeenCalled();
  });
});
