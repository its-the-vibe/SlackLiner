#!/bin/bash
# Example script to push messages to Redis for SlackLiner to process

REDIS_HOST="${REDIS_HOST:-localhost}"
REDIS_PORT="${REDIS_PORT:-6379}"
REDIS_LIST_KEY="${REDIS_LIST_KEY:-slack_messages}"

# Function to push a message
push_message() {
    local channel="$1"
    local text="$2"
    local event_type="$3"
    local event_payload="$4"
    
    if [ -z "$channel" ] || [ -z "$text" ]; then
        echo "Usage: push_message <channel> <text> [event_type] [event_payload_json]"
        return 1
    fi
    
    local json_message
    if [ -n "$event_type" ]; then
        # Message with metadata
        if [ -n "$event_payload" ]; then
            json_message=$(cat <<EOF
{"channel":"$channel","text":"$text","metadata":{"event_type":"$event_type","event_payload":$event_payload}}
EOF
)
        else
            json_message=$(cat <<EOF
{"channel":"$channel","text":"$text","metadata":{"event_type":"$event_type","event_payload":{}}}
EOF
)
        fi
    else
        # Simple message without metadata
        json_message=$(cat <<EOF
{"channel":"$channel","text":"$text"}
EOF
)
    fi
    
    docker exec slackliner-redis redis-cli RPUSH "$REDIS_LIST_KEY" "$json_message"
    echo "Message pushed to Redis queue"
}

# Example usage
if [ "$#" -ge 2 ]; then
    push_message "$1" "$2" "$3" "$4"
else
    echo "Usage: $0 <channel> <text> [event_type] [event_payload_json]"
    echo "Example (simple): $0 '#general' 'Hello from SlackLiner!'"
    echo "Example (with metadata): $0 '#general' 'Task created' 'task_created' '{\"task_id\":\"123\",\"priority\":\"high\"}'"
    exit 1
fi
