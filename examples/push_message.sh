#!/bin/bash
# Example script to push messages to Redis for SlackLiner to process

REDIS_HOST="${REDIS_HOST:-localhost}"
REDIS_PORT="${REDIS_PORT:-6379}"
REDIS_LIST_KEY="${REDIS_LIST_KEY:-slack_messages}"

# Function to push a message
push_message() {
    local channel="$1"
    local text="$2"
    
    if [ -z "$channel" ] || [ -z "$text" ]; then
        echo "Usage: push_message <channel> <text>"
        return 1
    fi
    
    local json_message=$(cat <<EOF
{"channel":"$channel","text":"$text"}
EOF
)
    
    docker exec slackliner-redis redis-cli RPUSH "$REDIS_LIST_KEY" "$json_message"
    echo "Message pushed to Redis queue"
}

# Example usage
if [ "$#" -eq 2 ]; then
    push_message "$1" "$2"
else
    echo "Usage: $0 <channel> <text>"
    echo "Example: $0 '#general' 'Hello from SlackLiner!'"
    exit 1
fi
