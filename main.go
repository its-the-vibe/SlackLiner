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

// SlackMessage represents the payload structure expected from Redis
type SlackMessage struct {
	Channel string `json:"channel"`
	Text    string `json:"text"`
}

func main() {
	// Load configuration from environment variables
	redisAddr := getEnv("REDIS_ADDR", "localhost:6379")
	redisPassword := getEnv("REDIS_PASSWORD", "")
	redisDB := 0 // Using default DB
	redisListKey := getEnv("REDIS_LIST_KEY", "slack_messages")
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
	go processMessages(ctx, rdb, slackClient, redisListKey)

	// Wait for shutdown signal
	<-sigChan
	log.Println("Shutting down gracefully...")
	cancel()
	time.Sleep(1 * time.Second) // Give goroutine time to finish current operation
}

func processMessages(ctx context.Context, rdb *redis.Client, slackClient *slack.Client, listKey string) {
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

			// Send to Slack
			log.Printf("Sending message to channel '%s': %s", msg.Channel, msg.Text)
			channelID, timestamp, err := slackClient.PostMessage(
				msg.Channel,
				slack.MsgOptionText(msg.Text, false),
			)
			if err != nil {
				log.Printf("Error posting to Slack: %v", err)
				continue
			}

			log.Printf("Message sent successfully to channel %s (timestamp: %s)", channelID, timestamp)
		}
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
