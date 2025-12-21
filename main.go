package main

import (
	"context"
	"encoding/json"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/slack-go/slack"
)

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

// SlackMessage represents the payload structure expected from Redis
type SlackMessage struct {
	Channel  string           `json:"channel"`
	Text     string           `json:"text"`
	Metadata *MessageMetadata `json:"metadata,omitempty"`
	TTL      int              `json:"ttl,omitempty"` // Time-to-live in seconds for automatic deletion via TimeBomb
}

func main() {
	// Load configuration from environment variables
	redisAddr := getEnv("REDIS_ADDR", "localhost:6379")
	redisPassword := getEnv("REDIS_PASSWORD", "")
	redisDB := 0 // Using default DB
	redisListKey := getEnv("REDIS_LIST_KEY", "slack_messages")
	timeBombChannel := getEnv("TIMEBOMB_REDIS_CHANNEL", "timebomb-messages")
	slackToken := getEnv("SLACK_BOT_TOKEN", "")

	if slackToken == "" {
		log.Fatal("SLACK_BOT_TOKEN environment variable is required")
	}

	// Initialize Redis client
	log.Printf("Connecting to Redis at %s...", redisAddr)
	rdb := redis.NewClient(&redis.Options{
		Addr:     redisAddr,
		Password: redisPassword,
		DB:       redisDB,
	})
	defer rdb.Close()

	// Test Redis connection
	ctx := context.Background()
	if err := rdb.Ping(ctx).Err(); err != nil {
		log.Fatalf("Failed to connect to Redis: %v", err)
	}
	log.Println("Connected to Redis successfully")

	// Initialize Slack client
	log.Println("Initializing Slack client...")
	slackClient := slack.New(slackToken)

	// Test Slack connection
	if _, err := slackClient.AuthTest(); err != nil {
		log.Fatalf("Failed to authenticate with Slack: %v", err)
	}
	log.Println("Slack authentication successful")

	// Setup graceful shutdown with context
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	// Start message processing loop
	log.Printf("Starting to listen for messages on Redis list '%s'...", redisListKey)
	go processMessages(ctx, rdb, slackClient, redisListKey, timeBombChannel)

	// Wait for shutdown signal
	<-sigChan
	log.Println("Shutting down gracefully...")
	cancel()
	time.Sleep(1 * time.Second) // Give goroutine time to finish current operation
}

func processMessages(ctx context.Context, rdb *redis.Client, slackClient *slack.Client, listKey string, timeBombChannel string) {
	for {
		select {
		case <-ctx.Done():
			log.Println("Message processing stopped")
			return
		default:
			// BLPOP blocks until a message is available or timeout occurs
			result, err := rdb.BLPop(ctx, 5*time.Second, listKey).Result()
			if err == redis.Nil {
				// Timeout, no message available
				continue
			} else if err != nil {
				log.Printf("Error reading from Redis: %v", err)
				time.Sleep(1 * time.Second)
				continue
			}

			// result[0] is the key, result[1] is the value
			if len(result) < 2 {
				log.Println("Invalid result from Redis BLPOP")
				continue
			}

			messageData := result[1]

			// Parse the message
			var msg SlackMessage
			if err := json.Unmarshal([]byte(messageData), &msg); err != nil {
				log.Printf("Error parsing message JSON: %v, data: %s", err, messageData)
				continue
			}

			// Validate message
			if msg.Channel == "" || msg.Text == "" {
				log.Printf("Invalid message: channel and text are required. Got: %+v", msg)
				continue
			}

			// Validate TTL if provided
			if msg.TTL < 0 {
				log.Printf("Invalid message: ttl must be non-negative if provided. Got: %+v", msg)
				continue
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
				continue
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
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
