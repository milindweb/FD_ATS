import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { MemoryRouter } from "react-router-dom";
import { beforeEach, describe, expect, it, vi } from "vitest";

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
  ListFDs: vi.fn().mockResolvedValue({
    items: [
      {
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
      },
    ],
    total: 1,
    page: 1,
    pageSize: 25,
  }),
  UpcomingMaturities: vi.fn().mockResolvedValue([
    {
      fdNumber: "FD-26-009",
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
  GetFD: vi.fn().mockResolvedValue({
    fd: {
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
    },
    history: [],
  }),
}));

vi.mock("./LoginPage", async () => {
  const { useEffect } = await import("react");
  return {
    // Auto-sign-in so route-level tests can exercise the gated app.
    LoginPage: ({ onSuccess }: { onSuccess: () => void }) => {
      useEffect(() => {
        onSuccess();
      }, []);
      return null;
    },
  };
});

import { GetFD, ListFDs } from "../../wailsjs/go/main/App";
import { AppShell } from "../components/layout/AppShell";
import { DashboardPage } from "./DashboardPage";
import App from "../App";

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

  it("shows KPIs and the FD list", async () => {
    renderDashboard();

    expect(await screen.findByText("Active FDs")).toBeInTheDocument();
    expect(await screen.findByText("Total Interest")).toBeInTheDocument();
    expect(await screen.findByText("Financial Year")).toBeInTheDocument();
    expect(await screen.findByText("₹1,00,000")).toBeInTheDocument(); // FY deposits (big value)
    expect(await screen.findByText(/FY 2026-27/)).toBeInTheDocument(); // FY label (meta)
    expect(await screen.findByText("₹1,75,000")).toBeInTheDocument(); // total deposit
    expect(await screen.findByText("FD-26-001")).toBeInTheDocument();
    expect(await screen.findByText("Upcoming Maturities")).toBeInTheDocument();
    expect(await screen.findByText("Maturities by Month")).toBeInTheDocument();
    expect(screen.queryByText("Maturing Soon")).not.toBeInTheDocument();
    expect(ListFDs).toHaveBeenCalledWith({ search: "", filter: "ALL", page: 1, pageSize: 25 });
  });

  it("passes the search term to the backend", async () => {
    renderDashboard();
    const search = await screen.findByLabelText("Search FDs");
    await userEvent.type(search, "asha");

    await vi.waitFor(() => {
      expect(ListFDs).toHaveBeenCalledWith({ search: "asha", filter: "ALL", page: 1, pageSize: 25 });
    });
  });

  it("navigates to the FD details screen when a row is clicked", async () => {
    render(<App />);

    await userEvent.click(await screen.findByText("FD-26-001"));

    expect(await screen.findByText("Deposit Details")).toBeInTheDocument();
    expect(GetFD).toHaveBeenCalledWith("FD-26-001");
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
