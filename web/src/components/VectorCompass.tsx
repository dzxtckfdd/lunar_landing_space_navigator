import React from 'react';

interface VectorCompassProps {
  sunAzimuth: number;
  sunElevation: number;
}

export const VectorCompass: React.FC<VectorCompassProps> = ({ sunAzimuth, sunElevation }) => {
  return (
    <div className="bg-[#0a0f12]/85 border border-[#1e2d3a]/60 rounded-lg p-2 px-3 flex items-center gap-3 backdrop-blur-xl pointer-events-auto">
      {/* Compass dial */}
      <div className="w-9 h-9 rounded-full border border-[#24394C]/60 relative flex justify-center items-center">
        {/* Cardinal ticks */}
        {[0, 90, 180, 270].map((deg) => (
          <div
            key={deg}
            className="absolute w-px h-1.5 bg-[#47525F]/50"
            style={{
              transform: `rotate(${deg}deg) translateY(-15px)`,
              transformOrigin: 'center center',
            }}
          />
        ))}
        {/* Sun needle */}
        <div
          className="absolute w-px h-3.5 origin-bottom"
          style={{
            transform: `rotate(${sunAzimuth}deg)`,
            background: 'linear-gradient(to top, transparent, #FCE873)',
            bottom: '50%',
            left: 'calc(50% - 0.5px)',
          }}
        />
        <div className="w-1 h-1 rounded-full bg-[#C9B188]" />
      </div>

      {/* Readout */}
      <div className="mono">
        <div className="text-[8px] uppercase tracking-widest text-[#47525F] leading-none">
          Solar Vector
        </div>
        <div className="text-[11px] font-medium text-white leading-tight mt-0.5">
          {sunAzimuth.toFixed(1)}° <span className="text-[#47525F]">AZ</span>{' '}
          {sunElevation >= 0 ? '+' : ''}{sunElevation.toFixed(2)}° <span className="text-[#47525F]">EL</span>
        </div>
      </div>
    </div>
  );
};
