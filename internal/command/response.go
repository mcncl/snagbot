package command

import (
	"encoding/json"
	"fmt"
)

// SlackResponse represents a response to be sent to Slack
type SlackResponse struct {
	ResponseType string `json:"response_type"`
	Text         string `json:"text"`
}

// NewEphemeralResponse creates a new ephemeral response (only visible to the user)
func NewEphemeralResponse(text string) *SlackResponse {
	return &SlackResponse{
		ResponseType: "ephemeral",
		Text:         text,
	}
}

// NewChannelResponse creates a new in-channel response (visible to everyone)
func NewChannelResponse(text string) *SlackResponse {
	return &SlackResponse{
		ResponseType: "in_channel",
		Text:         text,
	}
}

// ToJSON converts the response to JSON
func (r *SlackResponse) ToJSON() (string, error) {
	bytes, err := json.Marshal(r)
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}

// FormatSlackResponse formats a text response into a proper Slack response JSON
func FormatSlackResponse(text string, isEphemeral bool) string {
	var response *SlackResponse
	if isEphemeral {
		response = NewEphemeralResponse(text)
	} else {
		response = NewChannelResponse(text)
	}

	jsonStr, err := response.ToJSON()
	if err != nil {
		// Fallback if there's an error
		return fmt.Sprintf(`{"response_type": "ephemeral", "text": %q}`, text)
	}

	return jsonStr
}
