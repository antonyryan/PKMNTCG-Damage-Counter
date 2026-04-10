import {
  createContext,
  useCallback,
  useContext,
  useEffect,
  useMemo,
  useState,
  type PropsWithChildren,
} from "react";
import type {
  GameState,
  PlayerState,
  PokemonRef,
  Side,
  SlotState,
  SpecialStatus,
  Zone,
} from "../types";
import {
  computeDamage,
  computeNextStatuses,
  emptySlot,
  promoteBenchWithSwap,
} from "./gameRules";

const LOCAL_STORAGE_KEY = "pkmntcg-companion-state-v1";

// emptyPlayer builds the initial board for one side of the match.
const emptyPlayer = (): PlayerState => ({
  active: emptySlot(),
  bench: [emptySlot(), emptySlot(), emptySlot(), emptySlot(), emptySlot()],
  usedGX: false,
  usedVSTAR: false,
});

// initialState builds a full offline match snapshot.
const initialState = (): GameState => ({
  me: emptyPlayer(),
  opponent: emptyPlayer(),
  coinResult: null,
  dieResult: null,
  coinFlipping: false,
});

interface GameContextValue {
  state: GameState;
  setPokemon: (
    side: Side,
    zone: Zone,
    pokemon: PokemonRef,
    benchIndex?: number,
  ) => void;
  knockout: (side: Side, zone: Zone, benchIndex?: number) => void;
  adjustDamage: (
    side: Side,
    zone: Zone,
    amount: number,
    benchIndex?: number,
  ) => void;
  toggleStatus: (side: Side, status: SpecialStatus) => void;
  promoteBenchToActive: (side: Side, benchIndex: number) => void;
  toggleGX: (side: Side) => void;
  toggleVSTAR: (side: Side) => void;
  flipCoin: () => void;
  rollDie: () => void;
  resetGame: () => void;
}

const GameContext = createContext<GameContextValue | null>(null);

// loadState restores the last match from localStorage and falls back to a safe empty state.
function loadState(): GameState {
  const raw = localStorage.getItem(LOCAL_STORAGE_KEY);
  if (!raw) {
    return initialState();
  }

  try {
    const parsed = JSON.parse(raw) as GameState;
    if (!parsed.me || !parsed.opponent) {
      return initialState();
    }
    return parsed;
  } catch {
    return initialState();
  }
}

// updatePlayerSlot applies a slot-level mutation without duplicating active/bench branching logic.
function updatePlayerSlot(
  player: PlayerState,
  zone: Zone,
  updater: (slot: SlotState) => SlotState,
  benchIndex?: number,
): PlayerState {
  if (zone === "active") {
    return { ...player, active: updater(player.active) };
  }

  if (benchIndex === undefined || benchIndex < 0 || benchIndex > 4) {
    return player;
  }

  const nextBench = [...player.bench] as PlayerState["bench"];
  nextBench[benchIndex] = updater(player.bench[benchIndex]);
  return { ...player, bench: nextBench };
}

// GameProvider owns the match state, persistence, and all board mutations used by the UI.
export function GameProvider({ children }: PropsWithChildren) {
  const [state, setState] = useState<GameState>(loadState);

  useEffect(() => {
    // Persist every mutation immediately so reloading the page does not lose the current game.
    localStorage.setItem(LOCAL_STORAGE_KEY, JSON.stringify(state));
  }, [state]);

  const setPokemon = useCallback(
    (side: Side, zone: Zone, pokemon: PokemonRef, benchIndex?: number) => {
      setState((prev) => ({
        ...prev,
        [side]: updatePlayerSlot(
          prev[side],
          zone,
          (slot) => ({
            ...slot,
            pokemon,
            damage: slot.pokemon ? slot.damage : 0,
            statuses: zone === "active" ? slot.statuses : [],
          }),
          benchIndex,
        ),
      }));
    },
    [],
  );

  const knockout = useCallback(
    (side: Side, zone: Zone, benchIndex?: number) => {
      setState((prev) => ({
        ...prev,
        [side]: updatePlayerSlot(
          prev[side],
          zone,
          () => emptySlot(),
          benchIndex,
        ),
      }));
    },
    [],
  );

  const adjustDamage = useCallback(
    (side: Side, zone: Zone, amount: number, benchIndex?: number) => {
      setState((prev) => ({
        ...prev,
        [side]: updatePlayerSlot(
          prev[side],
          zone,
          (slot) => ({
            ...slot,
            damage: computeDamage(slot.damage, amount),
          }),
          benchIndex,
        ),
      }));
    },
    [],
  );

  const toggleStatus = useCallback((side: Side, status: SpecialStatus) => {
    setState((prev) => {
      const player = prev[side];
      const current = player.active.statuses;
      const nextStatuses = computeNextStatuses(current, status);

      return {
        ...prev,
        [side]: {
          ...player,
          active: {
            ...player.active,
            statuses: nextStatuses,
          },
        },
      };
    });
  }, []);

  const promoteBenchToActive = useCallback((side: Side, benchIndex: number) => {
    setState((prev) => {
      return {
        ...prev,
        [side]: promoteBenchWithSwap(prev[side], benchIndex),
      };
    });
  }, []);

  const toggleGX = useCallback((side: Side) => {
    setState((prev) => ({
      ...prev,
      [side]: {
        ...prev[side],
        usedGX: !prev[side].usedGX,
      },
    }));
  }, []);

  const toggleVSTAR = useCallback((side: Side) => {
    setState((prev) => ({
      ...prev,
      [side]: {
        ...prev[side],
        usedVSTAR: !prev[side].usedVSTAR,
      },
    }));
  }, []);

  const flipCoin = useCallback(() => {
    setState((prev) => ({ ...prev, coinFlipping: true }));

    let ticks = 0;
    // A short interval creates lightweight feedback before the final random result settles.
    const timer = window.setInterval(() => {
      ticks += 1;
      setState((prev) => ({
        ...prev,
        coinResult: Math.random() < 0.5 ? "Heads" : "Tails",
      }));

      if (ticks >= 8) {
        window.clearInterval(timer);
        setState((prev) => ({
          ...prev,
          coinFlipping: false,
          coinResult: Math.random() < 0.5 ? "Heads" : "Tails",
        }));
      }
    }, 75);
  }, []);

  const rollDie = useCallback(() => {
    setState((prev) => ({
      ...prev,
      dieResult: Math.floor(Math.random() * 6) + 1,
    }));
  }, []);

  const resetGame = useCallback(() => {
    localStorage.removeItem(LOCAL_STORAGE_KEY);
    setState(initialState());
  }, []);

  const value = useMemo<GameContextValue>(
    () => ({
      state,
      setPokemon,
      knockout,
      adjustDamage,
      toggleStatus,
      promoteBenchToActive,
      toggleGX,
      toggleVSTAR,
      flipCoin,
      rollDie,
      resetGame,
    }),
    [
      state,
      setPokemon,
      knockout,
      adjustDamage,
      toggleStatus,
      promoteBenchToActive,
      toggleGX,
      toggleVSTAR,
      flipCoin,
      rollDie,
      resetGame,
    ],
  );

  return <GameContext.Provider value={value}>{children}</GameContext.Provider>;
}

// useGame exposes the match store and guards against usage outside the provider tree.
export function useGame() {
  const context = useContext(GameContext);
  if (!context) {
    throw new Error("useGame must be used inside GameProvider");
  }
  return context;
}
