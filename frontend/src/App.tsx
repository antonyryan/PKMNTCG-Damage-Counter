import { Coins, Dice6, RotateCcw } from "lucide-react";
import { GameProvider, useGame } from "./context/GameContext";
import { PokemonSlot } from "./components/PokemonSlot";
import type { Side } from "./types";

// SideBoard renders one half of the table and optionally rotates it for the opposing player.
function SideBoard({ side, rotated }: { side: Side; rotated?: boolean }) {
  const { state } = useGame();
  const player = state[side];

  return (
    <section
      className={`relative grid h-1/2 grid-rows-[auto_1fr] gap-2 p-2 sm:p-3 ${rotated ? "rotate-180" : ""}`}
    >
      <div className="grid grid-cols-5 gap-2 rounded-2xl bg-white/30 p-2">
        <div className="col-start-3">
          <PokemonSlot side={side} zone="active" slot={player.active} />
        </div>
      </div>

      <div className="grid grid-cols-5 gap-2 rounded-2xl bg-white/30 p-2">
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

// Toolbox groups shared match actions that are not tied to a single board slot.
function Toolbox() {
  const { state, flipCoin, rollDie, toggleGX, toggleVSTAR, resetGame } =
    useGame();

  const markerClass = (used: boolean) =>
    `rounded-xl border px-2 py-2 text-xs font-black tracking-wide transition ${
      used
        ? "border-slate-400 bg-slate-300 text-slate-600"
        : "border-emerald-700 bg-emerald-500 text-white"
    }`;

  return (
    <aside className="pointer-events-none absolute inset-x-0 top-1/2 z-20 -translate-y-1/2 px-2">
      <div className="pointer-events-auto mx-auto flex max-w-[540px] flex-col gap-2 rounded-2xl border border-teal-900/15 bg-white/85 p-2 shadow-card backdrop-blur">
        <div className="grid grid-cols-2 gap-2">
          <button
            type="button"
            onClick={flipCoin}
            className="flex items-center justify-center gap-2 rounded-xl border border-teal-900/20 bg-board-panel px-3 py-3 text-sm font-semibold active:scale-95"
          >
            <Coins size={18} />
            {state.coinFlipping
              ? "Flipping..."
              : `Coin: ${state.coinResult ?? "-"}`}
          </button>
          <button
            type="button"
            onClick={rollDie}
            className="flex items-center justify-center gap-2 rounded-xl border border-teal-900/20 bg-board-panel px-3 py-3 text-sm font-semibold active:scale-95"
          >
            <Dice6 size={18} />
            Die: {state.dieResult ?? "-"}
          </button>
        </div>

        <div className="grid grid-cols-2 gap-2 text-center text-[10px] font-bold uppercase tracking-[0.18em] text-board-ink/60">
          <span>Opponent</span>
          <span>You</span>
        </div>

        <div className="grid grid-cols-2 gap-2">
          <button
            type="button"
            onClick={() => toggleGX("opponent")}
            className={markerClass(state.opponent.usedGX)}
          >
            GX
          </button>
          <button
            type="button"
            onClick={() => toggleGX("me")}
            className={markerClass(state.me.usedGX)}
          >
            GX
          </button>
        </div>

        <div className="grid grid-cols-2 gap-2">
          <button
            type="button"
            onClick={() => toggleVSTAR("opponent")}
            className={markerClass(state.opponent.usedVSTAR)}
          >
            VSTAR
          </button>
          <button
            type="button"
            onClick={() => toggleVSTAR("me")}
            className={markerClass(state.me.usedVSTAR)}
          >
            VSTAR
          </button>
        </div>

        <button
          type="button"
          onClick={resetGame}
          className="flex items-center justify-center gap-2 rounded-xl border border-rose-400 bg-rose-500 px-3 py-3 text-sm font-bold text-white active:scale-95"
        >
          <RotateCcw size={16} />
          New Game
        </button>
      </div>
    </aside>
  );
}

// Battlefield enforces the fixed two-half mobile layout used during a physical match.
function Battlefield() {
  return (
    <main className="relative h-dvh overflow-hidden bg-[radial-gradient(circle_at_20%_20%,rgba(226,113,29,0.25),transparent_40%),radial-gradient(circle_at_80%_75%,rgba(26,46,44,0.2),transparent_45%),linear-gradient(180deg,#f5f3ee_0%,#efe6d3_100%)] font-sans text-board-ink">
      <div className="pointer-events-none absolute inset-0 opacity-30 [background-image:linear-gradient(rgba(26,46,44,0.08)_1px,transparent_1px),linear-gradient(90deg,rgba(26,46,44,0.08)_1px,transparent_1px)] [background-size:18px_18px]" />
      <SideBoard side="opponent" rotated />
      <SideBoard side="me" />
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
