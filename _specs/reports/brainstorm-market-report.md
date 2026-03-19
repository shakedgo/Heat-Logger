# Product Improvement Report: Heat-Logger

## Tech Stack
- **Backend:** Go 1.23, Gin (HTTP framework), GORM (ORM), SQLite (single-file DB)
- **Frontend:** Vue 3 (Options API), Vite, Axios, SASS, Font Awesome
- **Testing:** Go testify with mocks
- **Infrastructure:** Local dev only (`air` hot-reload, no Docker/CI/deployment config)

## What This App Does
Heat-Logger predicts how long to run a water heater before a shower, learning from user satisfaction feedback. Users enter outdoor temperature and shower duration, get a predicted heating time via a Gaussian-kNN algorithm, then rate the result on a cold-to-hot scale (1-100, 50 = perfect). The system blends per-user history with global data to improve predictions over time.

## Target Audience
Individuals with storage/tank water heaters (not instant/tankless) in climates with meaningful seasonal temperature variation -- likely in Mediterranean, Middle Eastern, or continental European regions. The single-page, no-auth, household-scale design targets a technically comfortable individual who wants to stop guessing and wasting gas/electricity on heating.

## Market Context

This is a niche personal utility with no direct software-only competitor doing exactly this. The closest products are hardware-based:

