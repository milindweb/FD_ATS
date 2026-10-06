import { describe, expect, it } from "vitest";
import { statusLabel, statusTone } from "./statusTone";

describe("statusTone", () => {
  it("maps API status keys to visual tones", () => {
    expect(statusTone("ACTIVE")).toBe("success");
    expect(statusTone("CLOSED")).toBe("neutral");
    expect(statusTone("MATURED")).toBe("success");
    expect(statusTone("PREMATURE")).toBe("warning");
    expect(statusTone("RENEWED")).toBe("info");
  });

  it("defaults to neutral for unknown or empty keys", () => {
    expect(statusTone("SOMETHING")).toBe("neutral");
    expect(statusTone(null)).toBe("neutral");
    expect(statusTone(undefined)).toBe("neutral");
    expect(statusTone("")).toBe("neutral");
  });

  it("is case-insensitive", () => {
    expect(statusTone("active")).toBe("success");
    expect(statusTone("renewed")).toBe("info");
  });
});

describe("statusLabel", () => {
  it("humanises history event keys", () => {
    expect(statusLabel("OPEN")).toBe("Opened");
    expect(statusLabel("RENEW")).toBe("Renewed");
    expect(statusLabel("CLOSE")).toBe("Closed");
  });

  it("keeps status keys readable as-is", () => {
    expect(statusLabel("ACTIVE")).toBe("ACTIVE");
    expect(statusLabel("PREMATURE")).toBe("PREMATURE");
  });

  it("shows a dash for empty keys", () => {
    expect(statusLabel(null)).toBe("—");
  });
});
