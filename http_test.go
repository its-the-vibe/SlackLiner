package main

import (
	"encoding/json"
	"testing"
)

func TestMessageResponseParsing(t *testing.T) {
	tests := []struct {
		name      string
		jsonInput string
		wantErr   bool
		validate  func(*testing.T, MessageResponse)
	}{
		{
			name:      "valid message response",
			jsonInput: `{"channel":"C1234567890","ts":"1766282873.772199"}`,
			wantErr:   false,
			validate: func(t *testing.T, resp MessageResponse) {
				if resp.Channel != "C1234567890" {
					t.Errorf("Channel = %v, want C1234567890", resp.Channel)
				}
				if resp.TS != "1766282873.772199" {
					t.Errorf("TS = %v, want 1766282873.772199", resp.TS)
				}
			},
		},
		{
			name:      "response with channel name",
			jsonInput: `{"channel":"#general","ts":"1234567890.123456"}`,
			wantErr:   false,
			validate: func(t *testing.T, resp MessageResponse) {
				if resp.Channel != "#general" {
					t.Errorf("Channel = %v, want #general", resp.Channel)
				}
				if resp.TS != "1234567890.123456" {
					t.Errorf("TS = %v, want 1234567890.123456", resp.TS)
				}
			},
		},
		{
			name:      "invalid JSON",
			jsonInput: `{"channel":"C1234567890"`,
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var resp MessageResponse
			err := json.Unmarshal([]byte(tt.jsonInput), &resp)
			if (err != nil) != tt.wantErr {
				t.Errorf("json.Unmarshal() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && tt.validate != nil {
				tt.validate(t, resp)
			}
		})
	}
}

func TestMessageResponseMarshaling(t *testing.T) {
	tests := []struct {
		name    string
		resp    MessageResponse
		wantErr bool
	}{
		{
			name: "simple response",
			resp: MessageResponse{
				Channel: "C1234567890",
				TS:      "1766282873.772199",
			},
			wantErr: false,
		},
		{
			name: "response with channel name",
			resp: MessageResponse{
				Channel: "#general",
				TS:      "1234567890.123456",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data, err := json.Marshal(tt.resp)
			if (err != nil) != tt.wantErr {
				t.Errorf("json.Marshal() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				// Verify we can unmarshal back
				var resp MessageResponse
				if err := json.Unmarshal(data, &resp); err != nil {
					t.Errorf("Failed to unmarshal back: %v", err)
				}
				// Verify the values match
				if resp.Channel != tt.resp.Channel {
					t.Errorf("Channel = %v, want %v", resp.Channel, tt.resp.Channel)
				}
				if resp.TS != tt.resp.TS {
					t.Errorf("TS = %v, want %v", resp.TS, tt.resp.TS)
				}
			}
		})
	}
}
