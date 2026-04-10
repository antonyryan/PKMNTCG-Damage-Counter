import { describe, expect, it } from "vitest";
import { buildApiUrl } from "./client";

describe("buildApiUrl", () => {
  it("omits empty query params and preserves populated ones", () => {
    expect(buildApiUrl("/api/pokemon/search", { q: "pik", limit: 20, empty: "" }))
      .toBe("/api/pokemon/search?q=pik&limit=20");
  });
});
