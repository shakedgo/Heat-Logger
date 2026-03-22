const CACHE_KEY = 'heatLogger_weatherCache';
const CACHE_TTL_MS = 30 * 60 * 1000; // 30 minutes
const GEO_DENIED_KEY = 'heatLogger_geoDenied';
const GEO_TIMEOUT_MS = 10000;

/**
 * Fetch weather data (temperature + sunshine) using geolocation + Open-Meteo.
 * Returns { temperature, sunshineHours, source } or null on failure.
 */
export async function fetchWeatherData() {
  // Check if geo was previously denied
  if (localStorage.getItem(GEO_DENIED_KEY) === 'true') {
    return null;
  }

  // Check cache first
  const cached = getCachedWeather();
  if (cached) {
    // Re-calculate effective sunshine from cached forecast data
    return computeEffectiveSunshine(cached);
  }

  // Get user location
  let position;
  try {
    position = await getUserLocation();
  } catch (err) {
    if (err.code === 1) {
      // PERMISSION_DENIED
      localStorage.setItem(GEO_DENIED_KEY, 'true');
    }
    return null;
  }

  // Fetch from Open-Meteo
  try {
    const { latitude, longitude } = position.coords;
    const data = await fetchOpenMeteo(latitude, longitude);
    if (!data) return null;

    // Cache the raw forecast data
    const cacheEntry = {
      timestamp: Date.now(),
      forecast: data,
    };
    localStorage.setItem(CACHE_KEY, JSON.stringify(cacheEntry));

    return computeEffectiveSunshine(cacheEntry);
  } catch (err) {
    console.error('Weather fetch failed:', err);
    return null;
  }
}

/**
 * Clear the geo-denied flag and weather cache so a fresh attempt can be made.
 */
export function resetWeatherState() {
  localStorage.removeItem(GEO_DENIED_KEY);
  localStorage.removeItem(CACHE_KEY);
}

/**
 * Request user geolocation.
 */
function getUserLocation() {
  return new Promise((resolve, reject) => {
    if (!navigator.geolocation) {
      reject(new Error('Geolocation not supported'));
      return;
    }
    navigator.geolocation.getCurrentPosition(resolve, reject, {
      enableHighAccuracy: false,
      timeout: GEO_TIMEOUT_MS,
      maximumAge: CACHE_TTL_MS,
    });
  });
}

/**
 * Fetch today's weather from Open-Meteo daily API.
 * Returns { temperature, sunshineDurationSeconds, sunrise, sunset } or null.
 */
async function fetchOpenMeteo(lat, lon) {
  const url = `https://api.open-meteo.com/v1/forecast?latitude=${lat}&longitude=${lon}&daily=temperature_2m_mean,sunshine_duration,sunrise,sunset&timezone=auto&forecast_days=1`;
  const res = await fetch(url);
  if (!res.ok) return null;

  const json = await res.json();
  const daily = json?.daily;
  if (!daily) return null;

  const temperature = daily.temperature_2m_mean?.[0];
  const sunshineDurationSeconds = daily.sunshine_duration?.[0];
  const sunrise = daily.sunrise?.[0];
  const sunset = daily.sunset?.[0];

  if (temperature == null || sunshineDurationSeconds == null) return null;

  return { temperature, sunshineDurationSeconds, sunrise, sunset };
}

/**
 * Compute effective sunshine hours based on time of day.
 * - After sunset: use actual daily values (full measurement).
 * - Before sunrise: effective sunshine = 0.
 * - During day: pro-rate by elapsed daylight fraction.
 */
export function computeEffectiveSunshine(cacheEntry) {
  const { forecast } = cacheEntry;
  const { temperature, sunshineDurationSeconds, sunrise, sunset } = forecast;

  const totalSunshineHours = sunshineDurationSeconds / 3600;
  const roundedTemp = Math.round(temperature);

  // If we don't have sunrise/sunset, return full values
  if (!sunrise || !sunset) {
    return {
      temperature: roundedTemp,
      sunshineHours: parseFloat(totalSunshineHours.toFixed(1)),
      source: 'auto',
    };
  }

  const now = new Date();
  const sunriseTime = new Date(sunrise);
  const sunsetTime = new Date(sunset);

  let effectiveSunshine;

  if (now >= sunsetTime) {
    // After sunset: actual measured values
    effectiveSunshine = totalSunshineHours;
  } else if (now <= sunriseTime) {
    // Before sunrise: no sun yet
    effectiveSunshine = 0;
  } else {
    // During the day: pro-rate
    const totalDaylight = sunsetTime - sunriseTime;
    const elapsed = now - sunriseTime;
    const fraction = Math.max(0, Math.min(1, elapsed / totalDaylight));
    effectiveSunshine = totalSunshineHours * fraction;
  }

  return {
    temperature: roundedTemp,
    sunshineHours: parseFloat(effectiveSunshine.toFixed(1)),
    source: 'auto',
  };
}

/**
 * Get cached weather if still valid (within TTL).
 */
function getCachedWeather() {
  try {
    const raw = localStorage.getItem(CACHE_KEY);
    if (!raw) return null;
    const entry = JSON.parse(raw);
    if (Date.now() - entry.timestamp > CACHE_TTL_MS) {
      localStorage.removeItem(CACHE_KEY);
      return null;
    }
    return entry;
  } catch {
    localStorage.removeItem(CACHE_KEY);
    return null;
  }
}
