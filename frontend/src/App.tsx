import { useEffect, useRef, useState } from "react";
import { Coins, Dice6, RotateCcw } from "lucide-react";
import { getEvolutionOptions } from "./api/client";
import { GameProvider, useGame } from "./context/GameContext";
import { PokemonSlot } from "./components/PokemonSlot";
import {
  Sheet,
  SheetClose,
  SheetContent,
  SheetHeader,
  SheetTitle,
  SheetTrigger,
} from "./components/ui/sheet";
import {
  Dialog,
  DialogClose,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
} from "./components/ui/dialog";
import type { PokemonEvolutionOption, Side } from "./types";

// SideBoard renders one half of the table and optionally rotates it for the opposing player.
function SideBoard({ side, rotated }: { side: Side; rotated?: boolean }) {
  const { state } = useGame();
  const player = state[side];

  return (
    <section
      className={`relative grid h-1/2 grid-rows-[auto_0.82fr] gap-2 p-2 sm:p-3 md:grid-rows-[auto_1fr] ${rotated ? "rotate-180" : ""}`}
    >
      <div className="grid grid-cols-5 gap-2 rounded-2xl bg-white/30 p-2">
        <div className="col-start-2 col-end-5 mt-14 md:col-start-3 md:col-end-4 md:mt-0">
          <PokemonSlot side={side} zone="active" slot={player.active} />
        </div>
      </div>

      <div className="grid grid-cols-5 items-start gap-1 rounded-2xl bg-white/30 p-1.5 md:items-stretch md:gap-2 md:p-2">
        {player.bench.map((slot, index) => (
          <PokemonSlot
            key={`${side}-bench-${index}`}
            side={side}
            zone="bench"
            slot={slot}
            benchIndex={index}
          />
        ))}
      </div>
    </section>
  );
}

