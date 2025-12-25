package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"

	"github.com/redis/go-redis/v9"
	"github.com/slack-go/slack"
)

// startHTTPServer starts the HTTP server for posting messages via HTTP endpoint
func startHTTPServer(ctx context.Context, addr string, slackClient *slack.Client, rdb *redis.Client, timeBombChannel string) *http.Server {
	mux := http.NewServeMux()

	// POST /message endpoint
	mux.HandleFunc("/message", func(w http.ResponseWriter, r *http.Request) {
		handlePostMessage(ctx, w, r, slackClient, rdb, timeBombChannel)
	})

	// Health check endpoint
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	server := &http.Server{
		Addr:    addr,
		Handler: mux,
	}

	// Start server in a goroutine
	go func() {
		log.Printf("Starting HTTP server on %s", addr)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Printf("HTTP server error: %v", err)
		}
	}()

	// Handle graceful shutdown
	go func() {
		<-ctx.Done()
		log.Println("Shutting down HTTP server...")
		if err := server.Shutdown(context.Background()); err != nil {
			log.Printf("Error shutting down HTTP server: %v", err)
		}
	}()

	return server
}

// handlePostMessage handles POST requests to send Slack messages
func handlePostMessage(ctx context.Context, w http.ResponseWriter, r *http.Request, slackClient *slack.Client, rdb *redis.Client, timeBombChannel string) {
	// Only allow POST requests
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Parse request body
	var msg SlackMessage
	if err := json.NewDecoder(r.Body).Decode(&msg); err != nil {
		log.Printf("Error parsing request body: %v", err)
		http.Error(w, "Invalid JSON payload", http.StatusBadRequest)
		return
	}

	// Send message to Slack
	channelID, timestamp, err := sendSlackMessageWithResponse(ctx, slackClient, rdb, msg, timeBombChannel)
	if err != nil {
		log.Printf("Error sending message to Slack: %v", err)
		if err == ErrInvalidMessage || err == ErrInvalidTTL {
			http.Error(w, err.Error(), http.StatusBadRequest)
		} else {
			http.Error(w, "Failed to send message to Slack", http.StatusInternalServerError)
		}
		return
	}

	// Build and send response
	response := MessageResponse{
		Channel: channelID,
		TS:      timestamp,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Printf("Error encoding response: %v", err)
	}
}
