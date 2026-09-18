import { describe, it, expect } from "vitest";
import { percentage, expiresAtFromMinutes } from "../utils/formatters";

describe("percentage", () => {
  it("returns 0 when total is 0", () => {
    expect(percentage(5, 0)).toBe(0);
  });

  it("rounds to the nearest whole percent", () => {
    expect(percentage(1, 3)).toBe(33);
    expect(percentage(2, 3)).toBe(67);
  });

  it("handles 100%", () => {
    expect(percentage(10, 10)).toBe(100);
  });
});

describe("expiresAtFromMinutes", () => {
  it("returns null for no expiration", () => {
    expect(expiresAtFromMinutes(null)).toBeNull();
    expect(expiresAtFromMinutes(0)).toBeNull();
  });

  it("returns an ISO timestamp in the future for a given duration", () => {
    const before = Date.now();
    const iso = expiresAtFromMinutes(30);
    const diffMinutes = (new Date(iso).getTime() - before) / 60000;
    expect(diffMinutes).toBeGreaterThan(29);
    expect(diffMinutes).toBeLessThan(31);
  });
});
