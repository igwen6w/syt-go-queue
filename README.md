
# SYT Go Queue

SYT Go Queue is a robust, scalable asynchronous task processing system built in Go, specifically designed for handling Large Language Model (LLM) API calls. It leverages Redis-based queuing to manage high-throughput requests to LLM services, ensuring efficient processing and reliable storage of conversation results.

## Features

- **Asynchronous Processing**: Queue task production and consumption using asynq (Redis-based)
- **HTTP API**: RESTful endpoints for task creation and management
- **LLM Integration**: Built-in support for Deepseek LLM API with configurable parameters
- **Persistent Storage**: MySQL database for storing conversation history and results
- **Configurable Workers**: Adjustable concurrency and retry mechanisms
- **Structured Logging**: Comprehensive logging using zap logger
- **Error Handling**: Robust error handling with detailed error reporting
- **Callback Support**: Webhook notifications when tasks complete

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
│   ├── database/   # Database access and models
│   ├── handler/    # HTTP handlers
│   ├── logger/     # Logging utilities
│   ├── server/     # API server implementation
│   ├── task/       # Queue task definitions
│   ├── types/      # Common type definitions
│   └── worker/     # Worker implementation
└── pkg/            # Shared packages
```

## Database Schema

The system uses MySQL to store conversation data. The main table structure includes:

```sql
CREATE TABLE valuation_records (
    id BIGINT PRIMARY KEY,
    status VARCHAR(50),
    user_message TEXT,
    sys_message TEXT,
    report TEXT,
    failed_times INT DEFAULT 0,
    failed_info TEXT,
    progress VARCHAR(50),
    progress_info TEXT,
    current_task_node INT DEFAULT 0,
    callback_url VARCHAR(255),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
);
```

## Quick Start

### Prerequisites

- Go 1.24+
- Redis
- MySQL

### Installation

```bash
# Clone the repository
git clone https://github.com/igwen6w/syt-go-queue.git
cd syt-go-queue

# Install dependencies
go mod download
```

### Configuration

Create or modify `config/config.yaml`:

```yaml
app:
  name: syt-go-queue
  mode: development  # development, production
  port: 8080

redis:
  addr: localhost:6379
  password: ""
  db: 0

mysql:
  dsn: root:password@tcp(localhost:3306)/syt_queue?charset=utf8mb4&parseTime=True&loc=Local
  max_idle_conns: 10
  max_open_conns: 100

deepseek:
  api_key: your_api_key
  base_url: https://api.deepseek.com/v1
  timeout: 30s
  model: deepseek-chat
  max_tokens: 2000

queue:
  concurrency: 10  # Number of concurrent workers
  retry: 3         # Number of retries for failed tasks
  retention: 24h   # How long to keep completed tasks

logger:
  level: info       # debug, info, warn, error
  development: true # Pretty console output in development mode
```

### Running the Application

1. Start the API server:
   ```bash
   go run cmd/api/main.go --config=config/config.yaml
   ```

2. Start the worker:
   ```bash
   go run cmd/worker/main.go --config=config/config.yaml
   ```

### Building for Production

```bash
# Build the API server
go build -o bin/api cmd/api/main.go

# Build the worker
go build -o bin/worker cmd/worker/main.go
```

## API Documentation

### Create LLM Task

```http
POST /api/tasks/llm
Content-Type: application/json

{
    "table_name": "valuation_records",
    "id": 123
}
```

Response:
```json
{
    "code": 200,
    "message": "Success",
    "data": {
        "task_id": "task_123456",
        "status": "enqueued"
    }
}
```

## Testing

The project includes unit tests for critical components. To run the tests:

```bash
# Run all tests
go test ./...

# Run tests with coverage
go test -cover ./...
```

For testing, a separate configuration file `config/config_test.yaml` is used to avoid affecting production data.

## Deployment

### Docker

A Dockerfile is provided for containerized deployment:

```bash
# Build the Docker image
docker build -t syt-go-queue .

# Run the API server
docker run -p 8080:8080 -v $(pwd)/config:/app/config syt-go-queue api

# Run the worker
docker run -v $(pwd)/config:/app/config syt-go-queue worker
```

### Kubernetes

For Kubernetes deployment, sample manifests are available in the `deploy/k8s` directory.

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add some amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## License

MIT
