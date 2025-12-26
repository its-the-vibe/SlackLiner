package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/slack-go/slack"
)

func main() {
	// Load configuration from environment variables
	redisAddr := getEnv("REDIS_ADDR", "localhost:6379")
	redisPassword := getEnv("REDIS_PASSWORD", "")
	redisDB := 0 // Using default DB
	redisListKey := getEnv("REDIS_LIST_KEY", "slack_messages")
	redisReactionListKey := getEnv("REDIS_REACTION_LIST_KEY", "slack_reactions")
	timeBombChannel := getEnv("TIMEBOMB_REDIS_CHANNEL", "timebomb-messages")
	slackToken := getEnv("SLACK_BOT_TOKEN", "")
	httpAddr := getEnv("HTTP_ADDR", ":8080")

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

	// Start HTTP server
	startHTTPServer(ctx, httpAddr, slackClient, rdb, timeBombChannel)

	// Start message processing loop
	log.Printf("Starting to listen for messages on Redis list '%s'...", redisListKey)
	go processMessages(ctx, rdb, slackClient, redisListKey, timeBombChannel)

	// Start reaction processing loop
	log.Printf("Starting to listen for reactions on Redis list '%s'...", redisReactionListKey)
	go processReactions(ctx, rdb, slackClient, redisReactionListKey)

	// Wait for shutdown signal
	<-sigChan
	log.Println("Shutting down gracefully...")
	cancel()
	time.Sleep(1 * time.Second) // Give goroutines time to finish current operation
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
