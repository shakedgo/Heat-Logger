import { describe, it, beforeEach, mock } from 'node:test';
import assert from 'node:assert/strict';

// Mock localStorage
const storage = {};
const localStorageMock = {
  getItem: (k) => storage[k] ?? null,
  setItem: (k, v) => { storage[k] = String(v); },
  removeItem: (k) => { delete storage[k]; },
};

// Mock navigator.geolocation
let geoResult = null;
let geoError = null;
const geolocationMock = {
  getCurrentPosition: (success, error) => {
    if (geoError) error(geoError);
    else success(geoResult);
  },
};

// Mock fetch
let fetchResponse = null;
let fetchUrl = null;
const fetchMock = async (url) => {
  fetchUrl = url;
  if (!fetchResponse) throw new Error('Network error');
  return {
    ok: true,
    json: async () => fetchResponse,
  };
};

// Set up globals before importing
globalThis.localStorage = localStorageMock;
Object.defineProperty(globalThis, 'navigator', {
  value: { geolocation: geolocationMock },
  writable: true,
  configurable: true,
});
globalThis.fetch = fetchMock;

const { fetchWeatherData, computeEffectiveSunshine, resetWeatherState } = await import('../frontend/src/utils/weather.js');

function clearState() {
  for (const k of Object.keys(storage)) delete storage[k];
  geoResult = null;
  geoError = null;
  fetchResponse = null;
  fetchUrl = null;
}

describe('weather utility', () => {
  beforeEach(clearState);

  it('fetches weather data with valid coordinates', async () => {
    geoResult = { coords: { latitude: 32.08, longitude: 34.78 } };
    fetchResponse = {
      daily: {
        temperature_2m_mean: [22.5],
        sunshine_duration: [36000], // 10 hours in seconds
        sunrise: [new Date(Date.now() - 8 * 3600000).toISOString()],
        sunset: [new Date(Date.now() - 1 * 3600000).toISOString()], // after sunset
      },
    };

    const result = await fetchWeatherData();
    assert.ok(result);
    assert.equal(result.temperature, 23); // rounded from 22.5
    assert.equal(result.sunshineHours, 10.0); // after sunset: full value
    assert.equal(result.source, 'auto');
  });

  it('constructs correct Open-Meteo API URL', async () => {
    // Clear cache from previous test
    resetWeatherState();
    geoResult = { coords: { latitude: 31.5, longitude: 34.5 } };
    fetchResponse = {
      daily: {
        temperature_2m_mean: [20],
        sunshine_duration: [28800],
        sunrise: [new Date(Date.now() - 6 * 3600000).toISOString()],
        sunset: [new Date(Date.now() + 2 * 3600000).toISOString()],
      },
    };

    await fetchWeatherData();
    assert.ok(fetchUrl);
    assert.ok(fetchUrl.includes('latitude=31.5'));
    assert.ok(fetchUrl.includes('longitude=34.5'));
    assert.ok(fetchUrl.includes('temperature_2m_mean'));
    assert.ok(fetchUrl.includes('sunshine_duration'));
    assert.ok(fetchUrl.includes('sunrise'));
    assert.ok(fetchUrl.includes('sunset'));
  });

  it('returns null when geolocation is denied', async () => {
    geoError = { code: 1 }; // PERMISSION_DENIED
    const result = await fetchWeatherData();
    assert.equal(result, null);
    // Should remember the denial
    assert.equal(storage['heatLogger_geoDenied'], 'true');
  });

  it('returns null when API call throws error', async () => {
    geoResult = { coords: { latitude: 32, longitude: 34 } };
    fetchResponse = null; // will cause fetch to throw
    const result = await fetchWeatherData();
    assert.equal(result, null);
  });

  it('returns null when API response is malformed', async () => {
    geoResult = { coords: { latitude: 32, longitude: 34 } };
    fetchResponse = { daily: { temperature_2m_mean: [] } }; // missing data
    const result = await fetchWeatherData();
    assert.equal(result, null);
  });

  it('caches weather data for subsequent calls', async () => {
    geoResult = { coords: { latitude: 32, longitude: 34 } };
    const sunset = new Date(Date.now() - 1 * 3600000).toISOString();
    fetchResponse = {
      daily: {
        temperature_2m_mean: [25],
        sunshine_duration: [43200], // 12h
        sunrise: [new Date(Date.now() - 10 * 3600000).toISOString()],
        sunset: [sunset],
      },
    };

    const first = await fetchWeatherData();
    assert.ok(first);

    // Change fetch response — second call should use cache
    fetchResponse = {
      daily: { temperature_2m_mean: [99], sunshine_duration: [0], sunrise: [sunset], sunset: [sunset] },
    };

    const second = await fetchWeatherData();
    assert.ok(second);
    assert.equal(second.temperature, first.temperature, 'Should return cached temperature');
  });
});

describe('computeEffectiveSunshine', () => {
  it('before sunrise returns 0 sunshine', () => {
    const result = computeEffectiveSunshine({
      forecast: {
        temperature: 18,
        sunshineDurationSeconds: 36000, // 10h
        sunrise: new Date(Date.now() + 2 * 3600000).toISOString(), // sunrise in 2h
        sunset: new Date(Date.now() + 14 * 3600000).toISOString(),
      },
    });
    assert.equal(result.sunshineHours, 0);
    assert.equal(result.temperature, 18);
  });

  it('after sunset returns full sunshine value', () => {
    const result = computeEffectiveSunshine({
      forecast: {
        temperature: 25.4,
        sunshineDurationSeconds: 28800, // 8h
        sunrise: new Date(Date.now() - 14 * 3600000).toISOString(),
        sunset: new Date(Date.now() - 1 * 3600000).toISOString(), // sunset 1h ago
      },
    });
    assert.equal(result.sunshineHours, 8.0);
    assert.equal(result.temperature, 25); // rounded
  });

  it('midday returns approximately half the forecasted sunshine', () => {
    // Set sunrise 6h ago and sunset 6h from now — we're exactly at midday
    const now = Date.now();
    const result = computeEffectiveSunshine({
      forecast: {
        temperature: 22,
        sunshineDurationSeconds: 36000, // 10h
        sunrise: new Date(now - 6 * 3600000).toISOString(),
        sunset: new Date(now + 6 * 3600000).toISOString(),
      },
    });
    // At midday (50% of daylight elapsed) → ~5.0h
    assert.ok(result.sunshineHours >= 4.5 && result.sunshineHours <= 5.5,
      `Expected ~5h, got ${result.sunshineHours}`);
  });
});
