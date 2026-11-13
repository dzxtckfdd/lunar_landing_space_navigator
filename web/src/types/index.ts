export interface Preset {
  id: string;
  name: string;
  sub_region: string;
  latitude: number;
  longitude: number;
  elevation_m: number;
  average_slope_deg: number;
  distance_to_psr_km: number;
  description: string;
  scientific_significance: string;
  is_custom?: boolean;
}

export interface CandidateSite {
  id: number;
  cluster_region: string;
  latitude: number;
  longitude: number;
  elevation_m: number;
  slope_deg: number;
  distance_to_psr_km: number;
  is_safe_slope: boolean;
}

export interface HorizonPoint {
  azimuth: number;
  horizon_elevation_deg: number;
  distance_m: number;
  peak_elevation_m: number;
}

export interface HorizonProfile {
  observer_lat: number;
  observer_lon: number;
  observer_elevation_m: number;
  mast_height_m: number;
  horizon_profile: HorizonPoint[];
  min_horizon_elevation_deg: number;
  max_horizon_elevation_deg: number;
  computation_time_ms: number;
}

export interface EvaluationRequest {
  latitude: number;
  longitude: number;
  mast_height_m?: number;
  mission_start_utc?: string;
  duration_hours?: number;
  launch_slip_days?: number;
  solar_panel_type?: string;
  max_shadow_freeze_hours?: number;
}

export interface HourlyStep {
  time_offset_hours: number;
  timestamp_utc: string;
  sun_azimuth_deg: number;
  sun_elevation_deg: number;
  sun_horizon_mask_deg: number;
  is_illuminated: boolean;
  solar_flux_rel: number;
  earth_azimuth_deg: number;
  earth_elevation_deg: number;
  earth_horizon_mask_deg: number;
  is_visible_dte: boolean;
}

export interface MissionSummary {
  mission_start_utc: string;
  mission_end_utc: string;
  total_hours: number;
  illuminated_hours: number;
  solar_illumination_pct: number;
  continuous_sunlight_max_hours: number;
  longest_shadow_hours: number;
  thermal_failure_triggered: boolean;
  dte_connected_hours: number;
  dte_comm_pct: number;
  longest_comm_blackout_hours: number;
  feasibility_score: number;
}

export interface EvaluationResult {
  latitude: number;
  longitude: number;
  elevation_m: number;
  slope_deg: number;
  mission_summary: MissionSummary;
  timeline_hourly: HourlyStep[];
  computation_time_ms: number;
}

export interface RecommenderRequest {
  min_solar_illumination_pct?: number;
  min_dte_comm_pct?: number;
  max_slope_deg?: number;
  max_distance_to_psr_km?: number;
  mission_start_utc?: string;
  duration_hours?: number;
  launch_slip_days?: number;
  preferred_region?: string;
  limit?: number;
  mast_height_m?: number;
  solar_panel_type?: string;
  max_shadow_freeze_hours?: number;
}

export interface RecommendedSite {
  rank: number;
  title: string;
  cluster_region: string;
  latitude: number;
  longitude: number;
  elevation_m: number;
  slope_deg: number;
  distance_to_psr_km: number;
  solar_illumination_pct: number;
  continuous_sunlight_hours: number;
  dte_comm_pct: number;
  longest_blackout_hours: number;
  feasibility_score: number;
  badge: string;
}

export interface RecommenderResponse {
  candidates_scanned: number;
  matches_found: number;
  recommendations: RecommendedSite[];
  computation_time_ms: number;
}

export interface SavedScenario {
  id?: string;
  title: string;
  latitude: number;
  longitude: number;
  antenna_mast_height_m: number;
  mission_start_utc: string;
  mission_duration_hours: number;
  launch_slip_days: number;
  solar_illumination_pct: number;
  dte_comm_pct: number;
  feasibility_score: number;
  notes?: string;
  created_at?: string;
}
