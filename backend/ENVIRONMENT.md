# Environment Configuration

This document describes how to configure the Heat-Logger backend using environment variables.

## Quick Start

1. Copy the example environment file:
   ```bash
   cp .env.example .env
   ```

2. Or run the setup script:
   ```bash
   ./scripts/env-setup.sh
   ```

3. Edit the `.env` file to customize your configuration

## Environment Variables

### Server Configuration

| Variable | Default | Description |
|----------|---------|-------------|
| `SERVER_PORT` | `8080` | Port the server will listen on |
| `SERVER_HOST` | `localhost` | Host address the server will bind to |

### Database Configuration

| Variable | Default | Description |
|----------|---------|-------------|
| `DATABASE_PATH` | `./data.db` | Path to the SQLite database file |
| `DATABASE_DRIVER` | `sqlite` | Database driver to use |

### Prediction Service Configuration

| Variable | Default | Description |
|----------|---------|-------------|
| `PREDICTOR_VERSION` | `v2` | Version of prediction service to use (`v1` or `v2`) |
| `PREDICTION_MODEL_PATH` | `./models/` | Path to prediction model files |

### CORS Configuration

| Variable | Default | Description |
|----------|---------|-------------|
| `CORS_ALLOWED_ORIGINS` | `http://localhost:5173,http://localhost:3000,http://127.0.0.1:5173` | Comma-separated list of allowed origins |
| `CORS_ALLOWED_METHODS` | `GET,POST,PUT,DELETE,OPTIONS` | Comma-separated list of allowed HTTP methods |
| `CORS_ALLOWED_HEADERS` | `Origin,Content-Type,Accept,Authorization` | Comma-separated list of allowed headers |

### Logging Configuration

| Variable | Default | Description |
|----------|---------|-------------|
| `LOG_LEVEL` | `info` | Log level (`debug`, `info`, `warn`, `error`) |
| `LOG_FORMAT` | `text` | Log format (`text` or `json`) |

### Application Configuration

| Variable | Default | Description |
|----------|---------|-------------|
| `ENVIRONMENT` | `development` | Application environment (`development`, `staging`, `production`) |
| `GIN_MODE` | `debug` | Gin framework mode (`debug`, `release`, `test`) |

## Environment-Specific Configurations

### Development
```bash
ENVIRONMENT=development
GIN_MODE=debug
LOG_LEVEL=debug
```

### Production
```bash
ENVIRONMENT=production
GIN_MODE=release
LOG_LEVEL=warn
SERVER_HOST=0.0.0.0
```

### Staging
```bash
ENVIRONMENT=staging
GIN_MODE=release
LOG_LEVEL=info
```

## File Structure

```
backend/
├── .env                 # Your local environment variables (not committed)
├── .env.example        # Template for environment variables
├── internal/config/    # Configuration package
│   ├── config.go       # Main configuration structs and loading
│   └── env.go          # .env file loading utilities
└── scripts/
    └── env-setup.sh    # Environment setup script
```

## Usage in Code

```go
import "heat-logger/internal/config"

func main() {
    // Load configuration
    cfg, err := config.Load()
    if err != nil {
        log.Fatal("Failed to load configuration:", err)
    }

    // Use configuration
    serverAddr := cfg.GetServerAddress()
    isProduction := cfg.IsProduction()
    
    // Access specific config sections
    port := cfg.Server.Port
    dbPath := cfg.Database.Path
}
```

## Best Practices

1. **Never commit `.env` files** - They may contain sensitive information
2. **Use `.env.example`** - As a template for required environment variables
3. **Set sensible defaults** - All configuration should have reasonable defaults
4. **Validate configuration** - Check required values on startup
5. **Use environment-specific files** - Consider `.env.local`, `.env.production` for different environments

## Troubleshooting

### Configuration not loading
- Check that `.env` file exists in the backend directory
- Verify file permissions
- Check for syntax errors in `.env` file

### Environment variables not taking effect
- Restart the application after changing `.env`
- Check that the variable name matches exactly (case-sensitive)
- Verify the variable is not already set in your shell

### Database connection issues
- Ensure `DATABASE_PATH` points to a writable location
- Check file permissions for the database directory
