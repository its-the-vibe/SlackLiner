# SlackLiner

A simple Golang service that reads Slack message payloads from a Redis list and publishes them to Slack using the Slack API. It also supports adding emoji reactions to existing Slack messages and provides an HTTP API for direct message posting.

## Features

- 🚀 Written in Go 1.24
- 📦 Lightweight Docker container built from scratch
- 🔄 Uses Redis BLPOP for efficient message queue processing
- 🌐 HTTP API endpoint for direct message posting with timestamp response
- 💬 Slack App integration with bot token (supports dynamic channels)
- 🎨 Support for Slack Block Kit for rich, interactive messages
- 😄 Emoji reaction support for existing messages
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

4. **Send a test message via HTTP:**
   ```bash
   curl -X POST http://localhost:8080/message \
     -H "Content-Type: application/json" \
     -d '{"channel":"#general","text":"Hello from SlackLiner HTTP API!"}'
   ```

5. **Or send via Redis queue:**
   ```bash
   docker exec slackliner-redis redis-cli RPUSH slack_messages '{"channel":"#general","text":"Hello from SlackLiner!"}'
   ```

6. **Add a test reaction:**
   ```bash
   docker exec slackliner-redis redis-cli RPUSH slack_reactions '{"reaction":"thumbsup","channel":"#general","ts":"1234567890.123456"}'
   ```

## Configuration

Configure the service using environment variables:

| Variable | Description | Default | Required |
|----------|-------------|---------|----------|
| `SLACK_BOT_TOKEN` | Slack Bot User OAuth Token | - | ✅ |
| `HTTP_ADDR` | HTTP server address | `:8080` | ❌ |
| `REDIS_ADDR` | Redis server address | `localhost:6379` | ❌ |
| `REDIS_PASSWORD` | Redis password (if authentication is enabled) | - | ❌ |
| `REDIS_LIST_KEY` | Redis list key to read messages from | `slack_messages` | ❌ |
| `REDIS_REACTION_LIST_KEY` | Redis list key to read reactions from | `slack_reactions` | ❌ |
| `TIMEBOMB_REDIS_CHANNEL` | Redis Pub/Sub channel for TimeBomb integration | `timebomb-messages` | ❌ |

## HTTP API

### POST /message

Send a Slack message via HTTP POST request. The endpoint accepts the same JSON payload as the Redis queue method and returns the channel ID and message timestamp.

**Endpoint:** `POST http://localhost:8080/message`

