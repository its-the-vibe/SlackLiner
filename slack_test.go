package main

import (
	"context"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"

	"github.com/slack-go/slack"
)

func TestSendSlackMessageWithResponseValidatesDestination(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name string
		msg  SlackMessage
	}{
		{
			name: "missing channel and user_id",
			msg: SlackMessage{
				Text: "hello",
			},
		},
		{
			name: "both channel and user_id provided",
			msg: SlackMessage{
				Channel: "#general",
				UserID:  "U1234567890",
				Text:    "hello",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, _, err := sendSlackMessageWithResponse(ctx, nil, nil, tt.msg, "timebomb-channel")
			if err != ErrInvalidMessage {
				t.Fatalf("expected ErrInvalidMessage, got %v", err)
			}
		})
	}
}

func TestSendSlackMessageWithResponsePostsToChannelWithoutOpeningConversation(t *testing.T) {
	ctx := context.Background()

	var mu sync.Mutex
	openConversationCalls := 0
	postMessageCalls := 0
	lastPostedChannel := ""

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := r.ParseForm(); err != nil {
			t.Fatalf("failed to parse form: %v", err)
		}

		switch r.URL.Path {
		case "/conversations.open":
			openConversationCalls++
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"ok":true,"channel":{"id":"D1234567890"}}`))
		case "/chat.postMessage":
			mu.Lock()
			postMessageCalls++
			lastPostedChannel = r.PostForm.Get("channel")
			mu.Unlock()

			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"ok":true,"channel":"C1234567890","ts":"1234567890.123456","message":{"text":"ok"}}`))
		default:
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
	}))
	defer server.Close()

	slackClient := slack.New("test-token", slack.OptionAPIURL(server.URL+"/"))

	channelID, timestamp, err := sendSlackMessageWithResponse(ctx, slackClient, nil, SlackMessage{
		Channel: "#general",
		Text:    "hello channel",
	}, "timebomb-channel")
	if err != nil {
		t.Fatalf("sendSlackMessageWithResponse returned error: %v", err)
	}

	if channelID != "C1234567890" {
		t.Fatalf("channelID = %q, want %q", channelID, "C1234567890")
	}
	if timestamp != "1234567890.123456" {
		t.Fatalf("timestamp = %q, want %q", timestamp, "1234567890.123456")
	}
	if openConversationCalls != 0 {
		t.Fatalf("openConversationCalls = %d, want 0", openConversationCalls)
	}
	if postMessageCalls != 1 {
		t.Fatalf("postMessageCalls = %d, want 1", postMessageCalls)
	}

	mu.Lock()
	defer mu.Unlock()
	if lastPostedChannel != "#general" {
		t.Fatalf("lastPostedChannel = %q, want %q", lastPostedChannel, "#general")
	}
}

func TestSendSlackMessageWithResponseResolvesDMChannelFromUserID(t *testing.T) {
	ctx := context.Background()

	var mu sync.Mutex
	openConversationCalls := 0
	postMessageCalls := 0
	lastPostedChannel := ""
	lastOpenConversationUsers := ""

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := r.ParseForm(); err != nil {
			t.Fatalf("failed to parse form: %v", err)
		}

		switch r.URL.Path {
		case "/conversations.open":
			mu.Lock()
			openConversationCalls++
			lastOpenConversationUsers = r.PostForm.Get("users")
			mu.Unlock()

			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"ok":true,"channel":{"id":"D1234567890"}}`))
		case "/chat.postMessage":
			mu.Lock()
			postMessageCalls++
			lastPostedChannel = r.PostForm.Get("channel")
			mu.Unlock()

			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"ok":true,"channel":"D1234567890","ts":"1234567890.123456","message":{"text":"ok"}}`))
		default:
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
	}))
	defer server.Close()

	slackClient := slack.New("test-token", slack.OptionAPIURL(server.URL+"/"))

	channelID, timestamp, err := sendSlackMessageWithResponse(ctx, slackClient, nil, SlackMessage{
		UserID: "U1234567890",
		Text:   "hello DM",
	}, "timebomb-channel")
	if err != nil {
		t.Fatalf("sendSlackMessageWithResponse returned error: %v", err)
	}

	if channelID != "D1234567890" {
		t.Fatalf("channelID = %q, want %q", channelID, "D1234567890")
	}
	if timestamp != "1234567890.123456" {
		t.Fatalf("timestamp = %q, want %q", timestamp, "1234567890.123456")
	}

	mu.Lock()
	defer mu.Unlock()
	if openConversationCalls != 1 {
		t.Fatalf("openConversationCalls = %d, want 1", openConversationCalls)
	}
	if postMessageCalls != 1 {
		t.Fatalf("postMessageCalls = %d, want 1", postMessageCalls)
	}
	if lastOpenConversationUsers != "U1234567890" {
		t.Fatalf("lastOpenConversationUsers = %q, want %q", lastOpenConversationUsers, "U1234567890")
	}
	if lastPostedChannel != "D1234567890" {
		t.Fatalf("lastPostedChannel = %q, want %q", lastPostedChannel, "D1234567890")
	}
}
