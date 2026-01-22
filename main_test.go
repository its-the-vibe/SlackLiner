package main

import (
	"encoding/json"
	"testing"
)

func TestSlackMessageParsing(t *testing.T) {
	tests := []struct {
		name      string
		jsonInput string
		wantErr   bool
		validate  func(*testing.T, SlackMessage)
	}{
		{
			name:      "simple message without metadata",
			jsonInput: `{"channel":"#general","text":"Hello World"}`,
			wantErr:   false,
			validate: func(t *testing.T, msg SlackMessage) {
				if msg.Channel != "#general" {
					t.Errorf("Channel = %v, want #general", msg.Channel)
				}
				if msg.Text != "Hello World" {
					t.Errorf("Text = %v, want Hello World", msg.Text)
				}
				if msg.Metadata != nil {
					t.Errorf("Metadata should be nil for simple message")
				}
			},
		},
		{
			name: "message with metadata",
			jsonInput: `{
				"channel":"#general",
				"text":"Task created",
				"metadata":{
					"event_type":"task_created",
					"event_payload":{"task_id":"123","priority":"high"}
				}
			}`,
			wantErr: false,
			validate: func(t *testing.T, msg SlackMessage) {
				if msg.Channel != "#general" {
					t.Errorf("Channel = %v, want #general", msg.Channel)
				}
				if msg.Text != "Task created" {
					t.Errorf("Text = %v, want Task created", msg.Text)
				}
				if msg.Metadata == nil {
					t.Fatal("Metadata should not be nil")
				}
				if msg.Metadata.EventType != "task_created" {
					t.Errorf("EventType = %v, want task_created", msg.Metadata.EventType)
				}
				if taskID, ok := msg.Metadata.EventPayload["task_id"].(string); !ok || taskID != "123" {
					t.Errorf("task_id = %v, want 123", msg.Metadata.EventPayload["task_id"])
				}
				if priority, ok := msg.Metadata.EventPayload["priority"].(string); !ok || priority != "high" {
					t.Errorf("priority = %v, want high", msg.Metadata.EventPayload["priority"])
				}
			},
		},
		{
			name: "message with metadata and complex payload",
			jsonInput: `{
				"channel":"C1234567890",
				"text":"Order placed",
				"metadata":{
					"event_type":"order_created",
					"event_payload":{
						"order_id":"ORD-789",
						"amount":99.99,
						"items":3,
						"express":true
					}
				}
			}`,
			wantErr: false,
			validate: func(t *testing.T, msg SlackMessage) {
				if msg.Metadata == nil {
					t.Fatal("Metadata should not be nil")
				}
				if msg.Metadata.EventType != "order_created" {
					t.Errorf("EventType = %v, want order_created", msg.Metadata.EventType)
				}
				// Test number value
				if amount, ok := msg.Metadata.EventPayload["amount"].(float64); !ok || amount != 99.99 {
					t.Errorf("amount = %v, want 99.99", msg.Metadata.EventPayload["amount"])
				}
				// Test boolean value
				if express, ok := msg.Metadata.EventPayload["express"].(bool); !ok || !express {
					t.Errorf("express = %v, want true", msg.Metadata.EventPayload["express"])
				}
			},
		},
		{
			name:      "message with TTL",
			jsonInput: `{"channel":"#general","text":"Temporary message","ttl":3600}`,
			wantErr:   false,
			validate: func(t *testing.T, msg SlackMessage) {
				if msg.Channel != "#general" {
					t.Errorf("Channel = %v, want #general", msg.Channel)
				}
				if msg.Text != "Temporary message" {
					t.Errorf("Text = %v, want Temporary message", msg.Text)
				}
				if msg.TTL != 3600 {
					t.Errorf("TTL = %v, want 3600", msg.TTL)
				}
			},
		},
		{
			name: "message with TTL and metadata",
			jsonInput: `{
				"channel":"#alerts",
				"text":"Alert notification",
				"ttl":300,
				"metadata":{
					"event_type":"alert_created",
					"event_payload":{"severity":"high"}
				}
			}`,
			wantErr: false,
			validate: func(t *testing.T, msg SlackMessage) {
				if msg.Channel != "#alerts" {
					t.Errorf("Channel = %v, want #alerts", msg.Channel)
				}
				if msg.TTL != 300 {
					t.Errorf("TTL = %v, want 300", msg.TTL)
				}
				if msg.Metadata == nil {
					t.Fatal("Metadata should not be nil")
				}
				if msg.Metadata.EventType != "alert_created" {
					t.Errorf("EventType = %v, want alert_created", msg.Metadata.EventType)
				}
			},
		},
		{
			name:      "message with zero TTL (should be ignored)",
			jsonInput: `{"channel":"#general","text":"No deletion","ttl":0}`,
			wantErr:   false,
			validate: func(t *testing.T, msg SlackMessage) {
				if msg.TTL != 0 {
					t.Errorf("TTL = %v, want 0", msg.TTL)
				}
			},
		},
		{
			name:      "message with thread_ts",
			jsonInput: `{"channel":"#general","text":"Reply to thread","thread_ts":"1234567890.123456"}`,
			wantErr:   false,
			validate: func(t *testing.T, msg SlackMessage) {
				if msg.Channel != "#general" {
					t.Errorf("Channel = %v, want #general", msg.Channel)
				}
				if msg.Text != "Reply to thread" {
					t.Errorf("Text = %v, want Reply to thread", msg.Text)
				}
				if msg.ThreadTS != "1234567890.123456" {
					t.Errorf("ThreadTS = %v, want 1234567890.123456", msg.ThreadTS)
				}
			},
		},
		{
			name: "message with thread_ts and metadata",
			jsonInput: `{
				"channel":"#general",
				"text":"Thread reply with metadata",
				"thread_ts":"1234567890.123456",
				"metadata":{
					"event_type":"thread_reply",
					"event_payload":{"reply_type":"automated"}
				}
			}`,
			wantErr: false,
			validate: func(t *testing.T, msg SlackMessage) {
				if msg.ThreadTS != "1234567890.123456" {
					t.Errorf("ThreadTS = %v, want 1234567890.123456", msg.ThreadTS)
				}
				if msg.Metadata == nil {
					t.Fatal("Metadata should not be nil")
				}
				if msg.Metadata.EventType != "thread_reply" {
					t.Errorf("EventType = %v, want thread_reply", msg.Metadata.EventType)
				}
			},
		},
		{
			name:      "message with thread_ts and TTL",
			jsonInput: `{"channel":"#general","text":"Temporary thread reply","thread_ts":"1234567890.123456","ttl":300}`,
			wantErr:   false,
			validate: func(t *testing.T, msg SlackMessage) {
				if msg.ThreadTS != "1234567890.123456" {
					t.Errorf("ThreadTS = %v, want 1234567890.123456", msg.ThreadTS)
				}
				if msg.TTL != 300 {
					t.Errorf("TTL = %v, want 300", msg.TTL)
				}
			},
		},
		{
			name:      "invalid JSON",
			jsonInput: `{"channel":"#general"`,
			wantErr:   true,
		},
		{
			name: "message with blocks only",
			jsonInput: `{
				"channel":"#general",
				"blocks":[
					{
						"type":"section",
						"text":{
							"type":"mrkdwn",
							"text":"Hello, *world*! :tada:"
						}
					}
				]
			}`,
			wantErr: false,
			validate: func(t *testing.T, msg SlackMessage) {
				if msg.Channel != "#general" {
					t.Errorf("Channel = %v, want #general", msg.Channel)
				}
				if msg.Text != "" {
					t.Errorf("Text should be empty for blocks-only message, got: %v", msg.Text)
				}
				if len(msg.Blocks) == 0 {
					t.Fatal("Blocks should not be empty")
				}
			},
		},
		{
			name: "message with blocks and text",
			jsonInput: `{
				"channel":"#general",
				"text":"Fallback text",
				"blocks":[
					{
						"type":"section",
						"text":{
							"type":"mrkdwn",
							"text":"Rich *formatted* text"
						}
					}
				]
			}`,
			wantErr: false,
			validate: func(t *testing.T, msg SlackMessage) {
				if msg.Channel != "#general" {
					t.Errorf("Channel = %v, want #general", msg.Channel)
				}
				if msg.Text != "Fallback text" {
					t.Errorf("Text = %v, want Fallback text", msg.Text)
				}
				if len(msg.Blocks) == 0 {
					t.Fatal("Blocks should not be empty")
				}
			},
		},
		{
			name: "message with complex blocks",
			jsonInput: `{
				"channel":"#general",
				"blocks":[
					{
						"type":"section",
						"text":{
							"type":"mrkdwn",
							"text":"Hello, *world*! :tada:"
						}
					},
					{
						"type":"input",
						"element":{
							"type":"external_select",
							"placeholder":{
								"type":"plain_text",
								"text":"Select an item",
								"emoji":true
							},
							"action_id":"SlackCompose"
						},
						"label":{
							"type":"plain_text",
							"text":"Label",
							"emoji":true
						},
						"optional":false
					}
				]
			}`,
			wantErr: false,
			validate: func(t *testing.T, msg SlackMessage) {
				if msg.Channel != "#general" {
					t.Errorf("Channel = %v, want #general", msg.Channel)
				}
				if len(msg.Blocks) == 0 {
					t.Fatal("Blocks should not be empty")
				}
			},
		},
		{
			name: "message with blocks and metadata",
			jsonInput: `{
				"channel":"#general",
				"blocks":[
					{
						"type":"section",
						"text":{
							"type":"mrkdwn",
							"text":"Interactive message"
						}
					}
				],
				"metadata":{
					"event_type":"form_submitted",
					"event_payload":{"form_id":"123"}
				}
			}`,
			wantErr: false,
			validate: func(t *testing.T, msg SlackMessage) {
				if len(msg.Blocks) == 0 {
					t.Fatal("Blocks should not be empty")
				}
				if msg.Metadata == nil {
					t.Fatal("Metadata should not be nil")
				}
				if msg.Metadata.EventType != "form_submitted" {
					t.Errorf("EventType = %v, want form_submitted", msg.Metadata.EventType)
				}
			},
		},
		{
			name: "message with blocks and thread_ts",
			jsonInput: `{
				"channel":"#general",
				"blocks":[
					{
						"type":"section",
						"text":{
							"type":"mrkdwn",
							"text":"Thread reply with blocks"
						}
					}
				],
				"thread_ts":"1234567890.123456"
			}`,
			wantErr: false,
			validate: func(t *testing.T, msg SlackMessage) {
				if len(msg.Blocks) == 0 {
					t.Fatal("Blocks should not be empty")
				}
				if msg.ThreadTS != "1234567890.123456" {
					t.Errorf("ThreadTS = %v, want 1234567890.123456", msg.ThreadTS)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var msg SlackMessage
			err := json.Unmarshal([]byte(tt.jsonInput), &msg)
			if (err != nil) != tt.wantErr {
				t.Errorf("json.Unmarshal() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && tt.validate != nil {
				tt.validate(t, msg)
			}
		})
	}
}

func TestSlackMessageMarshaling(t *testing.T) {
	tests := []struct {
		name    string
		msg     SlackMessage
		wantErr bool
	}{
		{
			name: "simple message",
			msg: SlackMessage{
				Channel: "#general",
				Text:    "Hello",
			},
			wantErr: false,
		},
		{
			name: "message with metadata",
			msg: SlackMessage{
				Channel: "#general",
				Text:    "Task created",
				Metadata: &MessageMetadata{
					EventType: "task_created",
					EventPayload: map[string]interface{}{
						"task_id":  "123",
						"priority": "high",
					},
				},
			},
			wantErr: false,
		},
		{
			name: "message with TTL",
			msg: SlackMessage{
				Channel: "#general",
				Text:    "Temporary message",
				TTL:     3600,
			},
			wantErr: false,
		},
		{
			name: "message with TTL and metadata",
			msg: SlackMessage{
				Channel: "#alerts",
				Text:    "Alert notification",
				TTL:     300,
				Metadata: &MessageMetadata{
					EventType: "alert_created",
					EventPayload: map[string]interface{}{
						"severity": "high",
					},
				},
			},
			wantErr: false,
		},
		{
			name: "message with thread_ts",
			msg: SlackMessage{
				Channel:  "#general",
				Text:     "Reply to thread",
				ThreadTS: "1234567890.123456",
			},
			wantErr: false,
		},
		{
			name: "message with thread_ts and metadata",
			msg: SlackMessage{
				Channel:  "#general",
				Text:     "Thread reply with metadata",
				ThreadTS: "1234567890.123456",
				Metadata: &MessageMetadata{
					EventType: "thread_reply",
					EventPayload: map[string]interface{}{
						"reply_type": "automated",
					},
				},
			},
			wantErr: false,
		},
		{
			name: "message with thread_ts and TTL",
			msg: SlackMessage{
				Channel:  "#general",
				Text:     "Temporary thread reply",
				ThreadTS: "1234567890.123456",
				TTL:      300,
			},
			wantErr: false,
		},
		{
			name: "message with blocks only",
			msg: SlackMessage{
				Channel: "#general",
				Blocks:  json.RawMessage(`[{"type":"section","text":{"type":"mrkdwn","text":"Hello, *world*!"}}]`),
			},
			wantErr: false,
		},
		{
			name: "message with blocks and text",
			msg: SlackMessage{
				Channel: "#general",
				Text:    "Fallback text",
				Blocks:  json.RawMessage(`[{"type":"section","text":{"type":"mrkdwn","text":"Rich text"}}]`),
			},
			wantErr: false,
		},
		{
			name: "message with blocks and metadata",
			msg: SlackMessage{
				Channel: "#general",
				Blocks:  json.RawMessage(`[{"type":"section","text":{"type":"mrkdwn","text":"Interactive message"}}]`),
				Metadata: &MessageMetadata{
					EventType: "form_submitted",
					EventPayload: map[string]interface{}{
						"form_id": "123",
					},
				},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data, err := json.Marshal(tt.msg)
			if (err != nil) != tt.wantErr {
				t.Errorf("json.Marshal() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				// Verify we can unmarshal back
				var msg SlackMessage
				if err := json.Unmarshal(data, &msg); err != nil {
					t.Errorf("Failed to unmarshal back: %v", err)
				}
			}
		})
	}
}

func TestReactionMessageParsing(t *testing.T) {
	tests := []struct {
		name      string
		jsonInput string
		wantErr   bool
		validate  func(*testing.T, ReactionMessage)
	}{
		{
			name:      "valid reaction message",
			jsonInput: `{"reaction":"heart_eyes_cat","channel":"C1234567890","ts":"1766282873.772199"}`,
			wantErr:   false,
			validate: func(t *testing.T, msg ReactionMessage) {
				if msg.Reaction != "heart_eyes_cat" {
					t.Errorf("Reaction = %v, want heart_eyes_cat", msg.Reaction)
				}
				if msg.Channel != "C1234567890" {
					t.Errorf("Channel = %v, want C1234567890", msg.Channel)
				}
				if msg.TS != "1766282873.772199" {
					t.Errorf("TS = %v, want 1766282873.772199", msg.TS)
				}
				if msg.Remove {
					t.Errorf("Remove should be false for message without remove field")
				}
			},
		},
		{
			name:      "reaction with different emoji",
			jsonInput: `{"reaction":"thumbsup","channel":"#general","ts":"1234567890.123456"}`,
			wantErr:   false,
			validate: func(t *testing.T, msg ReactionMessage) {
				if msg.Reaction != "thumbsup" {
					t.Errorf("Reaction = %v, want thumbsup", msg.Reaction)
				}
				if msg.Channel != "#general" {
					t.Errorf("Channel = %v, want #general", msg.Channel)
				}
				if msg.TS != "1234567890.123456" {
					t.Errorf("TS = %v, want 1234567890.123456", msg.TS)
				}
			},
		},
		{
			name:      "reaction with tada emoji",
			jsonInput: `{"reaction":"tada","channel":"C9876543210","ts":"9999999999.999999"}`,
			wantErr:   false,
			validate: func(t *testing.T, msg ReactionMessage) {
				if msg.Reaction != "tada" {
					t.Errorf("Reaction = %v, want tada", msg.Reaction)
				}
			},
		},
		{
			name:      "reaction removal message",
			jsonInput: `{"reaction":"thumbsup","channel":"C1234567890","ts":"1766282873.772199","remove":true}`,
			wantErr:   false,
			validate: func(t *testing.T, msg ReactionMessage) {
				if msg.Reaction != "thumbsup" {
					t.Errorf("Reaction = %v, want thumbsup", msg.Reaction)
				}
				if msg.Channel != "C1234567890" {
					t.Errorf("Channel = %v, want C1234567890", msg.Channel)
				}
				if msg.TS != "1766282873.772199" {
					t.Errorf("TS = %v, want 1766282873.772199", msg.TS)
				}
				if !msg.Remove {
					t.Errorf("Remove = %v, want true", msg.Remove)
				}
			},
		},
		{
			name:      "reaction removal with remove false",
			jsonInput: `{"reaction":"heart","channel":"C1234567890","ts":"1766282873.772199","remove":false}`,
			wantErr:   false,
			validate: func(t *testing.T, msg ReactionMessage) {
				if msg.Reaction != "heart" {
					t.Errorf("Reaction = %v, want heart", msg.Reaction)
				}
				if msg.Remove {
					t.Errorf("Remove = %v, want false", msg.Remove)
				}
			},
		},
		{
			name:      "invalid JSON",
			jsonInput: `{"reaction":"heart"`,
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var msg ReactionMessage
			err := json.Unmarshal([]byte(tt.jsonInput), &msg)
			if (err != nil) != tt.wantErr {
				t.Errorf("json.Unmarshal() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && tt.validate != nil {
				tt.validate(t, msg)
			}
		})
	}
}

func TestReactionMessageMarshaling(t *testing.T) {
	tests := []struct {
		name    string
		msg     ReactionMessage
		wantErr bool
	}{
		{
			name: "simple reaction",
			msg: ReactionMessage{
				Reaction: "heart_eyes_cat",
				Channel:  "C1234567890",
				TS:       "1766282873.772199",
			},
			wantErr: false,
		},
		{
			name: "thumbsup reaction",
			msg: ReactionMessage{
				Reaction: "thumbsup",
				Channel:  "#general",
				TS:       "1234567890.123456",
			},
			wantErr: false,
		},
		{
			name: "tada reaction",
			msg: ReactionMessage{
				Reaction: "tada",
				Channel:  "C9876543210",
				TS:       "9999999999.999999",
			},
			wantErr: false,
		},
		{
			name: "reaction with remove true",
			msg: ReactionMessage{
				Reaction: "thumbsup",
				Channel:  "C1234567890",
				TS:       "1766282873.772199",
				Remove:   true,
			},
			wantErr: false,
		},
		{
			name: "reaction with remove false",
			msg: ReactionMessage{
				Reaction: "heart",
				Channel:  "C1234567890",
				TS:       "1766282873.772199",
				Remove:   false,
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data, err := json.Marshal(tt.msg)
			if (err != nil) != tt.wantErr {
				t.Errorf("json.Marshal() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				// Verify we can unmarshal back
				var msg ReactionMessage
				if err := json.Unmarshal(data, &msg); err != nil {
					t.Errorf("Failed to unmarshal back: %v", err)
				}
			}
		})
	}
}
