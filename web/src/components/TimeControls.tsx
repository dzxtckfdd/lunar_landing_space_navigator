import React, { useState } from 'react';

interface TimeControlsProps {
  currentDate: Date;
  onSetDate: (d: Date) => void;
  isPlaying: boolean;
  onTogglePlay: () => void;
  speed: number;
  onSetSpeed: (speed: number) => void;
  launchSlipDays: number;
  onSetLaunchSlip: (days: number) => void;
  missionDurationHours: number;
  missionHour: number;
  onScrubHour: (hour: number) => void;
}

const SPEED_PRESETS = [
  { label: '1×', value: 1 },
  { label: '10×', value: 10 },
  { label: '1m/s', value: 60 },
  { label: '1h/s', value: 3600 },
  { label: '1d/s', value: 86400 },
];

const MISSION_PRESETS = [
  { label: 'Artemis III Launch (15 Nov 2026)', date: new Date('2026-11-15T00:00:00Z') },
  { label: 'South Pole Peak Sun (21 Dec 2026)', date: new Date('2026-12-21T00:00:00Z') },
  { label: 'Low Sun Grazing (21 Jun 2027)', date: new Date('2027-06-21T00:00:00Z') },
  { label: 'Artemis IV Window (10 Sep 2028)', date: new Date('2028-09-10T00:00:00Z') },
  { label: 'Current Time (Now)', date: new Date() },
];

