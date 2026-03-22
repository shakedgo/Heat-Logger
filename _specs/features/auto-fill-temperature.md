# Spec for Auto-Fill Weather Data (Temperature + Sunshine)

branch: claude/feature/auto-fill-temperature
figma_component (if used): N/A

## Summary

Auto-fill weather data in the prediction form using the browser's Geolocation API and the Open-Meteo weather API (free, no API key required). Two values are fetched: the **average daily temperature** and **sunshine hours**. The goal is to always use the most accurate data available to produce the best prediction:

- **After sunset**: Open-Meteo's daily values for today reflect actual measurements — use them directly. The full day's real sunshine and real average temperature are known.
- **Before sunset**: The full day's data isn't available yet, so pro-rate today's forecasted sunshine by the fraction of daylight elapsed, and use the forecasted average temperature.

These values are sent as inputs to the prediction algorithm, which uses sunshine hours to estimate how much the sun has passively heated the water tank — a critical factor in Israel and other sunny climates where rooftop solar water heaters ("dud shemesh") are standard.

This requires changes on both frontend (auto-fill, new field, sunshine calculation) and backend (new model field, prediction algorithm update).

## Why Sunshine Hours Matter

In Israel and similar climates, most homes have a solar water heater on the roof. On sunny days, the sun heats the water significantly, meaning the electric boiler needs to run for less time (or not at all). On cloudy or short winter days, the boiler needs to compensate. The current prediction model only considers outdoor temperature and shower duration — it has no way to distinguish a sunny 20°C day (tank is warm) from a cloudy 20°C day (tank is cold). Adding sunshine hours closes this gap.

## Functional Requirements

### Frontend — Weather Auto-Fill
- On app load (or when InputForm mounts), request the user's location via `navigator.geolocation.getCurrentPosition`
- Use the returned coordinates to fetch today's data from the Open-Meteo daily forecast API: `daily=temperature_2m_mean,sunshine_duration,sunrise,sunset`
- Determine whether the current time is **after sunset** or **before sunset**:
  - **After sunset**: use the values directly — by this point Open-Meteo has updated today's data with actual measurements. Temperature and sunshine reflect what really happened today.
  - **Before sunset**: pro-rate sunshine by the fraction of daylight elapsed:
    - `elapsedFraction = clamp((now - sunrise) / (sunset - sunrise), 0, 1)`
    - `effectiveSunshine = forecastedSunshine × elapsedFraction`
    - Before sunrise: effective sunshine = 0
  - Temperature: use the daily value as-is (best available, becomes more accurate as the day progresses)
- `sunshine_duration` from Open-Meteo is in seconds — convert to hours
- Pre-fill the temperature field with the average daily temperature, rounded to the nearest integer
- Display the sunshine hours as a read-only informational field below the temperature input (e.g., "Sunshine today: 9.2h") — the user doesn't need to edit this, but should see it for transparency
- Show a brief loading indicator while fetching weather data
- Show a small visual indicator (e.g., location pin icon or "auto-detected" label) when values were auto-filled
- The user must always be able to manually override the temperature value
- Provide a refresh/re-detect button to manually trigger a new weather fetch
- If geolocation is denied, unavailable, or times out, fall back to the localStorage-persisted temperature; sunshine hours default to null (omitted from prediction request)
- If Open-Meteo fails or returns unexpected data, same fallback behavior
- Remember denied geolocation permission in localStorage to avoid repeated browser prompts
- Cache fetched weather data for 30 minutes to avoid redundant API calls on page refreshes (but recalculate effective sunshine from cached forecast data on each mount when before sunset, since the elapsed fraction changes)
- Continue saving temperature to localStorage on submit (existing behavior)

### Backend — New Sunshine Hours Field
- Add an optional `sunshineHours` field (float, nullable) to the `DailyRecord` model
- Add an optional `SunshineHours` field to the `PredictionRequest` struct
- Accept `sunshineHours` as an optional field in the `/api/calculate` and `/api/feedback` endpoints
- Validate `sunshineHours` range: 0 to 16 hours (max reasonable sunshine in a day)
- Store `sunshineHours` in the database when provided
- Existing records without sunshine data remain valid — the field is nullable and the algorithm must handle its absence

### Backend — Prediction Algorithm Update
- Add a third Gaussian distance dimension for sunshine hours to the v2 prediction algorithm, alongside duration and temperature
- When computing similarity weights, include: `wSun := gaussian(req.SunshineHours - r.SunshineHours, sigmaSun)` and multiply it into the overall weight `w = wDur * wTmp * wSun`
- Add a new config parameter `SigmaSun` (default 3.0 hours) controlling the Gaussian width for sunshine similarity
- When either the request or a historical record has no sunshine data (null/zero), skip the sunshine dimension for that comparison — effectively treat it as a neutral factor (weight = 1.0) so old records without sunshine data are not penalized
- Update `freqCellKey` to include rounded sunshine hours as a third dimension, so records with same duration+temp but different sunshine are in different frequency cells
- This graceful degradation means the algorithm improves as sunshine data accumulates, without breaking predictions on historical data

