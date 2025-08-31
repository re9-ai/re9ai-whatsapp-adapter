<<<<<<< HEAD
# WhatsApp Adapter

**Status**: PLANNED  
**Purpose**: WhatsApp Business API integration for multi-platform messaging  
**Stack**: Go, WhatsApp Business API, WebSocket  

## Overview

WhatsApp Adapter provides WhatsApp Business API integration for the re9.ai construction platform. This service enables users to interact with the platform through WhatsApp, handling message routing, media sharing, and business messaging for construction project management.

## Key Features

- WhatsApp Business API integration and webhook management
- Message format standardization for platform compatibility
- Media file handling (images, documents, voice messages, videos)
- Business messaging templates for construction workflows
- User authentication and session management via WhatsApp
- Real-time message delivery and status tracking
- Integration with chat orchestrator for unified messaging
- Support for WhatsApp Business features (quick replies, interactive buttons)

## Technology Stack

- **Runtime**: Go 1.21+
- **Framework**: Gin (HTTP framework)
- **WhatsApp API**: WhatsApp Business API
- **Database**: PostgreSQL, Redis
- **Real-time**: WebSocket
- **File Storage**: Amazon S3
- **Deployment**: Docker, Kubernetes

## Documentation

For detailed architecture, API specifications, and implementation guidance, see the [re9.ai Infrastructure Wiki](https://github.com/re9-ai/re9ai-terraform-infra/wiki).

## Repository Status

This repository contains the planned architecture for the WhatsApp Adapter. Implementation is currently in planning phase.

## Related Services

- [Chat Orchestrator](https://github.com/re9-ai/re9ai-chat-orchestrator) - Multi-platform coordination
- [Chat Service](https://github.com/re9-ai/re9ai-chat-service) - Core messaging functionality
- [Telegram Adapter](https://github.com/re9-ai/re9ai-telegram-adapter) - Telegram integration
=======
## 📚 Documentation & Wiki

This repository includes a comprehensive wiki submodule containing architectural and business documentation:

- **📖 Wiki Access**: The `wiki/` directory contains extensive documentation
- **🔄 Wiki Sync**: Use `./scripts/update-wiki.sh` to sync latest wiki changes
- **📋 Wiki Guide**: See [WIKI.md](WIKI.md) for detailed wiki management instructions

⚠️ **Important**: The wiki is read-only in this repository. Edit content through GitHub's wiki interface, then sync locally using the provided script.



# re9.ai WhatsApp Adapter

A Go-based WhatsApp Business API adapter using Twilio for the re9.ai platform. This service handles incoming and outgoing WhatsApp messages, media processing, and AI integration.

## Features

- **Twilio WhatsApp Business API Integration**: Send and receive WhatsApp messages
- **Media Handling**: Support for images, videos, audio, and documents
- **AI Integration**: Forward messages to chat orchestrator for AI processing
- **Database Storage**: PostgreSQL for message persistence
- **Redis Caching**: Fast message retrieval and rate limiting
- **Security**: Webhook signature verification and security headers
- **Health Checks**: Readiness and liveness probes
- **Structured Logging**: JSON logging for production environments

## Architecture

```
WhatsApp User → Twilio → WhatsApp Adapter → Chat Orchestrator → AI Services
                    ↓                           ↓
                Database ← Message Storage      AI Processing
                    ↓
                Redis Cache
```

## Prerequisites

- Go 1.21 or higher
- PostgreSQL 13+
- Redis 6+
- Twilio Account with WhatsApp Business API
- AWS S3 bucket (for media storage)

## Setup

### 1. Clone and Install Dependencies

```bash
git clone https://github.com/re9-ai/re9ai-whatsapp-adapter.git
cd re9ai-whatsapp-adapter
go mod download
```

### 2. Environment Configuration

Copy the example environment file and configure your settings:

```bash
cp .env.example .env
```

Edit `.env` with your configuration:

```bash
# Twilio Configuration
TWILIO_ACCOUNT_SID=ACxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx
TWILIO_AUTH_TOKEN=your_auth_token_here
TWILIO_WHATSAPP_FROM=whatsapp:+14155238886

# Database Configuration  
DATABASE_URL=postgres://user:password@localhost:5432/whatsapp_adapter

# Other required configurations...
```

### 3. Database Setup

Create the database and run migrations:

```sql
CREATE DATABASE whatsapp_adapter;
```

The service will automatically create the required tables on startup.

### 4. Twilio WhatsApp Configuration

1. **Get Twilio Credentials**:
   - Account SID from your [Twilio Console](https://console.twilio.com/)
   - Auth Token from the same page
   - WhatsApp Sandbox number (for development) or approved WhatsApp Business number

2. **Configure Webhook URL**:
   Set your webhook URL in Twilio Console:
   ```
   https://your-domain.com/webhooks/whatsapp/messages
   ```

3. **Webhook Configuration**:
   - Webhook URL: `https://your-domain.com/webhooks/whatsapp/messages`
   - Status Callback URL: `https://your-domain.com/webhooks/whatsapp/status`
   - HTTP Method: POST

## Running the Service

### Development

```bash
# Install dependencies
go mod download

# Run the service
go run main.go
```

### Production

```bash
# Build the application
go build -o whatsapp-adapter main.go

# Run the binary
./whatsapp-adapter
```

### Docker

```bash
# Build Docker image
docker build -t re9ai/whatsapp-adapter .

# Run container
docker run -p 8080:8080 --env-file .env re9ai/whatsapp-adapter
```

## API Endpoints

### Health Checks

- `GET /health` - Basic health check
- `GET /ready` - Readiness check (includes database and Redis connectivity)

### WhatsApp Webhooks

- `GET /webhooks/whatsapp/verify` - Webhook verification
- `POST /webhooks/whatsapp/messages` - Incoming messages
- `POST /webhooks/whatsapp/status` - Message status updates

### Message API

- `POST /api/v1/messages/send` - Send WhatsApp message
- `GET /api/v1/messages/:messageId` - Get message details
- `POST /api/v1/media/upload` - Upload media files

### Metrics

- `GET /metrics` - Prometheus metrics (TODO)

## Sending Messages

### Text Message

```bash
curl -X POST http://localhost:8080/api/v1/messages/send 
  -H "Content-Type: application/json" 
  -d '{
    "to": "whatsapp:+5511999999999",
    "content": "Hello from re9.ai!",
    "type": "text"
  }'
```

### Media Message

```bash
curl -X POST http://localhost:8080/api/v1/messages/send 
  -H "Content-Type: application/json" 
  -d '{
    "to": "whatsapp:+5511999999999",
    "content": "Check out this image!",
    "type": "image",
    "media_url": "https://example.com/image.jpg",
    "media_type": "image/jpeg"
  }'
```

### Template Message

```bash
curl -X POST http://localhost:8080/api/v1/messages/send 
  -H "Content-Type: application/json" 
  -d '{
    "to": "whatsapp:+5511999999999",
    "template": "HXb5b62575e6e4ff6129ad7c8efe1f983e",
    "variables": {
      "1": "12/1",
      "2": "3pm"
    }
  }'
```

## Configuration

### Environment Variables

| Variable | Description | Required | Default |
|----------|-------------|----------|---------|
| `PORT` | Server port | No | `8080` |
| `ENVIRONMENT` | Environment (development/production) | No | `development` |
| `LOG_LEVEL` | Log level (debug/info/warn/error) | No | `info` |
| `DATABASE_URL` | PostgreSQL connection string | Yes | - |
| `REDIS_URL` | Redis connection string | No | `redis://localhost:6379` |
| `TWILIO_ACCOUNT_SID` | Twilio Account SID | Yes | - |
| `TWILIO_AUTH_TOKEN` | Twilio Auth Token | Yes | - |
| `TWILIO_WHATSAPP_FROM` | WhatsApp sender number | Yes | - |
| `WHATSAPP_WEBHOOK_SECRET` | Webhook verification secret | Yes | - |
| `WHATSAPP_VERIFY_TOKEN` | Webhook verification token | Yes | - |
| `AWS_REGION` | AWS region for S3 | No | `us-east-1` |
| `S3_BUCKET_NAME` | S3 bucket for media storage | No | - |
| `CHAT_ORCHESTRATOR_URL` | Chat orchestrator service URL | No | `http://localhost:8081` |
| `AI_PROCESSING_URL` | AI processing service URL | No | `http://localhost:8082` |

## Development

### Project Structure

```
.
├── cmd/                    # Application entrypoints
├── internal/
│   ├── config/            # Configuration management
│   ├── handlers/          # HTTP handlers
│   ├── middleware/        # HTTP middleware
│   ├── models/           # Data models
│   └── services/         # Business logic services
├── pkg/
│   ├── database/         # Database utilities
│   ├── logger/           # Logging utilities
│   └── redis/            # Redis utilities
├── scripts/              # Build and deployment scripts
├── Dockerfile           # Docker configuration
├── go.mod              # Go module definition
└── main.go             # Application entry point
```

### Running Tests

```bash
# Run all tests
go test ./...

# Run tests with coverage
go test -cover ./...

# Run tests with race detection
go test -race ./...
```

### Code Quality

```bash
# Run linter
golangci-lint run

# Format code
go fmt ./...

# Vet code
go vet ./...
```

## Monitoring

### Health Checks

The service provides health check endpoints for monitoring:

- `/health` - Always returns 200 OK with basic service info
- `/ready` - Returns 200 OK only when all dependencies are available

### Logging

Structured logging is provided via logrus:
- JSON format in production
- Human-readable format in development
- Configurable log levels

### Metrics

Prometheus metrics endpoint is available at `/metrics` (implementation pending).

## Troubleshooting

### Common Issues

1. **Webhook Verification Failed**
   - Ensure `WHATSAPP_WEBHOOK_SECRET` matches Twilio configuration
   - Check webhook URL is accessible from internet

2. **Database Connection Failed**
   - Verify `DATABASE_URL` is correct
   - Ensure PostgreSQL is running and accessible
   - Check database permissions

3. **Message Sending Failed**
   - Verify Twilio credentials
   - Check WhatsApp number is approved for business use
   - Ensure recipient has opted in to receive messages

4. **Media Upload Failed**
   - Check AWS credentials and S3 bucket permissions
   - Verify bucket exists and is in the correct region

### Debug Mode

Enable debug logging:

```bash
export LOG_LEVEL=debug
```

## Security

- Webhook signature verification prevents unauthorized requests
- Security headers are automatically added to all responses
- Input validation on all endpoints
- Rate limiting to prevent abuse

## Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests for new functionality
5. Ensure all tests pass
6. Submit a pull request

## License

This project is licensed under the MIT License - see the LICENSE file for details.

## Support

For support and questions:
- Create an issue on GitHub
- Contact the re9.ai development team
- Check the [wiki](./wiki/) for additional documentation

This repository is a template for projects in the re9.ai stack.

## Getting Started

1. Clone this repository to start a new re9.ai project.
2. Run the setup script on first use:
   ```bash
   ./scripts/setup-default-project-template.sh [REPO_NAME]
   ```
   This script will initialize the recommended folder structure, add the wiki submodule, and set up automation scripts and documentation templates.
3. Reference the [re9.ai/wiki](https://github.com/re9-ai/wiki) repository for architecture and business documentation.
4. Update this README with your project-specific information.

## Included
- Basic folder structure
- Wiki submodule integration
- Setup and wiki management scripts
- Copilot instructions
- .gitignore

---

This template ensures all re9.ai repositories start with a consistent, maintainable foundation.

- Follow the organization’s standards for documentation and automation.
- Use the provided scripts for setup and wiki management.
- Reference `.github/copilotinstructions` for Copilot usage guidelines.

---

This template ensures all re9.ai repositories start with a consistent, maintainable, and well-documented foundation.
>>>>>>> 811635edb07e5f2f8f096b80838adf2ddfa179f4
