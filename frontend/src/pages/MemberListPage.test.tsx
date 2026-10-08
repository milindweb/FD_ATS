import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { MemoryRouter } from "react-router-dom";
import { beforeEach, describe, expect, it, vi } from "vitest";

vi.mock("../../wailsjs/go/main/App", () => ({
  SystemStatus: vi.fn().mockResolvedValue({ ready: true, error: "", authed: true }),
  ListMembers: vi.fn().mockResolvedValue({
    items: [
      {
        id: 1,
        genNo: "GEN-001",
        name: "Asha",
        dob: "1990-05-10",
        mobile: "9876543210",
        email: "",
        presentAddress: "",
        permanentAddress: "",
        employerName: "",
        department: "",
        designation: "Clerk",
        tokenNo: "125",
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
        activeFdCount: 2,
        activeFdAmount: 150000,
      },
    ],
    total: 1,
    page: 1,
    pageSize: 25,
  }),
  PickMemberImportPath: vi.fn().mockResolvedValue(""),
  PickMemberTemplatePath: vi.fn().mockResolvedValue(""),
  WriteMemberTemplate: vi.fn().mockResolvedValue(undefined),
  PreviewMemberImport: vi.fn().mockResolvedValue({
    total: 0,
    valid: 0,
    duplicateGen: 0,
    missingMandatory: 0,
    invalid: 0,
    existing: 0,
    rows: [],
  }),
  CommitMemberImport: vi.fn().mockResolvedValue({ imported: 0, skipped: 0, total: 0 }),
}));

import { ListMembers, PickMemberTemplatePath, WriteMemberTemplate } from "../../wailsjs/go/main/App";
import { MemberListPage } from "./MemberListPage";

function renderMembers() {
  return render(
    <MemoryRouter initialEntries={["/member"]}>
      <MemberListPage />
    </MemoryRouter>,
  );
}

describe("MemberListPage", () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it("shows the member list from the backend", async () => {
    renderMembers();

    expect(await screen.findByText("GEN-001")).toBeInTheDocument();
    expect(screen.getByText("Asha")).toBeInTheDocument();
    expect(screen.getByText("Clerk")).toBeInTheDocument();
    expect(screen.getByText("₹1,50,000")).toBeInTheDocument();
    expect(ListMembers).toHaveBeenCalledWith({ search: "", page: 1, pageSize: 25 });
  });

  it("shows the empty state when there are no members", async () => {
    vi.mocked(ListMembers).mockResolvedValue({
      items: [],
      total: 0,
      page: 1,
      pageSize: 25,
    } as unknown as Awaited<ReturnType<typeof ListMembers>>);
    renderMembers();

    expect(await screen.findByText("No members yet")).toBeInTheDocument();
  });

  it("passes the search term to the backend", async () => {
    renderMembers();
    const search = await screen.findByLabelText("Search members");
    await userEvent.type(search, "asha");

    await vi.waitFor(() => {
      expect(ListMembers).toHaveBeenCalledWith({ search: "asha", page: 1, pageSize: 25 });
    });
  });

  it("shows the upload guide and downloads the sample template (SRS §52.5)", async () => {
    renderMembers();
    await userEvent.click(await screen.findByText("Bulk Excel Upload"));

    expect(await screen.findByText("Mandatory columns:")).toBeInTheDocument();
    expect(screen.getByText(/Download sample template/)).toBeInTheDocument();

    vi.mocked(PickMemberTemplatePath).mockResolvedValue("C:\\tmp\\Member_Import_Template.xlsx");
    await userEvent.click(screen.getByText("Download sample template"));

    await vi.waitFor(() => {
      expect(WriteMemberTemplate).toHaveBeenCalledWith("C:\\tmp\\Member_Import_Template.xlsx");
    });
    expect(await screen.findByText(/Member_Import_Template\.xlsx/)).toBeInTheDocument();
  });

  it("does not write a template when the save dialog is cancelled", async () => {
    renderMembers();
    await userEvent.click(await screen.findByText("Bulk Excel Upload"));

    vi.mocked(PickMemberTemplatePath).mockResolvedValue("");
    await userEvent.click(await screen.findByText("Download sample template"));

    await vi.waitFor(() => {
      expect(PickMemberTemplatePath).toHaveBeenCalled();
    });
    expect(WriteMemberTemplate).not.toHaveBeenCalled();
    expect(screen.queryByText(/Sample template saved/)).not.toBeInTheDocument();
  });
});
