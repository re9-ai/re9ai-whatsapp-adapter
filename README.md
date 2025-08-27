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



# re9.ai Project Template

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
