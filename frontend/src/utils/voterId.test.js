import { describe, it, expect, beforeEach } from "vitest";
import { getVoterId } from "./voterId";

describe("getVoterId", () => {
  beforeEach(() => {
    localStorage.clear();
  });

  it("creates and persists a voter id on first call", () => {
    const id = getVoterId();
    expect(id).toBeTruthy();
    expect(localStorage.getItem("livepoll_voter_id")).toBe(id);
  });

  it("returns the same id on subsequent calls", () => {
    const first = getVoterId();
    const second = getVoterId();
    expect(second).toBe(first);
  });
});
