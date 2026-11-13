import React, { useEffect, useRef, useState } from 'react';
import { Preset, EvaluationResult, HorizonProfile } from '../types';
import { fetchHorizon } from '../services/api';

interface HorizonSkyDomeProps {
  activePreset: Preset;
  evalResult: EvaluationResult | null;
  currentDate: Date;
  missionHour: number;
  isOpen: boolean;
  onToggleOpen: () => void;
}

export const HorizonSkyDome: React.FC<HorizonSkyDomeProps> = ({
  activePreset,
  evalResult,
  missionHour,
  isOpen,
  onToggleOpen,
}) => {
  const canvasRef = useRef<HTMLCanvasElement | null>(null);
  const [horizonProfile, setHorizonProfile] = useState<HorizonProfile | null>(null);
  const [loadingHorizon, setLoadingHorizon] = useState(false);
  const [mastHeight, setMastHeight] = useState(2.5);

  // Fetch true LOLA 360° horizon when preset or mast height changes
  useEffect(() => {
    let cancelled = false;
    async function loadProfile() {
      setLoadingHorizon(true);
      const profile = await fetchHorizon(activePreset.latitude, activePreset.longitude, mastHeight);
      if (!cancelled && profile) {
        setHorizonProfile(profile);
      }
      setLoadingHorizon(false);
    }
    loadProfile();
    return () => { cancelled = true; };
  }, [activePreset.latitude, activePreset.longitude, mastHeight]);

  // Close on Escape key
  useEffect(() => {
    if (!isOpen) return;
    const handleKeyDown = (e: KeyboardEvent) => {
      if (e.key === 'Escape') onToggleOpen();
    };
    window.addEventListener('keydown', handleKeyDown);
    return () => window.removeEventListener('keydown', handleKeyDown);
  }, [isOpen, onToggleOpen]);

  // Current hourly step from evaluation timeline
  const currentStep =
    evalResult && evalResult.timeline_hourly && evalResult.timeline_hourly.length > 0
      ? evalResult.timeline_hourly[Math.min(evalResult.timeline_hourly.length - 1, Math.floor(missionHour))]
      : null;

  // Render 360° Panorama Canvas
  useEffect(() => {
    const canvas = canvasRef.current;
    if (!canvas || !isOpen) return;
    const ctx = canvas.getContext('2d');
    if (!ctx) return;

    const width = canvas.clientWidth;
    const height = canvas.clientHeight;
    canvas.width = width * 2;
    canvas.height = height * 2;

    ctx.save();
    ctx.scale(2, 2);
    ctx.clearRect(0, 0, width, height);

    // Coordinate conversion:
    // Azimuth 0..360 deg -> x: 40px .. (width - 20px)
    // Elevation -4° .. +12° -> y: (height - 30px) .. 20px
    const marginL = 40;
    const marginR = 20;
    const marginT = 25;
    const marginB = 30;
    const plotW = width - marginL - marginR;
    const plotH = height - marginT - marginB;

    const minEl = -4.0;
    const maxEl = 12.0;

    const azToX = (az: number) => marginL + (az / 360.0) * plotW;
    const elToY = (el: number) => marginT + plotH * (1.0 - (el - minEl) / (maxEl - minEl));

    // 1. Grid Background & Elevation Marks
    ctx.strokeStyle = 'rgba(255, 255, 255, 0.08)';
    ctx.lineWidth = 1;
    ctx.fillStyle = 'rgba(255, 255, 255, 0.4)';
    ctx.font = '10px "JetBrains Mono", monospace';
    ctx.textAlign = 'right';
    ctx.textBaseline = 'middle';

    for (let deg = -2; deg <= 10; deg += 2) {
      const y = elToY(deg);
      ctx.beginPath();
      ctx.moveTo(marginL, y);
      ctx.lineTo(width - marginR, y);
      ctx.stroke();

      ctx.fillText(`${deg > 0 ? '+' : ''}${deg}°`, marginL - 6, y);
    }

    // Zero-elevation dashed line (True Local Horizontal Plane)
    const yZero = elToY(0);
    ctx.strokeStyle = 'rgba(201, 177, 136, 0.5)';
    ctx.setLineDash([4, 4]);
    ctx.lineWidth = 1.2;
    ctx.beginPath();
    ctx.moveTo(marginL, yZero);
    ctx.lineTo(width - marginR, yZero);
    ctx.stroke();
    ctx.setLineDash([]);
    ctx.fillStyle = '#C9B188';
    ctx.textAlign = 'left';
    ctx.fillText('0° HORIZON TANGENT', marginL + 8, yZero - 5);

    // 2. Cardinal Azimuth Marks
    const cardinals = [
      { label: '0° N', az: 0 },
      { label: '45° NE', az: 45 },
      { label: '90° E', az: 90 },
      { label: '135° SE', az: 135 },
      { label: '180° S', az: 180 },
      { label: '225° SW', az: 225 },
      { label: '270° W', az: 270 },
      { label: '315° NW', az: 315 },
      { label: '360° N', az: 360 },
    ];

    ctx.strokeStyle = 'rgba(255, 255, 255, 0.08)';
    ctx.fillStyle = 'rgba(255, 255, 255, 0.5)';
    ctx.textAlign = 'center';
    ctx.textBaseline = 'top';

    cardinals.forEach(({ label, az }) => {
      const x = azToX(az);
      ctx.beginPath();
      ctx.moveTo(x, marginT);
      ctx.lineTo(x, height - marginB);
      ctx.stroke();

      ctx.fillText(label, x, height - marginB + 8);
    });

    // 3. Draw 360° LOLA Topographic Horizon Ridge
    if (horizonProfile && horizonProfile.horizon_profile.length > 0) {
      const pts = horizonProfile.horizon_profile;

      // Fill area under the horizon (lunar terrain)
      ctx.beginPath();
      ctx.moveTo(marginL, height - marginB);

      pts.forEach((pt) => {
        const x = azToX(pt.azimuth);
        const y = Math.max(marginT, Math.min(height - marginB, elToY(pt.horizon_elevation_deg)));
        ctx.lineTo(x, y);
      });

      // Wrap back to bottom right and left
      ctx.lineTo(azToX(360), height - marginB);
      ctx.closePath();

      const terrainGrad = ctx.createLinearGradient(0, marginT, 0, height - marginB);
      terrainGrad.addColorStop(0, 'rgba(40, 50, 65, 0.7)');
      terrainGrad.addColorStop(1, 'rgba(12, 18, 24, 0.95)');
      ctx.fillStyle = terrainGrad;
      ctx.fill();

      // Topographic ridge outline
      ctx.beginPath();
      pts.forEach((pt, i) => {
        const x = azToX(pt.azimuth);
        const y = Math.max(marginT, Math.min(height - marginB, elToY(pt.horizon_elevation_deg)));
        if (i === 0) ctx.moveTo(x, y);
        else ctx.lineTo(x, y);
      });
      ctx.strokeStyle = '#c9b188';
      ctx.lineWidth = 1.8;
      ctx.stroke();
    }

    // 4. Draw Solar Trajectory Curve (over 24-48 hours)
    if (evalResult && evalResult.timeline_hourly && evalResult.timeline_hourly.length > 0) {
      ctx.beginPath();
      ctx.strokeStyle = 'rgba(252, 232, 115, 0.3)';
      ctx.lineWidth = 1.5;
      ctx.setLineDash([2, 2]);

      evalResult.timeline_hourly.forEach((step, idx) => {
        const x = azToX(step.sun_azimuth_deg);
        const y = elToY(step.sun_elevation_deg);
        if (idx === 0) ctx.moveTo(x, y);
        else ctx.lineTo(x, y);
      });
      ctx.stroke();
      ctx.setLineDash([]);
    }

    // 5. Draw Current Sun Position
    if (currentStep) {
      const sunX = azToX(currentStep.sun_azimuth_deg);
      const sunY = elToY(currentStep.sun_elevation_deg);
      const isLit = currentStep.is_illuminated;

      // Vertical guideline from Sun down to horizon
      ctx.strokeStyle = isLit ? 'rgba(252, 232, 115, 0.4)' : 'rgba(244, 63, 94, 0.4)';
      ctx.lineWidth = 1;
      ctx.beginPath();
      ctx.moveTo(sunX, sunY);
      ctx.lineTo(sunX, height - marginB);
      ctx.stroke();

      // Glowing Sun Disc
      const sunGlow = ctx.createRadialGradient(sunX, sunY, 2, sunX, sunY, 14);
      sunGlow.addColorStop(0, '#ffffff');
      sunGlow.addColorStop(0.3, isLit ? '#fce873' : '#f87171');
      sunGlow.addColorStop(1, 'rgba(0, 0, 0, 0)');
      ctx.fillStyle = sunGlow;
      ctx.beginPath();
      ctx.arc(sunX, sunY, 14, 0, Math.PI * 2);
      ctx.fill();

      // Sun solid core
      ctx.fillStyle = isLit ? '#fce873' : '#f43f5e';
      ctx.beginPath();
      ctx.arc(sunX, sunY, 5, 0, Math.PI * 2);
      ctx.fill();

      // Sun tooltip text
      ctx.fillStyle = '#ffffff';
      ctx.font = 'bold 10px "JetBrains Mono", monospace';
      ctx.textAlign = 'center';
      ctx.fillText(
        `SUN: ${currentStep.sun_elevation_deg.toFixed(1)}° (${isLit ? 'LIT' : 'SHADOW'})`,
        sunX,
        Math.max(12, sunY - 18)
      );
    }

    // 6. Draw Current Earth Position (DTE Communications)
    if (currentStep) {
      const earthX = azToX(currentStep.earth_azimuth_deg);
      const earthY = elToY(currentStep.earth_elevation_deg);
      const isDTE = currentStep.is_visible_dte;

      // Vertical guideline from Earth
      ctx.strokeStyle = isDTE ? 'rgba(56, 189, 248, 0.4)' : 'rgba(244, 63, 94, 0.4)';
      ctx.lineWidth = 1;
      ctx.beginPath();
      ctx.moveTo(earthX, earthY);
      ctx.lineTo(earthX, height - marginB);
      ctx.stroke();

      // Earth Glowing Disc
      const earthGlow = ctx.createRadialGradient(earthX, earthY, 2, earthX, earthY, 12);
      earthGlow.addColorStop(0, '#e0f2fe');
      earthGlow.addColorStop(0.4, isDTE ? '#38bdf8' : '#fb7185');
      earthGlow.addColorStop(1, 'rgba(0, 0, 0, 0)');
      ctx.fillStyle = earthGlow;
      ctx.beginPath();
      ctx.arc(earthX, earthY, 12, 0, Math.PI * 2);
      ctx.fill();

      // Earth core
      ctx.fillStyle = isDTE ? '#38bdf8' : '#e11d48';
      ctx.beginPath();
      ctx.arc(earthX, earthY, 4, 0, Math.PI * 2);
      ctx.fill();

      // Earth tooltip text
      ctx.fillStyle = '#bae6fd';
      ctx.font = 'bold 10px "JetBrains Mono", monospace';
      ctx.textAlign = 'center';
      ctx.fillText(
        `EARTH: ${currentStep.earth_elevation_deg.toFixed(1)}° (${isDTE ? 'DTE ONLINE' : 'OCCLUDED'})`,
        earthX,
        Math.max(12, earthY - 16)
      );
    }

    ctx.restore();
  }, [horizonProfile, evalResult, currentStep, isOpen]);

  const summary = evalResult?.mission_summary;
  const solarPct = summary ? summary.solar_illumination_pct : 88.0;
  const dtePct = summary ? summary.dte_comm_pct : 74.0;
  const continuousSunH = summary ? summary.continuous_sunlight_max_hours : 210;
  const longestShadowH = summary ? summary.longest_shadow_hours : 24;
  const longestBlackoutH = summary ? summary.longest_comm_blackout_hours : 32;
  const isThermalFreeze = summary ? summary.thermal_failure_triggered : longestShadowH > 36;
  const score = summary ? summary.feasibility_score : 92.0;

  return (
    <>
      {/* Floating Toggle Button (Always accessible, minimal) */}
      <button
        type="button"
        onClick={onToggleOpen}
        className={`fixed top-20 right-6 z-20 flex items-center gap-2 px-4 py-2 rounded-full border shadow-2xl backdrop-blur-xl transition-all cursor-pointer pointer-events-auto ${
          isOpen
            ? 'bg-[#fce873] border-[#fce873] text-black font-bold'
            : 'bg-black/50 hover:bg-black/80 border-white/20 text-white font-medium'
        }`}
      >
        <span>{isOpen ? '✕' : '🏔️'}</span>
        <span className="text-xs uppercase tracking-wider">
          {isOpen ? 'CLOSE HORIZON' : 'HORIZON & METRICS'}
        </span>
      </button>

      {/* Expanded Modal / Drawer */}
      {isOpen && (
        <div className="fixed top-20 right-6 bottom-28 z-20 w-[620px] max-w-[calc(100vw-3rem)] bg-[#070b10]/95 border border-white/20 rounded-3xl shadow-2xl backdrop-blur-2xl flex flex-col overflow-hidden pointer-events-auto text-white animate-in fade-in slide-in-from-right-8 duration-200">
          {/* Header */}
          <div className="px-6 py-4 border-b border-white/10 flex justify-between items-center bg-white/5">
            <div>
              <div className="text-[10px] uppercase font-bold tracking-widest text-[#fce873]">
                TOPOGRAPHIC HORIZON & SOLAR PROFILE
              </div>
              <h2 className="text-lg font-bold text-white tracking-tight mt-0.5">
                {activePreset.name}
              </h2>
              <div className="text-xs text-white/50 font-mono mt-0.5">
                {Math.abs(activePreset.latitude).toFixed(3)}°S, {Math.abs(activePreset.longitude).toFixed(3)}°
                {activePreset.longitude >= 0 ? 'E' : 'W'} · Elev: {Math.round(activePreset.elevation_m)}m
              </div>
            </div>

            {/* Overall Feasibility Score Badge & Close Button */}
            <div className="flex items-center gap-4">
              <div className="text-right">
                <div className="text-[9px] uppercase tracking-wider text-white/40">
                  FEASIBILITY SCORE
                </div>
                <div
                  className={`text-2xl font-mono font-black ${
                    score >= 85 ? 'text-[#fce873]' : score >= 70 ? 'text-amber-400' : 'text-rose-500'
                  }`}
                >
                  {score.toFixed(1)}
                  <span className="text-xs font-normal text-white/40">/100</span>
                </div>
              </div>

              {/* Explicit Close Button */}
              <button
                type="button"
                onClick={onToggleOpen}
                className="w-8 h-8 rounded-full bg-white/10 hover:bg-white/20 border border-white/20 text-white/70 hover:text-white flex items-center justify-center transition-colors cursor-pointer text-sm font-bold shadow-lg"
                title="Close Window (Esc)"
              >
                ✕
              </button>
            </div>
          </div>

          {/* 360° Panorama Canvas Section */}
          <div className="p-4 flex-1 flex flex-col min-h-[220px]">
            <div className="flex justify-between items-center mb-1">
              <span className="text-[10px] uppercase font-semibold tracking-wider text-white/60">
                360° LOLA Terrain Silhouette & Celestial Sky Dome
              </span>
              <div className="flex items-center gap-2">
                <label className="text-[10px] text-white/40 uppercase">Mast Height:</label>
                <select
                  value={mastHeight}
                  onChange={(e) => setMastHeight(Number(e.target.value))}
                  className="bg-[#0b131a] text-white border border-white/20 rounded-md px-2 py-0.5 text-[11px] font-mono focus:outline-none focus:border-[#fce873] cursor-pointer"
                  style={{ colorScheme: 'dark' }}
                >
                  <option value={1.0} style={{ backgroundColor: '#0b131a', color: '#ffffff' }}>1.0m (Rover)</option>
                  <option value={2.5} style={{ backgroundColor: '#0b131a', color: '#ffffff' }}>2.5m (Lander)</option>
                  <option value={5.0} style={{ backgroundColor: '#0b131a', color: '#ffffff' }}>5.0m (Mast)</option>
                  <option value={10.0} style={{ backgroundColor: '#0b131a', color: '#ffffff' }}>10.0m (Tower)</option>
                </select>
              </div>
            </div>

            {/* Canvas */}
            <div className="relative flex-1 w-full bg-black/60 border border-white/10 rounded-2xl overflow-hidden">
              <canvas ref={canvasRef} className="w-full h-full block" />
              {loadingHorizon && (
                <div className="absolute inset-0 bg-black/70 flex items-center justify-center text-xs text-[#fce873] font-mono">
                  Raymarching 360° LOLA elevation grid...
                </div>
              )}
            </div>

            {/* Current Realtime Status Badges */}
            <div className="mt-2.5 flex items-center justify-between gap-2">
              <div
                className={`flex-1 px-3 py-1.5 rounded-xl border text-xs font-mono font-semibold flex items-center gap-2 ${
                  currentStep?.is_illuminated
                    ? 'bg-amber-500/15 border-amber-500/40 text-[#fce873]'
                    : 'bg-rose-500/15 border-rose-500/40 text-rose-300'
                }`}
              >
                <span>{currentStep?.is_illuminated ? '☀️' : '🌑'}</span>
                <span>
                  {currentStep?.is_illuminated
                    ? `SUN ELEV: +${currentStep.sun_elevation_deg.toFixed(1)}° (DIRECT SUNLIGHT)`
                    : `SUN ELEV: ${currentStep?.sun_elevation_deg.toFixed(1)}° (IN SHADOW)`}
                </span>
              </div>

              <div
                className={`flex-1 px-3 py-1.5 rounded-xl border text-xs font-mono font-semibold flex items-center gap-2 ${
                  currentStep?.is_visible_dte
                    ? 'bg-sky-500/15 border-sky-500/40 text-sky-300'
                    : 'bg-rose-500/15 border-rose-500/40 text-rose-300'
                }`}
              >
                <span>{currentStep?.is_visible_dte ? '🌍' : '🚫'}</span>
                <span>
                  {currentStep?.is_visible_dte
                    ? `EARTH: +${currentStep.earth_elevation_deg.toFixed(1)}° (DTE ONLINE)`
                    : `EARTH: ${currentStep?.earth_elevation_deg.toFixed(1)}° (OCCLUDED)`}
                </span>
              </div>
            </div>
          </div>

          {/* Mission Planner & Scientist Metrics Grid */}
          <div className="px-6 py-4 bg-white/5 border-t border-white/10">
            <div className="text-[10px] uppercase font-bold tracking-wider text-white/50 mb-2">
              Mission Critical Telemetry & Environmental Metrics
            </div>

            <div className="grid grid-cols-4 gap-2">
              {/* Solar % */}
              <div className="bg-white/5 p-2.5 rounded-xl border border-white/5">
                <div className="text-lg font-mono font-bold text-[#fce873] leading-none">
                  {solarPct.toFixed(1)}%
                </div>
                <div className="text-[9px] uppercase tracking-wider text-white/40 mt-1">
                  Solar Illumination
                </div>
              </div>

              {/* Continuous Daylight */}
              <div className="bg-white/5 p-2.5 rounded-xl border border-white/5">
                <div className="text-lg font-mono font-bold text-white leading-none">
                  {continuousSunH}h
                </div>
                <div className="text-[9px] uppercase tracking-wider text-white/40 mt-1">
                  Continuous Sun
                </div>
              </div>

              {/* Longest Night & Thermal Status */}
              <div
                className={`p-2.5 rounded-xl border ${
                  isThermalFreeze
                    ? 'bg-rose-500/20 border-rose-500/40'
                    : 'bg-white/5 border-white/5'
                }`}
              >
                <div
                  className={`text-lg font-mono font-bold leading-none ${
                    isThermalFreeze ? 'text-rose-400' : 'text-white'
                  }`}
                >
                  {longestShadowH}h
                </div>
                <div className="text-[9px] uppercase tracking-wider text-white/40 mt-1">
                  {isThermalFreeze ? 'CRITICAL FREEZE' : 'Max Shadow'}
                </div>
              </div>

              {/* DTE Comm % */}
              <div className="bg-white/5 p-2.5 rounded-xl border border-white/5">
                <div className="text-lg font-mono font-bold text-sky-400 leading-none">
                  {dtePct.toFixed(1)}%
                </div>
                <div className="text-[9px] uppercase tracking-wider text-white/40 mt-1">
                  DTE Link Windows
                </div>
              </div>

              {/* Comm Blackout */}
              <div className="bg-white/5 p-2.5 rounded-xl border border-white/5">
                <div className="text-lg font-mono font-bold text-white leading-none">
                  {longestBlackoutH}h
                </div>
                <div className="text-[9px] uppercase tracking-wider text-white/40 mt-1">
                  Max Blackout
                </div>
              </div>

              {/* Terrain Slope */}
              <div className="bg-white/5 p-2.5 rounded-xl border border-white/5">
                <div
                  className={`text-lg font-mono font-bold leading-none ${
                    activePreset.average_slope_deg <= 8.0 ? 'text-emerald-400' : 'text-amber-400'
                  }`}
                >
                  {activePreset.average_slope_deg.toFixed(1)}°
                </div>
                <div className="text-[9px] uppercase tracking-wider text-white/40 mt-1">
                  {activePreset.average_slope_deg <= 10.0 ? 'Slope (Safe)' : 'Slope (Caution)'}
                </div>
              </div>

              {/* Distance to Water Ice PSR */}
              <div className="bg-white/5 p-2.5 rounded-xl border border-white/5">
                <div className="text-lg font-mono font-bold text-cyan-300 leading-none">
                  {activePreset.distance_to_psr_km.toFixed(2)}km
                </div>
                <div className="text-[9px] uppercase tracking-wider text-white/40 mt-1">
                  Distance to Ice
                </div>
              </div>

              {/* Regional Elevation */}
              <div className="bg-white/5 p-2.5 rounded-xl border border-white/5">
                <div className="text-lg font-mono font-bold text-white leading-none">
                  +{Math.round(activePreset.elevation_m)}m
                </div>
                <div className="text-[9px] uppercase tracking-wider text-white/40 mt-1">
                  Surface Elev
                </div>
              </div>
            </div>

            {/* Scientific notes */}
            <div className="mt-3 text-[11px] text-white/70 leading-relaxed italic bg-black/40 p-2.5 rounded-xl border border-white/5">
              "{activePreset.scientific_significance}"
            </div>
          </div>
        </div>
      )}
    </>
  );
};