function PlayerMiniMenu({
  side,
  rotated,
  anchor,
}: {
  side: Side;
  rotated?: boolean;
  anchor: "left" | "right";
}) {
  const {
    state,
    adjustTeamDamage,
    evolvePokemon,
    flipCoin,
    rollDie,
    toggleGX,
    toggleVSTAR,
    isSessionReady,
    sessionError,
  } = useGame();
  const player = state[side];
  const title = side === "me" ? "You" : "Opponent";
  const [evolveMobileOpen, setEvolveMobileOpen] = useState(false);
  const [evolveDesktopOpen, setEvolveDesktopOpen] = useState(false);
  const [query, setQuery] = useState("");
  const [loading, setLoading] = useState(false);
  const [evolveResults, setEvolveResults] = useState<PokemonEvolutionOption[]>(
    [],
  );
  const lastTeamDamageTouchAtRef = useRef(0);

  const canEvolveActive = Boolean(player.active.pokemon);
  const canAdjustTeamDamage =
    Boolean(player.active.pokemon) ||
    player.bench.some((slot) => Boolean(slot.pokemon));
  const isEvolveOpen = evolveMobileOpen || evolveDesktopOpen;
  const TEAM_DAMAGE_TOUCH_GUARD_MS = 450;
  const evolveSheetClass =
    "z-[121] min-h-[36vh] max-h-[58vh] overflow-y-auto rounded-t-3xl border-t border-teal-900/15 bg-board-panel p-4 pb-6 md:hidden";

  const applyTeamDamage = (amount: number) => {
    adjustTeamDamage(side, amount);
  };

  const handleTeamDamageTouch = (amount: number) => {
    lastTeamDamageTouchAtRef.current = Date.now();
    applyTeamDamage(amount);
  };

  const handleTeamDamageClick = (amount: number) => {
    if (Date.now() - lastTeamDamageTouchAtRef.current < TEAM_DAMAGE_TOUCH_GUARD_MS) {
      return;
    }
    applyTeamDamage(amount);
  };

  useEffect(() => {
    if (!isEvolveOpen) {
      setQuery("");
      setEvolveResults([]);
      setLoading(false);
      return;
    }

    if (!isSessionReady) {
      setEvolveResults([]);
      setLoading(false);
      return;
    }

    if (!player.active.pokemon) {
      setEvolveResults([]);
      return;
    }

    let cancelled = false;
    const timer = window.setTimeout(async () => {
      setLoading(true);
      try {
        const results = await getEvolutionOptions(
          player.active.pokemon!.id,
          query,
        );
        if (!cancelled) {
          setEvolveResults(results);
        }
      } catch {
        if (!cancelled) {
          setEvolveResults([]);
        }
      } finally {
        if (!cancelled) {
          setLoading(false);
        }
      }
    }, 120);

    return () => {
      cancelled = true;
      window.clearTimeout(timer);
    };
  }, [isEvolveOpen, player.active.pokemon, query]);

  const evolveSearchPanel = (
    <div className="space-y-2">
      <input
        value={query}
        onChange={(e) => setQuery(e.target.value)}
        placeholder="Search Pokemon"
        className="w-full rounded-xl border border-teal-900/20 bg-white px-3 py-2 text-sm outline-none focus:border-board-accent"
        autoFocus
        disabled={!isSessionReady}
      />
      <div className="max-h-52 overflow-y-auto rounded-xl border border-teal-900/15 bg-white">
        {!isSessionReady && sessionError && (
          <p className="p-2 text-xs text-rose-600">{sessionError}</p>
        )}
        {loading && <p className="p-2 text-xs text-slate-500">Searching...</p>}
        {!loading && isSessionReady && evolveResults.length === 0 && (
          <p className="p-2 text-xs text-slate-500">No results</p>
        )}
        {evolveResults.map((pokemon) => (
          <button
            key={`${side}-active-evolve-${pokemon.id}`}
            type="button"
            className="flex w-full items-center justify-between gap-2 border-b border-slate-100 px-3 py-2 text-left text-sm last:border-none active:bg-slate-100"
            onClick={() => {
              evolvePokemon(side, "active", {
                id: pokemon.id,
                name: pokemon.name,
              });
              setEvolveMobileOpen(false);
              setEvolveDesktopOpen(false);
            }}
          >
            <span>{pokemon.name}</span>
            <span
              className={`rounded-full border px-2 py-0.5 text-[10px] font-semibold ${
                pokemon.action === "Evolve"
                  ? "border-emerald-300 bg-emerald-50 text-emerald-800"
                  : "border-amber-300 bg-amber-50 text-amber-800"
              }`}
            >
              {pokemon.action}
            </span>
          </button>
        ))}
      </div>
    </div>
  );

  const markerClass = (used: boolean) =>
    `rounded-lg border px-2 py-1 text-[10px] font-black tracking-wide transition ${
      used
        ? "border-slate-400 bg-slate-300 text-slate-600"
        : "border-emerald-700 bg-emerald-500 text-white"
    }`;

  return (
    <aside
      className={`pointer-events-none absolute top-1/2 z-20 -translate-y-1/2 ${anchor === "left" ? "left-2" : "right-2"}`}
    >
      <div
        className={`pointer-events-auto flex w-[132px] flex-col gap-1 rounded-2xl border border-teal-900/20 bg-white/90 p-2 shadow-card backdrop-blur ${rotated ? "rotate-180" : ""}`}
      >
        <p className="text-center text-[10px] font-bold uppercase tracking-[0.2em] text-board-ink/60">
          {title}
        </p>

        <div className="grid grid-cols-2 gap-1">
          <button
            type="button"
            onClick={flipCoin}
            className="flex items-center justify-center gap-1 rounded-lg border border-teal-900/20 bg-board-panel px-2 py-1 text-[10px] font-semibold active:scale-95"
          >
            <Coins size={12} />
            {state.coinFlipping ? "..." : (state.coinResult ?? "-")}
          </button>
          <button
            type="button"
            onClick={rollDie}
            className="flex items-center justify-center gap-1 rounded-lg border border-teal-900/20 bg-board-panel px-2 py-1 text-[10px] font-semibold active:scale-95"
          >
            <Dice6 size={12} />
            {state.dieResult ?? "-"}
          </button>
        </div>

        <div className="grid grid-cols-2 gap-1">
          <button
            type="button"
            onClick={() => toggleGX(side)}
            className={markerClass(player.usedGX)}
          >
            GX
          </button>
          <button
            type="button"
            onClick={() => toggleVSTAR(side)}
            className={markerClass(player.usedVSTAR)}
          >
            VSTAR
          </button>
        </div>

        <Sheet open={evolveMobileOpen} onOpenChange={setEvolveMobileOpen}>
          <SheetTrigger asChild>
            <button
              type="button"
              disabled={!canEvolveActive || !isSessionReady}
              className="rounded-lg border border-teal-900/20 bg-white px-2 py-1 text-[10px] font-semibold text-board-ink active:scale-95 disabled:cursor-not-allowed disabled:opacity-45 md:hidden"
            >
              Evolve Active
            </button>
          </SheetTrigger>
          <SheetContent side="bottom" className={evolveSheetClass}>
            <SheetHeader className="mb-3 flex-row items-center justify-between">
              <SheetTitle className="text-sm font-bold text-board-ink">
                Evolve Active
              </SheetTitle>
              <SheetClose asChild>
                <button
                  type="button"
                  className="rounded-lg border border-teal-900/20 bg-white px-2 py-1 text-xs font-semibold"
                >
                  Close
                </button>
              </SheetClose>
            </SheetHeader>
            {evolveSearchPanel}
          </SheetContent>
        </Sheet>

        <Dialog open={evolveDesktopOpen} onOpenChange={setEvolveDesktopOpen}>
          <DialogTrigger asChild>
            <button
              type="button"
              disabled={!canEvolveActive || !isSessionReady}
              className="hidden rounded-lg border border-teal-900/20 bg-white px-2 py-1 text-[10px] font-semibold text-board-ink active:scale-95 disabled:cursor-not-allowed disabled:opacity-45 md:block"
            >
              Evolve Active
            </button>
          </DialogTrigger>
          <DialogContent className="hidden md:block">
            <DialogHeader className="mb-3 flex-row items-center justify-between">
              <DialogTitle className="text-sm font-bold text-board-ink">
                Evolve Active
              </DialogTitle>
              <DialogClose asChild>
                <button
                  type="button"
                  className="rounded-lg border border-teal-900/20 bg-white px-2 py-1 text-xs font-semibold"
                >
                  Close
                </button>
              </DialogClose>
            </DialogHeader>
            {evolveSearchPanel}
          </DialogContent>
        </Dialog>

        <div className="grid grid-cols-2 gap-1">
          <button
            type="button"
            onTouchEnd={(e) => {
              e.preventDefault();
              handleTeamDamageTouch(-10);
            }}
            onClick={() => handleTeamDamageClick(-10)}
            disabled={!isSessionReady || !canAdjustTeamDamage}
            className="rounded-lg border border-teal-900/20 bg-white px-2 py-1 text-[10px] font-semibold text-board-ink active:scale-95 disabled:cursor-not-allowed disabled:opacity-45"
          >
            Team -10
          </button>
          <button
            type="button"
            onTouchEnd={(e) => {
              e.preventDefault();
              handleTeamDamageTouch(10);
            }}
            onClick={() => handleTeamDamageClick(10)}
            disabled={!isSessionReady || !canAdjustTeamDamage}
            className="rounded-lg border border-teal-900/20 bg-white px-2 py-1 text-[10px] font-semibold text-board-ink active:scale-95 disabled:cursor-not-allowed disabled:opacity-45"
          >
            Team +10
          </button>
        </div>
      </div>
    </aside>
  );
}