export const TimeControls: React.FC<TimeControlsProps> = ({
  currentDate,
  onSetDate,
  isPlaying,
  onTogglePlay,
  speed,
  onSetSpeed,
  launchSlipDays,
  onSetLaunchSlip,
  missionDurationHours,
  missionHour,
  onScrubHour,
}) => {
  const [showDatePicker, setShowDatePicker] = useState(false);
  const [dateInputVal, setDateInputVal] = useState(
    currentDate.toISOString().slice(0, 16)
  );

  // Format UTC date string for display (NASA Eyes style)
  const monthNames = ['JAN', 'FEB', 'MAR', 'APR', 'MAY', 'JUN', 'JUL', 'AUG', 'SEP', 'OCT', 'NOV', 'DEC'];
  const day = String(currentDate.getUTCDate()).padStart(2, '0');
  const month = monthNames[currentDate.getUTCMonth()];
  const year = currentDate.getUTCFullYear();
  const hours = String(currentDate.getUTCHours()).padStart(2, '0');
  const minutes = String(currentDate.getUTCMinutes()).padStart(2, '0');
  const seconds = String(currentDate.getUTCSeconds()).padStart(2, '0');

  const handleStep = (hoursDelta: number) => {
    const next = new Date(currentDate.getTime() + hoursDelta * 3600 * 1000);
    onSetDate(next);
  };

  const handleDateSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    const parsed = new Date(dateInputVal + ':00Z');
    if (!isNaN(parsed.getTime())) {
      onSetDate(parsed);
      setShowDatePicker(false);
    }
  };

  return (
    <div className="fixed bottom-6 left-1/2 -translate-x-1/2 z-20 flex flex-col items-center gap-2 select-none pointer-events-auto">
      {/* Date Picker Popover */}
      {showDatePicker && (
        <div className="mb-2 bg-[#0a0f14]/95 border border-white/20 rounded-2xl p-4 shadow-2xl backdrop-blur-2xl text-white w-[360px] animate-in fade-in zoom-in-95 duration-150">
          <div className="flex justify-between items-center pb-2 border-b border-white/10 mb-3">
            <span className="text-xs font-semibold uppercase tracking-wider text-white/70">
              Jump to Calendar Date (UTC)
            </span>
            <button
              type="button"
              onClick={() => setShowDatePicker(false)}
              className="text-white/40 hover:text-white text-sm cursor-pointer"
            >
              ✕
            </button>
          </div>

          <form onSubmit={handleDateSubmit} className="flex flex-col gap-3">
            <div>
              <label className="text-[10px] text-white/50 block mb-1 uppercase tracking-wider">
                Select Date & Time (2024–2035)
              </label>
              <input
                type="datetime-local"
                value={dateInputVal}
                onChange={(e) => setDateInputVal(e.target.value)}
                min="2024-01-01T00:00"
                max="2035-12-31T23:59"
                className="w-full bg-white/5 border border-white/20 rounded-lg px-3 py-1.5 text-xs font-mono text-white focus:outline-none focus:border-[#fce873]"
              />
            </div>

            <button
              type="submit"
              className="w-full bg-[#fce873] hover:bg-[#ffe247] text-black font-semibold text-xs py-1.5 rounded-lg transition-colors cursor-pointer"
            >
              Set Simulation Date
            </button>
          </form>

          {/* Quick Mission Bookmarks */}
          <div className="mt-3 pt-3 border-t border-white/10">
            <span className="text-[10px] text-white/40 uppercase tracking-wider block mb-1.5">
              Mission Presets
            </span>
            <div className="flex flex-col gap-1">
              {MISSION_PRESETS.map((p) => (
                <button
                  key={p.label}
                  type="button"
                  onClick={() => {
                    onSetDate(p.date);
                    setDateInputVal(p.date.toISOString().slice(0, 16));
                    setShowDatePicker(false);
                  }}
                  className="text-left text-[11px] text-white/80 hover:text-white hover:bg-white/10 px-2 py-1 rounded transition-colors flex items-center justify-between cursor-pointer"
                >
                  <span>{p.label}</span>
                  <span className="text-[10px] text-white/40 font-mono">
                    {p.date.getUTCFullYear()}
                  </span>
                </button>
              ))}
            </div>
          </div>
        </div>
      )}

      {/* Main NASA Eyes Floating Time Bar */}
      <div className="bg-[#080d12]/90 border border-white/15 rounded-2xl px-5 py-2.5 shadow-2xl backdrop-blur-2xl flex flex-col gap-2 min-w-[700px] max-w-[860px]">
        {/* Top Row: Date Display, Playback, Speed, Launch Slip */}
        <div className="flex items-center justify-between gap-4">
          {/* Calendar Date & Click to Change */}
          <button
            type="button"
            onClick={() => {
              setDateInputVal(currentDate.toISOString().slice(0, 16));
              setShowDatePicker((prev) => !prev);
            }}
            className="flex items-center gap-2 px-3 py-1 bg-white/5 hover:bg-white/10 border border-white/15 rounded-xl text-left cursor-pointer transition-colors group"
            title="Click to change date"
          >
            <span className="text-sm">📅</span>
            <div>
              <div className="text-[13px] font-mono font-bold text-white group-hover:text-[#fce873] transition-colors leading-tight">
                {day} {month} {year} · {hours}:{minutes}:{seconds} <span className="text-[10px] text-white/50">UTC</span>
              </div>
              <div className="text-[9px] uppercase tracking-wider text-white/40">
                CLICK TO JUMP DATE
              </div>
            </div>
          </button>

          {/* Steppers & Playback Controls */}
          <div className="flex items-center gap-1.5">
            <button
              type="button"
              onClick={() => handleStep(-24)}
              className="px-2 py-1 rounded-lg bg-white/5 hover:bg-white/15 text-[11px] font-mono text-white/80 hover:text-white transition-colors cursor-pointer"
              title="-24 hours"
            >
              -1d
            </button>
            <button
              type="button"
              onClick={() => handleStep(-1)}
              className="px-2 py-1 rounded-lg bg-white/5 hover:bg-white/15 text-[11px] font-mono text-white/80 hover:text-white transition-colors cursor-pointer"
              title="-1 hour"
            >
              -1h
            </button>

            {/* Play / Pause Toggle */}
            <button
              type="button"
              onClick={onTogglePlay}
              className={`px-3 py-1 rounded-lg text-xs font-semibold flex items-center gap-1.5 transition-all cursor-pointer ${
                isPlaying
                  ? 'bg-rose-600/80 hover:bg-rose-600 text-white shadow-[0_0_12px_rgba(225,29,72,0.4)]'
                  : 'bg-[#fce873] hover:bg-[#ffe247] text-black font-bold shadow-[0_0_12px_rgba(252,232,115,0.4)]'
              }`}
            >
              <span>{isPlaying ? '❚❚' : '▶'}</span>
              <span>{isPlaying ? 'PAUSE' : 'PLAY'}</span>
            </button>

            <button
              type="button"
              onClick={() => handleStep(1)}
              className="px-2 py-1 rounded-lg bg-white/5 hover:bg-white/15 text-[11px] font-mono text-white/80 hover:text-white transition-colors cursor-pointer"
              title="+1 hour"
            >
              +1h
            </button>
            <button
              type="button"
              onClick={() => handleStep(24)}
              className="px-2 py-1 rounded-lg bg-white/5 hover:bg-white/15 text-[11px] font-mono text-white/80 hover:text-white transition-colors cursor-pointer"
              title="+24 hours"
            >
              +1d
            </button>
          </div>

          {/* Speed Presets */}
          <div className="flex items-center bg-white/5 p-0.5 rounded-lg border border-white/10">
            {SPEED_PRESETS.map((s) => (
              <button
                key={s.value}
                type="button"
                onClick={() => onSetSpeed(s.value)}
                className={`px-2 py-0.5 rounded text-[10px] font-mono transition-colors cursor-pointer ${
                  speed === s.value
                    ? 'bg-white/20 text-[#fce873] font-bold'
                    : 'text-white/40 hover:text-white'
                }`}
              >
                {s.label}
              </button>
            ))}
          </div>

          {/* Launch Slip Stepper (Key Challenge Requirement) */}
          <div className="flex items-center gap-2 pl-2 border-l border-white/10">
            <div className="text-right">
              <div className="text-[9px] uppercase tracking-wider text-white/40">LAUNCH SLIP</div>
              <div className="text-[11px] font-mono font-bold text-[#fce873]">
                {launchSlipDays >= 0 ? `+${launchSlipDays}` : launchSlipDays} DAYS
              </div>
            </div>
            <div className="flex flex-col gap-0.5">
              <button
                type="button"
                onClick={() => onSetLaunchSlip(Math.min(14, launchSlipDays + 1))}
                className="w-5 h-3.5 bg-white/10 hover:bg-white/20 rounded text-[9px] flex items-center justify-center text-white cursor-pointer"
                title="+1 day slip"
              >
                ▲
              </button>
              <button
                type="button"
                onClick={() => onSetLaunchSlip(Math.max(-14, launchSlipDays - 1))}
                className="w-5 h-3.5 bg-white/10 hover:bg-white/20 rounded text-[9px] flex items-center justify-center text-white cursor-pointer"
                title="-1 day slip"
              >
                ▼
              </button>
            </div>
          </div>
        </div>

        {/* Bottom Row: Mission Hour Scrub Slider */}
        <div className="flex items-center gap-3 pt-1 border-t border-white/10">
          <span className="text-[10px] font-mono text-white/50 shrink-0">
            T+0h
          </span>
          <input
            type="range"
            min="0"
            max={missionDurationHours}
            step="1"
            value={missionHour}
            onChange={(e) => onScrubHour(Number(e.target.value))}
            className="w-full h-1.5 bg-white/15 rounded-lg appearance-none cursor-pointer accent-[#fce873]"
          />
          <span className="text-[10px] font-mono text-white/70 shrink-0 min-w-[70px] text-right">
            T+{Math.floor(missionHour)}h / {missionDurationHours}h
          </span>
        </div>
      </div>
    </div>
  );
};
