import React from 'react';

interface HeaderProps {
  currentDateUTC: string;
  isBackendConnected: boolean;
  onResetView?: () => void;
}

export const Header: React.FC<HeaderProps> = ({
  currentDateUTC,
  isBackendConnected,
}) => {
  return (
    <header className="fixed top-0 left-0 right-0 z-20 pointer-events-none flex flex-col items-center pt-4 px-6 select-none">
      {/* Brand matching user image reference */}
      <div className="pointer-events-auto flex items-center gap-3 bg-black/40 backdrop-blur-md px-5 py-2 rounded-full border border-white/10 shadow-2xl transition-all hover:bg-black/60">
        {/* Minimal Lunar lander icon matching image */}
        <svg
          className="w-7 h-7 text-white shrink-0"
          viewBox="0 0 32 32"
          fill="none"
          stroke="currentColor"
          strokeWidth="1.8"
          strokeLinecap="round"
          strokeLinejoin="round"
        >
          {/* Main cabin */}
          <circle cx="16" cy="11" r="5" />
          <path d="M11 11h10" />
          <path d="M16 6v-2" />
          {/* Legs & Pads */}
          <path d="M10 16l-3.5 10M22 16l3.5 10" />
          <path d="M5 26h5M22 26h5" />
          <path d="M13 16l-1 10M19 16l1 10" />
          {/* Descent stage thruster base */}
          <path d="M11 20h10" />
        </svg>

        <div className="flex items-baseline gap-2">
          <h1 className="text-[19px] font-medium tracking-tight text-white leading-none">
            Lunar Navigator <span className="font-light text-white/80">by Gemini</span>
          </h1>
        </div>

        {/* Subtle status dot */}
        <div
          className="w-1.5 h-1.5 rounded-full ml-1"
          style={{
            backgroundColor: isBackendConnected ? '#4ade80' : '#f87171',
            boxShadow: isBackendConnected
              ? '0 0 8px rgba(74, 222, 128, 0.8)'
              : '0 0 8px rgba(248, 113, 113, 0.8)',
          }}
          title={isBackendConnected ? 'Backend Connected' : 'Offline / Standalone Mode'}
        />
      </div>

      {/* Subtle date readout below */}
      <div className="mt-1 text-[11px] font-mono tracking-wider text-white/50 pointer-events-auto bg-black/20 px-3 py-0.5 rounded-full backdrop-blur-sm">
        {currentDateUTC}
      </div>
    </header>
  );
};
