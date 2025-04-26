
# SYT Go Queue

A queue system built with asynq for handling LLM API calls.

## Features

- Queue task production and consumption using asynq (Redis-based)
- HTTP API endpoint for task creation
- LLM (Deepseek) API integration
- MySQL storage for conversation results

## Architecture

```ascii
┌──────────────┐         ┌──────────────┐         ┌──────────────┐
│   PHP API    │ ──────> │  HTTP API    │ ──────> │ Redis Queue  │
└──────────────┘         └──────────────┘         └──────────────┘
                                                         │
                                                         │
                                                         ▼
┌──────────────┐         ┌──────────────┐         ┌──────────────┐
│    MySQL     │ <────── │  Consumer    │ <────── │ Asynq Worker │
└──────────────┘         └──────────────┘         └──────────────┘
                               │
                               │
                               ▼
                        ┌──────────────┐
                        │ Deepseek API │
                        └──────────────┘
```

## Project Structure

```
.
├── api/            # HTTP API handlers
├── cmd/            # Application entry points
│   ├── api/        # HTTP API server
│   └── worker/     # Asynq worker
├── config/         # Configuration files
├── internal/       # Internal packages
│   ├── handler/    # HTTP handlers
│   ├── model/      # Database models
│   ├── service/    # Business logic
│   └── task/       # Queue task definitions
└── pkg/            # Shared packages
```

## Quick Start

1. Prerequisites:
   - Go 1.24+
   - Redis
   - MySQL

2. Configuration:
   ```yaml
   # config/config.yaml
   redis:
     addr: localhost:6379
     password: ""
   
   mysql:
     dsn: root:password@tcp(localhost:3306)/dbname?charset=utf8mb4&parseTime=True&loc=Local
   
   deepseek:
     api_key: your_api_key
   ```

3. Run API server:
   ```bash
   go run cmd/api/main.go
   ```

4. Run worker:
   ```bash
   go run cmd/worker/main.go
   ```

## API Documentation

### Create LLM Task

```http
POST /api/v1/tasks/llm
Content-Type: application/json

{
    "prompt": "Your question here",
    "model": "deepseek-chat",
    "callback_url": "http://your-callback-url"
}
```

Response:
```json
{
    "task_id": "task_123456",
    "status": "enqueued"
}
```

## License

MIT
