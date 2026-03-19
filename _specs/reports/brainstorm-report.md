# Product Improvement Report: Heat-Logger

## Tech Stack
- **Language:** Go 1.23 (backend), JavaScript/Vue 3 (frontend)
- **Framework:** Gin (HTTP router), GORM (ORM), Vite (build tool)
- **Libraries:** Axios (HTTP client), Font Awesome (icons), SASS (styling), testify (testing)
- **Database:** SQLite (single file, `data.db`)
- **Infrastructure:** Local-only dev setup (`air` for hot-reload, no containerization or deployment config found)

## What This App Does
Heat-Logger predicts how long a user should run their water heater before taking a shower, based on shower duration and outside temperature. Users enter these two values, receive a predicted heating time, then report satisfaction on a cold-to-hot scale (1-100). The app learns from this feedback using a Gaussian-kNN algorithm (v2) to improve future predictions, blending per-user history with global data.

## Target Audience
Individuals with non-instant (storage/tank) water heaters in climates where outdoor temperature meaningfully affects heating needs. The app's terminology ("average temperature," satisfaction on a cold/hot spectrum) and the single-user-at-a-time flow suggest a household user -- likely someone in a region with significant seasonal temperature variation (e.g., Mediterranean, Middle Eastern, or continental European climate) who wants to avoid both cold showers and wasted energy/gas from over-heating.

## Market Context
This is a niche personal utility / smart-home optimization tool. There are no direct competitors doing exactly this. The closest analogues would be:
- **Smart water heater controllers** (e.g., Aquanta, Rheem EcoNet) -- hardware devices that optimize heating schedules, but they use thermostats rather than user feedback loops.
- **Energy monitoring apps** (e.g., Sense, Emporia) -- track energy usage but don't predict specific heating durations from user satisfaction.

Heat-Logger's differentiator is the ML-from-feedback loop -- it is a software-only solution that requires no hardware integration.

## Problems Identified

### 1. No Feedback on Prediction Quality Over Time
The app collects rich data (satisfaction scores per prediction) but never shows the user whether predictions are getting better. There is no dashboard, trend visualization, or accuracy metric visible anywhere in the UI. The `DailyRecord` model stores `Satisfaction`, `HeatingTime`, `ShowerDuration`, and `AverageTemperature` -- all the ingredients for a learning curve chart -- but the only visualization is the raw history list. This is the "data without visualization" gap: users submit feedback but never see whether it's working. This likely undermines trust in the system and reduces motivation to keep providing feedback.

**Evidence:** `frontend/src/components/HistoryList.vue` renders records as a flat list with no aggregation. The backlog (`_specs/improvements-backlog.md`, line 42) explicitly calls out "Analytics view" as a future architecture item.

### 2. No Guidance on What Temperature to Enter
The form asks for "Average Temperature (C)" with no explanation of what temperature is meant -- outdoor ambient? Water inlet? Room temperature? The prediction algorithm treats it as an ambient outdoor temperature (higher temp = less heating needed per `prediction_service.go` line 89: `tempFactor := -0.15`), but the UI gives no hint. Users guessing wrong will poison the ML model with inconsistent data.

**Evidence:** `frontend/src/components/InputForm.vue` lines 19-25 -- the temperature input has no placeholder, helper text, or tooltip. The label just says "Average Temperature (C)."

### 3. User Identity Is Fragile and Unprotected
The `userId` is a free-text field stored in `localStorage`. Any typo creates a new "user" with zero history, and anyone can read or overwrite another user's data by typing their ID. The backlog acknowledges this (`_specs/improvements-backlog.md` line 41: "Multi-user auth"). More immediately, there is no validation or normalization -- "User1", "user1", and " user1 " would all be different users, silently fracturing the ML model.

**Evidence:** `frontend/src/components/InputForm.vue` line 151 reads from `localStorage` with no normalization. `backend/internal/handler/record_handler.go` line 56-61 checks for empty `UserID` but does no trimming or case normalization.

