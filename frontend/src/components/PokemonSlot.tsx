import { useEffect, useMemo, useRef, useState, type ReactElement } from "react";
import {
  ArrowUpCircle,
  Flame,
  Moon,
  Plus,
  Skull,
  Sparkles,
  Waves,
} from "lucide-react";
import {
  Sheet,
  SheetClose,
  SheetContent,
  SheetHeader,
  SheetTitle,
  SheetTrigger,
} from "./ui/sheet";
import {
  Dialog,
  DialogClose,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
} from "./ui/dialog";
import {
  getEvolutionCandidates,
  searchPokemonCatalog,
} from "../data/pokemonEvolution";
import { useGame } from "../context/GameContext";
import type {
  PokemonSearchResult,
  Side,
  SlotState,
  SpecialStatus,
  Zone,
} from "../types";

interface PokemonSlotProps {
  side: Side;
  zone: Zone;
  slot: SlotState;
  benchIndex?: number;
}

// STATUS_BUTTONS keeps the active-slot controls declarative and easy to extend.
const STATUS_BUTTONS: {
  status: SpecialStatus;
  label: string;
  icon: ReactElement;
}[] = [
  { status: "sleep", label: "Sleep", icon: <Moon size={16} /> },
  { status: "confusion", label: "Conf", icon: <Sparkles size={16} /> },
  { status: "paralysis", label: "Para", icon: <Waves size={16} /> },
  { status: "poison", label: "Poison", icon: <Skull size={16} /> },
  { status: "burn", label: "Burn", icon: <Flame size={16} /> },
];

