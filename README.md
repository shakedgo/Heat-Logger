# Heat-Logger

A personal web application that predicts optimal water heater timing based on user feedback, shower duration, and environmental factors. The app learns from daily ratings (1-10 scale) and adjusts predictions accordingly.

## Features

- **Smart Prediction**: Uses machine learning to predict optimal water heater timing
- **Daily Tracking**: Log daily shower data including temperature, duration, and satisfaction rating
- **Learning System**: The app learns from your feedback to improve predictions over time
- **Historical Data**: View and manage your past entries
- **Weather Integration**: Considers environmental factors in predictions

## Tech Stack

### Frontend
- **Vue 3** with Options API (traditional export default pattern)
- **Vite** for build tooling
- **SCSS** for styling
- **Axios** for API communication
- **Pinia** for state management

### Backend
- **Go** with clean architecture
- **Chi Router** for HTTP routing
- **GORM** with SQLite for database
- **RESTful API** design

## Project Structure

```
Heat-Logger/
├── backend/           # Go backend application
│   ├── cmd/          # Application entry points
│   ├── internal/     # Private application code
│   └── data.json     # SQLite database
├── frontend/         # Vue 3 frontend application
│   ├── src/
│   │   ├── components/  # Vue components
│   │   ├── plugins/     # Vue plugins (API, etc.)
│   │   └── styles/      # SCSS stylesheets
│   └── public/       # Static assets
└── run-dev.sh        # Development startup script
```

## Development

### Prerequisites
- Go 1.21+
- Node.js 18+
- npm or yarn

### Quick Start

1. **Clone the repository**
   ```bash
   git clone <repository-url>
   cd Heat-Logger
   ```

2. **Start development servers**
   ```bash
   ./run-dev.sh
   ```

   This will start both the backend (Go) and frontend (Vue) development servers.

### Manual Setup

#### Backend
```bash
cd backend
go mod download
go run cmd/server/main.go
```

#### Frontend
```bash
cd frontend
npm install
npm run dev
```

## API Endpoints

- `POST /calculate` - Calculate heating time based on input data
- `POST /feedback` - Submit user feedback and save to history
- `GET /history` - Retrieve all historical records
- `POST /history/delete` - Delete a specific record
- `POST /history/deleteall` - Delete all records

## License

This is a personal project for learning and experimentation.
