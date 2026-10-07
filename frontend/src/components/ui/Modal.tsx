import { useEffect, type ReactNode } from "react";
import { useEscapeKey } from "../../lib/hooks";
import { Button } from "./Button";
import { IconButton } from "./Button";

export interface ModalProps {
  open: boolean;
  title: ReactNode;
  onClose: () => void;
  children: ReactNode;
  footer?: ReactNode;
  wide?: boolean;
}

/** Dialog with overlay, Escape support and focus on open (SRS §39). */
export function Modal({ open, title, onClose, children, footer, wide }: ModalProps) {
  useEscapeKey(open, onClose);

  useEffect(() => {
    if (!open) return;
    const previous = document.activeElement as HTMLElement | null;
    const timer = setTimeout(() => {
      document.querySelector<HTMLElement>(".hs-modal button, .hs-modal input")?.focus();
    }, 0);
    return () => {
      clearTimeout(timer);
      previous?.focus?.();
    };
  }, [open]);

  if (!open) return null;

  return (
    <div className="hs-modal-overlay" onMouseDown={(e) => e.target === e.currentTarget && onClose()}>
      <div className={`hs-modal ${wide ? "hs-modal--wide" : ""}`} role="dialog" aria-modal="true" aria-label={typeof title === "string" ? title : "Dialog"}>
        <header className="hs-modal__head">
          <h2 className="hs-modal__title">{title}</h2>
          <IconButton icon="close" label="Close dialog" onClick={onClose} />
        </header>
        <div className="hs-modal__body">{children}</div>
        {footer && <footer className="hs-modal__foot">{footer}</footer>}
      </div>
    </div>
  );
}

export interface ConfirmDialogProps {
  open: boolean;
  title: string;
  message: ReactNode;
  confirmLabel?: string;
  danger?: boolean;
  busy?: boolean;
  confirmDisabled?: boolean;
  onConfirm: () => void;
  onCancel: () => void;
}

/** Asks for confirmation before important actions (SRS §39). */
export function ConfirmDialog({
  open,
  title,
  message,
  confirmLabel = "Confirm",
  danger,
  busy,
  confirmDisabled,
  onConfirm,
  onCancel,
}: ConfirmDialogProps) {
  return (
    <Modal
      open={open}
      title={title}
      onClose={onCancel}
      footer={
        <>
          <Button onClick={onCancel} disabled={busy}>
            Cancel
          </Button>
          <Button variant={danger ? "danger" : "primary"} onClick={onConfirm} loading={busy} disabled={confirmDisabled}>
            {confirmLabel}
          </Button>
        </>
      }
    >
      {message}
    </Modal>
  );
}
