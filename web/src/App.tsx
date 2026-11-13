import { useState, useEffect, useRef, useCallback } from 'react';
import { Header } from './components/Header';
import { LunarViewer3D, LunarViewer3DHandle } from './components/LunarViewer3D';
import { TimeControls } from './components/TimeControls';
import { HorizonSkyDome } from './components/HorizonSkyDome';
import { SiteDrawer } from './components/SiteDrawer';
import {
  fetchHealth,
  fetchPresets,
  evaluateMission,
  recommendSites,
  DEFAULT_PRESETS,
} from './services/api';
import { Preset, EvaluationResult } from './types';

export function App() {
  const [officialPresets, setOfficialPresets] = useState<Preset[]>(DEFAULT_PRESETS);
  const [customSites, setCustomSites] = useState<Preset[]>([]);
  const [recommendedSites, setRecommendedSites] = useState<Preset[]>([]);
  const [selectedIndex, setSelectedIndex] = useState(0);

  // Time & Simulation State
  const [missionStartDate, setMissionStartDate] = useState<Date>(
    new Date('2026-11-15T00:00:00Z')
  );
  const [currentDate, setCurrentDate] = useState<Date>(
    new Date('2026-11-15T00:00:00Z')
  );
  const [isPlaying, setIsPlaying] = useState(false);
  const [speed, setSpeed] = useState(3600); // Default 1 hr / second
  const [launchSlipDays, setLaunchSlipDays] = useState(0);
  const missionDurationHours = 336; // 14 Earth days (lunar daylight period)

  // Modals & Panels State
  const [isHorizonOpen, setIsHorizonOpen] = useState(false);
  const [isPickingMode, setIsPickingMode] = useState(false);
  const [isRecommending, setIsRecommending] = useState(false);
  const [isBackendConnected, setIsBackendConnected] = useState(false);

  // Astrodynamics / Evaluation results
  const [evalResult, setEvalResult] = useState<EvaluationResult | null>(null);

  const viewerRef = useRef<LunarViewer3DHandle | null>(null);
  const speedRef = useRef(speed);
  useEffect(() => { speedRef.current = speed; }, [speed]);

  // Combine all sites into single unified list
  const allPresets = [...officialPresets, ...customSites, ...recommendedSites];
  const activePreset = allPresets[selectedIndex] || DEFAULT_PRESETS[0];

  // Elapsed hours from mission start date
  const missionHour = Math.max(
    0,
    Math.min(
      missionDurationHours,
      (currentDate.getTime() - missionStartDate.getTime()) / (3600 * 1000)
    )
  );

  // 1. Health check heartbeat
  useEffect(() => {
    let isMounted = true;
    async function checkHealth() {
      const health = await fetchHealth();
      if (isMounted) {
        setIsBackendConnected(health !== null && health.status === 'ready');
      }
    }
    checkHealth();
    const interval = setInterval(checkHealth, 4000);
    return () => {
      isMounted = false;
      clearInterval(interval);
    };
  }, []);

  // 2. Fetch presets from backend on mount
  useEffect(() => {
    async function loadPresets() {
      const data = await fetchPresets();
      if (data && data.length > 0) {
        setOfficialPresets(data);
      }
    }
    loadPresets();
  }, []);

  // 3. Re-evaluate mission whenever site, start date, or launch slip changes
  useEffect(() => {
    let cancelled = false;
    async function runEval() {
      if (!activePreset) return;

      const res = await evaluateMission({
        latitude: activePreset.latitude,
        longitude: activePreset.longitude,
        mission_start_utc: missionStartDate.toISOString(),
        duration_hours: missionDurationHours,
        launch_slip_days: launchSlipDays,
        mast_height_m: 2.5,
      });

      if (!cancelled && res) {
        setEvalResult(res);
      }
    }
    runEval();
    return () => { cancelled = true; };
  }, [activePreset.latitude, activePreset.longitude, missionStartDate, launchSlipDays]);

  // 4. Smooth playback loop for real-time and accelerated simulation
  useEffect(() => {
    if (!isPlaying) return;

    const interval = setInterval(() => {
      setCurrentDate((prev) => {
        // Advance time: speed seconds per tick (tick = 50ms = 0.05s)
        const advanceMs = speedRef.current * 50;
        const next = new Date(prev.getTime() + advanceMs);

        // Check if past mission duration
        const maxTime = missionStartDate.getTime() + missionDurationHours * 3600 * 1000;
        if (next.getTime() >= maxTime) {
          setIsPlaying(false);
          return new Date(maxTime);
        }
        return next;
      });
    }, 50);

    return () => clearInterval(interval);
  }, [isPlaying, missionStartDate]);

  // Handle setting arbitrary calendar date
  const handleSetDate = useCallback((newDate: Date) => {
    setCurrentDate(newDate);
    setMissionStartDate(newDate);
  }, []);

  // Handle fine scrubbing within mission duration
  const handleScrubHour = useCallback((hour: number) => {
    const newTime = missionStartDate.getTime() + hour * 3600 * 1000;
    setCurrentDate(new Date(newTime));
  }, [missionStartDate]);

  // Handle Auto-Recommender search
  const handleRunRecommender = useCallback(async (params: {
    minSun: number;
    minDTE: number;
    maxSlope: number;
    maxIce: number;
  }) => {
    setIsRecommending(true);
    const resp = await recommendSites({
      min_solar_illumination_pct: params.minSun,
      min_dte_comm_pct: params.minDTE,
      max_slope_deg: params.maxSlope,
      max_distance_to_psr_km: params.maxIce,
      mission_start_utc: missionStartDate.toISOString(),
      duration_hours: missionDurationHours,
      launch_slip_days: launchSlipDays,
      limit: 5,
    });

    if (resp && resp.recommendations && resp.recommendations.length > 0) {
      const newPresets: Preset[] = resp.recommendations.map((rec) => ({
        id: `rec-${rec.rank}-${rec.cluster_region.toLowerCase()}`,
        name: `#${rec.rank} ${rec.title}`,
        sub_region: rec.cluster_region,
        latitude: rec.latitude,
        longitude: rec.longitude,
        elevation_m: rec.elevation_m,
        average_slope_deg: rec.slope_deg,
        distance_to_psr_km: rec.distance_to_psr_km,
        description: `Pareto Rank ${rec.rank} Candidate: ${rec.badge}`,
        scientific_significance: `Feasibility Score ${rec.feasibility_score.toFixed(1)}/100 · Sun: ${rec.solar_illumination_pct.toFixed(0)}% · DTE: ${rec.dte_comm_pct.toFixed(0)}%`,
        is_custom: true,
      }));

      setRecommendedSites(newPresets);
      // Select the top recommended site
      const newTotal = officialPresets.length + customSites.length;
      setSelectedIndex(newTotal);
    }
    setIsRecommending(false);
  }, [missionStartDate, launchSlipDays, officialPresets.length, customSites.length]);

  // Handle picking custom coordinates on 3D Moon
  const handlePickCoordinates = useCallback((lat: number, lon: number) => {
    const newSite: Preset = {
      id: `custom-${Date.now()}`,
      name: `Custom Target (${lat.toFixed(1)}°S, ${lon.toFixed(1)}°)`,
      sub_region: 'User Explored Zone',
      latitude: lat,
      longitude: lon,
      elevation_m: 2100.0,
      average_slope_deg: 5.5,
      distance_to_psr_km: 2.0,
      description: 'Custom coordinates chosen by user on 3D lunar surface.',
      scientific_significance: 'Exploratory landing coordinate for localized astrodynamics evaluation.',
      is_custom: true,
    };
    setCustomSites((prev) => [...prev, newSite]);
    setSelectedIndex(allPresets.length);
    setIsPickingMode(false);
    setIsHorizonOpen(true); // Open horizon profile for newly picked point
  }, [allPresets.length]);

  return (
    <div className="relative w-screen h-screen overflow-hidden bg-[#020305] text-white">
      {/* 1. Minimal Header (Logo + Title from reference photo) */}
      <Header
        currentDateUTC={currentDate.toUTCString().replace('GMT', 'UTC')}
        isBackendConnected={isBackendConnected}
      />

      {/* 2. Fullscreen 3D Lunar Environment (Three.js WebGL) */}
      <LunarViewer3D
        ref={viewerRef}
        presets={allPresets}
        selectedIndex={selectedIndex}
        onSelectPreset={setSelectedIndex}
        currentDate={currentDate}
        isPickingMode={isPickingMode}
        onPickCoordinates={handlePickCoordinates}
        onExitPickingMode={() => setIsPickingMode(false)}
      />

      {/* 3. Floating Left Site Drawer & Auto-Recommender */}
      <SiteDrawer
        presets={allPresets}
        selectedIndex={selectedIndex}
        onSelectPreset={setSelectedIndex}
        onRunRecommender={handleRunRecommender}
        isRecommending={isRecommending}
        recommendations={recommendedSites}
        isPickingMode={isPickingMode}
        onTogglePickingMode={() => setIsPickingMode((prev) => !prev)}
        onViewSouthPole={() => viewerRef.current?.viewSouthPole()}
        onViewGlobalOrbit={() => viewerRef.current?.viewGlobalOrbit()}
        onViewTargetLock={() => viewerRef.current?.viewTargetLock()}
      />

      {/* 4. Floating Top-Right Horizon Sky Dome & Metrics Dashboard */}
      <HorizonSkyDome
        activePreset={activePreset}
        evalResult={evalResult}
        currentDate={currentDate}
        missionHour={missionHour}
        isOpen={isHorizonOpen}
        onToggleOpen={() => setIsHorizonOpen((prev) => !prev)}
      />

      {/* 5. Minimalist Bottom Time & Astrodynamics Controller (NASA Eyes style) */}
      <TimeControls
        currentDate={currentDate}
        onSetDate={handleSetDate}
        isPlaying={isPlaying}
        onTogglePlay={() => setIsPlaying((prev) => !prev)}
        speed={speed}
        onSetSpeed={setSpeed}
        launchSlipDays={launchSlipDays}
        onSetLaunchSlip={setLaunchSlipDays}
        missionDurationHours={missionDurationHours}
        missionHour={missionHour}
        onScrubHour={handleScrubHour}
      />

      {/* 6. Footer matching user reference image exactly */}
      <div className="fixed bottom-1 left-0 right-0 z-10 pointer-events-none text-center select-none pb-0.5">
        <span className="text-[11px] font-sans text-white/40 tracking-wider">
          NASA SPACE APP CHALLENGE 2026, Gemini team
        </span>
      </div>
    </div>
  );
}

export default App;
