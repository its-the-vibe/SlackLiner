package main

// MessageMetadata represents optional metadata to attach to a Slack message
type MessageMetadata struct {
	EventType    string                 `json:"event_type"`
	EventPayload map[string]interface{} `json:"event_payload"`
}

// TimeBombMessage represents the message structure to send to TimeBomb for scheduled deletion
type TimeBombMessage struct {
	Channel string `json:"channel"`
	TS      string `json:"ts"`
	TTL     int    `json:"ttl"`
}

// SlackMessage represents the payload structure expected from Redis for posting messages
type SlackMessage struct {
	Channel  string           `json:"channel"`
	Text     string           `json:"text"`
	ThreadTS string           `json:"thread_ts,omitempty"` // Thread timestamp to reply to an existing thread
	Metadata *MessageMetadata `json:"metadata,omitempty"`
	TTL      int              `json:"ttl,omitempty"` // Time-to-live in seconds for automatic deletion via TimeBomb
}

// ReactionMessage represents the payload structure for adding emoji reactions
type ReactionMessage struct {
	Reaction string `json:"reaction"` // Emoji name without colons (e.g., "heart_eyes_cat")
	Channel  string `json:"channel"`  // Channel ID (e.g., "C1234567890")
	TS       string `json:"ts"`       // Message timestamp
}