// PokemonSlot renders both empty and occupied board slots for active and bench positions.
export function PokemonSlot({
  side,
  zone,
  slot,
  benchIndex,
}: PokemonSlotProps) {
  const {
    setPokemon,
    knockout,
    adjustDamage,
    toggleStatus,
    promoteBenchToActive,
  } = useGame();
  const [addingMobileOpen, setAddingMobileOpen] = useState(false);
  const [addingDesktopOpen, setAddingDesktopOpen] = useState(false);
  const [evolvingMobileOpen, setEvolvingMobileOpen] = useState(false);
  const [evolvingDesktopOpen, setEvolvingDesktopOpen] = useState(false);
  const [menuOpen, setMenuOpen] = useState(false);
  const [query, setQuery] = useState("");
  const [addSearchResults, setAddSearchResults] = useState<PokemonSearchResult[]>([]);
  const lastMobileDamageTapAtRef = useRef(0);
  const isAdding =
    addingMobileOpen ||
    addingDesktopOpen ||
    evolvingMobileOpen ||
    evolvingDesktopOpen;
  const isEvolutionSearchOpen = evolvingMobileOpen || evolvingDesktopOpen;

  useEffect(() => {
    if (zone !== "bench" || !slot.pokemon) {
      setMenuOpen(false);
    }
  }, [zone, slot.pokemon]);

  useEffect(() => {
    if (!isAdding) {
      setQuery("");
      setAddSearchResults([]);
      return;
    }

    if (!addingMobileOpen && !addingDesktopOpen) {
      return;
    }

    // Keep a short debounce for smoother typing while using local dataset.
    const timer = window.setTimeout(() => {
      setAddSearchResults(searchPokemonCatalog(query));
    }, 100);

    return () => window.clearTimeout(timer);
  }, [query, isAdding]);

  const evolutionResults = useMemo(
    () => getEvolutionCandidates(slot.pokemon?.id ?? -1, query),
    [slot.pokemon?.id, query],
  );

  const chipClass = (active: boolean) =>
    `inline-flex items-center gap-1 rounded-full border px-2 py-1 text-[10px] font-semibold ${
      active
        ? "border-board-accent bg-board-accent text-white"
        : "border-teal-900/20 bg-white/70 text-board-ink"
    }`;

  const title = useMemo(() => {
    if (!slot.pokemon) {
      return zone === "active" ? "Active" : "Bench";
    }
    return slot.pokemon.name;
  }, [slot.pokemon, zone]);

  const benchMenuSheetSide = side === "opponent" ? "top" : "bottom";
  const benchMenuSheetClass =
    side === "opponent"
      ? "z-[121] min-h-[36vh] max-h-[58vh] overflow-y-auto rounded-b-3xl border-b border-teal-900/15 bg-board-panel p-4 pb-6 md:hidden"
      : "z-[121] min-h-[36vh] max-h-[58vh] overflow-y-auto rounded-t-3xl border-t border-teal-900/15 bg-board-panel p-4 pb-6 md:hidden";

  const iconToTopClass = side === "opponent" ? "rotate-180" : "";
  const opponentSheetOrientationClass = side === "opponent" ? "rotate-180" : "";
  const isBench = zone === "bench";

  const adjustMobileDamage = (amount: number) => {
    const now = performance.now();
    if (now - lastMobileDamageTapAtRef.current < 140) {
      return;
    }
    lastMobileDamageTapAtRef.current = now;
    adjustDamage(side, zone, amount, benchIndex);
  };

  const searchPanel = (
    <div className="space-y-2">
      <input
        value={query}
        onChange={(e) => setQuery(e.target.value)}
        placeholder="Search Pokemon"
        className="w-full rounded-xl border border-teal-900/20 bg-white px-3 py-2 text-sm outline-none focus:border-board-accent"
        autoFocus
      />
      <div className="max-h-52 overflow-y-auto rounded-xl border border-teal-900/15 bg-white">
        {(isEvolutionSearchOpen
          ? evolutionResults.length === 0
          : addSearchResults.length === 0) && (
          <p className="p-2 text-xs text-slate-500">No results</p>
        )}
        {(isEvolutionSearchOpen ? evolutionResults : addSearchResults).map((pokemon) => (
            <button
              key={pokemon.id}
              type="button"
              className="block w-full border-b border-slate-100 px-3 py-2 text-left text-sm last:border-none active:bg-slate-100"
              onClick={() => {
                setPokemon(
                  side,
                  zone,
                  { id: pokemon.id, name: pokemon.name },
                  benchIndex,
                );
                setAddingMobileOpen(false);
                setAddingDesktopOpen(false);
                setEvolvingMobileOpen(false);
                setEvolvingDesktopOpen(false);
              }}
            >
              {pokemon.name}
            </button>
          ))}
      </div>
    </div>
  );

  if (!slot.pokemon) {
    return (
      <div className="relative rounded-2xl border border-dashed border-teal-900/25 bg-white/40 p-2">
        <Sheet open={addingMobileOpen} onOpenChange={setAddingMobileOpen}>
          <SheetTrigger asChild>
            <button
              type="button"
              className="flex h-full min-h-20 w-full items-center justify-center rounded-xl bg-white/70 py-5 text-board-ink active:scale-95 md:hidden"
            >
              <Plus size={32} />
            </button>
          </SheetTrigger>

          <SheetContent
            side="bottom"
            className="z-[121] min-h-[36vh] max-h-[58vh] overflow-y-auto rounded-t-3xl border-t border-teal-900/15 bg-board-panel p-4 pb-6 md:hidden"
          >
            <SheetHeader className="mb-3 flex-row items-center justify-between">
              <SheetTitle className="text-sm font-bold text-board-ink">
                Add Pokemon
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

            {searchPanel}
          </SheetContent>
        </Sheet>

        <Dialog open={addingDesktopOpen} onOpenChange={setAddingDesktopOpen}>
          <DialogTrigger asChild>
            <button
              type="button"
              className="hidden h-full min-h-20 w-full items-center justify-center rounded-xl bg-white/70 py-5 text-board-ink active:scale-95 md:flex"
            >
              <Plus size={32} />
            </button>
          </DialogTrigger>
          <DialogContent className="hidden md:block">
            <DialogHeader className="mb-3 flex-row items-center justify-between">
              <DialogTitle className="text-sm font-bold text-board-ink">
                Add Pokemon
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
            {searchPanel}
          </DialogContent>
        </Dialog>
      </div>
    );
  }

  return (
    <div className="relative rounded-2xl border border-teal-900/20 bg-white/80 p-2 shadow-card backdrop-blur-sm">
      {isBench && (
        <button
          type="button"
          onClick={() => setMenuOpen(true)}
          aria-label={`Open bench menu for ${title}`}
          className="absolute inset-0 z-10 rounded-2xl md:hidden"
        />
      )}

      {isBench && (
        <Sheet open={menuOpen} onOpenChange={setMenuOpen}>
          <SheetContent side={benchMenuSheetSide} className={benchMenuSheetClass}>
            <div className={opponentSheetOrientationClass}>
              <SheetHeader className="mb-3 flex-row items-center justify-between">
                <SheetTitle className="truncate text-sm font-bold text-board-ink">
                  {title}
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

              <div className="mb-3 flex items-center justify-center gap-2">
                <button
                  type="button"
                  className="rounded-xl border border-teal-900/20 bg-white px-3 py-2 text-base font-bold active:scale-95"
                  onClick={() => adjustMobileDamage(-10)}
                >
                  -10
                </button>
                <div className="min-w-16 rounded-xl bg-board-ink px-4 py-2 text-center text-base font-black text-white">
                  {slot.damage}
                </div>
                <button
                  type="button"
                  className="rounded-xl border border-teal-900/20 bg-white px-3 py-2 text-base font-bold active:scale-95"
                  onClick={() => adjustMobileDamage(10)}
                >
                  +10
                </button>
              </div>

              <div className="grid grid-cols-2 gap-2">
                <SheetClose asChild>
                  <button
                    type="button"
                    onClick={() => promoteBenchToActive(side, benchIndex ?? 0)}
                    className="flex items-center justify-center gap-1 rounded-xl border border-teal-900/20 bg-board-accentSoft px-2 py-3 text-xs font-semibold text-board-ink active:scale-95"
                  >
                    <ArrowUpCircle size={15} className={iconToTopClass} />
                    Promote
                  </button>
                </SheetClose>

                <SheetClose asChild>
                  <button
                    type="button"
                    onClick={() => knockout(side, zone, benchIndex)}
                    className="flex items-center justify-center gap-1 rounded-xl border border-rose-300 bg-rose-100 px-2 py-3 text-xs font-semibold text-rose-900 active:scale-95"
                  >
                    <Skull size={15} />
                    Knock Out
                  </button>
                </SheetClose>
              </div>

              <div className="mt-2">
                <SheetClose asChild>
                  <button
                    type="button"
                    onClick={() => setEvolvingMobileOpen(true)}
                    className="w-full rounded-xl border border-teal-900/20 bg-white px-2 py-3 text-xs font-semibold text-board-ink active:scale-95"
                  >
                    Evolve
                  </button>
                </SheetClose>
              </div>
            </div>
          </SheetContent>
        </Sheet>
      )}

      {isBench && (
        <Sheet open={evolvingMobileOpen} onOpenChange={setEvolvingMobileOpen}>
          <SheetContent
            side="bottom"
            className="z-[121] min-h-[36vh] max-h-[58vh] overflow-y-auto rounded-t-3xl border-t border-teal-900/15 bg-board-panel p-4 pb-6 md:hidden"
          >
            <SheetHeader className="mb-3 flex-row items-center justify-between">
              <SheetTitle className="text-sm font-bold text-board-ink">
                Evolve Pokemon
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
            {searchPanel}
          </SheetContent>
        </Sheet>
      )}

      {isBench && (
        <Dialog open={evolvingDesktopOpen} onOpenChange={setEvolvingDesktopOpen}>
          <DialogContent className="hidden md:block">
            <DialogHeader className="mb-3 flex-row items-center justify-between">
              <DialogTitle className="text-sm font-bold text-board-ink">
                Evolve Pokemon
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
            {searchPanel}
          </DialogContent>
        </Dialog>
      )}

      <div className="mb-1 flex items-center justify-between gap-1">
        <p
          className={`truncate text-xs font-semibold tracking-wide text-board-ink ${
            zone === "active" ? "w-full text-center" : "text-left"
          }`}
        >
          {title}
        </p>
      </div>

      {isBench && (
        <div className="mb-2 flex items-center justify-center md:hidden">
          <div className="min-w-14 rounded-xl bg-board-ink px-3 py-2 text-center text-sm font-black text-white">
            {slot.damage}
          </div>
        </div>
      )}

      <div className={`${zone === "bench" ? "hidden md:block" : "block"}`}>
        <div className="my-2 flex items-center justify-center gap-2">
          <button
            type="button"
            className="rounded-xl border border-teal-900/20 bg-white px-3 py-2 text-base font-bold active:scale-95"
            onClick={() => adjustDamage(side, zone, -10, benchIndex)}
          >
            -10
          </button>
          <div className="min-w-14 rounded-xl bg-board-ink px-3 py-2 text-center text-base font-black text-white">
            {slot.damage}
          </div>
          <button
            type="button"
            className="rounded-xl border border-teal-900/20 bg-white px-3 py-2 text-base font-bold active:scale-95"
            onClick={() => adjustDamage(side, zone, 10, benchIndex)}
          >
            +10
          </button>
        </div>

        {zone === "active" && (
          <div className="mb-2 flex flex-wrap justify-center gap-1">
            {STATUS_BUTTONS.map((statusBtn) => (
              <button
                key={statusBtn.status}
                type="button"
                className={chipClass(slot.statuses.includes(statusBtn.status))}
                onClick={() => toggleStatus(side, statusBtn.status)}
              >
                {statusBtn.icon}
                <span>{statusBtn.label}</span>
              </button>
            ))}
          </div>
        )}

        <div className="flex items-center gap-2">
          {zone === "bench" && (
            <button
              type="button"
              onClick={() => promoteBenchToActive(side, benchIndex ?? 0)}
              className="flex flex-1 items-center justify-center gap-1 rounded-xl border border-teal-900/20 bg-board-accentSoft px-2 py-2 text-xs font-semibold text-board-ink active:scale-95"
            >
              <ArrowUpCircle size={15} className={iconToTopClass} />
              Promote
            </button>
          )}
          <button
            type="button"
            onClick={() => knockout(side, zone, benchIndex)}
            className="flex flex-1 items-center justify-center gap-1 rounded-xl border border-rose-300 bg-rose-100 px-2 py-2 text-xs font-semibold text-rose-900 active:scale-95"
          >
            <Skull size={15} />
            Knock Out
          </button>
        </div>

        {zone === "bench" && (
          <button
            type="button"
            onClick={() => setEvolvingDesktopOpen(true)}
            className="mt-2 w-full rounded-xl border border-teal-900/20 bg-white px-2 py-2 text-xs font-semibold text-board-ink active:scale-95"
          >
            Evolve
          </button>
        )}
      </div>
    </div>
  );
}
