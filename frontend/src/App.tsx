import { useEffect, useState } from "react";
import { Coins, Dice6, RotateCcw } from "lucide-react";
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
import type { PokemonSearchResult, Side } from "./types";

// SideBoard renders one half of the table and optionally rotates it for the opposing player.
function SideBoard({ side, rotated }: { side: Side; rotated?: boolean }) {
  const { state } = useGame();
  const player = state[side];

  return (
    <section
      className={`relative grid h-1/2 grid-rows-[auto_1fr] gap-2 p-2 sm:p-3 ${rotated ? "rotate-180" : ""}`}
    >
      <div className="grid grid-cols-5 gap-2 rounded-2xl bg-white/30 p-2">
        <div className="col-start-2 col-end-5 md:col-start-3 md:col-end-4">
          <PokemonSlot side={side} zone="active" slot={player.active} />
        </div>
      </div>

      <div className="grid grid-cols-5 gap-1 rounded-2xl bg-white/30 p-2 md:gap-2">
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

function PlayerMiniMenu({ side, rotated, anchor }: { side: Side; rotated?: boolean; anchor: "left" | "right" }) {
  const { state, flipCoin, rollDie, toggleGX, toggleVSTAR, setPokemon } =
    useGame();
  const player = state[side];
  const title = side === "me" ? "You" : "Opponent";
  const [evolveMobileOpen, setEvolveMobileOpen] = useState(false);
  const [evolveDesktopOpen, setEvolveDesktopOpen] = useState(false);
  const [query, setQuery] = useState("");
  const [loading, setLoading] = useState(false);
  const [results, setResults] = useState<PokemonSearchResult[]>([]);

  const canEvolveActive = Boolean(player.active.pokemon);
  const isEvolveOpen = evolveMobileOpen || evolveDesktopOpen;
  const evolveSheetClass =
    "z-[121] min-h-[36vh] max-h-[58vh] overflow-y-auto rounded-t-3xl border-t border-teal-900/15 bg-board-panel p-4 pb-6 md:hidden";

  useEffect(() => {
    if (!isEvolveOpen) {
      setQuery("");
      setResults([]);
      return;
    }

    const control = new AbortController();
    const timer = window.setTimeout(async () => {
      setLoading(true);
      try {
        const url = `/api/pokemon/search?q=${encodeURIComponent(query)}`;
        const response = await fetch(url, { signal: control.signal });
        if (!response.ok) {
          setResults([]);
          return;
        }
        const data = (await response.json()) as PokemonSearchResult[];
        setResults(data);
      } catch {
        setResults([]);
      } finally {
        setLoading(false);
      }
    }, 180);

    return () => {
      control.abort();
      window.clearTimeout(timer);
    };
  }, [query, isEvolveOpen]);

  const evolveSearchPanel = (
    <div className="space-y-2">
      <input
        value={query}
        onChange={(e) => setQuery(e.target.value)}
        placeholder="Search Pokemon"
        className="w-full rounded-xl border border-teal-900/20 bg-white px-3 py-2 text-sm outline-none focus:border-board-accent"
        autoFocus
      />
      <div className="max-h-52 overflow-y-auto rounded-xl border border-teal-900/15 bg-white">
        {loading && <p className="p-2 text-xs text-slate-500">Searching...</p>}
        {!loading && results.length === 0 && (
          <p className="p-2 text-xs text-slate-500">No results</p>
        )}
        {!loading &&
          results.map((pokemon) => (
            <button
              key={`${side}-active-evolve-${pokemon.id}`}
              type="button"
              className="block w-full border-b border-slate-100 px-3 py-2 text-left text-sm last:border-none active:bg-slate-100"
              onClick={() => {
                setPokemon(side, "active", { id: pokemon.id, name: pokemon.name });
                setEvolveMobileOpen(false);
                setEvolveDesktopOpen(false);
              }}
            >
              {pokemon.name}
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
            {state.coinFlipping ? "..." : state.coinResult ?? "-"}
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
              disabled={!canEvolveActive}
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
              disabled={!canEvolveActive}
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
  return (
    <main className="relative h-dvh overflow-hidden bg-[radial-gradient(circle_at_20%_20%,rgba(226,113,29,0.25),transparent_40%),radial-gradient(circle_at_80%_75%,rgba(26,46,44,0.2),transparent_45%),linear-gradient(180deg,#f5f3ee_0%,#efe6d3_100%)] font-sans text-board-ink">
      <div className="pointer-events-none absolute inset-0 opacity-30 [background-image:linear-gradient(rgba(26,46,44,0.08)_1px,transparent_1px),linear-gradient(90deg,rgba(26,46,44,0.08)_1px,transparent_1px)] [background-size:18px_18px]" />
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
