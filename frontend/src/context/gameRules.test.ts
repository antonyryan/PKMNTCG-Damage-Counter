import { describe, expect, it } from "vitest";
import type { PlayerState } from "../types";
import {
  computeDamage,
  computeNextStatuses,
  emptySlot,
  promoteBenchWithSwap,
} from "./gameRules";

function createPlayerState(): PlayerState {
  return {
    active: {
      pokemon: { id: 150, name: "Mewtwo" },
      damage: 30,
      statuses: ["poison", "sleep"],
    },
    bench: [
      {
        pokemon: { id: 25, name: "Pikachu" },
        damage: 10,
        statuses: ["burn"],
      },
      emptySlot(),
      emptySlot(),
      emptySlot(),
      emptySlot(),
    ],
    usedGX: false,
    usedVSTAR: false,
  };
}

describe("game rules", () => {
  it("enforces exclusivity for sleep/confusion/paralysis", () => {
    const current = ["poison", "sleep"] as const;
    const next = computeNextStatuses([...current], "confusion");

    expect(next).toContain("poison");
    expect(next).toContain("confusion");
    expect(next).not.toContain("sleep");
    expect(next).not.toContain("paralysis");
  });

  it("toggles independent statuses (poison and burn)", () => {
    const withPoison = computeNextStatuses([], "poison");
    const withBurn = computeNextStatuses(withPoison, "burn");
    const removePoison = computeNextStatuses(withBurn, "poison");

    expect(withBurn).toEqual(["poison", "burn"]);
    expect(removePoison).toEqual(["burn"]);
  });

  it("clamps damage at zero", () => {
    expect(computeDamage(20, -10)).toBe(10);
    expect(computeDamage(20, -30)).toBe(0);
    expect(computeDamage(0, 10)).toBe(10);
  });

  it("promotes bench pokemon to active with swap and clears statuses in both directions", () => {
    const player = createPlayerState();

    const next = promoteBenchWithSwap(player, 0);

    expect(next.active.pokemon?.name).toBe("Pikachu");
    expect(next.active.damage).toBe(10);
    expect(next.active.statuses).toEqual([]);

    expect(next.bench[0].pokemon?.name).toBe("Mewtwo");
    expect(next.bench[0].damage).toBe(30);
    expect(next.bench[0].statuses).toEqual([]);
  });

  it("does not change state when promoting an empty bench slot", () => {
    const player = createPlayerState();
    const next = promoteBenchWithSwap(player, 4);

    expect(next).toBe(player);
  });
});
