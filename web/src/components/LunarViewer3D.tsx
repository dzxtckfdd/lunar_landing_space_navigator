import { useEffect, useRef, useCallback, useImperativeHandle, forwardRef } from 'react';
import * as THREE from 'three';
import { OrbitControls } from 'three/examples/jsm/controls/OrbitControls.js';
import { Preset } from '../types';
import { getSubSolarPoint, getSubEarthPoint } from '../services/astro';

export interface LunarViewer3DHandle {
  viewSouthPole: () => void;
  viewGlobalOrbit: () => void;
  viewTargetLock: () => void;
}

interface LunarViewer3DProps {
  presets: Preset[];
  selectedIndex: number;
  onSelectPreset: (index: number) => void;
  currentDate: Date;
  isPickingMode?: boolean;
  onPickCoordinates?: (lat: number, lon: number) => void;
  onExitPickingMode?: () => void;
}

/** Convert lunar selenographic lat/lon (degrees) to 3D point on the Moon sphere (local space). */
function latLonToLocalSphere(latDeg: number, lonDeg: number, radius: number): THREE.Vector3 {
  const phi = THREE.MathUtils.degToRad(latDeg);
  const theta = THREE.MathUtils.degToRad(lonDeg);
  const cosPhi = Math.cos(phi);
  return new THREE.Vector3(
    radius * cosPhi * Math.sin(theta),
    radius * Math.sin(phi),
    radius * cosPhi * Math.cos(theta)
  );
}

/** Convert local 3D point on the Moon sphere back to selenographic lat/lon (degrees). */
function localSphereToLatLon(v: THREE.Vector3): { lat: number; lon: number } {
  const norm = v.clone().normalize();
  const lat = Math.asin(Math.max(-1, Math.min(1, norm.y))) * (180.0 / Math.PI);
  const lon = Math.atan2(norm.x, norm.z) * (180.0 / Math.PI);
  return { lat, lon };
}

/** Construct a high-resolution polar cap mesh using planar polar stereographic UVs to eliminate polar pinch. */
function createPolarCapGeometry(radius: number, minLatDeg: number): THREE.BufferGeometry {
  const rings = 64;
  const sectors = 128;
  const positions: number[] = [];
  const uvs: number[] = [];
  const indices: number[] = [];
  const normals: number[] = [];

  const maxAngleFromPole = THREE.MathUtils.degToRad(90 + minLatDeg); // e.g. for -75°, angle is 15°

  for (let r = 0; r <= rings; r++) {
    const frac = r / rings; // 0 at pole, 1 at boundary
    const angle = frac * maxAngleFromPole;
    const phi = -Math.PI / 2 + angle; // -90° to minLatDeg
    const cosPhi = Math.cos(phi);
    const sinPhi = Math.sin(phi);

    for (let s = 0; s <= sectors; s++) {
      const theta = (s / sectors) * Math.PI * 2;
      const cosTheta = Math.cos(theta);
      const sinTheta = Math.sin(theta);

      // 3D position on sphere surface
      const x = radius * cosPhi * sinTheta;
      const y = radius * sinPhi;
      const z = radius * cosPhi * cosTheta;
      positions.push(x, y, z);

      // Normal pointing outward
      normals.push(cosPhi * sinTheta, sinPhi, cosPhi * cosTheta);

      // Conformal polar stereographic planar UV mapping centered on pole (0.5, 0.5)
      const uvDist = frac * 0.485;
      const u = 0.5 + uvDist * sinTheta;
      const v = 0.5 + uvDist * cosTheta;
      uvs.push(u, v);
    }
  }

  // FRONT-FACING CCW winding order: (a, d, b) and (b, d, c)
  for (let r = 0; r < rings; r++) {
    for (let s = 0; s < sectors; s++) {
      const a = r * (sectors + 1) + s;
      const b = (r + 1) * (sectors + 1) + s;
      const c = (r + 1) * (sectors + 1) + (s + 1);
      const d = r * (sectors + 1) + (s + 1);

      indices.push(a, d, b);
      indices.push(b, d, c);
    }
  }

  const geo = new THREE.BufferGeometry();
  geo.setAttribute('position', new THREE.Float32BufferAttribute(positions, 3));
  geo.setAttribute('normal', new THREE.Float32BufferAttribute(normals, 3));
  geo.setAttribute('uv', new THREE.Float32BufferAttribute(uvs, 2));
  geo.setIndex(indices);
  return geo;
}

