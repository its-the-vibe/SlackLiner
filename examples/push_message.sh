#!/bin/bash
# Example script to push messages to Redis for SlackLiner to process

REDIS_HOST="${REDIS_HOST:-localhost}"
REDIS_PORT="${REDIS_PORT:-6379}"
REDIS_LIST_KEY="${REDIS_LIST_KEY:-slack_messages}"

# Function to push a message
push_message() {
    local channel="$1"
    local text="$2"
    local ttl="$3"
    local event_type="$4"
    local event_payload="$5"
    
    if [ -z "$channel" ] || [ -z "$text" ]; then
        echo "Usage: push_message <channel> <text> [ttl] [event_type] [event_payload_json]"
        return 1
    fi
    
    local json_message
    
    # Build JSON based on provided parameters
    if [ -n "$event_type" ]; then
        # Message with metadata
        if [ -n "$event_payload" ]; then
            if [ -n "$ttl" ]; then
                json_message=$(cat <<EOF
{"channel":"$channel","text":"$text","ttl":$ttl,"metadata":{"event_type":"$event_type","event_payload":$event_payload}}
EOF
)
            else
                json_message=$(cat <<EOF
{"channel":"$channel","text":"$text","metadata":{"event_type":"$event_type","event_payload":$event_payload}}
EOF
)
            fi
        else
            if [ -n "$ttl" ]; then
                json_message=$(cat <<EOF
{"channel":"$channel","text":"$text","ttl":$ttl,"metadata":{"event_type":"$event_type","event_payload":{}}}
EOF
)
            else
                json_message=$(cat <<EOF
{"channel":"$channel","text":"$text","metadata":{"event_type":"$event_type","event_payload":{}}}
EOF
)
            fi
        fi
    elif [ -n "$ttl" ]; then
        # Message with TTL but no metadata
        json_message=$(cat <<EOF
{"channel":"$channel","text":"$text","ttl":$ttl}
EOF
)
    else
        # Simple message without metadata or TTL
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
    push_message "$1" "$2" "$3" "$4" "$5"
else
    echo "Usage: $0 <channel> <text> [ttl] [event_type] [event_payload_json]"
    echo "Example (simple): $0 '#general' 'Hello from SlackLiner!'"
    echo "Example (with TTL): $0 '#general' 'This will be deleted in 5 minutes' 300"
    echo "Example (with metadata only): $0 '#general' 'Task created' 0 'task_created' '{\"task_id\":\"123\",\"priority\":\"high\"}'"
    echo "Example (with TTL and metadata): $0 '#alerts' 'High CPU alert' 600 'alert' '{\"severity\":\"high\",\"metric\":\"cpu\"}'"
    exit 1
fi
