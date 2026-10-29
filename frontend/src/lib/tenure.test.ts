import { describe, expect, it } from "vitest";
import { TENURE_PRESETS, activeTenurePreset, calendarTenureDays, tenurePresetDays } from "./tenure";

describe("calendarTenureDays", () => {
  it("lands on the anniversary date, not a fixed 1,095-day count", () => {
    expect(calendarTenureDays("2020-10-01", 36)).toBe(1095);
    expect(calendarTenureDays("2023-10-01", 36)).toBe(1096);
    expect(calendarTenureDays("2026-10-01", 36)).toBe(1096);
  });

  it("handles one and two year spans across leap years", () => {
    expect(calendarTenureDays("2023-10-01", 12)).toBe(366);
    expect(calendarTenureDays("2026-10-01", 12)).toBe(365);
    expect(calendarTenureDays("2023-10-01", 24)).toBe(731);
  });

  it("clamps month-end dates to the target month", () => {
    expect(calendarTenureDays("2024-01-31", 1)).toBe(29); // 29 Feb 2024
    expect(calendarTenureDays("2023-01-31", 1)).toBe(28); // 28 Feb 2023
  });

  it("returns 0 for invalid input", () => {
    expect(calendarTenureDays("", 36)).toBe(0);
    expect(calendarTenureDays("2023-13-01", 36)).toBe(0);
    expect(calendarTenureDays("2023-10-01", 0)).toBe(0);
  });
});

describe("tenurePresetDays", () => {
  it("uses the calendar day count for a valid start date", () => {
    expect(tenurePresetDays("2023-10-01", 36)).toBe(1096);
  });

  it("falls back to the old day count when the date is not usable yet", () => {
    expect(tenurePresetDays("", 36)).toBe(1095);
    expect(tenurePresetDays("", 6)).toBe(180);
  });
});

describe("activeTenurePreset", () => {
  it("matches the calendar day count for the current start date", () => {
    expect(activeTenurePreset("2023-10-01", "1096")).toEqual({ months: 36, label: "3 years" });
    expect(activeTenurePreset("2020-10-01", "1095")).toEqual({ months: 36, label: "3 years" });
  });

  it("returns nothing for a manually entered day count", () => {
    expect(activeTenurePreset("2023-10-01", "1095")).toBeUndefined();
    expect(activeTenurePreset("2023-10-01", "")).toBeUndefined();
  });
});

describe("TENURE_PRESETS", () => {
  it("offers the expected bank-style terms", () => {
    expect(TENURE_PRESETS.map((p) => p.label)).toEqual(["6 months", "1 year", "2 years", "3 years"]);
  });
});
