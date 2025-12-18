#!/usr/bin/env python3
"""
Example script to push messages to Redis for SlackLiner to process.
Requires: pip install redis
"""

import json
import os
import sys
import redis


def push_message(channel: str, text: str, redis_host: str = "localhost", 
                 redis_port: int = 6379, redis_list_key: str = "slack_messages") -> bool:
    """
    Push a Slack message to Redis queue.
    
    Args:
        channel: Slack channel name (e.g., '#general') or ID
        text: Message text to send
        redis_host: Redis host address
        redis_port: Redis port number
        redis_list_key: Redis list key to push to
    
    Returns:
        bool: True if successful, False otherwise
    """
    try:
        # Connect to Redis
        r = redis.Redis(host=redis_host, port=redis_port, decode_responses=True)
        
        # Create message payload
        message = {
            "channel": channel,
            "text": text
        }
        
        # Push to Redis list
        r.rpush(redis_list_key, json.dumps(message))
        print(f"✓ Message pushed to Redis queue '{redis_list_key}'")
        return True
        
    except redis.ConnectionError as e:
        print(f"✗ Failed to connect to Redis: {e}")
        return False
    except Exception as e:
        print(f"✗ Error: {e}")
        return False


if __name__ == "__main__":
    if len(sys.argv) != 3:
        print("Usage: python push_message.py <channel> <text>")
        print("Example: python push_message.py '#general' 'Hello from SlackLiner!'")
        sys.exit(1)
    
    channel = sys.argv[1]
    text = sys.argv[2]
    
    # Get config from environment variables
    redis_host = os.getenv("REDIS_HOST", "localhost")
    redis_port = int(os.getenv("REDIS_PORT", "6379"))
    redis_list_key = os.getenv("REDIS_LIST_KEY", "slack_messages")
    
    success = push_message(channel, text, redis_host, redis_port, redis_list_key)
    sys.exit(0 if success else 1)