## Figma Design Reference (only if referenced)

N/A

## Possible Edge Cases

- User denies geolocation permission: fall back to localStorage temperature, sunshine hours omitted from request
- Browser does not support Geolocation API: same fallback
- Open-Meteo API is down or returns an error: same fallback
- Open-Meteo returns temperature or sunshine in unexpected format: fall back for the malformed field, use the other if valid
- User is on HTTP (insecure context): geolocation may be unavailable — fall back gracefully
- Geolocation request times out (>10 seconds): fall back
- User has no localStorage value and geolocation also fails: leave temperature empty, omit sunshine hours
- User manually edits the auto-filled temperature: manual value takes precedence and is saved to localStorage
- Multiple rapid mounts/unmounts of InputForm: cancel in-flight requests to avoid stale updates
- Historical records have no sunshine data: algorithm skips sunshine dimension for those comparisons (neutral weight)
- Request has no sunshine data (geolocation denied): algorithm runs on just temperature + duration as it does today
- Sunshine hours of 0 (completely cloudy day): this is a valid data point, not the same as "missing" — distinguish null (missing) from 0 (no sunshine)
- Very early morning before sunrise: effective sunshine = 0, which is correct — the tank hasn't been heated by the sun yet
- After sunset: sunshine = actual measured value from Open-Meteo, reflecting the real solar heating the tank received that day
- Timezone handling: sunrise/sunset times from Open-Meteo must be compared to the user's local time, not UTC
- The v1 prediction algorithm does not use sunshine hours — it should accept but ignore the field to maintain compatibility
- freqCellKey with null sunshine: when sunshine is null, use a sentinel value (e.g., "?") for the cell key so null-sunshine records group together

## Acceptance Criteria

- When the user opens the app and allows geolocation, both the temperature and sunshine data are fetched from Open-Meteo
- Temperature field is pre-filled with the average daily temperature (rounded to nearest integer)
- Sunshine hours are displayed as a read-only field
- A loading indicator shows while weather data is being fetched
- A visual indicator shows values were auto-detected
- The user can manually override the temperature at any time
- A refresh button allows re-fetching weather data without reloading
- If geolocation/API fails, temperature falls back to localStorage; sunshine hours are omitted
- Denied geolocation permission is remembered across sessions
- Weather data is cached for 30 minutes (effective sunshine recalculated from cached forecast when before sunset)
- The `/api/calculate` and `/api/feedback` endpoints accept an optional `sunshineHours` field
- `sunshineHours` is validated to be between 0 and 16
- `DailyRecord` stores `sunshineHours` as a nullable float
- `PredictionRequest` includes an optional `SunshineHours` field
- The v2 prediction algorithm uses sunshine hours as a third similarity dimension when available
- `freqCellKey` includes sunshine hours as a third dimension
- Predictions remain accurate for historical records that lack sunshine data
- The v1 predictor accepts but ignores the sunshine field
- The `/api/history/export` CSV includes sunshine hours

## Open Questions

- Should sunshine duration (actual sun hours) or daylight duration (sunrise-to-sunset regardless of clouds) be used? Use sunshine duration — it's more accurate for solar heating since cloud cover directly affects how much the tank heats up
- What is the right `SigmaSun` default? 3.0 hours — a 3-hour difference in sunshine halves the similarity weight. Can be tuned later based on real data

## Testing Guidelines

Create a test file(s) in the ./tests folder for the new feature, and create meaningful tests for the following cases, without going too heavy:

### Frontend (weather fetch utility)
- Fetching weather data from Open-Meteo with valid coordinates returns rounded average temperature and sunshine data
- Sunshine calculation: before sunrise returns 0
- Sunshine calculation: after sunset returns full value (actual measurement)
- Sunshine calculation: midday returns approximately half of forecasted value (pro-rated)
- Fallback to defaults when API response is malformed
- Fallback to defaults when API call throws an error
- Correctly constructs the Open-Meteo API URL with daily parameters including sunrise/sunset
- Caching: a second call within 30 minutes returns cached forecast (recalculates sunshine when before sunset)
- Null sunshine hours when geolocation is denied

### Backend (prediction algorithm)
- Prediction with sunshine hours produces different results than without (when historical data has sunshine)
- Prediction gracefully handles null sunshine hours in the request (falls back to 2D similarity)
- Prediction gracefully handles historical records with null sunshine hours (skips sunshine dimension for those records)
- The `SigmaSun` parameter correctly controls Gaussian width for sunshine similarity
- freqCellKey includes sunshine as third dimension
- sunshineHours validation rejects values outside 0-16 range
