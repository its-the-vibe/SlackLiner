# SlackLiner

A simple Golang service that reads Slack message payloads from a Redis list and publishes them to Slack using the Slack API.

## Features

- 🚀 Written in Go 1.24
- 📦 Lightweight Docker container built from scratch
- 🔄 Uses Redis BLPOP for efficient message queue processing
- 💬 Slack App integration with bot token (supports dynamic channels)
- ⚙️ Fully configurable via environment variables
- 🛡️ Graceful shutdown handling

## Prerequisites

- Docker and Docker Compose
- A Slack App with a Bot Token (see [Slack App Setup](#slack-app-setup))

## Quick Start

1. **Clone the repository:**
   ```bash
   git clone https://github.com/its-the-vibe/SlackLiner.git
   cd SlackLiner
   ```

2. **Create a `.env` file with your Slack bot token:**
   ```bash
   echo "SLACK_BOT_TOKEN=xoxb-your-bot-token-here" > .env
   ```

3. **Start the services:**
   ```bash
   docker-compose up -d
   ```

4. **Send a test message:**
   ```bash
   docker exec slackliner-redis redis-cli RPUSH slack_messages '{"channel":"#general","text":"Hello from SlackLiner!"}'
   ```

## Configuration

Configure the service using environment variables:

| Variable | Description | Default | Required |
|----------|-------------|---------|----------|
| `SLACK_BOT_TOKEN` | Slack Bot User OAuth Token | - | ✅ |
| `REDIS_ADDR` | Redis server address | `localhost:6379` | ❌ |
| `REDIS_PASSWORD` | Redis password (if authentication is enabled) | - | ❌ |
| `REDIS_LIST_KEY` | Redis list key to read messages from | `slack_messages` | ❌ |

## Message Format

Messages in the Redis list should be JSON objects with the following structure:

```json
{
  "channel": "#general",
  "text": "Your message here"
}
```

### With Metadata (Optional)

You can also include custom metadata with your messages:

```json
{
  "channel": "#general",
  "text": "Your message here",
  "metadata": {
    "event_type": "task_created",
    "event_payload": {
      "id": "12345",
      "priority": "high",
      "assignee": "john"
    }
  }
}
```

### Field Descriptions

- **channel**: The Slack channel ID or name (e.g., `#general`, `C1234567890`)
- **text**: The message text to send
- **metadata** (optional): Custom metadata to attach to the message
  - **event_type**: A string identifier for the event type (max 255 characters)
  - **event_payload**: A JSON object containing custom data (max 50 keys, values must be strings, numbers, or booleans)
  
> **Note**: Metadata is useful for tracking message context, workflow states, or custom events. 
> See [Slack's metadata documentation](https://api.slack.com/reference/metadata) for more details.

## Slack App Setup

1. Go to [Slack API Apps](https://api.slack.com/apps)
2. Click "Create New App" → "From scratch"
3. Name your app and select your workspace
4. Navigate to "OAuth & Permissions"
5. Add the following Bot Token Scopes:
   - `chat:write` - Send messages
   - `chat:write.public` - Send messages to channels the app isn't in
6. Install the app to your workspace
7. Copy the "Bot User OAuth Token" (starts with `xoxb-`)
8. Use this token as your `SLACK_BOT_TOKEN`

## Usage Examples

### Using Docker Compose (Recommended)

The easiest way to use SlackLiner is with Docker Compose, which sets up both Redis and the service:

```bash
# Start services
docker-compose up -d

# Push a message to Redis
docker exec slackliner-redis redis-cli RPUSH slack_messages '{"channel":"#general","text":"Test message"}'

# View logs
docker-compose logs -f slackliner

# Stop services
docker-compose down
```

### Using with External Redis

If you want to use an external Redis instance:

```bash
docker build -t slackliner .
docker run -e SLACK_BOT_TOKEN=xoxb-your-token \
           -e REDIS_ADDR=redis-host:6379 \
           slackliner
```

### Manual Testing

Push messages to Redis from any Redis client:

```bash
# Using redis-cli - Simple message
redis-cli RPUSH slack_messages '{"channel":"#general","text":"Hello World!"}'

# Using redis-cli - Message with metadata
redis-cli RPUSH slack_messages '{"channel":"#general","text":"Task created: Fix bug #123","metadata":{"event_type":"task_created","event_payload":{"task_id":"123","priority":"high"}}}'

# Using Python
import redis
import json

r = redis.Redis(host='localhost', port=6379)

# Simple message
message = {"channel": "#general", "text": "Hello from Python!"}
r.rpush('slack_messages', json.dumps(message))

# Message with metadata
message_with_metadata = {
    "channel": "#general", 
    "text": "Task created: Fix bug #123",
    "metadata": {
        "event_type": "task_created",
        "event_payload": {
            "task_id": "123",
            "priority": "high",
            "assignee": "john"
        }
    }
}
r.rpush('slack_messages', json.dumps(message_with_metadata))
```

## Development

### Building Locally

```bash
go mod download
go build -o slackliner .
```

### Running Locally

```bash
export SLACK_BOT_TOKEN=xoxb-your-token
export REDIS_ADDR=localhost:6379
./slackliner
```

### Building Docker Image

```bash
docker build -t slackliner .
```

## Architecture

```
┌─────────────┐         ┌─────────────┐         ┌─────────────┐
│   Producer  │ RPUSH   │    Redis    │  BLPOP  │ SlackLiner  │
│   (Any App) ├────────►│    Queue    ├────────►│   Service   │
└─────────────┘         └─────────────┘         └──────┬──────┘
                                                        │
                                                        │ POST
                                                        ▼
                                                ┌─────────────┐
                                                │  Slack API  │
                                                └─────────────┘
```

1. **Producer**: Any application that can write to Redis (API server, cron job, etc.)
2. **Redis Queue**: Stores messages in a list using RPUSH
3. **SlackLiner**: Reads messages using BLPOP (blocking, efficient)
4. **Slack API**: Receives messages via the Slack Bot API

## License

MIT License - See LICENSE file for details
