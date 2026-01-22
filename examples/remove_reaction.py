#!/usr/bin/env python3
"""
Example script to remove emoji reactions via Redis for SlackLiner to process.
Requires: pip install redis
"""

import json
import os
import sys
import redis


def remove_reaction(reaction: str, channel: str, ts: str, redis_host: str = "localhost",
                    redis_port: int = 6379, redis_list_key: str = "slack_reactions") -> bool:
    """
    Remove an emoji reaction from a Slack message via Redis queue.
    
    Args:
        reaction: Emoji name without colons (e.g., 'thumbsup', 'heart_eyes_cat')
        channel: Slack channel ID (e.g., 'C1234567890')
        ts: Message timestamp (e.g., '1766282873.772199')
        redis_host: Redis host address
        redis_port: Redis port number
        redis_list_key: Redis list key to push to
    
    Returns:
        bool: True if successful, False otherwise
    """
    try:
        # Connect to Redis
        r = redis.Redis(host=redis_host, port=redis_port, decode_responses=True)
        
        # Create reaction removal payload
        reaction_message = {
            "reaction": reaction,
            "channel": channel,
            "ts": ts,
            "remove": True
        }
        
        # Push to Redis list
        r.rpush(redis_list_key, json.dumps(reaction_message))
        print(f"✓ Reaction '{reaction}' removal request pushed to Redis queue '{redis_list_key}'")
        print(f"  Channel: {channel}")
        print(f"  Timestamp: {ts}")
        return True
        
    except redis.ConnectionError as e:
        print(f"✗ Failed to connect to Redis: {e}")
        return False
    except Exception as e:
        print(f"✗ Error: {e}")
        return False


if __name__ == "__main__":
    if len(sys.argv) != 4:
        print("Usage: python remove_reaction.py <reaction> <channel> <timestamp>")
        print()
        print("Arguments:")
        print("  reaction   - Emoji name without colons (e.g., thumbsup, heart_eyes_cat, tada)")
        print("  channel    - Slack channel ID (e.g., C1234567890)")
        print("  timestamp  - Message timestamp (e.g., 1766282873.772199)")
        print()
        print("Examples:")
        print("  python remove_reaction.py thumbsup C1234567890 1766282873.772199")
        print("  python remove_reaction.py heart_eyes_cat C9876543210 1234567890.123456")
        print("  python remove_reaction.py tada C1234567890 9999999999.999999")
        print()
        print("Note: This removes a reaction from a message. The reaction must have been")
        print("      previously added to the message. You can get the timestamp (ts) from")
        print("      the message posting response or by inspecting the message in Slack's API.")
        sys.exit(1)
    
    reaction = sys.argv[1]
    channel = sys.argv[2]
    ts = sys.argv[3]
    
    # Get config from environment variables
    redis_host = os.getenv("REDIS_HOST", "localhost")
    redis_port = int(os.getenv("REDIS_PORT", "6379"))
    redis_list_key = os.getenv("REDIS_REACTION_LIST_KEY", "slack_reactions")
    
    success = remove_reaction(reaction, channel, ts, redis_host, redis_port, redis_list_key)
    sys.exit(0 if success else 1)
