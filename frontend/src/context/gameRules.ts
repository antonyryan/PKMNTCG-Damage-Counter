import type { PlayerState, SlotState, SpecialStatus } from "../types";

const EXCLUSIVE_STATUSES: SpecialStatus[] = ["sleep", "confusion", "paralysis"];

// computeNextStatuses applies the TCG rule where sleep, confusion, and paralysis replace each other.
export function computeNextStatuses(
  current: SpecialStatus[],
  status: SpecialStatus,
): SpecialStatus[] {
  const has = current.includes(status);

  if (EXCLUSIVE_STATUSES.includes(status)) {
    const withoutExclusive = current.filter(
      (s) => !EXCLUSIVE_STATUSES.includes(s),
    );
    return has ? withoutExclusive : [...withoutExclusive, status];
  }

  return has ? current.filter((s) => s !== status) : [...current, status];
}

// computeDamage prevents the UI from ever storing negative damage values.
export function computeDamage(currentDamage: number, amount: number): number {
  return Math.max(0, currentDamage + amount);
}

// promoteBenchWithSwap moves a bench Pokemon to active and clears active-only statuses on both sides of the swap.
export function promoteBenchWithSwap(
  player: PlayerState,
  benchIndex: number,
): PlayerState {
  const fromBench = player.bench[benchIndex];
  if (!fromBench || !fromBench.pokemon) {
    return player;
  }

  const nextBench = [...player.bench] as PlayerState["bench"];
  nextBench[benchIndex] = {
    ...player.active,
    statuses: [],
  };

  return {
    ...player,
    active: {
      ...fromBench,
      statuses: [],
    },
    bench: nextBench,
  };
}

// emptySlot creates the canonical empty state used across resets, knockouts, and initialization.
export function emptySlot(): SlotState {
  return {
    pokemon: null,
    damage: 0,
    statuses: [],
  };
}
