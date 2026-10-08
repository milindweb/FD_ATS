import { fireEvent, render, screen } from "@testing-library/react";
import { MemoryRouter, Route, Routes } from "react-router-dom";
import { beforeEach, describe, expect, it, vi } from "vitest";

vi.mock("../../wailsjs/go/main/App", () => ({
  GetFD: vi.fn(),
  PreviewFD: vi.fn(),
  EditFD: vi.fn(),
}));

vi.mock("../components/member/MemberPicker", () => ({
  MemberPicker: (p: any) => (
    <div data-testid="member-picker">
      <button type="button" data-testid="member-pick" onClick={() => p.onSelect({ id: 1, genNo: "GEN-1", name: "Test Member" })}>
        Pick member
      </button>
      <button type="button" data-testid="member-clear" onClick={() => p.onSelect({ id: 0, genNo: "", name: "" })}>
        Clear member
      </button>
    </div>
  ),
}));

import { EditFD, GetFD, PreviewFD } from "../../wailsjs/go/main/App";
import { EditFDPage } from "./EditFDPage";

const activeFd = {
  fdNumber: "FD-26-001",
  memberId: 1,
  fdFormNo: "FORM-1042",
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

    expect(await screen.findByTestId("member-picker")).toBeInTheDocument();
    expect(screen.getByLabelText(/FD Form No/)).toHaveValue("FORM-1042");
    expect(screen.getByLabelText(/Deposit Amount/)).toHaveValue(50000);
    expect(screen.getByLabelText(/^Tenure \(days\)/)).toHaveValue(365);

    fireEvent.change(screen.getByLabelText(/Deposit Amount/), { target: { value: "75000" } });

    const save = await screen.findByRole("button", { name: "Save Changes" });
    await vi.waitFor(() => {
      expect(save).toBeEnabled();
    });

    await vi.waitFor(() => {
      expect(PreviewFD).toHaveBeenCalledWith(
        expect.objectContaining({ memberId: 1, fdFormNo: "FORM-1042" }),
      );
    });

    fireEvent.click(save);

    await vi.waitFor(() => {
      expect(EditFD).toHaveBeenCalledWith({
        fdNumber: "FD-26-001",
        memberId: 1,
        fdFormNo: "FORM-1042",
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

    fireEvent.click(screen.getByTestId("member-clear"));
    expect(save).toBeDisabled();
    expect(screen.getByText("Select a linked member before saving.")).toBeInTheDocument();
  });

  it("requires a member before saving an FD with no linked member", async () => {
    vi.mocked(GetFD).mockResolvedValue({ fd: { ...activeFd, memberId: undefined }, history: [] } as never);

    renderEdit();

    const save = await screen.findByRole("button", { name: "Save Changes" });
    expect(save).toBeDisabled();
    expect(screen.getByText("This FD is not linked to a member yet — select one to continue.")).toBeInTheDocument();
    expect(screen.getByText("Select a linked member before saving.")).toBeInTheDocument();
    expect(PreviewFD).not.toHaveBeenCalled();

    fireEvent.click(screen.getByTestId("member-pick"));

    await vi.waitFor(() => {
      expect(save).toBeEnabled();
    });
  });

  it("refuses to edit a closed FD", async () => {
    vi.mocked(GetFD).mockResolvedValue({ fd: { ...activeFd, status: "CLOSED" }, history: [] } as never);

    renderEdit();

    expect(await screen.findByText("This FD cannot be edited")).toBeInTheDocument();
    expect(screen.queryByRole("button", { name: "Save Changes" })).not.toBeInTheDocument();
    expect(EditFD).not.toHaveBeenCalled();
  });
});
