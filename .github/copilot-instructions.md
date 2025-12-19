# SlackLiner Project Instructions

## Project Overview
SlackLiner is a lightweight Go service that reads Slack message payloads from a Redis list and publishes them to Slack using the Slack API. It uses Redis BLPOP for efficient message queue processing and supports optional message metadata.

**Key Technologies:**
- Go 1.24+
- Redis for message queuing
- Slack API (slack-go/slack library)
- Docker with multi-stage builds
- Docker Compose for local development

**Architecture:**
- Producer applications push JSON messages to a Redis list using RPUSH
- SlackLiner reads messages using BLPOP (blocking, efficient)
- Messages are sent to Slack via the Bot API with optional metadata support

## Coding Standards

### Go Conventions
- Follow standard Go formatting (use `go fmt`)
- Use meaningful variable and function names
- Keep functions focused and single-purpose
- Use `context.Context` for cancellation and timeouts
- Handle errors explicitly; never ignore errors
- Use `log.Printf` for logging with clear, actionable messages

### Code Style
- Use tabs for indentation (Go standard)
- Use camelCase for unexported names, PascalCase for exported names
- Group imports: standard library, then third-party, then local
- Add comments for exported functions and types
- Keep line length reasonable (typically under 100 characters)

### Error Handling
- Always check and handle errors appropriately
- Log errors with sufficient context for debugging
- Use `log.Fatal` only for initialization failures that prevent startup
- For runtime errors, log and continue processing when appropriate

## Testing Requirements

### Test Coverage
- Write unit tests for all business logic
- Test files should be named `*_test.go`
- Use table-driven tests for testing multiple scenarios
- Run tests with: `go test -v ./...`

### Test Structure
- Follow the existing pattern in `main_test.go`
- Use subtests with `t.Run()` for organized test cases
- Include both positive and negative test cases
- Test JSON marshaling/unmarshaling thoroughly

### Running Tests
```bash
# Run all tests
go test -v ./...

# Run tests with coverage
go test -v -cover ./...

# Run specific test
go test -v -run TestName
```

## Build and Deployment

### Local Development
```bash
# Install dependencies
go mod download

# Build locally
go build -o slackliner .

# Run locally (requires Redis and Slack token)
export SLACK_BOT_TOKEN=xoxb-your-token
export REDIS_ADDR=localhost:6379
./slackliner
```

### Docker
- Use the provided `Dockerfile` for building
- Multi-stage build: build in golang:1.24, run in scratch
- Keep the image minimal and secure
- Build command: `docker build -t slackliner .`

### Docker Compose
- Use `docker-compose.yml` for local testing
- Includes Redis service automatically
- Start with: `docker-compose up -d`
- View logs with: `docker-compose logs -f slackliner`

## Dependencies Management

### Adding Dependencies
- Use `go get` to add new dependencies
- Run `go mod tidy` to clean up dependencies
- Only add necessary, well-maintained libraries
- Current key dependencies:
  - `github.com/redis/go-redis/v9` - Redis client
  - `github.com/slack-go/slack` - Slack API client

### Updating Dependencies
- Review changes before updating major versions
- Test thoroughly after dependency updates
- Keep Go version in `go.mod` current

## Configuration

### Environment Variables
- All configuration via environment variables
- Required: `SLACK_BOT_TOKEN`
- Optional: `REDIS_ADDR`, `REDIS_PASSWORD`, `REDIS_LIST_KEY`
- Document new environment variables in README.md
- Use `getEnv()` helper for defaults

### Secrets Management
- Never commit secrets or tokens to the repository
- Use `.env` for local development (in `.gitignore`)
- Provide `.env.example` with placeholder values
- Document all required secrets in README.md

## Message Format

### JSON Structure
Messages must be valid JSON with:
- `channel` (required): Slack channel ID or name (e.g., "#general")
- `text` (required): Message text content
- `metadata` (optional): Custom metadata object
  - `event_type`: String identifier for event type
  - `event_payload`: Map of custom data (strings, numbers, booleans)

### Validation
- Validate that `channel` and `text` are non-empty
- Log and skip invalid messages (don't crash)
- Test JSON parsing with edge cases

## Security Considerations

### API Tokens
- Store Slack token securely in environment variables
- Never log the full token value
- Validate token on startup with `AuthTest()`

### Redis Connection
- Support Redis password authentication
- Handle Redis connection failures gracefully
- Use timeouts for Redis operations (currently 5s for BLPOP)

### Docker Security
- Use scratch base image for minimal attack surface
- Run as non-root user if possible
- Don't expose unnecessary ports

### Input Validation
- Validate all JSON inputs before processing
- Check for required fields (channel, text)
- Handle malformed JSON gracefully

## Graceful Shutdown

### Signal Handling
- Listen for SIGINT and SIGTERM
- Use context cancellation for clean shutdown
- Allow in-flight operations to complete (1s grace period)
- Log shutdown events clearly

## Common Patterns

### Logging
- Use descriptive log messages with context
- Log levels:
  - `log.Printf()` for informational messages
  - `log.Fatal()` for startup failures only
  - Include relevant data (channel, timestamp, error details)

### Context Usage
- Pass `context.Context` to functions that perform I/O
- Respect context cancellation
- Use context for timeouts and deadlines

### Redis Operations
- Use `BLPOP` for blocking reads (efficient)
- Handle `redis.Nil` (timeout) gracefully
- Retry on transient errors with backoff (1s delay)

## Documentation

### Code Comments
- Comment exported types and functions
- Explain non-obvious logic or business rules
- Keep comments up-to-date with code changes

### README Updates
- Update README.md when adding features
- Include usage examples for new functionality
- Document new environment variables
- Keep architecture diagram current

### Commit Messages
- Use conventional commit format when possible
- First line: concise summary (<50 chars)
- Body: explain what and why, not how
- Reference issues when applicable

## Examples and Testing

### Manual Testing
Use the provided examples in the `examples/` directory for manual testing scenarios.

### Test Messages
```bash
# Simple message
docker exec slackliner-redis redis-cli RPUSH slack_messages '{"channel":"#general","text":"Hello World"}'

# Message with metadata
docker exec slackliner-redis redis-cli RPUSH slack_messages '{"channel":"#general","text":"Task created","metadata":{"event_type":"task_created","event_payload":{"task_id":"123"}}}'
```

## Best Practices

1. **Keep it Simple**: This is a focused service; avoid scope creep
2. **Fail Fast**: Validate inputs and fail early with clear errors
3. **Be Explicit**: Prefer explicit code over clever tricks
4. **Test Thoroughly**: Test both success and failure paths
5. **Log Intelligently**: Log enough to debug, not so much it's noise
6. **Handle Errors**: Never ignore errors; handle them appropriately
7. **Document Changes**: Update README and comments when modifying behavior
