import React, { useState } from 'react';
import { Preset } from '../types';

interface LeftHUDProps {
  presets: Preset[];
  selectedIndex: number;
  onSelectPreset: (index: number) => void;
  onRunRecommender: (params: {
    minSun: number;
    minDTE: number;
    maxSlope: number;
    maxIce: number;
    slipDays: number;
  }) => void;
  isLoading: boolean;
  onAddCustomPreset: (name: string, lat: number, lon: number) => void;
  onDeleteCustomPreset: (id: string) => void;
  onTogglePickingMode: () => void;
  isPickingMode: boolean;
  recommendations?: Preset[];
}

export const LeftHUD: React.FC<LeftHUDProps> = ({
  presets,
  selectedIndex,
  onSelectPreset,
  onRunRecommender,
  isLoading,
  onAddCustomPreset,
  onDeleteCustomPreset,
  onTogglePickingMode,
  isPickingMode,
  recommendations = [],
}) => {
  const [activeTab, setActiveTab] = useState<'presets' | 'custom' | 'finder'>('presets');

  // Custom Site Form state
  const [customName, setCustomName] = useState('Custom Site Alpha');
  const [customLat, setCustomLat] = useState(-88.2);
  const [customLon, setCustomLon] = useState(45.0);

  // Smart Finder constraints state
  const [minSun, setMinSun] = useState(75);
  const [minDTE, setMinDTE] = useState(60);
  const [maxSlope, setMaxSlope] = useState(8);
  const [maxIce, setMaxIce] = useState(3.0);
  const [slipDays, setSlipDays] = useState(0);

  const officialPresets = presets.filter((p) => !p.is_custom && !p.id.startsWith('rec-'));
  const customPresets = presets.filter((p) => p.is_custom && !p.id.startsWith('rec-'));

  const handleAddSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    if (!customName.trim()) return;
    onAddCustomPreset(customName.trim(), customLat, customLon);
    setCustomName(`Custom Site #${customPresets.length + 2}`);
  };

  return (
    <div className="w-[310px] bg-[#080D0E]/90 border border-[#24394C]/60 rounded-xl backdrop-blur-xl shadow-2xl max-h-[calc(100vh-210px)] overflow-hidden flex flex-col pointer-events-auto">
      {/* 3 Navigation Tabs (100% English) */}
      <div className="flex border-b border-[#24394C]/60 bg-[#060A0C]/70">
        <button
          type="button"
          onClick={() => setActiveTab('presets')}
          className={`flex-1 py-2.5 text-[10px] font-semibold uppercase tracking-wider transition-all cursor-pointer ${
            activeTab === 'presets'
              ? 'text-white border-b-2 border-[#C9B188] bg-[#24394C]/30'
              : 'text-[#47525F] hover:text-[#C9B188]'
          }`}
        >
          ARTEMIS ({officialPresets.length})
        </button>
        <button
          type="button"
          onClick={() => setActiveTab('custom')}
          className={`flex-1 py-2.5 text-[10px] font-semibold uppercase tracking-wider transition-all cursor-pointer flex justify-center items-center gap-1.5 ${
            activeTab === 'custom'
              ? 'text-white border-b-2 border-[#C9B188] bg-[#24394C]/30'
              : 'text-[#47525F] hover:text-[#C9B188]'
          }`}
        >
          <span>CUSTOM</span>
          {customPresets.length > 0 && (
            <span className="w-4 h-4 rounded-full bg-[#891B1D] text-[9px] font-bold text-white flex items-center justify-center">
              {customPresets.length}
            </span>
          )}
        </button>
        <button
          type="button"
          onClick={() => setActiveTab('finder')}
          className={`flex-1 py-2.5 text-[10px] font-semibold uppercase tracking-wider transition-all cursor-pointer flex justify-center items-center gap-1.5 ${
            activeTab === 'finder'
              ? 'text-white border-b-2 border-[#C9B188] bg-[#24394C]/30'
              : 'text-[#47525F] hover:text-[#C9B188]'
          }`}
        >
          <span>AUTO-FIND</span>
          {recommendations.length > 0 && (
            <span className="w-4 h-4 rounded-full bg-[#FCE873] text-[9px] font-bold text-[#080D0E] flex items-center justify-center">
              {recommendations.length}
            </span>
          )}
        </button>
      </div>

      <div className="p-3 overflow-y-auto flex-1">
        {/* Tab 1: Official Artemis Zones */}
        {activeTab === 'presets' && (
          <div className="flex flex-col gap-1.5">
            {officialPresets.map((preset) => {
              const globalIndex = presets.findIndex((p) => p.id === preset.id);
              const isSelected = globalIndex === selectedIndex;
              return (
                <div
                  key={preset.id}
                  onClick={() => onSelectPreset(globalIndex)}
                  className={`px-3 py-2 rounded-lg cursor-pointer transition-all flex justify-between items-center group border ${
                    isSelected
                      ? 'bg-[#24394C]/80 border-[#FCE873] shadow-md'
                      : 'bg-[#24394C]/20 border-transparent hover:bg-[#24394C]/50 hover:border-[#47525F]/50'
                  }`}
                >
                  <div className="min-w-0">
                    <div className={`text-[12px] font-semibold leading-tight truncate ${
                      isSelected ? 'text-white' : 'text-[#C9B188] group-hover:text-white'
                    }`}>
                      {preset.name}
                    </div>
                    <div className="text-[10px] text-[#47525F] mt-0.5 truncate mono">
                      {Math.abs(preset.latitude).toFixed(1)}°S, {preset.longitude >= 0 ? '+' : ''}{preset.longitude.toFixed(1)}° · {preset.sub_region}
                    </div>
                  </div>
                  <div className={`text-[12px] font-bold mono shrink-0 ml-2 ${
                    isSelected ? 'text-[#FCE873]' : 'text-[#47525F]'
                  }`}>
                    {(95 - globalIndex * 3.5).toFixed(0)}
                  </div>
                </div>
              );
            })}
          </div>
        )}

        {/* Tab 2: Custom Sites Management */}
        {activeTab === 'custom' && (
          <div className="flex flex-col gap-3">
            {/* Pick on 3D Moon Button */}
            <button
              type="button"
              onClick={onTogglePickingMode}
              className={`w-full py-2.5 px-3 rounded-lg text-xs font-semibold tracking-wider uppercase border transition-all cursor-pointer flex items-center justify-center gap-2 ${
                isPickingMode
                  ? 'bg-[#891B1D] border-[#FCE873] text-white shadow-[0_0_12px_rgba(252,232,115,0.4)] animate-pulse'
                  : 'bg-[#24394C] border-[#C9B188] hover:bg-[#C9B188] hover:text-[#080D0E] text-white'
              }`}
            >
              <span>{isPickingMode ? '🎯' : '📍'}</span>
              <span>{isPickingMode ? 'Click on 3D Moon to Drop…' : 'Pick on 3D Moon'}</span>
            </button>

            {/* Manual Lat/Lon Entry Form */}
            <form onSubmit={handleAddSubmit} className="bg-[#24394C]/20 border border-[#47525F]/40 p-2.5 rounded-lg flex flex-col gap-2">
              <div className="text-[10px] font-semibold uppercase text-[#C9B188] tracking-wider">
                Manual Coordinate Entry
              </div>

              <div>
                <label className="text-[10px] text-[#47525F] block mb-0.5">Site Name</label>
                <input
                  type="text"
                  value={customName}
                  onChange={(e) => setCustomName(e.target.value)}
                  className="w-full bg-[#080D0E] border border-[#47525F]/60 rounded px-2 py-1 text-xs text-white outline-none focus:border-[#FCE873]"
                  placeholder="e.g. Shackleton Crater Base"
                />
              </div>

              <div className="grid grid-cols-2 gap-2">
                <div>
                  <div className="flex justify-between text-[10px] text-[#47525F]">
                    <span>Latitude (S)</span>
                    <span className="mono text-[#FCE873]">{customLat.toFixed(1)}°</span>
                  </div>
                  <input
                    type="range"
                    min="-89.9"
                    max="-80.0"
                    step="0.1"
                    value={customLat}
                    onChange={(e) => setCustomLat(Number(e.target.value))}
                    className="w-full accent-[#FCE873] cursor-pointer mt-1"
                  />
                </div>

                <div>
                  <div className="flex justify-between text-[10px] text-[#47525F]">
                    <span>Longitude</span>
                    <span className="mono text-[#FCE873]">{customLon.toFixed(1)}°</span>
                  </div>
                  <input
                    type="range"
                    min="-180.0"
                    max="180.0"
                    step="0.5"
                    value={customLon}
                    onChange={(e) => setCustomLon(Number(e.target.value))}
                    className="w-full accent-[#FCE873] cursor-pointer mt-1"
                  />
                </div>
              </div>

              <button
                type="submit"
                className="w-full mt-1 py-1.5 bg-[#24394C]/80 border border-[#C9B188]/60 hover:bg-[#C9B188] hover:text-[#080D0E] text-white text-[11px] font-semibold rounded uppercase tracking-wider transition-colors cursor-pointer"
              >
                + Add &amp; Evaluate Site
              </button>
            </form>

            {/* List of Custom Presets */}
            <div className="flex flex-col gap-1.5">
              <div className="text-[10px] uppercase font-semibold text-[#47525F] tracking-wider">
                User Custom Sites ({customPresets.length})
              </div>

              {customPresets.length === 0 ? (
                <div className="text-xs text-[#47525F] p-3 text-center border border-dashed border-[#47525F]/30 rounded-lg">
                  No custom sites added yet. Click on the 3D Moon or enter coordinates above.
                </div>
              ) : (
                customPresets.map((preset) => {
                  const globalIndex = presets.findIndex((p) => p.id === preset.id);
                  const isSelected = globalIndex === selectedIndex;
                  return (
                    <div
                      key={preset.id}
                      onClick={() => onSelectPreset(globalIndex)}
                      className={`px-3 py-2 rounded-lg cursor-pointer transition-all flex justify-between items-center group border ${
                        isSelected
                          ? 'bg-[#24394C]/80 border-[#FCE873] shadow-md'
                          : 'bg-[#24394C]/25 border-transparent hover:bg-[#24394C]/50 hover:border-[#47525F]/50'
                      }`}
                    >
                      <div className="min-w-0">
                        <div className={`text-[12px] font-semibold leading-tight truncate ${
                          isSelected ? 'text-white' : 'text-[#C9B188] group-hover:text-white'
                        }`}>
                          ★ {preset.name}
                        </div>
                        <div className="text-[10px] text-[#47525F] mt-0.5 truncate mono">
                          {Math.abs(preset.latitude).toFixed(2)}°S, {preset.longitude >= 0 ? '+' : ''}{preset.longitude.toFixed(2)}°
                        </div>
                      </div>

                      <button
                        type="button"
                        onClick={(e) => {
                          e.stopPropagation();
                          onDeleteCustomPreset(preset.id);
                        }}
                        className="text-[#47525F] hover:text-[#891B1D] px-1.5 py-0.5 rounded text-xs cursor-pointer ml-2"
                        title="Delete site"
                      >
                        ✕
                      </button>
                    </div>
                  );
                })
              )}
            </div>
          </div>
        )}

        {/* Tab 3: Smart Recommender Sliders & Pareto Results */}
        {activeTab === 'finder' && (
          <div className="flex flex-col gap-3">
            {[
              { label: 'Min Solar Illumination', value: minSun, set: setMinSun, min: 40, max: 95, unit: '%' },
              { label: 'Min Earth Link (DTE)', value: minDTE, set: setMinDTE, min: 30, max: 90, unit: '%' },
              { label: 'Max Terrain Slope', value: maxSlope, set: setMaxSlope, min: 3, max: 15, unit: '°' },
            ].map((s) => (
              <div key={s.label}>
                <div className="flex justify-between text-xs text-[#47525F] mb-1">
                  <span>{s.label}</span>
                  <span className="text-[#FCE873] font-semibold mono">{s.value}{s.unit}</span>
                </div>
                <input
                  type="range"
                  min={s.min}
                  max={s.max}
                  value={s.value}
                  onChange={(e) => s.set(Number(e.target.value))}
                  className="w-full accent-[#FCE873] cursor-pointer"
                />
              </div>
            ))}

            <div>
              <div className="flex justify-between text-xs text-[#47525F] mb-1">
                <span>Max Distance to Water Ice</span>
                <span className="text-[#FCE873] font-semibold mono">{maxIce.toFixed(1)} km</span>
              </div>
              <input
                type="range"
                min="0.5"
                max="10"
                step="0.5"
                value={maxIce}
                onChange={(e) => setMaxIce(Number(e.target.value))}
                className="w-full accent-[#FCE873] cursor-pointer"
              />
            </div>

            <div>
              <div className="flex justify-between text-xs text-[#47525F] mb-1">
                <span>Launch Slip Window</span>
                <span className="text-[#FCE873] font-semibold mono">
                  {slipDays > 0 ? `+${slipDays}` : slipDays} Days
                </span>
              </div>
              <input
                type="range"
                min="-7"
                max="7"
                value={slipDays}
                onChange={(e) => setSlipDays(Number(e.target.value))}
                className="w-full accent-[#FCE873] cursor-pointer"
              />
            </div>

            <button
              type="button"
              onClick={() => onRunRecommender({ minSun, minDTE, maxSlope, maxIce, slipDays })}
              disabled={isLoading}
              className="w-full mt-1 py-2.5 px-4 bg-[#24394C] border border-[#C9B188] hover:bg-[#C9B188] hover:text-[#080D0E] text-white text-xs font-semibold tracking-wider uppercase rounded-md transition-all flex justify-center items-center gap-2 cursor-pointer disabled:opacity-50"
            >
              {isLoading ? (
                <span className="animate-pulse">SCANNING POLAR CAP…</span>
              ) : (
                <>
                  <span>⌕</span>
                  <span>EXECUTE INVERSE SEARCH</span>
                </>
              )}
            </button>

            {/* Smart Finder Results List (Preserved cleanly) */}
            {recommendations.length > 0 && (
              <div className="flex flex-col gap-1.5 mt-2 pt-2 border-t border-[#24394C]/60">
                <div className="text-[10px] uppercase font-semibold text-[#FCE873] tracking-wider flex justify-between">
                  <span>Pareto-Optimal Sites ({recommendations.length})</span>
                  <span>Score</span>
                </div>
                {recommendations.map((rec) => {
                  const globalIndex = presets.findIndex((p) => p.id === rec.id);
                  const isSelected = globalIndex === selectedIndex;
                  return (
                    <div
                      key={rec.id}
                      onClick={() => {
                        if (globalIndex >= 0) onSelectPreset(globalIndex);
                      }}
                      className={`px-3 py-2 rounded-lg cursor-pointer transition-all flex justify-between items-center group border ${
                        isSelected
                          ? 'bg-[#24394C]/80 border-[#FCE873] shadow-md'
                          : 'bg-[#24394C]/25 border-transparent hover:bg-[#24394C]/50 hover:border-[#47525F]/50'
                      }`}
                    >
                      <div className="min-w-0">
                        <div className={`text-[12px] font-semibold leading-tight truncate ${
                          isSelected ? 'text-white' : 'text-[#C9B188] group-hover:text-white'
                        }`}>
                          {rec.name}
                        </div>
                        <div className="text-[10px] text-[#47525F] mt-0.5 truncate mono">
                          {Math.abs(rec.latitude).toFixed(2)}°S, {rec.longitude >= 0 ? '+' : ''}{rec.longitude.toFixed(2)}° · {rec.sub_region}
                        </div>
                      </div>
                      <div className="text-[12px] font-bold mono text-[#FCE873] ml-2 shrink-0">
                        92.5
                      </div>
                    </div>
                  );
                })}
              </div>
            )}
          </div>
        )}
      </div>
    </div>
  );
};
