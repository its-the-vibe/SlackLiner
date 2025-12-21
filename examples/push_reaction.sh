#!/bin/bash
# Example script to push emoji reactions to Redis for SlackLiner to process

REDIS_HOST="${REDIS_HOST:-localhost}"
REDIS_PORT="${REDIS_PORT:-6379}"
REDIS_REACTION_LIST_KEY="${REDIS_REACTION_LIST_KEY:-slack_reactions}"

# Function to push a reaction
push_reaction() {
    local reaction="$1"
    local channel="$2"
    local ts="$3"
    
    if [ -z "$reaction" ] || [ -z "$channel" ] || [ -z "$ts" ]; then
        echo "Usage: push_reaction <reaction> <channel> <timestamp>"
        return 1
    fi
    
    local json_message=$(cat <<EOF
{"reaction":"$reaction","channel":"$channel","ts":"$ts"}
EOF
)
    
    docker exec slackliner-redis redis-cli RPUSH "$REDIS_REACTION_LIST_KEY" "$json_message"
    echo "Reaction pushed to Redis queue"
}

# Example usage
if [ "$#" -eq 3 ]; then
    push_reaction "$1" "$2" "$3"
else
    echo "Usage: $0 <reaction> <channel> <timestamp>"
    echo ""
    echo "Arguments:"
    echo "  reaction   - Emoji name without colons (e.g., thumbsup, heart_eyes_cat, tada)"
    echo "  channel    - Slack channel ID (e.g., C1234567890)"
    echo "  timestamp  - Message timestamp (e.g., 1766282873.772199)"
    echo ""
    echo "Examples:"
    echo "  $0 thumbsup C1234567890 1766282873.772199"
    echo "  $0 heart_eyes_cat C9876543210 1234567890.123456"
    echo "  $0 tada C1234567890 9999999999.999999"
    echo ""
    echo "Note: You can get the timestamp (ts) from the message posting response"
    echo "      or by inspecting the message in Slack's API."
    exit 1
fi
