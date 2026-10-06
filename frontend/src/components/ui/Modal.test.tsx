import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { describe, expect, it, vi } from "vitest";
import { ConfirmDialog, Modal } from "./Modal";

describe("Modal", () => {
  it("renders title and body when open", () => {
    render(
      <Modal open title="Delete record" onClose={() => {}}>
        <p>Body text</p>
      </Modal>,
    );
    expect(screen.getByRole("dialog", { name: "Delete record" })).toBeInTheDocument();
    expect(screen.getByText("Body text")).toBeInTheDocument();
  });

  it("renders nothing when closed", () => {
    render(
      <Modal open={false} title="Hidden" onClose={() => {}}>
        <p>Body text</p>
      </Modal>,
    );
    expect(screen.queryByRole("dialog")).not.toBeInTheDocument();
  });

  it("closes on Escape", async () => {
    const onClose = vi.fn();
    render(
      <Modal open title="Escapable" onClose={onClose}>
        <p>Body</p>
      </Modal>,
    );
    await userEvent.keyboard("{Escape}");
    expect(onClose).toHaveBeenCalled();
  });
});

describe("ConfirmDialog", () => {
  it("asks for confirmation and reports the decision", async () => {
    const onConfirm = vi.fn();
    const onCancel = vi.fn();
    render(
      <ConfirmDialog
        open
        title="Confirm closure"
        message="Close this FD?"
        confirmLabel="Confirm Closure"
        onConfirm={onConfirm}
        onCancel={onCancel}
      />,
    );

    await userEvent.click(screen.getByRole("button", { name: "Cancel" }));
    expect(onCancel).toHaveBeenCalled();
    expect(onConfirm).not.toHaveBeenCalled();

    await userEvent.click(screen.getByRole("button", { name: "Confirm Closure" }));
    expect(onConfirm).toHaveBeenCalled();
  });
});
