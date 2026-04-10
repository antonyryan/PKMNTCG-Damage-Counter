import { useEffect, useMemo, useState, type ReactElement } from "react";
import {
  ArrowUpCircle,
  Flame,
  Moon,
  Plus,
  Skull,
  Sparkles,
  Waves,
} from "lucide-react";
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
  const [adding, setAdding] = useState(false);
  const [query, setQuery] = useState("");
  const [loading, setLoading] = useState(false);
  const [results, setResults] = useState<PokemonSearchResult[]>([]);

  useEffect(() => {
    if (!adding) {
      setQuery("");
      setResults([]);
      return;
    }

    const control = new AbortController();
    // Debounce search requests so the autocomplete stays responsive on mobile.
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
  }, [query, adding]);

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

  if (!slot.pokemon) {
    return (
      <div className="relative rounded-2xl border border-dashed border-teal-900/25 bg-white/40 p-2">
        {!adding && (
          <button
            type="button"
            onClick={() => setAdding(true)}
            className="flex h-full min-h-20 w-full items-center justify-center rounded-xl bg-white/70 py-5 text-board-ink active:scale-95"
          >
            <Plus size={32} />
          </button>
        )}

        {adding && (
          <div className="space-y-2">
            <input
              value={query}
              onChange={(e) => setQuery(e.target.value)}
              placeholder="Search Pokemon"
              className="w-full rounded-xl border border-teal-900/20 bg-white px-3 py-2 text-sm outline-none focus:border-board-accent"
              autoFocus
            />
            <div className="max-h-32 overflow-y-auto rounded-xl border border-teal-900/15 bg-white">
              {loading && (
                <p className="p-2 text-xs text-slate-500">Searching...</p>
              )}
              {!loading && results.length === 0 && (
                <p className="p-2 text-xs text-slate-500">No results</p>
              )}
              {!loading &&
                results.map((pokemon) => (
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
                      setAdding(false);
                    }}
                  >
                    {pokemon.name}
                  </button>
                ))}
            </div>
            <button
              type="button"
              className="w-full rounded-xl border border-slate-300 bg-white py-1 text-xs"
              onClick={() => setAdding(false)}
            >
              Cancel
            </button>
          </div>
        )}
      </div>
    );
  }

  return (
    <div className="rounded-2xl border border-teal-900/20 bg-white/80 p-2 shadow-card backdrop-blur-sm">
      <p className="truncate text-center text-xs font-semibold tracking-wide text-board-ink">
        {title}
      </p>

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
            <ArrowUpCircle size={15} />
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
    </div>
  );
}
