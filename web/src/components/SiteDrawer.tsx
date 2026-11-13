import React, { useState } from 'react';
import { Preset } from '../types';

interface SiteDrawerProps {
  presets: Preset[];
  selectedIndex: number;
  onSelectPreset: (idx: number) => void;
  onRunRecommender: (params: {
    minSun: number;
    minDTE: number;
    maxSlope: number;
    maxIce: number;
  }) => void;
  isRecommending: boolean;
  recommendations: Preset[];
  isPickingMode: boolean;
  onTogglePickingMode: () => void;
  onViewSouthPole: () => void;
  onViewGlobalOrbit: () => void;
  onViewTargetLock: () => void;
}

export const SiteDrawer: React.FC<SiteDrawerProps> = ({
  presets,
  selectedIndex,
  onSelectPreset,
  onRunRecommender,
  isRecommending,
  recommendations,
  isPickingMode,
  onTogglePickingMode,
  onViewSouthPole,
  onViewGlobalOrbit,
  onViewTargetLock,
}) => {
  const [isOpen, setIsOpen] = useState(false);
  const [activeTab, setActiveTab] = useState<'sites' | 'finder'>('sites');

  // Finder criteria state
  const [minSun, setMinSun] = useState(75);
  const [minDTE, setMinDTE] = useState(60);
  const [maxSlope, setMaxSlope] = useState(8);
  const [maxIce, setMaxIce] = useState(3.0);

  const officialPresets = presets.filter((p) => !p.id.startsWith('rec-'));

  const handleFinderSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    onRunRecommender({
      minSun,
      minDTE,
      maxSlope,
      maxIce,
    });
  };

  return (
    <>
      {/* Floating Top-Left Triggers */}
      <div className="fixed top-20 left-6 z-20 flex items-center gap-2 pointer-events-auto select-none">
        <button
          type="button"
          onClick={() => setIsOpen((prev) => !prev)}
          className={`flex items-center gap-2 px-4 py-2 rounded-full border shadow-2xl backdrop-blur-xl transition-all cursor-pointer ${
            isOpen
              ? 'bg-[#fce873] border-[#fce873] text-black font-bold'
              : 'bg-black/50 hover:bg-black/80 border-white/20 text-white font-medium'
          }`}
        >
          <span>🎯</span>
          <span className="text-xs uppercase tracking-wider">
            {isOpen ? 'CLOSE SITES' : 'LANDING SITES'}
          </span>
          <span className="text-[10px] font-mono px-1.5 py-0.2 bg-white/10 rounded-full">
            {presets[selectedIndex]?.name || 'SELECT'}
          </span>
        </button>

        {/* Quick Camera Views */}
        <div className="hidden sm:flex items-center bg-black/40 border border-white/15 rounded-full p-1 backdrop-blur-xl gap-1">
          <button
            type="button"
            onClick={onViewSouthPole}
            className="px-2.5 py-1 rounded-full text-[10px] font-mono uppercase text-white/70 hover:text-white hover:bg-white/10 transition-colors cursor-pointer"
            title="Focus camera straight down on Lunar South Pole"
          >
            SOUTH POLE
          </button>
          <button
            type="button"
            onClick={onViewGlobalOrbit}
            className="px-2.5 py-1 rounded-full text-[10px] font-mono uppercase text-white/70 hover:text-white hover:bg-white/10 transition-colors cursor-pointer"
            title="Orbit view showing Moon, Earth, and Sun"
          >
            GLOBAL ORBIT
          </button>
          <button
            type="button"
            onClick={onViewTargetLock}
            className="px-2.5 py-1 rounded-full text-[10px] font-mono uppercase text-[#fce873] hover:text-white hover:bg-white/10 transition-colors cursor-pointer"
            title="Target Lock active site"
          >
            TARGET LOCK
          </button>
        </div>

        {/* Pick on Moon toggle */}
        <button
          type="button"
          onClick={onTogglePickingMode}
          className={`px-3 py-2 rounded-full border text-xs font-semibold uppercase tracking-wider transition-all cursor-pointer flex items-center gap-1.5 ${
            isPickingMode
              ? 'bg-rose-600 border-rose-400 text-white shadow-[0_0_12px_rgba(225,29,72,0.6)] animate-pulse'
              : 'bg-black/40 hover:bg-black/80 border-white/20 text-white/80 hover:text-white'
          }`}
          title="Click anywhere on the 3D surface to drop a custom landing point"
        >
          <span>📍</span>
          <span>{isPickingMode ? 'CLICK ON MOON' : 'DROP PIN'}</span>
        </button>
      </div>

      {/* Slide-out Drawer */}
      {isOpen && (
        <div className="fixed top-32 left-6 bottom-28 z-20 w-[360px] bg-[#070b10]/95 border border-white/20 rounded-3xl shadow-2xl backdrop-blur-2xl flex flex-col overflow-hidden pointer-events-auto text-white animate-in fade-in slide-in-from-left-8 duration-200">
          {/* Tabs */}
          <div className="flex border-b border-white/10 bg-white/5">
            <button
              type="button"
              onClick={() => setActiveTab('sites')}
              className={`flex-1 py-3 text-xs font-semibold uppercase tracking-wider transition-colors cursor-pointer text-center ${
                activeTab === 'sites'
                  ? 'text-[#fce873] border-b-2 border-[#fce873] bg-white/5'
                  : 'text-white/50 hover:text-white'
              }`}
            >
              ARTEMIS SITES ({officialPresets.length})
            </button>
            <button
              type="button"
              onClick={() => setActiveTab('finder')}
              className={`flex-1 py-3 text-xs font-semibold uppercase tracking-wider transition-colors cursor-pointer text-center flex items-center justify-center gap-1.5 ${
                activeTab === 'finder'
                  ? 'text-[#fce873] border-b-2 border-[#fce873] bg-white/5'
                  : 'text-white/50 hover:text-white'
              }`}
            >
              <span>AUTO-FINDER</span>
              {recommendations.length > 0 && (
                <span className="w-4 h-4 rounded-full bg-[#fce873] text-[9px] font-bold text-black flex items-center justify-center">
                  {recommendations.length}
                </span>
              )}
            </button>
          </div>

          {/* Tab 1: Sites List */}
          {activeTab === 'sites' && (
            <div className="flex-1 overflow-y-auto p-3 flex flex-col gap-2">
              <div className="text-[10px] uppercase tracking-wider text-white/40 px-2 pt-1">
                Official Artemis III Candidate Zones
              </div>
              {officialPresets.map((preset) => {
                const isSelected = presets[selectedIndex]?.id === preset.id;
                return (
                  <button
                    key={preset.id}
                    type="button"
                    onClick={() => {
                      const realIndex = presets.findIndex((p) => p.id === preset.id);
                      if (realIndex !== -1) onSelectPreset(realIndex);
                    }}
                    className={`w-full p-3 rounded-2xl border text-left transition-all cursor-pointer flex flex-col gap-1 ${
                      isSelected
                        ? 'bg-[#fce873]/15 border-[#fce873] text-white shadow-[0_0_15px_rgba(252,232,115,0.2)]'
                        : 'bg-white/5 hover:bg-white/10 border-white/10 text-white/80'
                    }`}
                  >
                    <div className="flex justify-between items-start">
                      <span className="text-xs font-bold leading-tight">
                        {preset.name}
                      </span>
                      {isSelected && (
                        <span className="text-[9px] font-mono px-1.5 py-0.5 bg-[#fce873] text-black font-bold rounded">
                          SELECTED
                        </span>
                      )}
                    </div>
                    <div className="text-[10px] text-white/50 font-mono">
                      {Math.abs(preset.latitude).toFixed(2)}°S, {Math.abs(preset.longitude).toFixed(2)}° · {Math.round(preset.elevation_m)}m
                    </div>
                    <div className="flex items-center gap-2 mt-1 text-[10px] text-white/60">
                      <span>Slope: {preset.average_slope_deg.toFixed(1)}°</span>
                      <span>·</span>
                      <span>Ice: {preset.distance_to_psr_km.toFixed(1)}km</span>
                    </div>
                  </button>
                );
              })}
            </div>
          )}

          {/* Tab 2: Smart Recommender Finder */}
          {activeTab === 'finder' && (
            <div className="flex-1 overflow-y-auto p-4 flex flex-col gap-4">
              <form onSubmit={handleFinderSubmit} className="flex flex-col gap-3">
                <div className="text-[10px] uppercase font-bold tracking-wider text-[#fce873]">
                  Inverse Constraint Filtering
                </div>

                {/* Min Solar */}
                <div>
                  <div className="flex justify-between text-xs mb-1">
                    <span className="text-white/70">Min Sunlight</span>
                    <span className="font-mono text-[#fce873]">{minSun}%</span>
                  </div>
                  <input
                    type="range"
                    min="30"
                    max="95"
                    step="5"
                    value={minSun}
                    onChange={(e) => setMinSun(Number(e.target.value))}
                    className="w-full h-1 bg-white/20 rounded accent-[#fce873] cursor-pointer"
                  />
                </div>

                {/* Min DTE */}
                <div>
                  <div className="flex justify-between text-xs mb-1">
                    <span className="text-white/70">Min DTE Comms</span>
                    <span className="font-mono text-sky-400">{minDTE}%</span>
                  </div>
                  <input
                    type="range"
                    min="20"
                    max="90"
                    step="5"
                    value={minDTE}
                    onChange={(e) => setMinDTE(Number(e.target.value))}
                    className="w-full h-1 bg-white/20 rounded accent-sky-400 cursor-pointer"
                  />
                </div>

                {/* Max Slope */}
                <div>
                  <div className="flex justify-between text-xs mb-1">
                    <span className="text-white/70">Max Slope</span>
                    <span className="font-mono text-emerald-400">≤ {maxSlope}°</span>
                  </div>
                  <input
                    type="range"
                    min="3"
                    max="15"
                    step="1"
                    value={maxSlope}
                    onChange={(e) => setMaxSlope(Number(e.target.value))}
                    className="w-full h-1 bg-white/20 rounded accent-emerald-400 cursor-pointer"
                  />
                </div>

                {/* Max Distance to Ice */}
                <div>
                  <div className="flex justify-between text-xs mb-1">
                    <span className="text-white/70">Max Distance to PSR Ice</span>
                    <span className="font-mono text-cyan-300">≤ {maxIce}km</span>
                  </div>
                  <input
                    type="range"
                    min="0.5"
                    max="8.0"
                    step="0.5"
                    value={maxIce}
                    onChange={(e) => setMaxIce(Number(e.target.value))}
                    className="w-full h-1 bg-white/20 rounded accent-cyan-300 cursor-pointer"
                  />
                </div>

                <button
                  type="submit"
                  disabled={isRecommending}
                  className="w-full mt-2 bg-[#fce873] hover:bg-[#ffe247] disabled:opacity-50 text-black font-bold text-xs py-2 rounded-xl transition-all shadow-lg cursor-pointer"
                >
                  {isRecommending ? 'Running Spatial Filter...' : 'Find Top-5 Ranked Sites'}
                </button>
              </form>

              {/* Recommendations output */}
              {recommendations.length > 0 && (
                <div className="flex flex-col gap-2 pt-2 border-t border-white/10">
                  <div className="text-[10px] uppercase font-bold tracking-wider text-[#fce873]">
                    Recommended Sites (Pareto Ranked)
                  </div>
                  {recommendations.map((rec) => {
                    const isSelected = presets[selectedIndex]?.id === rec.id;
                    return (
                      <button
                        key={rec.id}
                        type="button"
                        onClick={() => {
                          const idx = presets.findIndex((p) => p.id === rec.id);
                          if (idx !== -1) onSelectPreset(idx);
                        }}
                        className={`w-full p-2.5 rounded-xl border text-left transition-all cursor-pointer ${
                          isSelected
                            ? 'bg-[#fce873]/20 border-[#fce873] text-white'
                            : 'bg-white/5 hover:bg-white/10 border-white/10 text-white/80'
                        }`}
                      >
                        <div className="text-xs font-bold">{rec.name}</div>
                        <div className="text-[10px] text-white/50 font-mono mt-0.5">
                          {Math.abs(rec.latitude).toFixed(2)}°S, {Math.abs(rec.longitude).toFixed(2)}° · Score: {rec.scientific_significance || '90+'}
                        </div>
                      </button>
                    );
                  })}
                </div>
              )}
            </div>
          )}
        </div>
      )}
    </>
  );
};
