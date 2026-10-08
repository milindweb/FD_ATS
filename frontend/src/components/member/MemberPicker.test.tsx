import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { beforeEach, describe, expect, it, vi } from "vitest";

vi.mock("../../../wailsjs/go/main/App", () => ({
  ListMembers: vi.fn().mockResolvedValue({
    items: [
      {
        id: 7,
        genNo: "GEN-007",
        name: "Ravi",
        dob: "",
        mobile: "9123456780",
        email: "",
        presentAddress: "",
        permanentAddress: "",
        employerName: "",
        department: "",
        designation: "Officer",
        tokenNo: "",
        nomineeName: "",
        nomineeRelationship: "",
        aadhaar: "",
        pan: "",
        bankName: "",
        accountNo: "",
        ifsc: "",
        profileRemarks: "",
        createdAt: "",
        updatedAt: "",
        activeFdCount: 0,
        activeFdAmount: 0,
      },
    ],
    total: 1,
    page: 1,
    pageSize: 8,
  }),
}));

import { ListMembers } from "../../../wailsjs/go/main/App";
import { MemberPicker } from "./MemberPicker";

function renderPicker(onSelect = vi.fn()) {
  const view = render(<MemberPicker selected={null} onSelect={onSelect} />);
  return { ...view, onSelect };
}

describe("MemberPicker", () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it("searches members and selects a result", async () => {
    const { onSelect } = renderPicker();
    const search = screen.getByLabelText(/Search member/);
    await userEvent.type(search, "ravi");

    const option = await screen.findByRole("option", { name: /GEN-007/ });
    await userEvent.click(option);

    expect(onSelect).toHaveBeenCalledWith({ id: 7, genNo: "GEN-007", name: "Ravi" });
    expect(ListMembers).toHaveBeenCalledWith({ search: "ravi", page: 1, pageSize: 8 });
    expect(screen.queryByRole("listbox")).not.toBeInTheDocument();
  });

  it("clears the chosen member with id 0", async () => {
    const onSelect = vi.fn();
    render(
      <MemberPicker selected={{ id: 7, genNo: "GEN-007", name: "Ravi" }} onSelect={onSelect} />,
    );

    await userEvent.click(screen.getByLabelText("Clear member selection"));

    expect(onSelect).toHaveBeenCalledWith({ id: 0, genNo: "", name: "" });
  });
});