**Request Body:** Same as [Message Formats](#message-formats) below

**Response:**
```json
{
  "channel": "C1234567890",
  "ts": "1766282873.772199"
}
```

The response includes:
- **channel**: The actual Slack channel ID where the message was posted
- **ts**: The message timestamp, which can be used for threading (as `thread_ts`) or adding reactions

**Example using curl:**
```bash
curl -X POST http://localhost:8080/message \
  -H "Content-Type: application/json" \
  -d '{
    "channel": "#general",
    "text": "Hello from HTTP API!"
  }'
```

**Example response:**
```json
{
  "channel": "C1234567890",
  "ts": "1766282873.772199"
}
```

**Using the response to post a threaded reply:**
```bash
# First message
RESPONSE=$(curl -s -X POST http://localhost:8080/message \
  -H "Content-Type: application/json" \
  -d '{"channel": "#general", "text": "Parent message"}')

# Extract the timestamp
TS=$(echo $RESPONSE | jq -r '.ts')

# Post a reply in the thread
curl -X POST http://localhost:8080/message \
  -H "Content-Type: application/json" \
  -d "{
    \"channel\": \"#general\",
    \"text\": \"Reply to thread\",
    \"thread_ts\": \"$TS\"
  }"
```

### GET /health

Health check endpoint that returns `OK` if the service is running.

**Example:**
```bash
curl http://localhost:8080/health
```

## Message Formats

### Posting Messages

Messages in the Redis list should be JSON objects with the following structure:

```json
{
  "channel": "#general",
  "text": "Your message here"
}
```

### With TTL (Optional)

You can include a `ttl` field to automatically delete the message after a specified number of seconds via [TimeBomb](https://github.com/its-the-vibe/TimeBomb):

```json
{
  "channel": "#general",
  "text": "This message will be deleted in 1 hour",
  "ttl": 3600
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

### Posting to a Thread (Optional)

You can reply to an existing message thread by including the `thread_ts` field with the timestamp of the parent message:

```json
{
  "channel": "#general",
  "text": "This is a reply in a thread",
  "thread_ts": "1234567890.123456"
}
```

> **Note**: The `thread_ts` value is the message timestamp (`ts`) returned when the original message was posted. Thread replies will appear nested under the parent message in Slack.

### With Blocks (Optional)

You can use Slack's [Block Kit](https://api.slack.com/block-kit) to create rich, interactive messages with structured layouts:

```json
{
  "channel": "#general",
  "blocks": [
    {
      "type": "section",
      "text": {
        "type": "mrkdwn",
        "text": "Hello, *world*! :tada:"
      }
    },
    {
      "type": "input",
      "element": {
        "type": "external_select",
        "placeholder": {
          "type": "plain_text",
          "text": "Select an item",
          "emoji": true
        },
        "action_id": "SlackCompose"
      },
      "label": {
        "type": "plain_text",
        "text": "Label",
        "emoji": true
      },
      "optional": false
    }
  ]
}
```

You can also combine blocks with text (text serves as a fallback for notifications and non-block surfaces):

```json
{
  "channel": "#general",
  "text": "Fallback text for notifications",
  "blocks": [
    {
      "type": "section",
      "text": {
        "type": "mrkdwn",
        "text": "Rich *formatted* message with blocks"
      }
    }
  ]
}
```

> **Note**: Either `text` or `blocks` must be provided (or both). When using blocks, the `text` field is optional and serves as a fallback for notifications and legacy clients.
> See [Slack's Block Kit documentation](https://api.slack.com/block-kit) for available block types and examples.

### Field Descriptions

- **channel**: The Slack channel ID or name (e.g., `#general`, `C1234567890`)
- **text** (optional if blocks provided): The message text to send - serves as fallback when blocks are provided
- **blocks** (optional): An array of [Block Kit](https://api.slack.com/block-kit) blocks for rich, interactive messages
- **thread_ts** (optional): Thread timestamp to reply to an existing thread - use the `ts` value from a previous message
- **ttl** (optional): Time-to-live in seconds - if provided, the message will be automatically deleted after this duration via [TimeBomb](https://github.com/its-the-vibe/TimeBomb)
- **metadata** (optional): Custom metadata to attach to the message
  - **event_type**: A string identifier for the event type (max 255 characters)
  - **event_payload**: A JSON object containing custom data (max 50 keys, values must be strings, numbers, or booleans)
  
> **Note**: Metadata is useful for tracking message context, workflow states, or custom events. 
> See [Slack's metadata documentation](https://api.slack.com/reference/metadata) for more details.

> **Note**: The TTL feature requires [TimeBomb](https://github.com/its-the-vibe/TimeBomb) to be running and connected to the same Redis instance. When a message with a TTL is sent, SlackLiner will publish the message details to the configured TimeBomb Redis channel for scheduled deletion.

### Adding Emoji Reactions

To add an emoji reaction to an existing Slack message, push a JSON object to the `slack_reactions` Redis list:

```json
{
  "reaction": "heart_eyes_cat",
  "channel": "C1234567890",
  "ts": "1766282873.772199"
}
```

#### Field Descriptions

- **reaction**: The emoji name without colons (e.g., `thumbsup`, `heart_eyes_cat`, `tada`)
- **channel**: The Slack channel ID (e.g., `C1234567890`)
- **ts**: The message timestamp (obtained when posting a message or from Slack's API)

> **Note**: To add reactions, you need the message timestamp (`ts`) which is returned when posting a message or can be retrieved from Slack's API. The channel should be the channel ID, not the channel name for reactions.

## Slack App Setup

1. Go to [Slack API Apps](https://api.slack.com/apps)
2. Click "Create New App" → "From scratch"
3. Name your app and select your workspace
4. Navigate to "OAuth & Permissions"
5. Add the following Bot Token Scopes:
   - `chat:write` - Send messages
   - `chat:write.public` - Send messages to channels the app isn't in
   - `reactions:write` - Add emoji reactions to messages
6. Install the app to your workspace
7. Copy the "Bot User OAuth Token" (starts with `xoxb-`)
8. Use this token as your `SLACK_BOT_TOKEN`

## Usage Examples

### Using HTTP API (Recommended for Synchronous Operations)

The HTTP API is ideal when you need to get the message timestamp back immediately (e.g., for threading or reactions):

```bash
# Start the service
docker-compose up -d

# Post a message and get the timestamp back
curl -X POST http://localhost:8080/message \
  -H "Content-Type: application/json" \
  -d '{"channel":"#general","text":"Hello from HTTP!"}'

# Response: {"channel":"C1234567890","ts":"1766282873.772199"}

# Use the timestamp to post a threaded reply
curl -X POST http://localhost:8080/message \
  -H "Content-Type: application/json" \
  -d '{
    "channel":"#general",
    "text":"This is a reply",
    "thread_ts":"1766282873.772199"
  }'

# Post a message with metadata
curl -X POST http://localhost:8080/message \
  -H "Content-Type: application/json" \
  -d '{
    "channel":"#general",
    "text":"Task created",
    "metadata":{
      "event_type":"task_created",
      "event_payload":{"task_id":"123","priority":"high"}
    }
  }'

# Post a message with blocks
curl -X POST http://localhost:8080/message \
  -H "Content-Type: application/json" \
  -d '{
    "channel":"#general",
    "blocks":[
      {
        "type":"section",
        "text":{
          "type":"mrkdwn",
          "text":"Hello, *world*! :tada:"
        }
      }
    ]
  }'

# Post a message with blocks and fallback text
curl -X POST http://localhost:8080/message \
  -H "Content-Type: application/json" \
  -d '{
    "channel":"#general",
    "text":"Fallback text for notifications",
    "blocks":[
      {
        "type":"section",
        "text":{
          "type":"mrkdwn",
          "text":"Rich *formatted* message with blocks"
        }
      },
      {
        "type":"divider"
      },
      {
        "type":"section",
        "fields":[
          {"type":"mrkdwn","text":"*Status:*\nCompleted"},
          {"type":"mrkdwn","text":"*Priority:*\nHigh"}
        ]
      }
    ]
  }'
```

### Using Docker Compose with Redis Queue

The Redis queue method is ideal for asynchronous, fire-and-forget messaging:

```bash
# Start services
docker-compose up -d

# Push a message to Redis
docker exec slackliner-redis redis-cli RPUSH slack_messages '{"channel":"#general","text":"Test message"}'

# Add a reaction to a message
docker exec slackliner-redis redis-cli RPUSH slack_reactions '{"reaction":"thumbsup","channel":"C1234567890","ts":"1766282873.772199"}'

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

# Using redis-cli - Message with TTL (auto-delete after 1 hour)
redis-cli RPUSH slack_messages '{"channel":"#general","text":"This message will self-destruct in 1 hour","ttl":3600}'

# Using redis-cli - Message with TTL and metadata
redis-cli RPUSH slack_messages '{"channel":"#general","text":"Alert: High CPU usage","ttl":300,"metadata":{"event_type":"alert","event_payload":{"severity":"high","metric":"cpu"}}}'

# Using redis-cli - Post to a thread
redis-cli RPUSH slack_messages '{"channel":"#general","text":"This is a reply in a thread","thread_ts":"1234567890.123456"}'

# Using redis-cli - Message with blocks
redis-cli RPUSH slack_messages '{"channel":"#general","blocks":[{"type":"section","text":{"type":"mrkdwn","text":"Hello, *world*! :tada:"}}]}'

# Using redis-cli - Message with blocks and text
redis-cli RPUSH slack_messages '{"channel":"#general","text":"Fallback text","blocks":[{"type":"section","text":{"type":"mrkdwn","text":"Rich *formatted* message"}}]}'

# Using redis-cli - Add emoji reaction
redis-cli RPUSH slack_reactions '{"reaction":"heart_eyes_cat","channel":"C1234567890","ts":"1766282873.772199"}'

# Using redis-cli - Add thumbsup reaction
redis-cli RPUSH slack_reactions '{"reaction":"thumbsup","channel":"#general","ts":"1234567890.123456"}'

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

# Message with TTL (auto-delete after 5 minutes)
message_with_ttl = {
    "channel": "#general",
    "text": "This temporary message will be deleted in 5 minutes",
    "ttl": 300
}
r.rpush('slack_messages', json.dumps(message_with_ttl))

# Message with TTL and metadata
message_with_ttl_and_metadata = {
    "channel": "#alerts",
    "text": "Alert: High CPU usage detected",
    "ttl": 600,
    "metadata": {
        "event_type": "alert",
        "event_payload": {
            "severity": "high",
            "metric": "cpu",
            "value": 95
        }
    }
}
r.rpush('slack_messages', json.dumps(message_with_ttl_and_metadata))

# Post to a thread
thread_reply = {
    "channel": "#general",
    "text": "This is a reply in a thread",
    "thread_ts": "1234567890.123456"
}
r.rpush('slack_messages', json.dumps(thread_reply))

# Post to a thread with metadata
thread_reply_with_metadata = {
    "channel": "#general",
    "text": "Thread reply with context",
    "thread_ts": "1234567890.123456",
    "metadata": {
        "event_type": "thread_reply",
        "event_payload": {
            "reply_type": "automated",
            "source": "bot"
        }
    }
}
r.rpush('slack_messages', json.dumps(thread_reply_with_metadata))

# Message with blocks only
message_with_blocks = {
    "channel": "#general",
    "blocks": [
        {
            "type": "section",
            "text": {
                "type": "mrkdwn",
                "text": "Hello, *world*! :tada:"
            }
        }
    ]
}
r.rpush('slack_messages', json.dumps(message_with_blocks))

# Message with blocks and fallback text
message_with_blocks_and_text = {
    "channel": "#general",
    "text": "Fallback text for notifications",
    "blocks": [
        {
            "type": "section",
            "text": {
                "type": "mrkdwn",
                "text": "Rich *formatted* message with blocks"
            }
        },
        {
            "type": "divider"
        },
        {
            "type": "section",
            "fields": [
                {"type": "mrkdwn", "text": "*Status:*\nCompleted"},
                {"type": "mrkdwn", "text": "*Priority:*\nHigh"}
            ]
        }
    ]
}
r.rpush('slack_messages', json.dumps(message_with_blocks_and_text))

# Add emoji reaction
reaction = {
    "reaction": "heart_eyes_cat",
    "channel": "C1234567890",
    "ts": "1766282873.772199"
}
r.rpush('slack_reactions', json.dumps(reaction))

# Add multiple reactions
reactions = [
    {"reaction": "thumbsup", "channel": "C1234567890", "ts": "1766282873.772199"},
    {"reaction": "tada", "channel": "C1234567890", "ts": "1766282873.772199"}
]
for reaction in reactions:
    r.rpush('slack_reactions', json.dumps(reaction))
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
                        ┌──────────────────┐
                        │   HTTP Client    │
                        │  (curl, app...)  │
                        └────────┬─────────┘
                                 │ POST /message
                                 │ (returns ts & channel)
                                 ▼
┌─────────────┐         ┌──────────────────┐         ┌─────────────┐
│   Producer  │ RPUSH   │  Redis Queues    │  BLPOP  │             │
│   (Any App) ├────────►│  - Messages      ├────────►│ SlackLiner  │
│             │         │  - Reactions     │         │   Service   │
└─────────────┘         └──────────────────┘         └──────┬──────┘
                                                             │
                                                             │ POST/React
                                                             ▼
                                                     ┌─────────────┐
                                                     │  Slack API  │
                                                     └─────────────┘
```

**Two Ways to Send Messages:**

1. **HTTP API** (synchronous): POST to `/message` endpoint and receive channel ID and timestamp in response
   - Ideal when you need the message timestamp for threading or reactions
   - Returns immediately with message details
   
2. **Redis Queue** (asynchronous): Push messages to Redis using RPUSH
   - Ideal for fire-and-forget messaging
   - Decoupled from the sender

**Components:**

1. **HTTP Client**: Any HTTP client (curl, application) that can POST JSON
2. **Producer**: Any application that can write to Redis (API server, cron job, etc.)
3. **Redis Queues**: 
   - `slack_messages` - Stores message payloads using RPUSH
   - `slack_reactions` - Stores reaction payloads using RPUSH
4. **SlackLiner Service**: 
   - HTTP server on port 8080 (configurable)
   - Reads from both queues using BLPOP (blocking, efficient)
   - Processes messages and posts to Slack channels
   - Processes reactions and adds emoji reactions to existing messages
5. **Slack API**: Receives messages and reactions via the Slack Bot API

## Project Structure

The codebase is organized into focused modules:

- `main.go` - Application entry point and configuration
- `types.go` - Type definitions for messages, reactions, and responses
- `http.go` - HTTP server and endpoint handlers
- `redis.go` - Redis queue processing logic
- `slack.go` - Slack API operations (posting messages, adding reactions)
- `main_test.go` - Comprehensive test suite

## License

MIT License - See LICENSE file for details
