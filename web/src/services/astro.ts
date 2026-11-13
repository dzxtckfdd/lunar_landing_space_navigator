// Astrodynamic calculations matching internal/astro in the Go backend
// Supports arbitrary dates from 2024 to 2035

export const SynodicMonthDays = 29.530588853;
export const MoonObliquityDeg = 1.5424;
export const MeanEarthMoonDistanceKm = 384400.0;
export const MeanSunDistanceKm = 149597870.7;
export const LunarMeanRadiusM = 1737400.0;
export const DegToRad = Math.PI / 180.0;
export const RadToDeg = 180.0 / Math.PI;

/** Convert Date to Julian Date */
export function getJulianDate(t: Date): number {
  let y = t.getUTCFullYear();
  let m = t.getUTCMonth() + 1; // 1-based
  const d = t.getUTCDate();

  if (m <= 2) {
    y -= 1;
    m += 12;
  }

  const a = Math.floor(y / 100);
  const b = 2 - a + Math.floor(a / 4);

  const dayFraction =
    (t.getUTCHours() + t.getUTCMinutes() / 60.0 + t.getUTCSeconds() / 3600.0) / 24.0;

  return (
    Math.floor(365.25 * (y + 4716)) +
    Math.floor(30.6001 * (m + 1)) +
    d +
    dayFraction +
    b -
    1524.5
  );
}

/** Days since J2000.0 (2000-01-01 12:00 UTC) */
export function getDaysSinceJ2000(t: Date): number {
  return getJulianDate(t) - 2451545.0;
}

export interface SubBodyPoint {
  latitudeDeg: number;
  longitudeDeg: number;
  distanceKm: number;
}

/** Calculate selenographic coordinates of subsolar point for any date */
export function getSubSolarPoint(t: Date): SubBodyPoint {
  const d = getDaysSinceJ2000(t);

  // Mean solar longitude on the ecliptic (degrees)
  let meanSunLonDeg = (280.46 + 0.9856474 * d) % 360.0;
  if (meanSunLonDeg < 0) meanSunLonDeg += 360.0;

  // Longitude of the Moon's ascending node (degrees)
  let nodeLonDeg = (125.04 - 0.0529538083 * d) % 360.0;
  if (nodeLonDeg < 0) nodeLonDeg += 360.0;

  const lambdaPrimeRad = (meanSunLonDeg - nodeLonDeg) * DegToRad;

  // Subsolar latitude: oscillates by ±1.5424°
  const subSolarLatDeg = MoonObliquityDeg * Math.sin(lambdaPrimeRad);

  // Subsolar longitude: travels 360° westward every synodic month (~29.53 days)
  let subSolarLonDeg = (180.0 - 360.0 * (d / SynodicMonthDays)) % 360.0;
  if (subSolarLonDeg > 180.0) subSolarLonDeg -= 360.0;
  else if (subSolarLonDeg < -180.0) subSolarLonDeg += 360.0;

  const meanAnomalySun = (357.528 + 0.9856003 * d) * DegToRad;
  const sunDistKm = MeanSunDistanceKm * (1.0 - 0.0167 * Math.cos(meanAnomalySun));

  return {
    latitudeDeg: subSolarLatDeg,
    longitudeDeg: subSolarLonDeg,
    distanceKm: sunDistKm,
  };
}

/** Calculate selenographic coordinates of subearth point (optical & physical libration) */
export function getSubEarthPoint(t: Date): SubBodyPoint {
  const d = getDaysSinceJ2000(t);

  // Mean anomalies and arguments for lunar libration
  const lPrime = (134.963 + 13.064993 * d) * DegToRad; // Moon's mean anomaly
  const f = (93.272 + 13.229350 * d) * DegToRad;       // Moon's argument of latitude

  // Optical libration in latitude (up to ~6.7°) and longitude (up to ~7.9°)
  const libLatDeg = 6.68 * Math.sin(f) + 0.17 * Math.sin(f - 2 * lPrime);
  const libLonDeg = 7.75 * Math.sin(lPrime) - 0.22 * Math.sin(2 * lPrime);

  const meanDistKm = MeanEarthMoonDistanceKm * (1.0 - 0.0549 * Math.cos(lPrime));

  return {
    latitudeDeg: libLatDeg,
    longitudeDeg: libLonDeg,
    distanceKm: meanDistKm,
  };
}