export const LunarViewer3D = forwardRef<LunarViewer3DHandle, LunarViewer3DProps>(({
  presets,
  selectedIndex,
  onSelectPreset,
  currentDate,
  isPickingMode = false,
  onPickCoordinates,
  onExitPickingMode,
}, ref) => {
  const mountRef = useRef<HTMLDivElement | null>(null);

  // References for live rendering loop
  const selectedIndexRef = useRef(selectedIndex);
  const presetsRef = useRef(presets);
  const currentDateRef = useRef(currentDate);
  const isPickingModeRef = useRef(isPickingMode);
  const controlsRef = useRef<OrbitControls | null>(null);
  const cameraRef = useRef<THREE.PerspectiveCamera | null>(null);
  const moonMeshRef = useRef<THREE.Mesh | null>(null);
  const sunLightRef = useRef<THREE.DirectionalLight | null>(null);
  const sunGroupRef = useRef<THREE.Group | null>(null);
  const earthMeshRef = useRef<THREE.Mesh | null>(null);
  const currentCamDistanceRef = useRef<number>(5.0);

  useEffect(() => { selectedIndexRef.current = selectedIndex; }, [selectedIndex]);
  useEffect(() => { presetsRef.current = presets; }, [presets]);
  useEffect(() => { currentDateRef.current = currentDate; }, [currentDate]);
  useEffect(() => { isPickingModeRef.current = isPickingMode; }, [isPickingMode]);

  // Smooth camera fly-to centering
  const animateCameraTo = useCallback((targetPos: THREE.Vector3, targetLookAt: THREE.Vector3, durationMs: number = 700) => {
    const controls = controlsRef.current;
    const camera = cameraRef.current;
    if (!controls || !camera) return;

    const startPos = camera.position.clone();
    const startLookAt = controls.target.clone();
    const startTime = performance.now();

    function step() {
      const now = performance.now();
      const elapsed = now - startTime;
      const progress = Math.min(elapsed / durationMs, 1.0);
      const ease = 1 - Math.pow(1 - progress, 3);

      camera!.position.lerpVectors(startPos, targetPos, ease);
      controls!.target.lerpVectors(startLookAt, targetLookAt, ease);
      controls!.update();

      if (progress < 1.0) {
        requestAnimationFrame(step);
      }
    }
    requestAnimationFrame(step);
  }, []);

  // Preset fly-to centering
  const flyToPreset = useCallback((index: number) => {
    const moon = moonMeshRef.current;
    if (!moon) return;

    const preset = presetsRef.current[index];
    if (!preset) return;

    moon.updateMatrixWorld(true);
    const localPos = latLonToLocalSphere(preset.latitude, preset.longitude, 2.6);
    const worldTarget = moon.localToWorld(localPos.clone());

    const moonCenter = moon.position.clone();
    const normal = worldTarget.clone().sub(moonCenter).normalize();
    const targetCamPos = worldTarget.clone().add(normal.multiplyScalar(0.9));

    animateCameraTo(targetCamPos, worldTarget, 650);
  }, [animateCameraTo]);

  const viewSouthPole = useCallback(() => {
    const moon = moonMeshRef.current;
    if (!moon) return;

    moon.updateMatrixWorld(true);
    const localSouthPole = new THREE.Vector3(0, -2.6, 0);
    const worldSouthPole = moon.localToWorld(localSouthPole.clone());

    const moonCenter = moon.position.clone();
    const poleNormal = worldSouthPole.clone().sub(moonCenter).normalize();
    const camPos = worldSouthPole.clone().add(poleNormal.multiplyScalar(1.35));

    animateCameraTo(camPos, worldSouthPole, 800);
  }, [animateCameraTo]);

  const viewGlobalOrbit = useCallback(() => {
    animateCameraTo(new THREE.Vector3(0, 1.8, 5.2), new THREE.Vector3(0, -0.4, 0), 800);
  }, [animateCameraTo]);

  const viewTargetLock = useCallback(() => {
    flyToPreset(selectedIndexRef.current);
  }, [flyToPreset]);

  // Expose camera controls to parent
  useImperativeHandle(ref, () => ({
    viewSouthPole,
    viewGlobalOrbit,
    viewTargetLock,
  }));

  // Fly camera whenever selected index changes
  useEffect(() => {
    flyToPreset(selectedIndex);
  }, [selectedIndex, flyToPreset]);

  // Main Three.js Scene Setup
  useEffect(() => {
    const mount = mountRef.current;
    if (!mount) return;

    // 1. Scene
    const scene = new THREE.Scene();
    scene.background = new THREE.Color(0x020305);

    // 2. Camera (Eyes on the Solar System framing)
    const camera = new THREE.PerspectiveCamera(
      36,
      window.innerWidth / window.innerHeight,
      0.05,
      1000
    );
    camera.position.set(0, 1.8, 5.2);
    cameraRef.current = camera;

    // 3. Renderer with ACES Filmic tone mapping
    const renderer = new THREE.WebGLRenderer({
      antialias: true,
      alpha: false,
      powerPreference: 'high-performance',
    });
    renderer.setSize(window.innerWidth, window.innerHeight);
    renderer.setPixelRatio(Math.min(window.devicePixelRatio, 2));
    renderer.toneMapping = THREE.ACESFilmicToneMapping;
    renderer.toneMappingExposure = 1.15;
    mount.appendChild(renderer.domElement);

    // 4. OrbitControls
    const controls = new OrbitControls(camera, renderer.domElement);
    controls.enableDamping = true;
    controls.dampingFactor = 0.08;
    controls.rotateSpeed = 0.45;
    controls.minDistance = 0.25;
    controls.maxDistance = 18.0;
    controls.target.set(0, -0.3, 0);
    controlsRef.current = controls;

    // 5. Starfield Background (Deep Space)
    const starCount = 3200;
    const starGeo = new THREE.BufferGeometry();
    const starPositions = new Float32Array(starCount * 3);
    for (let i = 0; i < starCount * 3; i += 3) {
      const radius = 140 + Math.random() * 200;
      const theta = Math.random() * Math.PI * 2;
      const phi = Math.acos(2 * Math.random() - 1);
      starPositions[i] = radius * Math.sin(phi) * Math.cos(theta);
      starPositions[i + 1] = radius * Math.sin(phi) * Math.sin(theta);
      starPositions[i + 2] = radius * Math.cos(phi);
    }
    starGeo.setAttribute('position', new THREE.BufferAttribute(starPositions, 3));
    const starMat = new THREE.PointsMaterial({
      color: 0xffffff,
      size: 0.18,
      transparent: true,
      opacity: 0.85,
    });
    scene.add(new THREE.Points(starGeo, starMat));

    // 6. Moon Mesh
    const moonRadius = 2.6;
    const moonGeo = new THREE.SphereGeometry(moonRadius, 128, 128);
    const textureLoader = new THREE.TextureLoader();
    const basePath = (import.meta.env.BASE_URL || '/').replace(/\/$/, '');

    const moonMap = textureLoader.load(`${basePath}/textures/moon_map.jpg`);
    const moonNormal = textureLoader.load(`${basePath}/textures/moon_normal.jpg`);
    moonMap.colorSpace = THREE.SRGBColorSpace;

    const moonMat = new THREE.MeshStandardMaterial({
      map: moonMap,
      normalMap: moonNormal,
      normalScale: new THREE.Vector2(0.85, 0.85),
      roughness: 0.94,
      metalness: 0.04,
    });
    const moonMesh = new THREE.Mesh(moonGeo, moonMat);
    scene.add(moonMesh);
    moonMeshRef.current = moonMesh;

    // 7. High-Resolution LOLA Polar Cap
    const lolaTexture = textureLoader.load(`${basePath}/textures/lola_south_pole.png`);
    lolaTexture.colorSpace = THREE.SRGBColorSpace;

    const polarCapGeo = createPolarCapGeometry(moonRadius + 0.003, -75);
    const polarCapMat = new THREE.MeshStandardMaterial({
      map: lolaTexture,
      roughness: 0.95,
      metalness: 0.02,
      polygonOffset: true,
      polygonOffsetFactor: -1.0,
      polygonOffsetUnits: -2.0,
    });
    const polarCapMesh = new THREE.Mesh(polarCapGeo, polarCapMat);
    moonMesh.add(polarCapMesh);

    // 8. Earth Mesh (with atmosphere and rotation)
    const earthRadius = 0.52;
    const earthDistance = 14.0;
    const earthGeo = new THREE.SphereGeometry(earthRadius, 64, 64);
    const earthTexture = textureLoader.load(`${basePath}/textures/earth_atmos_2048.jpg`);
    earthTexture.colorSpace = THREE.SRGBColorSpace;

    const earthMat = new THREE.MeshStandardMaterial({
      map: earthTexture,
      roughness: 0.65,
      metalness: 0.1,
    });
    const earthMesh = new THREE.Mesh(earthGeo, earthMat);
    scene.add(earthMesh);
    earthMeshRef.current = earthMesh;

    // 9. Sun Group & Lighting
    const sunDistance = 24.0;
    const sunGroup = new THREE.Group();
    const sunDiscRadius = 0.85;
    const sunDiscMat = new THREE.MeshBasicMaterial({ color: 0xffffff });
    const sunDisc = new THREE.Mesh(new THREE.SphereGeometry(sunDiscRadius, 32, 32), sunDiscMat);
    sunGroup.add(sunDisc);

    // Corona flare
    const sunCanvas = document.createElement('canvas');
    sunCanvas.width = 256;
    sunCanvas.height = 256;
    const sCtx = sunCanvas.getContext('2d')!;
    const sGrd = sCtx.createRadialGradient(128, 128, 16, 128, 128, 128);
    sGrd.addColorStop(0, 'rgba(255, 255, 255, 1.0)');
    sGrd.addColorStop(0.2, 'rgba(255, 245, 190, 0.85)');
    sGrd.addColorStop(0.55, 'rgba(255, 200, 100, 0.35)');
    sGrd.addColorStop(1, 'rgba(255, 180, 50, 0)');
    sCtx.fillStyle = sGrd;
    sCtx.fillRect(0, 0, 256, 256);
    const sunGlow = new THREE.Sprite(
      new THREE.SpriteMaterial({
        map: new THREE.CanvasTexture(sunCanvas),
        blending: THREE.AdditiveBlending,
        transparent: true,
      })
    );
    sunGlow.scale.set(7.0, 7.0, 1);
    sunGroup.add(sunGlow);

    scene.add(sunGroup);
    sunGroupRef.current = sunGroup;

    // Directional sunlight shining from the Sun onto the Moon
    const sunLight = new THREE.DirectionalLight(0xfff9f0, 3.8);
    sunLight.position.copy(sunGroup.position);
    scene.add(sunLight);
    sunLightRef.current = sunLight;

    // Deep ambient starlight
    const ambientLight = new THREE.AmbientLight(0x151c24, 0.25);
    scene.add(ambientLight);

    // 10. Interactive Raycaster for picking points
    const raycaster = new THREE.Raycaster();
    const mouse = new THREE.Vector2();

    const handleCanvasClick = (event: MouseEvent) => {
      if (!isPickingModeRef.current) return;

      const rect = renderer.domElement.getBoundingClientRect();
      mouse.x = ((event.clientX - rect.left) / rect.width) * 2 - 1;
      mouse.y = -((event.clientY - rect.top) / rect.height) * 2 + 1;

      raycaster.setFromCamera(mouse, camera);
      if (moonMeshRef.current) {
        const intersects = raycaster.intersectObject(moonMeshRef.current, false);
        if (intersects.length > 0) {
          const hitWorld = intersects[0].point;
          const hitLocal = moonMeshRef.current.worldToLocal(hitWorld.clone());
          const { lat, lon } = localSphereToLatLon(hitLocal);
          if (onPickCoordinates) {
            onPickCoordinates(lat, lon);
          }
        }
      }
    };
    renderer.domElement.addEventListener('click', handleCanvasClick);

    // 11. Main Astrodynamic Animation Loop
    const tempVec = new THREE.Vector3();
    let animId: number;

    const render = () => {
      animId = requestAnimationFrame(render);
      controls.update();

      const camDist = camera.position.distanceTo(controls.target);
      currentCamDistanceRef.current = camDist;

      const curDate = currentDateRef.current;

      // ── Physical Astrodynamics for arbitrary calendar Date ──
      // 1. Subsolar Point (Exact solar illumination of lunar surface)
      const subSolar = getSubSolarPoint(curDate);
      const sunLatRad = subSolar.latitudeDeg * (Math.PI / 180.0);
      const sunLonRad = subSolar.longitudeDeg * (Math.PI / 180.0);

      const sX = sunDistance * Math.cos(sunLatRad) * Math.sin(sunLonRad);
      const sY = sunDistance * Math.sin(sunLatRad);
      const sZ = sunDistance * Math.cos(sunLatRad) * Math.cos(sunLonRad);

      if (sunGroupRef.current) sunGroupRef.current.position.set(sX, sY, sZ);
      if (sunLightRef.current) sunLightRef.current.position.set(sX, sY, sZ);

      // 2. SubEarth Point & Earth Position (with real libration)
      const subEarth = getSubEarthPoint(curDate);
      const earthLatRad = subEarth.latitudeDeg * (Math.PI / 180.0);
      const earthLonRad = subEarth.longitudeDeg * (Math.PI / 180.0);

      // Earth position in space hovering above lunar limb
      const eX = earthDistance * Math.cos(earthLatRad) * Math.sin(earthLonRad);
      const eY = earthDistance * Math.sin(earthLatRad) + 1.8;
      const eZ = earthDistance * Math.cos(earthLatRad) * Math.cos(earthLonRad);
      earthMesh.position.set(eX, eY, eZ);

      // Earth rotation on its own axis based on UTC hour
      const utcHour = curDate.getUTCHours() + curDate.getUTCMinutes() / 60.0 + curDate.getUTCSeconds() / 3600.0;
      earthMesh.rotation.y = (utcHour / 24.0) * Math.PI * 2;

      // 3. Project 3D site markers to 2D screen coordinates
      if (moonMeshRef.current) {
        moonMeshRef.current.updateMatrixWorld(true);
        const selIdx = selectedIndexRef.current;
        const currentPresets = presetsRef.current;

        currentPresets.forEach((p, idx) => {
          const el = document.getElementById(`target-marker-${idx}`);
          if (!el) return;

          const localPos = latLonToLocalSphere(p.latitude, p.longitude, moonRadius);
          const worldPos = moonMeshRef.current!.localToWorld(localPos.clone());

          const camToPoint = worldPos.clone().sub(camera.position);
          const moonToPoint = worldPos.clone().sub(moonMeshRef.current!.position);
          const isFacingCamera = camToPoint.dot(moonToPoint) < 0;

          tempVec.copy(worldPos);
          tempVec.project(camera);

          if (isFacingCamera && tempVec.z < 1.0) {
            const sx = (tempVec.x * 0.5 + 0.5) * window.innerWidth;
            const sy = (-(tempVec.y * 0.5) + 0.5) * window.innerHeight;

            el.style.left = `${sx}px`;
            el.style.top = `${sy}px`;
            el.style.display = 'flex';

            if (idx === selIdx) {
              el.classList.add('active');
            } else {
              el.classList.remove('active');
            }
          } else {
            el.style.display = 'none';
          }
        });
      }

      renderer.render(scene, camera);
    };

    render();

    // 12. Responsive Resize
    const handleResize = () => {
      camera.aspect = window.innerWidth / window.innerHeight;
      camera.updateProjectionMatrix();
      renderer.setSize(window.innerWidth, window.innerHeight);
    };
    window.addEventListener('resize', handleResize);

    return () => {
      cancelAnimationFrame(animId);
      window.removeEventListener('resize', handleResize);
      renderer.domElement.removeEventListener('click', handleCanvasClick);
      renderer.dispose();
      if (mount.contains(renderer.domElement)) {
        mount.removeChild(renderer.domElement);
      }
    };
  }, []);

  return (
    <>
      {/* 3D WebGL Canvas */}
      <div
        ref={mountRef}
        className={`absolute inset-0 z-[1] ${isPickingMode ? 'cursor-crosshair' : 'cursor-grab active:cursor-grabbing'}`}
      />

      {/* Picking Mode Banner */}
      {isPickingMode && (
        <div className="fixed top-24 left-1/2 -translate-x-1/2 z-30 pointer-events-auto flex items-center gap-3 bg-rose-600/90 border border-white/40 px-5 py-2 rounded-full shadow-2xl backdrop-blur-xl animate-pulse">
          <span className="text-xs font-bold text-white tracking-wider uppercase">
            🎯 CLICK ON MOON SURFACE TO SET CUSTOM LANDING SPOT
          </span>
          <button
            type="button"
            onClick={onExitPickingMode}
            className="text-[11px] bg-black/40 hover:bg-black/60 text-white px-2.5 py-0.5 rounded cursor-pointer font-semibold"
          >
            CANCEL
          </button>
        </div>
      )}

      {/* 3D Target Markers (Matching User Uploaded Image) */}
      <div className="absolute inset-0 z-[4] pointer-events-none">
        {presets.map((preset, index) => {
          const isSelected = index === selectedIndex;
          return (
            <div
              key={preset.id}
              id={`target-marker-${index}`}
              onClick={(e) => {
                e.stopPropagation();
                onSelectPreset(index);
              }}
              className={`lunar-target-marker ${isSelected ? 'active' : ''}`}
              title={`${preset.name} (${Math.abs(preset.latitude).toFixed(2)}° S, ${Math.abs(preset.longitude).toFixed(2)}°)`}
            >
              <div className="inner-white-ring">
                <div className="center-red-dot" />
              </div>
              {isSelected && (
                <div className="marker-label">
                  <span>{preset.name}</span>
                </div>
              )}
            </div>
          );
        })}
      </div>
    </>
  );
});
