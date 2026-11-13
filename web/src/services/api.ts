import {
  Preset,
  EvaluationRequest,
  EvaluationResult,
  RecommenderRequest,
  RecommenderResponse,
  SavedScenario,
  HorizonProfile
} from '../types';

const API_BASE = '/api/v1';

// Default Artemis fallback presets for offline/independent viewing
export const DEFAULT_PRESETS: Preset[] = [
  {
    id: "malapert-mountain",
    name: "Malapert Mountain",
    sub_region: "Malapert Massif",
    latitude: -86.04,
    longitude: 2.71,
    elevation_m: 5050.0,
    average_slope_deg: 4.1,
    distance_to_psr_km: 1.82,
    description: "Massive volcanic/impact uplift plateau. Prime candidate for Peak of Eternal Light.",
    scientific_significance: "Optimal long-term solar power availability and direct vantage view of Earth."
  },
  {
    id: "shackleton-connecting-ridge",
    name: "Shackleton Connecting Ridge",
    sub_region: "South Pole Peak",
    latitude: -89.45,
    longitude: 120.12,
    elevation_m: 1240.0,
    average_slope_deg: 4.8,
    distance_to_psr_km: 0.85,
    description: "Elevated ridge between Shackleton and de Gerlache craters. Features high cumulative illumination.",
    scientific_significance: "Immediate access to cryogenic volatiles in Shackleton interior."
  },
  {
    id: "faustini-rim-a",
    name: "Faustini Rim Plateau",
    sub_region: "Faustini Crater Rim",
    latitude: -87.15,
    longitude: 84.30,
    elevation_m: 2100.0,
    average_slope_deg: 6.1,
    distance_to_psr_km: 1.40,
    description: "Sunlit ridge overlooking deep permanently shadowed cold traps in Faustini.",
    scientific_significance: "Direct exploration of ancient pre-Nectarian volatile reservoirs."
  },
  {
    id: "de-gerlache-rim-1",
    name: "de Gerlache Mountain",
    sub_region: "de Gerlache Crater",
    latitude: -88.50,
    longitude: -88.20,
    elevation_m: 1800.0,
    average_slope_deg: 5.2,
    distance_to_psr_km: 1.10,
    description: "Broad ridge with consistent seasonal illumination and gentle descent slopes.",
    scientific_significance: "Impact melt sheets and volatile stratigraphy."
  },
  {
    id: "nobile-crater-rim-1",
    name: "Nobile Ridge",
    sub_region: "Nobile Crater",
    latitude: -85.20,
    longitude: 53.50,
    elevation_m: 3200.0,
    average_slope_deg: 4.9,
    distance_to_psr_km: 2.50,
    description: "Broad ridge selected for VIPER rover target science, moderate slopes.",
    scientific_significance: "Direct rover traversability into volatile-rich terrain."
  }
];

export async function fetchHealth(): Promise<{ status: string; uptime_seconds: number } | null> {
  try {
    const controller = new AbortController();
    const timeout = setTimeout(() => controller.abort(), 2000);
    const res = await fetch(`${API_BASE}/health`, { signal: controller.signal });
    clearTimeout(timeout);
    if (res.ok) {
      const data = await res.json();
      if (data && data.status === 'ready') return data;
    }
  } catch (e) {
    // Server is down or unreachable
  }
  return null;
}

export async function fetchPresets(): Promise<Preset[]> {
  try {
    const res = await fetch(`${API_BASE}/presets`);
    if (res.ok) {
      const data = await res.json();
      if (Array.isArray(data) && data.length > 0) return data;
    }
  } catch (e) {
    console.warn("Using offline presets fallback", e);
  }
  return DEFAULT_PRESETS;
}

export async function fetchHorizon(lat: number, lon: number, mastHeightM: number = 2.5): Promise<HorizonProfile | null> {
  try {
    const res = await fetch(`${API_BASE}/horizon`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ latitude: lat, longitude: lon, mast_height_m: mastHeightM }),
    });
    if (res.ok) return await res.json();
  } catch (e) {
    console.warn("Fetch horizon failed", e);
  }
  return null;
}

export async function evaluateMission(req: EvaluationRequest): Promise<EvaluationResult | null> {
  try {
    const res = await fetch(`${API_BASE}/evaluate`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(req),
    });
    if (res.ok) return await res.json();
  } catch (e) {
    console.warn("Mission evaluate failed", e);
  }
  return null;
}

export async function recommendSites(req: RecommenderRequest): Promise<RecommenderResponse | null> {
  try {
    const res = await fetch(`${API_BASE}/recommend`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(req),
    });
    if (res.ok) return await res.json();
  } catch (e) {
    console.warn("Recommender query failed", e);
  }
  return null;
}

export async function saveScenario(scenario: SavedScenario): Promise<{ id: string; share_url: string } | null> {
  try {
    const res = await fetch(`${API_BASE}/scenarios`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(scenario),
    });
    if (res.ok) return await res.json();
  } catch (e) {
    console.warn("Save scenario failed", e);
  }
  return null;
}
