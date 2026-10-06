import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { beforeEach, describe, expect, it, vi, type Mock } from "vitest";

vi.mock("../lib/api", () => ({
  Login: vi.fn(),
  ResetCredentials: vi.fn(),
  SystemStatus: vi.fn().mockResolvedValue({ ready: true, error: "" }),
  errorMessage: (err: unknown) =>
    err instanceof Error ? err.message : String(err ?? "Something went wrong."),
}));

import { Login } from "../lib/api";
import { LoginPage } from "./LoginPage";

describe("LoginPage", () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it("signs in with valid credentials and calls onSuccess", async () => {
    (Login as Mock).mockResolvedValue(undefined);
    const onSuccess = vi.fn();
    render(<LoginPage onSuccess={onSuccess} />);

    expect(await screen.findByText("Sign in to continue")).toBeInTheDocument();
    await userEvent.type(screen.getByLabelText("Username"), "Admin");
    await userEvent.type(screen.getByLabelText("Password"), "0000");
    await userEvent.click(screen.getByRole("button", { name: "Sign in" }));

    await vi.waitFor(() => {
      expect(Login).toHaveBeenCalledWith({ username: "Admin", password: "0000" });
    });
    expect(onSuccess).toHaveBeenCalled();
  });

  it("shows the backend error when credentials are rejected", async () => {
    (Login as Mock).mockRejectedValue(new Error("Invalid username or password."));
    const onSuccess = vi.fn();
    render(<LoginPage onSuccess={onSuccess} />);

    await userEvent.type(await screen.findByLabelText("Username"), "Admin");
    await userEvent.type(screen.getByLabelText("Password"), "wrong");
    await userEvent.click(screen.getByRole("button", { name: "Sign in" }));

    expect(await screen.findByText("Invalid username or password.")).toBeInTheDocument();
    expect(onSuccess).not.toHaveBeenCalled();
  });

  it("opens the forgot-password form", async () => {
    render(<LoginPage onSuccess={vi.fn()} />);

    await userEvent.click(await screen.findByText("Forgot password?"));

    expect(await screen.findByText("Recovery code")).toBeInTheDocument();
    expect(screen.getByRole("button", { name: "Reset credentials" })).toBeInTheDocument();
    expect(screen.getByText("Back to sign in")).toBeInTheDocument();
  });
});
