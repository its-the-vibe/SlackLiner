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
                 redis_port: int = 6379, redis_list_key: str = "slack_messages",
                 ttl: int = None, metadata: dict = None) -> bool:
    """
    Push a Slack message to Redis queue.
    
    Args:
        channel: Slack channel name (e.g., '#general') or ID
        text: Message text to send
        redis_host: Redis host address
        redis_port: Redis port number
        redis_list_key: Redis list key to push to
        ttl: Optional time-to-live in seconds for automatic deletion via TimeBomb
        metadata: Optional metadata dict with 'event_type' and 'event_payload' keys
    
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
        
        # Add TTL if provided
        if ttl is not None and ttl > 0:
            message["ttl"] = ttl
        
        # Add metadata if provided
        if metadata:
            message["metadata"] = metadata
        
        # Push to Redis list
        r.rpush(redis_list_key, json.dumps(message))
        print(f"✓ Message pushed to Redis queue '{redis_list_key}'")
        if ttl:
            print(f"  with TTL: {ttl} seconds")
        if metadata:
            print(f"  with metadata event_type: {metadata.get('event_type', 'N/A')}")
        return True
        
    except redis.ConnectionError as e:
        print(f"✗ Failed to connect to Redis: {e}")
        return False
    except Exception as e:
        print(f"✗ Error: {e}")
        return False


if __name__ == "__main__":
    if len(sys.argv) < 3:
        print("Usage: python push_message.py <channel> <text> [ttl] [event_type] [event_payload_json]")
        print("Example (simple): python push_message.py '#general' 'Hello from SlackLiner!'")
        print("Example (with TTL): python push_message.py '#general' 'This will be deleted in 5 min' 300")
        print("Example (with metadata): python push_message.py '#general' 'Task created' '' 'task_created' '{\"task_id\":\"123\",\"priority\":\"high\"}'")
        print("Example (with TTL and metadata): python push_message.py '#alerts' 'High CPU alert' 600 'alert' '{\"severity\":\"high\",\"metric\":\"cpu\"}'")
        sys.exit(1)
    
    channel = sys.argv[1]
    text = sys.argv[2]
    
    # Parse optional TTL
    ttl = None
    if len(sys.argv) >= 4 and sys.argv[3]:
        try:
            ttl = int(sys.argv[3])
            if ttl < 0:
                print(f"✗ TTL must be non-negative, got: {ttl}")
                sys.exit(1)
        except ValueError as e:
            print(f"✗ Invalid TTL (must be an integer): {e}")
            sys.exit(1)
    
    # Parse optional metadata
    metadata = None
    if len(sys.argv) >= 5 and sys.argv[4]:
        event_type = sys.argv[4]
        event_payload = {}
        if len(sys.argv) >= 6:
            try:
                event_payload = json.loads(sys.argv[5])
            except json.JSONDecodeError as e:
                print(f"✗ Invalid JSON for event_payload: {e}")
                sys.exit(1)
        
        metadata = {
            "event_type": event_type,
            "event_payload": event_payload
        }
    
    # Get config from environment variables
    redis_host = os.getenv("REDIS_HOST", "localhost")
    redis_port = int(os.getenv("REDIS_PORT", "6379"))
    redis_list_key = os.getenv("REDIS_LIST_KEY", "slack_messages")
    
    success = push_message(channel, text, redis_host, redis_port, redis_list_key, ttl, metadata)
    sys.exit(0 if success else 1)
