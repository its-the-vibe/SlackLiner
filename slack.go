package main

import (
	"context"
	"encoding/json"
	"log"

	"github.com/redis/go-redis/v9"
	"github.com/slack-go/slack"
)

// sendSlackMessage sends a message to Slack and optionally publishes to TimeBomb for deletion
func sendSlackMessage(ctx context.Context, slackClient *slack.Client, rdb *redis.Client, msg SlackMessage, timeBombChannel string) {
	// Validate message
	if msg.Channel == "" || msg.Text == "" {
		log.Printf("Invalid message: channel and text are required. Got: %+v", msg)
		return
	}

	// Validate TTL if provided
	if msg.TTL < 0 {
		log.Printf("Invalid message: ttl must be non-negative if provided. Got: %+v", msg)
		return
	}

	// Send to Slack
	log.Printf("Sending message to channel '%s': %s", msg.Channel, msg.Text)

	// Build message options
	msgOptions := []slack.MsgOption{
		slack.MsgOptionText(msg.Text, false),
		slack.MsgOptionDisableLinkUnfurl(),
	}

	// Add metadata if provided
	if msg.Metadata != nil {
		log.Printf("Including metadata with event_type: %s", msg.Metadata.EventType)
		msgOptions = append(msgOptions, slack.MsgOptionMetadata(slack.SlackMetadata{
			EventType:    msg.Metadata.EventType,
			EventPayload: msg.Metadata.EventPayload,
		}))
	}

	channelID, timestamp, err := slackClient.PostMessage(msg.Channel, msgOptions...)
	if err != nil {
		log.Printf("Error posting to Slack: %v", err)
		return
	}

	log.Printf("Message sent successfully to channel %s (timestamp: %s)", channelID, timestamp)

	// If TTL is specified, publish to TimeBomb for scheduled deletion
	if msg.TTL > 0 {
		tbMsg := TimeBombMessage{
			Channel: channelID,
			TS:      timestamp,
			TTL:     msg.TTL,
		}

		tbPayload, err := json.Marshal(tbMsg)
		if err != nil {
			log.Printf("Error marshaling TimeBomb message: %v", err)
		} else {
			err = rdb.Publish(ctx, timeBombChannel, string(tbPayload)).Err()
			if err != nil {
				log.Printf("Error publishing to TimeBomb channel '%s': %v", timeBombChannel, err)
			} else {
				log.Printf("Published to TimeBomb for deletion: channel=%s, ts=%s, ttl=%ds", channelID, timestamp, msg.TTL)
			}
		}
	}
}

// addSlackReaction adds an emoji reaction to a Slack message
func addSlackReaction(slackClient *slack.Client, msg ReactionMessage) {
	// Validate message
	if msg.Reaction == "" || msg.Channel == "" || msg.TS == "" {
		log.Printf("Invalid reaction message: reaction, channel, and ts are required. Got: %+v", msg)
		return
	}

	// Add reaction to Slack
	log.Printf("Adding reaction '%s' to message in channel '%s' at timestamp '%s'", msg.Reaction, msg.Channel, msg.TS)

	itemRef := slack.ItemRef{
		Channel:   msg.Channel,
		Timestamp: msg.TS,
	}

	err := slackClient.AddReaction(msg.Reaction, itemRef)
	if err != nil {
		log.Printf("Error adding reaction to Slack: %v", err)
		return
	}

	log.Printf("Reaction '%s' added successfully to channel %s (timestamp: %s)", msg.Reaction, msg.Channel, msg.TS)
}
