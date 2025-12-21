package main

import (
	"context"
	"encoding/json"
	"log"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/slack-go/slack"
)

// processMessages reads messages from Redis list and sends them to Slack
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

			// Send message to Slack
			sendSlackMessage(ctx, slackClient, rdb, msg, timeBombChannel)
		}
	}
}

// processReactions reads reaction messages from Redis list and adds reactions to Slack messages
func processReactions(ctx context.Context, rdb *redis.Client, slackClient *slack.Client, listKey string) {
	for {
		select {
		case <-ctx.Done():
			log.Println("Reaction processing stopped")
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

			// Parse the reaction message
			var msg ReactionMessage
			if err := json.Unmarshal([]byte(messageData), &msg); err != nil {
				log.Printf("Error parsing reaction message JSON: %v, data: %s", err, messageData)
				continue
			}

			// Add reaction to Slack
			addSlackReaction(slackClient, msg)
		}
	}
}
