import { fireEvent, render, screen } from "@testing-library/react";
import { MemoryRouter } from "react-router-dom";
import { describe, expect, it, vi } from "vitest";

vi.mock("../../wailsjs/go/main/App", () => ({
  CreateFD: vi.fn(),
  PreviewFD: vi.fn(),
}));

vi.mock("../components/member/MemberPicker", () => ({
  MemberPicker: () => <div data-testid="member-picker" />,
}));

import { todayISO } from "../lib/format";
import { tenurePresetDays } from "../lib/tenure";
import { NewFDPage } from "./NewFDPage";

describe("NewFDPage tenure presets", () => {
  it("defaults to a calendar 3-year tenure and keeps presets in sync", () => {
    render(
      <MemoryRouter>
        <NewFDPage />
      </MemoryRouter>,
    );

    const start = screen.getByLabelText(/^Start Date/);
    const tenure = screen.getByLabelText(/^Tenure \(days\)/);
    expect(tenure).toHaveValue(tenurePresetDays(todayISO(), 36));

    fireEvent.change(start, { target: { value: "2028-01-01" } });
    // 01-01-2028 + 3 years = 01-01-2031 = 1,096 days (2028 is a leap year).
    expect(tenure).toHaveValue(1096);

    fireEvent.click(screen.getByRole("button", { name: "1 year" }));
    // 01-01-2028 + 1 year = 01-01-2029 = 366 days.
    expect(tenure).toHaveValue(366);
  });
});