### 4. History List Doesn't Scale and Lacks Filtering
All records are fetched in a single `GET /api/history` call with no pagination, and the frontend renders them all inside a `max-height: 520px` scrollable div. The backlog (item #11) already flags this. But even before hitting performance limits, the current list is hard to use: there is no search, no date filtering, and no way to see records for a specific temperature range or shower duration -- making it difficult for users with 50+ entries to find or review relevant past sessions.

**Evidence:** `backend/internal/services/record_service.go` line 35-39 -- `GetAllRecords()` fetches everything. `frontend/src/components/HistoryList.vue` lines 14-47 render all entries with no filtering controls. The backlog items #11 and #13 confirm this is recognized but unaddressed.

### 5. Form Resets After Feedback, Losing Context
When a user submits satisfaction feedback, the form clears all inputs including temperature and shower duration (backlog item #8). For someone logging multiple showers on the same day (e.g., different household members), this forces redundant re-entry. The temperature probably hasn't changed between showers.

**Evidence:** `frontend/src/components/InputForm.vue` lines 205-214 -- `resetForm()` clears `averageTemperature` and `showerDuration` to empty strings. Only `userId` is preserved.

## Proposed Improvements

### Quick Wins

#### 1. Preserve Temperature and Duration After Feedback Submission
- **Solves:** Problem #5
- **What:** Keep the temperature and shower duration fields populated after submitting feedback, so users only need to adjust the satisfaction slider for repeated sessions.
- **How:** In `frontend/src/components/InputForm.vue`, modify `resetForm()` to preserve `averageTemperature` and `showerDuration` alongside `userId`. Roughly 3 lines changed. Also store the last-used temperature in `localStorage` so it persists across sessions (the outdoor temperature usually doesn't change between showers on the same day).

#### 2. Add Temperature Input Context and Auto-Fill
- **Solves:** Problem #2
- **What:** Add helper text clarifying this is outdoor/ambient temperature, and optionally auto-fill it using the browser's Geolocation API + a free weather API (or at minimum, persist the last entered value).
- **How:** Add a `<small>` tag under the temperature input in `InputForm.vue` saying "Enter the current outdoor temperature." Persist the last value in `localStorage` (like `userId` already is). For a stretch version, use `navigator.geolocation` + Open-Meteo (free, no API key) to pre-fill. This is purely a frontend change.

#### 3. Normalize User IDs
- **Solves:** Problem #3
- **What:** Trim whitespace and lowercase the `userId` on both frontend and backend to prevent accidental identity fragmentation.
- **How:** In `InputForm.vue` `handleCalculate()`, add `this.formData.userId = this.formData.userId.trim().toLowerCase()` before saving to localStorage and emitting. On the backend in `backend/internal/handler/record_handler.go`, add `req.UserID = strings.TrimSpace(strings.ToLower(req.UserID))` in both `CalculateHeatingTime` and `SubmitFeedback`. Two files, ~4 lines total.

### Long-term Bets

#### 1. Prediction Accuracy Dashboard
- **Solves:** Problem #1
- **What:** Add a visual dashboard showing how prediction accuracy improves over time. Display average satisfaction trend, prediction vs. actual needed time, and streak of "Just perfect" scores. This turns the app from a utility into something users actively want to engage with -- gamification through visible progress.
- **How:** Create a new `StatsPanel.vue` component. Compute stats client-side from the existing `/api/history` data (no backend changes needed for v1). Show: (a) a sparkline or bar chart of satisfaction over the last 20 entries, (b) current average satisfaction, (c) "best streak" of near-perfect scores. For charts, use a lightweight library like `unovis` or hand-draw SVG sparklines (no heavy charting library needed). Place it between the input form and history list in `App.vue`.

#### 2. History Filtering with Time Ranges and Conditions
- **Solves:** Problem #4
- **What:** Add a filter bar above the history list with preset time ranges ("Last 7 days", "Last 30 days", "All") and optional temperature/duration range filters. Add server-side pagination to support growing datasets.
- **How:** Backend: Add query parameters to `GET /api/history` (`?from=&to=&page=&limit=`). Modify `GetAllRecords` in `record_service.go` to accept filter criteria. Frontend: Add a `HistoryFilters.vue` component with date presets and a "Search" button. Wire into `HistoryList.vue`. This touches `record_service.go`, `record_handler.go`, `router.go`, and two new/modified Vue components.

## Recommended Next Step
**Preserve Temperature and Duration After Feedback Submission**
1. In `frontend/src/components/InputForm.vue`, modify the `resetForm()` method to preserve `averageTemperature` and `showerDuration` (save them before clearing and restore after).
2. Add `localStorage.setItem('heatLogger_temperature', ...)` in `handleCalculate()` and read it back in `data()`, mirroring the existing `userId` persistence pattern.
3. Test the flow: submit prediction -> submit feedback -> verify temperature and duration fields retain their values, and that they persist across page reloads.