- **[Aquanta](https://aquanta.io/)** ($149): A retrofit smart controller that clamps onto existing water heaters. It learns usage patterns and schedules heating automatically. Key differentiator: hardware sensors measure actual water temperature and recovery rate -- no user input needed. It integrates with Alexa and supports time-of-use electricity rate optimization.
- **[TrickleStar TS2301](https://tricklestar.com/products/ts2301-wi-fi-electric-water-heater-controller)**: Wi-Fi controller focused on scheduling and remote on/off. No learning algorithm -- purely manual schedules.
- **[Rheem EcoNet](https://www.rheem.com/water-heating/articles/guide-to-using-the-app-to-save-money-with-your-smart-water-heater/)**: An app for Rheem's connected water heaters. Shows ready status, energy usage, vacation mode, and peak-hour scheduling. Built for their hardware ecosystem only.

**Heat-Logger's unique angle:** Software-only, zero hardware cost, works with any water heater, and uses a human-in-the-loop ML feedback mechanism rather than thermosensors. The tradeoff is that it requires manual input and active engagement. The key strategic question is: how do you minimize friction enough that users keep feeding the model?

## Problems Identified

### 1. No Energy/Cost Savings Visibility -- The Core Value Prop Is Invisible
The app helps users avoid over-heating (wasting energy) and under-heating (cold showers), but never quantifies the value it delivers. Hardware competitors like Aquanta prominently show energy usage data and estimated savings. Heat-Logger has all the data to approximate this -- it knows heating time deltas between predictions and could estimate energy savings relative to a fixed "always heat for X minutes" baseline -- but surfaces none of it. Without visible ROI, the app feels like a curiosity rather than a utility.

**Evidence:** No energy or cost-related fields in `backend/internal/models/record.go`. No aggregation logic anywhere in the backend. The `HistoryList.vue` shows raw records but no summaries.

### 2. The Feedback Loop Has No Memory of "What Actually Happened"
Users rate satisfaction on a cold/hot scale, but the app never asks what heating time the user *actually used*. The system assumes the user followed the prediction exactly, but in practice users often round up/down or ignore the suggestion. The prediction algorithm in `prediction_service_v2.go` (lines 347-388, `impliedTarget`) derives adjustments from `record.HeatingTime` -- but that's the *predicted* time, not the *actual* time. If the user heated for longer than predicted and reported "too hot," the model misinterprets the signal.

**Evidence:** `PredictionResponse` in `prediction_service.go` (line 37) returns `HeatingTime`. The feedback in `record_handler.go` (line 74-131) stores whatever `heatingTime` is in the submitted record -- which is the predicted value passed through from the frontend (`InputForm.vue` line 198: `heatingTime: this.latestHeatingTime`). There is no "actual time used" field.

### 3. No Seasonal Awareness or Weather Integration
The prediction algorithm treats temperature as a point-in-time input, but heating needs are deeply seasonal. A 15C day in autumn (water pipes still warm from summer) heats differently than a 15C day in spring (pipes cold from winter). The model has `Date` in every record but never uses month/season as a feature. Additionally, requiring users to manually enter temperature every time is a friction point that hardware competitors eliminate entirely.

**Evidence:** `prediction_service_v2.go` computes Gaussian distance on temperature and duration (line 178-179) but never considers `r.rec.Date` month or season. The `PredictionRequest` struct (line 31-34) has no season/month field. No weather API integration exists anywhere.

### 4. Cold-Start Is Discouraging -- New Users Get a Magic "30 Minutes" With No Explanation
When a new user (or one in a new temperature range) has no matching history, the v2 predictor returns a hardcoded 30 minutes (`prediction_service_v2.go` line 156). The v1 uses a linear formula (`prediction_service.go` lines 87-91). Neither explains to the user why this number was chosen or how many data points are needed before predictions become personalized. The UI shows "Initial estimate" but doesn't set expectations about the learning curve.

**Evidence:** `prediction_service_v2.go` line 156: `out := 30.0`. `InputForm.vue` line 57: displays "Initial estimate" when sampleCount is 0, but says nothing about what happens next.

### 5. The App Works Only When Open -- No Notification or Reminder System
Unlike Aquanta (which runs autonomously) or EcoNet (which sends push notifications), Heat-Logger requires the user to actively open the app, enter data, and provide feedback every single time. There is no reminder to log a shower, no notification when it's time to turn off the heater, and no timer. For a daily-use app, this is a major retention gap -- the value depends on consistent usage, but nothing prompts it.

**Evidence:** No service worker, no push notification infrastructure, no PWA manifest. `frontend/index.html` has no `<link rel="manifest">`. The app is purely request-response.

## Proposed Improvements

### Quick Wins

#### 1. Add an Estimated Savings Summary to the History View
- **Solves:** Problem #1
- **What:** Show a simple "estimated minutes saved this week/month" summary at the top of the history list by comparing each prediction to a configurable baseline heating time (e.g., the user's average before using the app, or a sensible default like 45 minutes).
- **How:** Compute this client-side from the existing `/api/history` data. In `HistoryList.vue`, add a computed property that sums `(baseline - record.heatingTime)` for records in the current week/month where satisfaction was near-perfect (40-60). Display it in a small summary card above the history entries. No backend changes needed. Could also show "average satisfaction this month" alongside it. Roughly 30-50 lines of new Vue code.
- **Differentiator:** Aquanta shows energy data prominently; this brings Heat-Logger closer to that value visibility without requiring hardware sensors.

#### 2. Add a Countdown Timer After Prediction
- **Solves:** Problem #5 (partially)
- **What:** After displaying the predicted heating time, offer a one-tap "Start Timer" button that counts down from the predicted minutes and plays an audio alert when done. This keeps users in the app during the heating window and naturally leads them to the feedback form.
- **How:** Add a `HeatingTimer.vue` component with a simple countdown using `setInterval`. Show it conditionally in `InputForm.vue` when `currentEntry.heatingTime` is set. Use the Web Audio API or a short `<audio>` element for the alert sound. When the timer finishes, auto-scroll to the satisfaction slider. Purely frontend, ~80-100 lines of new code.
- **Differentiator:** No hardware competitor offers this because they control the heater directly. For a software-only tool, a timer bridges the gap between "prediction" and "action."

#### 3. Add an "Actual Heating Time" Optional Field to Feedback
- **Solves:** Problem #2
- **What:** When submitting feedback, let users optionally adjust the heating time they actually used (pre-filled with the predicted value). This gives the ML model a more accurate signal without making the UX heavier for users who followed the prediction exactly.
- **How:** In `InputForm.vue`, add an editable number input pre-filled with `currentEntry.heatingTime` in the feedback form. In `handleFeedback()`, send the (potentially modified) value. Add a `<small>` hint: "Adjust if you heated for a different time than suggested." The backend already accepts any `heatingTime` float in the feedback payload -- no backend changes needed.

### Long-term Bets

#### 1. Progressive Web App with Timer Notifications
- **Solves:** Problem #5
- **What:** Convert Heat-Logger into a PWA with a web app manifest, service worker, and push notifications. After a user starts a heating timer, the app can send a browser notification when heating is done -- even if the tab is in the background. On mobile, this makes it feel like a native app with an install prompt.
- **How:** Add a `manifest.json` with app metadata and icons. Register a service worker (Vite has `vite-plugin-pwa`). Use the Notifications API for timer alerts. For daily reminders ("Log your shower?"), use the Push API with a lightweight backend push service. Modify `frontend/vite.config.js` to include the PWA plugin. This touches the build config, adds 2-3 new files, and requires HTTPS for production.
- **Differentiator:** None of the hardware competitors offer a free, installable PWA. This gives Heat-Logger a "native app feel" at zero distribution cost.

#### 2. Auto-Fill Temperature via Weather API and Seasonal Model Enhancement
- **Solves:** Problem #3
- **What:** Use the browser's Geolocation API + Open-Meteo (free, no API key) to auto-fill the current outdoor temperature. On the backend, enhance the prediction model to incorporate month/season as a feature -- so the same temperature in January vs. July can produce different predictions.
- **How:** Frontend: In `InputForm.vue`, add a "Use current weather" button that calls `navigator.geolocation.getCurrentPosition()`, then fetches `https://api.open-meteo.com/v1/forecast?latitude=X&longitude=Y&current_weather=true`. Auto-fill the temperature field. Backend: Add a `Month` or `Season` derived field to `PredictionRequest`. In `prediction_service_v2.go`, add a third Gaussian dimension for month-distance (circular, so December and January are close). This requires updating the `PredictionRequest` struct, adding ~20 lines to the v2 predictor, and ~30 lines to the frontend.

## Recommended Next Step
**Add a Countdown Timer After Prediction**
1. Create `frontend/src/components/HeatingTimer.vue` with a countdown display (minutes:seconds), a Start/Cancel button, and an audio beep on completion. Accept `duration` (minutes) as a prop.
2. In `frontend/src/components/InputForm.vue`, import `HeatingTimer` and render it below the prediction result when `currentEntry.heatingTime` is set and the user hasn't submitted feedback yet. Pass `currentEntry.heatingTime` as the duration prop.
3. When the timer finishes, emit an event or auto-scroll to the satisfaction slider and show a toast ("Heating complete! How was your shower?") to prompt immediate feedback -- closing the loop from prediction to feedback in a single session.

## Sources
- [Aquanta Smart Water Heater Controller](https://aquanta.io/)
- [TrickleStar Wi-Fi Water Heater Controller](https://tricklestar.com/products/ts2301-wi-fi-electric-water-heater-controller)
- [Rheem EcoNet App Guide](https://www.rheem.com/water-heating/articles/guide-to-using-the-app-to-save-money-with-your-smart-water-heater/)
- [Aquanta Review - TechHive](https://www.techhive.com/article/578640/aquanta-water-heater-controller-review.html)
- [Eco-Home Innovations: Smart Water Heating in 2025 - Eccotemp](https://www.eccotemp.com/blog/ecohome-innovations-how-smart-water-heating-is-saving-energy-in-2025/)
