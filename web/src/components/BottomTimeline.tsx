import React from 'react';

interface BottomTimelineProps {
  missionHour: number;
  isPlaying: boolean;
  speed: number;
  onScrub: (hour: number) => void;
  onTogglePlay: () => void;
  onSetSpeed: (speed: number) => void;
}

export const BottomTimeline: React.FC<BottomTimelineProps> = ({
  missionHour,
  isPlaying,
  speed,
  onScrub,
  onTogglePlay,
  onSetSpeed,
}) => {
  const currentDay = (missionHour / 24.0).toFixed(1);
  const totalDays = '14.0';

  return (
    <div className="w-full max-w-[840px] bg-[#080D0E]/90 border border-[#24394C]/60 rounded-xl p-3 px-5 backdrop-blur-xl shadow-2xl flex flex-col gap-2.5 pointer-events-auto select-none">
      <div className="flex justify-between items-center">
        {/* Playback Controls & Speed */}
        <div className="flex items-center gap-3">
          <button
            type="button"
            onClick={(e) => {
              e.preventDefault();
              e.stopPropagation();
              onTogglePlay();
            }}
            className={`px-3.5 py-1.5 rounded-md text-xs font-semibold flex items-center gap-2 border transition-all cursor-pointer ${
              isPlaying
                ? 'bg-[#891B1D]/40 border-[#891B1D] text-white shadow-[0_0_12px_rgba(137,27,29,0.5)]'
                : 'bg-[#24394C] border-[#C9B188] hover:bg-[#C9B188] hover:text-[#080D0E] text-white'
            }`}
          >
            <span>{isPlaying ? '❚❚' : '▶'}</span>
            <span className="tracking-wider">{isPlaying ? 'PAUSE' : 'PLAY'}</span>
          </button>

          {/* Speed Multipliers */}
          <div className="flex items-center gap-1 bg-[#080D0E]/70 p-0.5 rounded border border-[#47525F]/30">
            {[1, 10, 50].map((s) => (
              <button
                type="button"
                key={s}
                onClick={(e) => {
                  e.preventDefault();
                  e.stopPropagation();
                  onSetSpeed(s);
                }}
                className={`px-2.5 py-1 rounded text-[11px] font-semibold mono transition-colors cursor-pointer ${
                  speed === s
                    ? 'bg-[#24394C] text-[#FCE873] border border-[#FCE873]/50 shadow-sm'
                    : 'text-[#47525F] hover:text-white'
                }`}
              >
                {s}×
              </button>
            ))}
          </div>
        </div>

        {/* Telemetry Timeline Readout */}
        <div className="text-xs text-[#C9B188] font-medium mono tracking-wide">
          <span className="text-[#47525F] uppercase text-[10px] mr-2">MISSION ELAPSED:</span>
          <span className="text-white font-bold">T+{String(Math.floor(missionHour)).padStart(3, '0')}h</span>
          <span className="text-[#47525F] mx-1.5">/</span>
          <span>336h (DAY {currentDay} OF {totalDays})</span>
        </div>
      </div>

      {/* Scrub Bar Slider */}
      <div className="w-full flex items-center">
        <input
          type="range"
          min="0"
          max="336"
          step="1"
          value={missionHour}
          onChange={(e) => onScrub(Number(e.target.value))}
          className="w-full cursor-pointer accent-[#FCE873]"
        />
      </div>
    </div>
  );
};
