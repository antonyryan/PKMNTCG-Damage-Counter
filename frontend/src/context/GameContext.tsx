import {
  createContext,
  useCallback,
  useContext,
  useEffect,
  useMemo,
  useRef,
  useState,
  type PropsWithChildren,
} from "react";
import {
  applySessionAction,
  createOrLoadSession,
  type SessionActionRequest,
} from "../api/client";
import type {
  GameHistoryEntry,
  GameState,
  PlayerState,
  PokemonRef,
  Side,
  SpecialStatus,
  Zone,
} from "../types";

const SESSION_STORAGE_KEY = "pkmntcg-companion-session-id-v1";

// emptyPlayer builds the initial board for one side of the match.
const emptyPlayer = (): PlayerState => ({
  active: { pokemon: null, damage: 0, statuses: [] },
  bench: [
    { pokemon: null, damage: 0, statuses: [] },
    { pokemon: null, damage: 0, statuses: [] },
    { pokemon: null, damage: 0, statuses: [] },
    { pokemon: null, damage: 0, statuses: [] },
    { pokemon: null, damage: 0, statuses: [] },
  ],
  usedGX: false,
  usedVSTAR: false,
});

// initialState builds a safe placeholder snapshot until the backend session loads.
const initialState = (): GameState => ({
  me: emptyPlayer(),
  opponent: emptyPlayer(),
  coinResult: null,
  dieResult: null,
  coinFlipping: false,
});

interface GameContextValue {
  state: GameState;
  history: GameHistoryEntry[];
  isSessionReady: boolean;
  sessionError: string | null;
  setPokemon: (
    side: Side,
    zone: Zone,
    pokemon: PokemonRef,
    benchIndex?: number,
  ) => void;
  evolvePokemon: (
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
  retrySession: () => void;
}

const GameContext = createContext<GameContextValue | null>(null);

// GameProvider owns the match state, persistence, and all board mutations used by the UI.
export function GameProvider({ children }: PropsWithChildren) {
  const [state, setState] = useState<GameState>(initialState);
  const [history, setHistory] = useState<GameHistoryEntry[]>([]);
  const [sessionId, setSessionId] = useState<string | null>(null);
  const [isSessionReady, setIsSessionReady] = useState(false);
  const [sessionError, setSessionError] = useState<string | null>(null);
  const actionQueueRef = useRef(Promise.resolve());

  const bootstrapSession = useCallback(async () => {
    setIsSessionReady(false);
    setSessionError(null);

    try {
      const savedSessionId = localStorage.getItem(SESSION_STORAGE_KEY);
      const session = await createOrLoadSession(savedSessionId);
      localStorage.setItem(SESSION_STORAGE_KEY, session.sessionId);
      setSessionId(session.sessionId);
      setState(session.state);
      setHistory(session.history);
      setIsSessionReady(true);
    } catch (error) {
      console.error("Failed to initialize backend session", error);
      setSessionId(null);
      setState(initialState());
      setHistory([]);
      setSessionError("Could not connect to the backend session service.");
    }
  }, []);

  useEffect(() => {
    let cancelled = false;

    const run = async () => {
      await bootstrapSession();
      if (cancelled) {
        return;
      }
    };

    void run();

    return () => {
      cancelled = true;
    };
  }, [bootstrapSession]);

  const enqueueAction = useCallback(
    (request: SessionActionRequest, optimisticUpdate?: (prev: GameState) => GameState) => {
      if (!sessionId) {
        setSessionError("Session not ready yet. Try again in a moment.");
        return;
      }

      setSessionError(null);

      if (optimisticUpdate) {
        setState(optimisticUpdate);
      }

      actionQueueRef.current = actionQueueRef.current
        .catch(() => undefined)
        .then(async () => {
          const session = await applySessionAction(sessionId, request);
          setState(session.state);
          setHistory(session.history);
          setIsSessionReady(true);
        })
        .catch((error) => {
          console.error(`Failed to apply action ${request.type}`, error);
          setSessionError("Failed to sync the action with the backend.");
        });
    },
    [sessionId],
  );

  const setPokemon = useCallback(
    (side: Side, zone: Zone, pokemon: PokemonRef, benchIndex?: number) => {
      enqueueAction({
        type: "set-pokemon",
        side,
        zone,
        benchIndex,
        pokemonId: pokemon.id,
      });
    },
    [enqueueAction],
  );

  const evolvePokemon = useCallback(
    (side: Side, zone: Zone, pokemon: PokemonRef, benchIndex?: number) => {
      enqueueAction({
        type: "evolve-pokemon",
        side,
        zone,
        benchIndex,
        pokemonId: pokemon.id,
      });
    },
    [enqueueAction],
  );

  const knockout = useCallback(
    (side: Side, zone: Zone, benchIndex?: number) => {
      enqueueAction({ type: "knockout", side, zone, benchIndex });
    },
    [enqueueAction],
  );

  const adjustDamage = useCallback(
    (side: Side, zone: Zone, amount: number, benchIndex?: number) => {
      enqueueAction({ type: "adjust-damage", side, zone, amount, benchIndex });
    },
    [enqueueAction],
  );

  const toggleStatus = useCallback((side: Side, status: SpecialStatus) => {
    enqueueAction({ type: "toggle-status", side, status });
  }, [enqueueAction]);

  const promoteBenchToActive = useCallback((side: Side, benchIndex: number) => {
    enqueueAction({ type: "promote-bench", side, benchIndex });
  }, [enqueueAction]);

  const toggleGX = useCallback((side: Side) => {
    enqueueAction({ type: "toggle-gx", side });
  }, [enqueueAction]);

  const toggleVSTAR = useCallback((side: Side) => {
    enqueueAction({ type: "toggle-vstar", side });
  }, [enqueueAction]);

  const flipCoin = useCallback(() => {
    enqueueAction(
      { type: "flip-coin" },
      (prev) => ({
        ...prev,
        coinFlipping: true,
      }),
    );
  }, [enqueueAction]);

  const rollDie = useCallback(() => {
    enqueueAction({ type: "roll-die" });
  }, [enqueueAction]);

  const resetGame = useCallback(() => {
    enqueueAction({ type: "reset" });
  }, [enqueueAction]);

  const value = useMemo<GameContextValue>(
    () => ({
      state,
      history,
      isSessionReady,
      sessionError,
      setPokemon,
      evolvePokemon,
      knockout,
      adjustDamage,
      toggleStatus,
      promoteBenchToActive,
      toggleGX,
      toggleVSTAR,
      flipCoin,
      rollDie,
      resetGame,
      retrySession: () => {
        void bootstrapSession();
      },
    }),
    [
      state,
      history,
      isSessionReady,
      sessionError,
      setPokemon,
      evolvePokemon,
      knockout,
      adjustDamage,
      toggleStatus,
      promoteBenchToActive,
      toggleGX,
      toggleVSTAR,
      flipCoin,
      rollDie,
      resetGame,
      bootstrapSession,
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
