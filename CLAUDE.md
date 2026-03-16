# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

Heat-Logger is a full-stack app that predicts how long to heat water before a shower, using ML algorithms that learn from user satisfaction feedback. It consists of a Go backend (Gin + GORM + SQLite) and a Vue 3 frontend (Vite).

## Development Commands

### Start everything (recommended)
```bash
./run-dev.sh
```
This starts the backend with hot-reload (`air`) and the frontend dev server concurrently.

### Frontend only
```bash
cd frontend
npm run dev      # Dev server at localhost:5173
npm run build    # Production build
```

### Backend only
```bash
cd backend
air              # Hot-reload dev server at localhost:8080
go build -o ./tmp/main ./cmd/server/main.go  # Manual build
```

### Tests
```bash
cd backend
go test ./internal/services/...
```

### Environment setup
```bash
cp backend/.env.example backend/.env
# Edit backend/.env as needed
```

## Architecture

### Request flow
```
Vue 3 UI → Axios (localhost:8080/api) → Gin router → Handler → Service → GORM → SQLite
```

### Key endpoints
| Method | Path | Purpose |
|--------|------|---------|
| POST | `/api/calculate` | Predict heating time given duration + temperature |
| POST | `/api/feedback` | Save result with satisfaction score (1–100, 50=perfect) |
| GET  | `/api/history` | Fetch all records |
| GET  | `/api/history/export` | Export CSV |
| POST | `/api/history/delete` | Delete record by ID |
| POST | `/api/history/deleteall` | Delete all records |

### Prediction service versions

Two interchangeable prediction algorithms, selected via `PREDICTOR_VERSION` in `.env`:

- **v1** (`prediction_service.go`): Target-based with similarity matching, clustering, dynamic learning rates, success anchors, perfect score decay
- **v2** (`prediction_service_v2.go`, **default**): Gaussian-kNN with distance weighting, anchor boost for near-perfect scores, user/global blending, recency decay, step-capped safety bounds

Both implement the `Predictor` interface in `predictor.go`. The factory in `config` instantiates the right one.

### Database model
`DailyRecord` (table: `daily_records`) — stores `ShowerDuration`, `AverageTemperature`, `HeatingTime`, `Satisfaction` (1–100), `UserID` (default `"global"`), auto-migrated on startup.

### Configuration
All config is in `backend/internal/config/`. Environment variables (or `.env`) control port, DB path, CORS origins, predictor version, log level, Gin mode. See `backend/ENVIRONMENT.md` for the full reference.

### Frontend
Single-page Vue 3 app. Main components: `InputForm.vue` (prediction UI), `HistoryList.vue` (record browser), `UiToaster.vue` (notifications). API calls centralized in `src/plugins/api.js` (Axios instance pointed at `localhost:8080/api`).
