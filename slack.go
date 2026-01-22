package main

import (
	"context"
	"encoding/json"
	"log"

	"github.com/redis/go-redis/v9"
	"github.com/slack-go/slack"
)

// sendSlackMessageWithResponse sends a message to Slack and returns the channel and timestamp
// Returns error if message validation fails or Slack API call fails
func sendSlackMessageWithResponse(ctx context.Context, slackClient *slack.Client, rdb *redis.Client, msg SlackMessage, timeBombChannel string) (string, string, error) {
	// Validate message - channel is required, and either text or blocks must be provided
	if msg.Channel == "" {
		log.Printf("Invalid message: channel is required. Got: %+v", msg)
		return "", "", ErrInvalidMessage
	}

	if msg.Text == "" && len(msg.Blocks) == 0 {
		log.Printf("Invalid message: either text or blocks are required. Got: %+v", msg)
		return "", "", ErrInvalidMessage
	}

	// Validate TTL if provided
	if msg.TTL < 0 {
		log.Printf("Invalid message: ttl must be non-negative if provided. Got: %+v", msg)
		return "", "", ErrInvalidTTL
	}

	// Send to Slack
	if msg.Text != "" {
		log.Printf("Sending message to channel '%s': %s", msg.Channel, msg.Text)
	} else {
		log.Printf("Sending message with blocks to channel '%s'", msg.Channel)
	}

	// Build message options
	msgOptions := []slack.MsgOption{
		slack.MsgOptionDisableLinkUnfurl(),
	}

	// Add text if provided
	if msg.Text != "" {
		msgOptions = append(msgOptions, slack.MsgOptionText(msg.Text, false))
	}

	// Add blocks if provided
	if len(msg.Blocks) > 0 {
		var blocks slack.Blocks
		if err := json.Unmarshal(msg.Blocks, &blocks); err != nil {
			log.Printf("Error unmarshaling blocks JSON: %v. Raw blocks: %s", err, string(msg.Blocks))
			return "", "", err
		}
		if blocks.BlockSet != nil && len(blocks.BlockSet) > 0 {
			log.Printf("Including %d blocks", len(blocks.BlockSet))
			msgOptions = append(msgOptions, slack.MsgOptionBlocks(blocks.BlockSet...))
		} else {
			log.Printf("Warning: blocks field provided but no valid blocks found")
		}
	}

	// Add thread_ts if provided (for posting to a thread)
	if msg.ThreadTS != "" {
		log.Printf("Posting as reply to thread: %s", msg.ThreadTS)
		msgOptions = append(msgOptions, slack.MsgOptionTS(msg.ThreadTS))
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
		return "", "", err
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

	return channelID, timestamp, nil
}

// sendSlackMessage sends a message to Slack and optionally publishes to TimeBomb for deletion
// This is a wrapper around sendSlackMessageWithResponse that discards the return values
// for use with the Redis queue where we don't need to return the response to the caller
func sendSlackMessage(ctx context.Context, slackClient *slack.Client, rdb *redis.Client, msg SlackMessage, timeBombChannel string) {
	// Errors are already logged by sendSlackMessageWithResponse, so we don't need to handle them here
	_, _, _ = sendSlackMessageWithResponse(ctx, slackClient, rdb, msg, timeBombChannel)
}

// addSlackReaction adds or removes an emoji reaction to/from a Slack message
func addSlackReaction(slackClient *slack.Client, msg ReactionMessage) {
	// Validate message
	if msg.Reaction == "" || msg.Channel == "" || msg.TS == "" {
		log.Printf("Invalid reaction message: reaction, channel, and ts are required. Got: %+v", msg)
		return
	}

	itemRef := slack.ItemRef{
		Channel:   msg.Channel,
		Timestamp: msg.TS,
	}

	if msg.Remove {
		// Remove reaction from Slack
		log.Printf("Removing reaction '%s' from message in channel '%s' at timestamp '%s'", msg.Reaction, msg.Channel, msg.TS)

		err := slackClient.RemoveReaction(msg.Reaction, itemRef)
		if err != nil {
			log.Printf("Error removing reaction from Slack: %v", err)
			return
		}

		log.Printf("Reaction '%s' removed successfully from channel %s (timestamp: %s)", msg.Reaction, msg.Channel, msg.TS)
	} else {
		// Add reaction to Slack
		log.Printf("Adding reaction '%s' to message in channel '%s' at timestamp '%s'", msg.Reaction, msg.Channel, msg.TS)

		err := slackClient.AddReaction(msg.Reaction, itemRef)
		if err != nil {
			log.Printf("Error adding reaction to Slack: %v", err)
			return
		}

		log.Printf("Reaction '%s' added successfully to channel %s (timestamp: %s)", msg.Reaction, msg.Channel, msg.TS)
	}
}
