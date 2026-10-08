import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { MemoryRouter } from "react-router-dom";
import { beforeEach, describe, expect, it, vi } from "vitest";

const navigate = vi.fn();

vi.mock("react-router-dom", async (importOriginal) => {
  const actual = await importOriginal<typeof import("react-router-dom")>();
  return { ...actual, useNavigate: () => navigate };
});

vi.mock("../../wailsjs/go/main/App", () => ({
  SystemStatus: vi.fn().mockResolvedValue({ ready: true, error: "", authed: true }),
  DashboardStats: vi.fn().mockResolvedValue({
    totalFds: 2,
    activeFds: 1,
    activePrincipal: 175000,
    totalInterest: 8000,
    fyLabel: "FY 2026-27",
    fyDeposits: 100000,
    maturingToday: 0,
    maturing7: 1,
    maturing30: 2,
    maturing90: 3,
  }),
  UpcomingMaturities: vi.fn().mockResolvedValue([
    {
      fdNumber: "FD-9",
      customerName: "Asha",
      principal: 50000,
      maturityDate: "2027-01-01",
      maturityAmount: 52000,
      daysRemaining: 87,
      status: "ACTIVE",
    },
  ]),
  MaturityChart: vi.fn().mockResolvedValue([
    { period: "2026-11", label: "Nov", count: 1, amount: 52000 },
    ...Array.from({ length: 11 }, (_, i) => ({
      period: `2026-${String(i + 12).padStart(2, "0")}`,
      label: "M",
      count: 0,
      amount: 0,
    })),
  ]),
}));

import { AppShell } from "../components/layout/AppShell";
import { DashboardPage } from "./DashboardPage";

function renderDashboard() {
  return render(
    <MemoryRouter initialEntries={["/"]}>
      <DashboardPage />
    </MemoryRouter>,
  );
}

describe("DashboardPage", () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it("shows the KPI row, upcoming maturities and the chart (SRS §9.1)", async () => {
    renderDashboard();

    expect(await screen.findByText("Active FDs")).toBeInTheDocument();
    expect(await screen.findByText("Total Interest")).toBeInTheDocument();
    expect(await screen.findByText("Financial Year")).toBeInTheDocument();
    expect(await screen.findByText("₹1,00,000")).toBeInTheDocument(); // FY deposits (big value)
    expect(await screen.findByText(/FY 2026-27/)).toBeInTheDocument(); // FY label (meta)
    expect(await screen.findByText("₹1,75,000")).toBeInTheDocument(); // total deposit
    expect(await screen.findByText("Upcoming Maturities")).toBeInTheDocument();
    expect(await screen.findByText("Maturities by Month")).toBeInTheDocument();
    // The upcoming table uses the SRS member wording.
    expect(await screen.findByText("Member Name")).toBeInTheDocument();
  });

  it("offers the quick actions under the quick search (SRS §9.1)", async () => {
    renderDashboard();

    expect(await screen.findByText("New Member")).toBeInTheDocument();
    expect(screen.getByText("Create New FD")).toBeInTheDocument();
    expect(screen.getByText("View full FD list")).toBeInTheDocument();
    expect(screen.getByLabelText("Quick search")).toBeInTheDocument();
  });

  it("hands the quick-search query to the FD Master page on Enter", async () => {
    renderDashboard();

    const search = await screen.findByLabelText("Quick search");
    await userEvent.type(search, "asha{Enter}");

    expect(navigate).toHaveBeenCalledWith("/fds?q=asha");
  });

  it("opens the FD Master list when Enter is pressed with an empty query", async () => {
    renderDashboard();

    const search = await screen.findByLabelText("Quick search");
    await userEvent.type(search, "{Enter}");

    expect(navigate).toHaveBeenCalledWith("/fds");
  });
});

describe("AppShell", () => {
  it("renders header, main and footer around the page", async () => {
    render(
      <MemoryRouter>
        <AppShell>
          <div>Page content</div>
        </AppShell>
      </MemoryRouter>,
    );

    expect(screen.getByRole("banner")).toBeInTheDocument();
    expect(screen.getByRole("contentinfo")).toBeInTheDocument();
    expect(await screen.findByText("Page content")).toBeInTheDocument();
  });
});