export interface CelestialTopocentric {
  azimuthDeg: number;
  elevationDeg: number;
  distanceKm: number;
}

/**
 * Topocentric apparent Azimuth (0°–360° clockwise from North) and Elevation (-90° to +90°)
 * from observer selenographic coords to celestial body
 */
export function calculateTopocentric(
  obsLatDeg: number,
  obsLonDeg: number,
  obsElevM: number,
  tgtLatDeg: number,
  tgtLonDeg: number,
  tgtDistKm: number
): CelestialTopocentric {
  const obsLatRad = obsLatDeg * DegToRad;
  const obsLonRad = obsLonDeg * DegToRad;
  const tgtLatRad = tgtLatDeg * DegToRad;
  const tgtLonRad = tgtLonDeg * DegToRad;

  const rObsM = LunarMeanRadiusM + obsElevM;
  const rTgtM = tgtDistKm * 1000.0;

  const cosObsLat = Math.cos(obsLatRad);
  const sinObsLat = Math.sin(obsLatRad);
  const cosObsLon = Math.cos(obsLonRad);
  const sinObsLon = Math.sin(obsLonRad);

  const rObsX = rObsM * cosObsLat * cosObsLon;
  const rObsY = rObsM * cosObsLat * sinObsLon;
  const rObsZ = rObsM * sinObsLat;

  const cosTgtLat = Math.cos(tgtLatRad);
  const sinTgtLat = Math.sin(tgtLatRad);
  const cosTgtLon = Math.cos(tgtLonRad);
  const sinTgtLon = Math.sin(tgtLonRad);

  const rTgtX = rTgtM * cosTgtLat * cosTgtLon;
  const rTgtY = rTgtM * cosTgtLat * sinTgtLon;
  const rTgtZ = rTgtM * sinTgtLat;

  const rhoX = rTgtX - rObsX;
  const rhoY = rTgtY - rObsY;
  const rhoZ = rTgtZ - rObsZ;
  const rhoMag = Math.sqrt(rhoX * rhoX + rhoY * rhoY + rhoZ * rhoZ);
  if (rhoMag < 1.0) {
    return { azimuthDeg: 0, elevationDeg: 0, distanceKm: tgtDistKm };
  }

  // Local Zenith unit vector
  const uZ_X = cosObsLat * cosObsLon;
  const uZ_Y = cosObsLat * sinObsLon;
  const uZ_Z = sinObsLat;

  // Local North and East unit vectors
  let uN_X: number, uN_Y: number, uN_Z: number;
  let uE_X: number, uE_Y: number, uE_Z: number;

  if (obsLatDeg <= -89.9999) {
    uN_X = 1.0;
    uN_Y = 0.0;
    uN_Z = 0.0;
    uE_X = 0.0;
    uE_Y = 1.0;
    uE_Z = 0.0;
  } else {
    uN_X = -sinObsLat * cosObsLon;
    uN_Y = -sinObsLat * sinObsLon;
    uN_Z = cosObsLat;

    uE_X = -sinObsLon;
    uE_Y = cosObsLon;
    uE_Z = 0.0;
  }

  const zTopo = rhoX * uZ_X + rhoY * uZ_Y + rhoZ * uZ_Z;
  const xTopo = rhoX * uN_X + rhoY * uN_Y + rhoZ * uN_Z;
  const yTopo = rhoX * uE_X + rhoY * uE_Y + rhoZ * uE_Z;

  let sinElevation = zTopo / rhoMag;
  sinElevation = Math.max(-1.0, Math.min(1.0, sinElevation));
  const elevationDeg = Math.asin(sinElevation) * RadToDeg;

  let azimuthDeg = Math.atan2(yTopo, xTopo) * RadToDeg;
  if (azimuthDeg < 0.0) azimuthDeg += 360.0;

  return {
    azimuthDeg,
    elevationDeg,
    distanceKm: rhoMag / 1000.0,
  };
}
