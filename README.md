# Heat-Logger

A full-stack application that predicts optimal water heater timing based on user feedback, shower duration, and environmental factors. The system uses similarity-based machine learning algorithms that continuously improve through granular 1–100 satisfaction feedback.

## Project Overview

Heat-Logger learns from user behavior to optimize daily shower routines. Two interchangeable prediction algorithms are available, selected via environment config. The default (**v2**) uses a Gaussian-weighted k-NN approach; the legacy (**v1**) uses a target-based similarity model.

## Key Features

- **Dual ML Algorithms**: Two prediction backends (v1 target-based, v2 Gaussian-kNN), switchable via env var
- **Granular Feedback**: 1–100 satisfaction scale (50 = perfect) for precise learning
- **Historical Analysis**: Record browser with satisfaction tracking and CSV export
- **Undo Delete**: Restore accidentally deleted history records
- **Health Check**: `/api/health` endpoint for backend availability monitoring

## Technical Architecture

### Frontend (Vue 3)
- **Framework**: Vue 3 with Options API
- **Build Tool**: Vite
- **Styling**: SCSS with responsive design
- **HTTP Client**: Axios
- **Components**: `InputForm.vue`, `HistoryList.vue`, `UiToaster.vue`, `SkeletonCard.vue`

### Backend (Go)
- **Framework**: Gin
- **Database**: SQLite with GORM
- **Architecture**: Handler → Service → GORM → SQLite
- **Prediction**: Plugin-based via `Predictor` interface; factory selects v1 or v2 at startup

## Machine Learning Algorithms

### v1 — Target-Based Prediction (`prediction_service.go`)
Finds similar historical records (±2°C, ±3 min) and computes a weighted target heating time:

```go
if satisfaction < 50 {
    coldnessFactor := (50.0 - satisfaction) / 49.0
    targetTime = record.HeatingTime + coldnessFactor*4.0
} else if satisfaction > 50 {
    hotnessFactor := (satisfaction - 50.0) / 50.0
    targetTime = record.HeatingTime - hotnessFactor*4.0
}
// Final prediction = weighted average of all target times
```

Perfect scores (satisfaction=50) attract future predictions and decay when contradicted by newer feedback.

### v2 — Gaussian-kNN (`prediction_service_v2.go`, **default**)
- Distance-weighted k-NN with Gaussian kernel
- Anchor boost for near-perfect scores
- Blends user-specific and global data
- Recency decay and step-capped safety bounds

Both implement the `Predictor` interface in `predictor.go`. Switch with `PREDICTOR_VERSION=v1` in `.env`.

## Environment Configuration

```bash
cp backend/.env.example backend/.env
# or
cd backend && ./scripts/env-setup.sh
```

Key variables:

```bash
SERVER_PORT=8080
DATABASE_PATH=./data.db
PREDICTOR_VERSION=v2   # v1 or v2
ENVIRONMENT=development
GIN_MODE=debug
```

See `backend/ENVIRONMENT.md` for the full reference.

## Development Commands

```bash
# Start everything (backend hot-reload + frontend dev server)
./run-dev.sh

# Frontend only
cd frontend && npm run dev       # localhost:5173
cd frontend && npm run build

# Backend only
cd backend && air                # localhost:8080
cd backend && go build -o ./tmp/main ./cmd/server/main.go

# Tests
cd backend && go test ./internal/services/...
```

## Project Structure

```
Heat-Logger/
├── backend/
│   ├── cmd/server/main.go
│   ├── internal/
│   │   ├── config/
│   │   │   ├── config.go
│   │   │   └── env.go
│   │   ├── handler/record_handler.go
│   │   ├── models/record.go
│   │   ├── routes/router.go
│   │   └── services/
│   │       ├── predictor.go              # Predictor interface
│   │       ├── prediction_service.go     # v1 algorithm
│   │       ├── prediction_service_v2.go  # v2 algorithm (default)
│   │       └── record_service.go
│   ├── pkg/database/database.go
│   ├── scripts/env-setup.sh
│   ├── .env.example
│   ├── ENVIRONMENT.md
│   └── data.db
├── frontend/
│   ├── src/
│   │   ├── components/
│   │   │   ├── InputForm.vue        # Prediction UI and feedback
│   │   │   ├── HistoryList.vue      # Record browser
│   │   │   ├── UiToaster.vue        # Notifications
│   │   │   └── SkeletonCard.vue     # Loading placeholder
│   │   ├── plugins/api.js
│   │   └── main.js
│   └── index.html
├── _specs/
│   └── improvements-backlog.md
└── run-dev.sh
```

## API Endpoints

| Method | Path | Purpose |
|--------|------|---------|
| POST | `/api/calculate` | Predict heating time (duration + temperature) |
| POST | `/api/feedback` | Save result with satisfaction score (1–100) |
| GET | `/api/history` | Fetch all records |
| POST | `/api/history/delete` | Delete record by ID |
| POST | `/api/history/deleteall` | Delete all records |
| POST | `/api/history/restore` | Restore a deleted record |
| GET | `/api/history/export` | Export data as CSV |
| GET | `/api/health` | Backend health check |

### Example Requests

**Calculate Heating Time:**
```http
POST /api/calculate
{"duration": 15.5, "temperature": 22.0}

Response: {"heatingTime": 10.8}
```

**Submit Feedback:**
```http
POST /api/feedback
{
  "showerDuration": 15.5,
  "averageTemperature": 22.0,
  "heatingTime": 10.8,
  "satisfaction": 50
}
```

## Future Improvements

See [`_specs/improvements-backlog.md`](_specs/improvements-backlog.md) for the full backlog. Highlights:

- Delete-all confirmation showing record count
- Preserve form inputs after feedback submission
- Structured logging for prediction service v2
- Consistent API error response shape
- Pagination / virtual scrolling for large history
- Time-based history filtering
- Multi-user auth (UserID is currently client-generated)
- Analytics view (satisfaction trends, prediction accuracy)

## License

Developed for educational and personal use, showcasing modern web development practices and machine learning implementation.
