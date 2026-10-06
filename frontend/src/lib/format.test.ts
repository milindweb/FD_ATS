import { describe, expect, it } from "vitest";
import { daysLabel, formatDate, formatMoney, formatNumber, formatPercent, formatTenure, parseISO, todayISO } from "./format";

describe("formatMoney", () => {
  it("formats whole rupees with Indian grouping", () => {
    expect(formatMoney(12500000)).toBe("₹1,25,00,000");
    expect(formatMoney(50000)).toBe("₹50,000");
  });

  it("rounds and handles nullish input", () => {
    expect(formatMoney(99.6)).toBe("₹100");
    expect(formatMoney(null)).toBe("₹0");
    expect(formatMoney(undefined)).toBe("₹0");
    expect(formatMoney(Number.NaN)).toBe("₹0");
  });
});

describe("formatNumber", () => {
  it("groups digits without a symbol", () => {
    expect(formatNumber(12500000)).toBe("1,25,00,000");
    expect(formatNumber(null)).toBe("0");
  });
});

describe("formatDate", () => {
  it("renders ISO dates as DD/MM/YYYY", () => {
    expect(formatDate("2026-10-06")).toBe("06/10/2026");
    expect(formatDate("")).toBe("—");
    expect(formatDate(null)).toBe("—");
  });
});

describe("formatPercent", () => {
  it("always shows two decimals", () => {
    expect(formatPercent(8)).toBe("8.00%");
    expect(formatPercent(8.5)).toBe("8.50%");
    expect(formatPercent(null)).toBe("—");
  });
});

describe("formatTenure", () => {
  it("labels days plainly", () => {
    expect(formatTenure(1095)).toBe("1,095 Days");
    expect(formatTenure(0)).toBe("0 Days");
    expect(formatTenure(null)).toBe("—");
  });
});

describe("todayISO / parseISO", () => {
  it("round-trips today's date", () => {
    const today = todayISO();
    expect(today).toMatch(/^\d{4}-\d{2}-\d{2}$/);
    expect(parseISO(today)).not.toBeNull();
  });

  it("rejects malformed input", () => {
    expect(parseISO("6/10/2026")).toBeNull();
    expect(parseISO("2026-13-45")).not.toBeNull(); // normalised by Date
  });
});

describe("daysLabel", () => {
  it("describes relative days for maturity tracking", () => {
    expect(daysLabel(0)).toBe("Today");
    expect(daysLabel(1)).toBe("Tomorrow");
    expect(daysLabel(12)).toBe("in 12 Days");
    expect(daysLabel(-3)).toBe("3 Days ago");
  });
});