function CenterMenus() {
  return (
    <>
      <PlayerMiniMenu side="opponent" rotated anchor="left" />
      <PlayerMiniMenu side="me" anchor="right" />
    </>
  );
}

// Toolbox groups shared match actions that are not tied to a single board slot.
function Toolbox() {
  const { resetGame } = useGame();

  return (
    <aside className="pointer-events-none absolute bottom-2 inset-x-0 z-20 flex justify-center px-2">
      <button
        type="button"
        onClick={resetGame}
        className="pointer-events-auto flex items-center justify-center gap-2 rounded-xl border border-rose-400 bg-rose-500 px-3 py-2 text-xs font-bold text-white active:scale-95"
      >
        <RotateCcw size={14} />
        New Game
      </button>
    </aside>
  );
}

// Battlefield enforces the fixed two-half mobile layout used during a physical match.
function Battlefield() {
  const { isSessionReady, retrySession, sessionError } = useGame();

  return (
    <main className="relative h-dvh overflow-hidden bg-[radial-gradient(circle_at_20%_20%,rgba(226,113,29,0.25),transparent_40%),radial-gradient(circle_at_80%_75%,rgba(26,46,44,0.2),transparent_45%),linear-gradient(180deg,#f5f3ee_0%,#efe6d3_100%)] font-sans text-board-ink">
      <div className="pointer-events-none absolute inset-0 opacity-30 [background-image:linear-gradient(rgba(26,46,44,0.08)_1px,transparent_1px),linear-gradient(90deg,rgba(26,46,44,0.08)_1px,transparent_1px)] [background-size:18px_18px]" />
      {!isSessionReady && (
        <div className="absolute inset-x-2 top-2 z-30 rounded-xl border border-amber-300 bg-amber-50 px-3 py-2 text-xs font-medium text-amber-900 shadow-card">
          {sessionError ?? "Connecting to backend session..."}
          {sessionError && (
            <button
              type="button"
              onClick={retrySession}
              className="ml-3 rounded-md border border-amber-400 bg-white px-2 py-1 text-[11px] font-semibold"
            >
              Retry
            </button>
          )}
        </div>
      )}
      <SideBoard side="opponent" rotated />
      <SideBoard side="me" />
      <CenterMenus />
      <Toolbox />
    </main>
  );
}

// App mounts the board inside the global match state provider.
export default function App() {
  return (
    <GameProvider>
      <Battlefield />
    </GameProvider>
  );
}
