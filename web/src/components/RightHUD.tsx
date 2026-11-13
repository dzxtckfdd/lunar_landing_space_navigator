import React, { useEffect, useRef } from 'react';
import { Preset, EvaluationResult } from '../types';

interface RightHUDProps {
  activePreset: Preset;
  evalResult: EvaluationResult | null;
  missionHour: number;
  onSaveScenario: () => void;
}

export const RightHUD: React.FC<RightHUDProps> = ({
  activePreset,
  evalResult,
  missionHour,
  onSaveScenario,
}) => {
  const canvasRef = useRef<HTMLCanvasElement | null>(null);

  const solarPct = evalResult ? evalResult.mission_summary.solar_illumination_pct : 88.0;
  const dtePct = evalResult ? evalResult.mission_summary.dte_comm_pct : 72.0;
  const slopeDeg = evalResult ? evalResult.slope_deg : activePreset.average_slope_deg;
  const distIceKm = activePreset.distance_to_psr_km;
  const score = evalResult ? evalResult.mission_summary.feasibility_score : 91.5;
  const thermalDeath = evalResult ? evalResult.mission_summary.thermal_failure_triggered : false;

  // Draw 360° horizon canvas
  useEffect(() => {
    const canvas = canvasRef.current;
    if (!canvas) return;
    const ctx = canvas.getContext('2d');
    if (!ctx) return;

    const dw = canvas.clientWidth || 276;
    const dh = canvas.clientHeight || 48;
    canvas.width = dw * 2;
    canvas.height = dh * 2;

    ctx.save();
    ctx.scale(2, 2);
    ctx.clearRect(0, 0, dw, dh);

    // Zero-line
    ctx.strokeStyle = 'rgba(71, 82, 95, 0.4)';
    ctx.lineWidth = 0.5;
    ctx.setLineDash([3, 3]);
    ctx.beginPath();
    ctx.moveTo(0, dh * 0.6);
    ctx.lineTo(dw, dh * 0.6);
    ctx.stroke();
    ctx.setLineDash([]);

    // Horizon fill
    ctx.fillStyle = 'rgba(36, 57, 76, 0.4)';
    ctx.strokeStyle = '#C9B188';
    ctx.lineWidth = 1.2;

    ctx.beginPath();
    ctx.moveTo(0, dh);
    for (let az = 0; az <= 360; az += 2) {
      const x = (az / 360.0) * dw;
      const rim =
        Math.sin((az + activePreset.longitude) * (Math.PI / 60.0)) * 3.5 +
        Math.cos(az * (Math.PI / 30.0)) * 1.5 +
        activePreset.average_slope_deg * 0.4;
      ctx.lineTo(x, dh * 0.6 - rim * 2.5);
    }
    ctx.lineTo(dw, dh);
    ctx.closePath();
    ctx.fill();
    ctx.stroke();

    // Sun marker
    const sunAz = ((missionHour * 360.0) / 336.0) % 360.0;
    const sunX = (sunAz / 360.0) * dw;
    ctx.fillStyle = '#FCE873';
    ctx.beginPath();
    ctx.arc(sunX, dh * 0.6 - 6, 3, 0, Math.PI * 2);
    ctx.fill();

    ctx.restore();
  }, [activePreset, missionHour, evalResult]);

  return (
    <div className="w-[300px] bg-[#0a0f12]/85 border border-[#1e2d3a]/60 rounded-lg backdrop-blur-xl shadow-2xl pointer-events-auto">
      {/* Site header */}
      <div className="px-4 pt-3 pb-2 border-b border-[#1e2d3a]/60">
        <div className="text-[8px] uppercase tracking-[0.16em] text-[#891B1D] font-bold">
          ARTEMIS CANDIDATE
        </div>
        <div className="text-[15px] font-semibold text-white leading-tight mt-0.5">
          {activePreset.name}
        </div>
        <div className="text-[10px] text-[#47525F] mono mt-0.5">
          {Math.abs(activePreset.latitude).toFixed(2)}°S, {Math.abs(activePreset.longitude).toFixed(2)}°{activePreset.longitude >= 0 ? 'E' : 'W'} · {Math.round(activePreset.elevation_m)}m
        </div>
      </div>

      {/* KPI Grid */}
      <div className="grid grid-cols-2 gap-px bg-[#1e2d3a]/30 mx-px">
        {[
          { value: solarPct.toFixed(1), unit: '%', label: 'Solar' },
          { value: dtePct.toFixed(1), unit: '%', label: 'DTE Link' },
          { value: slopeDeg.toFixed(1), unit: '°', label: 'Slope' },
          { value: distIceKm.toFixed(1), unit: 'km', label: 'Ice Dist' },
        ].map((kpi) => (
          <div key={kpi.label} className="bg-[#0a0f12]/90 px-3 py-2">
            <div className="text-[15px] font-bold text-[#FCE873] mono leading-none">
              {kpi.value}<span className="text-[10px] text-[#47525F] ml-0.5">{kpi.unit}</span>
            </div>
            <div className="text-[9px] text-[#47525F] mt-0.5 uppercase tracking-wider">{kpi.label}</div>
          </div>
        ))}
      </div>

      {/* Horizon canvas */}
      <div className="px-3 pt-3 pb-2">
        <div className="flex justify-between items-center text-[9px] uppercase font-medium text-[#47525F] mb-1.5 tracking-wider">
          <span>360° Horizon</span>
          <span className="mono text-[#C9B188]">SUN {((missionHour * 360) / 336).toFixed(0)}°</span>
        </div>
        <canvas ref={canvasRef} className="w-full h-12 bg-[#060a0c] rounded block" />
      </div>

      {/* Score */}
      <div className="mx-3 mb-2 bg-[#24394C]/20 border border-[#24394C]/40 rounded p-2.5 flex justify-between items-center">
        <div>
          <div className="text-[9px] uppercase font-medium tracking-wider text-[#C9B188]">
            Feasibility
          </div>
          <div className="text-[10px] text-[#47525F] mt-0.5">
            {thermalDeath ? '⚠ Shadow Freeze Risk' : 'Nominal'}
          </div>
        </div>
        <div className={`text-2xl font-bold mono ${thermalDeath ? 'text-[#891B1D]' : 'text-[#FCE873]'}`}>
          {score.toFixed(1)}
        </div>
      </div>

      {/* Bookmark button */}
      <div className="px-3 pb-3">
        <button
          onClick={onSaveScenario}
          className="w-full py-1.5 bg-transparent border border-[#24394C]/50 hover:border-[#C9B188]/60 text-[#47525F] hover:text-[#C9B188] text-[10px] font-medium tracking-widest uppercase rounded transition-all cursor-pointer"
        >
          Bookmark Scenario
        </button>
      </div>
    </div>
  );
};
